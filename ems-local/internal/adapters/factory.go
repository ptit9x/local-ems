// Package adapters constructs protocol/vendor-specific device implementations
// behind the normalized devices interfaces.
package adapters

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vmo/ems-platform/ems-local/internal/config"
	"github.com/vmo/ems-platform/ems-local/internal/devices"
	"github.com/vmo/ems-platform/ems-local/internal/devices/drivers"
	"github.com/vmo/ems-platform/ems-local/internal/modbus"
	"github.com/vmo/ems-platform/ems-local/internal/ocpp"
)

// ModbusConnector is injectable so construction and failure behavior can be
// tested without opening sockets.
type ModbusConnector func(*modbus.Client) error

// Factory owns construction policy. The caller owns the returned registry and
// must call Close during shutdown.
type Factory struct {
	OCPP            *ocpp.CentralSystem
	Timeout         time.Duration
	ConnectOnCreate bool
	ConnectModbus   ModbusConnector
}

func NewFactory(central *ocpp.CentralSystem) *Factory {
	return &Factory{OCPP: central, Timeout: 3 * time.Second, ConnectOnCreate: true, ConnectModbus: func(client *modbus.Client) error { return client.Connect() }}
}

// Build creates and registers every configured site device. Construction is
// transactional: a failure closes all adapters already created in this call.
func (f *Factory) Build(cfg *config.Config) (*devices.DeviceRegistry, error) {
	registry := devices.NewDeviceRegistry()
	register := func(adapter devices.Device) error {
		if err := registry.Register(adapter); err != nil {
			return err
		}
		return nil
	}
	fail := func(err error) (*devices.DeviceRegistry, error) { registry.Close(); return nil, err }

	for _, item := range cfg.Devices.Meter {
		adapter, err := f.Create(devices.DeviceConfig{ID: item.ID, Type: devices.DeviceTypeMeter, Host: item.Host, Port: item.Port, UnitID: item.Unit, Protocol: "modbus", Brand: "generic", Extra: map[string]interface{}{"role": item.Role}})
		if err != nil {
			return fail(err)
		}
		if err := register(adapter); err != nil {
			return fail(err)
		}
	}
	for _, item := range cfg.Devices.BMS {
		adapter, err := f.Create(devices.DeviceConfig{ID: item.ID, Type: devices.DeviceTypeBMS, Host: item.Host, Port: item.Port, UnitID: item.Unit, Protocol: "modbus", Brand: "generic"})
		if err != nil {
			return fail(err)
		}
		if err := register(adapter); err != nil {
			return fail(err)
		}
	}
	for _, item := range cfg.Devices.PCS {
		adapter, err := f.Create(devices.DeviceConfig{ID: item.ID, Type: devices.DeviceTypePCS, Host: item.Host, Port: item.Port, UnitID: item.Unit, Protocol: "modbus", Brand: item.Brand, Extra: map[string]interface{}{"command_mode": item.CommandMode}})
		if err != nil {
			return fail(err)
		}
		if err := register(adapter); err != nil {
			return fail(err)
		}
	}
	for _, item := range cfg.Devices.EVCharger {
		adapter, err := f.Create(devices.DeviceConfig{ID: item.ID, Type: devices.DeviceTypeEVCharger, Host: item.Host, Port: item.Port, UnitID: item.Unit, Protocol: item.Protocol, Brand: "generic"})
		if err != nil {
			return fail(err)
		}
		if err := register(adapter); err != nil {
			return fail(err)
		}
	}
	return registry, nil
}

// Create selects an adapter by device type, protocol, and vendor profile.
func (f *Factory) Create(cfg devices.DeviceConfig) (devices.SnapshotDevice, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}
	protocol := strings.ToLower(cfg.Protocol)
	if protocol == "modbus-tcp" {
		protocol = "modbus"
	}
	if cfg.Type == devices.DeviceTypeEVCharger && protocol == "ocpp" {
		return f.createOCPP(cfg)
	}
	if protocol != "modbus" {
		return nil, fmt.Errorf("adapter factory: unsupported protocol %q for %s", cfg.Protocol, cfg.ID)
	}
	return f.createModbus(cfg)
}

func (f *Factory) createModbus(cfg devices.DeviceConfig) (devices.SnapshotDevice, error) {
	timeout := f.Timeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	client := modbus.NewClient(net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)), timeout)
	if f.ConnectOnCreate {
		connector := f.ConnectModbus
		if connector == nil {
			connector = func(c *modbus.Client) error { return c.Connect() }
		}
		if err := connector(client); err != nil {
			return nil, fmt.Errorf("adapter factory: connect %s: %w", cfg.ID, err)
		}
	}
	closeOnError := func(err error) (devices.SnapshotDevice, error) { _ = client.Close(); return nil, err }

	base := newManaged(cfg, client.Close)
	switch cfg.Type {
	case devices.DeviceTypeMeter:
		role, _ := cfg.Extra["role"].(string)
		adapter := devices.NewMeterAdapter(client, byte(cfg.UnitID), cfg.ID)
		base.poll = func() (devices.DataPoints, error) {
			value, err := adapter.Read()
			if err != nil {
				return nil, err
			}
			return points(map[string]interface{}{"role": role, "active_power_w": value.ActivePowerW, "voltage_l1_v": value.VoltageL1V, "voltage_l2_v": value.VoltageL2V, "voltage_l3_v": value.VoltageL3V, "current_l1_a": value.CurrentL1A, "current_l2_a": value.CurrentL2A, "current_l3_a": value.CurrentL3A, "frequency_hz": value.FrequencyHz}), nil
		}
	case devices.DeviceTypeBMS:
		adapter := devices.NewBMSAdapter(client, byte(cfg.UnitID))
		base.poll = func() (devices.DataPoints, error) {
			value, err := adapter.Read()
			if err != nil {
				return nil, err
			}
			return points(map[string]interface{}{"soc": value.SOC, "temperature_c": value.TemperatureC, "cell_voltage_min_mv": value.CellVoltageMinMV, "cell_voltage_max_mv": value.CellVoltageMaxMV, "status": value.Status.String(), "cycle_count": value.CycleCount}), nil
		}
	case devices.DeviceTypePCS:
		driver, err := drivers.LoadPCSDriver(config.PCSDeviceConfig{ID: cfg.ID, Host: cfg.Host, Port: cfg.Port, Unit: cfg.UnitID, Brand: cfg.Brand, CommandMode: stringExtra(cfg.Extra, "command_mode")}, client)
		if err != nil {
			return closeOnError(err)
		}
		base.vendor = driver.Brand()
		base.close = closeBoth(driver, client)
		base.poll = func() (devices.DataPoints, error) {
			value, err := driver.ReadState()
			if err != nil {
				return nil, err
			}
			return points(map[string]interface{}{"active_power_setpoint_w": value.ActivePowerSetpointW, "active_power_w": value.ActualActivePowerW, "reactive_power_var": value.ReactivePowerVar, "status": value.Status.String(), "command_mode": driver.CommandMode()}), nil
		}
		base.write = func(command devices.Command) error {
			if command.Name != "set_active_power_w" {
				return unsupported(command.Name)
			}
			value, err := intParam(command, "watts")
			if err != nil {
				return err
			}
			return driver.WriteSetpoint(value)
		}
	case devices.DeviceTypeEVCharger:
		adapter := devices.NewEVChargerAdapter(client, byte(cfg.UnitID))
		base.poll = func() (devices.DataPoints, error) {
			if err := adapter.Poll(); err != nil {
				return nil, err
			}
			return evPoints(adapter.Data()), nil
		}
		base.write = evWriter(adapter.WriteMaxPower, adapter.WriteStartStop)
	default:
		return closeOnError(fmt.Errorf("adapter factory: unsupported device type %q", cfg.Type))
	}
	return base, nil
}

func (f *Factory) createOCPP(cfg devices.DeviceConfig) (devices.SnapshotDevice, error) {
	if f.OCPP == nil {
		return nil, fmt.Errorf("adapter factory: OCPP central system is required for %s", cfg.ID)
	}
	adapter := devices.NewOCPPEVChargerAdapter(cfg.ID, f.OCPP)
	base := newManaged(cfg, nil)
	base.poll = func() (devices.DataPoints, error) {
		if err := adapter.Poll(); err != nil {
			return nil, err
		}
		return evPoints(adapter.Data()), nil
	}
	base.write = evWriter(adapter.WriteMaxPower, adapter.WriteStartStop)
	return base, nil
}

type managedAdapter struct {
	mu     sync.RWMutex
	cfg    devices.DeviceConfig
	vendor string
	health *devices.HealthTracker
	data   devices.DataPoints
	poll   func() (devices.DataPoints, error)
	write  func(devices.Command) error
	close  func() error
}

func newManaged(cfg devices.DeviceConfig, closeFn func() error) *managedAdapter {
	vendor := cfg.Brand
	if vendor == "" {
		vendor = "generic"
	}
	return &managedAdapter{cfg: cfg, vendor: vendor, health: devices.NewHealthTracker(), data: make(devices.DataPoints), close: closeFn}
}
func (a *managedAdapter) ID() string                   { return a.cfg.ID }
func (a *managedAdapter) Type() devices.DeviceType     { return a.cfg.Type }
func (a *managedAdapter) Protocol() string             { return a.cfg.Protocol }
func (a *managedAdapter) ProtocolType() string         { return a.cfg.Protocol }
func (a *managedAdapter) Vendor() string               { return a.vendor }
func (a *managedAdapter) Health() devices.DeviceHealth { return a.health.Health() }
func (a *managedAdapter) Status() devices.DeviceStatus {
	h := a.health.Health()
	if h.Connected && h.ConsecutiveErrs < 3 {
		return devices.DeviceStatusOnline
	}
	if h.ConsecutiveErrs >= 3 {
		return devices.DeviceStatusFault
	}
	return devices.DeviceStatusOffline
}
func (a *managedAdapter) Poll() error {
	data, err := a.poll()
	if err != nil {
		a.health.RecordError(err)
		return err
	}
	a.mu.Lock()
	a.data = clonePoints(data)
	a.mu.Unlock()
	a.health.RecordSuccess()
	return nil
}
func (a *managedAdapter) Snapshot() devices.DataPoints {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return clonePoints(a.data)
}
func (a *managedAdapter) Write(command devices.Command) error {
	if a.write == nil {
		return fmt.Errorf("device %s is read-only", a.cfg.ID)
	}
	if err := a.write(command); err != nil {
		a.health.RecordError(err)
		return err
	}
	a.health.RecordSuccess()
	return nil
}
func (a *managedAdapter) Close() error {
	if a.close != nil {
		return a.close()
	}
	return nil
}

var _ devices.SnapshotDevice = (*managedAdapter)(nil)
var _ devices.ControllableDevice = (*managedAdapter)(nil)
var _ io.Closer = (*managedAdapter)(nil)

type evCommands interface {
	WriteMaxPower(int) error
	WriteStartStop(bool) error
}

func evWriter(maxPower func(int) error, startStop func(bool) error) func(devices.Command) error {
	return func(command devices.Command) error {
		switch command.Name {
		case "set_max_power_w":
			value, err := intParam(command, "watts")
			if err != nil {
				return err
			}
			return maxPower(value)
		case "start":
			return startStop(true)
		case "stop":
			return startStop(false)
		default:
			return unsupported(command.Name)
		}
	}
}
func evPoints(value devices.EVChargerData) devices.DataPoints {
	return points(map[string]interface{}{"status": int(value.Status), "active_power_w": value.ActivePower, "energy_delivered_wh": value.EnergyDelivered, "max_power_limit_w": value.MaxPowerLimit, "vehicle_connected": value.VehicleConnected, "current_l1_a": value.CurrentL1, "current_l2_a": value.CurrentL2, "current_l3_a": value.CurrentL3})
}
func points(values map[string]interface{}) devices.DataPoints {
	now := time.Now().UTC()
	result := make(devices.DataPoints, len(values))
	for name, value := range values {
		result[name] = devices.DataPoint{Name: name, Value: value, Timestamp: now, Quality: "good"}
	}
	return result
}
func clonePoints(input devices.DataPoints) devices.DataPoints {
	result := make(devices.DataPoints, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}
func validateConfig(cfg devices.DeviceConfig) error {
	if cfg.ID == "" {
		return fmt.Errorf("adapter factory: device id is required")
	}
	if cfg.Type == "" {
		return fmt.Errorf("adapter factory: device type is required for %s", cfg.ID)
	}
	if cfg.Protocol == "" {
		return fmt.Errorf("adapter factory: protocol is required for %s", cfg.ID)
	}
	if strings.ToLower(cfg.Protocol) != "ocpp" && (cfg.Host == "" || cfg.Port <= 0 || cfg.UnitID < 0 || cfg.UnitID > 255) {
		return fmt.Errorf("adapter factory: invalid Modbus address for %s", cfg.ID)
	}
	return nil
}
func stringExtra(extra map[string]interface{}, key string) string {
	if extra == nil {
		return ""
	}
	value, _ := extra[key].(string)
	return value
}
func intParam(command devices.Command, key string) (int, error) {
	raw, ok := command.Params[key]
	if !ok {
		return 0, fmt.Errorf("command %s: missing %s", command.Name, key)
	}
	switch value := raw.(type) {
	case int:
		return value, nil
	case float64:
		return int(value), nil
	default:
		return 0, fmt.Errorf("command %s: %s must be numeric", command.Name, key)
	}
}
func unsupported(name string) error { return fmt.Errorf("unsupported command %q", name) }
func closeBoth(driver io.Closer, client io.Closer) func() error {
	return func() error {
		first := driver.Close()
		second := client.Close()
		if first != nil {
			return first
		}
		return second
	}
}
