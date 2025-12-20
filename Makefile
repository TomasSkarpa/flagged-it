# -------------------------------------------------------------------
# Project Makefile for "flagged-it"
#
# This Makefile standardizes building, testing, packaging, and cleaning
# across multiple platforms.
# Designed with portability and CI/CD pipelines in mind.
# -------------------------------------------------------------------

.PHONY: setup run debug clean build check web build-all build-release version

# -------------------------------------------------------------------
# Configurable variables and cross-platform ready commands
# -------------------------------------------------------------------

BINARY    := flagged-it
MAIN      := cmd/main.go
BUILD_DIR := build

# Default target OS/arch
GOOS   ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Windows binaries need ".exe" extension
ifeq ($(GOOS),windows)
  EXT := .exe
else
  EXT :=
endif

# Final output path for current build
OUT := $(BUILD_DIR)/$(BINARY)-$(GOOS)-$(GOARCH)$(EXT)

PLATFORMS = darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/386

# Detect Windows for cross-platform commands
ifeq ($(OS),Windows_NT)
  MKDIR := if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
  EXE_EXT := .exe
  # Windows-compatible env for Go build
  GO_BUILD = set GOOS=$(GOOS)&& set GOARCH=$(GOARCH)&& go build -o $(OUT) $(MAIN)
  RM_RF := if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR)
else
  MKDIR := mkdir -p $(BUILD_DIR)
  EXE_EXT :=
  GO_BUILD = GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(OUT) $(MAIN)
  RM_RF := rm -rf $(BUILD_DIR)
endif

OUT := $(BUILD_DIR)/$(BINARY)-$(GOOS)-$(GOARCH)$(EXE_EXT)

# -------------------------------------------------------------------
# Setup and development targets
# -------------------------------------------------------------------

setup:
	go mod tidy

run:
	go run $(MAIN)

# Run the app in debug mode
debug:
	go run $(MAIN) -v

# Run the app in web mode
web:
	@$(MKDIR)
	@GOOS=js GOARCH=wasm go build -o $(BUILD_DIR)/flagged-it.wasm $(MAIN)
	@if [ -f "$$(go env GOROOT)/lib/wasm/wasm_exec.js" ]; then \
		cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" $(BUILD_DIR)/; \
	elif [ -f "$$(go env GOROOT)/misc/wasm/wasm_exec.js" ]; then \
		cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" $(BUILD_DIR)/; \
	else \
		curl -s https://raw.githubusercontent.com/golang/go/master/misc/wasm/wasm_exec.js -o $(BUILD_DIR)/wasm_exec.js; \
	fi
	@cp index.html $(BUILD_DIR)/
	@cp -r assets $(BUILD_DIR)/
	@echo "Built WebAssembly to /$(BUILD_DIR)"
ifeq ($(shell uname -s),Darwin)
	@echo "Access from other devices: http://$$(ifconfig | grep "inet " | grep -v 127.0.0.1 | head -1 | awk '{print $$2}'):8080"
else
	@echo "Access from other devices: http://$$(hostname -I | awk '{print $$1}'):8080"
endif
	@cd $(BUILD_DIR) && python3 -m http.server 8080 --bind 0.0.0.0

# Remove build artifacts (safe for both Linux/macOS and Windows)
clean:
	go clean
	@$(RM_RF)
	@echo "Cleaned build artifacts"

build:
	@$(MKDIR)
	@$(GO_BUILD)
	@echo "Built $(OUT)"

# Static analysis and formatting
check:
	go vet ./...
	go fmt ./...

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@$(MKDIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%%/*}; \
		arch=$${platform##*/}; \
		ext=$$(if [ "$$os" = "windows" ]; then echo .exe; fi); \
		GOOS=$$os GOARCH=$$arch go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY)-$$os-$$arch$$ext $(MAIN); \
		echo "Built $(BINARY)-$$os-$$arch$$ext"; \
	done
	@echo "All binaries built in $(BUILD_DIR)/"

# Build with version info
VERSION ?= $(shell if [ -f VERSION ]; then cat VERSION; else git describe --tags --always --dirty 2>/dev/null || echo "dev"; fi)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS = -ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

build-release:
	@$(MKDIR)
	@echo "Building $(OUT) with version $(VERSION)..."
	@$(GO_BUILD) $(LDFLAGS)
	@echo "Built $(OUT)"

# Show current version
version:
	@if [ -f VERSION ]; then \
		echo "Current version: $$(cat VERSION)"; \
	else \
		echo "No VERSION file found. Current: dev"; \
	fi