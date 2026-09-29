---
name: ems-controller-generator
description: "Use when creating a new energy controller for the Local EMS project. Also use when asked to add a controller, implement a new energy strategy, or scaffold a controller package."
metadata:
  category: technique
  author: richard
  triggers: new controller, add controller, energy controller, scaffold controller,
    ems controller, constraint controller, battery controller, power controller
risk: low
source: local
---

# EMS Controller Generator

Generates a complete, ready-to-run controller package for the Local EMS project following the **Constraint-based Controller Chain** architecture.

## When to Use

- User asks to create a new energy controller
- User says "add controller", "new controller", "implement X strategy"
- User wants to add a new energy management algorithm
- Adding a new priority to the scheduler

## Architecture Rules (MUST FOLLOW)

> [!CAUTION]
> **NEVER** write directly to hardware. Controllers ONLY read `DeviceState` and add `Constraint`s.
> **NEVER** use if-else to override setpoints. The Resolver decides the final setpoint.
> **NEVER** change priority order of `limit_discharge` (P1) or `sell_to_grid_limit` (P2) — they are safety-critical.

### How Controllers Work

```
Scheduler → runs each Controller in priority order
  → each Controller reads DeviceState (read-only process image)
  → each Controller adds Constraint(s) to ConstraintCollector
  → Resolver aggregates all constraints → final setpoint
```

### Constraint Types (use `*int` — nil means "no limit")

| Field | Meaning | Example |
|-------|---------|---------|
| `MaxDischargePower` | Cap discharge watts | `intPtr(0)` = block discharge |
| `MaxChargePower` | Cap charge watts | `intPtr(2000000)` = max 2MW charge |
| `ForcePower` | Override setpoint | `intPtr(-500000)` = force 500kW charge |
| `Source` | Controller ID | `controllerID` |

### DeviceState Fields (read-only)

| Field | Type | Description |
|-------|------|-------------|
| `GridPowerW` | `int` | Grid meter: positive=buying, negative=selling |
| `SolarPowerW` | `int` | Solar meter: positive=generating |
| `LoadPowerW` | `int` | Load meter: positive=consuming |
| `BatterySOC` | `float64` | Battery SOC 0-100% |
| `BatteryTempC` | `float64` | Battery temperature in Celsius |
| `ESSActivePowerW` | `int` | Current ESS power: positive=discharge, negative=charge |
| `CellVoltageMinMV` | `int` | Min cell voltage in millivolts |
| `CellVoltageMaxMV` | `int` | Max cell voltage in millivolts |

### Priority Order (Scheduler)

| Priority | Controller | Purpose |
|----------|-----------|---------|
| P1 | `limit_discharge` | Hardware protection (safety-critical) |
| P2 | `sell_to_grid_limit` | Grid compliance (safety-critical) |
| P3 | `peak_shaving`, `time_of_use`, `ev_charging` | Cost optimization |
| P4 | `balancing` | Default self-consumption |

New controllers typically go at **P3** (optimization) unless they are safety-critical.

## Generation Steps

### Step 1: Create package directory

```
internal/engine/controllers/<controller_name>/
```

Package name MUST be lowercase, single word or snake_case.

### Step 2: Create controller file `<controller_name>.go`

Use this exact template:

```go
package <controller_name>

import (
	"github.com/vmo/local-ems/internal/engine/controllers"
	"github.com/vmo/local-ems/internal/engine/resolver"
)

const controllerID = "<controller_name>"

// Config holds tunable parameters for the <controller_name> controller.
type Config struct {
	// Add configurable parameters here with defaults in DefaultConfig()
	Enabled bool
}

// DefaultConfig returns the production default configuration.
func DefaultConfig() Config {
	return Config{
		Enabled: true,
	}
}

// Controller implements <describe what this controller does>.
type Controller struct {
	config  Config
	enabled bool
}

// New creates a new <controller_name> controller.
func New(cfg Config) *Controller {
	return &Controller{config: cfg, enabled: cfg.Enabled}
}

func (c *Controller) ID() string      { return controllerID }
func (c *Controller) IsEnabled() bool  { return c.enabled }

// Run reads the current device state and adds constraints to the collector.
func (c *Controller) Run(state controllers.DeviceState, collector *resolver.ConstraintCollector) error {
	// 1. Read state (DO NOT modify state)
	// 2. Evaluate conditions
	// 3. Add constraints via collector.Add()
	//
	// Example:
	//   collector.Add(resolver.Constraint{
	//       MaxDischargePower: intPtr(0),
	//       Source:            controllerID,
	//   })

	return nil
}

// Optionally implement PowerAdvisor if this controller suggests a desired power.
// Uncomment below if needed:
//
// func (c *Controller) DesiredPower(state controllers.DeviceState) int {
//     return 0
// }

func intPtr(v int) *int { return &v }
```

### Step 3: Create test file `<controller_name>_test.go`

Use this exact template:

```go
package <controller_name>

import (
	"testing"

	"github.com/vmo/local-ems/internal/engine/controllers"
	"github.com/vmo/local-ems/internal/engine/resolver"
)

func TestID(t *testing.T) {
	c := New(DefaultConfig())
	if c.ID() != "<controller_name>" {
		t.Fatalf("expected '<controller_name>', got %q", c.ID())
	}
}

func TestIsEnabled(t *testing.T) {
	c := New(DefaultConfig())
	if !c.IsEnabled() {
		t.Fatal("expected controller to be enabled by default")
	}
}

func TestNormalState_NoConstraint(t *testing.T) {
	c := New(DefaultConfig())
	collector := resolver.NewConstraintCollector()
	state := controllers.DeviceState{
		// Set "normal" values that should NOT trigger constraints
		BatterySOC: 50.0,
		GridPowerW: 0,
	}

	if err := c.Run(state, collector); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(collector.Constraints()) != 0 {
		t.Fatalf("expected 0 constraints in normal state, got %d", len(collector.Constraints()))
	}
}

func TestTriggerCondition_AddsConstraint(t *testing.T) {
	c := New(DefaultConfig())
	collector := resolver.NewConstraintCollector()
	state := controllers.DeviceState{
		// Set values that SHOULD trigger constraints
	}

	if err := c.Run(state, collector); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	constraints := collector.Constraints()
	if len(constraints) != 1 {
		t.Fatalf("expected 1 constraint, got %d", len(constraints))
	}
	// Verify constraint values
}

func TestEdgeCase_BoundaryValues(t *testing.T) {
	// Test exact boundary values (e.g., SOC == threshold)
	c := New(DefaultConfig())
	collector := resolver.NewConstraintCollector()
	state := controllers.DeviceState{
		// Set boundary values
	}

	if err := c.Run(state, collector); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify behavior at boundary
}
```

### Step 4: Register in Scheduler

In `cmd/ems-core/main.go`, add the controller to the scheduler in the correct priority position:

```go
import "<controller_name>pkg" "github.com/vmo/local-ems/internal/engine/controllers/<controller_name>"

// In controller setup:
ctrls := []controllers.Controller{
    limitDischargeCtrl,    // P1 — NEVER change position
    sellToGridLimitCtrl,   // P2 — NEVER change position
    peakShavingCtrl,       // P3
    timeOfUseCtrl,         // P3
    evChargingCtrl,        // P3
    newCtrl,               // P3 — add new controller here
    balancingCtrl,         // P4 — always last
}
```

### Step 5: Run tests

```bash
# Unit tests for the new controller
go test ./internal/engine/controllers/<controller_name>/... -v

# All tests to verify no conflicts
go test ./... -v
```

## Checklist

- [ ] Package created under `internal/engine/controllers/<name>/`
- [ ] Implements `Controller` interface: `ID()`, `Run()`, `IsEnabled()`
- [ ] Config struct with `DefaultConfig()`
- [ ] `Run()` ONLY reads state and adds constraints — NO hardware writes
- [ ] Uses `*int` (pointer) for constraint fields, `nil` = no limit
- [ ] `intPtr()` helper included
- [ ] Unit tests cover: normal state, trigger condition, boundary values
- [ ] Registered in Scheduler at correct priority
- [ ] All tests pass: `go test ./...`

## Examples from Project

### Simple constraint controller (limit_discharge pattern):
- Read a state field (SOC)
- If below threshold → add constraint blocking discharge
- If critical → add ForcePower constraint

### Pure advisor controller (balancing pattern):
- `Run()` adds NO constraints
- Implements `PowerAdvisor` interface
- `DesiredPower()` returns suggested power based on grid state

### Hybrid controller (peak_shaving pattern):
- Adds constraints (MaxDischargePower cap)
- Also implements `PowerAdvisor` for desired discharge amount
