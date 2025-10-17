# snag - Documentation Project

## Overview

This document outlines the documentation completion plan for the `snag` CLI tool. The documentation phase ensures users have comprehensive guides, proper licensing information, and troubleshooting resources.

**Status**: Phase 8 - In Progress (25%)
**Last Updated**: 2025-10-17

## About snag

`snag` is a CLI tool for intelligently fetching web page content using a browser engine, with smart session management and format conversion capabilities. Built in Go, it provides a single binary solution for retrieving web content as Markdown or HTML, with seamless authentication support through Chrome/Chromium browsers.

## Technology Stack

- **Language**: Go 1.21+
- **CLI Framework**: github.com/urfave/cli/v2
- **Browser Control**: github.com/go-rod/rod
- **HTML Conversion**: github.com/JohannesKaufmann/html-to-markdown/v2
- **License**: Mozilla Public License 2.0

## Current Documentation Status

### ✅ Completed Documentation

1. **README.md**: Core user documentation (95% complete)
   - Project description and features
   - Installation instructions (basic)
   - All 16 CLI flags documented with examples
   - Common usage examples
   - Authenticated site workflow
   - Platform support information

2. **Source Code**: All files have MPL 2.0 license headers

3. **CLI Help**: Built-in `--help` flag with comprehensive usage

4. **Internal Documentation**:
   - `docs/design.md`: Complete design decisions (923 lines)
   - `docs/notes.md`: Development notes
   - `PROJECT.md`: Implementation plan and status

### ⏳ Missing Documentation

1. **Troubleshooting Section**: No guide for common issues
2. **Third-Party Licenses**: No LICENSES/ directory
3. **Installation Guide**: Homebrew instructions pending (Phase 9)
4. **Advanced Usage**: No examples for complex scenarios
5. **Contributing Guide**: No CONTRIBUTING.md (post-MVP)

## Implementation Plan

### Task 1: Add Troubleshooting Section to README

**Description**: Create comprehensive troubleshooting guide for common issues

**Content to Add**:

```markdown
## Troubleshooting

### Browser Issues

#### "Browser not found" error

**Problem**: snag cannot locate Chrome/Chromium on your system.

**Solutions**:
- Install Chrome: https://www.google.com/chrome/
- Install Chromium: `brew install chromium` (macOS) or `sudo apt install chromium-browser` (Linux)
- Ensure Chrome/Chromium is in your system PATH

#### Browser launches but pages don't load

**Problem**: Browser starts but cannot navigate to URLs.

**Solutions**:
- Check your internet connection
- Verify the URL is correct and accessible
- Try with `--verbose` flag to see detailed logs
- Check firewall settings

#### "Failed to connect to existing browser" error

**Problem**: Attempting to connect to a running browser instance failed.

**Solutions**:
- Ensure Chrome is launched with `--remote-debugging-port=9222`
- Try a different port with `--port 9333`
- Kill existing Chrome processes and let snag launch a new instance

### Authentication Issues

#### "Authentication required" error

**Problem**: Page requires login but snag cannot authenticate.

**Solutions**:
- Use `--force-visible` to manually log in: `snag --force-visible https://example.com`
- Log in via the visible browser, then re-run snag (will use existing session)
- Check if the site uses cookies - browser session persists authentication

#### Login page detected but I'm already logged in

**Problem**: False positive on authentication detection.

**Solutions**:
- Use `--force-headless` to skip auth detection
- The page may have login forms in footer/header (false positive)
- Report the issue with the URL for investigation

### Timeout Issues

#### "Page load timeout" error

**Problem**: Page takes too long to load.

**Solutions**:
- Increase timeout: `snag --timeout 60 https://example.com`
- Use `--wait-for` to wait for specific element: `snag --wait-for ".content" https://example.com`
- Check if page has infinite loading/animations
- Try with `--verbose` to see what's happening

#### Page loads but content is missing

**Problem**: Page loads but dynamic content hasn't appeared.

**Solutions**:
- Use `--wait-for` with selector: `snag --wait-for "#main-content" https://example.com`
- Increase timeout to allow for slow loading
- Check if content requires JavaScript (some sites may not work)

### Output Issues

#### Output is empty or just has headers

**Problem**: Fetched page but content is missing.

**Solutions**:
- Try `--format html` to see raw HTML
- Use `--verbose` to check if page loaded correctly
- Page may require authentication (see auth issues)
- Content may be loaded dynamically (use `--wait-for`)

#### Markdown formatting looks wrong

**Problem**: Converted Markdown has formatting issues.

**Solutions**:
- Use `--format html` to get raw HTML instead
- Some complex HTML structures may not convert perfectly
- Report specific formatting issues for investigation

#### "Permission denied" when saving to file

**Problem**: Cannot write output file.

**Solutions**:
- Check file path is writable: `ls -la /path/to/output/`
- Ensure directory exists: `mkdir -p /path/to/output/`
- Check disk space: `df -h`
- Verify file permissions

### Performance Issues

#### snag is very slow

**Problem**: Takes a long time to fetch pages.

**Solutions**:
- First run launches browser (slower), subsequent runs reuse it
- Use existing Chrome session: launch Chrome with `--remote-debugging-port=9222`
- Slow websites naturally take longer
- Use `--verbose` to identify bottlenecks

#### High memory usage

**Problem**: Browser consuming too much memory.

**Solutions**:
- Close browser between runs: don't use existing session
- Close unused tabs in existing browser sessions
- Browser memory usage is normal (Chrome is memory-intensive)
- Restart snag with fresh browser instance

### Platform-Specific Issues

#### macOS: "Chrome.app cannot be opened" error

**Problem**: macOS security blocking Chrome launch.

**Solutions**:
- Open Chrome manually first: `open -a "Google Chrome"`
- Check System Preferences > Security & Privacy
- Allow Chrome in privacy settings

#### Linux: "No DISPLAY environment variable" error

**Problem**: Running in headless environment without display.

**Solutions**:
- Headless mode should work automatically
- Ensure Xvfb is installed for headless: `sudo apt install xvfb`
- Use `--force-headless` explicitly

### Getting Help

If you're still having issues:

1. Run with `--debug` flag to get detailed logs
2. Check existing issues: https://github.com/grantcarthew/snag/issues
3. Create new issue with:
   - snag version: `snag --version`
   - Operating system and version
   - Full command you ran
   - Complete error message
   - Output from `--debug` flag

## Common Questions

**Q: Does snag work with JavaScript-heavy sites?**
A: Yes, snag uses a real browser engine (Chrome) so JavaScript executes normally.

**Q: Can I use snag for web scraping?**
A: Yes, but be respectful of websites' terms of service and robots.txt.

**Q: Does snag support proxies?**
A: Not yet. This is a post-MVP feature.

**Q: Can I run snag in CI/CD pipelines?**
A: Yes, but you'll need Chrome/Chromium installed in your CI environment.

**Q: Does snag support Windows?**
A: Not currently tested. macOS and Linux are officially supported.
```

**Deliverables**:
- Updated README.md with troubleshooting section
- Covers all major error scenarios
- Provides actionable solutions
- Includes common questions

### Task 2: Create LICENSES Directory Structure

**Description**: Set up proper third-party license attribution

**Directory Structure**:
```
snag/
├── LICENSE              # MPL 2.0 (already exists)
└── LICENSES/
    ├── README.md        # Overview of dependencies and licenses
    ├── rod.LICENSE      # MIT License
    ├── urfave-cli.LICENSE  # MIT License
    └── html-to-markdown.LICENSE  # MIT License
```

**Subtasks**:
1. Create `LICENSES/` directory
2. Create `LICENSES/README.md` explaining the licenses
3. Download license files from dependencies
4. Verify license compatibility with MPL 2.0

**Deliverables**:
- `LICENSES/` directory with all third-party licenses
- Clear attribution and compliance

### Task 3: Create LICENSES/README.md

**Description**: Document all third-party dependencies and their licenses

**Content**:

```markdown
# Third-Party Licenses

`snag` is licensed under the Mozilla Public License 2.0 (see ../LICENSE).

This directory contains the licenses for third-party dependencies used in snag.

## Dependencies

### go-rod/rod (MIT License)
- **Purpose**: Chrome DevTools Protocol library for browser automation
- **Repository**: https://github.com/go-rod/rod
- **License**: MIT License
- **License File**: rod.LICENSE

### urfave/cli (MIT License)
- **Purpose**: CLI framework for building command-line applications
- **Repository**: https://github.com/urfave/cli
- **License**: MIT License
- **License File**: urfave-cli.LICENSE

### JohannesKaufmann/html-to-markdown (MIT License)
- **Purpose**: HTML to Markdown conversion library
- **Repository**: https://github.com/JohannesKaufmann/html-to-markdown
- **License**: MIT License
- **License File**: html-to-markdown.LICENSE

## License Compatibility

All dependencies use the MIT License, which is compatible with MPL 2.0:
- MIT is a permissive license allowing commercial use, modification, and distribution
- MIT-licensed code can be included in MPL 2.0 projects
- Proper attribution is maintained in this directory

## Generating License Files

To verify or update these licenses:

```bash
# Clone repositories and copy LICENSE files
gh repo clone go-rod/rod && cp rod/LICENSE LICENSES/rod.LICENSE
gh repo clone urfave/cli && cp cli/LICENSE LICENSES/urfave-cli.LICENSE
gh repo clone JohannesKaufmann/html-to-markdown && cp html-to-markdown/LICENSE LICENSES/html-to-markdown.LICENSE
```

## Acknowledgments

We thank the maintainers and contributors of these excellent open-source projects.
```

**Deliverables**:
- Clear documentation of all dependencies
- License compatibility statement
- Instructions for verification

### Task 4: Download Third-Party License Files

**Description**: Obtain license files from dependency repositories

**Commands to Run**:

```bash
# Create LICENSES directory
mkdir -p LICENSES

# go-rod/rod (MIT)
curl -L https://raw.githubusercontent.com/go-rod/rod/main/LICENSE -o LICENSES/rod.LICENSE

# urfave/cli (MIT)
curl -L https://raw.githubusercontent.com/urfave/cli/main/LICENSE -o LICENSES/urfave-cli.LICENSE

# JohannesKaufmann/html-to-markdown (MIT)
curl -L https://raw.githubusercontent.com/JohannesKaufmann/html-to-markdown/main/LICENSE -o LICENSES/html-to-markdown.LICENSE
```

**Verification**:
- Ensure each file contains proper MIT License text
- Verify copyright holders are correct
- Check file sizes are reasonable (MIT License ~1-2KB)

**Deliverables**:
- Three license files in LICENSES/ directory
- All files properly formatted and complete

### Task 5: Update README with License Information

**Description**: Add section about licensing to README

**Content to Add**:

```markdown
## License

`snag` is licensed under the [Mozilla Public License 2.0](LICENSE).

### Third-Party Licenses

This project uses the following open-source libraries:

- [go-rod/rod](https://github.com/go-rod/rod) - MIT License
- [urfave/cli](https://github.com/urfave/cli) - MIT License
- [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) - MIT License

See the [LICENSES](LICENSES/) directory for full license texts.
```

**Placement**: Add near end of README, before "Contributing" or "Support" sections

**Deliverables**:
- Updated README.md with license section
- Links to license files and repositories

### Task 6: Add Advanced Usage Examples

**Description**: Document complex usage scenarios

**Examples to Add**:

```markdown
## Advanced Usage

### Using an Existing Browser Session

Keep authentication across multiple snag calls:

```bash
# Start Chrome with remote debugging (in separate terminal)
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome \
  --remote-debugging-port=9222 \
  --user-data-dir=/tmp/chrome-profile

# In another terminal, use snag with existing session
snag https://authenticated-site.com
snag https://another-authenticated-site.com
snag https://third-site.com
```

All three commands share the same browser session and authentication state.

### Batch Processing URLs

Process multiple URLs with a shell loop:

```bash
# From file
while read url; do
  filename=$(echo "$url" | sed 's/[^a-zA-Z0-9]/_/g').md
  snag --quiet -o "$filename" "$url"
done < urls.txt

# From command line
for url in https://example.com https://example.org; do
  snag --quiet "$url" >> combined.md
  echo -e "\n---\n" >> combined.md
done
```

### Custom User Agent for API Documentation

Some sites serve different content based on user agent:

```bash
# Mobile user agent
snag --user-agent "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)" \
  https://example.com

# Bot user agent
snag --user-agent "MyBot/1.0 (+https://example.com/bot)" \
  https://api-docs.example.com

# Specific browser
snag --user-agent "Mozilla/5.0 (X11; Linux x86_64) Firefox/95.0" \
  https://example.com
```

### Waiting for Dynamic Content

Sites that load content via JavaScript:

```bash
# Wait for specific element to appear
snag --wait-for ".article-content" https://dynamic-site.com

# Combine with increased timeout for slow sites
snag --timeout 60 --wait-for "#loaded-indicator" https://slow-site.com
```

### Debugging Failed Fetches

Troubleshoot issues with verbose logging:

```bash
# See what's happening during fetch
snag --verbose https://problematic-site.com

# Full debug output including browser messages
snag --debug https://problematic-site.com 2> debug.log

# Open browser visibly to see what snag sees
snag --force-visible https://problematic-site.com
```

### Headless in CI/CD

Run in continuous integration environments:

```bash
# Explicit headless mode (default, but explicit for clarity)
snag --force-headless --timeout 30 https://example.com

# Combine with quiet mode for clean CI logs
snag --quiet --force-headless https://example.com > output.md
```

### Converting Specific Page Sections

Use browser DevTools to find selectors, then wait for them:

```bash
# Find selector in Chrome DevTools (Inspect Element)
# Then use it with snag
snag --wait-for "article.main-content" https://news-site.com
```
```

**Deliverables**:
- README.md with advanced usage section
- Real-world examples users can copy-paste
- Covers complex scenarios

### Task 7: Enhance Installation Section

**Description**: Improve installation instructions in README

**Current State**: Basic "download binary" instructions

**Improvements Needed**:

```markdown
## Installation

### Homebrew (macOS/Linux) - Coming Soon

```bash
# Will be available after v1.0.0 release
brew install grantcarthew/tap/snag
```

### Download Binary (Current Method)

Download the latest release for your platform:

**macOS**:
```bash
# Apple Silicon (M1/M2/M3)
curl -L https://github.com/grantcarthew/snag/releases/latest/download/snag-darwin-arm64 -o snag
chmod +x snag
sudo mv snag /usr/local/bin/

# Intel
curl -L https://github.com/grantcarthew/snag/releases/latest/download/snag-darwin-amd64 -o snag
chmod +x snag
sudo mv snag /usr/local/bin/
```

**Linux**:
```bash
# 64-bit
curl -L https://github.com/grantcarthew/snag/releases/latest/download/snag-linux-amd64 -o snag
chmod +x snag
sudo mv snag /usr/local/bin/

# ARM64
curl -L https://github.com/grantcarthew/snag/releases/latest/download/snag-linux-arm64 -o snag
chmod +x snag
sudo mv snag /usr/local/bin/
```

### Build from Source

Requires Go 1.21 or later:

```bash
git clone https://github.com/grantcarthew/snag.git
cd snag
go build -o snag
sudo mv snag /usr/local/bin/
```

### Prerequisites

snag requires Chrome or Chromium to be installed:

**macOS**:
```bash
# Chrome (recommended)
# Download from https://www.google.com/chrome/

# Or Chromium via Homebrew
brew install chromium
```

**Linux**:
```bash
# Ubuntu/Debian
sudo apt install chromium-browser

# Fedora
sudo dnf install chromium

# Arch Linux
sudo pacman -S chromium
```

### Verify Installation

```bash
snag --version
```
```

**Deliverables**:
- Comprehensive installation instructions
- Platform-specific commands
- Prerequisites clearly documented
- Build from source option included

### Task 8: Review and Polish All Documentation

**Description**: Final review pass on all documentation

**Checklist**:

- [ ] README.md is complete and accurate
- [ ] All CLI flags documented with examples
- [ ] Troubleshooting section is comprehensive
- [ ] Installation instructions work on all platforms
- [ ] License information is correct
- [ ] Links are not broken
- [ ] Examples are tested and working
- [ ] Spelling and grammar checked
- [ ] Markdown formatting is correct (CommonMark)
- [ ] Code blocks have proper syntax highlighting
- [ ] Section ordering is logical

**Tools to Use**:
```bash
# Check markdown formatting
markdownlint README.md

# Check for broken links (if tool available)
markdown-link-check README.md

# Spell check
aspell check README.md
```

**Deliverables**:
- Polished, professional README.md
- All documentation error-free
- Ready for v1.0.0 release

## Documentation Standards

### Markdown Style

- Use CommonMark specification
- ATX-style headings (`#` prefix)
- Fenced code blocks with language tags
- Consistent list formatting (use `-` for bullets)
- Blank lines between sections

### Code Examples

- Always include language tag: ` ```bash `
- Show expected output when relevant
- Use realistic, working examples
- Include error cases and solutions

### Voice and Tone

- Clear and concise
- Friendly but professional
- Action-oriented (start with verbs)
- Avoid jargon where possible
- Explain technical terms when used

## Success Criteria

- ✅ README.md has comprehensive troubleshooting section
- ✅ LICENSES/ directory created with all third-party licenses
- ✅ License compatibility documented
- ✅ Advanced usage examples added
- ✅ Installation instructions complete for all platforms
- ✅ All documentation reviewed and polished
- ✅ No broken links or formatting errors
- ✅ Ready for public release

## Timeline Estimate

- Task 1 (Troubleshooting): 2-3 hours
- Task 2-5 (Licenses): 1-2 hours
- Task 6 (Advanced examples): 1-2 hours
- Task 7 (Installation): 1 hour
- Task 8 (Review): 1 hour

**Total**: 6-9 hours of work

## Next Steps After Documentation

Once documentation is complete (Phase 8), proceed to:
- **Phase 9**: Distribution & Release - See PROJECT-release.md
- Multi-platform builds, Homebrew formula, v1.0.0 release

## Related Documents

- `PROJECT.md`: Main project implementation plan
- `PROJECT-testing.md`: Testing strategy and plan
- `PROJECT-release.md`: Release and distribution plan
- `README.md`: User-facing documentation
- `docs/design.md`: Design decisions and rationale
