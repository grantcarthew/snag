# snag

Intelligently fetch web page content using a browser engine.

## Overview

`snag` is a CLI tool that fetches web page content using Chrome/Chromium via the Chrome DevTools Protocol. It can connect to existing browser sessions, launch headless browsers, or open visible browsers for authenticated sessions. Output can be Markdown or HTML.

## Installation

### Homebrew

**Note:** There's a name conflict with an older deprecated tool in Homebrew core. You must use the full tap name:

```bash
brew tap grantcarthew/tap
brew install grantcarthew/tap/snag
```

### Go Install

```bash
go install github.com/grantcarthew/snag@latest
```

### Build from source

```bash
git clone https://github.com/grantcarthew/snag.git
cd snag
go build
./snag --version
```

## Usage

### Basic usage

```bash
# Fetch and convert to Markdown (default)
snag https://example.com

# Save to file
snag -o output.md https://example.com

# Get raw HTML
snag --format html https://example.com

# Quiet mode (only output content)
snag --quiet https://example.com
```

### Flags

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
  --force-headless           Force headless mode even if Chrome is running
  --force-visible            Force visible mode for authentication
  --open-browser             Open Chrome browser in visible state (no URL required)
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

## Examples

### Simple fetch to stdout

```bash
snag https://example.com
```

### Save to file

```bash
snag -o docs.md https://docs.example.com/api
```

### Get raw HTML

```bash
snag --format html https://example.com > page.html
```

### Verbose logging

```bash
snag --verbose https://example.com
```

### Wait for dynamic content

```bash
snag --wait-for ".content-loaded" https://dynamic-site.com
```

### Longer timeout

```bash
snag --timeout 60 https://slow-site.com
```

### Authenticated sites (use existing browser session)

```bash
# First, manually open Chrome and log in to the site
# Then run snag - it will connect to your existing session
snag https://private.example.com
```

### Force visible mode for authentication

```bash
snag --force-visible https://private.example.com
```

## How It Works

**Smart Browser Management:**

1. **Session Detection**: Auto-detects existing Chrome instance with remote debugging enabled
2. **Mode Selection**:
   - If Chrome running → Connect to existing session (preserves auth/cookies)
   - If no Chrome → Launch headless mode
   - If `--force-visible` → Launch visible mode for user authentication
3. **Tab Management**: Keep tabs open in visible mode, close in headless mode (or use `--close-tab`)

**Output:**

- **stdout**: Content only (HTML/Markdown) - enables piping
- **stderr**: All logs, warnings, errors, progress indicators

## Requirements

- Chrome, Chromium, Edge, or Brave browser installed
- macOS or Linux (Windows not currently supported)

## License

Mozilla Public License 2.0

## Author

Grant Carthew <grant@carthew.net>
