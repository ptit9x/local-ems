package balancing

import (
	"testing"

	"github.com/vmo/local-ems/internal/engine/controllers"
)

func TestController_DesiredPower(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		solarW  int
		loadW   int
		expected int
	}{
		{
			name:     "solar 5000W, load 3000W - charge 2000W",
			config:   DefaultConfig(),
			solarW:   5000,
			loadW:    3000,
			expected: -2000,
		},
		{
			name:     "solar 2000W, load 5000W - discharge 3000W",
			config:   DefaultConfig(),
			solarW:   2000,
			loadW:    5000,
			expected: 3000,
		},
		{
			name:     "solar 3000W, load 3000W - balanced",
			config:   DefaultConfig(),
			solarW:   3000,
			loadW:    3000,
			expected: 0,
		},
		{
			name:     "no solar, load 5000W - full discharge",
			config:   DefaultConfig(),
			solarW:   0,
			loadW:    5000,
			expected: 5000,
		},
		{
			name:     "solar 8000W, no load - full charge",
			config:   DefaultConfig(),
			solarW:   8000,
			loadW:    0,
			expected: -8000,
		},
		{
			name:     "disabled - no action",
			config:   Config{Enabled: false},
			solarW:   5000,
			loadW:    3000,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := New(tt.config)
			state := controllers.DeviceState{
				SolarPowerW: tt.solarW,
				LoadPowerW:  tt.loadW,
			}

			got := ctrl.DesiredPower(state)
			if got != tt.expected {
				t.Errorf("DesiredPower() = %d, want %d", got, tt.expected)
			}
		})
	}
}
