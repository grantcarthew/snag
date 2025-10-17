# snag - Project Implementation Plan

## Overview

`snag` is a CLI tool for intelligently fetching web page content using a browser engine, with smart session management and format conversion capabilities. Built in Go, it provides a single binary solution for retrieving web content as Markdown or HTML, with seamless authentication support through Chrome/Chromium browsers.

## Reference Documentation

**IMPORTANT**: Before starting implementation, review:
- ./docs/design.md - Complete design decisions and rationale
- ./docs/notes.md - License headers and development notes
- ./reference/rod/ - Chrome DevTools Protocol library documentation
- ./reference/urfave-cli/ - CLI framework documentation

## Technology Stack

- **Language**: Go 1.21+
- **CLI Framework**: github.com/urfave/cli/v2
- **Browser Control**: github.com/go-rod/rod
- **HTML Conversion**: github.com/JohannesKaufmann/html-to-markdown/v2
- **License**: Mozilla Public License 2.0

## Implementation Phases

### Phase 1: Foundation & Setup

- Initialize Go module (github.com/grantcarthew/snag)
- Create project files (main.go, browser.go, fetch.go, convert.go, logger.go, errors.go)
- Add MPL 2.0 license headers to all source files
- Implement errors.go with sentinel errors (ErrBrowserNotFound, ErrPageLoadTimeout, etc.)
- Implement logger.go (4 levels: quiet, normal, verbose, debug)
- Implement color support and emoji indicators
- Create basic README.md structure

### Phase 2: CLI Framework

- Import and configure urfave/cli/v2
- Define app metadata (name, version, description)
- Implement main.go CLI structure
- Define all 16 MVP flags with urfave/cli syntax
- Implement --version flag
- Implement --help text with examples
- Add flag validation logic
- Configure position-independent argument parsing

### Phase 3: Browser Management (browser.go)

- Import rod and launcher packages
- Implement browser instance detection (check localhost:9222)
- Implement launcher.LookPath() for browser discovery
- Implement connectToExisting() function
- Implement launchHeadless() function
- Implement launchVisible() function
- Implement port configuration (--port flag)
- Handle browser cleanup on exit
- Add error handling for browser not found

### Phase 4: Page Fetching (fetch.go)

- Implement page.Navigate() with URL
- Implement page load waiting with timeout
- Implement --timeout flag handling
- Implement --wait-for selector logic
- Implement auth detection for HTTP status (401, 403)
- Implement auth detection for login page patterns
- Implement HTML content extraction
- Implement --close-tab behavior
- Add error handling for navigation failures

### Phase 5: Content Processing (convert.go)

- Import html-to-markdown/v2 library
- Implement convertToMarkdown() function
- Implement --format html output (pass-through)
- Implement --format markdown output (default)
- Implement writeToStdout() function
- Implement writeToFile() function (--output flag)
- Add content size calculation and reporting
- Handle conversion errors gracefully

### Phase 6: CLI Integration

- Wire browser manager to CLI commands
- Wire page fetcher to CLI commands
- Wire content converter to CLI commands
- Implement --verbose logging throughout
- Implement --quiet mode (stderr suppression)
- Implement --debug mode with CDP tracing
- Implement --user-agent flag
- Implement --force-headless flag
- Implement --force-visible flag
- Implement --open-browser flag
- Connect all error handling to exit codes (0/1)
- Validate end-to-end flow

### Phase 7: Testing & Validation

- Create testdata/ directory
- Create HTML test fixtures (simple.html, auth.html, etc.)
- Implement test HTTP server for controlled tests
- Write integration test: simple page fetch
- Write integration test: HTML to Markdown conversion
- Write integration test: auth detection (401/403)
- Write integration test: timeout handling
- Write integration test: output to file
- Write integration test: output format selection
- Write integration test: browser connection modes
- Write integration test: CLI flag combinations
- Test verbose/quiet/debug logging modes
- Validate error messages are clear and actionable
- Test on real websites (example.com, etc.)

### Phase 8: Documentation

- Complete README.md with full description
- Document all 16 CLI flags with examples
- Add usage examples for common scenarios
- Add authenticated site workflow example
- Add troubleshooting section
- Create LICENSES/ directory
- Add third-party license files (rod, urfave/cli, html-to-markdown)
- Document platform support (macOS, Linux)

### Phase 9: Distribution & Release

- Create .github/workflows/release.yml
- Configure multi-platform builds:
  - darwin/arm64 (macOS Apple Silicon)
  - darwin/amd64 (macOS Intel)
  - linux/amd64 (Linux 64-bit)
  - linux/arm64 (Linux ARM)
- Set up GitHub release automation
- Generate checksums for binaries
- Create Homebrew tap repository (grantcarthew/homebrew-tap)
- Write Homebrew formula (snag.rb)
- Test Homebrew installation locally
- Test binary downloads on each platform
- Tag v1.0.0 release
- Publish to GitHub releases
- Update Homebrew tap
