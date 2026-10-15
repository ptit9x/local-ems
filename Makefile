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

# ---- Cloud EMS (NestJS + React, Phase B) ----

.PHONY: dev-cloud
dev-cloud:
	@echo "⏳ ems-cloud not yet initialized. See README.md for roadmap."

.PHONY: test-cloud
test-cloud:
	@echo "⏳ ems-cloud not yet initialized. See README.md for roadmap."

# ---- All ----

.PHONY: test-all
test-all: test-local test-cloud

.PHONY: clean
clean: clean-local
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
	@echo "    make clean          Clean all build artifacts"
	@echo ""
