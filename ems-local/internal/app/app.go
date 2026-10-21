// Package app provides the application lifecycle and dependency wiring for the
// EMS local controller. It was extracted from cmd/ems-core/main.go to keep the
// entrypoint minimal and make the wiring logic testable.
package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/vmo/ems-platform/ems-local/internal/adapters"
	"github.com/vmo/ems-platform/ems-local/internal/build"
	"github.com/vmo/ems-platform/ems-local/internal/config"
	"github.com/vmo/ems-platform/ems-local/internal/device_runtime"
	"github.com/vmo/ems-platform/ems-local/internal/devices"
	"github.com/vmo/ems-platform/ems-local/internal/devices/aggregator"
	"github.com/vmo/ems-platform/ems-local/internal/engine/controllers"
	"github.com/vmo/ems-platform/ems-local/internal/engine/controllers/balancing"
	"github.com/vmo/ems-platform/ems-local/internal/engine/controllers/ev_charging"
	"github.com/vmo/ems-platform/ems-local/internal/engine/controllers/limit_discharge"
	"github.com/vmo/ems-platform/ems-local/internal/engine/controllers/peak_shaving"
	"github.com/vmo/ems-platform/ems-local/internal/engine/controllers/sell_to_grid_limit"
	"github.com/vmo/ems-platform/ems-local/internal/engine/controllers/time_of_use"
	"github.com/vmo/ems-platform/ems-local/internal/engine/cycle"
	"github.com/vmo/ems-platform/ems-local/internal/engine/distributor"
	"github.com/vmo/ems-platform/ems-local/internal/engine/resolver"
	"github.com/vmo/ems-platform/ems-local/internal/engine/scheduler"
	"github.com/vmo/ems-platform/ems-local/internal/failover"
	"github.com/vmo/ems-platform/ems-local/internal/health"
	"github.com/vmo/ems-platform/ems-local/internal/ocpp"
	"github.com/vmo/ems-platform/ems-local/internal/simulator"
	"github.com/vmo/ems-platform/ems-local/internal/store"
	emsSync "github.com/vmo/ems-platform/ems-local/internal/sync"
	"github.com/vmo/ems-platform/ems-local/internal/ui"
)

// App holds all the wired-up services for the local EMS controller.
type App struct {
	cfg *config.Config

	db              *store.Store
	monitor         *health.Monitor
	agg             *aggregator.Aggregator
	sim             *simulator.Simulator
	dist            *distributor.Distributor
	cm              *cycle.Manager
	uiServer        *ui.Server
	syncer          *emsSync.Syncer
	cmdSub          *emsSync.CommandSubscriber
	heartbeat       *failover.HeartbeatPublisher
	failoverManager *failover.FailoverManager
	deviceRegistry  *devices.DeviceRegistry
	deviceRuntime   *device_runtime.Orchestrator
	ocppCentral     *ocpp.CentralSystem
	evCtrl          *ev_charging.EVChargingController

	failoverSubscribe func() error
	cycleCount        int
}

// New creates and wires all EMS services based on the given config.
func New(cfg *config.Config) (*App, error) {
	a := &App{cfg: cfg}

	nBMS := len(cfg.Devices.BMS)
	nPCS := len(cfg.Devices.PCS)
	nMeter := len(cfg.Devices.Meter)
	nEV := len(cfg.Devices.EVCharger)

	// --- Storage ---
	var storeOpts []store.Option
	if cfg.Storage.WALMode {
		storeOpts = append(storeOpts, store.WithWAL())
	}
	if cfg.Storage.RetentionDays > 0 {
		storeOpts = append(storeOpts, store.WithAutoRetention(cfg.Storage.RetentionDays))
	}
	var err error
	a.db, err = store.New(cfg.Storage.DSN, storeOpts...)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// --- Health Monitor ---
	a.monitor = health.NewMonitor()

	// --- Aggregator ---
	a.agg = aggregator.New(nBMS, nPCS, nMeter, nEV)

	// --- Physical device runtime ---
	if !cfg.Simulate {
		hasOCPP := false
		for _, charger := range cfg.Devices.EVCharger {
			if charger.Protocol == "ocpp" {
				hasOCPP = true
				break
			}
		}
		if hasOCPP {
			a.ocppCentral = ocpp.NewCentralSystem(cfg.OCPP.ListenAddress, nil)
			if err := a.ocppCentral.Start(); err != nil {
				a.db.Close()
				return nil, fmt.Errorf("start OCPP Central System: %w", err)
			}
			log.Printf("ocpp: Local Central System listening on %s", cfg.OCPP.ListenAddress)
		}
		factory := adapters.NewFactory(a.ocppCentral)
		a.deviceRegistry, err = factory.Build(cfg)
		if err != nil {
			a.Close()
			return nil, fmt.Errorf("build physical device registry: %w", err)
		}
		a.deviceRuntime = device_runtime.New(a.deviceRegistry, a.agg, cfg, func() bool {
			return a.failoverManager == nil || a.failoverManager.ControlOwner() == "local"
		})
	}

	// --- Distributor ---
	a.dist = distributor.New(distributor.Strategy(cfg.Engine.Distribution), nPCS)

	// --- Base Simulator ---
	a.sim = simulator.New(simulator.Config{
		InitialSOC:        cfg.Simulator.InitialSOC,
		SolarPeakW:        cfg.Simulator.SolarPeakW,
		LoadBaseW:         cfg.Simulator.LoadBaseW,
		LoadVarianceW:     cfg.Simulator.LoadVarianceW,
		BatteryCapacityWh: cfg.Simulator.BatteryCapacityWh,
		MaxChargeRateW:    cfg.Simulator.MaxChargeRateW,
		MaxDischargeRateW: cfg.Simulator.MaxDischargeRateW,
	})

	// --- Initialize aggregator with simulated device states ---
	if cfg.Simulate {
		initSimulatedDevices(a.agg, cfg)
	}

	// --- Constraint System ---
	collector := resolver.NewConstraintCollector()
	res := resolver.NewResolver(collector)

	// --- Controllers ---
	cc := cfg.Controllers
	limitDisch := limit_discharge.New(limit_discharge.Config{
		MinSOC: cc.LimitDischarge.MinSOC, ForceChargeSOC: cc.LimitDischarge.ForceChargeSOC,
		ForceChargeW: cc.LimitDischarge.ForceChargeW, Enabled: cc.LimitDischarge.Enabled,
	})
	sellToGrid := sell_to_grid_limit.New(sell_to_grid_limit.Config{
		MaxSellToGridPower: cc.SellToGrid.MaxSellToGridPower, Enabled: cc.SellToGrid.Enabled,
	})
	peakShave := peak_shaving.New(peak_shaving.Config{
		PeakThresholdW: cc.PeakShaving.PeakThresholdW, Enabled: cc.PeakShaving.Enabled,
	})
	touCtrl := time_of_use.New(buildTOUConfig(cc))
	a.evCtrl = ev_charging.New(cc.EVCharging.MaxSitePowerW)
	a.evCtrl.SetEnabled(cc.EVCharging.Enabled)
	balanceCtrl := balancing.New(balancing.Config{Enabled: cc.Balancing.Enabled})

	sched := scheduler.New(limitDisch, sellToGrid, peakShave, touCtrl, a.evCtrl, balanceCtrl)
	advisors := []cycle.PowerAdvisor{peakShave, touCtrl, a.evCtrl, balanceCtrl}

	// --- UI Server ---
	authCfg := ui.AuthConfig{Username: cfg.UI.Username, Password: cfg.UI.Password}
	portStr := fmt.Sprintf("%d", cfg.UI.Port)
	a.uiServer = ui.NewServer(":"+portStr, a.db, a.monitor, authCfg)
	a.uiServer.SetAggregator(a.agg)
	if err := a.uiServer.Start(); err != nil {
		a.Close()
		return nil, fmt.Errorf("start UI server: %w", err)
	}

	// --- Cloud Sync ---
	syncInterval := time.Duration(cfg.Sync.IntervalSeconds) * time.Second
	if syncInterval <= 0 {
		syncInterval = 5 * time.Minute
	}
	if cfg.Sync.Enabled {
		a.setupCloudSync(cfg, syncInterval, touCtrl, cc)
	} else {
		stubPub := emsSync.NewStubPublisher()
		a.syncer = emsSync.NewSyncer(a.db, stubPub, cfg.Site.ID, cfg.Sync.TopicPrefix, syncInterval)
		log.Println("sync: disabled (using stub publisher)")
	}

	// --- Cycle Manager ---
	a.cm = cycle.New(sched, collector, res, advisors, time.Second)
	a.cm.OnCycle(a.onCycle)

	return a, nil
}

// Run starts all background services and blocks until ctx is cancelled.
func (a *App) Run(ctx context.Context) {
	cfg := a.cfg
	nBMS := len(cfg.Devices.BMS)
	nPCS := len(cfg.Devices.PCS)
	nMeter := len(cfg.Devices.Meter)
	nEV := len(cfg.Devices.EVCharger)
	totalDevices := nBMS + nPCS + nMeter + nEV
	portStr := fmt.Sprintf("%d", cfg.UI.Port)

	a.syncer.Start(ctx)
	defer a.syncer.Stop()

	if a.failoverManager != nil && !cfg.Simulate {
		go func() {
			for ctx.Err() == nil && a.failoverManager.State() == failover.HandoverPending {
				if err := a.failoverManager.RequestRecovery(); err != nil {
					log.Printf("failover: startup recovery: %v", err)
				}
				select {
				case <-ctx.Done():
					return
				case <-time.After(2 * time.Second):
				}
			}
		}()
	}
	if a.heartbeat != nil {
		a.heartbeat.Start(ctx)
		defer a.heartbeat.Stop()
	}
	if a.cmdSub != nil {
		go func() {
			for ctx.Err() == nil {
				if err := a.cmdSub.Start(); err == nil {
					return
				}
				time.Sleep(2 * time.Second)
			}
		}()
	}
	if a.failoverSubscribe != nil {
		go func() {
			for ctx.Err() == nil {
				if err := a.failoverSubscribe(); err == nil {
					return
				}
				time.Sleep(2 * time.Second)
			}
		}()
	}
	if a.deviceRuntime != nil {
		a.deviceRuntime.Start(ctx, cfg.StandardInterval())
	}

	fmt.Printf("▶ Starting EMS cycle loop with %d devices... (Ctrl+C to stop)\n", totalDevices)
	fmt.Printf("📊 Dashboard: http://localhost:%s\n\n", portStr)

	a.cm.Start(ctx, func() controllers.DeviceState {
		sys := a.agg.Aggregate()
		solarPower, loadPower := sys.SolarPowerW, sys.LoadPowerW
		if cfg.Simulate {
			simState := a.sim.NextState()
			solarPower, loadPower = simState.SolarPowerW, simState.LoadPowerW
		}
		return controllers.DeviceState{
			GridPowerW:       sys.GridPowerW,
			SolarPowerW:      solarPower,
			LoadPowerW:       loadPower,
			BatterySOC:       sys.TotalSOC,
			BatteryTempC:     sys.MaxCellTempC,
			ESSActivePowerW:  sys.TotalActivePowerW,
			CellVoltageMinMV: sys.MinCellVoltMV,
			CellVoltageMaxMV: sys.MaxCellVoltMV,
		}
	}, func(setpoint int) {
		if cfg.Simulate {
			a.sim.ApplyPower(setpoint)
		}
	})
}

// Close releases all resources.
func (a *App) Close() {
	if a.uiServer != nil {
		a.uiServer.Stop()
	}
	if a.ocppCentral != nil {
		a.ocppCentral.Stop()
	}
	if a.deviceRegistry != nil {
		a.deviceRegistry.Close()
	}
	if a.db != nil {
		a.db.Close()
	}
}

// onCycle is the callback invoked after each control cycle completes.
func (a *App) onCycle(result cycle.CycleResult) {
	cfg := a.cfg
	nBMS := len(cfg.Devices.BMS)
	nPCS := len(cfg.Devices.PCS)
	nMeter := len(cfg.Devices.Meter)
	nEV := len(cfg.Devices.EVCharger)

	a.cycleCount++
	a.monitor.RecordCycle()

	// Store to SQLite
	if err := a.db.InsertCycleResult(result); err != nil {
		log.Printf("DB insert error: %v", err)
	}

	// Feed to cloud sync accumulator
	a.syncer.AddCycle(result)

	// Distribute setpoint across PCS units
	sysState := a.agg.Aggregate()
	distResult := a.dist.Distribute(result.ResolvedPower.Setpoint, sysState.PCS)
	if cfg.Simulate {
		for i, sp := range distResult.Setpoints {
			pcs := sysState.PCS[i]
			pcs.SetpointW = sp
			pcs.ActivePowerW = sp
			a.agg.UpdatePCS(i, pcs)
		}
	} else if a.deviceRuntime != nil {
		setpoints := make(map[string]int, len(distResult.Setpoints))
		for i, sp := range distResult.Setpoints {
			if i < len(sysState.PCS) {
				setpoints[sysState.PCS[i].ID] = sp
			}
		}
		if err := a.deviceRuntime.ApplyPCSSetpoints(setpoints); err != nil {
			log.Printf("pcs command: %v", err)
		}
	}

	// Update simulated device states
	if cfg.Simulate {
		updateSimulatedBMS(a.agg, sysState, result, nBMS, a.cycleCount)
		updateSimulatedMeters(a.agg, sysState, result, nMeter)
	}

	// Update EV charger simulation / DLM
	totalEVPower := 0
	demands := make([]ev_charging.ChargerDemand, 0, nEV)
	for _, ev := range sysState.EVCharger {
		demand := ev.MaxPowerLimit
		if demand <= 0 {
			demand = 22000
		}
		demands = append(demands, ev_charging.ChargerDemand{ID: ev.ID, DemandW: demand, Online: ev.Online, Charging: ev.Status == 1})
	}
	nonEVLoad := sysState.GridPowerW - sysState.TotalEVPowerW
	limits := a.evCtrl.AllocateLimits(nonEVLoad, demands)
	if cfg.Simulate {
		totalEVPower = updateSimulatedEVChargers(a.agg, sysState, limits, nEV)
	} else if a.deviceRuntime != nil {
		if err := a.deviceRuntime.ApplyEVLimits(limits); err != nil {
			log.Printf("ev dlm command: %v", err)
		}
		for _, ev := range sysState.EVCharger {
			totalEVPower += ev.ActivePowerW
		}
	}
	a.evCtrl.SetEVTotalPower(totalEVPower)

	// Broadcast to WebSocket clients
	a.uiServer.BroadcastCycleResult(result)

	// Console output (every 5th cycle)
	if a.cycleCount%5 == 0 {
		constraintStr := "-"
		if len(result.ResolvedPower.AppliedConstraints) > 0 {
			constraintStr = fmt.Sprintf("%v", result.ResolvedPower.AppliedConstraints)
		}
		fmt.Printf("[%s] ☀%6dW | ⚡%6dW | 🔌%6dW | 🔋%5.1f%% | ESS:%6dW | BMS:%d PCS:%d | %s\n",
			result.Timestamp.Format("15:04:05"),
			result.State.SolarPowerW, result.State.LoadPowerW, result.State.GridPowerW,
			result.State.BatterySOC, result.ResolvedPower.Setpoint,
			nBMS, nPCS, constraintStr,
		)
	}

	for _, err := range result.Errors {
		a.monitor.RecordError()
		fmt.Printf("  ⚠ ERROR: %v\n", err)
	}
}

// setupCloudSync wires MQTT publisher, syncer, command subscriber, heartbeat, and failover.
func (a *App) setupCloudSync(cfg *config.Config, syncInterval time.Duration, touCtrl *time_of_use.Controller, cc config.ControllersConfig) {
	mqttPub := emsSync.NewMQTTPublisher(emsSync.MQTTOptions{
		Broker: cfg.Sync.MQTTBroker, ClientID: cfg.Sync.ClientID,
		Username: cfg.Sync.Username, Password: cfg.Sync.Password,
		CACertPath: cfg.Sync.CACertPath, Insecure: cfg.Sync.Insecure, QoS: 1,
	})

	a.syncer = emsSync.NewSyncer(a.db, mqttPub, cfg.Site.ID, cfg.Sync.TopicPrefix, syncInterval)
	a.cmdSub = emsSync.NewCommandSubscriber(mqttPub, cfg.Site.ID, cfg.Sync.TopicPrefix)
	a.failoverManager = failover.NewFailoverManager(mqttPub, cfg.Site.ID, cfg.Sync.TopicPrefix)
	if !cfg.Simulate {
		a.failoverManager.BeginStartupRecovery()
	}
	a.heartbeat = failover.NewHeartbeatPublisher(mqttPub, cfg.Site.ID, cfg.Sync.TopicPrefix, 30*time.Second)
	a.heartbeat.SetFailoverProvider(func() (string, string, uint64) {
		return a.failoverManager.ControlOwner(), a.failoverManager.State().String(), a.failoverManager.LeaseEpoch()
	})
	a.heartbeat.SetEVProvider(func() []failover.EVChargerInfo {
		state := a.agg.Aggregate()
		result := make([]failover.EVChargerInfo, 0, len(state.EVCharger))
		for _, ev := range state.EVCharger {
			result = append(result, failover.EVChargerInfo{ID: ev.ID, Status: ev.Status, ActivePowerW: ev.ActivePowerW, VehicleConnected: ev.VehicleConnected, Protocol: "ocpp"})
		}
		return result
	})
	a.failoverSubscribe = func() error {
		topic := fmt.Sprintf("%s/%s/failover/control", cfg.Sync.TopicPrefix, cfg.Site.ID)
		return mqttPub.Subscribe(topic, func(_ string, payload []byte) {
			var command struct {
				Action string `json:"action"`
			}
			if err := json.Unmarshal(payload, &command); err != nil || command.Action == "" {
				log.Printf("failover: invalid control message")
				return
			}
			a.failoverManager.HandleCloudMessage(command.Action, payload)
		})
	}

	registerCommandHandlers(a.cmdSub, cfg, touCtrl, a.evCtrl, cc)

	log.Printf("sync: enabled — broker=%s interval=%s topic=%s/%s/telemetry",
		cfg.Sync.MQTTBroker, syncInterval, cfg.Sync.TopicPrefix, cfg.Site.ID)
}

// buildTOUConfig maps YAML TOU period config to the time_of_use.Config struct.
func buildTOUConfig(cc config.ControllersConfig) time_of_use.Config {
	var touPeriods []time_of_use.TariffPeriod
	for _, p := range cc.TimeOfUse.Periods {
		rate := time_of_use.MidPeak
		switch p.Rate {
		case "off_peak":
			rate = time_of_use.OffPeak
		case "on_peak":
			rate = time_of_use.OnPeak
		}
		touPeriods = append(touPeriods, time_of_use.TariffPeriod{
			StartHour: p.StartHour, EndHour: p.EndHour, Rate: rate,
		})
	}
	return time_of_use.Config{
		Periods: touPeriods, MaxChargeW: cc.TimeOfUse.MaxChargeW,
		MaxDischargeW: cc.TimeOfUse.MaxDischargeW, MinSOCToDischarge: cc.TimeOfUse.MinSOCToDischarge,
		Enabled: cc.TimeOfUse.Enabled,
	}
}

// registerCommandHandlers sets up remote command handlers for identify, restart,
// update_tariff, and update_config.
func registerCommandHandlers(cmdSub *emsSync.CommandSubscriber, cfg *config.Config, touCtrl *time_of_use.Controller, evCtrl *ev_charging.EVChargingController, cc config.ControllersConfig) {
	cmdSub.OnCommand("identify", func(cmd emsSync.RemoteCommand) error {
		log.Printf("command: identify — site=%s version=%s", cfg.Site.ID, build.Version)
		return nil
	})
	cmdSub.OnCommand("restart", func(cmd emsSync.RemoteCommand) error {
		log.Printf("command: restart accepted (simulator demo does not restart the process)")
		return nil
	})
	cmdSub.OnCommand("update_tariff", func(cmd emsSync.RemoteCommand) error {
		start, okStart := cmd.Payload["off_peak_start"].(float64)
		end, okEnd := cmd.Payload["off_peak_end"].(float64)
		if !okStart || !okEnd || start < 0 || start > 23 || end < 0 || end > 23 {
			return fmt.Errorf("invalid tariff hours")
		}
		updated := time_of_use.Config{Periods: []time_of_use.TariffPeriod{{StartHour: int(start), EndHour: int(end), Rate: time_of_use.OffPeak}}, MaxChargeW: cc.TimeOfUse.MaxChargeW, MaxDischargeW: cc.TimeOfUse.MaxDischargeW, MinSOCToDischarge: cc.TimeOfUse.MinSOCToDischarge, Enabled: true}
		touCtrl.UpdateConfig(updated)
		return nil
	})
	cmdSub.OnCommand("update_config", func(cmd emsSync.RemoteCommand) error {
		allowed := map[string]bool{"ev_max_site_power_w": true, "ev_enabled": true}
		for key := range cmd.Payload {
			if !allowed[key] {
				return fmt.Errorf("runtime config key %q is not hot-reloadable", key)
			}
		}
		maxPower := cc.EVCharging.MaxSitePowerW
		if raw, exists := cmd.Payload["ev_max_site_power_w"]; exists {
			value, ok := raw.(float64)
			if !ok || value <= 0 || value > 10_000_000 {
				return fmt.Errorf("invalid ev_max_site_power_w")
			}
			maxPower = int(value)
		}
		enabled := cc.EVCharging.Enabled
		if raw, exists := cmd.Payload["ev_enabled"]; exists {
			value, ok := raw.(bool)
			if !ok {
				return fmt.Errorf("invalid ev_enabled")
			}
			enabled = value
		}
		if err := evCtrl.UpdateConfig(maxPower, enabled); err != nil {
			return err
		}
		cc.EVCharging.MaxSitePowerW, cc.EVCharging.Enabled = maxPower, enabled
		log.Printf("command: runtime EV config applied atomically max_site_power_w=%d enabled=%t", maxPower, enabled)
		return nil
	})
}

// initSimulatedDevices populates the aggregator with initial simulated device states.
func initSimulatedDevices(agg *aggregator.Aggregator, cfg *config.Config) {
	for i, d := range cfg.Devices.BMS {
		agg.UpdateBMS(i, aggregator.BMSState{
			ID: d.ID, SOC: 70 + float64(i)*5, CapacityWh: 5000,
			MaxChargeW: 2500, MaxDischargeW: 5000,
			TempC: 25 + float64(i)*2, MinCellMV: 3200, MaxCellMV: 3400,
			Online: true, UpdatedAt: time.Now(),
		})
	}
	for i, d := range cfg.Devices.PCS {
		agg.UpdatePCS(i, aggregator.PCSState{
			ID: d.ID, MaxPowerW: 5000, Status: aggregator.PCSRunning,
			Online: true, UpdatedAt: time.Now(),
		})
	}
	for i, d := range cfg.Devices.Meter {
		agg.UpdateMeter(i, aggregator.MeterState{
			ID: d.ID, Role: d.Role, Online: true, FrequencyHz: 50.0,
			VoltageV: 400, UpdatedAt: time.Now(),
		})
	}
	for i, d := range cfg.Devices.EVCharger {
		status := 0
		power := 0
		connected := false
		if i%2 == 0 {
			status = 1
			power = 7000 + rand.Intn(4000)
			connected = true
		}
		agg.UpdateEVCharger(i, aggregator.EVChargerState{
			ID: d.ID, Status: status, ActivePowerW: power,
			MaxPowerLimit: 22000, VehicleConnected: connected,
			Online: true, UpdatedAt: time.Now(),
		})
	}
}

// updateSimulatedBMS updates simulated BMS SOC and temperature each cycle.
func updateSimulatedBMS(agg *aggregator.Aggregator, sysState aggregator.SystemState, result cycle.CycleResult, nBMS int, cycleCount int) {
	for i := 0; i < nBMS; i++ {
		bms := sysState.BMS[i]
		bmsPower := float64(result.ResolvedPower.Setpoint) / float64(nBMS)
		socDelta := -bmsPower / float64(bms.CapacityWh) * 100.0 / 3600.0
		bms.SOC = math.Max(0, math.Min(100, bms.SOC+socDelta))
		bms.TempC = 25 + 5*math.Sin(float64(cycleCount)/200.0) + float64(i)*2 + rand.Float64()
		bms.MinCellMV = 3100 + int(bms.SOC*5)
		bms.MaxCellMV = bms.MinCellMV + 50 + rand.Intn(30)
		bms.UpdatedAt = time.Now()
		agg.UpdateBMS(i, bms)
	}
}

// updateSimulatedMeters updates simulated meter readings each cycle.
func updateSimulatedMeters(agg *aggregator.Aggregator, sysState aggregator.SystemState, result cycle.CycleResult, nMeter int) {
	for i := 0; i < nMeter; i++ {
		m := sysState.Meter[i]
		if m.Role == "grid" {
			m.ActivePowerW = result.State.GridPowerW
		} else if m.Role == "solar" {
			m.ActivePowerW = result.State.SolarPowerW
		}
		m.FrequencyHz = 50.0 + (rand.Float64()-0.5)*0.1
		m.VoltageV = 400 + (rand.Float64()-0.5)*10
		m.CurrentA = float64(m.ActivePowerW) / m.VoltageV
		m.UpdatedAt = time.Now()
		agg.UpdateMeter(i, m)
	}
}

// updateSimulatedEVChargers updates simulated EV charger states and returns total EV power.
func updateSimulatedEVChargers(agg *aggregator.Aggregator, sysState aggregator.SystemState, limits map[string]int, nEV int) int {
	totalEVPower := 0
	for i := 0; i < nEV; i++ {
		ev := sysState.EVCharger[i]
		ev.MaxPowerLimit = limits[ev.ID]
		if ev.Status == 1 { // Charging
			ev.ActivePowerW = 7000 + rand.Intn(4000)
			if ev.ActivePowerW > ev.MaxPowerLimit {
				ev.ActivePowerW = ev.MaxPowerLimit
			}
			ev.EnergyDelivered += ev.ActivePowerW / 3600
			ev.CurrentL1 = float64(ev.ActivePowerW) / 230.0
			ev.CurrentL2 = ev.CurrentL1 * (0.95 + rand.Float64()*0.1)
			ev.CurrentL3 = ev.CurrentL1 * (0.95 + rand.Float64()*0.1)
			if rand.Intn(300) == 0 {
				ev.Status = 4
				ev.ActivePowerW = 0
			}
		} else if ev.Status == 0 && rand.Intn(120) == 0 {
			ev.Status = 1
			ev.VehicleConnected = true
			ev.EnergyDelivered = 0
		} else if ev.Status == 4 {
			ev.ActivePowerW = 0
			ev.CurrentL1 = 0
			ev.CurrentL2 = 0
			ev.CurrentL3 = 0
			if rand.Intn(60) == 0 {
				ev.Status = 0
				ev.VehicleConnected = false
			}
		}
		totalEVPower += ev.ActivePowerW
		ev.UpdatedAt = time.Now()
		agg.UpdateEVCharger(i, ev)
	}
	return totalEVPower
}
