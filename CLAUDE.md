# CLAUDE.md — Guidelines for AI Coding Agents

## Project Overview

This is the **EMS Platform** monorepo — a Battery Energy Storage System (BESS) management platform.

| Sub-project | Path | Description |
|-------------|------|-------------|
| **Local EMS** | `ems-local/` | On-premise Go controller (5MWh BESS), constraint-based control chain |
| **Cloud EMS** | `ems-cloud/` | Multi-tenant fleet management (NestJS + React) — Phase B, not started |
| **Contracts** | `contracts/` | Shared MQTT schemas (JSON Schema) for Local ↔ Cloud communication |
| **Docs** | `docs/` | Architecture, specs, plans, screenshots |

**Current Status:** Local EMS Phase 0–5 complete. Phase 6 (Multi-rate Polling), Phase 7 (Cloud Sync), Phase 8 (Production Hardening) are planned.

## Core Architecture

### Constraint-based Controller Chain

- **DO NOT** use if-else or direct overrides for charge/discharge decisions.
- Each controller **adds constraints** (power limits), NOT direct setpoints.
- The **ESS Power Resolver** aggregates all constraints → computes the final setpoint.
- The tightest constraint always wins (takes MIN of all MaxDischarge, MIN of all MaxCharge).

### Power Sign Conventions

| Direction | Sign | Example |
|-----------|------|---------|
| **Discharge** (battery → grid/load) | Positive (+) | `+50000` = 50kW discharge |
| **Charge** (grid/solar → battery) | Negative (−) | `-50000` = 50kW charge |
| **Grid buying** (importing from grid) | Positive (+) | `GridPowerW = +3000` |
| **Grid selling** (exporting to grid) | Negative (−) | `GridPowerW = -500` |
| **Solar generating** | Positive (+) | `SolarPowerW = +8000` |
| **Load consuming** | Positive (+) | `LoadPowerW = +5000` |

### Cycle Manager (1-second loop)

Each 1-second cycle executes sequentially:
1. Snapshot process image (channel values)
2. Compute aggregations (Sum)
3. Reset constraints
4. Run controllers in priority order
5. Resolver aggregates constraints → setpoint
6. Write setpoint to PCS via Modbus
7. Log & persist

### Controller Priority Order

| Priority | Controller | Package | Purpose |
|----------|-----------|---------|---------| 
| **P1** | `limit_discharge` | `ems-local/internal/engine/controllers/limit_discharge/` | Hardware protection — block discharge at low SOC, force charge at critical SOC |
| **P2** | `sell_to_grid_limit` | `ems-local/internal/engine/controllers/sell_to_grid_limit/` | Grid compliance — prevent exporting beyond allowed limit (default: zero-export) |
| **P3** | `peak_shaving` | `ems-local/internal/engine/controllers/peak_shaving/` | Cost optimization — discharge to shave grid demand peaks |
| **P3** | `time_of_use` | `ems-local/internal/engine/controllers/time_of_use/` | Cost optimization — charge at off-peak rates, discharge at peak rates |
| **P3** | `ev_charging` | `ems-local/internal/engine/controllers/ev_charging/` | DLM — manage site power budget for EV chargers |
| **P4** | `balancing` | `ems-local/internal/engine/controllers/balancing/` | Default — self-consumption optimization (solar → battery → load) |

### DeviceState (Process Image)

All controllers receive a read-only `DeviceState` snapshot each cycle:

```go
type DeviceState struct {
    GridPowerW       int     // Grid meter: positive=buying, negative=selling
    SolarPowerW      int     // Solar meter: positive=generating
    LoadPowerW       int     // Load meter: positive=consuming
    BatterySOC       float64 // Battery SOC 0-100%
    BatteryTempC     float64 // Battery temperature in Celsius
    ESSActivePowerW  int     // Current ESS power: positive=discharge, negative=charge
    CellVoltageMinMV int     // Min cell voltage in millivolts
    CellVoltageMaxMV int     // Max cell voltage in millivolts
}
```

### Constraint Struct

```go
type Constraint struct {
    MaxDischargePower *int   // Cap discharge watts, nil = no limit
    MaxChargePower    *int   // Cap charge watts, nil = no limit
    ForcePower        *int   // Override setpoint (negative=charge), nil = no force
    Source            string // Controller ID that created this constraint
}
```

**Pointer semantics:** `nil` = "no limit". Use `intPtr(0)` to set a zero limit.

### Device Adapters

| Device | Adapter | Reading Struct | Writable | Source |
|--------|---------|---------------|----------|--------|
| Grid/Solar Meter | `MeterAdapter` | `MeterReading` | No | `ems-local/internal/devices/registers.go` |
| Battery (BMS) | `BMSAdapter` | `BMSReading` | No | `ems-local/internal/devices/registers.go` |
| Inverter (PCS) | `PCSAdapter` / `ProfiledPCSAdapter` | `PCSReading` | Yes (`WriteSetpoint`) | `ems-local/internal/devices/drivers/` |
| EV Charger | `EVChargerAdapter` | `EVChargerData` | Yes (`WriteMaxPower`, `WriteStartStop`) | `ems-local/internal/devices/evcharger.go` |

### PCS Profile Registry

Multi-brand PCS support via `PCSProfileRegistry` in `ems-local/internal/devices/drivers/profile.go`:
- `generic` — Default register map (registers 0–3)
- `sungrow` — Sungrow PCS (registers 5000+, scale 1.0)
- `sinexcel` — Sinexcel PCS (registers 100+, scale 10.0)

Add new brands by calling `registry.Register(PCSProfile{...})` in `DefaultPCSProfileRegistry()`.

## Code Conventions

### Language & Style

- **Go 1.27** — Use the latest Go features.
- Follow [Effective Go](https://go.dev/doc/effective_go) and `go vet` / `golint`.
- Package names: lowercase, single word (e.g., `resolver`, `simulator`).
- Error handling: always wrap errors with context (`fmt.Errorf("doing X: %w", err)`).

### Directory Structure

```
ems-local/                 → Local EMS Go project
  cmd/ems-core/            → Entrypoint. Only contains main.go, wires up dependencies.
  internal/                → Business logic. NOT exported outside the module.
    config/                → YAML config loader + validation.
    devices/               → Device adapters (Meter, BMS, PCS, EV Charger).
      drivers/             → Multi-brand PCS drivers (Generic, Sungrow, Sinexcel, LiteOn).
      aggregator/          → System state aggregation (Sum).
    engine/                → Control engine.
      controllers/         → Controller implementations (one package per controller).
        limit_discharge/   → P1: Hardware protection.
        sell_to_grid_limit/→ P2: Grid compliance.
        peak_shaving/      → P3: Peak demand reduction.
        time_of_use/       → P3: Time-based charge/discharge.
        ev_charging/       → P3: EV charger DLM.
        balancing/         → P4: Self-consumption default.
      cycle/               → 1-second control loop.
      scheduler/           → Controller priority scheduler.
      resolver/            → Constraint → setpoint resolver.
      distributor/         → Power distribution across PCS units.
    failover/              → Heartbeat & fault detection.
    health/                → Watchdog & health checks.
    modbus/                → Zero-dependency Modbus TCP client/server.
    ocpp/                  → OCPP 1.6J protocol for EV Chargers.
    simulator/             → Hardware simulator (fake Modbus devices).
    store/                 → SQLite storage (WAL + auto-retention).
    sync/                  → Cloud sync (MQTT, store-and-forward).
    topology/              → BESS group & site topology.
    ui/                    → Embedded web dashboard (server, websocket, auth).
  web/                     → HTML/CSS/JS source (embedded via go:embed).
  config/                  → YAML configuration files (dev.yaml, production.yaml).
    pcs_profiles/          → PCS register profile YAMLs.
  scripts/                 → Install/uninstall scripts.
contracts/                 → Shared MQTT schemas (JSON Schema).
docs/                      → Documentation.
  architecture/            → System architecture diagrams.
  specs/                   → Requirements & Q&A.
  plans/                   → Design plans & roadmaps.
  screenshots/             → UI screenshots.
```

### Key Conventions

1. **Controllers**: Each controller lives in its own package under `ems-local/internal/engine/controllers/`. Must implement the `Controller` interface (`ID()`, `Run()`, `IsEnabled()`) defined at `ems-local/internal/engine/controllers/controller.go`. Optionally implement `PowerAdvisor` (`DesiredPower()`) for controllers that suggest a target setpoint.
2. **Constraints**: Use the `Constraint` struct with pointer fields (`*int`) — `nil` means no limit. Use `intPtr(v)` helper to create pointers.
3. **Devices**: Each device type (Meter, BMS, PCS, EV Charger) has its own adapter in `ems-local/internal/devices/`. Device lifecycle is managed via `ems-local/internal/devices/manager.go`. Device registry supports `Device`, `ControllableDevice`, and `ReconfigurableDevice` interfaces.
4. **PCS Profiles**: Multi-brand PCS register profiles are defined in `ems-local/internal/devices/drivers/profile.go`. Add new brands by registering a `PCSProfile` in `DefaultPCSProfileRegistry()`.
5. **Simulator**: Deterministic simulation (seed=42) in `ems-local/internal/simulator/`. Solar=sine wave, Load=random, SOC=ramp. Always test with the simulator before testing on real hardware.
6. **Config**: Uses YAML (`ems-local/config/dev.yaml`, `ems-local/config/production.yaml`). Loaded via `ems-local/internal/config/`. Supports site, devices, topology, polling, engine, storage, sync, and ui sections.
7. **Thread Safety**: Device adapters that store data for concurrent access use `sync.RWMutex`. Controllers with dynamic runtime config (e.g., `ev_charging`) also use mutex protection.
8. **Security Hardening** (local deployment):
   - **Session store**: Uses `sync.Mutex` (full lock) in `validate()` — never `RLock()` when a write (`delete`) may occur.
   - **WebSocket limit**: Max `100` concurrent connections (`maxWSClients` in `websocket.go`). New connections are rejected with `conn.Close()` when limit is reached.
   - **HTTP timeouts**: All HTTP servers must set `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, and `MaxHeaderBytes` to prevent Slowloris and connection leak attacks.
   - **OCPP frame limit**: WebSocket frames are capped at `1 MB` (`maxOCPPFrameSize` in `ocpp/server.go`). Oversized frames are rejected before buffer allocation.
   - When adding new network-facing code, always consider: connection limits, read timeouts, input size limits, and proper lock semantics.

### Testing

- **Unit tests**: Placed in the same package, in `*_test.go` files.
- Every NEW controller must have unit tests covering: normal state (no constraints), trigger condition, boundary values, disabled state.
- Run `cd ems-local && go test ./internal/... -count=1` to execute all unit tests.
- Run `cd ems-local && go test ./... -count=1` to execute everything.
- Or use `make test` from root or from `ems-local/`.

## Adding a New Controller

1. Create a new package under `ems-local/internal/engine/controllers/<controller_name>/`.
2. Implement the `Controller` interface: `ID() string`, `Run(state, collector) error`, `IsEnabled() bool`.
3. Add a `Config` struct with a `DefaultConfig()` function for tunable parameters.
4. The controller must only **read** the process image and **add constraints** — DO NOT write directly to hardware.
5. Optionally implement `PowerAdvisor` interface if the controller suggests a desired power setpoint.
6. Register the controller in the Scheduler with the appropriate priority (P1/P2 are safety-critical, P3 for optimization, P4 for defaults).
7. Write comprehensive unit tests covering: normal state, trigger conditions, boundary values, disabled controller.
8. Run integration tests to verify no conflicts with existing controllers.

## Adding a New Device Adapter

1. Define register constants in `ems-local/internal/devices/registers.go` (or in the adapter file for complex devices like `evcharger.go`).
2. Create a `<DeviceName>Reading` struct with engineering units.
3. Create a `<DeviceName>Adapter` with `Poll()` / `Read()` methods.
4. For writable devices, implement `Write<Parameter>()` methods.
5. Use `sync.RWMutex` if data is accessed concurrently.
6. Add the `DeviceType` constant to `device.go`.
7. Update YAML config schema.
8. Add simulator support if needed.

## Adding a New PCS Brand

1. Get register addresses from the vendor Modbus datasheet.
2. Create a `PCSProfile` struct with register addresses, scale factor, and status map.
3. Register in `DefaultPCSProfileRegistry()` in `ems-local/internal/devices/drivers/profile.go`.
4. Write unit tests verifying: profile registration, scale factor, status map completeness.
5. Update YAML config to allow the new brand name.

## Modifying Code

- **DO NOT** change controller priority order without fully understanding the impact. `limit_discharge` (P1) and `sell_to_grid_limit` (P2) are safety-critical.
- **DO NOT** bypass the constraint system. All charge/discharge decisions must go through the Resolver.
- **DO NOT** write directly to hardware from controllers. Use only `collector.Add(Constraint{...})`.
- Preserve comments and docstrings unrelated to your changes.
- Commit messages in English, following conventional commits (`feat:`, `fix:`, `refactor:`...).

## AI Skills (Code Generation)

The following skills are available in `.claude/skills/` (project-local) to accelerate development:

| Skill | Trigger | What It Generates |
|-------|---------|-------------------|
| `ems-controller-generator` | "new controller", "add controller" | Controller package with interface impl, config, tests, scheduler registration |
| `ems-device-adapter` | "new device", "add adapter" | Device adapter with register map, reading struct, poll/write, tests |
| `ems-simulator-scenario` | "simulation scenario", "replay test" | Simulation test with scenario config, cycle loop, assertions |
| `ems-modbus-register-map` | "register map", "new PCS brand" | PCS profile with register addresses, scale factor, status map, tests |

## Dependencies

- `github.com/mattn/go-sqlite3` — SQLite driver (requires CGO).
- `gopkg.in/yaml.v3` — YAML parser.
- Keep dependencies minimal. Do not add heavy frameworks if stdlib is sufficient.

## License

This project is licensed under the [MIT License](LICENSE).

## Reference Documentation

- [EMS MVP Plan](docs/plans/ems-mvp-plan.md) — Detailed design & task roadmap.
- [Cloud ERP Plan](docs/plans/cloud-erp-plan.md) — Cloud ERP plan (future phase).
- [Architecture Diagram](docs/architecture/architecture.jpg) — System architecture overview.
- [Contracts README](contracts/README.md) — MQTT schemas for Local ↔ Cloud.
