package ocpp

import (
	"encoding/json"
	"fmt"
	"time"
)

// RemoteStartTransaction sends a start charging command.
func (cp *ChargePoint) RemoteStartTransaction(connectorId int, idTag string) error {
	req := map[string]interface{}{
		"connectorId": connectorId,
		"idTag":       idTag,
	}
	return cp.call("RemoteStartTransaction", req)
}

// RemoteStopTransaction sends a stop charging command.
func (cp *ChargePoint) RemoteStopTransaction(transactionId int) error {
	req := map[string]interface{}{
		"transactionId": transactionId,
	}
	return cp.call("RemoteStopTransaction", req)
}

// SetChargingProfile sets a power limit.
type ChargingProfile struct {
	ChargingProfileId      int              `json:"chargingProfileId"`
	StackLevel             int              `json:"stackLevel"`
	ChargingProfilePurpose string           `json:"chargingProfilePurpose"`
	ChargingProfileKind    string           `json:"chargingProfileKind"`
	ChargingSchedule       ChargingSchedule `json:"chargingSchedule"`
}
type ChargingSchedule struct {
	ChargingRateUnit       string                   `json:"chargingRateUnit"` // "W" or "A"
	ChargingSchedulePeriod []ChargingSchedulePeriod `json:"chargingSchedulePeriod"`
}
type ChargingSchedulePeriod struct {
	StartPeriod int     `json:"startPeriod"`
	Limit       float64 `json:"limit"`
}

func (cp *ChargePoint) SetChargingProfile(connectorId int, profile ChargingProfile) error {
	req := map[string]interface{}{
		"connectorId":     connectorId,
		"csChargingProfiles": profile,
	}
	return cp.call("SetChargingProfile", req)
}

func (cp *ChargePoint) call(action string, payload interface{}) error {
	cp.mu.Lock()
	cp.txCounter++
	msgID := fmt.Sprintf("%d", cp.txCounter)
	ch := make(chan []byte, 1)
	cp.pending[msgID] = ch
	cp.mu.Unlock()

	defer func() {
		cp.mu.Lock()
		delete(cp.pending, msgID)
		cp.mu.Unlock()
	}()

	msg := []interface{}{MessageTypeCall, msgID, action, payload}
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if err := writeWSFrame(cp.conn, b); err != nil {
		return err
	}

	select {
	case <-ch:
		// simple implementation: assuming success if we got a reply
		return nil
	case <-time.After(10 * time.Second):
		return fmt.Errorf("timeout waiting for %s response", action)
	}
}
