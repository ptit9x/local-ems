package distributor

import (
	"testing"

	"github.com/vmo/local-ems/internal/devices/aggregator"
)

func makePCS(n int, maxW int) []aggregator.PCSState {
	pcs := make([]aggregator.PCSState, n)
	for i := range pcs {
		pcs[i] = aggregator.PCSState{
			ID:       "pcs-" + string(rune('0'+i)),
			MaxPowerW: maxW,
			Status:   aggregator.PCSRunning,
			Online:   true,
		}
	}
	return pcs
}

func TestEqualSplit_Basic(t *testing.T) {
	d := New(EqualSplit, 2)
	pcs := makePCS(2, 5000)

	result := d.Distribute(4000, pcs)

	// 4000 / 2 = 2000 each
	if result.Setpoints[0] != 2000 {
		t.Errorf("expected pcs-0 = 2000, got %d", result.Setpoints[0])
	}
	if result.Setpoints[1] != 2000 {
		t.Errorf("expected pcs-1 = 2000, got %d", result.Setpoints[1])
	}
	if result.TotalW != 4000 {
		t.Errorf("expected total 4000, got %d", result.TotalW)
	}
}

func TestEqualSplit_OddNumber(t *testing.T) {
	d := New(EqualSplit, 2)
	pcs := makePCS(2, 5000)

	result := d.Distribute(4001, pcs)

	// 4001 / 2 = 2000 each, remainder 1 goes to first
	if result.Setpoints[0] != 2001 {
		t.Errorf("expected pcs-0 = 2001, got %d", result.Setpoints[0])
	}
	if result.Setpoints[1] != 2000 {
		t.Errorf("expected pcs-1 = 2000, got %d", result.Setpoints[1])
	}
}

func TestEqualSplit_Clamped(t *testing.T) {
	d := New(EqualSplit, 2)
	pcs := makePCS(2, 1000) // max 1000W each

	result := d.Distribute(4000, pcs)

	// 4000 / 2 = 2000, but clamped to 1000 each
	if result.Setpoints[0] != 1000 {
		t.Errorf("expected pcs-0 = 1000 (clamped), got %d", result.Setpoints[0])
	}
	if result.TotalW != 2000 {
		t.Errorf("expected total 2000 (clamped), got %d", result.TotalW)
	}
	if !result.Clamped {
		t.Error("expected Clamped=true")
	}
}

func TestEqualSplit_Charging(t *testing.T) {
	d := New(EqualSplit, 2)
	pcs := makePCS(2, 5000)

	result := d.Distribute(-3000, pcs)

	// -3000 / 2 = -1500 each
	if result.Setpoints[0] != -1500 {
		t.Errorf("expected pcs-0 = -1500, got %d", result.Setpoints[0])
	}
	if result.TotalW != -3000 {
		t.Errorf("expected total -3000, got %d", result.TotalW)
	}
}

func TestEqualSplit_OneFaulted(t *testing.T) {
	d := New(EqualSplit, 2)
	pcs := makePCS(2, 5000)
	pcs[1].Status = aggregator.PCSFault // pcs-1 faulted

	result := d.Distribute(4000, pcs)

	// All 4000 goes to pcs-0
	if result.Setpoints[0] != 4000 {
		t.Errorf("expected pcs-0 = 4000, got %d", result.Setpoints[0])
	}
	if result.Setpoints[1] != 0 {
		t.Errorf("expected pcs-1 = 0 (faulted), got %d", result.Setpoints[1])
	}
}

func TestEqualSplit_AllOffline(t *testing.T) {
	d := New(EqualSplit, 2)
	pcs := makePCS(2, 5000)
	pcs[0].Online = false
	pcs[1].Online = false

	result := d.Distribute(4000, pcs)

	if result.TotalW != 0 {
		t.Errorf("expected total 0 (all offline), got %d", result.TotalW)
	}
}

func TestRoundRobin(t *testing.T) {
	d := New(RoundRobin, 2)
	pcs := makePCS(2, 3000) // max 3000 each

	// First call: starts at pcs-0
	r1 := d.Distribute(5000, pcs)
	// pcs-0 gets 3000 (clamped), pcs-1 gets 2000
	if r1.Setpoints[0] != 3000 {
		t.Errorf("round 1: expected pcs-0 = 3000, got %d", r1.Setpoints[0])
	}
	if r1.Setpoints[1] != 2000 {
		t.Errorf("round 1: expected pcs-1 = 2000, got %d", r1.Setpoints[1])
	}

	// Second call: starts at pcs-1 (rotated)
	r2 := d.Distribute(5000, pcs)
	if r2.Setpoints[1] != 3000 {
		t.Errorf("round 2: expected pcs-1 = 3000 (lead), got %d", r2.Setpoints[1])
	}
	if r2.Setpoints[0] != 2000 {
		t.Errorf("round 2: expected pcs-0 = 2000, got %d", r2.Setpoints[0])
	}
}

func TestMasterMode_SingleGroup(t *testing.T) {
	d := New(EqualSplit, 2)
	pcs := makePCS(2, 5000)

	d.SetGroups([]BESSGroup{
		{Name: "Group A", PCSIndices: []int{0, 1}, CommandMode: "master", MasterIndex: 0},
	})

	result := d.Distribute(4000, pcs)

	// Master mode: only pcs-0 (master) gets the full setpoint
	if result.Setpoints[0] != 4000 {
		t.Errorf("expected master pcs-0 = 4000, got %d", result.Setpoints[0])
	}
	if result.Setpoints[1] != 0 {
		t.Errorf("expected pcs-1 = 0 (not master), got %d", result.Setpoints[1])
	}
}

func TestMasterMode_MasterOffline(t *testing.T) {
	d := New(EqualSplit, 2)
	pcs := makePCS(2, 5000)
	pcs[0].Online = false // master is offline

	d.SetGroups([]BESSGroup{
		{Name: "Group A", PCSIndices: []int{0, 1}, CommandMode: "master", MasterIndex: 0},
	})

	result := d.Distribute(4000, pcs)

	// Master offline → fallback to first active PCS (pcs-1)
	if result.Setpoints[0] != 0 {
		t.Errorf("expected pcs-0 = 0 (offline), got %d", result.Setpoints[0])
	}
	if result.Setpoints[1] != 4000 {
		t.Errorf("expected pcs-1 = 4000 (fallback master), got %d", result.Setpoints[1])
	}
}

func TestMixedGroups(t *testing.T) {
	d := New(EqualSplit, 4)
	pcs := makePCS(4, 5000)

	d.SetGroups([]BESSGroup{
		{Name: "Group A", PCSIndices: []int{0, 1}, CommandMode: "master", MasterIndex: 0},
		{Name: "Group B", PCSIndices: []int{2, 3}, CommandMode: "individual"},
	})

	result := d.Distribute(8000, pcs)

	// 4 active PCS total. Group A has 2 active, Group B has 2 active.
	// Group A gets 8000*2/4 = 4000 → master pcs-0 gets 4000
	// Group B gets 8000*2/4 = 4000 → split: pcs-2=2000, pcs-3=2000
	if result.Setpoints[0] != 4000 {
		t.Errorf("expected Group A master pcs-0 = 4000, got %d", result.Setpoints[0])
	}
	if result.Setpoints[1] != 0 {
		t.Errorf("expected Group A pcs-1 = 0, got %d", result.Setpoints[1])
	}
	if result.Setpoints[2] != 2000 {
		t.Errorf("expected Group B pcs-2 = 2000, got %d", result.Setpoints[2])
	}
	if result.Setpoints[3] != 2000 {
		t.Errorf("expected Group B pcs-3 = 2000, got %d", result.Setpoints[3])
	}
	if result.TotalW != 8000 {
		t.Errorf("expected total 8000, got %d", result.TotalW)
	}
}

func TestNoGroups_Fallback(t *testing.T) {
	d := New(EqualSplit, 2)
	pcs := makePCS(2, 5000)

	// No groups configured → same as current EqualSplit
	result := d.Distribute(4000, pcs)

	if result.Setpoints[0] != 2000 {
		t.Errorf("expected pcs-0 = 2000, got %d", result.Setpoints[0])
	}
	if result.Setpoints[1] != 2000 {
		t.Errorf("expected pcs-1 = 2000, got %d", result.Setpoints[1])
	}
}

