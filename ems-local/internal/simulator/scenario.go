package simulator

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Scenario is a deterministic simulator script. Events are applied in offset order.
type Scenario struct {
	Name       string          `yaml:"name"`
	Seed       int64           `yaml:"seed"`
	Duration   time.Duration   `yaml:"-"`
	Step       time.Duration   `yaml:"-"`
	DurationS  int             `yaml:"duration_seconds"`
	StepMS     int             `yaml:"step_ms"`
	InitialSOC float64         `yaml:"initial_soc"`
	SolarPeakW int             `yaml:"solar_peak_w"`
	LoadBaseW  int             `yaml:"load_base_w"`
	Events     []ScenarioEvent `yaml:"events"`
}

type ScenarioEvent struct {
	AtSecond int    `yaml:"at_second"`
	Type     string `yaml:"type"`
	DeviceID string `yaml:"device_id,omitempty"`
	Value    int    `yaml:"value,omitempty"`
}

func LoadScenario(path string) (Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Scenario{}, fmt.Errorf("scenario: read: %w", err)
	}
	var s Scenario
	if err := yaml.Unmarshal(data, &s); err != nil {
		return Scenario{}, fmt.Errorf("scenario: parse: %w", err)
	}
	if s.Name == "" || s.DurationS <= 0 || s.StepMS <= 0 || s.InitialSOC < 0 || s.InitialSOC > 100 {
		return Scenario{}, fmt.Errorf("scenario: invalid required values")
	}
	last := -1
	for _, event := range s.Events {
		if event.AtSecond < last || event.AtSecond > s.DurationS {
			return Scenario{}, fmt.Errorf("scenario: events must be ordered and within duration")
		}
		last = event.AtSecond
	}
	s.Duration = time.Duration(s.DurationS) * time.Second
	s.Step = time.Duration(s.StepMS) * time.Millisecond
	return s, nil
}
