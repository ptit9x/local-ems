package store_test

import (
	"os"
	"testing"
	"time"

	"github.com/vmo/ems-platform/ems-local/internal/engine/cycle"
	"github.com/vmo/ems-platform/ems-local/internal/store"
)

func TestDiskUsage(t *testing.T) {
	// Create a temporary database file
	tempFile, err := os.CreateTemp("", "ems-test-db-*.sqlite")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	dbPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(dbPath)

	s, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer s.Close()

	// Insert a record so database file has some size
	cr := cycle.CycleResult{
		Timestamp:    time.Now(),
		DesiredPower: 1000,
	}
	cr.State.SolarPowerW = 500
	cr.State.LoadPowerW = 500
	cr.State.GridPowerW = 0
	cr.State.BatterySOC = 50.0
	cr.State.BatteryTempC = 25.0
	cr.ResolvedPower.Setpoint = 0

	if err := s.InsertCycleResult(cr); err != nil {
		t.Fatalf("failed to insert record: %v", err)
	}

	usage, err := s.DiskUsage()
	if err != nil {
		t.Fatalf("DiskUsage returned error: %v", err)
	}

	if usage.DatabaseBytes == 0 {
		t.Errorf("expected non-zero database size, got 0")
	}

	// For physical files, disk stats should be populated
	if usage.DiskTotalBytes == 0 {
		t.Errorf("expected non-zero DiskTotalBytes, got 0")
	}
}

func TestDiskUsageInMemory(t *testing.T) {
	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("failed to open memory store: %v", err)
	}
	defer s.Close()

	usage, err := s.DiskUsage()
	if err != nil {
		t.Fatalf("DiskUsage returned error: %v", err)
	}

	if usage.DatabaseBytes != 0 {
		t.Errorf("expected 0 database size for memory store, got %d", usage.DatabaseBytes)
	}
}
