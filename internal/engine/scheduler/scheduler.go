package scheduler

import (
	"fmt"

	"github.com/vmo/local-ems/internal/engine/controllers"
	"github.com/vmo/local-ems/internal/engine/resolver"
)

// Scheduler runs controllers in fixed priority order.
type Scheduler struct {
	controllers []controllers.Controller
}

// New creates a new Scheduler with the provided controllers in priority order.
func New(ctrls ...controllers.Controller) *Scheduler {
	return &Scheduler{
		controllers: ctrls,
	}
}

// RunAll executes all enabled controllers in order, adding constraints to collector.
// Errors from individual controllers are collected but do not stop execution.
func (s *Scheduler) RunAll(state controllers.DeviceState, collector *resolver.ConstraintCollector) []error {
	var errs []error
	for _, ctrl := range s.controllers {
		if !ctrl.IsEnabled() {
			continue
		}
		if err := ctrl.Run(state, collector); err != nil {
			errs = append(errs, fmt.Errorf("controller %s: %w", ctrl.ID(), err))
		}
	}
	return errs
}

// Controllers returns the ordered list of controllers.
func (s *Scheduler) Controllers() []controllers.Controller {
	return s.controllers
}
