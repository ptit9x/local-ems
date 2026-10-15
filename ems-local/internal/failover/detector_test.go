package failover

import (
	"encoding/json"
	"testing"

	emsSync "github.com/vmo/ems-platform/ems-local/internal/sync"
)

func TestFailoverManager(t *testing.T) {
	pub := emsSync.NewStubPublisher()
	pub.Connect()
	fm := NewFailoverManager(pub, "test-site", "ems")

	if fm.State() != FailoverNormal {
		t.Errorf("expected state normal, got %v", fm.State())
	}

	stateChanges := []FailoverState{}
	fm.OnStateChange(func(s FailoverState) {
		stateChanges = append(stateChanges, s)
	})

	// Normal -> Active
	fm.HandleCloudMessage("takeover", []byte(`{}`))
	if fm.State() != FailoverActive {
		t.Errorf("expected state active, got %v", fm.State())
	}

	// Active -> Recovery
	err := fm.RequestRecovery()
	if err != nil {
		t.Fatalf("failed to request recovery: %v", err)
	}
	if fm.State() != FailoverRecovery {
		t.Errorf("expected state recovery, got %v", fm.State())
	}

	messages := pub.GetMessages()
	foundRecovery := false
	for _, msg := range messages {
		if msg.Topic == "ems/test-site/failover/recovery" {
			foundRecovery = true
			var req recoveryRequest
			if err := json.Unmarshal(msg.Payload, &req); err != nil {
				t.Fatalf("failed to unmarshal request: %v", err)
			}
			if req.Action != "request_recovery" {
				t.Errorf("expected action request_recovery, got %s", req.Action)
			}
		}
	}
	if !foundRecovery {
		t.Errorf("expected recovery message to be published")
	}

	// Recovery -> Normal
	fm.HandleCloudMessage("handover_complete", []byte(`{}`))
	if fm.State() != FailoverNormal {
		t.Errorf("expected state normal, got %v", fm.State())
	}

	if len(stateChanges) != 3 {
		t.Errorf("expected 3 state changes, got %d", len(stateChanges))
	} else {
		if stateChanges[0] != FailoverActive {
			t.Errorf("expected first change to active")
		}
		if stateChanges[1] != FailoverRecovery {
			t.Errorf("expected second change to recovery")
		}
		if stateChanges[2] != FailoverNormal {
			t.Errorf("expected third change to normal")
		}
	}
}
