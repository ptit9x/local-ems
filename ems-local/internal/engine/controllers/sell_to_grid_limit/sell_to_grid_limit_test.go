package sell_to_grid_limit

import (
	"testing"

	"github.com/vmo/ems-platform/ems-local/internal/engine/controllers"
	"github.com/vmo/ems-platform/ems-local/internal/engine/resolver"
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
			name:   "grid buying 1000W, max=0 (zero-export) - limits discharge to 1000W",
			config: DefaultConfig(),
			state: controllers.DeviceState{
				GridPowerW:      1000,
				ESSActivePowerW: 0,
			},
			expectedCount: 1,
			checkFunc: func(t *testing.T, constraints []resolver.Constraint) {
				c := constraints[0]
				if c.MaxDischargePower == nil || *c.MaxDischargePower != 1000 {
					t.Errorf("expected MaxDischargePower=1000, got %v", c.MaxDischargePower)
				}
			},
		},
		{
			name:   "grid selling 500W, max=0 (zero-export) - forces charge of 500W via MaxDischargePower",
			config: DefaultConfig(),
			state: controllers.DeviceState{
				GridPowerW:      -500,
				ESSActivePowerW: 0,
			},
			expectedCount: 1,
			checkFunc: func(t *testing.T, constraints []resolver.Constraint) {
				c := constraints[0]
				if c.MaxDischargePower == nil || *c.MaxDischargePower != -500 {
					t.Errorf("expected MaxDischargePower=-500, got %v", c.MaxDischargePower)
				}
			},
		},
		{
			name:   "grid selling 500W, max=500 - limits discharge to 0W",
			config: Config{MaxSellToGridPower: 500, Enabled: true},
			state: controllers.DeviceState{
				GridPowerW:      -500,
				ESSActivePowerW: 0,
			},
			expectedCount: 1,
			checkFunc: func(t *testing.T, constraints []resolver.Constraint) {
				c := constraints[0]
				if c.MaxDischargePower == nil || *c.MaxDischargePower != 0 {
					t.Errorf("expected MaxDischargePower=0, got %v", c.MaxDischargePower)
				}
			},
		},
		{
			name:   "grid selling 1000W, ESS=500W, max=200 - limits discharge to -300W",
			config: Config{MaxSellToGridPower: 200, Enabled: true},
			state: controllers.DeviceState{
				GridPowerW:      -1000,
				ESSActivePowerW: 500, // currently discharging 500W
			},
			expectedCount: 1,
			checkFunc: func(t *testing.T, constraints []resolver.Constraint) {
				c := constraints[0]
				if c.MaxDischargePower == nil || *c.MaxDischargePower != -300 {
					t.Errorf("expected MaxDischargePower=-300, got %v", c.MaxDischargePower)
				}
			},
		},
		{
			name:   "grid = 0W - limits discharge to 0W",
			config: DefaultConfig(),
			state: controllers.DeviceState{
				GridPowerW:      0,
				ESSActivePowerW: 0,
			},
			expectedCount: 1,
			checkFunc: func(t *testing.T, constraints []resolver.Constraint) {
				c := constraints[0]
				if c.MaxDischargePower == nil || *c.MaxDischargePower != 0 {
					t.Errorf("expected MaxDischargePower=0, got %v", c.MaxDischargePower)
				}
			},
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
