package topology

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vmo/ems-platform/ems-local/internal/config"
)

func TestBuildFromConfig(t *testing.T) {
	topoCfg := config.TopologyConfig{
		PCCMeter: "meter-1",
		BESSGroups: []config.BESSGroupConfig{
			{Name: "group1", BMS: []string{"bms-1"}, PCS: []string{"pcs-1"}},
		},
	}
	devicesCfg := config.DevicesConfig{
		Meter: []config.MeterDeviceConfig{{ID: "meter-1", Role: "grid"}},
		BMS:   []config.BMSDeviceConfig{{ID: "bms-1"}},
		PCS:   []config.PCSDeviceConfig{{ID: "pcs-1"}},
	}

	mgr, err := BuildFromConfig(topoCfg, devicesCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mgr.GetPCCMeterID() != "meter-1" {
		t.Errorf("expected pcc meter meter-1, got %s", mgr.GetPCCMeterID())
	}
}

func TestValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		topo       Topology
		devices    map[string]DeviceInfo
		wantErrStr string
	}{
		{
			name: "missing pcc meter",
			topo: Topology{},
			wantErrStr: "pcc_meter is required",
		},
		{
			name: "pcc meter not registered",
			topo: Topology{PCCMeter: "meter-1"},
			wantErrStr: "pcc_meter meter-1 not registered",
		},
		{
			name: "pcc meter wrong type",
			topo: Topology{PCCMeter: "meter-1"},
			devices: map[string]DeviceInfo{"meter-1": {ID: "meter-1", Type: "BMS"}},
			wantErrStr: "pcc_meter meter-1 must be a Meter",
		},
		{
			name: "pcs not registered",
			topo: Topology{
				PCCMeter: "meter-1",
				BESSGroups: []BESSGroup{{Name: "g1", PCS: []string{"pcs-1"}}},
			},
			devices: map[string]DeviceInfo{"meter-1": {ID: "meter-1", Type: "Meter"}},
			wantErrStr: "pcs pcs-1 in group g1 not registered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(tt.topo)
			for _, d := range tt.devices {
				mgr.RegisterDevice(d)
			}
			err := mgr.Validate()
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != tt.wantErrStr {
				t.Errorf("expected error %q, got %q", tt.wantErrStr, err.Error())
			}
		})
	}
}

func TestUpdateTopology(t *testing.T) {
	mgr := NewManager(Topology{PCCMeter: "meter-1"})
	mgr.RegisterDevice(DeviceInfo{ID: "meter-1", Type: "Meter"})
	mgr.RegisterDevice(DeviceInfo{ID: "meter-2", Type: "Meter"})
	mgr.RegisterDevice(DeviceInfo{ID: "bms-1", Type: "BMS"})

	if err := mgr.Update(Topology{PCCMeter: "meter-2"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mgr.GetPCCMeterID() != "meter-2" {
		t.Errorf("expected pcc meter meter-2, got %s", mgr.GetPCCMeterID())
	}

	err := mgr.Update(Topology{PCCMeter: "bms-1"})
	if err == nil {
		t.Fatalf("expected error updating to wrong type")
	}
	if mgr.GetPCCMeterID() != "meter-2" {
		t.Errorf("expected pcc meter to remain meter-2, got %s", mgr.GetPCCMeterID())
	}
}

func TestAPIHandlers(t *testing.T) {
	mgr := NewManager(Topology{PCCMeter: "meter-1"})
	mgr.RegisterDevice(DeviceInfo{ID: "meter-1", Type: "Meter"})
	mgr.RegisterDevice(DeviceInfo{ID: "meter-2", Type: "Meter"})

	mux := http.NewServeMux()
	RegisterAPI(mux, mgr)

	// Test GET /api/topology
	req := httptest.NewRequest("GET", "/api/topology", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	var topo Topology
	json.NewDecoder(rr.Body).Decode(&topo)
	if topo.PCCMeter != "meter-1" {
		t.Errorf("expected pcc meter meter-1, got %s", topo.PCCMeter)
	}

	// Test PUT /api/topology
	newTopo := Topology{PCCMeter: "meter-2"}
	body, _ := json.Marshal(newTopo)
	req = httptest.NewRequest("PUT", "/api/topology", bytes.NewReader(body))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if mgr.GetPCCMeterID() != "meter-2" {
		t.Errorf("expected pcc meter updated to meter-2, got %s", mgr.GetPCCMeterID())
	}
}
