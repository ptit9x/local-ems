# Local EMS — Deployment Guide

## Table of Contents

- [System Requirements](#system-requirements)
- [Quick Start](#quick-start)
- [Linux Installation](#linux-installation)
- [Windows Installation](#windows-installation)
- [Configuration](#configuration)
- [Verification](#verification)
- [Management Commands](#management-commands)
- [Uninstallation](#uninstallation)
- [Troubleshooting](#troubleshooting)

---

## System Requirements

### Smart Home (Residential)

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| **Device** | Raspberry Pi 4 (2GB) | Raspberry Pi 5 (4GB) |
| **OS** | Linux ARM64 (Debian 11+) | Ubuntu 22.04 LTS / Raspbian |
| **CPU** | ARMv8 / 1 GHz | Quad-core / 1.5 GHz |
| **RAM** | 512 MB | 2 GB |
| **Storage** | 4 GB SD card | 32 GB SSD |
| **Network** | 1× Ethernet (to inverter) | 1× Ethernet + Wi-Fi |
| **Power** | 5V / 3A USB-C | UPS recommended |

### Factory (Industrial)

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| **Device** | Industrial PC / Mini PC | Rack-mount server |
| **OS** | Linux x86_64 (Debian 11+) or Windows 10/11 | Ubuntu 22.04 LTS Server |
| **CPU** | x86_64 / 2 cores / 1.5 GHz | 4 cores / 2+ GHz |
| **RAM** | 1 GB | 4 GB |
| **Storage** | 8 GB | 64 GB SSD |
| **Network** | 1× Ethernet (to Modbus devices) | 2× Ethernet (device LAN + WAN) |
| **Power** | Standard AC | UPS required |

### Network Requirements

- All BMS, PCS, Meter, and EV Charger devices must be on the **same LAN** as the EMS machine
- Modbus TCP port **502** must be reachable from EMS to each device
- Dashboard port **8080** (configurable) must be accessible from the operator's browser
- **No internet required** — EMS runs fully offline

## Quick Start

You have received a `.zip` file (e.g. `ems-smart-home-linux-amd64.zip`). The package contains:

```
ems-core          ← EMS application
.env              ← runtime parameters (editable)
config.yaml       ← device topology
install.sh        ← Linux installer
install.bat       ← Windows installer
uninstall.sh      ← Linux uninstaller
uninstall.bat     ← Windows uninstaller
data/             ← database (auto-created)
```

---

## Linux Installation

```bash
unzip ems-smart-home-linux-amd64.zip
cd ems-smart-home/
sudo bash install.sh
```

The installer will:
1. Copy files to `/opt/ems/`
2. Create a system service that starts on boot
3. Start EMS and print the dashboard URL

---

## Windows Installation

1. Extract the zip file
2. Right-click **install.bat** → **Run as Administrator**

The installer will:
1. Copy files to `C:\ems\`
2. Create a Windows service that starts on boot
3. Add a dashboard shortcut to the Desktop

---

## Configuration

Edit `.env` **before running the installer** to match your site.

### Key Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `EMS_UI_PORT` | `8080` | Dashboard port |
| `EMS_UI_PASSWORD` | `ems@2025` | Dashboard login password |
| `EMS_PEAK_THRESHOLD_W` | `50000` | Peak shaving threshold (Watts) |
| `EMS_TOU_MAX_CHARGE_W` | `3000` | Max charge power during off-peak (Watts) |
| `EMS_TOU_MAX_DISCHARGE_W` | `5000` | Max discharge power during on-peak (Watts) |
| `EMS_MIN_SOC` | `15.0` | Minimum battery SOC before blocking discharge (%) |
| `EMS_EV_ENABLED` | `auto` | Enable EV charging management (`true`/`false`) |
| `EMS_EV_MAX_SITE_POWER_W` | `50000` | Max power budget for EV chargers (Watts) |
| `EMS_RETENTION_DAYS` | `7` | Days of data to keep |

### Typical Values

| Parameter | Smart Home | Factory |
|-----------|-----------|---------|
| `EMS_PEAK_THRESHOLD_W` | `3000` | `250000` |
| `EMS_TOU_MAX_CHARGE_W` | `2000` | `500000` |
| `EMS_TOU_MAX_DISCHARGE_W` | `3000` | `500000` |
| `EMS_EV_ENABLED` | `false` | `true` |
| `EMS_RETENTION_DAYS` | `30` | `90` |

To change parameters after installation, edit `/opt/ems/.env` (Linux) or `C:\ems\.env` (Windows), then restart the service.

---

## Verification

1. Open a browser: `http://<machine-ip>:8080`
2. Login with the username/password from `.env`
3. Verify the dashboard shows real-time data

### Checklist

- [ ] Dashboard is accessible
- [ ] Battery SOC is displayed and stays above `EMS_MIN_SOC`
- [ ] Grid power stays ≥ 0 (zero-export is working)
- [ ] Solar and load readings update every few seconds

---

## Management Commands

### Linux

```bash
sudo systemctl status ems       # check status
sudo systemctl restart ems      # restart (after config change)
sudo systemctl stop ems         # stop
sudo journalctl -u ems -f       # view live logs
sudo journalctl -u ems --since "1 hour ago"   # recent logs
```

### Windows

```cmd
sc query ems       &REM check status
sc stop ems        &REM stop
sc start ems       &REM start
```

Or double-click `C:\ems\Start EMS.bat` to run manually.

---

## Uninstallation

### Linux

```bash
sudo bash /opt/ems/uninstall.sh
```

### Windows

Right-click `uninstall.bat` → **Run as Administrator**

Both uninstallers ask whether to keep or delete historical data.

---

## Troubleshooting

| Problem | Cause | Fix |
|---------|-------|-----|
| Can't access dashboard | Service not running or port blocked | Check service status, check firewall |
| SOC drops to 0% | `EMS_MIN_SOC` too low or controller disabled | Set `EMS_MIN_SOC=15` in `.env`, restart |
| Grid export (selling power) | Zero-export not active | Set `EMS_MAX_SELL_TO_GRID_W=0` in `.env`, restart |
| Database too large | Retention too long | Reduce `EMS_RETENTION_DAYS`, restart |
| Modbus timeout | Network or device issue | Check cables, verify device IP with `ping` |
