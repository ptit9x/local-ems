package simulator

import "testing"

func TestLoadFullDemoScenario(t *testing.T) {
	s, err := LoadScenario("../../config/scenarios/full-demo.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if s.Seed != 42 || len(s.Events) < 8 {
		t.Fatalf("unexpected scenario: %+v", s)
	}
}
