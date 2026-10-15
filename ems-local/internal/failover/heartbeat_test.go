package failover

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	emsSync "github.com/vmo/ems-platform/ems-local/internal/sync"
)

func TestHeartbeatPublisher(t *testing.T) {
	pub := emsSync.NewStubPublisher()
	pub.Connect()
	siteID := "test-site"
	topicPrefix := "ems"
	interval := 50 * time.Millisecond

	hb := NewHeartbeatPublisher(pub, siteID, topicPrefix, interval)
	
	evStates := []EVChargerInfo{
		{
			ID:               "ev-1",
			Status:           1,
			ActivePowerW:     5000,
			VehicleConnected: true,
			Protocol:         "modbus",
		},
	}
	hb.SetEVProvider(func() []EVChargerInfo {
		return evStates
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	hb.Start(ctx)
	
	time.Sleep(120 * time.Millisecond)
	hb.Stop()

	messages := pub.GetMessages()
	if len(messages) == 0 {
		t.Fatalf("expected at least one message to be published, got 0")
	}

	foundTopic := false
	for _, msg := range messages {
		if msg.Topic == "ems/test-site/heartbeat" {
			foundTopic = true
			
			var hMsg HeartbeatMessage
			err := json.Unmarshal(msg.Payload, &hMsg)
			if err != nil {
				t.Fatalf("failed to unmarshal payload: %v", err)
			}
			
			if hMsg.SiteID != "test-site" {
				t.Errorf("expected site_id test-site, got %s", hMsg.SiteID)
			}
			if hMsg.Status != "online" {
				t.Errorf("expected status online, got %s", hMsg.Status)
			}
			if len(hMsg.EVStates) != 1 {
				t.Errorf("expected 1 EV state, got %d", len(hMsg.EVStates))
			} else if hMsg.EVStates[0].ID != "ev-1" {
				t.Errorf("expected ev-1, got %s", hMsg.EVStates[0].ID)
			}
		}
	}

	if !foundTopic {
		t.Errorf("did not find heartbeat message in published messages")
	}
}
