# snag - Design Document

## Overview

`snag` is a CLI tool for intelligently fetching web page content using a browser engine, with smart session management and format conversion capabilities.

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

## Technology Stack

**Language: Go**

Rationale:

- Single binary distribution (no runtime dependencies)
- Excellent cross-platform support (macOS, Linux, Windows)
- Native Chrome DevTools Protocol libraries (`chromedp` or `rod`)
- HTML to Markdown conversion libraries (`goldmark`, `gomarkdown`)
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

Position independent - flags can appear before or after URL:

```bash
snag <url> [options]
snag [options] <url>     # Both work identically
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

- `chromedp` - High-level CDP library (recommended for simplicity)
- `rod` - Alternative, more control, steeper learning curve

**HTML to Markdown:**

- `github.com/JohannesKaufmann/html-to-markdown` - Popular, well-maintained
- `github.com/gomarkdown/markdown` - Alternative

**CLI Framework:**

- `github.com/spf13/cobra` - Industry standard
- `github.com/urfave/cli` - Simpler alternative

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

- Build for multiple platforms: darwin/arm64, darwin/amd64, linux/amd64, linux/arm64, windows/amd64
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

**Current (Bash + Node.js):**

- `get-webpage` script (Bash wrapper)
- `fetch-html.js` (Puppeteer core)
- Dependencies: Node.js, Puppeteer, html2markdown (Go binary), bash_modules

**Migration Path:**

1. Port Node.js logic to Go using `chromedp`
2. Implement same CLI interface
3. Add Markdown conversion
4. Test feature parity
5. Distribute as binary

**Advantages:**

- No Node.js/npm required
- No bash_modules dependency
- Single 5MB binary vs multi-file installation
- Faster startup (no Node runtime)
- Better error messages (Go's type system)

## Success Criteria

**MVP Complete When:**

- [ ] Fetch URL and output Markdown to stdout
- [ ] Detect and connect to existing Chrome instance
- [ ] Launch headless Chrome when needed
- [ ] Detect authentication requirements
- [ ] Launch visible Chrome for auth flows
- [ ] Save output to file with `-o` flag
- [ ] Support `--format html` for raw output
- [ ] Support `--format markdown` (default)
- [ ] Implement `--version` flag
- [ ] Implement `--quiet` mode
- [ ] Implement `--user-agent` custom headers
- [ ] Position-independent argument parsing
- [ ] Homebrew formula working
- [ ] Basic documentation (README, --help)
- [ ] Test suite (unit + integration tests)

**Quality Gates:**

- [ ] Cross-platform builds (macOS arm64/amd64, Linux amd64)
- [ ] Unit tests for core functions
- [ ] Integration test with real websites
- [ ] Error handling for common failures
- [ ] Clean logging with `--verbose` flag

## Design Decisions Made

### 1. CLI Arguments ✅

- **Decision**: Use 16 arguments for MVP (position independent)
- **Rationale**:
  - All arguments from original `get-webpage` preserved
  - Added `--version`, `--quiet`, `--user-agent` as essential features
  - Replaced `--html` with `--format` for extensibility
  - Position independence is Go CLI framework default behavior
- **Status**: Complete - see CLI Arguments section above

### 2. Output Formats ✅

- **Decision**: MVP supports `markdown` (default) and `html` only
- **Post-MVP**: `text`, `pdf` as separate enhancement projects
- **Rationale**:
  - Keeps MVP scope focused
  - `text` and `pdf` add complexity (plain text extraction, Chrome PDF API)
  - Extensible via `--format` flag design
- **Status**: Complete

### 3. Argument Parsing ✅

- **Decision**: Position independent flags (standard Go CLI behavior)
- **Examples**: Both `snag -v example.com` and `snag example.com -v` work
- **Status**: Complete

## Open Questions

1. **Tab management priority**: How important is `--list-tabs` and `--tab` for MVP? → Post-MVP (Phase 2)
2. **Markdown library**: Which Go library produces cleanest output? → Pending
3. **Chrome bundling**: Should we consider bundling Chromium? → No (too large)
4. **Windows support**: Priority for initial release? → Probably later (darwin + linux first)
5. **Config file**: Support `.snagrc` for defaults? → Future consideration (post-MVP)
6. **CDP Library**: chromedp vs rod? → Pending
7. **HTML→Markdown**: Embed library vs shell out? → Pending
8. **Testing Strategy**: Unit vs integration, mocks vs real browser → Pending (see NOTES.md)

## Next Steps

1. Initialize Go module in repository
2. Set up basic CLI structure with Cobra
3. Implement Chrome detection and connection
4. Port page fetch logic from Node.js to chromedp
5. Add HTML to Markdown conversion
6. Create Homebrew tap repository
7. Set up GitHub Actions for releases
8. Write comprehensive README
9. Tag v1.0.0 release

## References

**Inspiration:**

- `wget` - Classic web fetching
- `curl` - Flexible HTTP client
- `httpie` - User-friendly HTTP client
- `monolith` - Save complete web pages
- `shot-scraper` - Datasette's screenshot/HTML tool

**Current Implementation:**

- `~/bin/scripts/get-webpage` (Bash wrapper)
- `~/bin/scripts/lib/chromium/fetch-html.js` (Puppeteer core)
