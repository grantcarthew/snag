# PROJECT: Multiple URL Support

## Project Status

**Status:** ✅ COMPLETED (2025-10-22)
**Feature Type:** Enhancement
**Priority:** Medium
**Complexity:** Medium
**Implementation Time:** ~4 hours (vs. estimated 13-18 hours)

## Implementation Summary

Successfully implemented clean, DRY multiple URL support for snag, enabling batch content retrieval for AI agents and automation workflows. The implementation follows existing codebase patterns (especially `handleAllTabs`) and maintains backward compatibility.

### ✅ Completed Features

1. **Multiple URL Arguments**: `snag --force-headless -d ./output example.com go.dev`
2. **URL File Support**: `snag --url-file urls.txt`
3. **Combined Input**: `snag --url-file urls.txt extra-url.com`
4. **Comment Support**: Full-line (`# comment`, `// comment`) and inline (`url # comment`)
5. **Auto-HTTPS Prepending**: URLs without scheme get `https://` prepended
6. **Continue-on-Error**: Processes all URLs even if some fail
7. **Batch Summary**: "Batch complete: X succeeded, Y failed"
8. **Progress Indicators**: `[N/M]` format with success/failure markers
9. **Flag Validation**: Errors for conflicting flags (`--output`, `--close-tab` with multiple URLs)
10. **Flag Order Validation**: Helpful error if flags come after URLs

### Files Modified

1. **validate.go** (validate.go:244-317):

   - Added `loadURLsFromFile()` - Parses URL files with comment support
   - Comment parsing: `#` and `//` for full-line and inline comments
   - Auto-prepends `https://` to URLs without scheme
   - Validates URLs and logs warnings for invalid lines (line numbers included)

2. **errors.go** (errors.go:46-53):

   - `ErrNoValidURLs` - No valid URLs provided
   - `ErrOutputFlagConflict` - `--output` cannot be used with multiple URLs
   - `ErrCloseTabMultipleURLs` - `--close-tab` not supported with multiple URLs

3. **main.go** (main.go:9-351):

   - Added `--url-file` flag for URL file input
   - Updated `ArgsUsage` from `<url>` to `[url...]`
   - Implemented URL collection logic (file + command-line)
   - Added flag order validation (detects flags after URLs)
   - Routing logic: single URL vs. multiple URLs
   - `--open-browser` with URLs handler routing

4. **handlers.go** (handlers.go:456-673):
   - `handleOpenURLsInBrowser()` - Opens multiple URLs in browser tabs (no output)
   - `handleMultipleURLs()` - Batch URL fetching following `handleAllTabs()` pattern
   - Reuses existing functions: `generateOutputFilename()`, `processPageContent()`
   - Single timestamp for entire batch (reduces filename collisions)
   - Continue-on-error with success/failure tracking

### Testdata Created

Created comprehensive test files in `testdata/`:

1. **testdata/urls.txt** - Full-featured test file demonstrating:

   - Full-line comments (`#` and `//`)
   - Inline comments (`url # comment`, `url // comment`)
   - Auto-HTTPS prepending
   - Blank lines
   - URLs with paths and query parameters

2. **testdata/urls-invalid.txt** - Error handling test file:

   - Mix of valid and invalid URLs
   - URLs with spaces (should warn and skip)
   - Malformed URLs (should warn and skip)
   - Tests continue-on-error behavior

3. **testdata/urls-small.txt** - Quick 2-URL test file for rapid testing

## Key Design Decisions

### 1. CLI Interface

**Multiple URL Sources:**

```bash
snag --force-headless -d ./output url1 url2 url3  # Inline URLs (flags FIRST)
snag --url-file urls.txt                          # From file
snag --url-file urls.txt url4 url5                # Combined
```

**Important: Flags Must Come Before URLs** (urfave/cli convention)

```bash
✓ snag --force-headless -d ./output example.com go.dev
✗ snag example.com go.dev --force-headless -d ./output
```

**Validation Rules:**

- Zero URLs → Error: "No valid URLs provided"
- One URL → Single URL flow (backward compatible)
- Two or more URLs → Batch processing mode

### 2. URL File Format

**Supported Syntax:**

```
# Full-line comments (# or //)
// Both comment styles supported

example.com                    # Auto-prepends https://
https://go.dev/doc/            # Uses as-is
http://localhost:8080          # Preserves http://

  github.com/user/repo         # Whitespace trimmed

                               # Blank lines ignored

https://example.com # Inline comment with #
https://go.dev // Inline comment with //
```

**Processing Rules:**

1. `TrimSpace()` on each line
2. Skip empty lines
3. Skip full-line comments (`#` or `//`)
4. Handle inline comments (` #` or ` //`)
5. Warn and skip lines with space but no comment marker
6. Auto-prepend `https://` if missing protocol
7. Validate URL format
8. Log warnings for invalid URLs (with line numbers) and continue

### 3. Output Behavior

**Multiple URLs (2+) Always Save to Disk:**

```bash
snag url1 url2 url3                        # Saves to ./{auto-generated-names}
snag --force-headless -d ./results url1 url2  # Saves to ./results/
```

**Single URL (Backward Compatible):**

```bash
snag url1                                  # Outputs to stdout
snag --force-headless -o output.md url1    # Saves to output.md
snag --force-headless -d ./results url1    # Saves to ./results/{auto-generated}
```

**Filename Generation:**

- Format: `yyyy-mm-dd-hhmmss-{title-slug}.{ext}`
- Single timestamp for entire batch (reduces collisions)
- Conflict resolution: Appends `-1`, `-2`, etc. (existing `ResolveConflict()`)
- Extension: Auto-detected from format (`.md`, `.html`, `.txt`, `.pdf`, `.png`)

### 4. Error Handling

**Continue-on-Error Strategy:**

- Invalid URL in file → Warning to stderr, skip line, continue
- Network/timeout error → Error to stderr, skip URL, continue
- File write error → Error to stderr, skip URL, continue
- Browser connection error → Fatal error, exit immediately

**Exit Codes:**

- Exit 0: ALL URLs succeeded
- Exit 1: ANY URL failed

**Progress Indicators:**

```
Processing 5 URLs...
[1/5] Fetching: https://example.com
✓ Saved to /tmp/snag-test/2025-10-22-141739-example-domain.md (0.2 KB)
[2/5] Fetching: https://go.dev
✓ Saved to /tmp/snag-test/2025-10-22-141739-the-go-programming-language.md (14.0 KB)
[3/5] Fetching: invalid-url.com
✗ [3/5] Invalid URL - skipping: invalid URL
Batch complete: 2 succeeded, 1 failed
```

### 5. Flag Behavior with Multiple URLs

| Flag               | Single URL            | Multiple URLs          | Notes                      |
| ------------------ | --------------------- | ---------------------- | -------------------------- |
| `-o, --output`     | ✓ Saves to file       | ✗ Error                | Use `--output-dir` instead |
| `-d, --output-dir` | ✓ Auto-generates name | ✓ Auto-generates names | All files saved here       |
| `--format`         | ✓ Applies             | ✓ Applies to all       | All URLs use same format   |
| `--timeout`        | ✓ Applies             | ✓ Applies to each      | Per-URL timeout            |
| `--wait-for`       | ✓ Applies             | ✓ Applies to each      | Waits on each page         |
| `--close-tab`      | ✓ Closes after fetch  | ✗ Error                | Ambiguous for multiple     |
| `--open-browser`   | ✓ Opens, no output    | ✓ Opens all, no output | Just opens tabs            |
| `--tab`            | ✓ Fetches from tab    | ✗ Error                | Conflicts with URLs        |
| `--all-tabs`       | ✓ Fetches all tabs    | ✗ Error                | Conflicts with URLs        |
| `--list-tabs`      | ✓ Lists tabs          | ✗ Error                | Standalone only            |

### 6. --open-browser Behavior

**With URLs (New Feature):**

```bash
snag --open-browser example.com go.dev    # Opens tabs, NO output
# Use `snag --list-tabs` to see tabs
# Use `snag --tab <index>` to fetch content from tabs
```

**Without URLs (Existing):**

```bash
snag --open-browser                       # Opens blank browser
```

## Testing Results

### Manual Testing Completed

✅ **Multiple inline URLs:**

```bash
./snag --force-headless -d /tmp/snag-test example.com httpbin.org/html
# Result: 2 files saved with auto-generated names
```

✅ **URL file:**

```bash
./snag --url-file testdata/urls-small.txt --force-headless
# Result: Loaded 2 URLs, both saved successfully
```

✅ **Combined file + inline:**

```bash
./snag --url-file testdata/urls-small.txt --force-headless go.dev
# Result: Processed 3 URLs (2 from file + 1 inline)
```

✅ **Error handling (invalid URLs):**

```bash
./snag --url-file testdata/urls-invalid.txt --force-headless -d /tmp/snag-test
# Result: Warnings for invalid lines, continued processing valid URLs
```

✅ **Flag validation:**

```bash
./snag --force-headless -o output.md example.com go.dev
# Result: Error - "Cannot use --output with multiple URLs"

./snag --force-headless --close-tab example.com go.dev
# Result: Error - "Cannot use --close-tab with multiple URLs"
```

✅ **Flag order validation:**

```bash
./snag example.com --force-headless
# Result: Error - "Flags must come before URL arguments"
```

## Implementation Insights

### DRY Principles Applied

1. **Reused Existing Functions:**

   - `generateOutputFilename()` from output.go
   - `processPageContent()` from handlers.go
   - `validateURL()`, `validateFormat()`, `validateTimeout()`, `validateDirectory()` from validate.go
   - Pattern matching from `handleAllTabs()` in handlers.go

2. **Consistent Error Handling:**

   - Followed existing sentinel error pattern in errors.go
   - Used existing logger methods (Success, Error, Warning, Info, Verbose)
   - Maintained [N/M] progress indicator format from `handleAllTabs()`

3. **Minimal Code Duplication:**
   - `handleMultipleURLs()` mirrors `handleAllTabs()` structure (handlers.go:228-333)
   - Single timestamp for batch (same pattern as tabs)
   - Continue-on-error with success/failure counting (same pattern)

### Key Learnings

1. **urfave/cli Flag Ordering:**

   - Flags MUST come before positional arguments
   - Added validation to catch this common user error
   - Provides helpful error message with correct usage example

2. **URL File Parsing:**

   - Stream processing with `bufio.Scanner` (memory efficient)
   - Line number tracking for clear error messages
   - Continue-on-error approach for robustness

3. **Browser Lifecycle:**

   - Reuse browser connection for batch processing
   - Close pages in headless mode to save memory
   - Leave visible browser running for user convenience

4. **Timestamp Strategy:**
   - Single timestamp for entire batch reduces filename collisions
   - Follows same pattern as `handleAllTabs()` (handlers.go:273)

## Usage Examples

### Basic Batch Fetching

```bash
# Multiple URLs, saves to current directory
snag --force-headless example.com go.dev httpbin.org/html

# Multiple URLs, saves to specific directory
snag --force-headless -d ./output example.com go.dev

# From URL file
snag --url-file urls.txt --force-headless

# Combined file + inline URLs
snag --url-file urls.txt --force-headless github.com
```

### With Format Options

```bash
# HTML format
snag --force-headless --format html -d ./output example.com go.dev

# PDF format (auto-generates filenames)
snag --force-headless --format pdf example.com go.dev

# PNG screenshots
snag --force-headless --format png -d ./screenshots example.com go.dev
```

### Open URLs in Browser (No Output)

```bash
# Opens 3 tabs in browser, no content output
snag --open-browser example.com go.dev github.com

# Then fetch content from tabs
snag --list-tabs
snag --tab 1 -o example.md
snag --tab 2 -o godev.md
```

### With Advanced Options

```bash
# With timeout and wait-for selector
snag --force-headless --timeout 60 --wait-for ".content" -d ./output url1 url2

# With custom user agent
snag --force-headless --user-agent "Custom Agent" -d ./output url1 url2
```

## Dependencies

**No new external dependencies added.**

Uses existing:

- `github.com/urfave/cli/v2` - CLI framework
- `github.com/go-rod/rod` - Browser control
- `github.com/JohannesKaufmann/html-to-markdown/v2` - HTML to Markdown conversion
- Standard library: `bufio`, `os`, `path/filepath`, `strings`

## Future Enhancements (Not Implemented)

### Potential Improvements

1. **Parallel Processing:**

   - Currently sequential (one URL at a time)
   - Could add `--parallel <N>` flag for concurrent fetching
   - Would require careful browser tab management

2. **Retry Logic:**

   - Could add `--retry <N>` flag for failed URLs
   - Exponential backoff between retries
   - Retry only network/timeout errors, not validation errors

3. **URL File Formats:**

   - Support CSV with metadata (URL, custom filename, format)
   - Support JSON with per-URL configuration
   - Would enable per-URL timeout, format, wait-for settings

4. **Progress Bar:**

   - Replace `[N/M]` with animated progress bar
   - Show estimated time remaining
   - Requires terminal library (new dependency)

5. **Resume Failed:**
   - Generate `.failed.txt` with URLs that failed
   - Allow resuming batch from failed list
   - Useful for large batches with intermittent failures

## Documentation Updates Needed

### README.md

- Add "Batch Processing" section with examples
- Add URL file format documentation
- Add flag order note (flags before URLs)
- Add `--url-file` flag documentation

### AGENTS.md

- Update test commands with multiple URL examples
- Add troubleshooting for URL file errors
- Add note about flag ordering

### CHANGELOG.md

- Add entry for multiple URL support feature
- Note about flag ordering requirement

## License

Mozilla Public License 2.0 (consistent with existing snag project)

## References

- **AGENTS.md**: Snag architecture and conventions
- **README.md**: User-facing documentation
- **handlers.go:228-333**: `handleAllTabs()` pattern (template for batch processing)
- **output.go**: Filename generation and conflict resolution
- **validate.go**: Input validation functions
