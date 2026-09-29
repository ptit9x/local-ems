package resolver

import "testing"

func intPtr(v int) *int { return &v }

func TestResolve(t *testing.T) {
	tests := []struct {
		name             string
		constraints      []Constraint
		requestedPower   int
		expectedSetpoint int
		expectedClamped  bool
	}{
		{
			name:             "no constraints - pass through",
			constraints:      nil,
			requestedPower:   3000,
			expectedSetpoint: 3000,
			expectedClamped:  false,
		},
		{
			name: "single MaxDischargePower clamps discharge",
			constraints: []Constraint{
				{MaxDischargePower: intPtr(2000), Source: "test"},
			},
			requestedPower:   5000,
			expectedSetpoint: 2000,
			expectedClamped:  true,
		},
		{
			name: "multiple MaxDischargePower - strictest wins",
			constraints: []Constraint{
				{MaxDischargePower: intPtr(5000), Source: "a"},
				{MaxDischargePower: intPtr(2000), Source: "b"},
				{MaxDischargePower: intPtr(3000), Source: "c"},
			},
			requestedPower:   4000,
			expectedSetpoint: 2000,
			expectedClamped:  true,
		},
		{
			name: "MaxChargePower clamps charge",
			constraints: []Constraint{
				{MaxChargePower: intPtr(1500), Source: "test"},
			},
			requestedPower:   -3000,
			expectedSetpoint: -1500,
			expectedClamped:  true,
		},
		{
			name: "ForcePower overrides requested",
			constraints: []Constraint{
				{ForcePower: intPtr(-2000), Source: "force_charge"},
			},
			requestedPower:   3000,
			expectedSetpoint: -2000,
			expectedClamped:  true,
		},
		{
			name: "ForcePower clamped by MaxDischargePower",
			constraints: []Constraint{
				{MaxDischargePower: intPtr(0), ForcePower: intPtr(-2000), Source: "limit"},
			},
			requestedPower:   3000,
			expectedSetpoint: -2000, // force charge is negative, MaxDischarge=0 only clamps positive
			expectedClamped:  true,
		},
		{
			name: "conflict: two different MaxDischargePower - min wins",
			constraints: []Constraint{
				{MaxDischargePower: intPtr(5000), Source: "a"},
				{MaxDischargePower: intPtr(2000), Source: "b"},
			},
			requestedPower:   3000,
			expectedSetpoint: 2000,
			expectedClamped:  true,
		},
		{
			name: "SOC protection: MaxDischarge=0 blocks peak_shaving discharge",
			constraints: []Constraint{
				{MaxDischargePower: intPtr(0), Source: "limit_discharge"},
			},
			requestedPower:   3000,
			expectedSetpoint: 0,
			expectedClamped:  true,
		},
		{
			name: "requested within bounds - no clamping",
			constraints: []Constraint{
				{MaxDischargePower: intPtr(5000), Source: "a"},
				{MaxChargePower: intPtr(3000), Source: "b"},
			},
			requestedPower:   2000,
			expectedSetpoint: 2000,
			expectedClamped:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewConstraintCollector()
			for _, c := range tt.constraints {
				collector.Add(c)
			}

			r := NewResolver(collector)
			result := r.Resolve(tt.requestedPower)

			if result.Setpoint != tt.expectedSetpoint {
				t.Errorf("setpoint: got %d, want %d", result.Setpoint, tt.expectedSetpoint)
			}
			if result.Clamped != tt.expectedClamped {
				t.Errorf("clamped: got %v, want %v", result.Clamped, tt.expectedClamped)
			}
		})
	}
}

func TestConstraintCollector_Reset(t *testing.T) {
	c := NewConstraintCollector()
	c.Add(Constraint{MaxDischargePower: intPtr(100), Source: "a"})
	c.Add(Constraint{MaxDischargePower: intPtr(200), Source: "b"})

	if len(c.Constraints()) != 2 {
		t.Fatalf("expected 2 constraints, got %d", len(c.Constraints()))
	}

	c.Reset()
	if len(c.Constraints()) != 0 {
		t.Fatalf("expected 0 constraints after reset, got %d", len(c.Constraints()))
	}
}
