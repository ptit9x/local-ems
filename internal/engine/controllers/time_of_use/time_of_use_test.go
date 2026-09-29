package time_of_use

import (
	"testing"
	"time"

	"github.com/vmo/local-ems/internal/engine/controllers"
)

func fixedTime(hour int) func() time.Time {
	return func() time.Time {
		return time.Date(2025, 1, 15, hour, 30, 0, 0, time.UTC)
	}
}

func TestController_DesiredPower(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		hour     int
		soc      float64
		expected int
	}{
		{
			name:     "23:00 (OffPeak) - charge",
			config:   DefaultConfig(),
			hour:     23,
			soc:      50.0,
			expected: -3000,
		},
		{
			name:     "02:00 (OffPeak, after midnight) - charge",
			config:   DefaultConfig(),
			hour:     2,
			soc:      50.0,
			expected: -3000,
		},
		{
			name:     "10:00 (OnPeak), SOC=80% - discharge",
			config:   DefaultConfig(),
			hour:     10,
			soc:      80.0,
			expected: 5000,
		},
		{
			name:     "10:00 (OnPeak), SOC=15% (below min) - no action",
			config:   DefaultConfig(),
			hour:     10,
			soc:      15.0,
			expected: 0,
		},
		{
			name:     "14:00 (MidPeak) - no action",
			config:   DefaultConfig(),
			hour:     14,
			soc:      80.0,
			expected: 0,
		},
		{
			name:     "06:00 boundary (end of OffPeak) - MidPeak",
			config:   DefaultConfig(),
			hour:     6,
			soc:      50.0,
			expected: 0,
		},
		{
			name:     "22:00 boundary (start of OffPeak) - charge",
			config:   DefaultConfig(),
			hour:     22,
			soc:      50.0,
			expected: -3000,
		},
		{
			name:   "disabled - no action",
			config: Config{Enabled: false},
			hour:   23,
			soc:    50.0,
			expected: 0,
		},
		{
			name:     "18:00 (OnPeak evening), SOC=25% - discharge (above min 20%)",
			config:   DefaultConfig(),
			hour:     18,
			soc:      25.0,
			expected: 5000,
		},
		{
			name:     "18:00 (OnPeak evening), SOC=20% exactly - no discharge",
			config:   DefaultConfig(),
			hour:     18,
			soc:      20.0,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := New(tt.config)
			ctrl.NowFunc = fixedTime(tt.hour)
			state := controllers.DeviceState{BatterySOC: tt.soc}

			got := ctrl.DesiredPower(state)
			if got != tt.expected {
				t.Errorf("DesiredPower() = %d, want %d", got, tt.expected)
			}
		})
	}
}
