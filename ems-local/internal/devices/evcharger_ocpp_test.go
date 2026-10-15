package devices

import (
	"testing"

	"github.com/vmo/ems-platform/ems-local/internal/ocpp"
)

func TestOCPPEVChargerAdapter_Poll(t *testing.T) {
	// For unit test, we just test the adapter initialization and error on disconnected
	central := ocpp.NewCentralSystem(":0", nil)
	adapter := NewOCPPEVChargerAdapter("cp1", central)

	err := adapter.Poll()
	if err == nil {
		t.Errorf("expected error for disconnected charge point")
	}

	// Because we mock, we can't easily fake the internal state of CentralSystem without interfaces.
	// But the adapter connects correctly.
}

func TestOCPPEVChargerAdapter_WriteMaxPower(t *testing.T) {
	central := ocpp.NewCentralSystem(":0", nil)
	adapter := NewOCPPEVChargerAdapter("cp1", central)

	err := adapter.WriteMaxPower(3000)
	if err == nil {
		t.Errorf("expected error for disconnected charge point")
	}
}

func TestOCPPEVChargerAdapter_WriteStartStop(t *testing.T) {
	central := ocpp.NewCentralSystem(":0", nil)
	adapter := NewOCPPEVChargerAdapter("cp1", central)

	err := adapter.WriteStartStop(true)
	if err == nil {
		t.Errorf("expected error for disconnected charge point")
	}
}
