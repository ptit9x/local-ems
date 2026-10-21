package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
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

func main() {
	simulate := flag.Bool("simulate", false, "Run with simulated hardware")
	configPath := flag.String("config", "", "Path to YAML config file")
	uiPort := flag.String("ui-port", "", "HTTP port override")
	dbPath := flag.String("db", "", "Database path override")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("ems-core %s (%s) %s/%s\n", build.Version, build.BuildTime, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	// --- Load Config ---
	var cfg *config.Config
	if *configPath != "" {
		var err error
		cfg, err = config.Load(*configPath)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
	} else {
		cfg = config.DefaultDev()
	}

	// CLI overrides
	if *simulate {
		cfg.Simulate = true
	}
	if *uiPort != "" {
		// parse int
		fmt.Sscanf(*uiPort, "%d", &cfg.UI.Port)
	}
	if *dbPath != "" {
		cfg.Storage.DSN = *dbPath
	}

	nBMS := len(cfg.Devices.BMS)
	nPCS := len(cfg.Devices.PCS)
	nMeter := len(cfg.Devices.Meter)
	nEV := len(cfg.Devices.EVCharger)
	totalDevices := nBMS + nPCS + nMeter + nEV

	portStr := fmt.Sprintf("%d", cfg.UI.Port)

	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Printf("║   Local EMS %-11s  %-7s/%-22s║\n", build.Version, runtime.GOOS, runtime.GOARCH)
	fmt.Printf("║   %d BMS, %d PCS, %d Meter, %d EV Charger (total %d)       ║\n", nBMS, nPCS, nMeter, nEV, totalDevices)
	fmt.Printf("║   Dashboard: http://localhost:%-27s║\n", portStr)
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// --- Storage ---
	var storeOpts []store.Option
	if cfg.Storage.WALMode {
		storeOpts = append(storeOpts, store.WithWAL())
	}
	if cfg.Storage.RetentionDays > 0 {
		storeOpts = append(storeOpts, store.WithAutoRetention(cfg.Storage.RetentionDays))
	}
	db, err := store.New(cfg.Storage.DSN, storeOpts...)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// --- Health Monitor ---
	monitor := health.NewMonitor()

	// --- Aggregator ---
	agg := aggregator.New(nBMS, nPCS, nMeter, nEV)

	// --- Physical device runtime (simulation continues to use in-memory devices) ---
	var failoverManager *failover.FailoverManager
	var deviceRegistry *devices.DeviceRegistry
	var deviceRuntime *device_runtime.Orchestrator
	var ocppCentral *ocpp.CentralSystem
	if !cfg.Simulate {
		hasOCPP := false
		for _, charger := range cfg.Devices.EVCharger {
			if charger.Protocol == "ocpp" {
				hasOCPP = true
				break
			}
		}
		if hasOCPP {
			ocppCentral = ocpp.NewCentralSystem(cfg.OCPP.ListenAddress, nil)
			if err := ocppCentral.Start(); err != nil {
				log.Fatalf("Failed to start Local OCPP Central System: %v", err)
			}
			defer ocppCentral.Stop()
			log.Printf("ocpp: Local Central System listening on %s", cfg.OCPP.ListenAddress)
		}
		factory := adapters.NewFactory(ocppCentral)
		var err error
		deviceRegistry, err = factory.Build(cfg)
		if err != nil {
			log.Fatalf("Failed to build physical device registry: %v", err)
		}
		defer deviceRegistry.Close()
		deviceRuntime = device_runtime.New(deviceRegistry, agg, cfg, func() bool {
			return failoverManager == nil || failoverManager.ControlOwner() == "local"
		})
	}

	// --- Distributor ---
	dist := distributor.New(distributor.Strategy(cfg.Engine.Distribution), nPCS)

	// --- Base Simulator (for time-of-day solar/load patterns) ---
	sim := simulator.New(simulator.Config{
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
			status := 0 // Available by default
			power := 0
			connected := false
			if i%2 == 0 { // Simulate: half are charging
				status = 1                     // Charging
				power = 7000 + rand.Intn(4000) // 7-11 kW
				connected = true
			}
			agg.UpdateEVCharger(i, aggregator.EVChargerState{
				ID: d.ID, Status: status, ActivePowerW: power,
				MaxPowerLimit: 22000, VehicleConnected: connected,
				Online: true, UpdatedAt: time.Now(),
			})
		}
	}

	// --- Constraint System ---
	collector := resolver.NewConstraintCollector()
	res := resolver.NewResolver(collector)

	// --- Controllers (all config-driven via YAML + .env) ---
	cc := cfg.Controllers

	limitDisch := limit_discharge.New(limit_discharge.Config{
		MinSOC:         cc.LimitDischarge.MinSOC,
		ForceChargeSOC: cc.LimitDischarge.ForceChargeSOC,
		ForceChargeW:   cc.LimitDischarge.ForceChargeW,
		Enabled:        cc.LimitDischarge.Enabled,
	})

	sellToGrid := sell_to_grid_limit.New(sell_to_grid_limit.Config{
		MaxSellToGridPower: cc.SellToGrid.MaxSellToGridPower,
		Enabled:            cc.SellToGrid.Enabled,
	})

	peakShave := peak_shaving.New(peak_shaving.Config{
		PeakThresholdW: cc.PeakShaving.PeakThresholdW,
		Enabled:        cc.PeakShaving.Enabled,
	})

	// Map TOU period config strings → TariffRate enum
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
			StartHour: p.StartHour,
			EndHour:   p.EndHour,
			Rate:      rate,
		})
	}
	touCtrl := time_of_use.New(time_of_use.Config{
		Periods:           touPeriods,
		MaxChargeW:        cc.TimeOfUse.MaxChargeW,
		MaxDischargeW:     cc.TimeOfUse.MaxDischargeW,
		MinSOCToDischarge: cc.TimeOfUse.MinSOCToDischarge,
		Enabled:           cc.TimeOfUse.Enabled,
	})

	evCtrl := ev_charging.New(cc.EVCharging.MaxSitePowerW)
	evCtrl.SetEnabled(cc.EVCharging.Enabled)

	balanceCtrl := balancing.New(balancing.Config{
		Enabled: cc.Balancing.Enabled,
	})

	sched := scheduler.New(
		limitDisch,  // P1: Hardware protection
		sellToGrid,  // P2: Grid compliance
		peakShave,   // P3: Cost optimization (advisory)
		touCtrl,     // P3: Cost optimization (advisory)
		evCtrl,      // P3: EV Dynamic Load Management
		balanceCtrl, // P4: Default self-consumption
	)

	advisors := []cycle.PowerAdvisor{peakShave, touCtrl, evCtrl, balanceCtrl}

	// --- UI Server ---
	authCfg := ui.AuthConfig{
		Username: cfg.UI.Username,
		Password: cfg.UI.Password,
	}
	uiServer := ui.NewServer(":"+portStr, db, monitor, authCfg)
	uiServer.SetAggregator(agg)
	if err := uiServer.Start(); err != nil {
		log.Fatalf("Failed to start UI server: %v", err)
	}
	defer uiServer.Stop()

	// --- Cloud Sync (MQTT store-and-forward) ---
	var syncer *emsSync.Syncer
	var cmdSub *emsSync.CommandSubscriber
	var heartbeat *failover.HeartbeatPublisher
	var failoverSubscribe func() error
	syncInterval := time.Duration(cfg.Sync.IntervalSeconds) * time.Second
	if syncInterval <= 0 {
		syncInterval = 5 * time.Minute
	}

	if cfg.Sync.Enabled {
		mqttPub := emsSync.NewMQTTPublisher(emsSync.MQTTOptions{
			Broker:     cfg.Sync.MQTTBroker,
			ClientID:   cfg.Sync.ClientID,
			Username:   cfg.Sync.Username,
			Password:   cfg.Sync.Password,
			CACertPath: cfg.Sync.CACertPath,
			Insecure:   cfg.Sync.Insecure,
			QoS:        1,
		})

		syncer = emsSync.NewSyncer(db, mqttPub, cfg.Site.ID, cfg.Sync.TopicPrefix, syncInterval)

		// Register command handlers
		cmdSub = emsSync.NewCommandSubscriber(mqttPub, cfg.Site.ID, cfg.Sync.TopicPrefix)
		failoverManager = failover.NewFailoverManager(mqttPub, cfg.Site.ID, cfg.Sync.TopicPrefix)
		if !cfg.Simulate {
			failoverManager.BeginStartupRecovery()
		}
		heartbeat = failover.NewHeartbeatPublisher(mqttPub, cfg.Site.ID, cfg.Sync.TopicPrefix, 30*time.Second)
		heartbeat.SetFailoverProvider(func() (string, string, uint64) {
			return failoverManager.ControlOwner(), failoverManager.State().String(), failoverManager.LeaseEpoch()
		})
		heartbeat.SetEVProvider(func() []failover.EVChargerInfo {
			state := agg.Aggregate()
			result := make([]failover.EVChargerInfo, 0, len(state.EVCharger))
			for _, ev := range state.EVCharger {
				result = append(result, failover.EVChargerInfo{ID: ev.ID, Status: ev.Status, ActivePowerW: ev.ActivePowerW, VehicleConnected: ev.VehicleConnected, Protocol: "ocpp"})
			}
			return result
		})
		failoverSubscribe = func() error {
			topic := fmt.Sprintf("%s/%s/failover/control", cfg.Sync.TopicPrefix, cfg.Site.ID)
			return mqttPub.Subscribe(topic, func(_ string, payload []byte) {
				var command struct {
					Action string `json:"action"`
				}
				if err := json.Unmarshal(payload, &command); err != nil || command.Action == "" {
					log.Printf("failover: invalid control message")
					return
				}
				failoverManager.HandleCloudMessage(command.Action, payload)
			})
		}
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

		log.Printf("sync: enabled — broker=%s interval=%s topic=%s/%s/telemetry",
			cfg.Sync.MQTTBroker, syncInterval, cfg.Sync.TopicPrefix, cfg.Site.ID)
	} else {
		stubPub := emsSync.NewStubPublisher()
		syncer = emsSync.NewSyncer(db, stubPub, cfg.Site.ID, cfg.Sync.TopicPrefix, syncInterval)
		log.Println("sync: disabled (using stub publisher)")
	}

	// --- Cycle Manager ---
	cm := cycle.New(sched, collector, res, advisors, time.Second)

	cycleCount := 0
	cm.OnCycle(func(result cycle.CycleResult) {
		cycleCount++
		monitor.RecordCycle()

		// Store to SQLite
		if err := db.InsertCycleResult(result); err != nil {
			log.Printf("DB insert error: %v", err)
		}

		// Feed to cloud sync accumulator
		syncer.AddCycle(result)

		// Distribute setpoint across PCS units
		sysState := agg.Aggregate()
		distResult := dist.Distribute(result.ResolvedPower.Setpoint, sysState.PCS)
		if cfg.Simulate {
			for i, sp := range distResult.Setpoints {
				pcs := sysState.PCS[i]
				pcs.SetpointW = sp
				pcs.ActivePowerW = sp // in simulation, output = setpoint
				agg.UpdatePCS(i, pcs)
			}
		} else if deviceRuntime != nil {
			setpoints := make(map[string]int, len(distResult.Setpoints))
			for i, sp := range distResult.Setpoints {
				if i < len(sysState.PCS) {
					setpoints[sysState.PCS[i].ID] = sp
				}
			}
			if err := deviceRuntime.ApplyPCSSetpoints(setpoints); err != nil {
				log.Printf("pcs command: %v", err)
			}
		}

		// Update BMS simulation (SOC drift, temp drift)
		if cfg.Simulate {
			for i := 0; i < nBMS; i++ {
				bms := sysState.BMS[i]
				// Each BMS handles a portion of total ESS power
				bmsPower := float64(result.ResolvedPower.Setpoint) / float64(nBMS)
				socDelta := -bmsPower / float64(bms.CapacityWh) * 100.0 / 3600.0
				bms.SOC = math.Max(0, math.Min(100, bms.SOC+socDelta))
				bms.TempC = 25 + 5*math.Sin(float64(cycleCount)/200.0) + float64(i)*2 + rand.Float64()
				bms.MinCellMV = 3100 + int(bms.SOC*5)
				bms.MaxCellMV = bms.MinCellMV + 50 + rand.Intn(30)
				bms.UpdatedAt = time.Now()
				agg.UpdateBMS(i, bms)
			}

			// Update meters
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

		// Update EV charger simulation
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
		limits := evCtrl.AllocateLimits(nonEVLoad, demands)
		if cfg.Simulate {
			for i := 0; i < nEV; i++ {
				ev := sysState.EVCharger[i]
				ev.MaxPowerLimit = limits[ev.ID]
				// Simulate charging dynamics
				if ev.Status == 1 { // Charging
					ev.ActivePowerW = 7000 + rand.Intn(4000) // 7-11 kW variation
					if ev.ActivePowerW > ev.MaxPowerLimit {
						ev.ActivePowerW = ev.MaxPowerLimit
					}
					ev.EnergyDelivered += ev.ActivePowerW / 3600 // Wh per second
					ev.CurrentL1 = float64(ev.ActivePowerW) / 230.0
					ev.CurrentL2 = ev.CurrentL1 * (0.95 + rand.Float64()*0.1)
					ev.CurrentL3 = ev.CurrentL1 * (0.95 + rand.Float64()*0.1)
					// Randomly finish charging (1 in 300 cycles ≈ every 5 min)
					if rand.Intn(300) == 0 {
						ev.Status = 4 // Finishing
						ev.ActivePowerW = 0
					}
				} else if ev.Status == 0 && rand.Intn(120) == 0 {
					// Available → vehicle arrives and starts charging
					ev.Status = 1
					ev.VehicleConnected = true
					ev.EnergyDelivered = 0
				} else if ev.Status == 4 {
					ev.ActivePowerW = 0
					ev.CurrentL1 = 0
					ev.CurrentL2 = 0
					ev.CurrentL3 = 0
					// Finishing → vehicle disconnects
					if rand.Intn(60) == 0 {
						ev.Status = 0
						ev.VehicleConnected = false
					}
				}
				totalEVPower += ev.ActivePowerW
				ev.UpdatedAt = time.Now()
				agg.UpdateEVCharger(i, ev)
			}
		} else if deviceRuntime != nil {
			if err := deviceRuntime.ApplyEVLimits(limits); err != nil {
				log.Printf("ev dlm command: %v", err)
			}
			for _, ev := range sysState.EVCharger {
				totalEVPower += ev.ActivePowerW
			}
		}
		// Feed EV total power to DLM controller
		evCtrl.SetEVTotalPower(totalEVPower)

		// Broadcast to WebSocket clients (include per-device info)
		uiServer.BroadcastCycleResult(result)

		// Console output (every 5th cycle)
		if cycleCount%5 == 0 {
			constraintStr := "-"
			if len(result.ResolvedPower.AppliedConstraints) > 0 {
				constraintStr = fmt.Sprintf("%v", result.ResolvedPower.AppliedConstraints)
			}

			fmt.Printf("[%s] ☀%6dW | ⚡%6dW | 🔌%6dW | 🔋%5.1f%% | ESS:%6dW | BMS:%d PCS:%d | %s\n",
				result.Timestamp.Format("15:04:05"),
				result.State.SolarPowerW,
				result.State.LoadPowerW,
				result.State.GridPowerW,
				result.State.BatterySOC,
				result.ResolvedPower.Setpoint,
				nBMS, nPCS,
				constraintStr,
			)
		}

		for _, err := range result.Errors {
			monitor.RecordError()
			fmt.Printf("  ⚠ ERROR: %v\n", err)
		}
	})

	// --- Graceful Shutdown ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\n⏹ Shutting down EMS...")
		cancel()
	}()

	// --- Run ---
	syncer.Start(ctx)
	defer syncer.Stop()
	if failoverManager != nil && !cfg.Simulate {
		go func() {
			for ctx.Err() == nil && failoverManager.State() == failover.HandoverPending {
				if err := failoverManager.RequestRecovery(); err != nil {
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
	if heartbeat != nil {
		heartbeat.Start(ctx)
		defer heartbeat.Stop()
	}
	if cmdSub != nil {
		go func() {
			for ctx.Err() == nil {
				if err := cmdSub.Start(); err == nil {
					return
				}
				time.Sleep(2 * time.Second)
			}
		}()
	}
	if failoverSubscribe != nil {
		go func() {
			for ctx.Err() == nil {
				if err := failoverSubscribe(); err == nil {
					return
				}
				time.Sleep(2 * time.Second)
			}
		}()
	}
	if deviceRuntime != nil {
		deviceRuntime.Start(ctx, cfg.StandardInterval())
	}

	fmt.Printf("▶ Starting EMS cycle loop with %d devices... (Ctrl+C to stop)\n", totalDevices)
	fmt.Printf("📊 Dashboard: http://localhost:%s\n\n", portStr)

	cm.Start(ctx, func() controllers.DeviceState {
		// Get aggregated state and map to DeviceState for controllers
		sys := agg.Aggregate()
		solarPower, loadPower := sys.SolarPowerW, sys.LoadPowerW
		if cfg.Simulate {
			simState := sim.NextState()
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
			sim.ApplyPower(setpoint)
		}
	})
}
