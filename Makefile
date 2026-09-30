# ============================================================
#  Local EMS — Makefile
# ============================================================
#  Usage:
#    make              → build for current OS
#    make all          → build all platforms
#    make linux        → build Linux x86_64
#    make linux-arm    → build Linux ARM64 (Raspberry Pi)
#    make windows      → build Windows .exe
#    make macos        → build macOS (Apple Silicon)
#    make macos-intel  → build macOS (Intel)
#    make docker-all   → cross-compile all via Docker (recommended)
#    make clean        → remove dist/ directory
#    make test         → run tests
#    make run          → run simulation mode
# ============================================================

APP_NAME    := ems-core
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME  := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
MODULE      := github.com/vmo/local-ems
MAIN_PKG    := ./cmd/ems-core
DIST_DIR    := dist

# Linker flags: bake version info into binary
LDFLAGS := -s -w \
  -X '$(MODULE)/internal/build.Version=$(VERSION)' \
  -X '$(MODULE)/internal/build.BuildTime=$(BUILD_TIME)'

# ---- Default: build for current OS ----

.PHONY: build
build:
	@echo "▶ Building $(APP_NAME) for $$(go env GOOS)/$$(go env GOARCH)..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=1 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME) $(MAIN_PKG)
	@echo "✅ $(DIST_DIR)/$(APP_NAME)"

# ---- Platform-specific (native, requires local C compiler) ----

.PHONY: linux
linux:
	@echo "▶ Building $(APP_NAME) for linux/amd64..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
		go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_PKG)
	@echo "✅ $(DIST_DIR)/$(APP_NAME)-linux-amd64"

.PHONY: linux-arm
linux-arm:
	@echo "▶ Building $(APP_NAME) for linux/arm64 (Raspberry Pi)..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 \
		CC=aarch64-linux-gnu-gcc \
		go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-linux-arm64 $(MAIN_PKG)
	@echo "✅ $(DIST_DIR)/$(APP_NAME)-linux-arm64"

.PHONY: windows
windows:
	@echo "▶ Building $(APP_NAME) for windows/amd64..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
		CC=x86_64-w64-mingw32-gcc \
		go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe $(MAIN_PKG)
	@echo "✅ $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe"

.PHONY: macos
macos:
	@echo "▶ Building $(APP_NAME) for darwin/arm64 (Apple Silicon)..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
		go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-darwin-arm64 $(MAIN_PKG)
	@echo "✅ $(DIST_DIR)/$(APP_NAME)-darwin-arm64"

.PHONY: macos-intel
macos-intel:
	@echo "▶ Building $(APP_NAME) for darwin/amd64 (Intel Mac)..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
		go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_PKG)
	@echo "✅ $(DIST_DIR)/$(APP_NAME)-darwin-amd64"

# ---- Build all platforms via Docker (no local cross-compiler needed) ----

.PHONY: docker-all
docker-all: docker-linux docker-linux-arm docker-windows
	@echo ""
	@echo "════════════════════════════════════════"
	@echo "  ✅ All builds complete in $(DIST_DIR)/"
	@echo "════════════════════════════════════════"
	@ls -lh $(DIST_DIR)/$(APP_NAME)-*

.PHONY: docker-linux
docker-linux:
	@echo "▶ [Docker] Building linux/amd64..."
	@mkdir -p $(DIST_DIR)
	docker run --rm \
		-v "$(CURDIR)":/src -w /src \
		-e CGO_ENABLED=1 \
		-e GOOS=linux -e GOARCH=amd64 \
		golang:1.27 \
		go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_PKG)
	@echo "✅ $(DIST_DIR)/$(APP_NAME)-linux-amd64"

.PHONY: docker-linux-arm
docker-linux-arm:
	@echo "▶ [Docker] Building linux/arm64..."
	@mkdir -p $(DIST_DIR)
	docker run --rm \
		-v "$(CURDIR)":/src -w /src \
		-e CGO_ENABLED=1 \
		-e GOOS=linux -e GOARCH=arm64 \
		-e CC=aarch64-linux-gnu-gcc \
		goreleaser/goreleaser-cross:v1.27 \
		go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-linux-arm64 $(MAIN_PKG)
	@echo "✅ $(DIST_DIR)/$(APP_NAME)-linux-arm64"

.PHONY: docker-windows
docker-windows:
	@echo "▶ [Docker] Building windows/amd64..."
	@mkdir -p $(DIST_DIR)
	docker run --rm \
		-v "$(CURDIR)":/src -w /src \
		-e CGO_ENABLED=1 \
		-e GOOS=windows -e GOARCH=amd64 \
		-e CC=x86_64-w64-mingw32-gcc \
		goreleaser/goreleaser-cross:v1.27 \
		go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe $(MAIN_PKG)
	@echo "✅ $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe"

# ---- Release: ready-to-deploy installer packages ----

# Helper: create a bundle dir with common files
define bundle
	@mkdir -p $(DIST_DIR)/$(1)/data
	@cp scripts/install.sh scripts/uninstall.sh $(DIST_DIR)/$(1)/ 2>/dev/null || true
	@cp scripts/install.bat scripts/uninstall.bat $(DIST_DIR)/$(1)/ 2>/dev/null || true
	@chmod +x $(DIST_DIR)/$(1)/install.sh $(DIST_DIR)/$(1)/uninstall.sh 2>/dev/null || true
endef

.PHONY: release
release: release-smart-home release-factory
	@echo ""
	@echo "════════════════════════════════════════════════"
	@echo "  ✅ All release packages ready in $(DIST_DIR)/"
	@echo "════════════════════════════════════════════════"
	@ls -lh $(DIST_DIR)/*.zip

.PHONY: release-smart-home
release-smart-home: build
	@echo "▶ Packaging Smart Home installer..."
	@rm -rf $(DIST_DIR)/ems-smart-home
	@mkdir -p $(DIST_DIR)/ems-smart-home
	@cp $(DIST_DIR)/$(APP_NAME) $(DIST_DIR)/ems-smart-home/
	@cp .env.smart-home $(DIST_DIR)/ems-smart-home/.env
	@cp config/dev.yaml $(DIST_DIR)/ems-smart-home/config.yaml
	$(call bundle,ems-smart-home)
	@cd $(DIST_DIR) && zip -rq ems-smart-home-$$(go env GOOS)-$$(go env GOARCH).zip ems-smart-home/
	@rm -rf $(DIST_DIR)/ems-smart-home
	@echo "✅ $(DIST_DIR)/ems-smart-home-$$(go env GOOS)-$$(go env GOARCH).zip"

.PHONY: release-factory
release-factory: build
	@echo "▶ Packaging Factory installer..."
	@rm -rf $(DIST_DIR)/ems-factory
	@mkdir -p $(DIST_DIR)/ems-factory
	@cp $(DIST_DIR)/$(APP_NAME) $(DIST_DIR)/ems-factory/
	@cp .env.factory $(DIST_DIR)/ems-factory/.env
	@cp config/production.yaml $(DIST_DIR)/ems-factory/config.yaml
	$(call bundle,ems-factory)
	@cd $(DIST_DIR) && zip -rq ems-factory-$$(go env GOOS)-$$(go env GOARCH).zip ems-factory/
	@rm -rf $(DIST_DIR)/ems-factory
	@echo "✅ $(DIST_DIR)/ems-factory-$$(go env GOOS)-$$(go env GOARCH).zip"

# Release for a specific platform (cross-compile + package)
.PHONY: release-linux
release-linux: linux
	@echo "▶ Packaging Linux x86_64 installers..."
	@for profile in smart-home factory; do \
		rm -rf $(DIST_DIR)/ems-$$profile; \
		mkdir -p $(DIST_DIR)/ems-$$profile/data; \
		cp $(DIST_DIR)/$(APP_NAME)-linux-amd64 $(DIST_DIR)/ems-$$profile/$(APP_NAME); \
		cp scripts/install.sh scripts/uninstall.sh $(DIST_DIR)/ems-$$profile/; \
		chmod +x $(DIST_DIR)/ems-$$profile/install.sh $(DIST_DIR)/ems-$$profile/uninstall.sh $(DIST_DIR)/ems-$$profile/$(APP_NAME); \
		if [ "$$profile" = "smart-home" ]; then \
			cp .env.smart-home $(DIST_DIR)/ems-$$profile/.env; \
			cp config/dev.yaml $(DIST_DIR)/ems-$$profile/config.yaml; \
		else \
			cp .env.factory $(DIST_DIR)/ems-$$profile/.env; \
			cp config/production.yaml $(DIST_DIR)/ems-$$profile/config.yaml; \
		fi; \
		cd $(DIST_DIR) && zip -rq ems-$$profile-linux-amd64.zip ems-$$profile/ && cd ..; \
		rm -rf $(DIST_DIR)/ems-$$profile; \
		echo "✅ $(DIST_DIR)/ems-$$profile-linux-amd64.zip"; \
	done

.PHONY: release-linux-arm
release-linux-arm: linux-arm
	@echo "▶ Packaging Linux ARM64 installers..."
	@for profile in smart-home factory; do \
		rm -rf $(DIST_DIR)/ems-$$profile; \
		mkdir -p $(DIST_DIR)/ems-$$profile/data; \
		cp $(DIST_DIR)/$(APP_NAME)-linux-arm64 $(DIST_DIR)/ems-$$profile/$(APP_NAME); \
		cp scripts/install.sh scripts/uninstall.sh $(DIST_DIR)/ems-$$profile/; \
		chmod +x $(DIST_DIR)/ems-$$profile/install.sh $(DIST_DIR)/ems-$$profile/uninstall.sh $(DIST_DIR)/ems-$$profile/$(APP_NAME); \
		if [ "$$profile" = "smart-home" ]; then \
			cp .env.smart-home $(DIST_DIR)/ems-$$profile/.env; \
			cp config/dev.yaml $(DIST_DIR)/ems-$$profile/config.yaml; \
		else \
			cp .env.factory $(DIST_DIR)/ems-$$profile/.env; \
			cp config/production.yaml $(DIST_DIR)/ems-$$profile/config.yaml; \
		fi; \
		cd $(DIST_DIR) && zip -rq ems-$$profile-linux-arm64.zip ems-$$profile/ && cd ..; \
		rm -rf $(DIST_DIR)/ems-$$profile; \
		echo "✅ $(DIST_DIR)/ems-$$profile-linux-arm64.zip"; \
	done

.PHONY: release-windows
release-windows: windows
	@echo "▶ Packaging Windows installers..."
	@for profile in smart-home factory; do \
		rm -rf $(DIST_DIR)/ems-$$profile; \
		mkdir -p $(DIST_DIR)/ems-$$profile/data; \
		cp $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe $(DIST_DIR)/ems-$$profile/$(APP_NAME).exe; \
		cp scripts/install.bat scripts/uninstall.bat $(DIST_DIR)/ems-$$profile/; \
		if [ "$$profile" = "smart-home" ]; then \
			cp .env.smart-home $(DIST_DIR)/ems-$$profile/.env; \
			cp config/dev.yaml $(DIST_DIR)/ems-$$profile/config.yaml; \
		else \
			cp .env.factory $(DIST_DIR)/ems-$$profile/.env; \
			cp config/production.yaml $(DIST_DIR)/ems-$$profile/config.yaml; \
		fi; \
		cd $(DIST_DIR) && zip -rq ems-$$profile-windows-amd64.zip ems-$$profile/ && cd ..; \
		rm -rf $(DIST_DIR)/ems-$$profile; \
		echo "✅ $(DIST_DIR)/ems-$$profile-windows-amd64.zip"; \
	done

# ---- Run ----

.PHONY: run
run:
	@echo "▶ Running EMS in simulation mode..."
	EMS_SIMULATE=true go run $(MAIN_PKG) --config config/dev.yaml

.PHONY: run-smart-home
run-smart-home:
	@echo "▶ Running EMS (Smart Home profile)..."
	EMS_ENV_FILE=.env.smart-home go run $(MAIN_PKG) --simulate --config config/dev.yaml

.PHONY: run-factory
run-factory:
	@echo "▶ Running EMS (Factory profile)..."
	EMS_ENV_FILE=.env.factory go run $(MAIN_PKG) --simulate --config config/dev.yaml

# ---- Test ----

.PHONY: test
test:
	@echo "▶ Running all tests..."
	go test ./internal/... -count=1

.PHONY: test-verbose
test-verbose:
	go test ./internal/... -count=1 -v

# ---- Clean ----

.PHONY: clean
clean:
	rm -rf $(DIST_DIR)
	@echo "✅ Cleaned $(DIST_DIR)/"

# ---- Help ----

.PHONY: help
help:
	@echo ""
	@echo "Local EMS — Build & Release Commands"
	@echo "══════════════════════════════════════════════════════════"
	@echo "  Build (binary only):"
	@echo "    make              Build for current OS"
	@echo "    make linux        Build Linux x86_64"
	@echo "    make linux-arm    Build Linux ARM64 (Raspberry Pi)"
	@echo "    make windows      Build Windows .exe"
	@echo "    make macos        Build macOS Apple Silicon"
	@echo "    make docker-all   Cross-compile all via Docker"
	@echo "  ──────────────────────────────────────────────────────"
	@echo "  Release (installer .zip for customers):"
	@echo "    make release              Package for current OS"
	@echo "    make release-linux        Package Linux x86_64 .zip"
	@echo "    make release-linux-arm    Package Linux ARM64 .zip"
	@echo "    make release-windows      Package Windows .zip"
	@echo "  ──────────────────────────────────────────────────────"
	@echo "  Run:"
	@echo "    make run              Simulation (dev config)"
	@echo "    make run-smart-home   Simulation (smart home)"
	@echo "    make run-factory      Simulation (factory)"
	@echo "  ──────────────────────────────────────────────────────"
	@echo "    make test         Run tests"
	@echo "    make clean        Remove dist/"
	@echo ""
