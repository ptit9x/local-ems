package cycle

import (
	"context"
	"log"
	"time"

	"github.com/vmo/local-ems/internal/engine/controllers"
	"github.com/vmo/local-ems/internal/engine/resolver"
	"github.com/vmo/local-ems/internal/engine/scheduler"
)

// PowerAdvisor is implemented by controllers that suggest a desired power setpoint.
type PowerAdvisor interface {
	DesiredPower(state controllers.DeviceState) int
}

// CycleResult holds the output of a single cycle execution.
type CycleResult struct {
	Timestamp       time.Time
	State           controllers.DeviceState
	ResolvedPower   resolver.ResolvedPower
	DesiredPower    int
	Errors          []error
	CycleDurationMs int64
}

// Manager orchestrates the read→process→write cycle loop.
type Manager struct {
	scheduler *scheduler.Scheduler
	collector *resolver.ConstraintCollector
	resolver  *resolver.Resolver
	advisors  []PowerAdvisor
	interval  time.Duration
	onCycle   func(result CycleResult)
}

// New creates a new CycleManager.
func New(
	sched *scheduler.Scheduler,
	collector *resolver.ConstraintCollector,
	res *resolver.Resolver,
	advisors []PowerAdvisor,
	interval time.Duration,
) *Manager {
	return &Manager{
		scheduler: sched,
		collector: collector,
		resolver:  res,
		advisors:  advisors,
		interval:  interval,
	}
}

// OnCycle registers a callback that is invoked after each cycle completes.
func (m *Manager) OnCycle(fn func(result CycleResult)) {
	m.onCycle = fn
}

// RunOnce executes a single cycle: reset → run controllers → collect advice → resolve.
func (m *Manager) RunOnce(state controllers.DeviceState) CycleResult {
	start := time.Now()

	// 1. Reset constraints from previous cycle
	m.collector.Reset()

	// 2. Run all controllers (adds constraints)
	errs := m.scheduler.RunAll(state, m.collector)

	// 3. Collect desired power from advisors (first non-zero wins, priority order)
	desiredPower := 0
	for _, advisor := range m.advisors {
		if p := advisor.DesiredPower(state); p != 0 {
			desiredPower = p
			break
		}
	}

	// 4. Resolve final setpoint from desired power + constraints
	resolved := m.resolver.Resolve(desiredPower)

	return CycleResult{
		Timestamp:       start,
		State:           state,
		ResolvedPower:   resolved,
		DesiredPower:    desiredPower,
		Errors:          errs,
		CycleDurationMs: time.Since(start).Milliseconds(),
	}
}

// Start runs the cycle loop until the context is cancelled.
// stateProvider returns the current device state each cycle.
// powerApplier applies the resolved setpoint to the hardware/simulator.
func (m *Manager) Start(
	ctx context.Context,
	stateProvider func() controllers.DeviceState,
	powerApplier func(setpointW int),
) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Cycle manager stopped.")
			return
		case <-ticker.C:
			state := stateProvider()
			result := m.RunOnce(state)
			powerApplier(result.ResolvedPower.Setpoint)

			if m.onCycle != nil {
				m.onCycle(result)
			}
		}
	}
}
