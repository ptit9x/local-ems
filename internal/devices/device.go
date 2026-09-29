package devices

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// DeviceType represents the type of a device.
type DeviceType string

const (
	DeviceTypeBMS       DeviceType = "BMS"
	DeviceTypePCS       DeviceType = "PCS"
	DeviceTypeMeter     DeviceType = "Meter"
	DeviceTypeEVCharger DeviceType = "EVCharger"
)

// DeviceStatus represents the current status of a device.
type DeviceStatus string

const (
	DeviceStatusOnline  DeviceStatus = "Online"
	DeviceStatusOffline DeviceStatus = "Offline"
	DeviceStatusFault   DeviceStatus = "Fault"
	DeviceStatusUnknown DeviceStatus = "Unknown"
)

// DataPoint represents a single data point from a device.
type DataPoint struct {
	Name      string
	Value     interface{}
	Timestamp time.Time
	Quality   string
}

// DataPoints is a collection of data points mapped by their name.
type DataPoints map[string]DataPoint

// Command represents a command to be written to a device.
type Command struct {
	Name   string
	Params map[string]interface{}
}

// DeviceConfig holds the configuration for a device.
type DeviceConfig struct {
	ID       string
	Type     DeviceType
	Host     string
	Port     int
	UnitID   int
	Protocol string
	Brand    string
	Extra    map[string]interface{}
}

// Device defines the base interface for all devices.
type Device interface {
	ID() string
	Type() DeviceType
	Protocol() string
	Status() DeviceStatus
	Poll() error
}

// ControllableDevice defines a device that can accept commands.
type ControllableDevice interface {
	Device
	Write(cmd Command) error
}

// ReconfigurableDevice defines a device that can be reconfigured dynamically.
type ReconfigurableDevice interface {
	Device
	Reconfigure(newConfig DeviceConfig) error
}

// DeviceRegistry manages a thread-safe collection of devices.
type DeviceRegistry struct {
	mu      sync.RWMutex
	devices map[string]Device
}

// NewDeviceRegistry creates a new DeviceRegistry.
func NewDeviceRegistry() *DeviceRegistry {
	return &DeviceRegistry{
		devices: make(map[string]Device),
	}
}

// Register adds a device to the registry.
func (r *DeviceRegistry) Register(dev Device) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := dev.ID()
	if _, exists := r.devices[id]; exists {
		return fmt.Errorf("registering device: device with ID %s already exists", id)
	}
	r.devices[id] = dev
	return nil
}

// Unregister removes a device from the registry.
func (r *DeviceRegistry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.devices[id]; !exists {
		return fmt.Errorf("unregistering device: device with ID %s not found", id)
	}
	delete(r.devices, id)
	return nil
}

// Get retrieves a device by its ID.
func (r *DeviceRegistry) Get(id string) (Device, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dev, exists := r.devices[id]
	return dev, exists
}

// GetByType retrieves all devices of a specific type.
func (r *DeviceRegistry) GetByType(t DeviceType) []Device {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Device
	for _, dev := range r.devices {
		if dev.Type() == t {
			result = append(result, dev)
		}
	}
	return result
}

// Reconfigure attempts to reconfigure an existing device.
func (r *DeviceRegistry) Reconfigure(id string, newConfig DeviceConfig) error {
	r.mu.RLock()
	dev, exists := r.devices[id]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("reconfiguring device: device with ID %s not found", id)
	}

	reconfDev, ok := dev.(ReconfigurableDevice)
	if !ok {
		return fmt.Errorf("reconfiguring device: device %s does not support dynamic reconfiguration", id)
	}

	if err := reconfDev.Reconfigure(newConfig); err != nil {
		return fmt.Errorf("reconfiguring device %s: %w", id, err)
	}

	return nil
}

// All returns a slice of all registered devices.
func (r *DeviceRegistry) All() []Device {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Device, 0, len(r.devices))
	for _, dev := range r.devices {
		result = append(result, dev)
	}
	return result
}

// Close closes all devices that implement io.Closer.
func (r *DeviceRegistry) Close() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, dev := range r.devices {
		if closer, ok := dev.(io.Closer); ok {
			_ = closer.Close()
		}
	}
}
