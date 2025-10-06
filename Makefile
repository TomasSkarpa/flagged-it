# -------------------------------------------------------------------
# Project Makefile for "flagged-it"
#
# This Makefile standardizes building, testing, packaging, and cleaning
# across multiple platforms.
# Designed with portability and CI/CD pipelines in mind.
# -------------------------------------------------------------------

.PHONY: setup run debug clean build check

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

# Run the app with verbose build/run output for troubleshooting
debug:
	go run -v $(MAIN)

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