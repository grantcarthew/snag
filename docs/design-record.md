# snag - Design Record

> **Document Type**: Design Record
> **Purpose**: Documents the design decisions, rationale, and architecture of the snag CLI tool
> **Audience**: Contributors, maintainers, and anyone interested in understanding why snag works the way it does

## Overview

`snag` is a CLI tool for intelligently fetching web page content using a browser engine, with smart session management and format conversion capabilities.

This document captures the design decisions made during snag's development, the alternatives considered, and the rationale behind each choice.

## Name Rationale

**Chosen: `snag`**

- **Short & memorable**: 4 letters, easy to type
- **Action-oriented**: Implies "quickly grab something"
- **Natural language**: "Snag that page for me"
- **Not taken**: Available on GitHub, npm, Homebrew
- **Format agnostic**: Works for both Markdown and HTML output
- **Expandable**: Can snag pages, tabs, sessions, etc.

Example usage:

```bash
snag https://example.com
snag https://app.internal.com -o docs.md
snag --list-tabs
snag --tab 3
```

Rejected alternatives: `web2md` (misleading with --html), `grab` (too generic), `mdl` (cryptic), `wg`/`pg` (too short), `websnap`/`pageget` (too long)

## Design Decisions Summary

**14 major design decisions documented below:**

| #   | Decision            | Choice                                                      |
| --- | ------------------- | ----------------------------------------------------------- |
| 1   | CLI Arguments       | 16 arguments, standard ordering: `snag [options] <url>`     |
| 2   | Output Formats      | Markdown (default), HTML                                    |
| 3   | Argument Parsing    | Options before URL (simplicity over flexibility)            |
| 4   | Platform Support    | macOS (arm64/amd64), Linux (amd64/arm64) - Windows deferred |
| 5   | Config File         | No config file support (permanent choice)                   |
| 6   | HTML→Markdown       | `html-to-markdown/v2` (embedded library)                    |
| 7   | License Attribution | `LICENSES/` directory                                       |
| 8   | CLI Framework       | `urfave/cli/v2`                                             |
| 9   | CDP Library         | `rod`                                                       |
| 10  | Browser Discovery   | rod's `launcher.LookPath()` (Chromium-based only)           |
| 11  | Logging Strategy    | Custom logger (4 levels, colors, emojis)                    |
| 12  | Error Handling      | Exit 0/1, sentinel errors, clear messages                   |
| 13  | Project Structure   | Flat structure at root                                      |
| 14  | Testing Strategy    | Integration tests with real browser                         |

See detailed rationale in [Design Decisions Made](#design-decisions-made) section below.

## Technology Stack

**Language: Go**

Rationale:

- Single binary distribution (no runtime dependencies)
- Excellent cross-platform support (macOS, Linux)
- Native Chrome DevTools Protocol library (`rod`)
- HTML to Markdown conversion library (`html-to-markdown`)
- Simple Homebrew formula (just binary distribution)
- ~5MB binary size
- GitHub Actions for multi-platform builds

Alternative considered:

- Node.js + pkg: Faster to ship (reuse existing code), but larger binaries (~50MB), still has complexity

## What It Does

**Core Functionality:**

- Fetch web page content using Chrome/Chromium via CDP (Chrome DevTools Protocol)
- Convert HTML to Markdown (or keep raw HTML)
- Smart browser session management
- Handle authenticated sessions gracefully
- Support both headless and visible browser modes

**Smart Behaviors:**

1. **Session Detection**: Auto-detect existing Chrome instance with remote debugging enabled
2. **Mode Selection**:
   - If Chrome running → Connect to existing session (preserves auth/cookies)
   - If no Chrome → Launch headless mode
   - If auth required → Launch visible mode for user authentication
3. **Authentication Handling**: Detect auth requirements (401/403, login pages, OAuth redirects)
4. **Tab Management**: Keep tabs open in visible mode, close in headless mode (configurable)

**Output Formats:**

- Markdown (default) - clean, readable
- HTML (raw) - via `--format html` flag

## What It Does NOT Do

**Explicit Non-Goals:**

- ❌ Remote control browser (clicking, form filling, multi-step workflows)
- ❌ Web scraping framework
- ❌ JavaScript execution/testing
- ❌ Screenshot capture
- ❌ Performance profiling

**Philosophy**: `snag` is a **passive observer** for content retrieval, not an automation framework. For browser automation, use Puppeteer, Playwright, or Selenium.

## Feature Set

### Phase 1: Core (MVP)

**Required:**

- Fetch URL and return content
- Smart Chrome session detection
- Three operation modes: connect, headless, visible
- Authentication detection and handling
- HTML to Markdown conversion
- Output to stdout or file

**CLI Arguments (MVP - v1.0.0):**

Standard flag ordering - options before URL:

```bash
snag [options] <url>
```

**Core Arguments:**

```
  <url>                      URL to fetch (required, unless --open-browser)
  --version                  Display version information
  -h, --help                 Show help message and exit
```

**Output Control:**

```
  -o, --output <file>        Save output to file instead of stdout
  --format <type>            Output format: markdown (default) | html
```

**Page Loading:**

```
  -t, --timeout <seconds>    Page load timeout in seconds (default: 30)
  -w, --wait-for <selector>  Wait for CSS selector before extracting content
```

**Browser Control:**

```
  -p, --port <port>          Chrome remote debugging port (default: 9222)
  -c, --close-tab            Close the browser tab after fetching content
  -fh, --force-headless      Force headless mode even if Chrome is running
  -fv, --force-visible       Force visible mode for authentication
  -ob, --open-browser        Open Chrome browser in visible state (no URL required)
```

**Logging/Debugging:**

```
  -v, --verbose              Enable verbose logging output
  -q, --quiet                Suppress all output except errors and content
  --debug                    Enable debug output
```

**Request Control:**

```
  --user-agent <string>      Custom user agent string (bypass headless detection)
```

**Total MVP Arguments:** 16

### Phase 2: Tab Management (Future)

**Proposed:**

```bash
snag --list-tabs                    # List all open tabs
snag --tab <index>                  # Get content from specific tab
snag --tab <url-pattern>            # Match tab by URL pattern
```

**Use Cases:**

- Get content from already-open authenticated tab
- Fetch from multiple tabs in parallel
- Avoid opening new tabs unnecessarily

### Phase 3+: Post-MVP Features

Each feature below will be a separate project/enhancement after MVP release:

**Text Format Support:**

```
  --format text              Extract plain text only (strip all HTML)
```

**PDF Export:**

```
  --format pdf               Export page as PDF using Chrome rendering
```

**Screenshot Capture:**

```
  --screenshot <file>        Save screenshot of the page (PNG/JPG)
```

**JavaScript Control:**

```
  --no-js                    Disable JavaScript execution (faster for static content)
```

**Cookie Management:**

```
  --cookies <file>           Load/save cookies from JSON file
```

**Advanced Headers:**

```
  --header <key:value>       Add custom HTTP headers (repeatable flag)
```

**Redirect Control:**

```
  --max-redirects <n>        Limit number of HTTP redirects
```

**Proxy Support:**

```
  --proxy <url>              Route requests through proxy server
```

**Other Considerations:**

- `--user-data-dir <path>`: Use specific Chrome profile
- Batch processing from file list
- JSON structured output mode

## Architecture

### Component Design

```
┌─────────────────────────────────────────────┐
│                CLI Interface                │
│  - Argument parsing                         │
│  - Help/usage display                       │
└─────────────┬───────────────────────────────┘
              │
┌─────────────▼───────────────────────────────┐
│           Browser Manager                   │
│  - Detect existing Chrome instance          │
│  - Launch Chrome (headless/visible)         │
│  - Connect via CDP                          │
└─────────────┬───────────────────────────────┘
              │
┌─────────────▼───────────────────────────────┐
│           Page Fetcher                      │
│  - Navigate to URL                          │
│  - Wait for page load                       │
│  - Detect authentication requirements       │
│  - Extract HTML content                     │
└─────────────┬───────────────────────────────┘
              │
┌─────────────▼───────────────────────────────┐
│         Content Converter                   │
│  - HTML → Markdown conversion               │
│  - Content formatting                       │
│  - Output writing                           │
└─────────────────────────────────────────────┘
```

### Key Libraries (Go)

**Chrome DevTools Protocol:**

- `github.com/go-rod/rod` - Simpler API, better resource efficiency, stable architecture

**HTML to Markdown:**

- `github.com/JohannesKaufmann/html-to-markdown/v2` - Well-maintained, proven output quality, CommonMark support

**CLI Framework:**

- `github.com/urfave/cli/v2` - Smaller binary size, simpler architecture, dynamic autocompletion

### Chrome Discovery

Auto-detect Chrome/Chromium paths:

```go
var chromePaths = []string{
    "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
    "/Applications/Chromium.app/Contents/MacOS/Chromium",
    "/usr/bin/chromium",
    "/usr/bin/chromium-browser",
    "/usr/bin/google-chrome",
    "/usr/bin/google-chrome-stable",
}
```

Check for running instance:

```bash
curl -s http://localhost:9222/json/version
```

## Distribution Strategy

### Primary: Homebrew

```bash
brew install grantcarthew/tap/snag
```

**Homebrew Formula:**

```ruby
class Snag < Formula
  desc "Intelligently fetch web page content with browser engine"
  homepage "https://github.com/grantcarthew/snag"
  url "https://github.com/grantcarthew/snag/releases/download/v1.0.0/snag-1.0.0.tar.gz"
  sha256 "..."

  depends_on :macos

  def install
    bin.install "snag"
  end

  test do
    system "#{bin}/snag", "--help"
  end
end
```

### Secondary: Direct Download

```bash
# macOS ARM64
curl -L https://github.com/grantcarthew/snag/releases/latest/download/snag-macos-arm64 -o snag
chmod +x snag
mv snag /usr/local/bin/

# macOS AMD64
curl -L https://github.com/grantcarthew/snag/releases/latest/download/snag-macos-amd64 -o snag

# Linux
curl -L https://github.com/grantcarthew/snag/releases/latest/download/snag-linux-amd64 -o snag
```

### GitHub Actions Release Workflow

- Build for multiple platforms: darwin/arm64, darwin/amd64, linux/amd64, linux/arm64
- Create GitHub release with binaries
- Auto-generate checksums
- Update Homebrew tap automatically

## Dependencies

**Runtime:**

- Chrome or Chromium browser (user-installed)

**Build:**

- Go 1.21+
- Standard library

**External:**

- None (single static binary)

## User Experience

### Example: Simple Fetch

```bash
$ snag https://example.com
# Markdown output to stdout
```

### Example: Authenticated Site

```bash
$ snag https://private.example.com
# Auto-detects auth needed
# Launches visible Chrome for login
# User signs in
# User runs command again
$ snag https://private.example.com
# Connects to existing session
# Fetches content successfully
```

### Example: Save to File

```bash
$ snag https://docs.example.com/api -o api-docs.md
✓ Fetched 12,543 characters
✓ Saved to api-docs.md
```

### Example: Use Existing Chrome

```bash
# User has Chrome open with --remote-debugging-port=9222
$ snag https://example.com
# Connects to existing session
# Preserves all cookies/auth
# Leaves tab open
```

## Migration from Current Implementation

**Design Philosophy:**

- Single-purpose tool: fetch web page content via browser engine
- Smart session management: connect to existing Chrome or launch new instance
- Format flexibility: Markdown (default) or raw HTML output
- Unix philosophy: do one thing well, output to stdout for piping

**Implementation Benefits:**

- No runtime dependencies (single static binary)
- Small binary size (~5MB)
- Fast startup (compiled Go vs interpreted JavaScript)
- Cross-platform support (macOS, Linux)
- Type-safe error handling

## Design Decisions

### 1. CLI Arguments

- **Decision**: Use 16 arguments with standard flag ordering: `snag [options] <url>`
- **Rationale**:
  - Essential CLI features: `--version`, `--quiet`, `--user-agent`, `--format`
  - Standard flag-then-argument pattern keeps implementation simple
  - Avoids parsing complexity of position-independent arguments
  - Consistent with most CLI tools (curl, wget, etc.)

### 2. Output Formats

- **Decision**: MVP supports `markdown` (default) and `html` only
- **Future Consideration**: `text`, `pdf` as separate enhancement projects
- **Rationale**:
  - Keeps MVP scope focused
  - `text` and `pdf` add complexity (plain text extraction, Chrome PDF API)
  - Extensible via `--format` flag design

### 3. Argument Parsing

- **Decision**: Standard flag ordering - options before URL: `snag [options] <url>`
- **Rationale**:
  - Simpler implementation and maintenance
  - Consistent with most CLI tools
  - urfave/cli v2 doesn't natively support position-independent args
  - Avoiding additional parsing complexity was prioritized

### 4. Platform Support

- **Decision**: MVP targets macOS and Linux only; Windows is future consideration
- **MVP Platforms**:
  - darwin/arm64 (macOS Apple Silicon)
  - darwin/amd64 (macOS Intel)
  - linux/amd64 (Linux 64-bit)
  - linux/arm64 (Linux ARM - Raspberry Pi, servers)
- **Post-MVP**: Windows support (requires Windows-specific path handling)
- **Rationale**:
  - Primary development/use on macOS and Linux
  - Windows adds complexity (path conventions, file handling)
  - Can add later without breaking existing users

### 5. Config File Support

- **Decision**: No config file support - permanent design choice
- **Rationale**:
  - CLI flags are sufficient for all use cases
  - Most users will use defaults (30s timeout, markdown format, auto-detect Chromium)
  - Power users can use shell aliases: `alias snag='snag --verbose --timeout 60'`
  - Avoids complexity: file parsing, precedence rules, file location conventions
  - Keeps tool simple and focused

### 6. HTML→Markdown Conversion

- **Decision**: Embed `github.com/JohannesKaufmann/html-to-markdown/v2` Go library
- **Library Choice**: `html-to-markdown` v2 (MIT license)
- **Rationale**:
  - Current `html2markdown` CLI is a wrapper around this exact library
  - Proven output quality (already using it via CLI)
  - No external dependencies (single binary)
  - Simple API (~40 lines for conversion)
  - Well-maintained, modern v2 with plugin support
  - Supports CommonMark, tables, strikethrough
- **Implementation**:
  ```go
  import htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
  markdown, err := htmltomarkdown.ConvertString(htmlContent)
  ```

### 7. License Attribution

- **Decision**: Use `LICENSES/` directory for third-party license attribution
- **Approach**:
  - Create `LICENSES/` directory in repository
  - Include each dependency's license as separate file (e.g., `LICENSES/html-to-markdown.txt`)
  - Visible in GitHub, included in source releases
  - Complies with MIT license requirements (attribution for html-to-markdown)
- **Post-MVP**: Consider `snag --licenses` command to print embedded licenses
- **Automation**: Use `go-licenses` tool during build/release process

### 8. CLI Framework Choice

- **Decision**: Use `github.com/urfave/cli` for CLI framework
- **Library**: urfave/cli v2 (MIT license)
- **Rationale**:
  - Smaller binary size compared to Cobra (important for single-binary tool)
  - Simpler, less boilerplate code
  - Better dynamic bash autocompletion (can autocomplete argument values)
  - Still supports subcommands for Phase 2 (--list-tabs, --tab)
  - Declarative, clean API
  - Widely used (23.6k GitHub stars), well-maintained
  - Less globals-heavy than Cobra's architectural pattern
- **Alternatives Considered**:
  - Cobra: More feature-rich but larger binaries, more dependencies
  - Coral: Cobra fork with fewer dependencies but less mature (431 stars)
  - Standard library flag package: Too basic, no subcommand support
- **Reference**: `reference/urfave-cli/`

### 9. CDP Library Choice

- **Decision**: Use `github.com/go-rod/rod` for Chrome DevTools Protocol
- **Library**: rod (MIT license)
- **Rationale**:
  - Simpler, more intuitive API compared to chromedp
  - Better resource efficiency (uses less CPU/memory)
  - More stable architecture with consistent CDP versioning
  - Auto-wait elements feature reduces error handling complexity
  - Chained context design for intuitive timeout/cancel
  - Debugging-friendly with auto input tracing
  - API more closely resembles Puppeteer (easier porting)
  - Perfect fit for passive content fetching use case
- **Alternatives Considered**:
  - chromedp: Faster raw speed but steeper learning curve, more resource usage
  - Direct CDP: Too low-level, too much work
- **Reference**: `reference/rod/`

### 10. Chrome/Chromium Discovery

- **Decision**: Use rod's built-in `launcher.LookPath()` for browser discovery
- **Approach**:
  - **Three-tier strategy**:
    1. Try connecting to existing browser instance (port 9222)
    2. If not running, use `launcher.LookPath()` to find system browser
    3. Launch browser in appropriate mode (headless/visible)
  - No auto-download (users should have browser installed)
  - No config file needed (rod handles discovery automatically)
  - No environment variable needed (rod's paths are comprehensive)
- **Supported Browsers** (Chromium-based only):
  - Google Chrome
  - Chromium
  - Microsoft Edge
  - Brave Browser
  - Chrome Canary
- **Platform Coverage**:
  - macOS: `/Applications/*.app`, `/usr/bin/*`
  - Linux: `/usr/bin/*`, system PATH
- **Firefox NOT Supported**:
  - Firefox deprecated CDP support (end of 2024)
  - Firefox moved to WebDriver BiDi protocol
  - rod is CDP-only, no Firefox compatibility
- **Rationale**:
  - rod's `LookPath()` is comprehensive and battle-tested
  - Covers all major Chromium browsers and installation paths
  - Cross-platform support built-in
  - Zero maintenance - rod team keeps paths updated
  - Clear error message if no browser found
- **Implementation Pattern**:
  ```go
  // Try connect to existing browser first
  browser := rod.New().ControlURL("ws://localhost:9222")
  if err := browser.Connect(); err == nil {
      // Connected to existing browser
  } else {
      // Launch new browser
      path, exists := launcher.LookPath()
      if !exists {
          return errors.New("no Chromium-based browser found")
      }
      url := launcher.New().Bin(path).Headless(!visible).MustLaunch()
      browser = rod.New().ControlURL(url).MustConnect()
  }
  ```
- **Reference**: `reference/rod/lib/launcher/browser.go:202-251`

### 11. Logging & Output Strategy

- **Decision**: Simple custom logger with colored output, no external dependencies
- **Output Routing**:
  - **stdout**: Content only (HTML/Markdown) - enables piping
  - **stderr**: All logs, warnings, errors, progress indicators
- **Log Levels**:
  - **Quiet** (`--quiet`): Only fatal errors to stderr
  - **Normal** (default): Key operations with emoji indicators
  - **Verbose** (`--verbose`): Detailed operation logs
  - **Debug** (`--debug`): Everything + CDP messages, timing info
- **Color Support**:
  - Auto-detect TTY: `isatty.IsTerminal(os.Stderr.Fd())`
  - Respect `NO_COLOR` environment variable
  - Colors: ✓ Green (success), ⚠ Yellow (warnings), ✗ Red (errors), Cyan (info)
- **Emoji Usage**:
  - Use emojis by default (✓ ⚠ ✗) when colors enabled
  - No detection needed - modern terminals support UTF-8
  - Degrades gracefully on ancient terminals
- **Format Examples**:

  ```
  Normal mode:
  Connecting to Chrome on port 9222...
  ✓ Connected to existing Chrome instance
  Fetching https://example.com...
  ✓ Success (12.5 KB)

  Verbose mode:
  Connecting to Chrome on port 9222...
  ✓ Connected to existing Chrome instance
  Navigating to https://example.com...
  Waiting for page load (timeout: 30s)...
  ✓ Page loaded (2.3s)
  Extracting HTML content...
  ✓ Extracted 45.2 KB HTML
  Converting to Markdown...
  ✓ Converted to 12.5 KB Markdown
  ✓ Success

  Quiet mode:
  (no output unless fatal error)
  ```

- **No Timestamps**: CLI tools are short-lived, timestamps add noise
- **Progress Indicators**: For operations > 1 second, show elapsed time
- **Why Custom Logger**:
  - Standard `log` package too basic (always adds timestamps, no levels)
  - `log/slog` too verbose/structured for CLI (designed for JSON logging)
  - Custom logger: ~100 lines, exactly what we need
  - Zero external dependencies (no Zap/Zerolog needed)
- **Implementation Pattern**:

  ```go
  type Logger struct {
      level  LogLevel
      color  bool
      writer io.Writer  // os.Stderr
  }

  func (l *Logger) Success(msg string) {
      if l.level >= LevelNormal {
          prefix := "✓"
          if l.color {
              prefix = green + "✓" + reset
          }
          fmt.Fprintf(l.writer, "%s %s\n", prefix, msg)
      }
  }
  ```

### 12. Error Handling & Exit Codes

- **Decision**: Simple exit codes (0/1) with clear error messages, sentinel errors for internal logic
- **Exit Codes**:
  - **0**: Success (content fetched and output)
  - **1**: Any error (network, browser, auth, timeout, validation, conversion)
- **Rationale**:
  - Modern CLI best practice: keep it simple (gh, kubectl use 0/1)
  - Multiple exit codes are hard to document/discover
  - Error messages more useful than exit code numbers
  - Most scripts just check `$? != 0` (worked or didn't work)
  - Complexity without benefit
- **Sentinel Errors** (for internal logic/testing, not exit codes):
  ```go
  var (
      ErrBrowserNotFound    = errors.New("no Chromium-based browser found")
      ErrPageLoadTimeout    = errors.New("page load timeout exceeded")
      ErrAuthRequired       = errors.New("authentication required")
      ErrInvalidURL         = errors.New("invalid URL")
      ErrConversionFailed   = errors.New("HTML to Markdown conversion failed")
  )
  ```
- **Error Wrapping** (for context):
  ```go
  if err := page.Navigate(url); err != nil {
      return fmt.Errorf("failed to navigate to %s: %w", url, err)
  }
  ```
- **Error Messages** (clear + actionable):

  ```
  ✗ Authentication required for https://private.example.com
    Try: snag https://private.example.com --force-visible

  ✗ Page load timeout exceeded (5s)
    The page took too long to load. Try increasing timeout with --timeout

  ✗ No Chromium-based browser found
    Install Chrome, Chromium, Edge, or Brave to use snag
  ```

- **Main Function Pattern**:
  ```go
  func main() {
      if err := run(); err != nil {
          logger.Error(err.Error())
          os.Exit(1)
      }
      // Success (implicit exit 0)
  }
  ```
- **No Panic**: For expected errors, only truly unrecoverable situations
- **Benefits**:
  - Simple for users and scripts
  - Error messages do the heavy lifting
  - No documentation burden for exit codes
  - Go idiomatic (standard error handling patterns)
  - Easy to test (check for specific sentinel errors)

### 13. Project Structure

- **Decision**: Start with flat structure at repository root, refactor later if needed
- **Structure**:
  ```
  snag/
  ├── main.go              # CLI entry point (urfave/cli setup)
  ├── browser.go           # Browser management (rod)
  ├── fetch.go             # Page fetching logic
  ├── convert.go           # HTML to Markdown conversion
  ├── logger.go            # Custom logger
  ├── errors.go            # Sentinel errors
  ├── testdata/            # Test fixtures
  │   ├── simple.html
  │   └── auth-page.html
  ├── integration_test.go  # Real browser tests
  ├── go.mod
  ├── go.sum
  ├── LICENSE
  └── README.md
  ```
- **Build Command**:
  ```bash
  go build -o snag
  ```
- **Why Flat Structure**:
  - Simple and easy to navigate
  - Perfect for focused single-binary CLI
  - No over-engineering for MVP (<2000 lines expected)
  - Simpler Homebrew formula (`go build` vs `go build ./cmd/snag`)
  - Easy to refactor to `internal/` packages later if needed
  - Matches Go philosophy: "start simple, refactor as needed"
- **When to Refactor to internal/**:
  - Code grows beyond ~2000 lines
  - Phase 2/3 adds significant complexity
  - Multiple contributors need clear boundaries
  - Want to prevent external imports
- **No pkg/ Directory**: Not building a reusable library
- **Distribution Benefits**:
  - ✅ Simpler build commands
  - ✅ Clearer for contributors (main.go in root)
  - ✅ Less boilerplate in Homebrew formula
  - ✅ Standard for single-binary tools

### 14. Testing Strategy

- **Decision**: Integration tests with real Chrome/Chromium browser
- **Test Approach**:
  - **Integration tests**: Real browser via rod
  - **Test fixtures**: HTML files in `testdata/`
  - **Test server**: Local HTTP server for controlled tests
  - **Real websites**: Test against public sites (example.com, etc.)
- **Test Coverage**:
  - Normal page fetch (HTML → Markdown conversion)
  - Authentication detection (401, 403, login forms)
  - Page load timeout handling
  - Invalid URLs and error conditions
  - Browser connection modes (connect, headless, visible)
  - Output formats (markdown, html)
  - CLI flag handling (--timeout, --wait-for, etc.)
- **Test Structure**:

  ```go
  func TestFetchPage(t *testing.T) {
      // Start local test HTTP server
      // Launch real Chrome with rod
      // Fetch page
      // Assert Markdown output
      // Cleanup
  }

  func TestAuthDetection(t *testing.T) {
      // Serve page with 401 status
      // Attempt fetch
      // Assert ErrAuthRequired
  }
  ```

- **Why Real Browser**:
  - Tests actual CDP integration
  - Validates JavaScript execution
  - Catches real-world issues
  - Tests browser detection/connection
  - No mocking complexity
- **Test Data**: `testdata/` directory for HTML fixtures
- **CI/CD Considerations**:
  - Install Chrome/Chromium in CI environment
  - Run tests in headless mode
  - GitHub Actions has Chrome pre-installed
- **Rationale**:
  - Blackbox testing matches user experience
  - Real browser catches integration issues early
  - Simple test setup (no complex mocking)
  - Validates end-to-end flow

## Implementation Notes

The design outlined in this document was successfully implemented with all 14 design decisions realized in the initial release.

**Key Implementation Outcomes:**

- Flat project structure (main.go, browser.go, fetch.go, convert.go, logger.go, errors.go, validate.go)
- Custom logger with 4 levels and color/emoji support
- Integration tests with real Chrome/Chromium browser
- Multi-platform builds: darwin/arm64, darwin/amd64, linux/amd64, linux/arm64
- Single binary distribution (~5MB)

**Future Enhancements Under Consideration:**

- Tab management features (Phase 2)
- Additional output formats (text, pdf)
- Screenshot capabilities
- Cookie management

## References

**Inspiration:**

- `wget` - Classic web fetching
- `curl` - Flexible HTTP client
- `httpie` - User-friendly HTTP client
- `monolith` - Save complete web pages
- `shot-scraper` - Datasette's screenshot/HTML tool
