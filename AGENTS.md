# AGENTS.md

## Project Overview

`snag` is a CLI tool that intelligently fetches web page content using Chrome/Chromium via the Chrome DevTools Protocol (CDP). Built for AI agents to consume web content efficiently.

**Key Features:**

- Auto-detect and connect to existing Chrome instances
- Launch headless or visible browser modes
- Handle authenticated sessions gracefully
- Multiple output formats: Markdown (default), HTML, Text, PDF, PNG
- Multiple URL support: Process multiple URLs in a single command
- Tab management: List, select, and fetch from existing browser tabs
- Pattern matching: Select tabs by index, exact URL, substring, or regex
- Output directory support: Auto-generate filenames with timestamps
- Single binary distribution, no runtime dependencies

**Technology Stack:**

- Language: Go 1.25.3
- CLI Framework: github.com/urfave/cli/v2 v2.27.7
- Browser Control: github.com/go-rod/rod v0.116.2 (Chrome DevTools Protocol)
- HTML to Markdown: github.com/JohannesKaufmann/html-to-markdown/v2 v2.4.0
- HTML to Text: github.com/k3a/html2text v1.2.1

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

# Test basic content fetching
snag https://example.com                # Fetch page as Markdown
snag --format html https://example.com  # Fetch as HTML
snag --format text https://example.com  # Fetch as plain text
snag --format pdf https://example.com   # Save as PDF (auto-generates filename)
snag --format png https://example.com   # Save as PNG screenshot (auto-generates filename)
snag -o output.md https://example.com   # Save to file

# Test multiple URL fetching
snag https://example.com https://google.com               # Fetch multiple URLs
snag -d output/ https://example.com https://google.com    # Save multiple with auto-generated names
snag -o combined.md https://example.com https://google.com # Combine into single file

# Test tab management features (Phase 2)
snag --open-browser                     # Open persistent browser (with DevTools enabled)
snag --list-tabs                        # List all open tabs
snag https://example.com                # Fetch URL (creates new tab)
snag --list-tabs                        # List tabs again (should show example.com)

# Tab selection by index (1-based)
snag --tab 1                            # Fetch from first tab
snag -t 2                               # Fetch from second tab (short form)

# Tab selection by pattern
snag -t "example.com"                   # Exact URL match (case-insensitive)
snag -t "example"                       # Substring/contains match
snag -t "https://.*\.com"               # Regex pattern match

# Tab features with output options
snag -t 1 --format html                 # Fetch tab 1 as HTML
snag -t "github" -o repo.md             # Fetch tab matching "github", save to file
snag -t 1 --wait-for ".content"         # Wait for selector in existing tab

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

**Strategy:**

- Integration tests with real Chrome/Chromium (no mocking)
- Tab tests may show minor isolation issues (not functional bugs)

**Running Tests:**

```bash
go test -v                           # All tests
go test -v -run TestBrowser_ListTabs # Specific test
go test -v -cover                    # With coverage
```

**Requirements:**

- Chrome, Chromium, Edge, or Brave installed
- 13 tab-related tests must pass

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

## Architecture

**File Structure:**

- `main.go` - CLI (urfave/cli), flags, handlers
- `browser.go` - Browser/tab management (rod)
- `fetch.go` - Page fetching, CDP operations
- `convert.go` - HTML to Markdown
- `logger.go` - Custom logger (4 levels, stderr only)
- `errors.go` - Sentinel errors
- `validate.go` - Input validation

**Key Tab Code Locations:**

```
browser.go:404-434    # ListTabs()
browser.go:434-463    # GetTabByIndex()
browser.go:473-544    # GetTabByPattern() with page.Info() caching
main.go:345-383       # handleListTabs()
main.go:412-534       # handleTabFetch()
```

**Browser Modes:**

1. Connect to existing Chrome (auto-detect)
2. Launch headless (if none found)
3. Launch visible (`--force-visible` for auth)
4. Open only (`--open-browser`)

**Tab Management (Phase 2):**

Phase 2 adds efficient tab management capabilities to work with existing browser tabs:

- **List Tabs** (`--list-tabs`, `-l`): Display all open tabs with index, URL, and title
- **Select by Index** (`--tab <index>`, `-t <index>`): Fetch content from specific tab (1-based indexing)
- **Select by Pattern** (`--tab <pattern>`, `-t <pattern>`): Match tabs by URL (exact/substring/regex)

**Tab Selection Examples:**

```bash
# List all open tabs
$ snag --list-tabs
Available tabs in Chrome (3 tabs):
  [1] https://github.com/grantcarthew/snag - grantcarthew/snag: Intelligent web content fetcher
  [2] https://example.com - Example Domain
  [3] https://go.dev/doc/ - Documentation - The Go Programming Language

# Fetch from tab by index (1-based)
snag --tab 1        # First tab
snag -t 3           # Third tab

# Fetch by exact URL (case-insensitive)
snag -t "https://github.com/grantcarthew/snag"
snag -t "EXAMPLE.COM"  # Case-insensitive match

# Fetch by substring/contains
snag -t "github"       # Contains "github"
snag -t "dashboard"    # Contains "dashboard"

# Fetch by regex pattern
snag -t "https://.*\.com"              # Regex: https:// + anything + .com
snag -t ".*/dashboard"                 # Regex: any URL ending with /dashboard
snag -t "(github|gitlab)\.com"         # Regex: github.com or gitlab.com

# With output options
snag -t 1 --format html                # HTML format
snag -t "docs" -o reference.md         # Save to file
snag -t 2 --wait-for ".loaded"         # Wait for selector
```

**Pattern Matching Rules (Progressive Fallthrough):**

1. **Integer** → Tab index (1-based, e.g., "1" = first tab)
2. **Exact match** → Case-insensitive URL match (`strings.EqualFold`)
3. **Contains match** → Case-insensitive substring (`strings.Contains`)
4. **Regex match** → Case-insensitive regex (`(?i)` flag)
5. **Error** → No match found (`ErrNoTabMatch`)

First matching tab wins if multiple tabs match.

**Key Implementation Details:**

- Tab features require existing browser connection (won't auto-launch)
- `--tab` and `<url>` argument are mutually exclusive
- Tab indexes are 1-based for user display (converted internally to 0-based)
- All pattern matching is case-insensitive
- Performance optimized: Single-pass page.Info() caching (browser.go:487-507)
- Tab listing output goes to stdout (enables piping)

**Use Cases:**

1. **Authenticated Sessions**: Fetch from authenticated tabs without re-authentication
2. **Reduce Tab Clutter**: Reuse existing tabs instead of creating new ones
3. **Quick Access**: List tabs to find content quickly
4. **Pattern Workflows**: Match tabs by URL patterns for automation

## Dependencies

**Direct Dependencies:**

- `github.com/urfave/cli/v2` v2.27.7 - CLI framework
- `github.com/go-rod/rod` v0.116.2 - Chrome DevTools Protocol
- `github.com/JohannesKaufmann/html-to-markdown/v2` v2.4.0 - HTML to Markdown conversion
- `github.com/k3a/html2text` v1.2.1 - HTML to plain text conversion

**Runtime Requirements:**

- Chromium-based browser (Chrome, Chromium, Edge, Brave)
- macOS (arm64/amd64) or Linux (amd64/arm64)

## Troubleshooting

**Quick fixes:**

- Browser not found: Install Chromium-based browser (Firefox NOT supported)
- Connection refused: Try different port `--port 9223`
- Timeout: Increase with `--timeout 60`
- Auth required: Use `--force-visible`, authenticate manually, reconnect
- Empty output: Try `--format html` or `--wait-for <selector>`
- Tab errors: Run `snag --list-tabs` first to see available tabs

**Debug logging:**

```bash
snag --verbose <url>   # Verbose logs to stderr
snag --debug <url>     # CDP message logs
```

## Design Philosophy

**Core Principles:**

- Single binary, no config files
- Passive observer (fetch content, don't automate)
- Unix philosophy: stdout = content, stderr = logs
- Clear, actionable error messages

**Non-Goals:**

- Browser automation (use Puppeteer/Playwright instead)
- Web scraping framework
- JavaScript execution/testing

## Important Implementation Notes

**Critical Bug Fix - Remote Debugging Port:**

- ALWAYS explicitly set `--remote-debugging-port` flag (browser.go:259-260)
- Rod's launcher won't set it for default port 9222, causing random port selection
- Test with both default and custom ports

**Tab Indexing:**

- User-facing: 1-based (tabs [1], [2], [3]...)
- Internal: 0-based (converted in TabInfo struct and GetTabByIndex)

**Performance - GetTabByPattern():**

- Caches `page.Info()` results in single pass (browser.go:487-507)
- Reduces network calls from 3N to N (3x improvement for 10 tabs)
- Do not modify pattern matching without preserving this optimization

## Release Process

Follow the comprehensive guide in `docs/release-process.md`.

**Quick reference**:

```bash
export VERSION="0.0.4"
go test -v ./...                    # Run tests first
# Then follow docs/release-process.md for full steps
```

## License

Mozilla Public License 2.0

Third-party licenses in `LICENSES/` directory.

## Current Development Status

**Recently Completed:**

- ✅ Multiple URL support (commit: 7dddb31)
- ✅ Text format support (plain text extraction)
- ✅ PDF format support (Chrome PDF rendering)
- ✅ PNG format support (full-page screenshots)
- ✅ Output directory support with auto-generated filenames
- ✅ Format refactoring (md/html/text/pdf/png)

**In Progress:**

- 🚧 Argument handling analysis (see PROJECT.md and docs/argument-handling.md)
- 🚧 URL file input (`--url-file FILE`)
- 🚧 All-tabs fetching (`--all-tabs`)

**Planned:**

- Future enhancements tracked in GitHub issues

## Additional Resources

- **README.md**: User-facing documentation and usage examples
- **PROJECT.md**: Current project status and work tracking
- **docs/design-record.md**: Comprehensive design decisions and rationale (26 design decisions documented)
- **docs/argument-handling.md**: Complete argument compatibility matrix
- **docs/release-process.md**: Step-by-step release guide for AI agents
- **docs/testing.md**: Testing documentation and strategies
- **Repository**: https://github.com/grantcarthew/snag
- **Issues**: https://github.com/grantcarthew/snag/issues
