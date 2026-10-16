# ============================================================
#  EMS Platform — Root Makefile (Orchestrator)
# ============================================================
#  Delegates to sub-project Makefiles.
#
#  Usage:
#    make                → build ems-local
#    make test           → test ems-local
#    make run            → run ems-local in simulation mode
#    make build-edge     → build ems-local for current OS
#    make help           → show all targets
# ============================================================

# ---- Local EMS (Go on-premise controller) ----

.PHONY: build build-edge
build build-edge:
	@$(MAKE) -C ems-local build

.PHONY: test test-local
test test-local:
	@$(MAKE) -C ems-local test

.PHONY: run dev-local
run dev-local:
	@$(MAKE) -C ems-local run

.PHONY: run-smart-home
run-smart-home:
	@$(MAKE) -C ems-local run-smart-home

.PHONY: run-factory
run-factory:
	@$(MAKE) -C ems-local run-factory

.PHONY: release
release:
	@$(MAKE) -C ems-local release

.PHONY: clean-local
clean-local:
	@$(MAKE) -C ems-local clean

# ---- Cloud EMS (NestJS + React) ----

.PHONY: dev-cloud
dev-cloud:
	@cd ems-cloud && docker compose up -d && npm run dev

.PHONY: test-cloud
test-cloud:
	@cd ems-cloud && npm run db:generate && npm run build && npm test

.PHONY: test-contracts
test-contracts:
	@cd ems-cloud && npm run test:contracts

.PHONY: clean-cloud
clean-cloud:
	@cd ems-cloud && docker compose down

# ---- All ----

.PHONY: test-all
test-all: test-local test-contracts test-cloud

.PHONY: demo
demo:
	@cd ems-cloud && docker compose up -d --build
	@echo "Dashboard: http://localhost:5173  API: http://localhost:3000/api"
	@echo "Login: admin@demo.local / demo1234"
	@cd ems-local && mkdir -p data && EMS_SYNC_ENABLED=true EMS_SYNC_INTERVAL_SEC=10 EMS_MQTT_BROKER=tcp://localhost:1883 go run ./cmd/ems-core --simulate --config config/cdi-singapore.yaml --db data/full-sim.db

.PHONY: demo-down
demo-down:
	@cd ems-cloud && docker compose down

.PHONY: demo-check
demo-check:
	@cd ems-cloud && npm run test:smoke

.PHONY: demo-reset
demo-reset:
	@echo "WARNING: deleting simulator database volumes and local demo data"
	@cd ems-cloud && docker compose down -v
	@rm -f ems-local/data/full-sim.db ems-local/data/full-sim.db-shm ems-local/data/full-sim.db-wal

.PHONY: clean
clean: clean-local clean-cloud
	@echo "✅ All clean"

# ---- Help ----

.PHONY: help
help:
	@echo ""
	@echo "EMS Platform — Top-Level Commands"
	@echo "══════════════════════════════════════════════════════════"
	@echo "  Local EMS (Go):"
	@echo "    make build          Build ems-local for current OS"
	@echo "    make test           Run ems-local tests"
	@echo "    make run            Run ems-local in simulation mode"
	@echo "    make run-smart-home Run (smart home profile)"
	@echo "    make run-factory    Run (factory profile)"
	@echo "    make release        Package installer .zips"
	@echo "    make clean-local    Remove ems-local/dist/"
	@echo "  ──────────────────────────────────────────────────────"
	@echo "  Cloud EMS (Phase B):"
	@echo "    make dev-cloud      Start cloud dev server (coming soon)"
	@echo "    make test-cloud     Run cloud tests (coming soon)"
	@echo "  ──────────────────────────────────────────────────────"
	@echo "  All:"
	@echo "    make test-all       Run all tests (local + cloud)"
	@echo "    make demo           Start full simulator demo"
	@echo "    make demo-down      Stop demo services"
	@echo "    make demo-check     Smoke-test a running demo"
	@echo "    make demo-reset     Delete all demo data"
	@echo "    make clean          Clean all build artifacts"
	@echo ""
