# flagged-it
Guess countries, letters, or facts – native Windows &amp; macOS app in Go.

## Setup Instructions

### Prerequisites
Install Go 1.21 or later:
- **macOS**: `brew install go`
- **Windows**: Download from https://golang.org/dl/
- **Linux**: `sudo apt install golang-go` (Ubuntu/Debian) or `sudo yum install golang` (CentOS/RHEL)

### Quick Setup
```bash
make setup
```

## Running the App

### Development Mode
```bash
make run
```

### Build Executable
```bash
make build
./flagged-it
```

### Build for Specific Platforms
```bash
make build-macos    # Build for macOS
make build-windows  # Build for Windows (run on Windows)
make build-linux    # Build for Linux (run on Linux)
```

### Build for All Platforms
```bash
make build-all
```
Executables will be in `build/` folder.

## Packaging for Distribution

### Package for Specific Platforms
```bash
make package-macos    # Creates .app bundle + zip
make package-windows  # Creates .exe + zip (run on Windows)
make package-linux    # Creates binary + tar.gz (run on Linux)
```

### Package All
```bash
make package-all
```
Creates ready-to-distribute files in `build/` folder.

## Debugging

### View Logs
Add print statements in your code:
```go
fmt.Println("Debug:", variable)
```

### Run with Verbose Output
```bash
make debug
```

### Check Code Quality
```bash
make check
```

### Run Tests
```bash
make test
```

## Project Structure
- `cmd/` - Application entry point
- `internal/app/` - Main app logic
- `internal/games/` - Game modules (each developer works here)
- `internal/ui/` - User interface components

## Code Conventions

### Commit Messages
Use lowercase prefixes:
```
feat: add country guessing game
fix: resolve navigation bug
ref: refactor dashboard component
docs: update setup instructions
```
