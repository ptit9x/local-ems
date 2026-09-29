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

# ---- Package: binary + config + .env template → zip ----

.PHONY: package-smart-home
package-smart-home: build
	@echo "▶ Packaging Smart Home bundle..."
	@mkdir -p $(DIST_DIR)/ems-smart-home
	@cp $(DIST_DIR)/$(APP_NAME) $(DIST_DIR)/ems-smart-home/
	@cp .env.smart-home $(DIST_DIR)/ems-smart-home/.env
	@cp config/dev.yaml $(DIST_DIR)/ems-smart-home/config.yaml
	@mkdir -p $(DIST_DIR)/ems-smart-home/data
	@cd $(DIST_DIR) && zip -r ems-smart-home-$$(go env GOOS)-$$(go env GOARCH).zip ems-smart-home/
	@rm -rf $(DIST_DIR)/ems-smart-home
	@echo "✅ $(DIST_DIR)/ems-smart-home-$$(go env GOOS)-$$(go env GOARCH).zip"

.PHONY: package-factory
package-factory: build
	@echo "▶ Packaging Factory bundle..."
	@mkdir -p $(DIST_DIR)/ems-factory
	@cp $(DIST_DIR)/$(APP_NAME) $(DIST_DIR)/ems-factory/
	@cp .env.factory $(DIST_DIR)/ems-factory/.env
	@cp config/production.yaml $(DIST_DIR)/ems-factory/config.yaml
	@mkdir -p $(DIST_DIR)/ems-factory/data
	@cd $(DIST_DIR) && zip -r ems-factory-$$(go env GOOS)-$$(go env GOARCH).zip ems-factory/
	@rm -rf $(DIST_DIR)/ems-factory
	@echo "✅ $(DIST_DIR)/ems-factory-$$(go env GOOS)-$$(go env GOARCH).zip"

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
	@echo "Local EMS — Build Commands"
	@echo "══════════════════════════════════════════════════════════"
	@echo "  make              Build for current OS"
	@echo "  make linux        Build Linux x86_64"
	@echo "  make linux-arm    Build Linux ARM64 (Raspberry Pi)"
	@echo "  make windows      Build Windows .exe"
	@echo "  make macos        Build macOS Apple Silicon"
	@echo "  make macos-intel  Build macOS Intel"
	@echo "  make docker-all   Cross-compile all via Docker"
	@echo "  ──────────────────────────────────────────────────────"
	@echo "  make package-smart-home   Package Smart Home bundle"
	@echo "  make package-factory      Package Factory bundle"
	@echo "  ──────────────────────────────────────────────────────"
	@echo "  make run              Run simulation (dev config)"
	@echo "  make run-smart-home   Run simulation (smart home)"
	@echo "  make run-factory      Run simulation (factory)"
	@echo "  ──────────────────────────────────────────────────────"
	@echo "  make test         Run tests"
	@echo "  make clean        Remove dist/"
	@echo ""
