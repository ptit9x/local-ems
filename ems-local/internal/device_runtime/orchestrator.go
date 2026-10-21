// Package device_runtime connects the protocol-neutral DeviceRegistry to the
// acquisition aggregator and command paths used by the EMS control loop.
package device_runtime

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/vmo/ems-platform/ems-local/internal/config"
	"github.com/vmo/ems-platform/ems-local/internal/devices"
	"github.com/vmo/ems-platform/ems-local/internal/devices/aggregator"
)

type Orchestrator struct {
	registry   *devices.DeviceRegistry
	aggregator *aggregator.Aggregator
	indexes    map[string]deviceIndex
	canControl func() bool
	retries    int
	retryDelay time.Duration
	mu         sync.Mutex
	lastEV     map[string]int
}

type deviceIndex struct {
	typ   devices.DeviceType
	index int
	role  string
}

func New(registry *devices.DeviceRegistry, agg *aggregator.Aggregator, cfg *config.Config, canControl func() bool) *Orchestrator {
	if canControl == nil {
		canControl = func() bool { return true }
	}
	o := &Orchestrator{registry: registry, aggregator: agg, indexes: make(map[string]deviceIndex), canControl: canControl, retries: 3, retryDelay: 100 * time.Millisecond, lastEV: make(map[string]int)}
	for i, d := range cfg.Devices.BMS {
		o.indexes[d.ID] = deviceIndex{typ: devices.DeviceTypeBMS, index: i}
	}
	for i, d := range cfg.Devices.PCS {
		o.indexes[d.ID] = deviceIndex{typ: devices.DeviceTypePCS, index: i}
	}
	for i, d := range cfg.Devices.Meter {
		o.indexes[d.ID] = deviceIndex{typ: devices.DeviceTypeMeter, index: i, role: d.Role}
	}
	for i, d := range cfg.Devices.EVCharger {
		o.indexes[d.ID] = deviceIndex{typ: devices.DeviceTypeEVCharger, index: i}
	}
	return o
}

func (o *Orchestrator) SetRetryPolicy(attempts int, delay time.Duration) {
	if attempts > 0 {
		o.retries = attempts
	}
	if delay >= 0 {
		o.retryDelay = delay
	}
}

func (o *Orchestrator) Start(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Second
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		o.PollOnce()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				o.PollOnce()
			}
		}
	}()
}

func (o *Orchestrator) PollOnce() {
	for _, dev := range o.registry.All() {
		if err := dev.Poll(); err != nil {
			log.Printf("device-runtime: poll %s: %v", dev.ID(), err)
		}
		snapshot, ok := dev.(devices.SnapshotDevice)
		if !ok {
			continue
		}
		o.applySnapshot(snapshot)
	}
}

func (o *Orchestrator) applySnapshot(dev devices.SnapshotDevice) {
	idx, ok := o.indexes[dev.ID()]
	if !ok {
		return
	}
	p := dev.Snapshot()
	now := time.Now()
	online := dev.Status() == devices.DeviceStatusOnline
	switch idx.typ {
	case devices.DeviceTypeBMS:
		o.aggregator.UpdateBMS(idx.index, aggregator.BMSState{ID: dev.ID(), SOC: f64(p, "soc"), TempC: f64(p, "temperature_c"), MinCellMV: integer(p, "cell_voltage_min_mv"), MaxCellMV: integer(p, "cell_voltage_max_mv"), CapacityWh: integerDefault(p, "capacity_wh", 1), MaxChargeW: integer(p, "max_charge_w"), MaxDischargeW: integer(p, "max_discharge_w"), Online: online, UpdatedAt: now})
	case devices.DeviceTypePCS:
		status := aggregator.PCSStandby
		if !online {
			status = aggregator.PCSOffline
		} else if stringValue(p, "status") == "Running" {
			status = aggregator.PCSRunning
		}
		o.aggregator.UpdatePCS(idx.index, aggregator.PCSState{ID: dev.ID(), ActivePowerW: integer(p, "active_power_w"), SetpointW: integer(p, "active_power_setpoint_w"), MaxPowerW: integer(p, "max_power_w"), Status: status, Online: online, UpdatedAt: now})
	case devices.DeviceTypeMeter:
		o.aggregator.UpdateMeter(idx.index, aggregator.MeterState{ID: dev.ID(), Role: idx.role, ActivePowerW: integer(p, "active_power_w"), VoltageV: f64(p, "voltage_l1_v"), CurrentA: f64(p, "current_l1_a"), FrequencyHz: f64(p, "frequency_hz"), Online: online, UpdatedAt: now})
	case devices.DeviceTypeEVCharger:
		o.aggregator.UpdateEVCharger(idx.index, aggregator.EVChargerState{ID: dev.ID(), Status: integer(p, "status"), ActivePowerW: integer(p, "active_power_w"), EnergyDelivered: integer(p, "energy_delivered_wh"), MaxPowerLimit: integer(p, "max_power_limit_w"), VehicleConnected: boolean(p, "vehicle_connected"), CurrentL1: f64(p, "current_l1_a"), CurrentL2: f64(p, "current_l2_a"), CurrentL3: f64(p, "current_l3_a"), Online: online, UpdatedAt: now})
	}
}

func (o *Orchestrator) ApplyPCSSetpoints(setpoints map[string]int) error {
	if !o.canControl() {
		return fmt.Errorf("device-runtime: local control lease is not active")
	}
	for id, watts := range setpoints {
		if err := o.writeWithRetry(id, devices.Command{Name: "set_active_power_w", Params: map[string]interface{}{"watts": watts}}); err != nil {
			return err
		}
	}
	return nil
}

func (o *Orchestrator) ApplyEVLimits(limits map[string]int) error {
	if !o.canControl() {
		return fmt.Errorf("device-runtime: local control lease is not active")
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	for id, watts := range limits {
		if previous, ok := o.lastEV[id]; ok && previous == watts {
			continue
		}
		if err := o.writeWithRetry(id, devices.Command{Name: "set_max_power_w", Params: map[string]interface{}{"watts": watts}}); err != nil {
			return err
		}
		o.lastEV[id] = watts
	}
	return nil
}

func (o *Orchestrator) writeWithRetry(id string, command devices.Command) error {
	dev, ok := o.registry.Get(id)
	if !ok {
		return fmt.Errorf("device-runtime: device %s not found", id)
	}
	control, ok := dev.(devices.ControllableDevice)
	if !ok {
		return fmt.Errorf("device-runtime: device %s is not controllable", id)
	}
	var err error
	for attempt := 1; attempt <= o.retries; attempt++ {
		if err = control.Write(command); err == nil {
			return nil
		}
		if attempt < o.retries && o.retryDelay > 0 {
			time.Sleep(o.retryDelay)
		}
	}
	return fmt.Errorf("device-runtime: %s command %s failed after %d attempts: %w", id, command.Name, o.retries, err)
}

func raw(p devices.DataPoints, key string) interface{} {
	if v, ok := p[key]; ok {
		return v.Value
	}
	return nil
}
func f64(p devices.DataPoints, key string) float64 {
	switch v := raw(p, key).(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	}
	return 0
}
func integer(p devices.DataPoints, key string) int { return int(math.Round(f64(p, key))) }
func integerDefault(p devices.DataPoints, key string, fallback int) int {
	v := integer(p, key)
	if v == 0 {
		return fallback
	}
	return v
}
func boolean(p devices.DataPoints, key string) bool       { v, _ := raw(p, key).(bool); return v }
func stringValue(p devices.DataPoints, key string) string { v, _ := raw(p, key).(string); return v }
