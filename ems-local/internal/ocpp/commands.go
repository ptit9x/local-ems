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
		"connectorId":        connectorId,
		"csChargingProfiles": profile,
	}
	return cp.call("SetChargingProfile", req)
}

func (cp *ChargePoint) call(action string, payload interface{}) error {
	return cp.callWithPolicy(action, payload, 3, 10*time.Second)
}

func (cp *ChargePoint) callWithPolicy(action string, payload interface{}, attempts int, timeout time.Duration) error {
	if attempts < 1 {
		attempts = 1
	}
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

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := writeWSFrame(cp.conn, b); err != nil {
			lastErr = err
			continue
		}
		select {
		case response := <-ch:
			if err := validateCallResponse(action, response); err != nil {
				return err
			}
			return nil
		case <-time.After(timeout):
			lastErr = fmt.Errorf("timeout waiting for %s response", action)
		}
	}
	return fmt.Errorf("%s failed after %d attempts: %w", action, attempts, lastErr)
}

func validateCallResponse(action string, response []byte) error {
	var parts []json.RawMessage
	if err := json.Unmarshal(response, &parts); err != nil || len(parts) < 3 {
		return fmt.Errorf("%s: invalid OCPP response", action)
	}
	var messageType int
	_ = json.Unmarshal(parts[0], &messageType)
	if messageType == MessageTypeCallError {
		var code, description string
		_ = json.Unmarshal(parts[2], &code)
		if len(parts) > 3 {
			_ = json.Unmarshal(parts[3], &description)
		}
		return fmt.Errorf("%s: OCPP error %s: %s", action, code, description)
	}
	if messageType != MessageTypeCallResult {
		return fmt.Errorf("%s: unexpected OCPP message type %d", action, messageType)
	}
	var payloadResult map[string]interface{}
	if len(parts) > 2 && json.Unmarshal(parts[2], &payloadResult) == nil {
		if status, ok := payloadResult["status"].(string); ok && status != "Accepted" {
			return fmt.Errorf("%s rejected: %s", action, status)
		}
	}
	return nil
}
