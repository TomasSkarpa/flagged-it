# flagged-it
Learn about countries playing different game modes – web application built with Go backend and Svelte frontend.

## Setup Instructions

## Quick Start

### Prerequisites

Install the following tools:

**Go 1.21+**: [golang.org/dl](https://golang.org/dl/) or `brew install go` (macOS)

**Node.js 24+**: [nodejs.org](https://nodejs.org/) or `brew install node` (macOS)

**Make**: Usually pre-installed on macOS/Linux, or `winget install ezwinports.make` (Windows)

### Setup & Run

```bash
# 1. Install all dependencies (Go + Node.js)
make setup

# 2. Start both API and web servers
make dev
```

The application will be available at:
- **Frontend**: http://localhost:5173
- **API**: http://localhost:8080

### Available Commands

- `make setup` - Install Go and Node.js dependencies
- `make dev` - Start both API and web development servers
- `make check` - Run code quality checks
- `make clean` - Remove build artifacts
- `make version` - Show current version

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
refactor: refactor dashboard component
enhance: github action workflow new logic
chore: update from node v20 to node v24 on svelte ap
docs: update setup instructions
```

## Attributions

This project uses the following open source libraries and assets:

- **[Twemoji](https://twemoji.twitter.com/)** - Flag graphics by Twitter (MIT License)

See the `licenses/` directory for full license texts.
