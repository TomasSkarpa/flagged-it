# flagged-it
Learn about countries playing different game modes – native Windows, macOS, Linux and web application in Go.

## Setup Instructions

Install Go 1.21 or later:
- **macOS**: `brew install go`
- **Windows**: Download from https://golang.org/dl/
- **Linux**: `sudo apt install golang-go` (Ubuntu/Debian) or `sudo yum install golang` (CentOS/RHEL)

Install Make:
- **macOS**: `brew install make`
- **Windows**: Download from https://gnuwin32.sourceforge.net/packages/make.htm or `winget install ezwinports.make`
- **Linux**: `sudo apt install make`

Make Targets
- `make setup` - Install dependencies
- `make run` - Run application
- `make web` - Run local web application
- `make debug` - Run with verbose output (and possibility to edit the json data)
- `make build` - Builds binary based on your system
- `make build-all` - Build binaries for all platforms (Windows, macOS, Linux)
- `make build-release` - Build with version information
- `make version` - Show current version
- `make check` - Format and analyze code
- `make clean` - Remove build artifacts

## Releases

This project uses automatic semantic versioning. See [RELEASING.md](RELEASING.md) for detailed release instructions.

**Quick release**: Go to Actions → "Auto Version Bump" → Run workflow, select bump type (patch/minor/major).

## Versioning Conventions & Contribution

### How to Contribute
Create a pull request, while following the conventions below.

### Branch Naming Convention
Use lowercase and kebab-case for branch names, such as:
- /country-guessing-game
- /dashboard-refactor

### Commit Messages
Use lowercase prefixes:
```
feat: add country guessing game
fix: resolve navigation bug
ref: refactor dashboard component
docs: update setup instructions
```

## Attributions

This project uses the following open source libraries and assets:

- **[Fyne](https://fyne.io/)** - Cross-platform GUI toolkit for Go (BSD 3-Clause License)
- **[Twemoji](https://twemoji.twitter.com/)** - Flag graphics by Twitter (MIT License)

See the `licenses/` directory for full license texts.
