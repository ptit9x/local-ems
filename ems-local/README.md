# ⚡ Local EMS — On-Premise Energy Management System

Constraint-based Battery Energy Storage System (BESS) controller written in Go. It runs the control loop independently of Cloud and buffers telemetry locally when MQTT is unavailable. The executable currently supports simulation mode only; physical-site commissioning is not yet enabled in `cmd/ems-core`.

## Quick Start

```bash
# Prerequisites: Go 1.27+, GCC/Clang (for SQLite CGO)
# macOS: xcode-select --install
# Ubuntu: sudo apt install build-essential

# Run in simulation mode (no hardware needed)
go run ./cmd/ems-core --simulate

# Open http://localhost:8080
```

## Architecture

```
cmd/ems-core/main.go        ← Entry point
internal/
├── engine/                  ← Core: Constraint-based Controller Chain
│   ├── controllers/         ←   6 independent controllers (balancing, peak shaving, ...)
│   ├── cycle/               ←   1-second control loop
│   ├── distributor/         ←   Splits setpoint across 24 PCS units
│   ├── resolver/            ←   Aggregates constraints → final setpoint
│   └── scheduler/           ←   Priority-ordered controller execution
├── devices/                 ← Hardware abstraction (BMS, PCS, Meter, EV Charger)
│   ├── drivers/             ←   Multi-brand PCS drivers (Generic, Sungrow, Sinexcel, LiteOn)
│   └── aggregator/          ←   Power summation across devices
├── modbus/                  ← Modbus TCP client, server, connection pool
├── ocpp/                    ← EV Charger OCPP 1.6J WebSocket server
├── sync/                    ← MQTT store-and-forward to Cloud
├── store/                   ← SQLite local storage with retention & outbox
├── simulator/               ← Hardware simulator for development
├── failover/                ← Heartbeat & fault detection
└── ui/                      ← Embedded web dashboard (dark/light, EN/VN)
web/                         ← HTML/CSS/JS source (embedded via go:embed)
config/                      ← YAML configs & PCS register profiles
```

The simulator supports time-varying solar/load behavior and a deterministic scenario at `config/scenarios/full-demo.yaml`. The CDI demo topology contains 4 BMS, 24 PCS, 3 meters, and 6 EV chargers.

## Control Strategy (Constraint Chain)

Each 1-second cycle:
1. Read all devices → `DeviceState` snapshot
2. Run controllers in priority order → each adds `Constraint`
3. Resolver picks tightest constraint → final setpoint
4. Distributor splits setpoint across PCS units
5. Write to hardware via Modbus

| Priority | Controller | Purpose |
|----------|-----------|---------|
| P1 | `limit_discharge` | Block discharge at low SOC, force charge at critical |
| P2 | `sell_to_grid_limit` | Prevent grid export beyond allowed limit |
| P3 | `peak_shaving` | Discharge to shave grid demand peaks |
| P3 | `time_of_use` | Charge off-peak, discharge peak hours |
| P3 | `ev_charging` | Dynamic Load Management for EV chargers |
| P4 | `balancing` | Default self-consumption optimization |

## Build & Test

```bash
make build          # Build for current OS
make test           # Run all tests
make docker-all     # Cross-compile via Docker (Linux/Windows/ARM)
make run            # Run simulation mode
```

## Configuration

```bash
go run ./cmd/ems-core --config config/production.yaml
go run ./cmd/ems-core --simulate                        # No hardware needed
go run ./cmd/ems-core --config config/cdi-singapore.yaml --simulate
```

Runtime flags include `--config`, `--simulate`, `--ui-port`, `--db`, and `--version`. Configuration precedence is environment variables, the selected `.env` file, YAML values, then defaults.

## Adding a New PCS Brand

1. Create `config/pcs_profiles/your-brand.yaml` with register map
2. Set `brand: "your-brand"` in site config
3. No code changes needed — the driver loader picks it up automatically

For brands with non-standard protocols, implement the `PCSDriver` interface in `internal/devices/drivers/`.

## Relation to Cloud Platform

This binary syncs telemetry to the Cloud via MQTT using the schema defined in [`../contracts/`](../contracts/). See the root [README](../README.md) for full platform overview.
