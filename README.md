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
- `make check` - Format and analyze code
- `make clean` - Remove build artifacts

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
