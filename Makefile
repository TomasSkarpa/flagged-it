# -------------------------------------------------------------------
# Project Makefile for "flagged-it"
#
# Quick start: make setup && make dev
# -------------------------------------------------------------------

.PHONY: setup dev clean check wails wails-setup wails-dev wails-build

# -------------------------------------------------------------------
# Setup - Install all dependencies
# -------------------------------------------------------------------

setup:
	@echo "Installing Go dependencies..."
	@go mod tidy
	@echo ""
	@echo "Installing Node.js dependencies..."
	@cd web && npm install
	@echo ""
	@echo "Setup complete! Run 'make dev' to start."

# -------------------------------------------------------------------
# Development - Start both API and web servers
# -------------------------------------------------------------------

dev:
	@echo "Starting development servers..."
	@echo ""
	@echo "Starting API server in background..."
	@cd cmd/web && go run main.go -dev > /tmp/flagged-it-api.log 2>&1 & echo $$! > /tmp/flagged-it-api.pid
	@sleep 2
	@echo "API server started on http://localhost:8080"
	@echo "(logs: /tmp/flagged-it-api.log)"
	@echo ""
	@echo "Starting Svelte dev server..."
	@cd web && npm run dev -- --host

# -------------------------------------------------------------------
# Clean build artifacts and temp files
# -------------------------------------------------------------------

clean:
	@echo "Cleaning build artifacts..."
	@go clean
	@rm -rf build
	@rm -rf web/.svelte-kit
	@rm -rf web/build
	@rm -rf cmd/desktop/build
	@rm -rf cmd/desktop/frontend
	@rm -f /tmp/flagged-it-api.log
	@rm -f /tmp/flagged-it-api.pid
	@echo "Done"

# -------------------------------------------------------------------
# Code quality checks
# -------------------------------------------------------------------

check:
	@echo "Running Go checks..."
	@mkdir -p cmd/desktop/frontend/build
	@echo '<!DOCTYPE html><html><head><title>Flagged It</title></head><body><h1>Flagged It</h1></body></html>' > cmd/desktop/frontend/build/index.html
	@go vet ./...
	@go fmt ./...
	@echo "Done"

# -------------------------------------------------------------------
# Wails Desktop App
# -------------------------------------------------------------------

wails-setup:
	@echo "Installing Wails CLI..."
	@go install github.com/wailsapp/wails/v2/cmd/wails@latest
	@echo ""
	@echo "Checking Wails installation..."
	@wails doctor
	@echo ""
	@echo "Wails setup complete!"

wails-dev:
	@echo "Starting Wails development mode..."
	@echo "Building frontend for desktop..."
	@cd web && BUILD_TARGET=desktop npm run build
	@echo ""
	@echo "Copying frontend build to desktop..."
	@rm -rf cmd/desktop/frontend
	@mkdir -p cmd/desktop/frontend
	@cp -r web/build cmd/desktop/frontend/
	@echo ""
	@echo "Starting Wails dev..."
	@cd cmd/desktop && wails dev

wails-build:
	@echo "Building Wails desktop app..."
	@echo ""
	@echo "Building frontend for desktop..."
	@cd web && BUILD_TARGET=desktop npm run build
	@echo ""
	@echo "Copying frontend build to desktop..."
	@rm -rf cmd/desktop/frontend
	@mkdir -p cmd/desktop/frontend
	@cp -r web/build cmd/desktop/frontend/
	@echo ""
	@echo "Building desktop app..."
	@cd cmd/desktop && wails build
	@echo ""
	@echo "Build complete! Check cmd/desktop/build/"

wails: wails-build