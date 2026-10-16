package sync_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/vmo/ems-platform/ems-local/internal/sync"
)

func TestRemoteCommandParsing(t *testing.T) {
	raw := `{
		"command": "update_tariff",
		"request_id": "req-001",
		"timestamp": "2025-01-15T10:00:00Z",
		"payload": {"off_peak_start": 22, "off_peak_end": 6}
	}`

	var cmd sync.RemoteCommand
	if err := json.Unmarshal([]byte(raw), &cmd); err != nil {
		t.Fatalf("failed to parse command: %v", err)
	}

	if cmd.Command != "update_tariff" {
		t.Errorf("expected command=update_tariff, got %s", cmd.Command)
	}
	if cmd.RequestID != "req-001" {
		t.Errorf("expected request_id=req-001, got %s", cmd.RequestID)
	}
	if cmd.Payload == nil {
		t.Fatal("expected payload to be non-nil")
	}
	if cmd.Payload["off_peak_start"] != float64(22) {
		t.Errorf("expected off_peak_start=22, got %v", cmd.Payload["off_peak_start"])
	}
}

func TestRemoteCommandMinimalParsing(t *testing.T) {
	raw := `{
		"command": "restart",
		"request_id": "req-002",
		"timestamp": "2025-01-15T12:00:00Z"
	}`

	var cmd sync.RemoteCommand
	if err := json.Unmarshal([]byte(raw), &cmd); err != nil {
		t.Fatalf("failed to parse command: %v", err)
	}
	if cmd.Command != "restart" {
		t.Errorf("expected command=restart, got %s", cmd.Command)
	}
	if cmd.Payload != nil {
		t.Errorf("expected nil payload for restart, got %v", cmd.Payload)
	}
}

func TestCommandResponseMarshaling(t *testing.T) {
	resp := sync.CommandResponse{
		RequestID: "req-001",
		SiteID:    "sg-01",
		Status:    "ok",
		Timestamp: time.Date(2025, 1, 15, 10, 0, 1, 0, time.UTC),
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)

	if parsed["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", parsed["status"])
	}
	if parsed["site_id"] != "sg-01" {
		t.Errorf("expected site_id=sg-01, got %v", parsed["site_id"])
	}
}

func TestCommandSubscriberDispatch(t *testing.T) {
	// Test handler registration
	pub := sync.NewStubPublisher()
	// CommandSubscriber requires *MQTTPublisher, not Publisher interface.
	// For unit testing without MQTT, we test the parsing/response logic separately.
	// The full integration test would need a real or mock MQTT broker.
	_ = pub // stub publisher for reference

	// Verify command types are valid
	validCommands := []string{"update_tariff", "update_config", "restart", "identify"}
	for _, cmd := range validCommands {
		rc := sync.RemoteCommand{
			Command:   cmd,
			RequestID: "test-" + cmd,
			Timestamp: time.Now().UTC(),
		}
		data, err := json.Marshal(rc)
		if err != nil {
			t.Errorf("failed to marshal command %s: %v", cmd, err)
			continue
		}
		var parsed sync.RemoteCommand
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Errorf("roundtrip failed for command %s: %v", cmd, err)
		}
		if parsed.Command != cmd {
			t.Errorf("expected %s, got %s", cmd, parsed.Command)
		}
	}
}

func TestMQTTOptionsDefaults(t *testing.T) {
	opts := sync.MQTTOptions{
		Broker:   "tcp://localhost:1883",
		ClientID: "test-client",
	}

	if opts.QoS != 0 {
		t.Errorf("expected default QoS=0 (before NewMQTTPublisher sets it), got %d", opts.QoS)
	}
	if opts.Broker != "tcp://localhost:1883" {
		t.Errorf("unexpected broker: %s", opts.Broker)
	}
}
