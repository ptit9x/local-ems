---
name: ems-simulator-scenario
description: "Use when creating a new simulation test scenario for the Local EMS project. Also use when adding replay tests, simulation data, or verifying controller behavior under specific conditions."
metadata:
  category: technique
  author: richard
  triggers: simulation scenario, replay test, simulator test, test scenario,
    verify controller, simulation data, fake data, ems test
risk: low
source: local
---

# EMS Simulator Scenario Generator

Generates simulation test scenarios that verify controller behavior under specific energy conditions, using the project's Simulator + Controller Chain architecture.

## When to Use

- Creating a new test scenario (e.g., "test cloud cover drop", "test peak demand spike")
- Verifying controller behavior under edge conditions
- Adding replay tests in `test/simulation/`
- User says "add simulation", "test scenario", "verify behavior"

## Architecture Context

### Simulator Design

The [Simulator](file:///Users/richard/vmo_bid_cdi/internal/simulator/simulator.go) generates `DeviceState` each tick (1 second):

```go
type Simulator struct {
    tick           int
    battSOC        float64
    essActivePower int
    config         Config
    rng            *rand.Rand  // seed=42, deterministic
}

type Config struct {
    InitialSOC        float64 // default: 50.0
    SolarPeakW        int     // default: 10000
    LoadBaseW          int     // default: 5000
    LoadVarianceW      int     // default: 3000
    BatteryCapacityWh int     // default: 5000000 (5MWh)
    MaxChargeRateW    int     // default: 50000
    MaxDischargeRateW int     // default: 50000
}
```

### Data Generation Patterns

| Signal | Method | Formula |
|--------|--------|---------|
| Solar | Sine wave | `peak * sin(π * tickInDay / 86400)`, 0 at night |
| Load | Random | `base + random(-variance/2, +variance/2)`, min 500W |
| SOC | Ramp | Updates from `ApplyPower()`: `deltaSOC = power / (3600 * capacity) * 100` |
| Temperature | Noise | `25°C + random(-1, +1)` |
| Cell Voltage | Linear | `3200mV + SOC/100 * 1000mV`, ±10mV variance |
| Grid | Calculated | `Load - Solar + ESS` |

### Test Flow

```
SimConfig → Simulator.NextState() → CycleManager.RunOnce(state) → Assert CycleResult
    └→ Simulator.ApplyPower(result.Setpoint) → loop
```

## Generation Steps

### Step 1: Define Scenario Parameters

Identify the scenario:
- **What condition** are you testing? (low SOC, high demand, cloud cover, night, etc.)
- **Which controllers** should activate?
- **What is the expected outcome?** (constraint applied, setpoint clamped, etc.)

### Step 2: Create Test File

Create `test/simulation/<scenario_name>_test.go`:

```go
package simulation

import (
	"testing"

	"github.com/vmo/local-ems/internal/engine/controllers"
	"github.com/vmo/local-ems/internal/engine/controllers/balancing"
	"github.com/vmo/local-ems/internal/engine/controllers/limit_discharge"
	"github.com/vmo/local-ems/internal/engine/controllers/sell_to_grid_limit"
	// ... import all controllers needed
	"github.com/vmo/local-ems/internal/engine/cycle"
	"github.com/vmo/local-ems/internal/engine/resolver"
	"github.com/vmo/local-ems/internal/engine/scheduler"
	"github.com/vmo/local-ems/internal/simulator"
)

func TestScenario_<ScenarioName>(t *testing.T) {
	// 1. Configure simulator for the specific scenario
	simCfg := simulator.Config{
		InitialSOC:        <value>,  // Set SOC for scenario
		SolarPeakW:        <value>,
		LoadBaseW:          <value>,
		LoadVarianceW:      0,       // 0 for deterministic tests
		BatteryCapacityWh: 5000000,
		MaxChargeRateW:    50000,
		MaxDischargeRateW: 50000,
	}
	sim := simulator.New(simCfg)

	// 2. Set up controller chain (same priority order as production)
	ldCtrl := limit_discharge.New(limit_discharge.DefaultConfig())
	stglCtrl := sell_to_grid_limit.New(sell_to_grid_limit.DefaultConfig())
	// ... other controllers
	balCtrl := balancing.New(balancing.DefaultConfig())

	ctrls := []controllers.Controller{
		ldCtrl,    // P1
		stglCtrl,  // P2
		// ... P3 controllers
		balCtrl,   // P4
	}

	sched := scheduler.New(ctrls)
	collector := resolver.NewConstraintCollector()
	res := resolver.New(collector)

	advisors := []cycle.PowerAdvisor{balCtrl}
	mgr := cycle.New(sched, collector, res, advisors, 0)

	// 3. Run simulation for N cycles
	const numCycles = <N>
	for i := 0; i < numCycles; i++ {
		state := sim.NextState()
		result := mgr.RunOnce(state)
		sim.ApplyPower(result.ResolvedPower.Setpoint)

		// 4. Assert expected behavior at specific ticks
		switch {
		case i == 0:
			// Initial state assertions
			if result.ResolvedPower.Setpoint > 0 {
				t.Logf("tick %d: discharging %dW", i, result.ResolvedPower.Setpoint)
			}
		case i >= <trigger_tick>:
			// Assert constraint activated
			// Example: SOC dropped below threshold
			if state.BatterySOC < 15.0 && result.ResolvedPower.Setpoint > 0 {
				t.Errorf("tick %d: SOC=%.1f%% but still discharging %dW",
					i, state.BatterySOC, result.ResolvedPower.Setpoint)
			}
		}
	}

	// 5. Final state assertions
	finalSOC := sim.SOC()
	if finalSOC < 0 || finalSOC > 100 {
		t.Errorf("final SOC out of bounds: %.1f%%", finalSOC)
	}
}
```

### Step 3: Common Scenario Templates

#### Template A: Low SOC Protection
```go
simCfg := simulator.Config{
    InitialSOC:   12.0,  // Start near threshold
    SolarPeakW:   0,     // No solar (worst case)
    LoadBaseW:     8000,  // High load
    LoadVarianceW: 0,
    // ...
}
// Assert: limit_discharge blocks discharge, force charge at SOC < 10%
```

#### Template B: Zero-Export Compliance
```go
simCfg := simulator.Config{
    InitialSOC:   80.0,
    SolarPeakW:   20000, // High solar
    LoadBaseW:     2000,  // Low load → excess exported
    LoadVarianceW: 0,
    // ...
}
// Assert: sell_to_grid_limit increases charge to absorb excess
```

#### Template C: Peak Demand Spike
```go
simCfg := simulator.Config{
    InitialSOC:   60.0,
    SolarPeakW:   5000,
    LoadBaseW:     15000, // Very high load
    LoadVarianceW: 0,
    // ...
}
// Assert: peak_shaving discharges to offset grid demand
```

#### Template D: Night Charge Schedule
```go
simCfg := simulator.Config{
    InitialSOC:   30.0,
    SolarPeakW:   0,     // Night, no solar
    LoadBaseW:     2000,
    LoadVarianceW: 0,
    // ...
}
// Assert: time_of_use charges during off-peak hours
```

#### Template E: Controller Priority Conflict
```go
simCfg := simulator.Config{
    InitialSOC:   13.0,  // Below MinSOC
    SolarPeakW:   0,
    LoadBaseW:     20000, // Peak demand → peak_shaving wants discharge
    LoadVarianceW: 0,
    // ...
}
// Assert: limit_discharge (P1) wins → discharge blocked despite peak demand
```

### Step 4: Run Tests

```bash
# Run specific scenario
go test ./test/simulation/ -run TestScenario_<Name> -v

# Run all simulation tests
go test ./test/simulation/... -v

# Run with coverage
go test ./test/simulation/... -v -cover
```

## Checklist

- [ ] Scenario clearly defined: condition, expected behavior
- [ ] Simulator config uses `LoadVarianceW: 0` for deterministic tests
- [ ] All controllers included in scheduler (same priority as production)
- [ ] Assertions at specific ticks verify constraint activation
- [ ] Final state assertions verify system stability
- [ ] No flaky tests (deterministic RNG with seed=42)
- [ ] All tests pass: `go test ./test/simulation/... -v`

## Common Assertion Patterns

```go
// Verify discharge blocked
if result.ResolvedPower.Setpoint > 0 {
    t.Errorf("expected no discharge, got %dW", result.ResolvedPower.Setpoint)
}

// Verify charging
if result.ResolvedPower.Setpoint >= 0 {
    t.Errorf("expected charge (negative), got %dW", result.ResolvedPower.Setpoint)
}

// Verify force charge
if result.ResolvedPower.Setpoint > -500000 {
    t.Errorf("expected force charge ≤ -500kW, got %dW", result.ResolvedPower.Setpoint)
}

// Verify setpoint clamped to range
if result.ResolvedPower.Setpoint < -50000 || result.ResolvedPower.Setpoint > 50000 {
    t.Errorf("setpoint out of hardware range: %dW", result.ResolvedPower.Setpoint)
}

// Verify SOC stability over N cycles
if math.Abs(finalSOC - initialSOC) > 5.0 {
    t.Errorf("SOC drifted too much: %.1f%% → %.1f%%", initialSOC, finalSOC)
}
```
