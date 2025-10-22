# PROJECT: Multiple URL Support

## Project Overview

Add support for fetching multiple URLs in a single snag invocation, enabling batch content retrieval for AI agents and automation workflows.

**Feature Type:** Enhancement
**Status:** Planning (Design Complete)
**Priority:** Medium
**Estimated Complexity:** Medium

## Current State

snag currently supports fetching content from:

- Single URL argument: `snag https://example.com` → stdout
- Single tab selection: `snag --tab 1` or `snag -t "pattern"`
- Tab listing: `snag --list-tabs`
- Batch tab fetching: `snag --all-tabs -d ./results`

**Limitations:**

- Only one URL can be provided per invocation (excluding tab features)
- Batch URL processing requires shell loops or scripts
- No built-in support for URL lists or files

## Goals

### Primary Objectives

1. **Multiple URL Arguments**: Accept multiple URLs via command line

   ```bash
   snag https://example.com https://go.dev https://github.com
   ```

2. **URL File Support**: Load URLs from a text file

   ```bash
   snag --url-file urls.txt
   ```

3. **Flexible Input**: Allow combining file and inline URLs
   ```bash
   snag --url-file common-urls.txt https://new-site.com
   ```

### Success Criteria

- Fetch and save content from multiple URLs sequentially
- Support mixing URLs with format/timeout/wait-for flags
- Handle errors gracefully with continue-on-error behavior
- Maintain backward compatibility with single URL usage
- Clear progress indicators and summary output

## Design Decisions (CONFIRMED)

### 1. CLI Interface

**Multiple URL Sources:**

```bash
snag url1 url2 url3                           # Inline URLs
snag --url-file urls.txt                      # From file
snag --url-file urls.txt url4 url5            # Combined (allowed)
```

**Validation Rules:**

- Zero URLs provided → Error: "No URLs provided. Use --url-file or provide URLs as arguments"
- One URL → Works normally (backward compatible)
- Two or more URLs → Batch processing mode

**Flag Conflicts:**

```bash
# These combinations ERROR:
snag --tab 1 url1                             # ERROR: Cannot mix --tab with URLs
snag --all-tabs url1                          # ERROR: Cannot mix --all-tabs with URLs
snag --list-tabs url1                         # ERROR: --list-tabs is standalone
snag url1 url2 -o output.md                   # ERROR: Use --output-dir for multiple URLs
snag url1 url2 --close-tab                    # ERROR: --close-tab not supported with multiple URLs

# These work:
snag url1 url2 -d ./results                   # ✓ Save to directory
snag url1 url2 --format html                  # ✓ Format applies to all
snag url1 url2 --timeout 60                   # ✓ Timeout applies to each
snag url1 url2 --wait-for ".content"          # ✓ Wait applies to each page
```

### 2. URL File Format

**File Specification:**

```
# Full-line comments (# or //)
// Both comment styles supported

# URLs can be provided with or without protocol
example.com                    # Auto-prepends https://
https://go.dev/doc/            # Uses as-is
http://localhost:8080          # Preserves http://

  github.com/user/repo         # Whitespace trimmed

                               # Blank lines ignored

https://example.com # Inline comment with #
https://go.dev // Inline comment with //
```

**Processing Rules (in order):**

1. `TrimSpace()` on each line
2. If empty → skip
3. If starts with `#` or `//` → skip (full-line comment)
4. Check for inline comments:
   - Look for `" #"` or `" //"` (space + marker)
   - If found: substring before marker, TrimSpace again
   - If space exists but NO comment marker → **log warning, skip line**
5. If missing `http://` or `https://` → prepend `"https://"`
6. Validate URL format
7. If invalid → **log warning to stderr, skip line**
8. If valid → add to URL list

**Example Error Handling:**

```
# urls.txt content:
https://example.com # Production     → ✓ https://example.com
github.com // Main repo              → ✓ https://github.com
https://go.dev extra text            → ⚠️ Skip - space without comment
https://not valid url                → ⚠️ Skip - invalid URL after https:// prepend
example.com                          → ✓ https://example.com
```

### 3. Output Behavior

**BREAKING CHANGE: --open-browser**

Current behavior (will change):

```bash
snag --open-browser https://example.com    # Opens browser AND outputs content
```

New behavior:

```bash
snag --open-browser https://example.com    # ONLY opens in browser, NO output
snag --open-browser url1 url2 url3         # Opens all in tabs, NO output
```

**Rationale:** --open-browser is for persistent browser sessions. To fetch content from opened tabs, use `snag --tab <index>`.

**Multiple URLs (2+) Always Save to Disk:**

```bash
# Headless mode (default)
snag url1 url2 url3                        # Saves to ./{auto-generated-names}
snag url1 url2 -d ./results                # Saves to ./results/{auto-generated-names}

# Open browser mode (just opens tabs)
snag --open-browser url1 url2 url3         # Opens tabs, NO output
```

**Single URL (Backward Compatible):**

```bash
snag url1                                  # Outputs to stdout
snag url1 -o output.md                     # Saves to output.md
snag url1 -d ./results                     # Saves to ./results/{auto-generated}
snag --open-browser url1                   # Opens in browser, NO output (BREAKING)
```

**Filename Generation:**

- Already implemented in `output.go`
- Format: `yyyy-mm-dd-hhmmss-{title-slug}.{ext}`
- Conflict resolution: Appends `-1`, `-2`, etc.
- Extension: Auto-detected from format (`.md`, `.html`, `.txt`, `.pdf`, `.png`)

### 4. Error Handling

**Continue-on-Error Strategy:**

```go
// Pseudocode
successCount := 0
failureCount := 0

for each url in urls {
    if err := fetchAndSave(url); err != nil {
        logger.Error("[%d/%d] Failed: %s - %v", current, total, url, err)
        failureCount++
        continue  // Continue processing remaining URLs
    }
    logger.Success("[%d/%d] Saved: %s", current, total, url)
    successCount++
}

// Summary
logger.Success("Batch complete: %d succeeded, %d failed", successCount, failureCount)

// Exit code
if failureCount > 0 {
    return exit code 1
}
return exit code 0
```

**Exit Codes:**

- Exit 0: ALL URLs succeeded
- Exit 1: ANY URL failed

**Error Types to Handle:**

- Invalid URL in file → Warning to stderr, skip line, continue
- Network/timeout error → Error to stderr, skip URL, continue
- File write error → Error to stderr, skip URL, continue
- Browser connection error → Fatal error, exit immediately

### 5. Progress Indicators

**Format (Option A - Minimal):**

```
[1/5] Fetching example.com...
✓ Saved: 2025-10-22-124752-example-domain.md

[2/5] Fetching go.dev...
✓ Saved: 2025-10-22-124753-go-programming-language.md

[3/5] Fetching invalid-url.com...
✗ Failed: timeout exceeded

[4/5] Fetching github.com...
✓ Saved: 2025-10-22-124755-github.md

Batch complete: 3 succeeded, 1 failed
```

**Logging Levels:**

- Normal: Progress indicators + summary
- Quiet (`-q`): Only errors + summary
- Verbose (`-v`): Detailed timing and URLs
- Debug (`--debug`): CDP messages

### 6. Flag Behavior with Multiple URLs

| Flag               | Single URL            | Multiple URLs          | Notes                      |
| ------------------ | --------------------- | ---------------------- | -------------------------- |
| `-o, --output`     | ✓ Saves to file       | ✗ Error                | Use `--output-dir` instead |
| `-d, --output-dir` | ✓ Auto-generates name | ✓ Auto-generates names | All files saved here       |
| `--format`         | ✓ Applies             | ✓ Applies to all       | All URLs use same format   |
| `--timeout`        | ✓ Applies             | ✓ Applies to each      | Per-URL timeout            |
| `--wait-for`       | ✓ Applies             | ✓ Applies to each      | Waits on each page         |
| `--close-tab`      | ✓ Closes after fetch  | ✗ Error                | Ambiguous for multiple     |
| `--open-browser`   | ✓ Opens, no output    | ✓ Opens all, no output | BREAKING CHANGE            |
| `--tab`            | ✓ Fetches from tab    | ✗ Error                | Conflicts with URLs        |
| `--all-tabs`       | ✓ Fetches all tabs    | ✗ Error                | Conflicts with URLs        |
| `--list-tabs`      | ✓ Lists tabs          | ✗ Error                | Standalone only            |

## Implementation Plan

### Phase 1: Core Multiple URL Support (4-5 hours)

**Files to Modify:**

- `main.go`: Update argument parsing and validation
- `handlers.go`: Add batch URL handler function
- `validate.go`: Add URL file loader and validators

**New Components:**

```go
// validate.go
// loadURLsFromFile reads and parses a URL file
// Returns list of valid URLs (invalid lines are logged and skipped)
func loadURLsFromFile(filename string) ([]string, error)

// validateURLList validates and normalizes a list of URLs
// Prepends https:// if missing, logs warnings for invalid URLs
func validateURLList(urls []string) ([]string, error)

// handlers.go
// handleMultipleURLs orchestrates batch URL fetching
func handleMultipleURLs(urls []string, config *Config) error

// Reuse existing functions:
// - generateOutputFilename() (already exists)
// - processPageContent() (already exists)
```

**Validation Logic:**

```go
// In main.go action handler
func(c *cli.Context) error {
    var urls []string

    // Load from file if provided
    if urlFile := c.String("url-file"); urlFile != "" {
        fileURLs, err := loadURLsFromFile(urlFile)
        if err != nil {
            return err
        }
        urls = append(urls, fileURLs...)
    }

    // Add command-line URLs
    urls = append(urls, c.Args().Slice()...)

    // Validate URL count
    if len(urls) == 0 {
        return errors.New("no URLs provided")
    }

    // Check for conflicting flags
    if len(urls) > 1 {
        // Error if -o specified
        if c.String("output") != "" {
            return errors.New("--output cannot be used with multiple URLs. Use --output-dir instead")
        }
        // Error if --close-tab specified
        if c.Bool("close-tab") {
            return errors.New("--close-tab not supported with multiple URLs")
        }
    }

    // Check tab conflicts
    if c.IsSet("tab") || c.Bool("all-tabs") || c.Bool("list-tabs") {
        if len(urls) > 0 {
            return errors.New("cannot mix tab features with URL arguments")
        }
    }

    // Route to appropriate handler
    if len(urls) == 1 {
        return handleSingleURL(urls[0], config)  // Existing behavior
    } else {
        return handleMultipleURLs(urls, config)  // New batch handler
    }
}
```

### Phase 2: File Loading and Validation (2-3 hours)

**URL File Loader:**

```go
// validate.go
func loadURLsFromFile(filename string) ([]string, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to open URL file: %w", err)
    }
    defer file.Close()

    var urls []string
    scanner := bufio.NewScanner(file)
    lineNum := 0

    for scanner.Scan() {
        lineNum++
        line := strings.TrimSpace(scanner.Text())

        // Skip empty lines
        if line == "" {
            continue
        }

        // Skip full-line comments
        if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
            continue
        }

        // Handle inline comments
        hasComment := false
        for _, marker := range []string{" #", " //"} {
            if idx := strings.Index(line, marker); idx != -1 {
                line = strings.TrimSpace(line[:idx])
                hasComment = true
                break
            }
        }

        // Check for space without comment (formatting error)
        if !hasComment && strings.Contains(line, " ") {
            logger.Warning("Line %d: URL contains space without comment marker - skipping: %s", lineNum, line)
            continue
        }

        // Auto-prepend https:// if missing protocol
        if !strings.HasPrefix(line, "http://") && !strings.HasPrefix(line, "https://") {
            line = "https://" + line
        }

        // Validate URL
        if err := validateURL(line); err != nil {
            logger.Warning("Line %d: Invalid URL - skipping: %s", lineNum, err)
            continue
        }

        urls = append(urls, line)
    }

    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("error reading file: %w", err)
    }

    if len(urls) == 0 {
        return nil, errors.New("no valid URLs found in file")
    }

    return urls, nil
}
```

### Phase 3: Batch URL Handler (3-4 hours)

**Batch Processing Function:**

```go
// handlers.go
func handleMultipleURLs(urls []string, config *Config) error {
    // Determine output directory
    outputDir := config.OutputDir
    if outputDir == "" {
        outputDir = "." // Default to current directory
    }

    // Validate output directory
    if err := validateDirectory(outputDir); err != nil {
        return err
    }

    // Create browser manager
    bm := NewBrowserManager(BrowserOptions{
        Port:          config.Port,
        ForceHeadless: config.ForceHeadless,
        ForceVisible:  config.ForceVisible,
        OpenBrowser:   config.OpenBrowser,
        UserAgent:     config.UserAgent,
    })

    browserManager = bm

    // Connect to browser
    _, err := bm.Connect()
    if err != nil {
        browserManager = nil
        return err
    }

    defer func() {
        bm.Close()
        browserManager = nil
    }()

    // If --open-browser, just open tabs and return
    if config.OpenBrowser {
        return openURLsInBrowser(bm, urls)
    }

    // Otherwise, fetch and save each URL
    return fetchAndSaveURLs(bm, urls, config, outputDir)
}

func openURLsInBrowser(bm *BrowserManager, urls []string) error {
    logger.Info("Opening %d URLs in browser...", len(urls))

    for i, url := range urls {
        logger.Info("[%d/%d] Opening: %s", i+1, len(urls), url)

        page, err := bm.NewPage()
        if err != nil {
            logger.Error("[%d/%d] Failed to create page: %v", i+1, len(urls), err)
            continue
        }

        _, err = page.Navigate(url)
        if err != nil {
            logger.Error("[%d/%d] Failed to navigate: %v", i+1, len(urls), err)
            continue
        }

        logger.Success("[%d/%d] Opened: %s", i+1, len(urls), url)
    }

    logger.Info("Browser will remain open. Press Ctrl+C to exit.")
    return nil
}

func fetchAndSaveURLs(bm *BrowserManager, urls []string, config *Config, outputDir string) error {
    successCount := 0
    failureCount := 0
    timestamp := time.Now() // Single timestamp for batch

    logger.Info("Processing %d URLs...", len(urls))

    for i, url := range urls {
        current := i + 1
        total := len(urls)

        logger.Info("[%d/%d] Fetching: %s", current, total, url)

        // Create new page for this URL
        page, err := bm.NewPage()
        if err != nil {
            logger.Error("[%d/%d] Failed to create page: %v", current, total, err)
            failureCount++
            continue
        }

        // Fetch page content
        fetcher := NewPageFetcher(page, config.Timeout)
        _, err = fetcher.Fetch(FetchOptions{
            URL:     url,
            Timeout: config.Timeout,
            WaitFor: config.WaitFor,
        })
        if err != nil {
            logger.Error("[%d/%d] Failed to fetch: %v", current, total, err)
            bm.ClosePage(page) // Clean up failed page
            failureCount++
            continue
        }

        // Get page info for filename generation
        info, err := page.Info()
        if err != nil {
            logger.Error("[%d/%d] Failed to get page info: %v", current, total, err)
            bm.ClosePage(page)
            failureCount++
            continue
        }

        // Generate output filename
        outputPath, err := generateOutputFilename(
            info.Title, url, config.Format,
            timestamp, outputDir,
        )
        if err != nil {
            logger.Error("[%d/%d] Failed to generate filename: %v", current, total, err)
            bm.ClosePage(page)
            failureCount++
            continue
        }

        // Process and save content
        if err := processPageContent(page, config.Format, outputPath); err != nil {
            logger.Error("[%d/%d] Failed to save content: %v", current, total, err)
            bm.ClosePage(page)
            failureCount++
            continue
        }

        // Success
        filename := filepath.Base(outputPath)
        logger.Success("[%d/%d] Saved: %s", current, total, filename)

        // Close page in headless mode
        bm.ClosePage(page)
        successCount++
    }

    // Summary
    logger.Success("Batch complete: %d succeeded, %d failed", successCount, failureCount)

    if failureCount > 0 {
        return fmt.Errorf("batch processing completed with %d failures", failureCount)
    }

    return nil
}
```

### Phase 4: CLI Integration (1-2 hours)

**Update main.go flags:**

```go
// Add new flag
&cli.StringFlag{
    Name:    "url-file",
    Aliases: []string{"f"},
    Usage:   "Read URLs from `FILE` (one per line, supports comments)",
},
```

**Update argument handling:**

- Change from `c.Args().First()` to `c.Args().Slice()`
- Combine file URLs + inline URLs
- Route to appropriate handler based on count

### Phase 5: Testing (3-4 hours)

**Test Coverage in `cli_test.go`:**

```go
// Basic multiple URL tests
func TestCLI_MultipleURLs_Inline(t *testing.T)
func TestCLI_MultipleURLs_File(t *testing.T)
func TestCLI_MultipleURLs_Combined(t *testing.T)

// File format tests
func TestCLI_URLFile_Comments(t *testing.T)
func TestCLI_URLFile_BlankLines(t *testing.T)
func TestCLI_URLFile_InlineComments(t *testing.T)
func TestCLI_URLFile_InvalidURLs(t *testing.T)
func TestCLI_URLFile_AutoHTTPS(t *testing.T)
func TestCLI_URLFile_SpaceWithoutComment(t *testing.T)

// Output tests
func TestCLI_MultipleURLs_OutputDir(t *testing.T)
func TestCLI_MultipleURLs_CurrentDir(t *testing.T)
func TestCLI_MultipleURLs_Format(t *testing.T)

// Error handling tests
func TestCLI_MultipleURLs_PartialFailure(t *testing.T)
func TestCLI_MultipleURLs_AllFail(t *testing.T)

// Conflict tests
func TestCLI_MultipleURLs_WithOutputFlag_Error(t *testing.T)
func TestCLI_MultipleURLs_WithCloseTab_Error(t *testing.T)
func TestCLI_MultipleURLs_WithTab_Error(t *testing.T)
func TestCLI_URLFile_WithListTabs_Error(t *testing.T)

// Edge cases
func TestCLI_URLFile_Empty(t *testing.T)
func TestCLI_URLFile_AllInvalid(t *testing.T)
func TestCLI_ZeroURLs_Error(t *testing.T)

// Open browser tests
func TestCLI_OpenBrowser_MultipleURLs(t *testing.T)
func TestCLI_OpenBrowser_SingleURL_NoOutput(t *testing.T) // BREAKING CHANGE test
```

**Manual Testing Script:**

```bash
#!/bin/bash
# scripts/test-multi-url.sh

set -e

echo "=== Test 1: Multiple inline URLs ==="
./snag https://example.com https://go.dev --force-headless

echo "=== Test 2: URL file basic ==="
cat > /tmp/urls.txt <<EOF
# Test URLs
example.com
https://go.dev/doc/
github.com/grantcarthew/snag
EOF

./snag --url-file /tmp/urls.txt --force-headless

echo "=== Test 3: URL file with inline comments ==="
cat > /tmp/urls-comments.txt <<EOF
example.com # Production site
go.dev // Documentation
github.com  Main repo
EOF

./snag -f /tmp/urls-comments.txt --force-headless

echo "=== Test 4: Combined file + inline ==="
./snag -f /tmp/urls.txt https://anthropic.com --force-headless

echo "=== Test 5: Output directory ==="
mkdir -p /tmp/snag-test
./snag -f /tmp/urls.txt -d /tmp/snag-test --force-headless

echo "=== Test 6: Invalid URLs (should skip) ==="
cat > /tmp/urls-invalid.txt <<EOF
example.com
invalid url with spaces
https://go.dev
not-a-valid-url
EOF

./snag -f /tmp/urls-invalid.txt --force-headless

echo "=== Test 7: Error - multiple URLs with -o ==="
./snag example.com go.dev -o output.md --force-headless 2>&1 || echo "Expected error ✓"

echo "=== Test 8: Error - multiple URLs with --close-tab ==="
./snag example.com go.dev --close-tab --force-headless 2>&1 || echo "Expected error ✓"

echo "=== Test 9: Open browser with multiple URLs ==="
./snag --open-browser example.com go.dev

echo "All tests complete!"
```

## New Sentinel Errors

```go
// errors.go
var (
    ErrURLFileNotFound       = errors.New("URL file not found")
    ErrNoValidURLs           = errors.New("no valid URLs provided")
    ErrInvalidURLFile        = errors.New("invalid URL file format")
    ErrOutputNotAllowed      = errors.New("--output cannot be used with multiple URLs")
    ErrCloseTabNotAllowed    = errors.New("--close-tab not supported with multiple URLs")
    ErrTabConflict           = errors.New("cannot mix tab features with URL arguments")
)
```

## Breaking Changes

### --open-browser Behavior Change

**Before:**

```bash
snag --open-browser https://example.com
# Opens browser AND outputs content to stdout
```

**After:**

```bash
snag --open-browser https://example.com
# ONLY opens browser, NO content output
# To fetch content: snag --tab 1
```

**Migration Guide:**

```bash
# Old workflow:
snag --open-browser https://example.com > output.md

# New workflow:
snag --open-browser https://example.com   # Opens in browser
snag --tab 1 -o output.md                 # Then fetch content
```

**Rationale:**

- `--open-browser` is for persistent sessions and manual interaction
- Separating "open" from "fetch" provides clearer semantics
- Aligns with multiple URL behavior (open all tabs without fetching)

## Documentation Updates

### README.md Updates

**Add Multiple URL Examples:**

````markdown
## Batch Processing

Fetch multiple URLs in one command:

```bash
# Multiple URLs (saves to current directory with auto-generated names)
snag https://example.com https://go.dev https://github.com

# From a file
snag --url-file urls.txt

# Combine file and inline URLs
snag --url-file common.txt https://new-site.com

# Save to specific directory
snag --url-file urls.txt --output-dir ./results

# Open multiple URLs in browser (no output)
snag --open-browser url1 url2 url3
```
````

**URL File Format:**

```
# Comments start with # or //
example.com                    # Auto-adds https://
https://go.dev/doc/            # Uses as-is
http://localhost:8080          # Preserves http://

// Inline comments supported
github.com // Repository

# Blank lines ignored
```

````

### AGENTS.md Updates

**Update "Build and Test Commands" section:**

```markdown
# Test multiple URL features
snag example.com go.dev github.com           # Multiple inline URLs
snag --url-file urls.txt                     # From file
snag -f urls.txt -d ./results                # With output dir

# Error cases
snag url1 url2 -o out.md                     # ERROR: Use --output-dir
snag url1 url2 --close-tab                   # ERROR: Not supported
snag --tab 1 url1                            # ERROR: Cannot mix
````

**Add to Troubleshooting:**

```markdown
- URL file errors: Check format (one URL per line, comments with # or //)
- Multiple URLs fail: Check --output-dir is writable directory
- "Space without comment" warning: URL contains space, add # or // for comment
```

### Design Docs

**Add to `docs/design-record.md`:**

```markdown
## Multiple URL Support (2025-10-22)

**Decision:** Support multiple URLs via inline arguments and --url-file

**Rationale:**

- AI agents benefit from batch content fetching
- URL files enable reproducible workflows
- Combines well with existing tab management

**Implementation:**

- Continue-on-error: Process all URLs even if some fail
- Exit 1 if ANY URL fails (standard batch operation behavior)
- Auto-save to disk (never stdout for multiple URLs)
- Progress indicators: [N/M] format with success/failure markers

**Breaking Change:** --open-browser no longer outputs content

- Separates "open persistent browser" from "fetch content"
- Use --tab to fetch from opened tabs
- Aligns with multiple URL behavior
```

## Timeline Estimate

- **Phase 1 (Core)**: 4-5 hours

  - Argument parsing: 1 hour
  - Validation logic: 2 hours
  - Integration: 1-2 hours

- **Phase 2 (File Loading)**: 2-3 hours

  - File parser: 1-2 hours
  - Comment handling: 1 hour

- **Phase 3 (Batch Handler)**: 3-4 hours

  - Batch orchestration: 2 hours
  - Error handling: 1 hour
  - Progress indicators: 1 hour

- **Phase 4 (CLI Integration)**: 1-2 hours

  - Flag updates: 0.5 hours
  - Routing logic: 0.5 hours
  - --open-browser changes: 1 hour

- **Phase 5 (Testing)**: 3-4 hours
  - Unit tests: 2 hours
  - Integration tests: 1-2 hours

**Total Estimate**: 13-18 hours

## Dependencies

**No new external dependencies required.**

Uses existing:

- `github.com/urfave/cli/v2` - CLI framework
- `github.com/go-rod/rod` - Browser control
- Standard library: `bufio`, `os`, `path/filepath`, `strings`

## Risks and Mitigations

### Risk 1: Breaking Change Impact

**Risk:** --open-browser behavior change breaks existing workflows
**Mitigation:**

- Document in CHANGELOG and README prominently
- Migration guide provided
- Behavior is more logical and consistent

### Risk 2: File Parsing Edge Cases

**Risk:** Complex URL formats may not parse correctly
**Mitigation:**

- Comprehensive test suite for edge cases
- Clear error messages with line numbers
- Continue-on-error prevents total failure

### Risk 3: Memory Usage with Large Files

**Risk:** Loading thousands of URLs into memory
**Mitigation:**

- Stream processing (scanner, not ReadAll)
- Sequential fetching (not parallel)
- Browser pages closed after each fetch in headless mode

### Risk 4: Filename Conflicts

**Risk:** Auto-generated filenames may collide
**Mitigation:**

- Already handled by existing `ResolveConflict()` function
- Appends `-1`, `-2`, etc. automatically
- Single timestamp for batch reduces collisions

## Success Metrics

### Functional Requirements

- ✅ Accept 0+ URLs via command line arguments
- ✅ Read and parse URL file with comments and blank lines
- ✅ Combine file URLs and inline URLs
- ✅ Auto-prepend https:// if missing
- ✅ Save each URL to auto-generated filename
- ✅ Support --output-dir for custom save location
- ✅ Continue on error by default (log failures to stderr)
- ✅ Exit code 0 if all succeed, 1 if any fail
- ✅ Progress indicators: [N/M] format
- ✅ Summary: X succeeded, Y failed
- ✅ Error on conflicting flags (-o, --close-tab, --tab, etc.)
- ✅ --open-browser just opens tabs, no output
- ✅ All existing tests pass
- ✅ New tests achieve >80% coverage

### Non-Functional Requirements

- ✅ Performance: <100ms overhead per additional URL
- ✅ Memory: Constant memory usage (stream file reading)
- ✅ Error messages: Clear, actionable, with line numbers
- ✅ Documentation: Complete usage examples and migration guide
- ✅ Backward compatible: Single URL behavior unchanged (except --open-browser)

## Open Questions

All questions resolved during design discussion ✅

## License

Mozilla Public License 2.0 (consistent with existing snag project)

## References

- **AGENTS.md**: Current snag architecture and conventions
- **README.md**: User-facing documentation
- **docs/design-record.md**: Design decisions and rationale
- **handlers.go**: Existing `handleAllTabs()` function (similar batch pattern)
- **output.go**: Existing filename generation and conflict resolution
