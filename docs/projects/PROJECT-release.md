# snag - Distribution & Release Project

## Overview

This document outlines the distribution and release plan for the `snag` CLI tool. The release phase covers multi-platform builds, binary distribution, package management integration, and version 1.0.0 release procedures.

**Status**: Phase 9 - Not Started (0%)
**Last Updated**: 2025-10-17

## About snag

`snag` is a CLI tool for intelligently fetching web page content using a browser engine, with smart session management and format conversion capabilities. Built in Go, it provides a single binary solution for retrieving web content as Markdown or HTML, with seamless authentication support through Chrome/Chromium browsers.

## Technology Stack

- **Language**: Go 1.21+
- **Build Tool**: Go toolchain
- **CI/CD**: GitHub Actions
- **Package Manager**: Homebrew (macOS/Linux)
- **Distribution**: GitHub Releases
- **License**: Mozilla Public License 2.0

## Release Goals

### Version 1.0.0 Objectives

1. **Multi-Platform Support**: Binaries for macOS (Intel/ARM) and Linux (amd64/ARM64)
2. **Automated Builds**: GitHub Actions for consistent, reproducible builds
3. **Easy Installation**: Homebrew formula for one-command installation
4. **Verified Binaries**: SHA256 checksums for all releases
5. **Semantic Versioning**: Follow semver for predictable upgrades
6. **Professional Distribution**: GitHub Releases with release notes

## Target Platforms

### Tier 1 (Fully Supported)

- **macOS ARM64** (Apple Silicon: M1/M2/M3)
- **macOS AMD64** (Intel Macs)
- **Linux AMD64** (64-bit Linux)

### Tier 2 (Best Effort)

- **Linux ARM64** (Raspberry Pi, ARM servers)

### Future Platforms (Post-MVP)

- Windows AMD64
- Windows ARM64
- FreeBSD

## Implementation Plan

### Task 1: Create GitHub Actions Workflow

**Description**: Set up automated build and release pipeline

**File**: `.github/workflows/release.yml`

**Workflow Triggers**:

- Push to tags matching `v*.*.*` (e.g., v1.0.0)
- Manual workflow dispatch for testing

**Content**:

```yaml
name: Release

on:
  push:
    tags:
      - "v*.*.*"
  workflow_dispatch:

permissions:
  contents: write

jobs:
  build:
    name: Build and Release
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.21"

      - name: Get version
        id: version
        run: |
          if [[ $GITHUB_REF == refs/tags/* ]]; then
            VERSION=${GITHUB_REF#refs/tags/}
          else
            VERSION=$(git describe --tags --always --dirty)
          fi
          echo "version=$VERSION" >> $GITHUB_OUTPUT
          echo "Building version: $VERSION"

      - name: Build binaries
        run: |
          VERSION="${{ steps.version.outputs.version }}"

          # Build for all platforms
          GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Version=$VERSION" -o dist/snag-darwin-arm64
          GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=$VERSION" -o dist/snag-darwin-amd64
          GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$VERSION" -o dist/snag-linux-amd64
          GOOS=linux GOARCH=arm64 go build -ldflags "-X main.Version=$VERSION" -o dist/snag-linux-arm64

          # Make all binaries executable
          chmod +x dist/*

      - name: Generate checksums
        run: |
          cd dist
          sha256sum * > SHA256SUMS
          cat SHA256SUMS

      - name: Create release
        uses: softprops/action-gh-release@v1
        if: startsWith(github.ref, 'refs/tags/')
        with:
          files: |
            dist/snag-darwin-arm64
            dist/snag-darwin-amd64
            dist/snag-linux-amd64
            dist/snag-linux-arm64
            dist/SHA256SUMS
          draft: false
          prerelease: false
          generate_release_notes: true
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Deliverables**:

- `.github/workflows/release.yml` file
- Automated builds for 4 platforms
- SHA256 checksum generation
- Automatic GitHub release creation

### Task 2: Add Version Information to main.go

**Description**: Embed version string in binary

**Changes to main.go**:

```go
package main

import (
    "fmt"
    "os"
    "github.com/urfave/cli/v2"
)

// Version is set by ldflags during build
var Version = "dev"

func main() {
    app := &cli.App{
        Name:    "snag",
        Version: Version,
        Usage:   "fetch web page content using a browser engine",
        // ... rest of app configuration
    }

    if err := app.Run(os.Args); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

**Build Command**:

```bash
go build -ldflags "-X main.Version=v1.0.0" -o snag
```

**Verification**:

```bash
./snag --version
# Output: snag version v1.0.0
```

**Deliverables**:

- Version variable in main.go
- Build process sets version via ldflags
- `--version` flag shows correct version

### Task 3: Test Local Multi-Platform Builds

**Description**: Verify builds work for all platforms before automating

**Build Script**: `scripts/build.sh`

```bash
#!/bin/bash
set -e

VERSION=${1:-"dev"}
DIST_DIR="dist"

echo "Building snag version $VERSION"

# Clean previous builds
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# Build for all platforms
echo "Building for macOS ARM64..."
GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Version=$VERSION" -o "$DIST_DIR/snag-darwin-arm64"

echo "Building for macOS AMD64..."
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=$VERSION" -o "$DIST_DIR/snag-darwin-amd64"

echo "Building for Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$VERSION" -o "$DIST_DIR/snag-linux-amd64"

echo "Building for Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -ldflags "-X main.Version=$VERSION" -o "$DIST_DIR/snag-linux-arm64"

# Make all binaries executable
chmod +x "$DIST_DIR"/*

# Generate checksums
echo "Generating checksums..."
cd "$DIST_DIR"
sha256sum * > SHA256SUMS
cd ..

echo "Build complete! Binaries in $DIST_DIR/"
ls -lh "$DIST_DIR/"
```

**Testing**:

```bash
# Test build script
chmod +x scripts/build.sh
./scripts/build.sh v1.0.0-test

# Verify binaries
file dist/*
dist/snag-darwin-arm64 --version
dist/snag-linux-amd64 --version

# Check checksums
cd dist && sha256sum -c SHA256SUMS
```

**Deliverables**:

- `scripts/build.sh` script
- Local builds working for all platforms
- Checksums verified

### Task 4: Create Homebrew Tap Repository

**Description**: Set up homebrew-tap repository for package distribution

**Repository**: `grantcarthew/homebrew-tap`

**Steps**:

1. **Create Repository on GitHub**:

   - Repository name: `homebrew-tap`
   - Description: "Homebrew formulae for Grant Carthew's projects"
   - Public repository
   - Initialize with README

2. **Create Initial README**:

````markdown
# Homebrew Tap

Official Homebrew tap for Grant Carthew's projects.

## Installation

```bash
brew tap grantcarthew/tap
```
````

## Available Formulae

### snag

Fetch web page content using a browser engine.

```bash
brew install grantcarthew/tap/snag
```

See [snag repository](https://github.com/grantcarthew/snag) for more information.

````

**Deliverables**:
- GitHub repository: `grantcarthew/homebrew-tap`
- Initial README.md
- Repository ready for formula

### Task 5: Create Homebrew Formula

**Description**: Write Homebrew formula for snag

**File**: `homebrew-tap/Formula/snag.rb`

**Content**:

```ruby
class Snag < Formula
  desc "Fetch web page content using a browser engine"
  homepage "https://github.com/grantcarthew/snag"
  version "1.0.0"
  license "MPL-2.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/grantcarthew/snag/releases/download/v1.0.0/snag-darwin-arm64"
      sha256 "CHECKSUM_ARM64"  # Replace with actual checksum
    else
      url "https://github.com/grantcarthew/snag/releases/download/v1.0.0/snag-darwin-amd64"
      sha256 "CHECKSUM_AMD64"  # Replace with actual checksum
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/grantcarthew/snag/releases/download/v1.0.0/snag-linux-arm64"
      sha256 "CHECKSUM_LINUX_ARM64"  # Replace with actual checksum
    else
      url "https://github.com/grantcarthew/snag/releases/download/v1.0.0/snag-linux-amd64"
      sha256 "CHECKSUM_LINUX_AMD64"  # Replace with actual checksum
    end
  end

  def install
    if OS.mac?
      if Hardware::CPU.arm?
        bin.install "snag-darwin-arm64" => "snag"
      else
        bin.install "snag-darwin-amd64" => "snag"
      end
    else
      if Hardware::CPU.arm?
        bin.install "snag-linux-arm64" => "snag"
      else
        bin.install "snag-linux-amd64" => "snag"
      end
    end
  end

  def caveats
    <<~EOS
      snag requires Chrome or Chromium to be installed.

      Install Chromium with:
        brew install chromium

      Or download Chrome from:
        https://www.google.com/chrome/
    EOS
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/snag --version")
  end
end
````

**Formula Development Notes**:

- Checksums must be replaced with actual SHA256 values after release
- Formula requires one binary per platform (not tarball)
- Caveats remind users about Chrome/Chromium requirement

**Deliverables**:

- `Formula/snag.rb` in homebrew-tap repository
- Formula tested locally
- Ready for v1.0.0 release

### Task 6: Test Homebrew Installation Locally

**Description**: Verify formula works before public release

**Local Testing Steps**:

```bash
# Install formula from local tap
brew tap grantcarthew/tap
brew install --verbose --debug grantcarthew/tap/snag

# Verify installation
which snag
snag --version

# Test basic functionality
snag https://example.com

# Uninstall for testing
brew uninstall snag

# Test upgrade (after making changes)
brew upgrade snag

# Audit formula for issues
brew audit --strict snag
brew style snag
```

**Testing Checklist**:

- [ ] Formula syntax is correct
- [ ] Installation completes without errors
- [ ] Binary is executable and in PATH
- [ ] `snag --version` shows correct version
- [ ] Basic functionality works (fetch a page)
- [ ] Uninstall works cleanly
- [ ] Formula passes `brew audit`
- [ ] Formula passes `brew style`

**Common Issues and Fixes**:

- **Wrong checksum**: Update SHA256 in formula
- **Binary not executable**: Check `chmod +x` in build
- **Path issues**: Verify `bin.install` command
- **Missing dependencies**: Add to formula if needed

**Deliverables**:

- Verified working Homebrew formula
- Installation tested on macOS
- All audit checks passing

### Task 7: Create Release Checklist

**Description**: Document release process for consistency

**File**: `docs/RELEASE_PROCESS.md`

**Content**:

````markdown
# Release Process

## Pre-Release Checklist

- [ ] All MVP features complete and tested
- [ ] Documentation complete (README, LICENSES, etc.)
- [ ] Version number decided (following semver)
- [ ] CHANGELOG.md updated with release notes
- [ ] All tests passing (once implemented)
- [ ] Build script tested locally
- [ ] Homebrew formula tested locally

## Release Steps

### 1. Prepare Release

```bash
# Update version in files if needed
VERSION="v1.0.0"

# Update CHANGELOG.md
# Add release notes under ## [1.0.0] - 2025-MM-DD

# Commit changes
git add .
git commit -m "chore: prepare for $VERSION release"
git push
```
````

### 2. Create Git Tag

```bash
VERSION="v1.0.0"

# Create annotated tag
git tag -a $VERSION -m "Release $VERSION"

# Push tag (triggers GitHub Actions)
git push origin $VERSION
```

### 3. Monitor GitHub Actions

- Go to Actions tab in GitHub repository
- Watch "Release" workflow
- Verify all builds succeed
- Check that release is created

### 4. Verify Release

- Go to Releases page
- Verify all 4 binaries are attached
- Verify SHA256SUMS file is attached
- Download and test a binary:
  ```bash
  curl -L https://github.com/grantcarthew/snag/releases/download/$VERSION/snag-darwin-arm64 -o snag
  chmod +x snag
  ./snag --version
  ./snag https://example.com
  ```

### 5. Update Homebrew Formula

```bash
# Get checksums from release
curl -L https://github.com/grantcarthew/snag/releases/download/$VERSION/SHA256SUMS

# Update Formula/snag.rb with:
# - New version number
# - New URL for each platform
# - New SHA256 checksums

# Commit and push
cd homebrew-tap
git add Formula/snag.rb
git commit -m "snag: update to $VERSION"
git push
```

### 6. Test Homebrew Installation

```bash
# Update tap
brew update

# Install or upgrade
brew upgrade snag
# or
brew install grantcarthew/tap/snag

# Verify
snag --version
```

### 7. Update README Installation Instructions

Update README.md with correct version numbers in download URLs.

### 8. Announce Release

- [ ] Post to GitHub Discussions (if enabled)
- [ ] Share on relevant platforms (Twitter, etc.)
- [ ] Update project website (if applicable)

## Post-Release

- [ ] Monitor GitHub issues for bug reports
- [ ] Respond to user feedback
- [ ] Plan next release features

## Versioning Guide

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (1.0.0 → 2.0.0): Breaking changes
- **MINOR** (1.0.0 → 1.1.0): New features (backward compatible)
- **PATCH** (1.0.0 → 1.0.1): Bug fixes

## Rollback Procedure

If critical issues found after release:

1. **Delete broken release**:

   ```bash
   gh release delete $VERSION
   git tag -d $VERSION
   git push origin :refs/tags/$VERSION
   ```

2. **Fix issues in code**

3. **Release patch version**:
   ```bash
   git tag -a v1.0.1 -m "Release v1.0.1 (fixes critical bug)"
   git push origin v1.0.1
   ```

````

**Deliverables**:
- Complete release process documentation
- Checklist for consistent releases
- Rollback procedure for emergencies

### Task 8: Create CHANGELOG.md

**Description**: Document release history and changes

**File**: `CHANGELOG.md`

**Content**:

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2025-MM-DD

### Added
- Initial release of snag CLI tool
- Intelligent web page content fetching using Chrome/Chromium
- Browser session management (auto-launch, existing instance detection)
- HTML to Markdown conversion with html-to-markdown/v2
- Multiple output formats: Markdown (default), HTML
- Authentication detection (HTTP 401/403, login form patterns)
- Automatic browser mode selection (headless/visible for auth)
- File output support with `-o/--output` flag
- Custom user agent with `--user-agent` flag
- Page load timeout control with `--timeout` flag
- Wait for selector with `--wait-for` flag
- Browser visibility control (`--force-headless`, `--force-visible`)
- Open browser without fetching with `--open-browser` flag
- Tab management with `--close-tab` flag
- Custom port with `--port` flag
- Four logging levels: quiet, normal, verbose, debug
- Color output with emoji indicators (auto-detection)
- Comprehensive CLI help with examples
- Mozilla Public License 2.0
- Multi-platform support: macOS (Intel/ARM), Linux (amd64/ARM64)
- Homebrew formula for easy installation

### Technical
- Built with Go 1.21+
- CLI framework: github.com/urfave/cli/v2
- Browser control: github.com/go-rod/rod
- HTML conversion: github.com/JohannesKaufmann/html-to-markdown/v2
- Clean separation: stdout (content) / stderr (logs)
- Proper exit codes: 0 (success), 1 (error)
- Helpful error messages with suggestions

### Documentation
- Comprehensive README with usage examples
- Troubleshooting guide for common issues
- Advanced usage examples
- Third-party license attribution
- Design decisions documented in docs/design.md

[Unreleased]: https://github.com/grantcarthew/snag/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/grantcarthew/snag/releases/tag/v1.0.0
````

**Deliverables**:

- CHANGELOG.md file
- Release notes for v1.0.0
- Template for future releases

### Task 9: Test Release on Linux

**Description**: Verify Linux builds work correctly

**Testing Environments**:

- Ubuntu 22.04 LTS (recommended)
- Arch Linux (optional)
- Debian 12 (optional)

**Test Plan**:

```bash
# Download Linux binary
VERSION="v1.0.0"
curl -L https://github.com/grantcarthew/snag/releases/download/$VERSION/snag-linux-amd64 -o snag
chmod +x snag

# Verify checksum
curl -L https://github.com/grantcarthew/snag/releases/download/$VERSION/SHA256SUMS | grep linux-amd64
sha256sum snag

# Test version
./snag --version

# Install Chromium
sudo apt install chromium-browser  # Ubuntu/Debian
# or
sudo pacman -S chromium  # Arch

# Test basic functionality
./snag https://example.com
./snag --format html https://example.com
./snag -o output.md https://example.com

# Test verbose mode
./snag --verbose https://example.com

# Test headless mode
./snag --force-headless https://example.com
```

**Deliverables**:

- Linux binary tested and working
- Installation process verified on Linux
- Any platform-specific issues documented

### Task 10: Execute v1.0.0 Release

**Description**: Perform the actual v1.0.0 release

**Pre-Flight Checklist**:

- [ ] All code complete and committed
- [ ] Documentation complete
- [ ] CHANGELOG.md updated
- [ ] Version number finalized
- [ ] Local builds tested
- [ ] GitHub Actions workflow tested
- [ ] Homebrew formula ready

**Release Steps**:

```bash
# 1. Final commit
git add .
git commit -m "chore: prepare for v1.0.0 release"
git push

# 2. Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0 - Initial public release"
git push origin v1.0.0

# 3. Monitor GitHub Actions
# - Go to Actions tab
# - Watch release workflow
# - Verify all builds succeed

# 4. Verify GitHub Release
# - Check Releases page
# - Download and test binaries
# - Verify checksums

# 5. Update Homebrew formula
cd ../homebrew-tap
# Update Formula/snag.rb with v1.0.0 checksums
git add Formula/snag.rb
git commit -m "snag: release v1.0.0"
git push

# 6. Test Homebrew installation
brew update
brew install grantcarthew/tap/snag
snag --version

# 7. Announce release
# - Update project website
# - Post to relevant communities
# - Share on social media
```

**Deliverables**:

- v1.0.0 release published on GitHub
- Homebrew formula updated and working
- Binaries available for all platforms
- Release announced publicly

## Distribution Channels

### Primary Channel: GitHub Releases

- Official source for all releases
- Contains all platform binaries
- Includes SHA256 checksums
- Release notes and changelog

### Secondary Channel: Homebrew

- Easiest installation method for macOS/Linux users
- Automatic dependency management
- Simple upgrade path
- Integrated with system package manager

### Future Channels (Post-v1.0.0)

- **apt repository**: Debian/Ubuntu packages
- **yum/dnf repository**: RedHat/Fedora packages
- **AUR**: Arch User Repository
- **Snapcraft**: Universal Linux packages
- **Docker Hub**: Container images

## Success Criteria

- ✅ GitHub Actions workflow builds for 4 platforms
- ✅ All binaries build successfully
- ✅ SHA256 checksums generated
- ✅ GitHub Release created automatically
- ✅ Homebrew tap repository created
- ✅ Homebrew formula working
- ✅ Installation tested on macOS
- ✅ Installation tested on Linux
- ✅ Version information embedded in binary
- ✅ CHANGELOG.md created
- ✅ Release process documented
- ✅ v1.0.0 released publicly

## Timeline Estimate

- Task 1 (GitHub Actions): 2-3 hours
- Task 2 (Version info): 30 minutes
- Task 3 (Local builds): 1-2 hours
- Task 4 (Homebrew tap): 30 minutes
- Task 5 (Formula): 2-3 hours
- Task 6 (Test Homebrew): 1-2 hours
- Task 7 (Release checklist): 1 hour
- Task 8 (CHANGELOG): 30 minutes
- Task 9 (Linux testing): 1-2 hours
- Task 10 (Release execution): 1-2 hours

**Total**: 11-16 hours of work

## Post-Release Maintenance

### Version Updates

- Patch releases for bug fixes
- Minor releases for new features
- Major releases for breaking changes

### Homebrew Updates

- Update formula for each release
- Maintain checksums
- Test installation process

### Support

- Monitor GitHub issues
- Respond to bug reports
- Help users with installation issues

## Related Documents

- `PROJECT.md`: Main project implementation plan
- `PROJECT-testing.md`: Testing strategy and plan
- `PROJECT-documentation.md`: Documentation completion plan
- `README.md`: User-facing documentation
- `docs/design.md`: Design decisions and rationale
