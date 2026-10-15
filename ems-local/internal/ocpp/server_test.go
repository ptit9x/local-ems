package ocpp

import (
	"encoding/json"
	"testing"
	"time"
)

type mockHandler struct{}

func (m *mockHandler) OnBootNotification(cpID string, req BootNotificationReq) BootNotificationConf {
	return BootNotificationConf{Status: "Accepted", Interval: 300, CurrentTime: time.Now().Format(time.RFC3339)}
}
func (m *mockHandler) OnHeartbeat(cpID string) HeartbeatConf {
	return HeartbeatConf{CurrentTime: time.Now().Format(time.RFC3339)}
}
func (m *mockHandler) OnStatusNotification(cpID string, req StatusNotificationReq) StatusNotificationConf {
	return StatusNotificationConf{}
}
func (m *mockHandler) OnMeterValues(cpID string, req MeterValuesReq) MeterValuesConf {
	return MeterValuesConf{}
}
func (m *mockHandler) OnStartTransaction(cpID string, req StartTransactionReq) StartTransactionConf {
	return StartTransactionConf{TransactionId: 1, IdTagInfo: IdTagInfo{Status: "Accepted"}}
}
func (m *mockHandler) OnStopTransaction(cpID string, req StopTransactionReq) StopTransactionConf {
	return StopTransactionConf{IdTagInfo: &IdTagInfo{Status: "Accepted"}}
}

func TestCentralSystem_StartStop(t *testing.T) {
	cs := NewCentralSystem("127.0.0.1:0", &mockHandler{})
	err := cs.Start()
	if err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	err = cs.Stop()
	if err != nil {
		t.Fatalf("failed to stop: %v", err)
	}
}

func TestChargePointActionHandling(t *testing.T) {
	cs := NewCentralSystem(":0", &mockHandler{})
	
	// Create dummy charge point
	cp := &ChargePoint{
		ID:              "cp1",
		ConnectorStatus: make(map[int]string),
	}

	// Test BootNotification handling
	payload := `{"chargePointVendor":"VendorX", "chargePointModel":"ModelY"}`
	res := cs.handleAction(cp, "BootNotification", json.RawMessage(payload))
	
	conf, ok := res.(BootNotificationConf)
	if !ok {
		t.Fatalf("expected BootNotificationConf, got %T", res)
	}
	if conf.Status != "Accepted" {
		t.Errorf("expected Accepted, got %s", conf.Status)
	}
	if cp.Vendor != "VendorX" {
		t.Errorf("expected vendor VendorX, got %s", cp.Vendor)
	}

	// Test Heartbeat
	res2 := cs.handleAction(cp, "Heartbeat", json.RawMessage(`{}`))
	_, ok = res2.(HeartbeatConf)
	if !ok {
		t.Fatalf("expected HeartbeatConf")
	}

	// Test StatusNotification
	statusPayload := `{"connectorId": 1, "errorCode": "NoError", "status": "Charging"}`
	cs.handleAction(cp, "StatusNotification", json.RawMessage(statusPayload))
	if cp.ConnectorStatus[1] != "Charging" {
		t.Errorf("expected Charging, got %s", cp.ConnectorStatus[1])
	}
}
