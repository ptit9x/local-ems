package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/vmo/ems-platform/ems-local/internal/app"
	"github.com/vmo/ems-platform/ems-local/internal/build"
	"github.com/vmo/ems-platform/ems-local/internal/config"
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

	// --- Wire and start application ---
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	defer application.Close()

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

	application.Run(ctx)
}
