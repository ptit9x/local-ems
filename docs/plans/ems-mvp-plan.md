# Local EMS (Energy Management System) System Design

## Goal
Build a Local EMS system in Go (Golang) for a 5MWh BESS. The system supports 100% hardware simulation (Simulator), Modbus/OCPP communication, independent offline operation without network connectivity (Store-and-Forward + Local UI), and integrated advanced energy control algorithms following the OpenEMS architecture.

## Software Architecture (Clean Architecture + Constraint-based Controller Chain)

### 1. File Structure
```text
/cmd
  /ems-core             # Main process (entrypoint)
/internal
  /config               # YAML/SQLite Config loader
  /simulator            # SIMULATOR: Initialize Fake Modbus Servers (Meters, BMS, PCS)
  /modbus               # Modbus Client (TCP/RTU) connecting to hardware
  /devices              # Hardware adapters (Meter, BMS, PCS)
  /ocpp                 # WebSocket handling for EV Chargers
  /channel              # Channel system: safe concurrent read/write with Process Image snapshot each cycle
  /engine               # DECISION ENGINE
     /cycle             # Cycle Manager: 1s main loop (read → process → write)
     /scheduler         # Scheduler: determines controller execution order (FixedOrder)
     /resolver          # ESS Power Resolver: aggregates constraints → final setpoint written to PCS
     /controllers
        /limit_discharge      # Prevent deep battery discharge (configurable: minSoc, forceChargeSoc)
        /sell_to_grid_limit   # Limit feed-in power to grid (configurable: maxSellToGridPower)
        /peak_shaving         # Peak load shaving (battery compensation when office load is high)
        /time_of_use          # Charge at off-peak rates, discharge at on-peak rates
        /balancing            # Self-consumption balancing
  /health               # Watchdog & health check for the process
  /sync                 # Local InfluxDB storage & MQTT sync when online
  /ui                   # Local HMI (REST API / WebSockets)
/web                    # Frontend HMI source code
/test
  /integration          # Integration tests for Controller Chain
  /simulation           # Replay data → verify output
```

### 2. Algorithmic Model (Inherited from OpenEMS — Constraint Accumulation)

Instead of a giant "if-else" function or an "absolute override" pattern, we use a **Constraint-based Controller Chain** architecture similar to OpenEMS:

**Operating Principle:**
1. **Scheduler** determines the controller execution order (higher priority → runs first).
2. **Every controller runs**, and places **constraints** on the ESS (e.g., `MaxDischargePower ≤ 0`, `MaxChargePower ≤ 2000W`).
3. Higher priority controllers place constraints first — their constraints will be **tighter** and prioritized for retention.
4. After all controllers have finished executing, the **ESS Power Resolver** aggregates all constraints to determine the final feasible setpoint to write to the PCS.

```go
// Each controller adds constraints, does NOT override directly
type Constraint struct {
    MaxDischargePower *int   // Discharge power limit (W), nil = unlimited
    MaxChargePower    *int   // Charge power limit (W), nil = unlimited
    ForcePower        *int   // Forced setpoint (used only for force charge scenarios)
    Source            string // Name of the controller creating the constraint
}

// Resolver selects the MIN value of all MaxDischarge, MIN of all MaxCharge
// → ensures the tightest constraint always wins
```

**Controller Execution Order (Scheduler FixedOrder):**
*   **Priority 1 (Hardware Protection):** `limit_discharge`. SOC < `minSoc` (default: 15%) → constraint `MaxDischargePower = 0`. SOC < `forceChargeSoc` (default: 10%) → constraint `ForcePower = chargeRate`. Subsequent controllers still run, but their constraints will be superseded by the tighter constraint in the resolver.
*   **Priority 2 (Grid Compliance):** `sell_to_grid_limit`. If the Grid Meter reports sell-to-grid power exceeding `maxSellToGridPower` (default: 0W for zero-export) → constraint increases charge or decreases discharge accordingly.
*   **Priority 3 (Cost Optimization):** `peak_shaving` — constraint discharges battery when load exceeds the peak clipping threshold. `time_of_use` — constraint charges during off-peak rate hours, discharges during on-peak rate hours.
*   **Priority 4 (Default):** `balancing`. Ensures solar power prioritizes charging the battery, falling back to the grid when insufficient.

### 3. Execution Cycle (Cycle Manager)

Each cycle (~1 second), Cycle Manager executes sequentially:
```text
┌─────────────────────────────────────────────────┐
│ 1. BEFORE_PROCESS_IMAGE                         │
│    → Snapshot all channel values                │
│                                                 │
│ 2. AFTER_PROCESS_IMAGE                          │
│    → Calculate Sum (total power, SOC...)        │
│                                                 │
│ 3. BEFORE_CONTROLLERS                           │
│    → Reset constraint list                      │
│                                                 │
│ 4. EXECUTE_CONTROLLERS                          │
│    → Scheduler calls each controller in order   │
│    → Each controller reads process image, adds  │
│      constraints to resolver                    │
│                                                 │
│ 5. AFTER_CONTROLLERS                            │
│    → Resolver aggregates constraints → setpoint │
│                                                 │
│ 6. EXECUTE_WRITE                                │
│    → Write setpoint to PCS via Modbus           │
│                                                 │
│ 7. AFTER_WRITE                                  │
│    → Log, write to InfluxDB, emit events        │
└─────────────────────────────────────────────────┘
```

## Tasks Roadmap (Development Phases)

### Phase 0: Spike — Validate Constraint Resolver Architecture (1-2 days)
- [x] Task 0.1: Small PoC consisting of 2 fake controllers + constraint resolver. Run on hardcoded data, verify the tightest constraint always wins.
- [x] Task 0.2: Write unit tests for resolver: multiple overlapping constraints → correct output.

### Phase 1: Foundation & Simulation (Simulator)
- [x] Task 1.1: Initialize Go project, setup directory structure, `go.mod`.
- [ ] Task 1.2: Write `/internal/channel` module — Channel system with Process Image (thread-safe snapshot for concurrent read).
- [x] Task 1.3: Write `/internal/simulator` module to create Fake Modbus Servers for Meter, BMS, PCS.
- [x] Task 1.4: Build simulation data generator script: Sinusoidal for Solar PV, Random for Office Load, SOC ramp for BMS.
- [x] Task 1.5: Write unit tests for simulator (valid data ranges, Modbus registers in correct format).

### Phase 2: Device Connectivity & Core Communication
- [x] Task 2.1: Write Modbus Client reading parameters from Fake Meter (Grid & Solar).
- [x] Task 2.2: Write Adapter reading Fake BMS status (SOC, Temperature, Cell voltage limits).
- [x] Task 2.3: Write Adapter writing charge/discharge commands to Fake PCS (Active/Reactive Power setpoint).
- [x] Task 2.4: Write unit tests for each adapter (mock Modbus server → verify parsed values).

### Phase 3: Control Algorithms (Decision Engine)
- [x] Task 3.1: Code `Controller` interface + `Constraint` struct. Write Cycle Manager (1s loop).
- [x] Task 3.2: Code `Scheduler` (FixedOrder) — accepts list of controller IDs, calls in sequence.
- [x] Task 3.3: Code `ESS Power Resolver` — receives []Constraint → calculates final setpoint.
- [x] Task 3.4: Implement `limit_discharge` (configurable `minSoc`, `forceChargeSoc`) + unit tests.
- [x] Task 3.5: Implement `sell_to_grid_limit` (configurable `maxSellToGridPower`, default 0) + unit tests.
- [x] Task 3.6: Implement `peak_shaving` (configurable `peakThreshold`) + unit tests.
- [x] Task 3.7: Implement `time_of_use` (configurable hourly electricity tariff schedule) + unit tests.
- [x] Task 3.8: Implement `balancing` (self-consumption optimization) + unit tests.
- [x] Task 3.9: Integration test: assemble all controllers → Scheduler → Resolver → run on simulator data.

### Phase 4: Local UI & Offline Sync
- [x] Task 4.1: Embed SQLite, establish buffer service. *(InfluxDB → SQLite for MVP)*
- [ ] Task 4.2: Code `sync` module to publish MQTT when online (Store-and-Forward). *(deferred to post-MVP)*
- [x] Task 4.3: Build REST API + WebSocket for Local Web UI (Display real-time power flow).
- [x] Task 4.4: Code `/internal/health` — watchdog heartbeat, restart policy when a controller panics.

### Phase 5: EV Charger & Multi-device (Q&A Update)
- [x] Task 5.1: Device Plugin Architecture — `Device`, `ControllableDevice` interfaces, `DeviceRegistry`
- [x] Task 5.2: EV Charger Adapter — Modbus read/write for EV charger (status, power, start/stop)
- [x] Task 5.3: Multi-brand PCS Register Map — PCSProfile registry (generic/sungrow/sinexcel)
- [x] Task 5.4: EV Charging Controller — DLM constraint controller (P3 priority)
- [x] Task 5.5: Configurable Device Topology — YAML config for site, bess_groups, topology
- [x] Task 5.6: EV Charger Dashboard — Widget, device chips, detail page on dashboard
- [x] Task 5.7: Aggregator EV Charger — EVChargerState, UpdateEVCharger, EV counts/totals
- [x] Task 5.8: Distributor master/individual mode — BESS group command mode support
- [x] Task 5.9: SQLite WAL mode + 7-day auto-retention
- [ ] Task 5.10: EV active control — auto WriteMaxPower from DLM controller

### Phase 6: Real Hardware & Multi-rate Polling (Planned)
- [ ] Task 6.1: Multi-rate poller (200ms/1s/5s goroutines)
- [ ] Task 6.2: Remove simulation-only guard
- [ ] Task 6.3: Robust data integrity (WAL + retry)

### Phase 7: Cloud Sync (Planned)
- [ ] Task 7.1: MQTT publisher module
- [ ] Task 7.2: Store-and-forward buffer
- [ ] Task 7.3: Cloud failover signal for EV Charger

### Phase 8: Production Hardening (Planned)
- [ ] Task 8.1: Config hot-reload
- [ ] Task 8.2: Per-device health monitoring

## Done When
- [x] `go run ./cmd/ems-core --simulate` starts successfully, simulator continuously generates sinusoidal/random data.
- [x] **Zero-export:** Fake Grid Meter = -500W (sell) → sell_to_grid_limit constraint activates → PCS increases charge → grid power ≥ 0W within ≤ 3 cycles.
- [x] **Prevent battery depletion:** SOC decreases to 15% → discharge constraint = 0. SOC decreases to 10% → force charge activates.
- [x] **Peak shaving:** Load exceeds peak threshold → PCS discharges to compensate, load seen from grid drops below threshold.
- [x] **Time-of-use:** During off-peak rate hours → automatically charge. During on-peak rate hours → automatically discharge.
- [x] **Constraint conflict:** When peak_shaving requests discharge but SOC < minSoc → limit_discharge constraint wins, discharge = 0.
- [x] Local SQLite writes data continuously every cycle. *(MQTT sync deferred to post-MVP)*
- [x] All unit tests pass. Integration test coverage > 80%.
- [x] EV Charger visible on dashboard with real-time status updates
- [x] Constraint-based EV DLM controller in scheduler
- [ ] Cloud sync via MQTT every 5 minutes
- [ ] Multi-rate hardware polling (200ms fast meter)
