# AGENTS.md

## Project Overview

`snag` is a CLI tool that intelligently fetches web page content using Chrome/Chromium via the Chrome DevTools Protocol (CDP). Built for AI agents to consume web content efficiently.

**Key Features:**

- Auto-detect and connect to existing Chrome instances
- Launch headless or visible browser modes
- Handle authenticated sessions gracefully
- Convert HTML to Markdown (default) or output raw HTML
- Single binary distribution, no runtime dependencies

**Technology Stack:**

- Language: Go 1.25.3
- CLI Framework: github.com/urfave/cli/v2
- Browser Control: github.com/go-rod/rod (Chrome DevTools Protocol)
- HTML to Markdown: github.com/JohannesKaufmann/html-to-markdown/v2

## Setup Commands

```bash
# Clone repository
git clone https://github.com/grantcarthew/snag.git
cd snag

# Install dependencies
go mod download

# Build binary
go build -o snag

# Run snag
./snag --version
./snag --help
```

## Build and Test Commands

```bash
# Build for current platform
go build -o snag

# Build with version info
go build -ldflags "-X main.version=1.0.0" -o snag

# Run all tests (integration tests with real browser)
go test -v

# Run specific test
go test -v -run TestFetchPage

# Run with coverage
go test -v -cover

# Cross-platform builds
GOOS=darwin GOARCH=arm64 go build -o snag-darwin-arm64
GOOS=darwin GOARCH=amd64 go build -o snag-darwin-amd64
GOOS=linux GOARCH=amd64 go build -o snag-linux-amd64
GOOS=linux GOARCH=arm64 go build -o snag-linux-arm64

# Code quality checks
go vet ./...
gofmt -l .

# Clean build artifacts
rm -f snag snag-*
```

## Code Style Guidelines

**Go Conventions:**

- Follow standard Go formatting: use `gofmt` or `goimports`
- Use Go 1.25.3+ features and idioms
- Keep functions focused and small
- Use descriptive variable names

**Project-Specific Patterns:**

- Flat project structure at root (main.go, browser.go, fetch.go, convert.go, logger.go, errors.go, validate.go)
- Custom logger for CLI output (logger.go)
- Sentinel errors for internal logic (errors.go)
- Exit codes: 0 (success), 1 (any error)
- Output routing (critical for piping):
  - stdout: Content only (HTML/Markdown)
  - stderr: All logs, warnings, errors, progress indicators

**Naming Conventions:**

- Exported constants use PascalCase: `FormatMarkdown`, `FormatHTML`
- Sentinel errors: `ErrBrowserNotFound`, `ErrPageLoadTimeout`, etc.
- Functions: Use descriptive verbs - `validateURL()`, `fetchPage()`, `convertToMarkdown()`

**Error Handling:**

- Use sentinel errors defined in errors.go
- Wrap errors with context: `fmt.Errorf("failed to navigate to %s: %w", url, err)`
- Clear, actionable error messages via logger
- Never panic for expected errors

**Logging:**

- Use custom Logger with 4 levels (quiet, normal, verbose, debug)
- Log to stderr only (stdout reserved for content)
- Success messages: `logger.Success("Connected to existing Chrome instance")`
- Errors: `logger.Error("Failed to connect")`
- Info messages: `logger.Info("Fetching https://example.com...")`
- Verbose: `logger.Verbose("Navigating to URL...")`
- Debug: `logger.Debug("CDP message: ...")`

**License Headers:**

- Add MPL 2.0 header to all Go source files when creating them
- Apply this header to every new `.go` file created
- See `LICENSE` file Exhibit A (lines 355-367) for reference

```go
// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main
```

## Development Workflow

**Branch Management:**

- Main branch: `main`
- Development branch pattern: `feature/description` or `fix/description`
- Always work on feature branches, PR to main

**Commit Message Format:**

Follow conventional commits style:

```
feat(browser): add support for custom user agents
fix(fetch): resolve WebSocket URL from HTTP endpoint
docs: reorganise project documentation
chore(deps): update rod to v0.116.2
```

Types: `feat`, `fix`, `docs`, `chore`, `test`, `refactor`

**Pull Request Guidelines:**

- Title format: `[component] Brief description`
- Run tests and build before committing
- Ensure code is formatted with `gofmt`
- Update documentation if changing CLI interface
- Keep PRs focused on single feature/fix

## Testing Instructions

**Test Strategy:**

- Integration tests with real Chrome/Chromium browser
- Test fixtures in `testdata/` directory (when added)
- No mocking - test against real browser via rod

**Running Tests:**

```bash
# Run all tests
go test -v

# Run specific test
go test -v -run TestFetchPage

# Run with coverage
go test -v -cover

# Verbose test output with debug logging
go test -v -args --debug
```

**CI/CD:**

- Tests run in GitHub Actions with Chrome pre-installed
- Tests execute in headless mode
- Multi-platform builds: darwin/arm64, darwin/amd64, linux/amd64, linux/arm64

**Test Requirements:**

- Chrome, Chromium, Edge, or Brave browser installed
- Browser discoverable via standard installation paths
- Supported platforms: macOS (arm64/amd64), Linux (amd64/arm64)
- Tests use real browser via rod (no mocking)

## Security Considerations

**Browser Security:**

- browser.go contains security comments about browser handling
- Never bypass certificate validation in production
- User agent customization available via `--user-agent` flag to bypass headless detection
- Close tabs in headless mode by default (configurable with `--close-tab`)

**Input Validation:**

- URL validation via validate.go module
- Format validation (markdown/html only)
- Timeout bounds checking (positive integers)
- Port validation (1-65535 range)

**Authentication:**

- Visible browser mode persists user sessions for authenticated sites
- No credential storage - relies on browser's session management
- Users authenticate manually in visible browser mode

**Output Handling:**

- Content written to stdout (can be piped)
- Logs written to stderr (prevents content leakage)
- File output uses standard file permissions

## Architecture and Components

**File Organization:**

```
snag/
├── main.go          # CLI entry point, urfave/cli setup, config struct
├── browser.go       # Browser management, rod integration
├── fetch.go         # Page fetching logic, CDP operations
├── convert.go       # HTML to Markdown conversion
├── logger.go        # Custom logger with color/emoji support
├── errors.go        # Sentinel error definitions
├── validate.go      # Input validation (URL, format)
├── *_test.go        # Test files (integration tests)
├── go.mod           # Go module dependencies
├── go.sum           # Dependency checksums
├── README.md        # User documentation
├── AGENTS.md        # This file - agent instructions
├── LICENSES/        # Third-party license files
├── docs/            # Design documentation
└── testdata/        # Test fixtures
```

**Key Components:**

- **CLI Interface** (main.go): urfave/cli framework with 16 flags
- **Browser Manager** (browser.go): Detect, launch, and connect to Chrome via rod
- **Page Fetcher** (fetch.go): Navigate, wait, extract HTML content
- **Content Converter** (convert.go): HTML to Markdown using html-to-markdown/v2 with table and strikethrough plugins
- **Logger** (logger.go): Custom colored output with 4 levels
- **Validator** (validate.go): Input validation for URLs and configuration

**Browser Operation Modes:**

1. **Connect Mode**: Auto-detect existing Chrome with remote debugging enabled
2. **Headless Mode**: Launch headless Chrome if no instance found
3. **Visible Mode**: Launch visible Chrome for authentication (`--force-visible`)
4. **Open Browser Mode**: Just open browser without fetching (`--open-browser`)

## Dependencies

**Direct Dependencies:**

- `github.com/urfave/cli/v2` - CLI framework
- `github.com/go-rod/rod` - Chrome DevTools Protocol library
- `github.com/JohannesKaufmann/html-to-markdown/v2` - HTML to Markdown conversion
- `golang.org/x/net` - Network utilities

**Runtime Requirements:**

- Chrome, Chromium, Microsoft Edge, or Brave browser
- macOS (arm64/amd64) or Linux (amd64/arm64)
- Remote debugging port available (default: 9222)

**Configuration:**

- All configuration via CLI flags only
- No config files (`.snagrc`, etc.)
- Power users can use shell aliases for defaults

## Troubleshooting

**Common Issues:**

1. **"No Chromium-based browser found"**

   - Install Chrome, Chromium, Edge, or Brave
   - rod's `launcher.LookPath()` checks standard installation paths
   - Supported browsers: Chrome, Chromium, Edge, Brave (Chromium-based only)
   - Firefox is NOT supported (deprecated CDP support)

2. **"Connection refused on port 9222"**

   - Another process may be using the debugging port
   - Try different port: `snag --port 9223 <url>`
   - Check if Chrome is running with `--remote-debugging-port`

3. **"Page load timeout exceeded"**

   - Increase timeout: `snag --timeout 60 <url>`
   - Check network connectivity
   - Try verbose mode: `snag --verbose <url>`

4. **"Authentication required"**

   - Use visible mode: `snag --force-visible <url>`
   - Manually authenticate in the browser window
   - Run command again - snag will connect to existing session

5. **Empty or incorrect output**
   - Try `--format html` to see raw HTML
   - Use `--wait-for <selector>` for dynamic content
   - Enable debug logging: `snag --debug <url>` (output to stderr)

**Debug Mode:**

```bash
# Enable verbose logging to stderr
snag --verbose https://example.com

# Enable debug logging with CDP messages
snag --debug https://example.com

# Quiet mode (only errors and content)
snag --quiet https://example.com
```

## Design Philosophy

**Core Principles:**

- **Simplicity**: Single binary, no config files, sensible defaults
- **Smart defaults**: Auto-detect browser, default to Markdown, 30s timeout
- **Passive observer**: Fetch content, don't automate (not Puppeteer/Playwright)
- **Unix philosophy**: Do one thing well, pipe-friendly (stdout = content, stderr = logs)
- **Clear errors**: Actionable error messages with suggestions

**Non-Goals:**

- Browser automation (clicking, form filling, multi-step workflows)
- Web scraping framework
- JavaScript execution/testing
- Screenshot capture (post-MVP feature)
- Performance profiling

**Future Enhancements:**

See docs/design-record.md for Phase 2+ features:

- Tab management (`--list-tabs`, `--tab <index>`)
- Additional formats (text, PDF)
- Screenshot support
- Cookie management
- Proxy support

## Release Process

**For AI Agents**: When performing a release, follow the comprehensive step-by-step guide in `docs/release-process.md`.

**Release Steps Summary**:

1. Pre-release checks (tests, build verification)
2. Determine version number (semver)
3. Update CHANGELOG.md (create if doesn't exist)
4. Commit and tag release
5. Build multi-platform binaries (darwin/linux, arm64/amd64)
6. Create GitHub release with binaries
7. Update Homebrew tap formula
8. Test installation
9. Post-release tasks

**Quick Release Command Reference**:

```bash
# Set version
export VERSION="0.0.4"

# Tag and build
git tag -a "v${VERSION}" -m "Release v${VERSION}"
git push origin "v${VERSION}"

# Build binaries
mkdir -p dist
GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=${VERSION}" -o "dist/snag-darwin-arm64"
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=${VERSION}" -o "dist/snag-darwin-amd64"
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=${VERSION}" -o "dist/snag-linux-amd64"
GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=${VERSION}" -o "dist/snag-linux-arm64"
cd dist && sha256sum * > SHA256SUMS && cd ..

# Create GitHub release (using gh CLI)
gh release create "v${VERSION}" --title "v${VERSION}" \
  --notes "Release v${VERSION}" \
  dist/snag-darwin-arm64 \
  dist/snag-darwin-amd64 \
  dist/snag-linux-amd64 \
  dist/snag-linux-arm64 \
  dist/SHA256SUMS

# Update Homebrew tap (manual edit required)
# See docs/release-process.md Step 8 for detailed instructions
```

**Important Notes**:

- Always run tests before releasing: `go test -v ./...`
- Homebrew tap is at `./reference/homebrew-tap/`
- Update `Formula/snag.rb` with new version and tarball SHA256
- Test installation: `brew reinstall grantcarthew/tap/snag`

**Full documentation**: `docs/release-process.md`

## License

Mozilla Public License 2.0

Third-party licenses in `LICENSES/` directory.

## Additional Resources

- **README.md**: User-facing documentation and usage examples
- **docs/design-record.md**: Comprehensive design decisions and rationale
- **docs/release-process.md**: Step-by-step release guide for AI agents
- **Repository**: https://github.com/grantcarthew/snag
- **Issues**: https://github.com/grantcarthew/snag/issues
