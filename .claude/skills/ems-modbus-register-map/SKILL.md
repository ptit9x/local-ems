---
name: ems-modbus-register-map
description: "Use when creating a new PCS register map profile for the Local EMS project. Also use when adding support for a new PCS brand, creating a register profile, or mapping Modbus registers for a new inverter."
metadata:
  category: technique
  author: richard
  triggers: register map, pcs profile, new brand, modbus register, inverter profile,
    sungrow, sinexcel, huawei, goodwe, pcs brand
risk: low
source: local
---

# EMS Modbus Register Map Generator

Generates PCS (Power Conversion System) register map profiles for new inverter brands, using the `PCSProfileRegistry` system.

## When to Use

- Adding support for a new PCS/inverter brand (Huawei, Goodwe, SMA, etc.)
- Creating a Modbus register mapping from a vendor datasheet
- User says "add PCS brand", "register map for X", "support new inverter"

## Architecture Context

### PCSProfile System

The project uses a [PCSProfileRegistry](file:///Users/richard/vmo_bid_cdi/internal/devices/pcs_profile.go) to abstract PCS register differences across brands:

```go
type PCSProfile struct {
    Brand             string           // Unique brand identifier
    ActivePowerSetReg uint16           // Register to write active power setpoint
    ActivePowerActReg uint16           // Register to read actual active power
    ReactivePowerReg  uint16           // Register to read reactive power
    StatusReg         uint16           // Register to read device status
    RegisterCount     uint16           // Total registers to read in one batch
    ScaleFactor       float64          // Multiply raw value by this (e.g., 10.0 for 0.1kW units)
    StatusMap         map[uint16]PCSStatus // Map vendor status codes to internal PCSStatus
    CommandMode       string           // "master" or "individual"
    Description       string           // Human-readable description
}
```

### Existing Profiles

| Brand | SetReg | ActReg | ReactiveReg | StatusReg | Scale | Mode |
|-------|--------|--------|-------------|-----------|-------|------|
| `generic` | 0 | 1 | 2 | 3 | 1.0 | master |
| `sungrow` | 5040 | 5008 | 5010 | 5000 | 1.0 | master |
| `sinexcel` | 100 | 102 | 104 | 110 | 10.0 | master |

### How Profiles Are Used

```
ProfiledPCSAdapter
  ├── Poll()
  │   ├── ReadHoldingRegisters(statusReg, 1) → PCSStatus
  │   ├── ReadHoldingRegisters(activePowerActReg, 1) → int * scaleFactor
  │   └── ReadHoldingRegisters(reactivePowerReg, 1) → int * scaleFactor
  └── WriteActivePower(watts)
      └── WriteMultipleRegisters(activePowerSetReg, watts / scaleFactor)
```

## Generation Steps

### Step 1: Gather Register Information from Vendor Datasheet

You need these registers at minimum:

| Required Info | Description | Example |
|---------------|-------------|---------|
| Active Power Setpoint | Write register for charge/discharge command | 5040 |
| Actual Active Power | Read register for current output | 5008 |
| Reactive Power | Read register for reactive output | 5010 |
| Status | Read register for running/fault state | 5000 |
| Scale Factor | Unit conversion (1.0 = W, 10.0 = 0.1kW, 100.0 = 0.01kW) | 1.0 |
| Status Codes | Vendor status values → PCSStatus mapping | 0→Stopped, 1→Running, 2→Fault |

### Step 2: Add Profile to Registry

In [pcs_profile.go](file:///Users/richard/vmo_bid_cdi/internal/devices/pcs_profile.go), add the new profile in `DefaultPCSProfileRegistry()`:

```go
r.Register(PCSProfile{
    Brand:             "<brand_name>",  // lowercase, no spaces
    ActivePowerSetReg: <register>,      // uint16: write setpoint register
    ActivePowerActReg: <register>,      // uint16: read actual power register
    ReactivePowerReg:  <register>,      // uint16: read reactive power register
    StatusReg:         <register>,      // uint16: read status register
    RegisterCount:     <count>,         // uint16: total batch read size
    ScaleFactor:       <factor>,        // float64: raw * factor = watts
    StatusMap: map[uint16]PCSStatus{
        <vendor_code>: PCSStopped,   // Map vendor codes to standard status
        <vendor_code>: PCSRunning,
        <vendor_code>: PCSFault,
        // Add vendor-specific states as needed
    },
    CommandMode: "master",  // or "individual" for multi-PCS setups
    Description: "<Brand Name> PCS profile — <model or series>",
})
```

### Step 3: Add Unit Tests

In [pcs_profile_test.go](file:///Users/richard/vmo_bid_cdi/internal/devices/pcs_profile_test.go), add tests:

```go
func TestProfile_<Brand>_Registered(t *testing.T) {
    registry := DefaultPCSProfileRegistry()
    profile, err := registry.Get("<brand_name>")
    if err != nil {
        t.Fatalf("profile not found: %v", err)
    }
    if profile.Brand != "<brand_name>" {
        t.Errorf("expected brand '<brand_name>', got %q", profile.Brand)
    }
}

func TestProfile_<Brand>_ScaleFactor(t *testing.T) {
    registry := DefaultPCSProfileRegistry()
    profile, _ := registry.Get("<brand_name>")
    if profile.ScaleFactor <= 0 {
        t.Errorf("scale factor must be positive, got %f", profile.ScaleFactor)
    }
}

func TestProfile_<Brand>_StatusMap(t *testing.T) {
    registry := DefaultPCSProfileRegistry()
    profile, _ := registry.Get("<brand_name>")

    // Verify essential statuses are mapped
    requiredStatuses := []uint16{<stopped_code>, <running_code>, <fault_code>}
    for _, code := range requiredStatuses {
        if _, ok := profile.StatusMap[code]; !ok {
            t.Errorf("status code %d not mapped", code)
        }
    }
}

func TestProfile_<Brand>_RegistersNonOverlapping(t *testing.T) {
    registry := DefaultPCSProfileRegistry()
    profile, _ := registry.Get("<brand_name>")

    // Verify setpoint and read registers are different
    if profile.ActivePowerSetReg == profile.ActivePowerActReg {
        t.Error("setpoint and actual registers must be different")
    }
}
```

### Step 4: Update YAML Config

In `config/dev.yaml` and `config/production.yaml`, allow the new brand:

```yaml
devices:
  pcs:
    - id: pcs-0
      host: 192.168.1.100
      port: 502
      unit: 1
      brand: <brand_name>    # ← new brand
      command_mode: master
```

### Step 5: Add Simulator Support (optional)

If you need to simulate this brand, add registers to the fake Modbus server in `internal/simulator/simulator.go`.

### Step 6: Run Tests

```bash
# Profile tests
go test ./internal/devices/... -run TestProfile -v

# All tests
go test ./... -v
```

## Common Brand Reference

Use these as starting points when mapping real datasheets:

### Huawei SUN2000 (example)
```go
r.Register(PCSProfile{
    Brand:             "huawei",
    ActivePowerSetReg: 40429,
    ActivePowerActReg: 32080,
    ReactivePowerReg:  32082,
    StatusReg:         32089,
    RegisterCount:     50,
    ScaleFactor:       1000.0, // kW → W
    StatusMap: map[uint16]PCSStatus{
        0:     PCSStopped,
        1:     PCSRunning,
        2:     PCSFault,
        0x100: PCSStopped, // Standby
    },
    CommandMode: "master",
    Description: "Huawei SUN2000 series ESS inverter",
})
```

### Goodwe ET Series (example)
```go
r.Register(PCSProfile{
    Brand:             "goodwe",
    ActivePowerSetReg: 45200,
    ActivePowerActReg: 35107,
    ReactivePowerReg:  35108,
    StatusReg:         35100,
    RegisterCount:     15,
    ScaleFactor:       1.0,
    StatusMap: map[uint16]PCSStatus{
        0: PCSStopped,
        1: PCSRunning,
        3: PCSFault,
    },
    CommandMode: "individual",
    Description: "Goodwe ET series hybrid inverter",
})
```

## Checklist

- [ ] Register addresses verified against vendor datasheet
- [ ] Scale factor correct (raw register value * factor = watts)
- [ ] Status codes mapped (at least: Stopped, Running, Fault)
- [ ] SetReg ≠ ActReg (write vs read registers are different)
- [ ] Profile registered in `DefaultPCSProfileRegistry()`
- [ ] Unit tests verify: registration, scale factor, status map
- [ ] YAML config updated with new brand option
- [ ] All tests pass: `go test ./...`

> [!WARNING]
> **Always verify registers against the actual vendor Modbus documentation.**
> Register addresses, scale factors, and status codes vary between firmware versions.
> Never assume a profile from one model works for another — even within the same brand.
