// Package config provides YAML-based configuration for the Local EMS.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type SiteConfig struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

type Config struct {
	Site     SiteConfig     `yaml:"site"`
	Simulate bool           `yaml:"simulate"`
	Devices  DevicesConfig  `yaml:"devices"`
	Topology TopologyConfig `yaml:"topology"`
	Polling  PollingConfig  `yaml:"polling"`
	Engine   EngineConfig   `yaml:"engine"`
	Storage  StorageConfig  `yaml:"storage"`
	Sync     SyncConfig     `yaml:"sync"`
	UI       UIConfig       `yaml:"ui"`
}

type DevicesConfig struct {
	BMS       []BMSDeviceConfig       `yaml:"bms"`
	PCS       []PCSDeviceConfig       `yaml:"pcs"`
	Meter     []MeterDeviceConfig     `yaml:"meter"`
	EVCharger []EVChargerDeviceConfig `yaml:"ev_charger"`
}

type BMSDeviceConfig struct {
	ID   string `yaml:"id"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Unit int    `yaml:"unit"`
}

type PCSDeviceConfig struct {
	ID          string `yaml:"id"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Unit        int    `yaml:"unit"`
	Brand       string `yaml:"brand,omitempty"`
	CommandMode string `yaml:"command_mode,omitempty"`
}

type MeterDeviceConfig struct {
	ID     string `yaml:"id"`
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	Unit   int    `yaml:"unit"`
	Role   string `yaml:"role,omitempty"`
	FastMs int    `yaml:"fast_ms,omitempty"`
}

type EVChargerDeviceConfig struct {
	ID       string `yaml:"id"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Unit     int    `yaml:"unit"`
	Protocol string `yaml:"protocol"` // "modbus" or "ocpp"
}

type TopologyConfig struct {
	PCCMeter   string            `yaml:"pcc_meter"`
	SolarMeter string            `yaml:"solar_meter,omitempty"`
	LoadMeter  string            `yaml:"load_meter,omitempty"`
	BESSGroups []BESSGroupConfig `yaml:"bess_groups"`
}

type BESSGroupConfig struct {
	Name        string   `yaml:"name"`
	BMS         []string `yaml:"bms"`
	PCS         []string `yaml:"pcs"`
	CommandMode string   `yaml:"command_mode"` // "master" or "individual"
}

type SyncConfig struct {
	Enabled         bool   `yaml:"enabled"`
	MQTTBroker      string `yaml:"mqtt_broker,omitempty"`
	ClientID        string `yaml:"client_id,omitempty"`
	TopicPrefix     string `yaml:"topic_prefix,omitempty"`
	IntervalSeconds int    `yaml:"interval_seconds,omitempty"`
	BufferDays      int    `yaml:"buffer_days,omitempty"`
}

type PollingConfig struct {
	FastMs     int `yaml:"fast_ms"`
	StandardMs int `yaml:"standard_ms"`
	SlowMs     int `yaml:"slow_ms"`
}

type EngineConfig struct {
	CycleMs      int    `yaml:"cycle_ms"`
	Distribution string `yaml:"distribution"`
}

type StorageConfig struct {
	Driver        string `yaml:"driver"`
	DSN           string `yaml:"dsn,omitempty"`
	Path          string `yaml:"path,omitempty"`
	RetentionDays int    `yaml:"retention_days,omitempty"`
	WALMode       bool   `yaml:"wal_mode,omitempty"`
}

type UIConfig struct {
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// FastInterval returns the fast polling interval.
func (c *Config) FastInterval() time.Duration {
	if c.Polling.FastMs <= 0 {
		return 200 * time.Millisecond
	}
	return time.Duration(c.Polling.FastMs) * time.Millisecond
}

// StandardInterval returns the standard polling interval.
func (c *Config) StandardInterval() time.Duration {
	if c.Polling.StandardMs <= 0 {
		return 1 * time.Second
	}
	return time.Duration(c.Polling.StandardMs) * time.Millisecond
}

// SlowInterval returns the slow polling interval.
func (c *Config) SlowInterval() time.Duration {
	if c.Polling.SlowMs <= 0 {
		return 5 * time.Second
	}
	return time.Duration(c.Polling.SlowMs) * time.Millisecond
}

// Load reads a YAML config file and returns a Config.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}

	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: validate: %w", err)
	}

	return cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Site.ID == "" {
		c.Site.ID = "default-site"
	}
	if c.Storage.RetentionDays == 0 {
		c.Storage.RetentionDays = 7
	}
	if c.Sync.IntervalSeconds == 0 {
		c.Sync.IntervalSeconds = 300
	}
	if c.Sync.BufferDays == 0 {
		c.Sync.BufferDays = 7
	}

	for i := range c.Devices.PCS {
		if c.Devices.PCS[i].Brand == "" {
			c.Devices.PCS[i].Brand = "generic"
		}
		if c.Devices.PCS[i].CommandMode == "" {
			c.Devices.PCS[i].CommandMode = "individual"
		}
	}

	for i := range c.Devices.EVCharger {
		if c.Devices.EVCharger[i].Protocol == "" {
			c.Devices.EVCharger[i].Protocol = "modbus"
		}
	}
}

func (c *Config) validate() error {
	if len(c.Devices.BMS) == 0 {
		return fmt.Errorf("at least 1 BMS required")
	}
	if len(c.Devices.PCS) == 0 {
		return fmt.Errorf("at least 1 PCS required")
	}
	if len(c.Devices.Meter) == 0 {
		return fmt.Errorf("at least 1 Meter required")
	}

	seen := make(map[string]bool)
	checkDevice := func(id, host string, port int) error {
		if id == "" {
			return fmt.Errorf("device missing id")
		}
		if seen[id] {
			return fmt.Errorf("duplicate device id: %s", id)
		}
		seen[id] = true
		if host == "" {
			return fmt.Errorf("device %s: missing host", id)
		}
		if port <= 0 {
			return fmt.Errorf("device %s: invalid port", id)
		}
		return nil
	}

	for _, d := range c.Devices.BMS {
		if err := checkDevice(d.ID, d.Host, d.Port); err != nil {
			return err
		}
	}
	for _, d := range c.Devices.PCS {
		if err := checkDevice(d.ID, d.Host, d.Port); err != nil {
			return err
		}
	}
	for _, d := range c.Devices.Meter {
		if err := checkDevice(d.ID, d.Host, d.Port); err != nil {
			return err
		}
	}
	for _, d := range c.Devices.EVCharger {
		if err := checkDevice(d.ID, d.Host, d.Port); err != nil {
			return err
		}
	}

	// Topology checks
	if c.Topology.PCCMeter != "" && !seen[c.Topology.PCCMeter] {
		return fmt.Errorf("topology: pcc_meter %s not found in devices", c.Topology.PCCMeter)
	}
	if c.Topology.SolarMeter != "" && !seen[c.Topology.SolarMeter] {
		return fmt.Errorf("topology: solar_meter %s not found in devices", c.Topology.SolarMeter)
	}
	if c.Topology.LoadMeter != "" && !seen[c.Topology.LoadMeter] {
		return fmt.Errorf("topology: load_meter %s not found in devices", c.Topology.LoadMeter)
	}

	for _, group := range c.Topology.BESSGroups {
		for _, bmsID := range group.BMS {
			if !seen[bmsID] {
				return fmt.Errorf("topology: bms %s in group %s not found in devices", bmsID, group.Name)
			}
		}
		for _, pcsID := range group.PCS {
			if !seen[pcsID] {
				return fmt.Errorf("topology: pcs %s in group %s not found in devices", pcsID, group.Name)
			}
		}
	}

	if c.Sync.Enabled && c.Sync.MQTTBroker == "" {
		return fmt.Errorf("sync is enabled but mqtt_broker is missing")
	}

	return nil
}

// DefaultDev returns a dev config with simulated devices.
func DefaultDev() *Config {
	return &Config{
		Site: SiteConfig{
			ID:   "dev-site",
			Name: "Development Site",
		},
		Devices: DevicesConfig{
			BMS: []BMSDeviceConfig{
				{ID: "bms-0", Host: "127.0.0.1", Port: 5020, Unit: 1},
				{ID: "bms-1", Host: "127.0.0.1", Port: 5020, Unit: 2},
			},
			PCS: []PCSDeviceConfig{
				{ID: "pcs-0", Host: "127.0.0.1", Port: 5020, Unit: 3, Brand: "generic", CommandMode: "individual"},
				{ID: "pcs-1", Host: "127.0.0.1", Port: 5020, Unit: 4, Brand: "generic", CommandMode: "individual"},
			},
			Meter: []MeterDeviceConfig{
				{ID: "meter-0", Host: "127.0.0.1", Port: 5020, Unit: 5, Role: "grid"},
				{ID: "meter-1", Host: "127.0.0.1", Port: 5020, Unit: 6, Role: "solar"},
			},
			EVCharger: []EVChargerDeviceConfig{
				{ID: "ev-0", Host: "127.0.0.1", Port: 5020, Unit: 7, Protocol: "modbus"},
				{ID: "ev-1", Host: "127.0.0.1", Port: 5020, Unit: 8, Protocol: "modbus"},
			},
		},
		Topology: TopologyConfig{
			PCCMeter: "meter-0",
			BESSGroups: []BESSGroupConfig{
				{
					Name:        "group-0",
					BMS:         []string{"bms-0"},
					PCS:         []string{"pcs-0"},
					CommandMode: "individual",
				},
				{
					Name:        "group-1",
					BMS:         []string{"bms-1"},
					PCS:         []string{"pcs-1"},
					CommandMode: "individual",
				},
			},
		},
		Polling: PollingConfig{
			FastMs:     200,
			StandardMs: 1000,
			SlowMs:     5000,
		},
		Engine: EngineConfig{
			CycleMs:      1000,
			Distribution: "equal",
		},
		Storage: StorageConfig{
			Driver:        "sqlite",
			DSN:           "/tmp/ems-dev.db",
			RetentionDays: 7,
			WALMode:       true,
		},
		Sync: SyncConfig{
			Enabled:         false,
			MQTTBroker:      "tcp://localhost:1883",
			ClientID:        "dev-ems",
			TopicPrefix:     "ems/dev",
			IntervalSeconds: 300,
			BufferDays:      7,
		},
		UI: UIConfig{
			Port:     8080,
			Username: "admin",
			Password: "ems@2025",
		},
		Simulate: true,
	}
}
