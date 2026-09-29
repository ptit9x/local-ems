package scheduler

import (
	"errors"
	"testing"

	"github.com/vmo/local-ems/internal/engine/controllers"
	"github.com/vmo/local-ems/internal/engine/resolver"
)

// mockController is a test helper.
type mockController struct {
	id      string
	enabled bool
	runErr  error
	ran     bool
}

func (m *mockController) ID() string      { return m.id }
func (m *mockController) IsEnabled() bool  { return m.enabled }
func (m *mockController) Run(state controllers.DeviceState, c *resolver.ConstraintCollector) error {
	m.ran = true
	return m.runErr
}

func TestScheduler_RunAll(t *testing.T) {
	t.Run("all controllers run in order", func(t *testing.T) {
		a := &mockController{id: "a", enabled: true}
		b := &mockController{id: "b", enabled: true}
		c := &mockController{id: "c", enabled: true}

		sched := New(a, b, c)
		collector := resolver.NewConstraintCollector()
		errs := sched.RunAll(controllers.DeviceState{}, collector)

		if len(errs) != 0 {
			t.Errorf("expected no errors, got %v", errs)
		}
		if !a.ran || !b.ran || !c.ran {
			t.Error("expected all controllers to have run")
		}
	})

	t.Run("disabled controllers are skipped", func(t *testing.T) {
		a := &mockController{id: "a", enabled: true}
		b := &mockController{id: "b", enabled: false}
		c := &mockController{id: "c", enabled: true}

		sched := New(a, b, c)
		collector := resolver.NewConstraintCollector()
		sched.RunAll(controllers.DeviceState{}, collector)

		if !a.ran {
			t.Error("expected controller a to have run")
		}
		if b.ran {
			t.Error("expected controller b to be skipped")
		}
		if !c.ran {
			t.Error("expected controller c to have run")
		}
	})

	t.Run("error in one controller doesn't stop others", func(t *testing.T) {
		a := &mockController{id: "a", enabled: true}
		b := &mockController{id: "b", enabled: true, runErr: errors.New("boom")}
		c := &mockController{id: "c", enabled: true}

		sched := New(a, b, c)
		collector := resolver.NewConstraintCollector()
		errs := sched.RunAll(controllers.DeviceState{}, collector)

		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errs))
		}
		if !c.ran {
			t.Error("expected controller c to run despite b's error")
		}
	})

	t.Run("empty scheduler returns no errors", func(t *testing.T) {
		sched := New()
		collector := resolver.NewConstraintCollector()
		errs := sched.RunAll(controllers.DeviceState{}, collector)

		if len(errs) != 0 {
			t.Errorf("expected no errors, got %v", errs)
		}
	})
}
