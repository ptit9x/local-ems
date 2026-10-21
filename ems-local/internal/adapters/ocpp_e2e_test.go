package adapters

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vmo/ems-platform/ems-local/internal/devices"
	"github.com/vmo/ems-platform/ems-local/internal/ocpp"
)

func TestSixOCPPChargersLifecycleAndDLMCommands(t *testing.T) {
	central := ocpp.NewCentralSystem("127.0.0.1:0", nil)
	if err := central.Start(); err != nil {
		t.Fatal(err)
	}
	defer central.Stop()
	factory := NewFactory(central)

	var clients []*websocket.Conn
	var adapters []devices.SnapshotDevice
	for i := 1; i <= 6; i++ {
		id := fmt.Sprintf("ev-%d", i)
		conn, _, err := websocket.DefaultDialer.Dial("ws://"+central.Address()+"/ocpp/"+id, nil)
		if err != nil {
			t.Fatalf("connect %s: %v", id, err)
		}
		clients = append(clients, conn)
		chargerCall(t, conn, fmt.Sprintf("boot-%d", i), "BootNotification", map[string]interface{}{"chargePointVendor": "Simulator", "chargePointModel": "E2E"})
		chargerCall(t, conn, fmt.Sprintf("status-%d", i), "StatusNotification", map[string]interface{}{"connectorId": 1, "errorCode": "NoError", "status": "Charging"})
		chargerCall(t, conn, fmt.Sprintf("start-%d", i), "StartTransaction", map[string]interface{}{"connectorId": 1, "idTag": "E2E", "meterStart": 0, "timestamp": time.Now().UTC().Format(time.RFC3339)})
		chargerCall(t, conn, fmt.Sprintf("meter-%d", i), "MeterValues", map[string]interface{}{"connectorId": 1, "transactionId": i, "meterValue": []interface{}{map[string]interface{}{"timestamp": time.Now().UTC().Format(time.RFC3339), "sampledValue": []interface{}{map[string]interface{}{"value": fmt.Sprintf("%d", 6000+i*100), "measurand": "Power.Active.Import", "unit": "W"}}}}})
		adapter, err := factory.Create(devices.DeviceConfig{ID: id, Type: devices.DeviceTypeEVCharger, Protocol: "ocpp"})
		if err != nil {
			t.Fatal(err)
		}
		if err := adapter.Poll(); err != nil {
			t.Fatal(err)
		}
		if got := adapter.Snapshot()["active_power_w"].Value.(int); got != 6000+i*100 {
			t.Fatalf("%s power=%d", id, got)
		}
		adapters = append(adapters, adapter)
	}
	defer func() {
		for _, conn := range clients {
			_ = conn.Close()
		}
	}()

	var wg sync.WaitGroup
	for _, conn := range clients {
		wg.Add(1)
		go func(c *websocket.Conn) {
			defer wg.Done()
			_, raw, err := c.ReadMessage()
			if err != nil {
				return
			}
			var call []json.RawMessage
			if json.Unmarshal(raw, &call) != nil || len(call) < 3 {
				return
			}
			var uid, action string
			_ = json.Unmarshal(call[1], &uid)
			_ = json.Unmarshal(call[2], &action)
			if action != "SetChargingProfile" {
				return
			}
			_ = c.WriteJSON([]interface{}{ocpp.MessageTypeCallResult, uid, map[string]interface{}{"status": "Accepted"}})
		}(conn)
	}
	for i, adapter := range adapters {
		control := adapter.(devices.ControllableDevice)
		if err := control.Write(devices.Command{Name: "set_max_power_w", Params: map[string]interface{}{"watts": 7000 + i*100}}); err != nil {
			t.Fatalf("command %d: %v", i, err)
		}
	}
	wg.Wait()
	for i, conn := range clients {
		chargerCall(t, conn, fmt.Sprintf("stop-%d", i+1), "StopTransaction", map[string]interface{}{"transactionId": i + 1, "meterStop": 7000, "timestamp": time.Now().UTC().Format(time.RFC3339), "reason": "Local"})
	}
	if got := len(central.ListChargePoints()); got != 6 {
		t.Fatalf("connected chargers=%d", got)
	}
}

func chargerCall(t *testing.T, conn *websocket.Conn, uid, action string, payload interface{}) {
	t.Helper()
	if err := conn.WriteJSON([]interface{}{ocpp.MessageTypeCall, uid, action, payload}); err != nil {
		t.Fatal(err)
	}
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var response []json.RawMessage
	if err := json.Unmarshal(raw, &response); err != nil || len(response) < 3 {
		t.Fatalf("invalid %s response: %s", action, raw)
	}
	var messageType int
	_ = json.Unmarshal(response[0], &messageType)
	if messageType != ocpp.MessageTypeCallResult {
		t.Fatalf("%s response type=%d", action, messageType)
	}
}
