package peak_shaving

import (
	"testing"

	"github.com/vmo/local-ems/internal/engine/controllers"
)

func TestController_DesiredPower(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		gridW    int
		expected int
	}{
		{
			name:     "grid 30kW, threshold 50kW - no peak",
			config:   DefaultConfig(),
			gridW:    30000,
			expected: 0,
		},
		{
			name:     "grid 70kW, threshold 50kW - discharge 20kW",
			config:   DefaultConfig(),
			gridW:    70000,
			expected: 20000,
		},
		{
			name:     "grid 50kW exactly - no peak",
			config:   DefaultConfig(),
			gridW:    50000,
			expected: 0,
		},
		{
			name:     "disabled - no peak regardless",
			config:   Config{PeakThresholdW: 50000, Enabled: false},
			gridW:    80000,
			expected: 0,
		},
		{
			name:     "custom threshold 20kW, grid 25kW - discharge 5kW",
			config:   Config{PeakThresholdW: 20000, Enabled: true},
			gridW:    25000,
			expected: 5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := New(tt.config)
			state := controllers.DeviceState{GridPowerW: tt.gridW}

			got := ctrl.DesiredPower(state)
			if got != tt.expected {
				t.Errorf("DesiredPower() = %d, want %d", got, tt.expected)
			}
		})
	}
}
