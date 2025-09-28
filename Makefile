.PHONY: setup run build clean test

# Setup project dependencies
setup:
	go mod tidy

# Run in development mode
run:
	go run cmd/main.go

# Build executable for current platform
build:
	go build -o flagged-it cmd/main.go

# Build for macOS
build-macos:
	go build -o build/flagged-it-macos cmd/main.go

# Build for Windows (run on Windows)
build-windows:
	go build -o build/flagged-it.exe cmd/main.go

# Build for Linux (run on Linux)
build-linux:
	go build -o build/flagged-it-linux cmd/main.go

# Build for all platforms
build-all: build-macos
	@echo "Note: Windows and Linux builds must be run on their respective platforms"

# Package as macOS app bundle
package-macos: build-macos
	mkdir -p build/Flagged-It.app/Contents/MacOS
	mkdir -p build/Flagged-It.app/Contents/Resources
	cp build/flagged-it-macos build/Flagged-It.app/Contents/MacOS/flagged-it
	echo '<?xml version="1.0" encoding="UTF-8"?>\n<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">\n<plist version="1.0">\n<dict>\n\t<key>CFBundleExecutable</key>\n\t<string>flagged-it</string>\n\t<key>CFBundleIdentifier</key>\n\t<string>com.flagged-it.app</string>\n\t<key>CFBundleName</key>\n\t<string>Flagged It</string>\n\t<key>CFBundleVersion</key>\n\t<string>1.0</string>\n</dict>\n</plist>' > build/Flagged-It.app/Contents/Info.plist
	cd build && zip -r Flagged-It-macOS.zip Flagged-It.app

# Package for Windows (run on Windows)
package-windows: build-windows
	cd build && zip flagged-it-windows.zip flagged-it.exe

# Package for Linux (run on Linux)
package-linux: build-linux
	cd build && tar -czf flagged-it-linux.tar.gz flagged-it-linux

# Package for current platform only
package-all: package-macos
	@echo "All packages created in build/ folder:"
	@ls -la build/*.zip build/*.tar.gz build/*.app 2>/dev/null || true

# Clean build artifacts
clean:
	rm -f flagged-it build/*

# Run tests
test:
	go test ./...

# Check code quality
check:
	go vet ./...
	go fmt ./...

# Debug with verbose output
debug:
	go run -v cmd/main.go