# snag

Intelligently fetch web page content using a browser engine.

## Why snag?

**Built for AI agents to consume web content efficiently.**

Modern AI agents need web content in clean, token-efficient formats. snag solves this by:

- **Markdown output** - AI models work better with markdown than HTML (70% fewer tokens)
- **Real browser rendering** - Handles JavaScript, dynamic content, lazy loading automatically
- **Authentication support** - Access private/authenticated pages through persistent browser sessions
- **Content archival** - Build reference libraries of web content for future AI agent use
- **Simple CLI interface** - One command, clean output, no complex automation scripts

**Perfect for:**

- AI coding assistants fetching documentation
- Building knowledge bases from authenticated sites
- Capturing dynamic web content for analysis
- Piping web content into AI processing pipelines

## Quick Start

```bash
# Install via Homebrew
brew tap grantcarthew/tap
brew install grantcarthew/tap/snag

# Fetch a page as markdown
snag example.com

# Save to file
snag docs.example.com > docs.md
```

That's it! snag auto-detects your Chromium-based browser and handles everything else.

## Installation

### Prerequisites

snag requires a Chromium-based (Chrome) browser:

**Linux:**

```bash
# Ubuntu/Debian
sudo apt update && sudo apt install chromium-browser

# Fedora
sudo dnf check-update && sudo dnf install chromium

# Arch Linux
sudo pacman -Sy chromium

# Homebrew
brew install chromium
```

**macOS:**

```bash
# Chromium (recommended) via Homebrew
# or Chrome - download from https://www.google.com/chrome/
brew install chromium
```

**Supported browsers:** Chrome, Chromium, Microsoft Edge, Brave, other Chromium-based browsers

### Install snag

**Homebrew (Linux/macOS):**

Note: There's a name conflict with an older deprecated tool. Use the full tap name:

```bash
brew tap grantcarthew/tap
brew install grantcarthew/tap/snag
```

**Go Install:**

```bash
go install github.com/grantcarthew/snag@latest
```

**Build from Source:**

```bash
git clone https://github.com/grantcarthew/snag.git
cd snag
go build
./snag --version
```

## Usage

### Basic Examples

```bash
# Fetch page as Markdown (default)
snag example.com
snag https://example.com

# Save to file
snag -o output.md https://example.com
snag example.com > output.md

# Get raw HTML instead
snag --format html https://example.com

# Quiet mode (content only, no logs)
snag --quiet https://example.com

# Wait for dynamic content to load
snag --wait-for ".content-loaded" https://dynamic-site.com

# Increase timeout for slow sites
snag --timeout 60 https://slow-site.com

# Verbose logging for debugging
snag --verbose https://example.com
```

## Common Scenarios

### AI Agent Documentation Fetching

```bash
# Fetch API documentation for AI context
snag https://api.example.com/docs > api-reference.md

# Pipe directly to AI assistant
snag --quiet https://docs.python.org/3/library/os.html | your-ai-tool
```

### Building a Knowledge Base

```bash
# Save multiple pages to a reference directory
snag -o reference/golang-basics.md https://go.dev/doc/tutorial/getting-started
snag -o reference/golang-concurrency.md https://go.dev/doc/effective_go#concurrency
snag -o reference/golang-errors.md https://go.dev/blog/error-handling-and-go
```

### Fetching Dynamic Content

```bash
# Wait for JavaScript to render content
snag --wait-for "#main-content" https://single-page-app.com

# Give slow sites more time
snag --timeout 90 --wait-for ".loaded" https://heavy-site.com
```

### Batch Processing URLs

```bash
# Process URLs from a file
while read url; do
  filename=$(echo "$url" | sed 's/[^a-zA-Z0-9]/_/g').md
  snag --quiet -o "$filename" "$url"
done < urls.txt

# Combine multiple pages
for url in https://example.com/page1 https://example.com/page2; do
  snag --quiet "$url" >> combined.md
  echo -e "\n---\n" >> combined.md
done
```

### CI/CD Integration

```bash
# Fetch documentation in CI pipeline
snag --force-headless --timeout 30 https://docs.example.com > docs.md

# Quiet mode for clean logs
snag --quiet --force-headless https://example.com > output.md
```

## Authentication

snag makes it easy to fetch content from authenticated/private sites using persistent browser sessions.

### Method 1: Visible Browser Mode

Open a browser, authenticate manually, then snag connects to it:

```bash
# Step 1: Open browser in visible mode and log in manually
# Note: Using the --open-browser (-b) switch enables the required DevTools protocol
snag --open-browser

# Step 2: In the browser window, navigate to your site and log in
# (Leave the browser open)

# Step 3: Fetch authenticated content - snag reuses your session
snag https://private.example.com

# Step 4: Fetch more pages with the same session
snag https://private.example.com/dashboard
snag https://private.example.com/settings
```

### Method 2: Force Visible Mode

Let snag launch the browser for you:

```bash
# Launch visible browser and navigate to page
snag --force-visible https://private.example.com

# Authenticate in the browser window that opens
# Then close it or leave it running

# Subsequent calls reuse the session
snag https://private.example.com/other-page
```

### Method 3: Existing Chromium Session

Keep one browser session for multiple snag calls:

```bash
# Terminal 1: Start Chromium with remote debugging
chromium --remote-debugging-port=9222 --user-data-dir=/tmp/chromium-profile

# Log in to your sites manually in this browser

# Terminal 2: Use snag with the existing session
snag https://authenticated-site1.com
snag https://authenticated-site2.com
snag https://authenticated-site3.com
```

All three commands share authentication state - no repeated logins required!

## Advanced Usage

### Custom User Agent

Bypass headless detection or mimic specific browsers:

```bash
# Linux Firefox user agent
snag --user-agent "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0" \
  https://example.com

# Custom bot identifier
snag --user-agent "MyBot/1.0 (+https://example.com/bot)" \
  https://api-docs.example.com
```

### Debugging Failed Fetches

```bash
# See what's happening during fetch
snag --verbose https://problematic-site.com

# Full debug output including browser messages
snag --debug https://problematic-site.com 2> debug.log

# Open browser visibly to see what snag sees
snag --force-visible https://problematic-site.com
```

### Tab Management

```bash
# Close tab after fetching (default in headless mode)
snag --close-tab https://example.com

# Keep tab open (default in visible mode)
snag https://example.com
```

### Custom Remote Debugging Port

```bash
# Use different port if 9222 is busy
snag --port 9223 https://example.com

# Connect to Chromium running on custom port
chromium --remote-debugging-port=9223 &
snag --port 9223 https://example.com
```

## CLI Reference

### Core Arguments

```
<url>                      URL to fetch (required, unless --open-browser)
--version                  Display version information
-h, --help                 Show help message and exit
```

### Output Control

```
-o, --output <file>        Save output to file instead of stdout
--format <type>            Output format: markdown (default) | html
```

### Page Loading

```
-t, --timeout <seconds>    Page load timeout in seconds (default: 30)
-w, --wait-for <selector>  Wait for CSS selector before extracting content
```

### Browser Control

```
-p, --port <port>          Chromium remote debugging port (default: 9222)
-c, --close-tab            Close the browser tab after fetching content
--force-headless           Force headless mode even if Chromium is running
--force-visible            Force visible mode for authentication
--open-browser             Open Chromium browser in visible state (no URL required)
```

### Logging/Debugging

```
--verbose                  Enable verbose logging output
-q, --quiet                Suppress all output except errors and content
--debug                    Enable debug output with CDP messages
```

### Request Control

```
--user-agent <string>      Custom user agent string (bypass headless detection)
```

## Troubleshooting

### Browser Issues

**"Browser not found" error**

snag cannot locate Chrome/Chromium on your system.

Solutions:

- Install Chromium: `brew install chromium`
- Install Chrome from https://www.google.com/chrome/
- Ensure Chromium/Chrome is in your system PATH

**"Failed to connect to existing browser"**

Cannot connect to running browser instance.

Solutions:

- Ensure Chromium/Chrome is launched with `--remote-debugging-port=9222`
- Try different port: `snag --port 9223 https://example.com`
- Kill existing Chromium/Chrome processes and let snag launch a new instance

### Authentication Issues

**"Authentication required" error**

Page requires login but snag cannot authenticate.

Solutions:

- Use `--force-visible` to manually log in: `snag --force-visible https://example.com`
- Open browser with `snag --open-browser`, log in, then run snag again
- Browser session persists authentication across snag calls

### Timeout Issues

**"Page load timeout" error**

Page takes too long to load.

Solutions:

- Increase timeout: `snag --timeout 60 https://example.com`
- Use `--wait-for` for specific element: `snag --wait-for ".content" https://example.com`
- Check network connectivity
- Try `--verbose` to see what's happening

**Page loads but content is missing**

Dynamic content hasn't appeared yet.

Solutions:

- Use `--wait-for` with selector: `snag --wait-for "#main-content" https://example.com`
- Increase timeout to allow for slow loading
- Inspect page with `--format html` to see raw output

### Output Issues

**Output is empty or incomplete**

Fetched page but content is missing.

Solutions:

- Try `--format html` to see raw HTML
- Use `--verbose` to check if page loaded correctly
- Page may require authentication (see authentication section)
- Content may be loaded dynamically (use `--wait-for`)

**Markdown formatting looks wrong**

Converted markdown has formatting issues.

Solutions:

- Use `--format html` to get raw HTML instead
- Some complex HTML structures may not convert perfectly
- Report specific issues at https://github.com/grantcarthew/snag/issues

### Platform-Specific Issues

**Linux: "No DISPLAY environment variable"**

Running in headless environment without display.

Solutions:

- Headless mode should work automatically
- Ensure Xvfb is installed: `sudo apt install xvfb`
- Use `--force-headless` explicitly

**macOS: "Chromium.app cannot be opened"**

macOS security blocking Chromium/Chrome launch.

Solutions:

- Open Chromium manually first: `open -a Chromium` or `open -a "Google Chrome"`
- Check System Preferences > Security & Privacy
- Allow the browser in privacy settings

### Getting Help

Still having issues?

1. Run with `--debug` flag for detailed logs
2. Check existing issues: https://github.com/grantcarthew/snag/issues
3. Create new issue with:
   - snag version: `snag --version`
   - Operating system and version
   - Full command you ran
   - Complete error message
   - Output from `--debug` flag

## How It Works

### Smart Browser Management

1. **Session Detection**: Auto-detects existing Chromium-based browser instance with remote debugging enabled
2. **Mode Selection**:
   - If Chromium browser is running → Connect to existing session (preserves auth/cookies)
   - If no browser found → Launch headless mode
   - If `--force-visible` → Launch visible mode for authentication
3. **Tab Management**: Keep tabs open in visible mode, close in headless mode (or use `--close-tab`)

### Output Routing

- **stdout**: Content only (HTML/Markdown) - enables piping to other tools
- **stderr**: All logs, warnings, errors, progress indicators

This design makes snag perfect for shell pipelines and AI agent integration.

### Technology

- **Language**: Go 1.21+
- **Browser Control**: Chrome DevTools Protocol via [go-rod/rod](https://github.com/go-rod/rod)
- **HTML Conversion**: [html-to-markdown/v2](https://github.com/JohannesKaufmann/html-to-markdown)
- **CLI Framework**: [urfave/cli/v2](https://github.com/urfave/cli)

## Contributing

Contributions welcome! Please:

1. Check existing issues: https://github.com/grantcarthew/snag/issues
2. Create issue for bugs or feature requests
3. Submit pull requests against `main` branch

### Reporting Issues

Include:

- snag version: `snag --version`
- Operating system and version
- Full command and error message
- Output from `--debug` flag

## License

`snag` is licensed under the [Mozilla Public License 2.0](LICENSE).

### Third-Party Licenses

This project uses the following open-source libraries:

- [go-rod/rod](https://github.com/go-rod/rod) - MIT License
- [urfave/cli](https://github.com/urfave/cli) - MIT License
- [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) - MIT License

See the [LICENSES](LICENSES/) directory for full license texts.

## Author

Grant Carthew <grant@carthew.net>
