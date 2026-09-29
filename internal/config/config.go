// Package config provides YAML-based configuration for the Local EMS.
package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type SiteConfig struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

type Config struct {
	Site        SiteConfig        `yaml:"site"`
	Simulate    bool              `yaml:"simulate"`
	Devices     DevicesConfig     `yaml:"devices"`
	Topology    TopologyConfig    `yaml:"topology"`
	Polling     PollingConfig     `yaml:"polling"`
	Engine      EngineConfig      `yaml:"engine"`
	Storage     StorageConfig     `yaml:"storage"`
	Sync        SyncConfig        `yaml:"sync"`
	UI          UIConfig          `yaml:"ui"`
	Controllers ControllersConfig `yaml:"controllers"`
	Simulator   SimulatorConfig   `yaml:"simulator"`
}

// --- Controller Configs (tunable via YAML + env vars) ---

type ControllersConfig struct {
	LimitDischarge LimitDischargeConfig `yaml:"limit_discharge"`
	SellToGrid     SellToGridConfig     `yaml:"sell_to_grid_limit"`
	PeakShaving    PeakShavingConfig    `yaml:"peak_shaving"`
	TimeOfUse      TimeOfUseConfig      `yaml:"time_of_use"`
	EVCharging     EVChargingConfig     `yaml:"ev_charging"`
	Balancing      BalancingConfig       `yaml:"balancing"`
}

type LimitDischargeConfig struct {
	Enabled        bool    `yaml:"enabled"`
	MinSOC         float64 `yaml:"min_soc"`
	ForceChargeSOC float64 `yaml:"force_charge_soc"`
	ForceChargeW   int     `yaml:"force_charge_w"`
}

type SellToGridConfig struct {
	Enabled            bool `yaml:"enabled"`
	MaxSellToGridPower int  `yaml:"max_sell_to_grid_w"`
}

type PeakShavingConfig struct {
	Enabled        bool `yaml:"enabled"`
	PeakThresholdW int  `yaml:"peak_threshold_w"`
}

type TOUPeriodConfig struct {
	StartHour int    `yaml:"start_hour"`
	EndHour   int    `yaml:"end_hour"`
	Rate      string `yaml:"rate"` // "off_peak", "on_peak", "mid_peak"
}

type TimeOfUseConfig struct {
	Enabled           bool              `yaml:"enabled"`
	MaxChargeW        int               `yaml:"max_charge_w"`
	MaxDischargeW     int               `yaml:"max_discharge_w"`
	MinSOCToDischarge float64           `yaml:"min_soc_to_discharge"`
	Periods           []TOUPeriodConfig `yaml:"periods"`
}

type EVChargingConfig struct {
	Enabled       bool `yaml:"enabled"`
	MaxSitePowerW int  `yaml:"max_site_power_w"`
}

type BalancingConfig struct {
	Enabled bool `yaml:"enabled"`
}

// --- Simulator Config (tunable via YAML + env vars) ---

type SimulatorConfig struct {
	InitialSOC        float64 `yaml:"initial_soc"`
	SolarPeakW        int     `yaml:"solar_peak_w"`
	LoadBaseW         int     `yaml:"load_base_w"`
	LoadVarianceW     int     `yaml:"load_variance_w"`
	BatteryCapacityWh int     `yaml:"battery_capacity_wh"`
	MaxChargeRateW    int     `yaml:"max_charge_rate_w"`
	MaxDischargeRateW int     `yaml:"max_discharge_rate_w"`
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
// Priority (highest wins): ENV vars > .env file > YAML values > defaults.
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
	cfg.applyEnvOverrides()

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

	// --- Controller defaults (only fill if YAML didn't set them) ---
	ld := &c.Controllers.LimitDischarge
	if ld.MinSOC == 0 {
		ld.MinSOC = 15.0
	}
	if ld.ForceChargeSOC == 0 {
		ld.ForceChargeSOC = 10.0
	}
	if ld.ForceChargeW == 0 {
		ld.ForceChargeW = 2000
	}
	if !ld.Enabled {
		ld.Enabled = true // safety controller defaults to enabled
	}

	stg := &c.Controllers.SellToGrid
	if !stg.Enabled {
		stg.Enabled = true // grid compliance defaults to enabled
	}
	// MaxSellToGridPower=0 is valid (zero-export), so no zero-check

	ps := &c.Controllers.PeakShaving
	if ps.PeakThresholdW == 0 {
		ps.PeakThresholdW = 50000
	}
	if !ps.Enabled {
		ps.Enabled = true
	}

	tou := &c.Controllers.TimeOfUse
	if tou.MaxChargeW == 0 {
		tou.MaxChargeW = 3000
	}
	if tou.MaxDischargeW == 0 {
		tou.MaxDischargeW = 5000
	}
	if tou.MinSOCToDischarge == 0 {
		tou.MinSOCToDischarge = 20.0
	}
	if !tou.Enabled {
		tou.Enabled = true
	}
	if len(tou.Periods) == 0 {
		tou.Periods = []TOUPeriodConfig{
			{StartHour: 22, EndHour: 6, Rate: "off_peak"},
			{StartHour: 9, EndHour: 12, Rate: "on_peak"},
			{StartHour: 17, EndHour: 21, Rate: "on_peak"},
		}
	}

	ev := &c.Controllers.EVCharging
	if ev.MaxSitePowerW == 0 {
		ev.MaxSitePowerW = 50000
	}
	// EVCharging.Enabled defaults to false — opt-in based on whether EV chargers exist
	if len(c.Devices.EVCharger) > 0 && !ev.Enabled {
		ev.Enabled = true
	}

	bal := &c.Controllers.Balancing
	if !bal.Enabled {
		bal.Enabled = true
	}

	// --- Simulator defaults ---
	sim := &c.Simulator
	if sim.InitialSOC == 0 {
		sim.InitialSOC = 50.0
	}
	if sim.SolarPeakW == 0 {
		sim.SolarPeakW = 10000
	}
	if sim.LoadBaseW == 0 {
		sim.LoadBaseW = 5000
	}
	if sim.LoadVarianceW == 0 {
		sim.LoadVarianceW = 3000
	}
	if sim.BatteryCapacityWh == 0 {
		sim.BatteryCapacityWh = 5000000
	}
	if sim.MaxChargeRateW == 0 {
		sim.MaxChargeRateW = 50000
	}
	if sim.MaxDischargeRateW == 0 {
		sim.MaxDischargeRateW = 50000
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
	cfg := &Config{
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
	cfg.applyDefaults()
	cfg.applyEnvOverrides()
	return cfg
}

// --- .env file loader & environment variable overrides ---

// loadDotEnv reads a .env file and sets environment variables.
// Does NOT override variables that are already set in the environment.
// Looks for EMS_ENV_FILE env var first, then falls back to ".env" in cwd.
func loadDotEnv() {
	path := os.Getenv("EMS_ENV_FILE")
	if path == "" {
		path = ".env"
	}

	f, err := os.Open(path)
	if err != nil {
		return // .env is optional — no error if missing
	}
	defer f.Close()

	log.Printf("config: loading env from %s", path)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		// Strip surrounding quotes
		value = strings.Trim(value, "\"'")
		// Don't override existing env vars
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

// applyEnvOverrides reads EMS_* environment variables and overrides config values.
// Called after applyDefaults, so env vars always win over YAML and defaults.
func (c *Config) applyEnvOverrides() {
	loadDotEnv()

	// --- Core ---
	envBool("EMS_SIMULATE", &c.Simulate)
	envInt("EMS_UI_PORT", &c.UI.Port)
	envStr("EMS_UI_USERNAME", &c.UI.Username)
	envStr("EMS_UI_PASSWORD", &c.UI.Password)
	envStr("EMS_DB_DSN", &c.Storage.DSN)
	envInt("EMS_RETENTION_DAYS", &c.Storage.RetentionDays)

	// --- Controllers ---
	envFloat("EMS_MIN_SOC", &c.Controllers.LimitDischarge.MinSOC)
	envFloat("EMS_FORCE_CHARGE_SOC", &c.Controllers.LimitDischarge.ForceChargeSOC)
	envInt("EMS_FORCE_CHARGE_W", &c.Controllers.LimitDischarge.ForceChargeW)
	envInt("EMS_MAX_SELL_TO_GRID_W", &c.Controllers.SellToGrid.MaxSellToGridPower)
	envInt("EMS_PEAK_THRESHOLD_W", &c.Controllers.PeakShaving.PeakThresholdW)
	envBool("EMS_PEAK_SHAVING_ENABLED", &c.Controllers.PeakShaving.Enabled)
	envInt("EMS_TOU_MAX_CHARGE_W", &c.Controllers.TimeOfUse.MaxChargeW)
	envInt("EMS_TOU_MAX_DISCHARGE_W", &c.Controllers.TimeOfUse.MaxDischargeW)
	envFloat("EMS_TOU_MIN_SOC", &c.Controllers.TimeOfUse.MinSOCToDischarge)
	envBool("EMS_TOU_ENABLED", &c.Controllers.TimeOfUse.Enabled)
	envInt("EMS_EV_MAX_SITE_POWER_W", &c.Controllers.EVCharging.MaxSitePowerW)
	envBool("EMS_EV_ENABLED", &c.Controllers.EVCharging.Enabled)
	envBool("EMS_BALANCING_ENABLED", &c.Controllers.Balancing.Enabled)

	// --- Simulator ---
	envFloat("EMS_SIM_INITIAL_SOC", &c.Simulator.InitialSOC)
	envInt("EMS_SIM_SOLAR_PEAK_W", &c.Simulator.SolarPeakW)
	envInt("EMS_SIM_LOAD_BASE_W", &c.Simulator.LoadBaseW)
	envInt("EMS_SIM_LOAD_VARIANCE_W", &c.Simulator.LoadVarianceW)
	envInt("EMS_SIM_BATTERY_CAPACITY_WH", &c.Simulator.BatteryCapacityWh)
	envInt("EMS_SIM_MAX_CHARGE_W", &c.Simulator.MaxChargeRateW)
	envInt("EMS_SIM_MAX_DISCHARGE_W", &c.Simulator.MaxDischargeRateW)
}

// --- env var helpers (stdlib only, no dependencies) ---

func envStr(key string, dst *string) {
	if v := os.Getenv(key); v != "" {
		*dst = v
	}
}

func envInt(key string, dst *int) {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			*dst = n
		}
	}
}

func envFloat(key string, dst *float64) {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			*dst = f
		}
	}
}

func envBool(key string, dst *bool) {
	if v := os.Getenv(key); v != "" {
		switch strings.ToLower(v) {
		case "true", "1", "yes":
			*dst = true
		case "false", "0", "no":
			*dst = false
		}
	}
}
