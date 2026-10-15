package sync_test

import (
	"context"
	"testing"
	"time"

	"github.com/vmo/ems-platform/ems-local/internal/engine/controllers"
	"github.com/vmo/ems-platform/ems-local/internal/engine/cycle"
	"github.com/vmo/ems-platform/ems-local/internal/engine/resolver"
	"github.com/vmo/ems-platform/ems-local/internal/store"
	"github.com/vmo/ems-platform/ems-local/internal/sync"
)

func TestSyncer(t *testing.T) {
	st, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer st.Close()

	pub := sync.NewStubPublisher()
	syncer := sync.NewSyncer(st, pub, "site-1", "ems", 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	syncer.Start(ctx)

	// Add cycles
	syncer.AddCycle(cycle.CycleResult{
		Timestamp: time.Now(),
		State: controllers.DeviceState{
			SolarPowerW:  1000,
			LoadPowerW:   500,
			BatterySOC:   50.0,
		},
		ResolvedPower: resolver.ResolvedPower{
			Setpoint: 500,
		},
	})

	// Wait for tick
	time.Sleep(100 * time.Millisecond)

	syncer.Stop()

	// Check publisher messages
	msgs := pub.GetMessages()
	if len(msgs) == 0 {
		t.Fatalf("expected messages published, got 0")
	}

	if msgs[0].Topic != "ems/site-1/telemetry" {
		t.Errorf("unexpected topic: %s", msgs[0].Topic)
	}
}

func TestSyncerReplay(t *testing.T) {
	st, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer st.Close()

	// Insert pending records manually
	for i := 0; i < 5; i++ {
		st.InsertOutbox(`{"test":true}`)
	}

	pub := sync.NewStubPublisher()
	syncer := sync.NewSyncer(st, pub, "site-1", "ems", 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	syncer.Start(ctx)
	time.Sleep(100 * time.Millisecond)
	syncer.Stop()

	msgs := pub.GetMessages()
	if len(msgs) != 5 {
		t.Errorf("expected 5 replayed messages, got %d", len(msgs))
	}
}
