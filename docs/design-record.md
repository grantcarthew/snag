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

**28 major design decisions documented below:**

| #   | Decision                  | Choice                                                          |
| --- | ------------------------- | --------------------------------------------------------------- |
| 1   | CLI Arguments             | 21 arguments, standard ordering: `snag [options] <url>`         |
| 2   | Output Formats            | Five formats: md, html, text, pdf, png                          |
| 3   | Argument Parsing          | Options before URL (simplicity over flexibility)                |
| 4   | Platform Support          | macOS (arm64/amd64), Linux (amd64/arm64) - Windows deferred     |
| 5   | Config File               | No config file support (permanent choice)                       |
| 6   | HTML→Markdown             | `html-to-markdown/v2` (embedded library)                        |
| 7   | License Attribution       | `LICENSES/` directory                                           |
| 8   | CLI Framework             | `urfave/cli/v2`                                                 |
| 9   | CDP Library               | `rod`                                                           |
| 10  | Browser Discovery         | rod's `launcher.LookPath()` (Chromium-based only)               |
| 11  | Logging Strategy          | Custom logger (4 levels, colors, emojis)                        |
| 12  | Error Handling            | Exit 0/1, sentinel errors, clear messages                       |
| 13  | Project Structure         | Flat structure at root                                          |
| 14  | Testing Strategy          | Integration tests with real browser                             |
| 15  | Flag Assignment           | `-t` moved from `--timeout` to `--tab` (more frequently used)   |
| 16  | Tab Indexing              | 1-based indexing (first tab is [1], not [0])                    |
| 17  | Pattern Matching          | Progressive fallthrough (exact → contains → regex)              |
| 18  | Case Sensitivity          | Case-insensitive matching for all modes                         |
| 19  | Regex Support             | Full regex patterns (not just wildcards)                        |
| 20  | Pattern Simplicity        | No regex detection needed (try all methods in order)            |
| 21  | Multiple Matches          | First match wins (predictable, simple)                          |
| 22  | Performance               | Single-pass page.Info() caching (3x improvement)                |
| 23  | Format Name Normalization | `md` (not "markdown") for consistency with file extensions      |
| 24  | Format Alias Support      | Case-insensitive formats with aliases (markdown→md, txt→text)   |
| 25  | Screenshot → PNG Format   | Remove `--screenshot` flag, use `--format png` (consistency)    |
| 26  | Binary Format Safety      | Auto-generate filenames for PDF/PNG (prevent stdout corruption) |
| 27  | Tab Range Support         | Support range syntax `N-M` for fetching multiple tabs           |
| 28  | Tab Ordering              | Sort tabs by URL→Title→ID (predictable, stable ordering)        |

See detailed rationale in design decisions section below.

## Detailed Argument Specifications

For complete argument validation rules, interaction matrices, error messages, and usage examples, see the [argument documentation](arguments/):

- **General**: [Validation Rules](arguments/validation.md) | [README](arguments/README.md)
- **Content Sources**: [URL](arguments/url.md) | [URL File](arguments/url-file.md) | [Tab](arguments/tab.md) | [All Tabs](arguments/all-tabs.md)
- **Output**: [Output](arguments/output.md) | [Output Dir](arguments/output-dir.md) | [Format](arguments/format.md)
- **Browser**: [Open Browser](arguments/open-browser.md) | [Force Headless](arguments/force-headless.md) | [List Tabs](arguments/list-tabs.md) | [Close Tab](arguments/close-tab.md)
- **Page Control**: [Wait For](arguments/wait-for.md) | [Timeout](arguments/timeout.md) | [Port](arguments/port.md)
- **Request**: [User Agent](arguments/user-agent.md) | [User Data Dir](arguments/user-data-dir.md)
- **Logging**: [Verbose](arguments/verbose.md) | [Quiet](arguments/quiet.md) | [Debug](arguments/debug.md)
- **Info**: [Help](arguments/help.md) | [Version](arguments/version.md)

The design record documents **WHY** decisions were made; the argument specs document **HOW** they work.

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

- Markdown (default) - clean, readable text format
- HTML - raw HTML output via `--format html`
- Text - plain text only (strips all HTML) via `--format text`
- PDF - visual rendering as document via `--format pdf`
- PNG - visual rendering as image via `--format png`

## What It Does NOT Do

**Explicit Non-Goals:**

- ❌ Remote control browser (clicking, form filling, multi-step workflows)
- ❌ Web scraping framework
- ❌ JavaScript execution/testing
- ❌ Performance profiling
- ❌ Video recording or animated GIF capture

**Philosophy**: `snag` is a **passive observer** for content retrieval, not an automation framework. For browser automation, use Puppeteer, Playwright, or Selenium.

## Feature Set

### Core Features

- Fetch URL and return content
- Smart Chrome session detection
- Three operation modes: connect, headless, visible
- Authentication detection and handling
- HTML to Markdown conversion
- Output to stdout or file

**CLI Arguments:**

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
  -d, --output-dir <dir>     Save files with auto-generated names to directory
  --format <FORMAT>          Output format: md (default) | html | text | pdf | png
```

**Page Loading:**

```
  --timeout <seconds>        Page load timeout in seconds (default: 30)
  -w, --wait-for <selector>  Wait for CSS selector before extracting content
```

**Browser Control:**

```
  -p, --port <port>          Chrome remote debugging port (default: 9222)
  -c, --close-tab            Close the browser tab after fetching content
  --force-headless           Force headless mode even if Chrome is running
  -b, --open-browser         Open Chrome browser in visible state (no URL required)
  -l, --list-tabs            List all open tabs in browser
  -t, --tab <pattern>        Fetch content from existing tab (index or pattern)
  -a, --all-tabs             Process all open tabs
  --user-data-dir <dir>      Custom browser profile directory
```

**URL Input:**

```
  --url-file <file>          Read URLs from file (one per line)
```

**Logging/Debugging:**

```
  --verbose                  Enable verbose logging output
  -q, --quiet                Suppress all output except errors and content
  --debug                    Enable debug output
```

**Request Control:**

```
  --user-agent <string>      Custom user agent string (bypass headless detection)
```

### Tab Management Features

```bash
snag --list-tabs                    # List all open tabs (1-based indexing)
snag -l                             # Short alias

snag --tab <index>                  # Get content from specific tab (1-based)
snag -t <index>                     # Short alias
snag -t <pattern>                   # Match tab by regex/URL pattern
```

**Key Design Decisions:**

1. **Flag Assignment**: Moved `-t` alias from `--timeout` to `--tab` (more frequently used)
2. **Tab Indexing**: 1-based indexing (first tab is [1], not [0]) - more intuitive for users
3. **Pattern Matching**: Progressive fallthrough with full regex support:
   - Range (N-M) → Multiple consecutive tabs (e.g., `1-3` = tabs 1, 2, 3)
   - Integer → Tab index (1-based)
   - Exact URL match (case-insensitive)
   - Substring/contains match (case-insensitive)
   - Regex match (case-insensitive, fallback)
   - Error if no matches
4. **Case Sensitivity**: All matching is case-insensitive for better UX
5. **Regex Support**: Full regex patterns (not just wildcards) - same complexity, more power
6. **Pattern Simplicity**: No regex detection needed - try all methods in order for every pattern
7. **Multiple Matches**: First match wins (predictable, simple)
8. **Performance**: Single-pass page.Info() caching (3x improvement over naive implementation)

**Use Cases:**

- Get content from already-open authenticated tab without creating new tabs
- Work with existing browser sessions (preserves cookies, auth state)
- Pattern-based tab selection (regex, exact match, or substring)
- Range-based batch processing of consecutive tabs (e.g., `1-5`)
- List all tabs to see what's available

**Examples:**

```bash
# List tabs (1-based indexing)
snag --list-tabs
# Output:
#   [1] https://github.com/grantcarthew/snag - snag repo
#   [2] https://go.dev/doc/ - Go docs
#   [3] https://app.internal.com/dashboard - Dashboard

# Fetch by index (1-based)
snag -t 1                                # First tab
snag -t 3                                # Third tab

# Fetch by range (1-based)
snag -t 1-3                              # First three tabs (auto-save)
snag -t 4-6 -d ./output/                 # Tabs 4-6 to directory

# Fetch by exact URL (case-insensitive)
snag -t https://github.com/grantcarthew/snag
snag -t GITHUB.COM                       # Case-insensitive

# Fetch by regex pattern
snag -t "github\.com/.*"                 # Regex: github.com/ + anything
snag -t ".*/dashboard"                   # Regex: ends with /dashboard
snag -t "(github|gitlab)\.com"           # Regex: alternation

# Fetch by substring (fallback)
snag -t "dashboard"                      # Contains "dashboard"
snag -t "github"                         # Contains "github"
```

**Technical Implementation:**

- New `TabInfo` struct (index, URL, title, ID) - browser.go:49-55
- `ListTabs()` - Get all tabs from browser - browser.go:404-434
- `GetTabByIndex(index int)` - Select tab by 1-based index - browser.go:434-463
- `GetTabByPattern(pattern string)` - Progressive fallthrough matching with caching - browser.go:473-544
- `handleListTabs()` - CLI handler for --list-tabs - main.go:345-383
- `handleTabFetch()` - CLI handler for --tab - main.go:412-534
- Browser requirement: Existing Chrome instance with remote debugging
- Integration tests: cli_test.go (13 tab-related tests)

**Rationale:**

- **1-based indexing**: UI tool for humans, not a programming API
- **Pattern matching order**: Exact → contains → regex prioritizes common cases for performance
- **No regex detection**: Simpler code, more predictable behavior (try all methods)
- **Full regex**: Same implementation cost as wildcards, more flexibility
- **Progressive fallthrough**: Maximizes chances of finding the right tab
- **Case-insensitive**: Better UX, users don't worry about capitalization
- **First match wins**: Simple, predictable, documented behavior
- **Performance optimization**: Caching eliminates redundant network calls (3x improvement)

### Additional Output Formats

```bash
snag --format text <url>              # Plain text extraction (strips all HTML)
snag --format pdf <url>               # PDF export using Chrome rendering
snag --format png <url>               # PNG screenshot capture (full page)
```

**Format Features:**

- Text format uses plain text extraction (no HTML/Markdown)
- PDF format uses Chrome's native PDF rendering API
- PNG format captures full-page screenshots (replaced `--screenshot` flag)
- All formats auto-generate filenames for binary outputs (PDF, PNG)
- Format name normalization: `md` (not "markdown") for consistency
- Alias support: `markdown` → `md`, `txt` → `text` (backward compatibility)
- Case-insensitive format input

### Future Features Under Consideration

**Tab Enhancements:**

```
  --tab-all <pattern>        Fetch from all matching tabs (batch processing)
  --list-tabs --format json  JSON output for scripting
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

## Design Philosophy

- Single-purpose tool: fetch web page content via browser engine
- Smart session management: connect to existing Chrome or launch new instance
- Format flexibility: Markdown (default) or raw HTML output
- Unix philosophy: do one thing well, output to stdout for piping

**Technical Benefits:**

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

- **Decision**: Support five output formats: `md`, `html`, `text`, `pdf`, and `png`
- **Rationale**:
  - Started with essential text formats (Markdown, HTML)
  - Added `text` for plain text extraction (no HTML/Markdown)
  - Added `pdf` using Chrome's native PDF rendering API
  - Added `png` for visual capture (replaces dedicated `--screenshot` flag)
  - All formats treated consistently via `--format` flag
  - Extensible design allows future format additions

### 3. Argument Parsing

- **Decision**: Standard flag ordering - options before URL: `snag [options] <url>`
- **Rationale**:
  - Simpler implementation and maintenance
  - Consistent with most CLI tools
  - urfave/cli v2 doesn't natively support position-independent args
  - Avoiding additional parsing complexity was prioritized

**See Also**: For complete validation order and cross-cutting rules, see [arguments/validation.md](arguments/validation.md)

### 4. Platform Support

- **Decision**: Target macOS and Linux; Windows is future consideration
- **Supported Platforms**:
  - darwin/arm64 (macOS Apple Silicon)
  - darwin/amd64 (macOS Intel)
  - linux/amd64 (Linux 64-bit)
  - linux/arm64 (Linux ARM - Raspberry Pi, servers)
- **Future**: Windows support (requires Windows-specific path handling)
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
- **Future Consideration**: `snag --licenses` command to print embedded licenses
- **Automation**: Use `go-licenses` tool during build/release process

### 8. CLI Framework Choice

- **Decision**: Use `github.com/urfave/cli` for CLI framework
- **Library**: urfave/cli v2 (MIT license)
- **Rationale**:
  - Smaller binary size compared to Cobra (important for single-binary tool)
  - Simpler, less boilerplate code
  - Better dynamic bash autocompletion (can autocomplete argument values)
  - Supports subcommands (e.g., --list-tabs, --tab)
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
  - **Quiet** (`--quiet`): Only errors to stderr (all error types, no warnings)
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
  (no output unless error occurs)
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

**See Also**: For output file specifications and validation rules, see [arguments/output.md](arguments/output.md) and [arguments/output-dir.md](arguments/output-dir.md)

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
    Try: snag --open-browser https://private.example.com

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
  - No over-engineering for small codebase
  - Simpler Homebrew formula (`go build` vs `go build ./cmd/snag`)
  - Easy to refactor to `internal/` packages later if needed
  - Matches Go philosophy: "start simple, refactor as needed"
- **When to Refactor to internal/**:
  - Code grows beyond ~2000 lines
  - Features add significant complexity
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

### 15. Flag Assignment

- **Decision**: Move `-t` alias from `--timeout` to `--tab`
- **Rationale**:
  - `--tab` will be used far more frequently than custom timeouts
  - Most users will use default 30s timeout (rarely need to change)
  - Shorter alias should go to more commonly used flag
  - Power users who need custom timeouts can type `--timeout 60`
- **Breaking Change**: Yes (users using `-t` for timeout will need to use `--timeout`)
- **Migration**: Document in release notes, minimal impact expected

### 16. Tab Indexing

- **Decision**: Use 1-based indexing (first tab is [1], not [0])
- **Rationale**:
  - snag is a UI tool for humans, not a programming API
  - 1-based indexing is more intuitive for end users
  - Matches how users think about lists ("first tab", "second tab")
  - Most CLI tools that list items use 1-based indexing
  - Internal arrays still use 0-based (just offset for display/input)
- **Implementation**: Convert user input (1-based) to array index (0-based) internally
- **Output**: Display tabs as [1], [2], [3], etc.

### 17. Pattern Matching

- **Decision**: Progressive fallthrough matching with 4 stages
- **Matching Process**:
  1. **Integer check**: Parse as int → Use as tab index (1-based)
  2. **Exact match**: Try case-insensitive exact URL match
  3. **Contains match**: Try case-insensitive substring search
  4. **Regex match**: Compile and match as fallback (case-insensitive)
  5. **Error**: If no matches found at any stage
- **Rationale**:
  - **Order changed during implementation**: Originally planned regex-first, changed to exact → contains → regex
  - Most common cases (exact URL, substring) hit first for better performance
  - Simpler patterns get priority over complex regex
  - Regex as fallback catches advanced use cases
  - Progressive fallthrough maximizes chances of finding the right tab
  - Simple patterns work automatically (no need to learn regex)
  - Power users can use full regex when needed
  - Forgiving approach: "try everything before failing"
- **First match wins**: If multiple tabs match, return first one (predictable, simple)
- **Implementation**: browser.go:473-544 (GetTabByPattern function)

### 18. Case Sensitivity

- **Decision**: All pattern matching is case-insensitive
- **Implementation**:
  - Regex: Use `(?i)` flag
  - Exact match: Use `strings.EqualFold()`
  - Contains match: Convert both strings to lowercase
- **Rationale**:
  - Better user experience (don't worry about capitalization)
  - URLs are typically lowercase but users might type them differently
  - Matches user expectations (most search is case-insensitive)
  - No performance penalty (minimal overhead)

### 19. Regex Support

- **Decision**: Support full regex patterns (not just wildcards)
- **Rationale**:
  - Implementation complexity is identical (using `regexp` package internally anyway)
  - Maximum flexibility for power users
  - Simple users can still use basic patterns (substring matching fallback)
  - Can document common patterns in README
  - No artificial limitations
- **Alternative Considered**: Wildcard-only (`*`) - rejected as same implementation cost
- **User Support**: Provide clear examples and error messages for invalid regex

### 20. Pattern Simplicity

- **Decision**: No regex character detection - simply try all matching methods in order
- **Rationale**:
  - **Simplified during implementation**: Originally planned `hasRegexChars()` detection, removed as unnecessary
  - Cleaner implementation: just try exact → contains → regex for every pattern
  - No detection logic needed, no edge cases to handle
  - User suggestion led to simpler, more elegant solution
  - Performance impact negligible (string comparisons are fast)
  - More predictable behavior (always tries all methods)
- **Alternative Rejected**: Detecting regex chars first, then routing to specific matcher
- **Implementation**: No `hasRegexChars()` function needed

### 21. Multiple Matches

- **Decision**: First match wins when multiple tabs match pattern
- **Rationale**:
  - Simple and predictable behavior
  - Consistent with other tools (grep, find, etc.)
  - Users can use `--list-tabs` to see tab order
  - Can add `--tab-all` flag in future for multiple matches
- **Documentation**: Clearly document that first match is returned
- **Verbose Mode**: Show which tab matched and why

### 22. Performance Optimization

- **Decision**: Single-pass page.Info() caching in GetTabByPattern()
- **Problem Identified**: Multiple `page.Info()` calls repeated for same pages (network round-trips)
- **Solution**: Cache page.Info() results once, iterate over cached data
- **Impact**: 3x reduction in network calls
  - Before: 3N calls for N tabs (worst case: exact + contains + regex each call Info())
  - After: N calls for N tabs (single pass to build cache, then iterate)
  - Example: 10 tabs = 30 calls → 10 calls
- **Implementation**:
  - Local `pageCache` struct stores page, URL, and index
  - Single loop at browser.go:487-507
  - All pattern matching operates on cached data
- **Rationale**:
  - Identified during code review after initial implementation
  - Significant performance improvement with minimal code complexity
  - Network calls are expensive compared to memory operations
  - Maintains exact same behavior, just faster
- **Code Location**: browser.go:487-507 (GetTabByPattern function)

**See Also**: For complete tab selection specification and examples, see [arguments/tab.md](arguments/tab.md) and [arguments/list-tabs.md](arguments/list-tabs.md)

### 23. Format Name Normalization

- **Decision**: Change `FormatMarkdown` from "markdown" to "md"
- **Rationale**:
  - **Consistency**: All other formats use short names matching file extensions (html, text, pdf, png)
  - **Only outlier**: "markdown" was the only long-form name (8 characters vs 2-4 for others)
  - **Matches extension**: Files are saved as `.md`, so format should be `md`
  - **Less typing**: Shorter format name for most commonly used format
  - **Predictability**: Users can guess format from file extension (`.md` → `md`, `.pdf` → `pdf`)
- **Breaking Change**: Yes, but acceptable pre-v1.0
- **Migration Path**: Backward compatibility via alias support (see Decision 24)
- **Implementation**:
  - Changed `FormatMarkdown = "md"` constant (was "markdown")
  - Updated all format validation and help text
  - Updated tests to use canonical "md" name
- **Impact**: More consistent, predictable CLI interface

### 24. Format Alias Support

- **Decision**: Support case-insensitive format input with backward-compatible aliases
- **Aliases Supported**:
  - `"markdown"` → `"md"` (backward compatibility)
  - `"txt"` → `"text"` (common alias)
- **Case Insensitivity**: All format inputs converted to lowercase before validation
  - `"MARKDOWN"` → `"markdown"` → `"md"` ✅
  - `"Png"` → `"png"` ✅
  - `"HTML"` → `"html"` ✅
- **Implementation**:
  - `normalizeFormat()` function in validate.go:124-141
  - Called before format validation in all handlers
  - Lowercase conversion + alias mapping
- **Rationale**:
  - **Better UX**: Users don't worry about capitalization
  - **Smooth migration**: Existing scripts using "markdown" continue working
  - **Common expectations**: Users expect "txt" to work for text files
  - **No complexity penalty**: Simple map lookup, negligible performance impact
  - **Future-proof**: Easy to add more aliases if needed
- **Code Location**: validate.go:124-141 (normalizeFormat function)

### 25. Screenshot → PNG Format Refactor

- **Decision**: Remove `--screenshot` flag, make PNG a regular format via `--format png`
- **Before**:
  ```bash
  snag --screenshot https://example.com       # Special flag
  snag --format pdf https://example.com       # Format flag
  ```
- **After**:
  ```bash
  snag --format png https://example.com       # PNG is just another format
  snag --format pdf https://example.com       # Consistent approach
  ```
- **Rationale**:
  - **Consistency**: All visual outputs (PDF, PNG) treated as formats, not special cases
  - **Eliminates code smell**: Removed parameter interdependency between `screenshot bool` and `format string`
  - **Simpler logic**: No special-case handling throughout codebase
  - **Semantic consistency**: PDF and PNG are both "visual captures" (not content extraction)
  - **Cleaner Config struct**: Removed redundant `Screenshot bool` field
  - **One way to do it**: Eliminates confusion about screenshot vs format png
- **Breaking Change**: Yes, but acceptable pre-v1.0
  - Old: `snag --screenshot https://example.com`
  - New: `snag --format png https://example.com`
- **Implementation Impact**:
  - Removed `--screenshot` CLI flag (main.go:93-96)
  - Removed `Screenshot bool` from Config struct (handlers.go:28)
  - Removed `screenshot` parameter from 2 helper functions
  - Updated 7 call sites in handlers.go
  - Updated formats.go to use `FormatPNG` constant
  - Simplified conditional logic (no special cases)
- **Benefits**:
  - 19 lines removed from handlers.go
  - Cleaner function signatures
  - More maintainable codebase
  - Consistent user experience

### 26. Binary Format Safety

- **Decision**: Auto-generate filenames for binary formats (PDF, PNG) when no output file specified
- **Behavior**:
  - Text formats (md, html, text): Output to stdout by default
  - Binary formats (pdf, png): Auto-generate filename to current directory
- **Rationale**:
  - **Terminal corruption prevention**: Binary data to stdout corrupts terminal display
  - **Better UX**: No need to remember `-o` flag for binary formats
  - **Safety first**: Prevents accidental terminal damage
  - **Sensible defaults**: Users expect binary files to be saved
- **Implementation**:
  - Check if output file not specified: `if config.OutputFile == ""`
  - Check if binary format: `if format == FormatPDF || format == FormatPNG`
  - Auto-generate filename with timestamp and page title
  - Save to current directory (handlers.go:118, 441)
- **Auto-Generated Filename Format**:
  ```
  yyyy-mm-dd-hhmmss-{page-title-slug}.{ext}
  Example: 2025-10-22-142033-github-snag-repo.png
  ```
- **User Override**: Can still use `-o` or `-d` to specify custom location
- **Code Location**: handlers.go:116-133, 439-450

**See Also**: For complete format specification, validation rules, and interaction matrix, see [arguments/format.md](arguments/format.md)

### 27. Tab Range Support

- **Decision**: Support range syntax `N-M` in `--tab` flag for fetching multiple consecutive tabs
- **Syntax**: `--tab 1-3` fetches tabs 1, 2, and 3
- **Rationale**:
  - **Intuitive**: Range notation is common and expected in CLI tools
  - **Complements existing features**: Works alongside single tab (`--tab 1`) and all tabs (`--all-tabs`)
  - **Easy to implement**: Simple regex pattern detection (`^\d+-\d+$`) before other pattern matching
  - **Useful for workflows**: Batch process specific subset of tabs without processing all
  - **Consistent with 1-based indexing**: Ranges use same indexing as single tab selection
- **Design Constraints**:
  - **No reverse ranges**: `3-1` is invalid (error: "start must be <= end")
  - **No non-contiguous ranges**: Only `N-M` syntax, not `1,3,5` (keeps implementation simple)
  - **Fail fast**: Stops at first missing tab in range (e.g., `1-100` fails when tab doesn't exist)
  - **Range detection first**: Checked before integer/string pattern matching in priority order
- **Behavior**:
  - Multiple tabs fetched → Behaves like `--all-tabs` (auto-save, no stdout)
  - Cannot use `--output` flag with ranges (error)
  - Can use `--output-dir` to specify save directory
  - Without output flags: Auto-saves to current directory with generated filenames
- **Validation Rules**:
  - Both start and end must be >= 1
  - Start must be <= end
  - Both indices must exist in browser (fails at first missing tab)
  - Empty range parts invalid: `1-`, `-3` → Error
- **Examples**:
  ```bash
  snag --tab 1-3                  # Fetch tabs 1, 2, 3 (auto-save)
  snag -t 4-6 -d ./output/        # Tabs 4-6 to directory
  snag -t 1-5 --format pdf        # Tabs 1-5 as PDFs
  snag -t 2-2                     # Single tab (equivalent to --tab 2)
  ```
- **Integration**: Fits naturally into existing progressive fallthrough pattern matching as highest priority check
- **Code Location**: Handler in `main.go` (range detection before pattern matching)

**See Also**: For complete tab range specification and examples, see [arguments/tab.md](arguments/tab.md)

### 28. Tab Ordering

- **Decision**: Sort tabs by URL (primary), Title (secondary), ID (tertiary) for all tab operations
- **Sort Order**: Alphabetical multi-level sort provides stable, predictable ordering
- **Rationale**:
  - **CDP Limitation**: Chrome DevTools Protocol returns tabs in unpredictable internal order (not visual left-to-right)
  - **Predictability**: Same tab order every time, regardless of CDP's internal state
  - **Reproducibility**: Automation workflows can rely on consistent tab indices
  - **Multi-level stability**: URL→Title→ID ensures unique sort position even for identical URLs
  - **User clarity**: Clear, documented ordering (alphabetical by URL) vs. unexplained random order
- **Impact on All Tab Operations**:
  - `--list-tabs`: Displays tabs in sorted order with "sorted by URL" message
  - `--tab <index>`: Selects tab by sorted position (tab [1] = first alphabetically by URL)
  - `--tab <range>`: Processes tabs in sorted order
  - `--all-tabs`: Processes tabs in sorted order
- **Performance**:
  - Single `page.Info()` call per tab (efficient network usage)
  - In-memory sorting after info fetch (no additional CDP calls)
  - `getSortedPagesWithInfo()` caches info for reuse across operations
- **Tradeoffs**:
  - **Pro**: Predictable, stable, reproducible automation
  - **Pro**: Same order across all tab operations (consistency)
  - **Pro**: No configuration needed (opinionated design)
  - **Con**: Tab indices don't match visual browser order
  - **Accepted**: Visual order is unreliable anyway (CDP limitation), sorted order is better than random
- **User Communication**:
  - `--list-tabs` output shows "sorted by URL" in header
  - Success messages mention "from sorted order (by URL)" for clarity
  - Documentation clearly explains tab [1] ≠ first visual tab
- **Examples**:
  ```bash
  # Visual browser order (left to right):
  #   - example.com
  #   - google.com/search
  #   - github.com/repo

  # snag --list-tabs output (sorted by URL):
  #   [1] https://example.com
  #   [2] https://github.com/repo
  #   [3] https://google.com/search

  snag --tab 1    # Fetches example.com (first alphabetically)
  snag --tab 1-2  # Fetches example.com and github.com
  ```
- **Code Location**:
  - Sort function: `browser.go:409-451` (`getSortedPagesWithInfo()`)
  - Used by: `ListTabs()`, `GetTabByIndex()`, `GetTabsByRange()`
- **Rejected Alternatives**:
  - **No sorting (use CDP order)**: Rejected - unpredictable, not reproducible
  - **Visual tab order**: Not possible via CDP
  - **Configurable sort order**: Rejected - adds complexity, opinionated design is better
  - **Sort by title only**: Less stable (multiple tabs can have same title)

**See Also**: For tab ordering documentation, see [arguments/list-tabs.md](arguments/list-tabs.md)

## References

**Inspiration:**

- `wget` - Classic web fetching
- `curl` - Flexible HTTP client
- `httpie` - User-friendly HTTP client
- `monolith` - Save complete web pages
- `shot-scraper` - Datasette's screenshot/HTML tool
