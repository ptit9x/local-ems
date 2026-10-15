package drivers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vmo/ems-platform/ems-local/internal/config"
	"github.com/vmo/ems-platform/ems-local/internal/devices"
	"github.com/vmo/ems-platform/ems-local/internal/modbus"
	"gopkg.in/yaml.v3"
)

// yamlProfile structure for parsing custom YAML PCS profiles.
type yamlProfile struct {
	Brand       string            `yaml:"brand"`
	Registers   map[string]uint16 `yaml:"registers"`
	ScaleFactor float64           `yaml:"scale_factor"`
	CommandMode string            `yaml:"command_mode"`
	StatusMap   map[uint16]uint16 `yaml:"status_map"`
	Description string            `yaml:"description"`
}

// LoadPCSDriver loads the appropriate PCSDriver based on the configuration.
func LoadPCSDriver(cfg config.PCSDeviceConfig, client *modbus.Client) (PCSDriver, error) {
	registry := DefaultPCSProfileRegistry()

	// 1. Try to load from built-in registry
	if profile, err := registry.Get(cfg.Brand); err == nil {
		return NewGenericPCSDriver(client, byte(cfg.Unit), profile), nil
	}

	// 2. Try to load from YAML file
	yamlPath := filepath.Join("config", "pcs_profiles", fmt.Sprintf("%s.yaml", cfg.Brand))
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load driver for brand '%s': not in registry and %s not found: %w", cfg.Brand, yamlPath, err)
	}

	var yp yamlProfile
	if err := yaml.Unmarshal(data, &yp); err != nil {
		return nil, fmt.Errorf("failed to parse YAML profile for brand '%s': %w", cfg.Brand, err)
	}

	// Map status
	statusMap := make(map[uint16]devices.PCSStatus)
	for k, v := range yp.StatusMap {
		statusMap[k] = devices.PCSStatus(v)
	}

	customProfile := PCSProfile{
		Brand:             yp.Brand,
		ActivePowerSetReg: yp.Registers["active_power_setpoint"],
		ActivePowerActReg: yp.Registers["active_power_actual"],
		ReactivePowerReg:  yp.Registers["reactive_power"],
		StatusReg:         yp.Registers["status"],
		RegisterCount:     yp.Registers["register_count"],
		ScaleFactor:       yp.ScaleFactor,
		StatusMap:         statusMap,
		CommandMode:       yp.CommandMode,
		Description:       yp.Description,
	}

	return NewGenericPCSDriver(client, byte(cfg.Unit), customProfile), nil
}
