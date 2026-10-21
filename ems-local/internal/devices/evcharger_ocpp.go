package devices

import (
	"fmt"
	"sync"

	"github.com/vmo/ems-platform/ems-local/internal/ocpp"
)

type OCPPEVChargerAdapter struct {
	cpID    string
	central *ocpp.CentralSystem
	mu      sync.RWMutex
	data    EVChargerData
}

func NewOCPPEVChargerAdapter(cpID string, central *ocpp.CentralSystem) *OCPPEVChargerAdapter {
	return &OCPPEVChargerAdapter{
		cpID:    cpID,
		central: central,
	}
}

func (a *OCPPEVChargerAdapter) Poll() error {
	cp := a.central.GetChargePoint(a.cpID)
	if cp == nil {
		return fmt.Errorf("charge point %s not connected", a.cpID)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	snapshot := cp.Snapshot()
	// Map OCPP status to internal status
	switch snapshot.Status {
	case "Available":
		a.data.Status = StatusAvailable
	case "Charging":
		a.data.Status = StatusCharging
	case "SuspendedEV":
		a.data.Status = StatusSuspendedEV
	case "SuspendedEVSE":
		a.data.Status = StatusSuspendedEVSE
	case "Finishing":
		a.data.Status = StatusFinishing
	case "Faulted":
		a.data.Status = StatusFault
	}
	a.data.ActivePower = snapshot.ActivePowerW
	a.data.EnergyDelivered = snapshot.EnergyWh
	a.data.CurrentL1 = snapshot.CurrentL1A
	a.data.CurrentL2 = snapshot.CurrentL2A
	a.data.CurrentL3 = snapshot.CurrentL3A
	a.data.VehicleConnected = snapshot.Status != "Available" && snapshot.Status != "Unavailable"

	return nil
}

func (a *OCPPEVChargerAdapter) Data() EVChargerData {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.data
}

func (a *OCPPEVChargerAdapter) WriteMaxPower(watts int) error {
	cp := a.central.GetChargePoint(a.cpID)
	if cp == nil {
		return fmt.Errorf("charge point %s not connected", a.cpID)
	}

	profile := ocpp.ChargingProfile{
		ChargingProfileId:      1,
		StackLevel:             0,
		ChargingProfilePurpose: "ChargePointMaxProfile",
		ChargingProfileKind:    "Absolute",
		ChargingSchedule: ocpp.ChargingSchedule{
			ChargingRateUnit: "W",
			ChargingSchedulePeriod: []ocpp.ChargingSchedulePeriod{
				{
					StartPeriod: 0,
					Limit:       float64(watts),
				},
			},
		},
	}

	return cp.SetChargingProfile(0, profile) // connector 0 applies to whole CP
}

func (a *OCPPEVChargerAdapter) WriteStartStop(start bool) error {
	cp := a.central.GetChargePoint(a.cpID)
	if cp == nil {
		return fmt.Errorf("charge point %s not connected", a.cpID)
	}

	if start {
		return cp.RemoteStartTransaction(1, "EMS")
	} else {
		transactionID := cp.Snapshot().TransactionID
		if transactionID == 0 {
			return fmt.Errorf("charge point %s has no active transaction", a.cpID)
		}
		return cp.RemoteStopTransaction(transactionID)
	}
}
