package sell_to_grid_limit

import (
	"testing"

	"github.com/vmo/local-ems/internal/engine/controllers"
	"github.com/vmo/local-ems/internal/engine/resolver"
)

func TestController_Run(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		state         controllers.DeviceState
		expectedCount int
		checkFunc     func(t *testing.T, constraints []resolver.Constraint)
	}{
		{
			name:   "grid buying 1000W - no constraint",
			config: DefaultConfig(),
			state: controllers.DeviceState{
				GridPowerW:      1000,
				ESSActivePowerW: 0,
			},
			expectedCount: 0,
		},
		{
			name:   "grid selling 500W, max=0 (zero-export) - constraint to absorb",
			config: DefaultConfig(),
			state: controllers.DeviceState{
				GridPowerW:      -500,
				ESSActivePowerW: 0,
			},
			expectedCount: 1,
			checkFunc: func(t *testing.T, constraints []resolver.Constraint) {
				c := constraints[0]
				if c.Source != "sell_to_grid_limit" {
					t.Errorf("expected source sell_to_grid_limit, got %s", c.Source)
				}
				// ESS at 0, need to absorb 500W → new limit = 0 - 500 = -500 → must charge
				if c.MaxDischargePower == nil || *c.MaxDischargePower != 0 {
					t.Errorf("expected MaxDischargePower=0, got %v", c.MaxDischargePower)
				}
				if c.ForcePower == nil || *c.ForcePower != -500 {
					t.Errorf("expected ForcePower=-500, got %v", c.ForcePower)
				}
			},
		},
		{
			name:   "grid selling 500W, max=500 - no constraint (within limit)",
			config: Config{MaxSellToGridPower: 500, Enabled: true},
			state: controllers.DeviceState{
				GridPowerW:      -500,
				ESSActivePowerW: 0,
			},
			expectedCount: 0,
		},
		{
			name:   "grid selling 1000W, max=200 - absorb 800W excess",
			config: Config{MaxSellToGridPower: 200, Enabled: true},
			state: controllers.DeviceState{
				GridPowerW:      -1000,
				ESSActivePowerW: 500, // currently discharging 500W
			},
			expectedCount: 1,
			checkFunc: func(t *testing.T, constraints []resolver.Constraint) {
				c := constraints[0]
				// sell=1000, max=200, excess=800, newLimit=500-800=-300 → charge
				if c.MaxDischargePower == nil || *c.MaxDischargePower != 0 {
					t.Errorf("expected MaxDischargePower=0, got %v", c.MaxDischargePower)
				}
				if c.ForcePower == nil || *c.ForcePower != -300 {
					t.Errorf("expected ForcePower=-300, got %v", c.ForcePower)
				}
			},
		},
		{
			name:   "grid = 0W - no constraint",
			config: DefaultConfig(),
			state: controllers.DeviceState{
				GridPowerW:      0,
				ESSActivePowerW: 0,
			},
			expectedCount: 0,
		},
		{
			name:   "disabled - no constraint",
			config: Config{MaxSellToGridPower: 0, Enabled: false},
			state: controllers.DeviceState{
				GridPowerW:      -1000,
				ESSActivePowerW: 0,
			},
			expectedCount: 0,
		},
		{
			name:   "grid selling but ESS discharging enough to just reduce",
			config: DefaultConfig(),
			state: controllers.DeviceState{
				GridPowerW:      -300,
				ESSActivePowerW: 2000, // discharging 2000W
			},
			expectedCount: 1,
			checkFunc: func(t *testing.T, constraints []resolver.Constraint) {
				c := constraints[0]
				// sell=300, max=0, excess=300, newLimit=2000-300=1700 → reduce discharge
				if c.MaxDischargePower == nil || *c.MaxDischargePower != 1700 {
					t.Errorf("expected MaxDischargePower=1700, got %v", c.MaxDischargePower)
				}
				if c.ForcePower != nil {
					t.Errorf("expected no ForcePower, got %v", c.ForcePower)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := New(tt.config)
			collector := resolver.NewConstraintCollector()

			err := ctrl.Run(tt.state, collector)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			constraints := collector.Constraints()
			if len(constraints) != tt.expectedCount {
				t.Fatalf("expected %d constraints, got %d", tt.expectedCount, len(constraints))
			}

			if tt.checkFunc != nil && len(constraints) > 0 {
				tt.checkFunc(t, constraints)
			}
		})
	}
}
