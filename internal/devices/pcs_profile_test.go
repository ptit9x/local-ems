package devices

import (
	"testing"
	"time"

	"github.com/vmo/local-ems/internal/modbus"
)

func TestDefaultRegistry_HasProfiles(t *testing.T) {
	r := DefaultPCSProfileRegistry()
	
	brands := []string{"generic", "sungrow", "sinexcel"}
	for _, brand := range brands {
		if _, err := r.Get(brand); err != nil {
			t.Errorf("expected to find profile %s, got error: %v", brand, err)
		}
	}
}

func TestRegistry_GetUnknown(t *testing.T) {
	r := DefaultPCSProfileRegistry()
	_, err := r.Get("unknown")
	if err == nil {
		t.Errorf("expected error getting unknown profile, got nil")
	}
}

func TestRegistry_Register(t *testing.T) {
	r := &PCSProfileRegistry{profiles: make(map[string]PCSProfile)}
	p := PCSProfile{Brand: "custom"}
	err := r.Register(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	got, err := r.Get("custom")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Brand != "custom" {
		t.Errorf("expected brand custom, got %s", got.Brand)
	}
}

func TestRegistry_List(t *testing.T) {
	r := DefaultPCSProfileRegistry()
	list := r.List()
	if len(list) != 3 {
		t.Errorf("expected 3 profiles, got %d", len(list))
	}
}

func TestProfiledPCSAdapter_Poll(t *testing.T) {
	srv := modbus.NewServer(":0")
	srv.Start()
	defer srv.Stop()
	
	unitID := byte(1)
	srv.AllocateRegisters(unitID, 20)
	
	// Default generic registers:
	// ActivePowerActReg: 1
	// ReactivePowerReg: 2
	// StatusReg: 3
	
	srv.SetRegister(unitID, 1, uint16(500)) // ActPower
	srv.SetRegister(unitID, 2, uint16(200)) // ReactPower
	srv.SetRegister(unitID, 3, uint16(1))   // Status
	
	client := modbus.NewClient(srv.Addr(), 1*time.Second)
	if err := client.Connect(); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()
	
	r := DefaultPCSProfileRegistry()
	prof := r.DefaultProfile()
	
	adapter := NewProfiledPCSAdapter(client, unitID, prof)
	err := adapter.Poll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	data := adapter.Data()
	if data.ActualActivePowerW != 500 {
		t.Errorf("expected ActualActivePowerW 500, got %d", data.ActualActivePowerW)
	}
	if data.ReactivePowerVar != 200 {
		t.Errorf("expected ReactivePowerVar 200, got %d", data.ReactivePowerVar)
	}
	if data.Status != PCSStatus(1) {
		t.Errorf("expected Status 1, got %v", data.Status)
	}
}

func TestProfiledPCSAdapter_WriteActivePower(t *testing.T) {
	srv := modbus.NewServer(":0")
	srv.Start()
	defer srv.Stop()
	
	unitID := byte(1)
	srv.AllocateRegisters(unitID, 20)
	
	client := modbus.NewClient(srv.Addr(), 1*time.Second)
	if err := client.Connect(); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()
	
	r := DefaultPCSProfileRegistry()
	prof := r.DefaultProfile()
	// prof.ActivePowerSetReg is 0
	
	adapter := NewProfiledPCSAdapter(client, unitID, prof)
	err := adapter.WriteActivePower(1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	val := srv.GetRegister(unitID, 0)
	
	if val != 1000 {
		t.Errorf("expected register value 1000, got %d", val)
	}
}
