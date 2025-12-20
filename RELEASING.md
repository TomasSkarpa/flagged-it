# Release Guide

This project uses manual releases triggered from GitHub Actions. You have full control over when and what version to release.

## How Versioning Works

- **VERSION file**: Tracks the current version (e.g., `v1.2.3`)
- **Manual releases**: Trigger releases from GitHub Actions UI
- **Git tags**: Created automatically when you release
- **GitHub Releases**: Created automatically with all platform binaries

## Version Format

Uses [Semantic Versioning](https://semver.org/): `MAJOR.MINOR.PATCH`

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes

## How to Release

### Step 1: Go to GitHub Actions

1. Navigate to your repository on GitHub
2. Click on the **Actions** tab
3. Find **"Build and Release"** workflow in the left sidebar
4. Click **"Run workflow"** button (top right)

### Step 2: Choose Your Release Method

You have two options when running the workflow:

#### Option A: Auto-Bump Version (Easiest - Recommended)

1. Leave **version** field **empty**
2. Select **bump_type**:
   - **patch**: Bug fixes (1.0.0 → 1.0.1) - *default*
   - **minor**: New features (1.0.0 → 1.1.0)
   - **major**: Breaking changes (1.0.0 → 2.0.0)
3. Click **"Run workflow"**

The workflow will:
- Automatically read current version from `VERSION` file
- Bump it based on your selection (patch/minor/major)
- Build binaries for all platforms
- Create a git tag
- Create a GitHub Release with all binaries

#### Option B: Specify Exact Version

1. Enter the exact version in **version** field (e.g., `v1.2.3` or `1.2.3`)
2. Leave **bump_type** as `none` (or any value, it will be ignored)
3. Click **"Run workflow"**

The workflow will:
- Use your specified version exactly
- Update the `VERSION` file
- Build binaries for all platforms
- Create a git tag
- Create a GitHub Release with all binaries

### Example Workflow

**Scenario**: You want to release version 1.2.0

1. Go to Actions → Build and Release → Run workflow
2. Enter `v1.2.0` in the version field
3. Set bump_type to `none`
4. Click Run workflow
5. Wait for the build to complete (~5-10 minutes)
6. Check the Releases page - your release is ready!

## What Gets Built

Each release includes binaries for:

- **Windows**: `flagged-it-windows-amd64.exe` (64-bit) and `flagged-it-windows-386.exe` (32-bit)
- **macOS**: `flagged-it-darwin-amd64` (Intel) and `flagged-it-darwin-arm64` (Apple Silicon)
- **Linux**: `flagged-it-linux-amd64` (64-bit) and `flagged-it-linux-arm64` (ARM64)

All are packaged as `.zip` (Windows) or `.tar.gz` (macOS/Linux).

## Local Testing

Build locally for testing:

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build with version info
make build-release

# Check current version
make version
```

## Release Checklist

- [ ] All tests pass
- [ ] Code is formatted (`make check`)
- [ ] Update CHANGELOG.md (if you maintain one)
- [ ] Go to Actions → Build and Release → Run workflow
- [ ] Enter version or select bump type
- [ ] Wait for build to complete
- [ ] Verify release on GitHub Releases page
- [ ] Test downloaded binaries on target platforms

## Troubleshooting

**Version not bumping?**
- Check that the VERSION file exists and has the correct format (`v1.2.3`)
- Ensure workflow has write permissions

**Release not created?**
- Check that a git tag was created
- Verify the release workflow ran successfully
- Check Actions tab for errors

**Wrong version in binary?**
- Version is set at build time via `-ldflags`
- Rebuild after version changes

