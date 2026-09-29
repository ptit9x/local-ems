# Local EMS Deployment Guide

Deployment guide for the Local EMS system in two scenarios: **Smart Home** (Residential) and **Factory** (Industrial/Factory).

## Table of Contents

- [1. Configuration Architecture Overview](#1-configuration-architecture-overview)
- [2. Hardware Requirements](#2-hardware-requirements)
- [3. System Configuration](#3-system-configuration)
- [4. Smart Home Deployment](#4-smart-home-deployment)
- [5. Factory Deployment](#5-factory-deployment)
- [6. Environment Variables](#6-environment-variables)
- [7. System Verification](#7-system-verification)
- [8. Operations & Maintenance](#8-operations--maintenance)

---

## 1. Configuration Architecture Overview

The system uses configuration layers in order of priority (highest → lowest):

```
┌──────────────────────────────────┐
│  1. Environment variables (ENV)  │  ← Highest priority
├──────────────────────────────────┤
│  2. .env file                    │  ← Overrides YAML
├──────────────────────────────────┤
│  3. YAML config file             │  ← Main configuration
├──────────────────────────────────┤
│  4. Default values (defaults)    │  ← Fallback
└──────────────────────────────────┘
```

**Deployment Workflow:**

1. Select the appropriate YAML file (`config/dev.yaml`, `config/production.yaml`, or create a new one).
2. Copy the matching `.env` template (`.env.smart-home` or `.env.factory`) to `.env`.
3. Adjust variables in `.env` according to actual site parameters.
4. Run EMS with the `--config` flag.

> **Note:** The `.env` file is listed in `.gitignore` — never commit it to git as it contains passwords.

---

## 2. Hardware Requirements

### 2.1 Scenario Comparison

| Parameter | Smart Home | Factory |
|-----------|------------|---------|
| BESS Capacity | 5 – 30 kWh | 500 kWh – 10 MWh |
| BMS | 1 | 2 – 4 |
| PCS / Inverter | 1 (hybrid) | 4 – 24 |
| Power Meter | 1 (grid) | 3 (grid + solar + load) |
| EV Charger | 0 – 1 | 2 – 20 |
| Solar PV | 3 – 15 kWp | 100 kWp – 5 MWp |
| EMS Host Machine | Raspberry Pi 4/5 | Industrial PC (x86) |
| Communication | Modbus TCP/RTU | Modbus TCP |

### 2.2 EMS Host Requirements

| Parameter | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | ARM64 or x86_64 | Quad-core |
| RAM | 512 MB | 2 GB |
| Disk | 4 GB | 32 GB SSD |
| Network | 1x Ethernet | 2x Ethernet (device LAN + WAN) |
| OS | Linux (kernel 4.x+) | Ubuntu 22.04 LTS / Debian 12 |

### 2.3 Build Software Requirements

- **Go 1.27+**
- **GCC** (required for SQLite CGO driver)

```bash
# macOS
xcode-select --install

# Ubuntu/Debian
sudo apt install build-essential
```

---

## 3. System Configuration

### 3.1 File Structure

```
.
├── .env                    # Environment variables (DO NOT commit)
├── .env.smart-home         # Smart home template
├── .env.factory            # Factory template
├── config/
│   ├── dev.yaml            # Dev config (2 devices of each type)
│   └── production.yaml     # Production config (4 BMS, 24 PCS)
```

### 3.2 YAML File Structure

The YAML file defines the **device topology** (list of BMS, PCS, Meter, EV Charger, BESS group topology). The `.env` file defines **operational parameters** (controller thresholds, simulator scales, passwords).

```yaml
# Sections in YAML file:
site:          # Site ID and name
simulate:      # true/false — simulation mode
devices:       # Device list (BMS, PCS, Meter, EV Charger)
topology:      # BESS group configuration, meter assignment
polling:       # Device polling frequency (fast/standard/slow)
engine:        # Control cycle, distribution strategy
storage:       # Database driver, DSN, retention
sync:          # MQTT cloud sync (optional)
ui:            # Dashboard port, username, password
controllers:   # Controller parameters (can be overridden by .env)
simulator:     # Simulator parameters (can be overridden by .env)
```

### 3.3 Controller Configuration in YAML

All controller parameters can be configured directly in YAML:

```yaml
controllers:
  limit_discharge:
    enabled: true
    min_soc: 15.0            # Minimum % SOC to allow discharge
    force_charge_soc: 10.0   # % SOC triggering forced charge
    force_charge_w: 2000     # Forced charge power (W)

  sell_to_grid_limit:
    enabled: true
    max_sell_to_grid_w: 0    # 0 = zero-export (no feed-in to grid)

  peak_shaving:
    enabled: true
    peak_threshold_w: 50000  # Peak shaving threshold (W)

  time_of_use:
    enabled: true
    max_charge_w: 3000       # Maximum charge power during off-peak (W)
    max_discharge_w: 5000    # Maximum discharge power during peak (W)
    min_soc_to_discharge: 20.0
    periods:
      - { start_hour: 22, end_hour: 4,  rate: off_peak }  # Off-peak
      - { start_hour: 9,  end_hour: 12, rate: on_peak }   # Morning peak
      - { start_hour: 17, end_hour: 21, rate: on_peak }   # Afternoon peak

  ev_charging:
    enabled: true
    max_site_power_w: 50000  # EV power budget (W)

  balancing:
    enabled: true
```

---

## 4. Smart Home Deployment

### 4.1 Step 1 — Prepare YAML Config

Use `config/dev.yaml` as the base, adjusting `devices` to match the actual hardware IP/port:

```yaml
# config/smart-home.yaml
site:
  id: "home-001"
  name: "Villa Alpha - District 2"

simulate: false   # false when real hardware is connected

devices:
  bms:
    - { id: bms-0, host: "192.168.1.100", port: 502, unit: 1 }
  pcs:
    - { id: pcs-0, host: "192.168.1.100", port: 502, unit: 2, brand: sungrow, command_mode: master }
  meter:
    - { id: meter-grid, host: "192.168.1.101", port: 502, unit: 1, role: grid }
  ev_charger: []

topology:
  pcc_meter: meter-grid
  bess_groups:
    - name: "home-bess"
      bms: [bms-0]
      pcs: [pcs-0]
      command_mode: master

polling:
  fast_ms: 1000
  standard_ms: 1000
  slow_ms: 5000

engine:
  cycle_ms: 1000
  distribution: equal

storage:
  driver: sqlite
  dsn: "data/ems.db"
  retention_days: 30
  wal_mode: true

sync:
  enabled: false

ui:
  port: 8080
  username: admin
  password: changeme
```

### 4.2 Step 2 — Create .env File

```bash
cp .env.smart-home .env
```

Edit `.env` for the actual site:

```bash
# .env — key variables to adjust:
EMS_SIMULATE=false                 # disable simulation
EMS_UI_PASSWORD=<strong-password>
EMS_PEAK_THRESHOLD_W=3000          # 3kW — suitable for residential
EMS_TOU_MAX_CHARGE_W=2000          # suitable for 10kWh battery
EMS_TOU_MAX_DISCHARGE_W=3000
EMS_EV_ENABLED=false               # true if EV charger is present
EMS_SIM_BATTERY_CAPACITY_WH=10000  # when running simulation test
```

### 4.3 Step 3 — Build & Run

**Run simulation first to verify:**

```bash
# Build
go build -o ems-core ./cmd/ems-core

# Run simulation (keep EMS_SIMULATE=true in .env)
./ems-core --config config/smart-home.yaml

# Open dashboard
open http://localhost:8080
```

**Run on real hardware (after connecting hardware):**

```bash
# Edit .env: EMS_SIMULATE=false
# Or override directly:
EMS_SIMULATE=false ./ems-core --config config/smart-home.yaml
```

### 4.4 Step 4 — Deploy to Raspberry Pi

```bash
# Cross-compile for ARM64
CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc \
  GOOS=linux GOARCH=arm64 \
  go build -o ems-arm64 ./cmd/ems-core

# Copy to Pi
scp ems-arm64 config/smart-home.yaml .env pi@192.168.1.50:~/ems/

# SSH into Pi
ssh pi@192.168.1.50
cd ~/ems
mkdir -p data

# Run
./ems-arm64 --config smart-home.yaml
```

**Systemd Configuration (Auto-start):**

```ini
# /etc/systemd/system/ems.service
[Unit]
Description=Local EMS
After=network.target

[Service]
Type=simple
User=pi
WorkingDirectory=/home/pi/ems
ExecStart=/home/pi/ems/ems-arm64 --config smart-home.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable ems
sudo systemctl start ems
sudo systemctl status ems
```

### 4.5 Recommended Parameters for Smart Home

| Environment Variable | Value | Description |
|----------------------|-------|-------------|
| `EMS_PEAK_THRESHOLD_W` | 3000 – 5000 | Peak threshold suitable for households |
| `EMS_TOU_MAX_CHARGE_W` | 1500 – 3000 | Does not exceed battery C-rate |
| `EMS_TOU_MAX_DISCHARGE_W` | 2000 – 5000 | Suitable for single-phase inverter |
| `EMS_FORCE_CHARGE_W` | 1000 – 2000 | Gentle forced charge |
| `EMS_EV_ENABLED` | false | Enable if 7-22kW wallbox is installed |
| `EMS_RETENTION_DAYS` | 30 | 30 days is sufficient for households |

---

## 5. Factory Deployment

### 5.1 Step 1 — Prepare YAML Config

Use `config/production.yaml` as the base. Adjust device IP/ports, PCS brands, and BESS group topology:

```yaml
# config/factory.yaml (example: 5MWh factory, 24 PCS, 4 BMS)
site:
  id: "factory-001"
  name: "Binh Duong Factory - VSIP IP"

simulate: false

devices:
  bms:
    - { id: bms-0, host: 10.0.1.10, port: 502, unit: 1 }
    - { id: bms-1, host: 10.0.1.11, port: 502, unit: 1 }
    - { id: bms-2, host: 10.0.1.12, port: 502, unit: 1 }
    - { id: bms-3, host: 10.0.1.13, port: 502, unit: 1 }
  pcs:
    # 24 PCS — see production.yaml for full reference
    - { id: pcs-0,  host: 10.0.2.10, port: 502, unit: 1, brand: sungrow, command_mode: individual }
    - { id: pcs-1,  host: 10.0.2.11, port: 502, unit: 1, brand: sungrow, command_mode: individual }
    # ... (add pcs-2 through pcs-23)
  meter:
    - { id: meter-grid,  host: 10.0.3.10, port: 502, unit: 1, role: grid, fast_ms: 200 }
    - { id: meter-solar, host: 10.0.3.11, port: 502, unit: 2, role: solar }
    - { id: meter-load,  host: 10.0.3.12, port: 502, unit: 3, role: load }
  ev_charger:
    - { id: evc-0, host: 10.0.4.10, port: 502, unit: 1, protocol: modbus }
    # ... (add evc-1 through evc-5)

topology:
  pcc_meter: meter-grid
  solar_meter: meter-solar
  load_meter: meter-load
  bess_groups:
    - name: group-0
      bms: [bms-0]
      pcs: [pcs-0, pcs-1, pcs-2, pcs-3, pcs-4, pcs-5]
      command_mode: individual
    - name: group-1
      bms: [bms-1]
      pcs: [pcs-6, pcs-7, pcs-8, pcs-9, pcs-10, pcs-11]
      command_mode: individual
    - name: group-2
      bms: [bms-2]
      pcs: [pcs-12, pcs-13, pcs-14, pcs-15, pcs-16, pcs-17]
      command_mode: individual
    - name: group-3
      bms: [bms-3]
      pcs: [pcs-18, pcs-19, pcs-20, pcs-21, pcs-22, pcs-23]
      command_mode: individual

polling:
  fast_ms: 200
  standard_ms: 1000
  slow_ms: 5000

engine:
  cycle_ms: 1000
  distribution: equal

storage:
  driver: sqlite
  dsn: "data/ems.db"
  retention_days: 90
  wal_mode: true

sync:
  enabled: false
  mqtt_broker: "ssl://cloud.example.com:8883"
  client_id: "factory-ems-001"
  topic_prefix: "ems/factory"
  interval_seconds: 300

ui:
  port: 8080
  username: admin
  password: changeme
```

### 5.2 Step 2 — Create .env File

```bash
cp .env.factory .env
```

Edit `.env` according to factory parameters:

```bash
# .env — key variables to adjust:
EMS_SIMULATE=false
EMS_UI_PASSWORD=<strong-password>
EMS_PEAK_THRESHOLD_W=250000        # 250kW — below 300kW contract
EMS_TOU_MAX_CHARGE_W=500000        # 500kW
EMS_TOU_MAX_DISCHARGE_W=500000     # 500kW
EMS_FORCE_CHARGE_W=50000           # 50kW force charge
EMS_EV_ENABLED=true
EMS_EV_MAX_SITE_POWER_W=150000     # 150kW EV budget
EMS_RETENTION_DAYS=90
```

### 5.3 Step 3 — Build & Run

```bash
# Build
go build -o ems-core ./cmd/ems-core

# Test with simulation first
EMS_SIMULATE=true ./ems-core --config config/factory.yaml

# Check dashboard
open http://localhost:8080

# Run in production (after verification)
EMS_SIMULATE=false ./ems-core --config config/factory.yaml
```

### 5.4 Step 4 — Configure systemd

```ini
# /etc/systemd/system/ems.service
[Unit]
Description=Local EMS - Factory
After=network.target

[Service]
Type=simple
User=ems
WorkingDirectory=/opt/ems
ExecStart=/opt/ems/ems-core --config /opt/ems/config/factory.yaml
Restart=always
RestartSec=3
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

```bash
sudo useradd -r -s /sbin/nologin ems
sudo mkdir -p /opt/ems/data
sudo cp ems-core config/factory.yaml .env /opt/ems/
sudo chown -R ems:ems /opt/ems

sudo systemctl enable ems
sudo systemctl start ems
journalctl -u ems -f   # monitor logs
```

### 5.5 Recommended Parameters for Factory

| Environment Variable | Value | Description |
|----------------------|-------|-------------|
| `EMS_PEAK_THRESHOLD_W` | 200000 – 400000 | Set to 80-85% of utility contract capacity |
| `EMS_TOU_MAX_CHARGE_W` | 300000 – 500000 | Based on total PCS capacity |
| `EMS_TOU_MAX_DISCHARGE_W` | 300000 – 500000 | Based on total PCS capacity |
| `EMS_FORCE_CHARGE_W` | 30000 – 100000 | 5-10% of battery capacity |
| `EMS_EV_MAX_SITE_POWER_W` | 100000 – 200000 | EV power budget at PCC |
| `EMS_MIN_SOC` | 10.0 – 20.0 | Minimum SOC before blocking discharge |
| `EMS_RETENTION_DAYS` | 90 | 3 months for analysis & audit |

### 5.6 Industrial EVN Tariff (Reference)

Configure `time_of_use.periods` in YAML according to EVN tariffs:

| Time Window | Type | Multiplier | Config rate |
|-------------|------|------------|-------------|
| 22:00 – 04:00 | Off-peak | 0.5x | `off_peak` |
| 04:00 – 09:30 | Normal | 1.0x | `mid_peak` |
| 09:30 – 11:30 | Peak | 1.7x | `on_peak` |
| 11:30 – 17:00 | Normal | 1.0x | `mid_peak` |
| 17:00 – 20:00 | Peak | 1.7x | `on_peak` |
| 20:00 – 22:00 | Normal | 1.0x | `mid_peak` |

```yaml
controllers:
  time_of_use:
    periods:
      - { start_hour: 22, end_hour: 4,  rate: off_peak }
      - { start_hour: 9,  end_hour: 11, rate: on_peak }
      - { start_hour: 17, end_hour: 20, rate: on_peak }
```

> **Note:** Period configuration is hour-accurate (minutes not yet supported). Set 9h instead of 9:30 for safety.

---

## 6. Environment Variables

### 6.1 Complete List

#### Core

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `EMS_ENV_FILE` | string | `.env` | Custom .env file path |
| `EMS_SIMULATE` | bool | `false` | Hardware simulation mode |
| `EMS_UI_PORT` | int | `8080` | Web dashboard port |
| `EMS_UI_USERNAME` | string | `admin` | Login username |
| `EMS_UI_PASSWORD` | string | `ems@2025` | Login password |
| `EMS_DB_DSN` | string | `/tmp/ems-dev.db` | Database file path |
| `EMS_RETENTION_DAYS` | int | `7` | Data retention days |

#### Controller — Battery Protection

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `EMS_MIN_SOC` | float | `15.0` | Minimum SOC — blocks discharge below this level |
| `EMS_FORCE_CHARGE_SOC` | float | `10.0` | SOC triggering forced charge |
| `EMS_FORCE_CHARGE_W` | int | `2000` | Forced charge power (W) |

#### Controller — Grid

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `EMS_MAX_SELL_TO_GRID_W` | int | `0` | Max feed-in to grid (0 = zero-export) |

#### Controller — Peak Shaving

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `EMS_PEAK_THRESHOLD_W` | int | `50000` | Peak shaving threshold (W) |
| `EMS_PEAK_SHAVING_ENABLED` | bool | `true` | Enable/disable peak shaving |

#### Controller — Time-of-Use

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `EMS_TOU_MAX_CHARGE_W` | int | `3000` | Off-peak charge power (W) |
| `EMS_TOU_MAX_DISCHARGE_W` | int | `5000` | Peak discharge power (W) |
| `EMS_TOU_MIN_SOC` | float | `20.0` | Minimum SOC for TOU discharge |
| `EMS_TOU_ENABLED` | bool | `true` | Enable/disable TOU |

#### Controller — EV Charging

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `EMS_EV_MAX_SITE_POWER_W` | int | `50000` | EV power budget at PCC (W) |
| `EMS_EV_ENABLED` | bool | `auto` | Auto-enabled if EV charger exists in YAML |
| `EMS_BALANCING_ENABLED` | bool | `true` | Enable/disable self-consumption balancing |

#### Simulator

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `EMS_SIM_INITIAL_SOC` | float | `50.0` | Initial SOC (%) |
| `EMS_SIM_SOLAR_PEAK_W` | int | `10000` | Solar peak power (W) |
| `EMS_SIM_LOAD_BASE_W` | int | `5000` | Base load (W) |
| `EMS_SIM_LOAD_VARIANCE_W` | int | `3000` | Load variance ±W |
| `EMS_SIM_BATTERY_CAPACITY_WH` | int | `5000000` | Battery capacity (Wh) |
| `EMS_SIM_MAX_CHARGE_W` | int | `50000` | Max charge rate (W) |
| `EMS_SIM_MAX_DISCHARGE_W` | int | `50000` | Max discharge rate (W) |

### 6.2 Usage Methods

**Method 1 — .env File (Recommended):**

```bash
cp .env.smart-home .env    # or .env.factory
vi .env                    # edit
./ems-core --config config/dev.yaml
```

**Method 2 — Specify Custom .env File:**

```bash
EMS_ENV_FILE=/opt/ems/production.env ./ems-core --config config/factory.yaml
```

**Method 3 — Inline Override (Quick Test):**

```bash
EMS_PEAK_THRESHOLD_W=100000 EMS_TOU_MAX_CHARGE_W=200000 ./ems-core --simulate
```

**Method 4 — Docker:**

```bash
docker run -d \
  -e EMS_SIMULATE=true \
  -e EMS_PEAK_THRESHOLD_W=3000 \
  -e EMS_UI_PASSWORD=secret \
  -v ./config:/app/config \
  local-ems --config /app/config/smart-home.yaml
```

---

## 7. System Verification

### 7.1 Run Tests

```bash
# Unit tests (all controllers + config)
go test ./internal/... -count=1

# Config tests only (verify .env loading)
go test ./internal/config/... -v -count=1
```

### 7.2 Simulation Verification

Run simulation and verify on the console:

```bash
EMS_SIMULATE=true ./ems-core --config config/smart-home.yaml
```

Expected output (every 5 seconds):

```
[14:30:05] ☀ 3200W | ⚡ 1800W | 🔌  -400W | 🔋 65.2% | ESS: -1400W | BMS:1 PCS:1 | -
```

| Column | Meaning | Expected / Valid Value |
|--------|---------|------------------------|
| ☀ | Solar power | > 0 during daytime |
| ⚡ | Load power | Always > 0 |
| 🔌 | Grid power | ≥ 0 if zero-export is active |
| 🔋 | Battery SOC | 10% – 100% |
| ESS | Setpoint | Negative = charge, Positive = discharge |

### 7.3 Verification Checklist

- [ ] Dashboard is accessible at `http://<ip>:8080`
- [ ] Login succeeds with username/password from `.env`
- [ ] Real-time power flow chart displays properly
- [ ] SOC does not drop below `EMS_MIN_SOC`
- [ ] Grid power ≥ 0 (zero-export working)
- [ ] Peak shaving activates when load exceeds threshold
- [ ] TOU charges battery during off-peak hours (22:00-04:00)
- [ ] TOU discharges battery during peak hours (09:00-12:00, 17:00-21:00)

---

## 8. Operations & Maintenance

### 8.1 Log Management

```bash
# View real-time logs (systemd)
journalctl -u ems -f

# View logs from the past hour
journalctl -u ems --since "1 hour ago"
```

### 8.2 Database

```bash
# Check disk usage
du -sh /opt/ems/data/ems.db

# Manual backup (SQLite safe copy)
sqlite3 /opt/ems/data/ems.db ".backup /backup/ems-$(date +%Y%m%d).db"
```

Data is automatically purged after `EMS_RETENTION_DAYS` days.

### 8.3 Runtime Configuration Changes

Edit `.env` file and restart the service:

```bash
vi /opt/ems/.env
sudo systemctl restart ems
```

Or temporarily override (without editing file):

```bash
sudo systemctl stop ems
EMS_PEAK_THRESHOLD_W=200000 /opt/ems/ems-core --config /opt/ems/config/factory.yaml
```

### 8.4 Software Updates

```bash
# Build new version
go build -o ems-core ./cmd/ems-core

# Deploy
sudo systemctl stop ems
sudo cp ems-core /opt/ems/ems-core
sudo systemctl start ems
```

### 8.5 Troubleshooting

| Symptom | Cause | Resolution |
|---------|-------|------------|
| Dashboard not accessible | EMS is not running or port is blocked | Run `systemctl status ems`, check firewall rules |
| SOC drops to 0% | `EMS_MIN_SOC` set too low or controller is disabled | Check `.env`, verify `EMS_MIN_SOC=15` |
| Negative grid power (grid export) | `sell_to_grid_limit` is disabled | Verify `EMS_MAX_SELL_TO_GRID_W=0` |
| Database size too large | Retention period is too long | Reduce `EMS_RETENTION_DAYS` |
| Modbus timeout | Network issue or device fault | Check Ethernet cables, ping host |
