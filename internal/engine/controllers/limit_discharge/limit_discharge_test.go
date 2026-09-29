package limit_discharge

import (
	"testing"

	"github.com/vmo/local-ems/internal/engine/controllers"
	"github.com/vmo/local-ems/internal/engine/resolver"
)

func TestController_Run(t *testing.T) {
	defaultCfg := DefaultConfig()

	tests := []struct {
		name          string
		config        Config
		soc           float64
		expectedCount int
		checkFunc     func(t *testing.T, constraints []resolver.Constraint)
	}{
		{
			name:          "SOC = 80% -> no constraints added",
			config:        defaultCfg,
			soc:           80.0,
			expectedCount: 0,
		},
		{
			name:          "SOC = 14% (below minSOC 15%) -> MaxDischargePower = 0",
			config:        defaultCfg,
			soc:           14.0,
			expectedCount: 1,
			checkFunc: func(t *testing.T, constraints []resolver.Constraint) {
				c := constraints[0]
				if c.Source != "limit_discharge" {
					t.Errorf("expected source limit_discharge, got %s", c.Source)
				}
				if c.MaxDischargePower == nil || *c.MaxDischargePower != 0 {
					t.Errorf("expected MaxDischargePower to be 0, got %v", c.MaxDischargePower)
				}
				if c.ForcePower != nil {
					t.Errorf("expected ForcePower to be nil, got %v", c.ForcePower)
				}
			},
		},
		{
			name:          "SOC = 9% (below forceChargeSOC 10%) -> ForcePower = -2000 AND MaxDischargePower = 0",
			config:        defaultCfg,
			soc:           9.0,
			expectedCount: 1,
			checkFunc: func(t *testing.T, constraints []resolver.Constraint) {
				c := constraints[0]
				if c.Source != "limit_discharge" {
					t.Errorf("expected source limit_discharge, got %s", c.Source)
				}
				if c.MaxDischargePower == nil || *c.MaxDischargePower != 0 {
					t.Errorf("expected MaxDischargePower to be 0, got %v", c.MaxDischargePower)
				}
				if c.ForcePower == nil || *c.ForcePower != -2000 {
					t.Errorf("expected ForcePower to be -2000, got %v", c.ForcePower)
				}
			},
		},
		{
			name:          "SOC = 15% exactly (boundary) -> no constraints",
			config:        defaultCfg,
			soc:           15.0,
			expectedCount: 0,
		},
		{
			name:          "SOC = 10% exactly (boundary) -> MaxDischargePower = 0 only",
			config:        defaultCfg,
			soc:           10.0,
			expectedCount: 1,
			checkFunc: func(t *testing.T, constraints []resolver.Constraint) {
				c := constraints[0]
				if c.MaxDischargePower == nil || *c.MaxDischargePower != 0 {
					t.Errorf("expected MaxDischargePower to be 0, got %v", c.MaxDischargePower)
				}
				if c.ForcePower != nil {
					t.Errorf("expected ForcePower to be nil, got %v", c.ForcePower)
				}
			},
		},
		{
			name: "Disabled controller -> no constraints regardless of SOC",
			config: Config{
				MinSOC:         15.0,
				ForceChargeSOC: 10.0,
				ForceChargeW:   2000,
				Enabled:        false,
			},
			soc:           5.0,
			expectedCount: 0,
		},
		{
			name: "Custom config values work correctly",
			config: Config{
				MinSOC:         25.0,
				ForceChargeSOC: 20.0,
				ForceChargeW:   3000,
				Enabled:        true,
			},
			soc:           19.0,
			expectedCount: 1,
			checkFunc: func(t *testing.T, constraints []resolver.Constraint) {
				c := constraints[0]
				if c.MaxDischargePower == nil || *c.MaxDischargePower != 0 {
					t.Errorf("expected MaxDischargePower to be 0, got %v", c.MaxDischargePower)
				}
				if c.ForcePower == nil || *c.ForcePower != -3000 {
					t.Errorf("expected ForcePower to be -3000, got %v", c.ForcePower)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := New(tt.config)
			state := controllers.DeviceState{BatterySOC: tt.soc}
			collector := resolver.NewConstraintCollector()

			err := ctrl.Run(state, collector)
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
