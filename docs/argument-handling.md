# Argument Handling Reference

**Purpose:** Complete specification of all argument/flag combinations and their interactions.

**Status:** Current implementation + Planned features (marked with 🚧)

**Last Updated:** 2025-10-23

---

## Table of Contents

1. [All Arguments and Flags](#all-arguments-and-flags)
2. [Quick Reference Matrix](#quick-reference-matrix)
3. [Mutually Exclusive Combinations](#mutually-exclusive-combinations)
4. [Mode-Based Behavior](#mode-based-behavior)
5. [Output Routing Rules](#output-routing-rules)
6. [Special Cases and Edge Cases](#special-cases-and-edge-cases)
7. [Validation Order](#validation-order)
8. [Undefined Behaviors](#undefined-behaviors)

---

## Design Decisions by Argument

This section documents the detailed design decisions for each argument, including validation rules, error handling, and interaction behaviors.

### `<url>` Positional Argument ✅

**Status:** Complete (2025-10-22)

#### Validation Rules

**Protocol Handling:**
- Auto-add `https://` if no protocol is present
- Valid schemes: `http`, `https`, `file` only
- Invalid schemes (e.g., `ftp://`, `data:`) → Error in validation

**URL Validation:**
- Validate using Go's `url.Parse()` before passing to browser
- Must have valid URL characters only
- Localhost and private IPs are allowed (e.g., `http://localhost:3000`, `http://192.168.1.1`)
- Connection failures are handled by browser, not validation

**Error Messages:**
- Invalid URL format: `"Invalid URL format: {url}"`
- Invalid scheme: `"Invalid URL scheme '{scheme}'. Supported: http, https, file"`

#### Multiple URLs Behavior

**With No Output Flag:**
```bash
snag https://example.com https://google.com
```
- Behavior: Auto-generate filenames in current directory
- Browser mode: Headless if no browser open
- Each URL gets separate file: `{timestamp}-{slug}.{ext}`

**With `--output FILE`:**
```bash
snag -o output.md https://example.com https://google.com
```
- Behavior: **Error** - Cannot combine multiple sources into single output file
- Error message: `"Cannot use --output with multiple URLs. Use --output-dir instead"`

**With `--output-dir DIR`:**
```bash
snag -d output/ https://example.com https://google.com
```
- Behavior: Auto-generate separate filenames in specified directory
- Browser mode: Headless if no browser open
- Each URL gets separate file in directory

#### Interaction Matrix

**Content Source Conflicts:**

| Combination | Behavior | Rationale |
|-------------|----------|-----------|
| `<url>` + `--url-file` | **Merge** both sources | Allow combining CLI URLs with file URLs |
| `<url>` + `--tab` | **Error** | Mutually exclusive content sources |
| `<url>` + `--all-tabs` | **Error** | Mutually exclusive content sources |
| `<url>` + `--list-tabs` | **Error** | List-tabs is standalone action only |

**Browser Mode:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `<url>` + `--open-browser` | Navigate to URL, **do not fetch** content, leave browser open | Opens URL in tab for manual interaction |
| `<url>` + `--force-headless` | Force headless mode even if browser already open | Override auto-detection |

**Output Control:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `<url>` + `--format` | Works normally | Apply format to fetched content |
| `<url>` + `--timeout` | Works normally | Apply timeout to page load |
| `<url>` + `--wait-for` | Works normally | Wait for selector after navigation |
| `<url>` + `--port` | Works normally | Use specified remote debugging port |
| `<url>` + `--user-agent` | Works normally | Set user agent for new page |

**Special Behaviors:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `<url>` + `--close-tab` | Close tab if browser visible; **ignored** if headless | Headless mode already closes tabs by default |

**Logging Flags:**
- `--verbose`, `--quiet`, `--debug`: All work normally with `<url>`

#### Examples

**Valid:**
```bash
snag https://example.com                           # Fetch to stdout
snag example.com                                   # Auto-adds https://
snag http://localhost:3000                         # Local development
snag file:///path/to/file.html                     # Local file
snag https://example.com -o page.md                # Save to file
snag https://example.com --open-browser            # Open URL, no fetch
```

**Invalid:**
```bash
snag ftp://example.com                             # ERROR: Invalid scheme
snag https://example.com --tab 1                   # ERROR: Conflicting sources
snag https://example.com --all-tabs                # ERROR: Conflicting sources
snag https://example.com --list-tabs               # ERROR: List-tabs standalone
snag url1 url2 -o out.md                           # ERROR: Multiple URLs need -d
```

### `--url-file FILE` ✅

**Status:** Complete (2025-10-22)

#### Validation Rules

**File Access:**
- File must exist and be readable
- File path can be relative or absolute
- Permission denied → Error: `"failed to open URL file: {error}"`
- File not found → Error: `"Failed to open URL file: {filename}"`
- Path is directory → Error: `"failed to open URL file: {error}"`

**File Format:**
- One URL per line
- Blank lines are ignored
- Full-line comments: Lines starting with `#` or `//`
- Inline comments: Text after ` #` or ` //` (space + marker)
- Auto-prepend `https://` if no protocol present (same as `<url>` argument)
- Invalid URLs are skipped with warning (continues processing)

**Content Validation:**
- Empty file + no inline URLs → Error: `ErrNoValidURLs`
- Empty file + inline URLs → Process inline URLs only
- Only comments/blank lines → Error: `ErrNoValidURLs`
- All invalid URLs → Warning for each, then Error: `ErrNoValidURLs`
- Mixed valid/invalid URLs → Warning for invalid, continue with valid
- No size limit (10,000+ URLs will process sequentially)

**URL Validation (per line):**
- URLs with space but no comment marker → Warning and skip: `"Line {N}: URL contains space without comment marker - skipping"`
- Invalid URL format → Warning and skip: `"Line {N}: Invalid URL - skipping"`
- Valid schemes: `http`, `https`, `file` (same as `<url>` argument)

**Multiple Files:**
- Multiple `--url-file` flags → Error: `"Only one --url-file allowed"`
- User must manually merge files if needed

**Error Messages:**
- File not found: `"Failed to open URL file: {filename}"`
- No valid URLs: `"no valid URLs found"`
- Invalid URL in file: `"Line {N}: Invalid URL - skipping: {url}"`
- Space without comment: `"Line {N}: URL contains space without comment marker - skipping: {line}"`
- Multiple url-file flags: `"Only one --url-file allowed"`

#### Behavior

**Basic Usage:**
```bash
snag --url-file urls.txt
```
- Loads all valid URLs from file
- Auto-saves to current directory with auto-generated names (never stdout)
- Processes as batch operation

**File Format Example:**
```
# Comments start with # or //
example.com                           # Auto-adds https://
https://go.dev                        # Explicit protocol
httpbin.org/html // Inline comment   # Space + // marker

# Blank lines ignored
go.dev/doc/
```

**Combining with CLI URLs:**
```bash
snag --url-file urls.txt https://example.com https://go.dev
```
- Merges file URLs with command-line URLs
- File URLs loaded first, then CLI args appended
- All URLs processed as single batch
- Auto-saves all to current directory

**Invalid URLs Handling:**
- Invalid URLs in file are skipped with warnings
- Processing continues with valid URLs
- Exit code 0 if at least one URL succeeds
- Exit code 1 if all URLs fail

#### Interaction Matrix

**Content Source Conflicts:**

| Combination | Behavior | Rationale |
|-------------|----------|-----------|
| `--url-file` + `<url>` arguments | **Merge** both sources | Allows combining file with additional URLs |
| `--url-file` + another `--url-file` | **Error** | Only one --url-file flag allowed, user must merge files |
| `--url-file` + `--tab` | **Error** | Mutually exclusive content sources |
| `--url-file` + `--all-tabs` | **Error** | Mutually exclusive content sources |
| `--url-file` + `--list-tabs` | **Error** | List-tabs is standalone action only |

**Output Control:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--url-file` alone | Auto-save to current dir | Each URL gets auto-generated filename |
| `--url-file` + `--output FILE` | **Error** | Cannot combine multiple sources into single file |
| `--url-file` + `--output-dir DIR` | Works normally | Save all to specified directory |
| `--url-file` + `--format md/html/text` | Works normally, auto-save | Apply format to all URLs, save with generated names |
| `--url-file` + `--format pdf/png` | Works normally, auto-save | Binary formats always auto-save |

**Browser Mode:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--url-file` + `--open-browser` | Opens all URLs in tabs, **no fetch** | Only --open-browser prevents fetching |
| `--url-file` + `--force-headless` | Works normally, auto-save | Force headless, fetch all URLs |

**Page Loading:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--url-file` + `--timeout` | Works normally | Applied to each URL individually |
| `--url-file` + `--wait-for` | Works normally | Wait for selector on every page |
| `--url-file` + `--user-agent` | Works normally | Applied to all new pages |
| `--url-file` + `--port` | Works normally | Use specified port for browser |

**Special Behaviors:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--url-file` + `--close-tab` | **Error** | Ambiguous for batch operations |

**Logging Flags:**
- `--verbose`: Works normally - show verbose logs for all URLs
- `--quiet`: Works normally - suppress all except errors
- `--debug`: Works normally - show debug logs for all URLs

#### Examples

**Valid:**
```bash
snag --url-file urls.txt                           # Batch to current dir
snag --url-file urls.txt -d ./output               # Batch to specific dir
snag --url-file urls.txt https://example.com       # Combined sources
snag --url-file urls.txt --format html             # HTML format, auto-save
snag --url-file urls.txt --force-headless          # Force headless
snag --url-file urls.txt -d ./out --format pdf     # PDF batch to directory
snag --url-file urls.txt --timeout 60              # 60s timeout per URL
snag --url-file urls.txt --wait-for ".content"     # Wait on every page
snag --url-file urls.txt --open-browser            # Open in tabs, no fetch
```

**Invalid:**
```bash
snag --url-file /nonexistent.txt                   # ERROR: File not found
snag --url-file empty.txt                          # ERROR: No valid URLs (if no inline URLs)
snag --url-file urls.txt -o output.md              # ERROR: Multiple URLs need -d
snag --url-file urls.txt --tab 1                   # ERROR: Conflicting sources
snag --url-file urls.txt --all-tabs                # ERROR: Conflicting sources
snag --url-file urls.txt --list-tabs               # ERROR: List-tabs standalone
snag --url-file urls.txt --close-tab               # ERROR: Ambiguous for batch
snag --url-file f1.txt --url-file f2.txt           # ERROR: Only one --url-file allowed
```

**File Format Examples:**

`urls.txt`:
```
# Example URL file
example.com
https://go.dev/doc/
httpbin.org/html # Testing endpoint

// Comment with //
go.dev

# Blank lines are fine


https://www.iana.org/
```

**Invalid URLs Behavior:**

`mixed.txt`:
```
example.com                    # Valid
invalid url with spaces        # Skipped with warning
go.dev                         # Valid
```

Output:
```
Line 2: URL contains space without comment marker - skipping: invalid url with spaces
Processing 2 URLs...
[1/2] Fetching: https://example.com
✓ Saved to ./2025-10-22-124752-example-domain.md
[2/2] Fetching: https://go.dev
✓ Saved to ./2025-10-22-124752-the-go-programming-language.md
Batch complete: 2 succeeded, 0 failed
```

#### Implementation Details

**Location:**
- Flag definition: `main.go:78-81`
- Handler logic: `main.go:200-207`
- URL loading: `validate.go:251-317`

**Processing Order:**
1. Load URLs from file (if `--url-file` provided)
2. Append command-line URL arguments
3. Validate all URLs in merged list
4. Process as batch operation with auto-generated filenames

**Key Functions:**
- `loadURLsFromFile(filename)` - Reads and validates URL file (validate.go:251-317)
- File validation happens before CLI URL validation
- Invalid URLs logged with line numbers for debugging

**Output Behavior:**
- URLs from `--url-file` always trigger batch mode (auto-save, never stdout)
- Even single URL from file gets auto-generated filename
- Combines with inline URLs for total count

### `--output-dir DIRECTORY` / `-d` ✅

**Status:** Complete (2025-10-22)

#### Validation Rules

**Path Handling:**
- Both relative and absolute paths supported
- Paths with spaces supported (user must quote in shell: `-d "my output dir"`)
- Path validation using Go's `os.Stat()` and `fileInfo.IsDir()`
- Empty string → Use current working directory (pwd)
- Relative paths resolved relative to current directory

**Directory Validation:**
- Directory must exist → Error: `"Output directory does not exist: {path}"`
- Path exists but is a file (not directory) → Error: `"Output directory path is a file, not a directory: {path}"`
- Permission denied (no write access) → Error: `"Cannot write to output directory: permission denied"`

**Multiple Directory Conflicts:**
- Multiple `-d` flags → Error: `"Only one --output-dir flag allowed"`

**Error Messages:**
- Directory doesn't exist: `"Output directory does not exist: {path}"`
- Path is file: `"Output directory path is a file, not a directory: {path}"`
- Permission denied: `"Cannot write to output directory: permission denied"`
- Empty after trim: Uses current directory (no error)
- Multiple flags: `"Only one --output-dir flag allowed"`

#### Behavior

**Basic Usage:**
```bash
snag https://example.com -d ./output
```
- Fetches URL content
- Generates filename automatically: `yyyy-mm-dd-hhmmss-{page-title-slug}.{ext}`
- Writes to `./output/{generated-filename}`
- Creates file if it doesn't exist

**Filename Generation:**
```bash
snag https://example.com -d ./docs
# Creates: ./docs/2025-10-22-214530-example-domain.md
```
- Format: `yyyy-mm-dd-hhmmss-{title-slug}.{ext}`
- Extension matches `--format` flag
- Page title used for slug (sanitized, lowercased, hyphens)

**Collision Handling:**
```bash
snag https://example.com -d ./output  # file.md
snag https://example.com -d ./output  # file-1.md (if collision)
snag https://example.com -d ./output  # file-2.md (if collision)
```
- Append counter suffix: `-1`, `-2`, `-3`, etc.
- Ensures no file overwriting

**Empty String Behavior:**
```bash
snag https://example.com -d ""
# Uses current working directory (pwd)
```

#### Interaction Matrix

**Output Destination Conflicts:**

| Combination | Behavior | Rationale |
|-------------|----------|-----------|
| `-d directory/` + `-o file.md` | **Error** | Mutually exclusive output destinations |
| Multiple `-d` flags | **Error** | Only one output directory allowed |

**Error messages:**
- `-d` + `-o`: `"Cannot use --output and --output-dir together"`
- Multiple `-d`: `"Only one --output-dir flag allowed"`

**Content Source Interactions:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `-d ./dir` + single `<url>` | Works normally | Auto-generated filename in directory |
| `-d ./dir` + multiple `<url>` | Works normally | Each URL gets separate auto-generated file |
| `-d ./dir` + `--url-file` | Works normally | Each URL from file gets separate file |
| `-d ./dir` + `--tab <pattern>` | Works normally | Tab content saved with auto-generated name |
| `-d ./dir` + `--all-tabs` | Works normally | Each tab gets separate auto-generated file |

**Special Operation Modes:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `-d ./dir` + `--list-tabs` | **Error** | List-tabs outputs to stdout only, no content |
| `-d ./dir` + `--open-browser` (no URL/url-file) | `-d` ignored | Only opens browser, nothing to fetch/save |
| `-d ./dir` + `--open-browser` + `<url>` | Works normally | Opens browser, fetches URL, saves to directory |

**Error messages:**
- `--list-tabs`: `"Cannot use --output-dir with --list-tabs (informational output only)"`

**Format Combinations:**

All format combinations work normally:

| Scenario | Behavior | Output Extension |
|----------|----------|------------------|
| `-d ./dir --format md` | Auto-save as Markdown | `.md` |
| `-d ./dir --format html` | Auto-save as HTML | `.html` |
| `-d ./dir --format text` | Auto-save as plain text | `.txt` |
| `-d ./dir --format pdf` | Auto-save as PDF | `.pdf` |
| `-d ./dir --format png` | Auto-save as PNG | `.png` |

**Compatible Flags:**

All these flags work normally with `-d`:

- ✅ `--format` - Apply format (extension matches format)
- ✅ `--timeout` - Apply timeout to page load
- ✅ `--wait-for` - Wait for selector before extraction
- ✅ `--close-tab` - Close tab after fetching
- ✅ `--force-headless` - Force headless browser mode
- ✅ `--verbose` / `--quiet` / `--debug` - Logging levels
- ✅ `--user-agent` - Custom user agent (applies to all new pages)
- ✅ `--port` - Remote debugging port

#### Examples

**Valid:**
```bash
snag https://example.com -d ./output                 # Basic usage
snag https://example.com -d ./docs/pages             # Nested directory
snag https://example.com -d /tmp/snag                # Absolute path
snag https://example.com -d "my output dir"          # Path with spaces
snag https://example.com -d ""                       # Current directory
snag https://example.com -d ./output --format html   # HTML format
snag https://example.com -d ./output --format pdf    # PDF format
snag url1 url2 url3 -d ./batch                       # Multiple URLs
snag --url-file urls.txt -d ./results                # URL file
snag --tab 1 -d ./tabs                               # From existing tab
snag --all-tabs -d ./all                             # All tabs
snag https://example.com -d ./out --timeout 60       # With timeout
snag https://example.com -d ./out --wait-for ".content"  # With wait
```

**Invalid:**
```bash
snag https://example.com -d /nonexistent/dir         # ERROR: Directory doesn't exist
snag https://example.com -d ./existing-file.txt      # ERROR: Path is file, not directory
snag https://example.com -d /root/restricted         # ERROR: Permission denied
snag https://example.com -d ./dir -o file.md         # ERROR: -d and -o conflict
snag https://example.com -d ./dir1 -d ./dir2         # ERROR: Multiple -d flags
snag --list-tabs -d ./tabs                           # ERROR: List-tabs standalone
snag --open-browser -d ./output                      # OK but -d ignored (nothing to fetch)
```

#### Implementation Details

**Location:**
- Flag definition: `main.go:68-71`
- Handler logic: Various handler functions
- Path validation: `validate.go` functions
- Filename generation: `output.go:generateOutputFilename()`

**Processing Flow:**
1. Validate output directory path (exists, is directory, writable)
2. Check for conflicts (`-o`, multiple `-d`)
3. Fetch content from source(s)
4. Generate filename(s) with timestamp and page title
5. Check for collisions → append counter if needed
6. Write to directory

**Filename Generation:**
- Function: `generateOutputFilename(pageTitle, format, outputDir string)`
- Timestamp: Single timestamp for batch operations
- Title slug: Sanitized, lowercase, hyphens for spaces
- Extension: Matches format (`.md`, `.html`, `.txt`, `.pdf`, `.png`)
- Collision resolution: Append `-1`, `-2`, etc.

**Special Behaviors:**
- Empty string for directory → Resolves to current working directory
- Relative paths → Resolved relative to current directory
- `--open-browser` without URL/url-file → Directory flag ignored (nothing to save)

### `--timeout SECONDS` ✅

**Status:** Complete (2025-10-22)

#### Validation Rules

**Value Validation:**
- Must be a positive integer (> 0)
- Negative value → Error: `"Timeout must be positive"`
- Zero value → Error: `"Timeout must be positive"`
- Non-integer (decimal) → Error: `"Timeout must be an integer"`
- Non-numeric string → Error: `"Invalid timeout value: {value}"`
- Empty value → Parse error from CLI framework
- Extremely large values (e.g., 999999999) → **Allowed** (user responsibility)

**Multiple Timeout Flags:**
- Multiple `--timeout` flags → Error: `"Only one --timeout flag allowed"`

**Error Messages:**
- Negative/zero: `"Timeout must be positive"`
- Non-integer: `"Timeout must be an integer"`
- Non-numeric: `"Invalid timeout value: {value}"`
- Multiple flags: `"Only one --timeout flag allowed"`

#### Behavior

**Scope:**
- Timeout applies **only to page navigation** (not to format conversion, PDF/PNG generation, or selector waiting)
- Default timeout: 30 seconds (if flag not specified)

**Basic Usage:**
```bash
snag https://example.com --timeout 60
```
- Sets 60-second timeout for initial page navigation
- If page doesn't load within 60s → Error: `"Page load timeout"`
- Does not affect `--wait-for` selector timeout
- Does not affect format conversion time

**Multiple URLs:**
```bash
snag url1 url2 url3 --timeout 45
```
- Timeout applied **per-URL individually** (not total operation time)
- Each URL gets 45 seconds to load
- Total operation could take 135+ seconds for 3 URLs

**With `--url-file`:**
```bash
snag --url-file urls.txt --timeout 60
```
- Timeout applied **per-URL individually**
- Each URL in file gets 60 seconds to load

#### Interaction Matrix

**Content Source Interactions:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--timeout` + single `<url>` | Works normally | Standard timeout for navigation |
| `--timeout` + multiple `<url>` | Works normally | Applied per-URL individually |
| `--timeout` + `--url-file` | Works normally | Applied per-URL individually |
| `--timeout` + `--tab` | Works with warning | Timeout applies to `--wait-for` if present, otherwise no effect |
| `--timeout` + `--all-tabs` | Works with warning | Timeout applies to `--wait-for` if present, otherwise no effect |

**Warning messages:**
- `--tab` (no --wait-for): `"Warning: --timeout has no effect without --wait-for when using --tab"`
- `--all-tabs` (no --wait-for): `"Warning: --timeout has no effect without --wait-for when using --all-tabs"`

**Special Operation Modes:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--timeout` + `--list-tabs` | **Error** | List-tabs standalone, no navigation |
| `--timeout` + `--open-browser` (no URL) | Timeout **ignored** | No navigation occurs |

**Error messages:**
- `--list-tabs`: `"Cannot use --timeout with --list-tabs (no navigation)"`

**Timing-Related Interactions:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--timeout` + `--wait-for` | Works normally | Timeout for navigation only, `--wait-for` has separate logic |

**Output Control:**

All output flags work normally with `--timeout`:

| Combination | Behavior |
|-------------|----------|
| `--timeout` + `--output` | Works normally |
| `--timeout` + `--output-dir` | Works normally |
| `--timeout` + `--format` (all formats) | Works normally - timeout applies to navigation, not conversion |

**Browser Mode:**

All browser mode flags work normally with `--timeout`:

| Combination | Behavior |
|-------------|----------|
| `--timeout` + `--force-headless` | Works normally |
| `--timeout` + `--close-tab` | Works normally |
| `--timeout` + `--port` | Works normally |

**Logging Flags:**

All logging flags work normally:
- `--verbose`: Works normally - verbose logging of timeout behavior
- `--quiet`: Works normally - suppress timeout messages
- `--debug`: Works normally - debug logging of timeout behavior

**Miscellaneous:**

| Combination | Behavior |
|-------------|----------|
| `--timeout` + `--user-agent` | Works normally - independent concerns |

#### Examples

**Valid:**
```bash
snag https://example.com --timeout 60               # 60-second timeout
snag https://example.com --timeout 120              # 2-minute timeout
snag https://example.com --timeout 5                # Short 5-second timeout
snag https://example.com --timeout 999999999        # Very long timeout (allowed)
snag url1 url2 url3 --timeout 45                    # 45s per URL
snag --url-file urls.txt --timeout 30               # 30s per URL in file
snag https://example.com --timeout 60 -o page.md    # With output file
snag https://example.com --timeout 60 --format pdf  # With PDF format
snag https://example.com --timeout 60 --wait-for ".content"  # With wait-for
snag --open-browser https://example.com --timeout 30  # Timeout ignored (no nav in open-browser)
```

**Invalid:**
```bash
snag https://example.com --timeout -30              # ERROR: Negative value
snag https://example.com --timeout 0                # ERROR: Zero value
snag https://example.com --timeout 45.5             # ERROR: Non-integer
snag https://example.com --timeout abc              # ERROR: Non-numeric
snag https://example.com --timeout                  # ERROR: Missing value
snag https://example.com --timeout 30 --timeout 60  # ERROR: Multiple flags
snag --list-tabs --timeout 30                       # ERROR: No navigation
```

**With Warnings:**
```bash
snag --tab 1 --timeout 30                           # ⚠️  Warning: no effect without --wait-for
snag --all-tabs --timeout 30                        # ⚠️  Warning: no effect without --wait-for
snag --tab 1 --timeout 30 --wait-for ".content"     # OK: timeout applies to selector
```

**Ignored (No Error):**
```bash
snag --open-browser --timeout 30                    # Timeout ignored, browser opens
```

#### Implementation Details

**Location:**
- Flag definition: `main.go` (CLI framework)
- Timeout validation: `validate.go` functions
- Timeout application: Browser navigation functions in `fetch.go` and `browser.go`

**Processing Flow:**
1. Validate timeout value (positive integer only)
2. Check for conflicts (multiple flags, tab operations, list-tabs)
3. Apply timeout to page navigation via CDP
4. If timeout expires → Return timeout error
5. Continue with content extraction (not subject to navigation timeout)

**Scope Clarification:**
- Navigation timeout: Time for page to load and become ready
- Does NOT include:
  - Format conversion time (HTML → Markdown, Text)
  - PDF/PNG generation time (handled separately by Chrome)
  - `--wait-for` selector waiting (separate timeout logic)
  - Network retry delays
  - Browser launch time

**Default Behavior:**
- No `--timeout` flag → Default 30 seconds
- Configurable per-operation via flag

### `--format FORMAT` / `-f` ✅

**Status:** Complete (2025-10-22)

#### Validation Rules

**Format Names:**
- Valid formats: `md`, `html`, `text`, `pdf`, `png`
- Format aliases: `markdown` accepted as alias for `md`
- Case-insensitive matching: `HTML`, `Html`, `html` all valid
- Invalid format name → Error: `"Invalid format '{format}'. Supported: md, html, text, pdf, png"`
- Empty string → Error: `"Format cannot be empty"`
- Typos/close matches → Error (no fuzzy matching)

**Multiple Format Flags:**
- Multiple `--format` flags → Error: `"Only one --format flag allowed"`

**Format Without Value:**
- `--format` with no value → Parse error from CLI framework

**Error Messages:**
- Invalid format: `"Invalid format '{format}'. Supported: md, html, text, pdf, png"`
- Empty format: `"Format cannot be empty"`
- Multiple flags: `"Only one --format flag allowed"`

#### Behavior

**Basic Usage:**
```bash
snag https://example.com --format html
```
- Fetches URL content
- Converts to specified format
- Outputs according to output rules (stdout or file)

**Text Formats (md, html, text):**
```bash
snag https://example.com --format markdown       # Markdown (alias)
snag https://example.com --format html           # HTML
snag https://example.com --format text           # Plain text
```
- Output to stdout by default
- Can be redirected with `-o` or `-d`
- Suitable for piping

**Binary Formats (pdf, png):**
```bash
snag https://example.com --format pdf            # PDF
snag https://example.com --format png            # Screenshot PNG
```
- **Never output to stdout** (would corrupt terminal)
- Auto-generate filename in current directory if no `-o` or `-d`
- Always saved to file

**Extension Mismatch Warning:**
```bash
snag https://example.com --format html -o page.md
```
- **Warning message:** `"Warning: Writing HTML format to file with .md extension"`
- User intent honored (file written as requested)
- Exit code 0 (warning, not error)

**Default Format:**
- No `--format` flag → Defaults to `md` (Markdown)

#### Interaction Matrix

**Content Source Interactions:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--format` + single `<url>` | Works normally | Apply format to fetched content |
| `--format` + multiple `<url>` | Works normally | Apply same format to all URLs |
| `--format` + `--url-file` | Works normally | Apply same format to all URLs in file |
| `--format` + `--tab` | Works normally | Apply format to tab content |
| `--format` + `--all-tabs` | Works normally | Apply same format to all tabs |

**Output Destination Interactions:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--format text` (no output flag) | Output to stdout | Standard behavior for text formats |
| `--format html` (no output flag) | Output to stdout | Standard behavior for text formats |
| `--format md` (no output flag) | Output to stdout | Standard behavior for text formats |
| `--format pdf` (no output flag) | Auto-save to current dir | Binary formats never go to stdout |
| `--format png` (no output flag) | Auto-save to current dir | Binary formats never go to stdout |
| `--format` + `-o file.md` (matching ext) | Write to file | Normal operation |
| `--format html` + `-o file.md` (mismatch) | Write to file with **warning** | Extension mismatch warning shown |
| `--format pdf` + `-o file.pdf` | Write to file | Normal operation |
| `--format` + `-d directory/` | Auto-generate with correct extension | Extension matches format |

**Extension Mapping with `-d`:**

| Format | Auto-generated Extension |
|--------|-------------------------|
| `md` / `markdown` | `.md` |
| `html` | `.html` |
| `text` | `.txt` |
| `pdf` | `.pdf` |
| `png` | `.png` |

**Special Operation Modes:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--format` + `--list-tabs` | Format **ignored** | List-tabs only displays tab list, no content |
| `--format` + `--open-browser` (no URL) | Format **ignored** | Open-browser only opens browser, no content |

**Browser Mode Interactions:**

All work normally:

| Combination | Behavior |
|-------------|----------|
| `--format` + `--force-headless` | Works normally |
| `--format` + `--open-browser` + URL | Works normally (current behavior) |

**Page Loading Interactions:**

All work normally:

| Combination | Behavior |
|-------------|----------|
| `--format` + `--timeout` | Works normally - timeout applies to page load |
| `--format` + `--wait-for` | Works normally - wait before format conversion |
| `--format` + `--user-agent` | Works normally - UA set for new pages |
| `--format` + `--port` | Works normally - use specified port |
| `--format` + `--close-tab` | Works normally - close after fetching |

**Logging Flags:**

All work normally:
- `--verbose`: Format output, verbose logs to stderr
- `--quiet`: Format output, suppress logs (errors only)
- `--debug`: Format output, debug logs to stderr

#### Examples

**Valid:**
```bash
snag https://example.com --format html              # HTML to stdout
snag https://example.com --format markdown          # Markdown alias
snag https://example.com --format text              # Plain text
snag https://example.com --format pdf               # PDF auto-saved
snag https://example.com --format png               # PNG screenshot auto-saved
snag https://example.com -f HTML                    # Case-insensitive
snag https://example.com --format html -o page.html # HTML to file
snag https://example.com --format pdf -o doc.pdf    # PDF to file
snag https://example.com --format html -d ./output  # Auto-generated .html
snag url1 url2 --format pdf -d ./docs               # Batch PDF generation
snag --url-file urls.txt --format text              # Text format for all
snag --tab 1 --format html                          # HTML from tab
snag --all-tabs --format pdf -d ./tabs              # All tabs as PDFs
```

**Invalid:**
```bash
snag https://example.com --format invalid           # ERROR: Invalid format
snag https://example.com --format ""                # ERROR: Empty format
snag https://example.com --format                   # ERROR: Missing value
snag https://example.com --format html --format pdf # ERROR: Multiple flags
snag https://example.com --format markdwon          # ERROR: Typo (no fuzzy match)
```

**With Warnings:**
```bash
snag https://example.com --format html -o page.md   # ⚠️ Extension mismatch
snag https://example.com --format pdf -o doc.txt    # ⚠️ Binary to text extension
snag https://example.com --format text -o file      # ⚠️ No extension
```

#### Implementation Details

**Location:**
- Flag definition: `main.go` (CLI framework)
- Format validation: `validate.go` functions
- Format conversion: `convert.go` (Markdown/HTML/Text) and `fetch.go` (PDF/PNG)
- Extension mapping: Output generation functions

**Format Conversion Flow:**
1. Validate format name (case-insensitive, check against valid list)
2. Fetch page content (HTML)
3. Convert based on format:
   - `md` / `markdown`: HTML → Markdown (html-to-markdown library)
   - `html`: Raw HTML (pass-through)
   - `text`: HTML → Plain text (html2text library)
   - `pdf`: Chrome PDF rendering via CDP
   - `png`: Full-page screenshot via CDP
4. Route output based on format type and output flags

**Binary Format Handling:**
- PDF/PNG use Chrome's native rendering capabilities
- Never sent to stdout (would corrupt terminal)
- Auto-generate filename if no output destination specified
- Filename pattern: `yyyy-mm-dd-hhmmss-{title-slug}.{ext}`

**Extension Mismatch Detection:**
- Check output file extension against format
- Emit warning to stderr if mismatch detected
- Does not prevent operation (user intent honored)
- Warning format: `"Warning: Writing {format} format to file with {ext} extension"`

**Format Aliases:**
- `markdown` → `md` (implemented in validation)
- Case normalization happens during validation

### `--output FILE` / `-o` ✅

**Status:** Complete (2025-10-22)

#### Validation Rules

**Path Handling:**
- Both relative and absolute paths supported
- Paths with spaces supported (user must quote in shell: `-o "my file.md"`)
- Parent directory must exist → Error: `"Output path invalid: parent directory does not exist"`
- Permission denied → Error: `"Failed to write output file: permission denied"`
- Directory provided instead of file → Error: `"Output path is a directory, not a file"`
- Empty string → Error: `"Output file path cannot be empty"`

**File Validation:**
- File existence check before fetching (only for permission/path validation)
- File overwrite behavior: Silently overwrite (standard Unix `cp` behavior)
- Read-only existing file → Error: `"Cannot write to read-only file: {path}"`

**Multiple Output Conflicts:**
- Multiple `-o` flags → Error: `"Only one --output flag allowed"`

**Error Messages:**
- Invalid path: `"Output path invalid: parent directory does not exist"`
- Permission denied: `"Failed to write output file: permission denied"`
- Directory provided: `"Output path is a directory, not a file"`
- Empty string: `"Output file path cannot be empty"`
- Read-only file: `"Cannot write to read-only file: {path}"`
- Multiple flags: `"Only one --output flag allowed"`

#### Behavior

**Basic Usage:**
```bash
snag https://example.com -o output.md
```
- Fetches URL content
- Writes to specified file path
- Overwrites file if it exists
- Creates file if it doesn't exist (parent directory must exist)

**Format Interactions:**
```bash
snag https://example.com -o file.md --format html
```
- **Warning message:** `"Warning: Writing HTML format to file with .md extension"`
- User gets what they requested (no error)
- Applies to all mismatched extensions

**Binary Formats:**
```bash
snag https://example.com -o output.pdf --format pdf   # Normal
snag https://example.com -o file.md --format pdf     # Warning
```
- PDF/PNG to text extension → **Warning message:** `"Warning: Writing PDF format to file with .md extension"`
- User intent honored (file written as requested)

**No Extension:**
```bash
snag https://example.com -o myfile --format markdown
```
- **Warning message:** `"Warning: Output file has no extension, expected .md for markdown format"`
- File created without extension as requested

#### Interaction Matrix

**Content Source Interactions:**

| Combination | Behavior | Rationale |
|-------------|----------|-----------|
| `-o file.md` + single `<url>` | Works normally | Standard single-file output |
| `-o file.md` + multiple `<url>` | **Error** | Ambiguous: cannot concatenate multiple sources |
| `-o file.md` + `--url-file` | **Error** | Ambiguous: cannot concatenate multiple sources |
| Multiple `-o` flags | **Error** | Only one output destination allowed |

**Error messages:**
- Multiple URLs + `-o`: `"Cannot use --output with multiple URLs. Use --output-dir instead"`
- `--url-file` + `-o`: `"Cannot use --output with --url-file. Use --output-dir instead"`
- Multiple `-o`: `"Only one --output flag allowed"`

**Output Destination Conflicts:**

| Combination | Behavior | Rationale |
|-------------|----------|-----------|
| `-o file.md` + `-d directory/` | **Error** | Mutually exclusive output destinations |

**Error message:**
- `"Cannot use --output and --output-dir together"`

**Special Operation Modes:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `-o file.txt` + `--list-tabs` | **Error** | List-tabs outputs to stdout only |
| `-o file.md` + `--open-browser` (no URL) | **Error** | Nothing to fetch |
| `-o file.md` + `--tab <pattern>` | Works normally | Fetch from tab, save to file |
| `-o file.md` + `--tab <pattern>` (no browser) | **Error** | Tab requires running browser |

**Error messages:**
- `--list-tabs`: `"Cannot use --output with --list-tabs (informational output only)"`
- `--open-browser` only: `"Cannot use --output without content source (URL or --tab)"`
- `--tab` no browser: `"No browser instance running with remote debugging"`

**Format Combinations:**

All format combinations work, with warnings for mismatches:

| Scenario | Behavior | Warning |
|----------|----------|---------|
| `-o file.md --format html` | Write HTML to .md file | ⚠️ Yes |
| `-o file.pdf --format pdf` | Write PDF to .pdf file | No |
| `-o file.md --format pdf` | Write PDF bytes to .md | ⚠️ Yes |
| `-o myfile` (no extension) | Write to extensionless file | ⚠️ Yes |

**File Overwriting:**

| Scenario | Behavior |
|----------|----------|
| File doesn't exist | Create new file |
| File exists (writable) | Silently overwrite |
| File exists (read-only) | **Error** |

**Compatible Flags:**

All these flags work normally with `-o`:

- ✅ `--format` - Apply format (with extension mismatch warnings)
- ✅ `--timeout` - Apply timeout to page load
- ✅ `--wait-for` - Wait for selector before extraction
- ✅ `--close-tab` - Close tab after fetching
- ✅ `--force-headless` - Force headless browser mode
- ✅ `--verbose` / `--quiet` / `--debug` - Logging levels
- ✅ `--user-agent` - Custom user agent
- ✅ `--port` - Remote debugging port

#### Examples

**Valid:**
```bash
snag https://example.com -o page.md                  # Basic usage
snag https://example.com -o ./output/page.md         # Relative path
snag https://example.com -o /tmp/page.md             # Absolute path
snag https://example.com -o "my file.md"             # Path with spaces
snag https://example.com -o page.html --format html  # Matching extension
snag https://example.com -o page.pdf --format pdf    # PDF format
snag --tab 1 -o content.md                           # From existing tab
snag https://example.com -o page.md --timeout 60     # With timeout
snag https://example.com -o page.md --wait-for ".content"  # With wait
```

**Invalid:**
```bash
snag url1 url2 -o out.md                             # ERROR: Multiple URLs
snag --url-file urls.txt -o out.md                   # ERROR: Multiple sources
snag https://example.com -o out.md -o out2.md        # ERROR: Multiple -o flags
snag https://example.com -o file.md -d ./dir         # ERROR: -o and -d conflict
snag --list-tabs -o tabs.txt                         # ERROR: List-tabs standalone
snag --open-browser -o file.md                       # ERROR: Nothing to fetch
snag --tab 1 -o file.md                              # ERROR: No browser running
snag https://example.com -o /root/file.md            # ERROR: Permission denied
snag https://example.com -o /nonexistent/dir/file.md # ERROR: Parent doesn't exist
snag https://example.com -o ./existing-directory/    # ERROR: Path is directory
snag https://example.com -o ""                       # ERROR: Empty string
```

**With Warnings:**
```bash
snag https://example.com -o file.md --format html    # ⚠️ Extension mismatch
snag https://example.com -o file.txt --format pdf    # ⚠️ Binary to text ext
snag https://example.com -o myfile                   # ⚠️ No extension
```

#### Implementation Details

**Location:**
- Flag definition: `main.go:64-67`
- Handler logic: `main.go:208-235`
- Path validation: `validate.go` functions
- File writing: Various handler functions

**Processing Flow:**
1. Validate output path (parent exists, not a directory, permissions)
2. Check for conflicts (`-d`, multiple URLs, `--url-file`)
3. Fetch content from source
4. Write to file (overwrite if exists)
5. Check for extension mismatch → emit warning if needed

**Warning Implementation:**
- Extension warnings logged to stderr
- Format: `"Warning: Writing {format} format to file with {ext} extension"`
- Does not prevent operation (user intent honored)
- Exit code remains 0 if operation succeeds

### `--port PORT` / `-p` ✅

**Status:** Complete (2025-10-23)

#### Invalid Values

**Valid port range:**
- Ports 1024-65535 allowed (non-privileged ports only)
- Excludes privileged ports 1-1023 (require root/admin)
- Error immediately if outside valid range

**Validation errors:**

| Value | Behavior | Error Message |
|-------|----------|---------------|
| Negative (e.g., `-1`) | Error immediately | "Port must be between 1024 and 65535" |
| Zero | Error immediately | "Port must be between 1024 and 65535" |
| Below 1024 (e.g., `80`, `443`) | Error immediately | "Port must be between 1024 and 65535 (privileged ports not allowed)" |
| Above 65535 (e.g., `70000`) | Error immediately | "Port must be between 1024 and 65535" |
| Non-integer (e.g., `9222.5`) | Error immediately | "Invalid port value: must be an integer" |
| Non-numeric (e.g., `abc`) | Error immediately | "Invalid port value: {value}" |
| Empty string | Fall back to default (9222) | No error |
| Multiple `--port` flags | Error immediately | "Only one --port flag allowed" |
| Port in use (by non-Chrome process) | Error at connection time | "Failed to connect to port {port}: {error}" |

**Default behavior:**
- No `--port` specified → Use default port 9222

#### Behavior

**Port-specific connection:**
```bash
snag --port 9223 https://example.com
```
- Attempts to connect **only** to port 9223
- No auto-detection of browsers on other ports
- If browser responds on that port → Connect to it
- If no browser responds → Launch new browser on that port

**Connection strategy:**
1. Try to connect to specified port
2. If connection fails, attempt to launch on specified port
3. Launch may fail if another Chrome instance already running (Chrome locks profile)

#### Interaction Matrix

**Browser Mode Interactions:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--port 9222` (default) | Connect/launch on 9222 | Standard behavior |
| `--port 9223` + no browser | Launch on port 9223 | New instance |
| `--port 9223` + browser on 9223 | Connect to existing | Reuse instance |
| `--port 9223` + browser on 9222 | No cross-detection | 9222 browser not detected, try 9223 |
| `--port` + `--force-headless` | Launch headless on port | Works normally |
| `--port` + `--open-browser` | Open browser on port | Works normally |
| `--port` + `--list-tabs` | List tabs from port | Works normally |
| `--port` + `--tab` | Fetch tab from port | Works normally |
| `--port` + `--all-tabs` | Fetch all tabs from port | Works normally |

**Port + User Data Directory Pairing:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--port 9222` + `--user-data-dir ~/.snag/profile1` | Launch/connect with profile1 on port 9222 | Works normally |
| Same profile, different ports | Attempt launch, Chrome will error | Chrome locks user data directories - **documented limitation** |
| Same port, different profiles | Port in use error | Cannot run multiple browsers on same port |
| Connecting to existing + different `--user-data-dir` | **Warn**, continue connecting | "Warning: Ignoring --user-data-dir (browser already running on port X)" |
| Neither specified | Port 9222 + default profile | Current behavior preserved |

**Multi-instance support:**
- Different ports + different `--user-data-dir` → Multiple isolated browser instances
- Same port → Only one instance (conflict)
- Same profile → Chrome prevents (directory lock)

**Content Source Interactions:**

All content source flags work normally with `--port`:

| Combination | Behavior |
|-------------|----------|
| `--port` + `<url>` | Works normally - fetch URL from browser on port |
| `--port` + `--url-file` | Works normally - use browser on port for all URLs |
| `--port` + multiple URLs | Works normally - use browser on port for all |

**Output Control:**

All output flags work normally with `--port`:

| Combination | Behavior |
|-------------|----------|
| `--port` + `--output` | Works normally - just controls browser connection |
| `--port` + `--output-dir` | Works normally |
| `--port` + `--format` | Works normally |

**Page Loading:**

All page loading flags work normally with `--port`:

| Combination | Behavior |
|-------------|----------|
| `--port` + `--timeout` | Works normally |
| `--port` + `--wait-for` | Works normally |
| `--port` + `--user-agent` | Works normally - applies when launching browser |

**Browser Control:**

| Combination | Behavior |
|-------------|----------|
| `--port` + `--close-tab` | Works normally |

**Logging Flags:**
- All logging flags work normally with `--port`

#### Examples

**Valid:**
```bash
snag --port 9222 https://example.com              # Default port (explicit)
snag --port 9223 https://example.com              # Custom port
snag -p 4444 https://example.com                  # Short form
snag --port 9223 --list-tabs                      # List tabs on custom port
snag --port 9223 --tab 1                          # Fetch from tab on custom port
snag --port 9223 --open-browser                   # Open browser on custom port

# Multi-instance with user-data-dir
snag --open-browser --port 9222 --user-data-dir ~/.snag/personal
snag --open-browser --port 9223 --user-data-dir ~/.snag/work

# Fetch from different instances
snag --port 9222 --tab "gmail" -o personal.md
snag --port 9223 --tab "gmail" -o work.md
```

**Invalid:**
```bash
snag --port -1 https://example.com                # ERROR: Negative
snag --port 0 https://example.com                 # ERROR: Zero
snag --port 80 https://example.com                # ERROR: Privileged port
snag --port 70000 https://example.com             # ERROR: Above 65535
snag --port 9222.5 https://example.com            # ERROR: Non-integer
snag --port abc https://example.com               # ERROR: Non-numeric
snag --port 9222 --port 9223 https://example.com  # ERROR: Multiple flags
```

**With Warnings:**
```bash
# Browser on 9222 using profile1, connect with different profile
snag --port 9222 --user-data-dir ~/.snag/profile2 <url>
# ⚠️ Warning: Ignoring --user-data-dir (browser already running on port 9222)
```

**Documented Limitations:**
```bash
# Same profile on different ports - Chrome will error
snag --open-browser --port 9222 --user-data-dir ~/.snag/profile1
snag --open-browser --port 9223 --user-data-dir ~/.snag/profile1
# Chrome error: Profile directory is locked
```

#### Implementation Details

**Location:**
- Flag definition: `main.go` (CLI framework)
- Port validation: `validate.go` (range 1024-65535)
- Connection logic: `browser.go` (BrowserManager.Connect methods)
- Launch logic: `browser.go:240-283` (launchBrowser function)

**How it works:**
1. Validate port is in range 1024-65535
2. Check for multiple `--port` flags
3. When connecting:
   - Try to connect to specified port only
   - If fails, attempt to launch browser on that port
4. When launching:
   - Set `--remote-debugging-port={port}` flag (browser.go:261)
   - Launch browser with port configuration

**Port-specific behavior:**
- `--port` specified → Only try that specific port, no auto-detection
- No `--port` → Try default 9222, then auto-detect, then launch
- Port conflicts handled by Chrome/OS (connection failure)

**Multi-instance considerations:**
- Different ports enable multiple browser instances
- Requires different `--user-data-dir` for each instance (Chrome locks profiles)
- Each instance independent (sessions, cookies, authentication)

---

### `--verbose` ✅

**Status:** Complete (2025-10-23)

#### Validation Rules

**Boolean Flag:**
- No value required (presence = enabled)
- No validation errors possible

**Multiple Flag Conflicts:**
- Multiple `--verbose` flags → Last flag honored (standard Unix behavior)
- `--verbose` + `--quiet` → Last flag wins
- `--verbose` + `--debug` → Last flag wins

#### Behavior

**Logging Level:**
- Enables verbose logging output to stderr
- Shows additional information about operations:
  - Browser connection details
  - Page navigation steps
  - Content conversion progress
  - File writing confirmations
- Does not affect stdout content output

**Basic Usage:**
```bash
snag https://example.com --verbose
```
- Outputs page content to stdout (as normal)
- Logs verbose messages to stderr:
  - "Connecting to Chrome on port 9222..."
  - "Navigating to https://example.com..."
  - "Waiting for page load..."
  - "Converting HTML to Markdown..."
  - "Content written to stdout"

#### Interaction Matrix

**Logging Level Priority (Last Flag Wins):**

| Combination | Effective Level | Rationale |
|-------------|----------------|-----------|
| `--verbose` | Verbose | Standard verbose mode |
| `--verbose --quiet` | Quiet | Last flag wins (Unix standard) |
| `--quiet --verbose` | Verbose | Last flag wins |
| `--verbose --debug` | Debug | Last flag wins |
| `--debug --verbose` | Verbose | Last flag wins |
| `--verbose --quiet --debug` | Debug | Last flag wins |

**All Other Flags:**
- Works normally with all other flags
- Simply controls stderr logging verbosity
- No conflicts or special behaviors

#### Examples

**Valid:**
```bash
snag https://example.com --verbose                  # Verbose logging
snag https://example.com --verbose -o page.md       # Verbose + file output
snag --url-file urls.txt --verbose                  # Verbose batch processing
snag --tab 1 --verbose                              # Verbose tab fetch
snag --list-tabs --verbose                          # Verbose tab listing
snag https://example.com --quiet --verbose          # Verbose wins (last flag)
```

**No Invalid Combinations:**
- Boolean flag, no invalid values
- Works with everything

#### Implementation Details

**Location:**
- Flag definition: `main.go` (CLI framework)
- Logger initialization: `main.go:181-187`
- Logging level: `logger.go`

**Processing:**
1. Check if `--verbose` flag is present
2. If multiple logging flags, last one wins
3. Initialize logger with verbose level
4. All subsequent operations use verbose logging

**Logging Behavior:**
- Normal output: Important messages only
- Verbose output: All operational details
- Logs go to stderr (stdout reserved for content)

---

### `--quiet` / `-q` ✅

**Status:** Complete (2025-10-23)

#### Validation Rules

**Boolean Flag:**
- No value required (presence = enabled)
- No validation errors possible

**Multiple Flag Conflicts:**
- Multiple `--quiet` flags → Last flag honored (standard Unix behavior)
- `--quiet` + `--verbose` → Last flag wins
- `--quiet` + `--debug` → Last flag wins

#### Behavior

**Logging Level:**
- Suppresses all logging output to stderr except errors
- Shows only:
  - Fatal errors
  - Critical validation failures
  - Operation completion status (success/failure)
- Ideal for scripting and automation

**Basic Usage:**
```bash
snag https://example.com --quiet
```
- Outputs page content to stdout (as normal)
- Suppresses all stderr messages except errors
- Silent operation on success

#### Interaction Matrix

**Logging Level Priority (Last Flag Wins):**

| Combination | Effective Level | Rationale |
|-------------|----------------|-----------|
| `--quiet` | Quiet | Suppress all but errors |
| `--quiet --verbose` | Verbose | Last flag wins (Unix standard) |
| `--verbose --quiet` | Quiet | Last flag wins |
| `--quiet --debug` | Debug | Last flag wins |
| `--debug --quiet` | Quiet | Last flag wins |
| `--quiet --verbose --debug` | Debug | Last flag wins |

**All Other Flags:**
- Works normally with all other flags
- Simply controls stderr logging verbosity
- No conflicts or special behaviors

#### Examples

**Valid:**
```bash
snag https://example.com --quiet                    # Silent operation
snag https://example.com -q -o page.md              # Silent file save
snag --url-file urls.txt --quiet                    # Silent batch processing
snag --tab 1 --quiet                                # Silent tab fetch
snag --list-tabs --quiet                            # Silent tab listing (shows tabs only)
snag https://example.com --verbose --quiet          # Quiet wins (last flag)
```

**No Invalid Combinations:**
- Boolean flag, no invalid values
- Works with everything

#### Implementation Details

**Location:**
- Flag definition: `main.go` (CLI framework)
- Logger initialization: `main.go:181-187`
- Logging level: `logger.go`

**Processing:**
1. Check if `--quiet` flag is present
2. If multiple logging flags, last one wins
3. Initialize logger with quiet level
4. All subsequent operations suppress non-error logs

**Logging Behavior:**
- Quiet output: Errors only
- Exit code 0 on success (silent)
- Exit code 1 on failure (error message shown)

---

### `--debug` ✅

**Status:** Complete (2025-10-23)

#### Validation Rules

**Boolean Flag:**
- No value required (presence = enabled)
- No validation errors possible

**Multiple Flag Conflicts:**
- Multiple `--debug` flags → Last flag honored (standard Unix behavior)
- `--debug` + `--verbose` → Last flag wins
- `--debug` + `--quiet` → Last flag wins

#### Behavior

**Logging Level:**
- Enables maximum logging output to stderr
- Shows all verbose information plus:
  - Chrome DevTools Protocol (CDP) messages
  - Browser connection debugging
  - Internal state information
  - Detailed error traces
- For troubleshooting and development

**Basic Usage:**
```bash
snag https://example.com --debug
```
- Outputs page content to stdout (as normal)
- Logs extensive debug information to stderr
- Includes CDP protocol messages

#### Interaction Matrix

**Logging Level Priority (Last Flag Wins):**

| Combination | Effective Level | Rationale |
|-------------|----------------|-----------|
| `--debug` | Debug | Maximum logging |
| `--debug --verbose` | Verbose | Last flag wins (Unix standard) |
| `--verbose --debug` | Debug | Last flag wins |
| `--debug --quiet` | Quiet | Last flag wins |
| `--quiet --debug` | Debug | Last flag wins |
| `--debug --quiet --verbose` | Verbose | Last flag wins |

**All Other Flags:**
- Works normally with all other flags
- Simply controls stderr logging verbosity
- No conflicts or special behaviors

#### Examples

**Valid:**
```bash
snag https://example.com --debug                    # Debug logging
snag https://example.com --debug -o page.md         # Debug + file output
snag --url-file urls.txt --debug                    # Debug batch processing
snag --tab 1 --debug                                # Debug tab fetch
snag --list-tabs --debug                            # Debug tab listing
snag https://example.com --verbose --debug          # Debug wins (last flag)
```

**No Invalid Combinations:**
- Boolean flag, no invalid values
- Works with everything

#### Implementation Details

**Location:**
- Flag definition: `main.go` (CLI framework)
- Logger initialization: `main.go:181-187`
- Logging level: `logger.go`

**Processing:**
1. Check if `--debug` flag is present
2. If multiple logging flags, last one wins
3. Initialize logger with debug level
4. All subsequent operations use debug logging
5. CDP messages logged via rod's debug capabilities

**Logging Behavior:**
- Debug output: Everything (verbose + CDP messages + internals)
- Extremely detailed for troubleshooting
- Logs go to stderr (stdout reserved for content)

---

### `--help` / `-h` ✅

**Status:** Complete (2025-10-23)

#### Validation Rules

**Boolean Flag:**
- No value required (presence = enabled)
- No validation errors possible

**Priority Behavior:**
- Takes absolute priority over all other flags
- Displays help and exits immediately
- Exit code 0 (success)

#### Behavior

**Basic Usage:**
```bash
snag --help
snag -h
```
- Displays comprehensive help message
- Shows all available flags and usage
- Exits with code 0
- No other operations performed

**Help Message Contents:**
- Tool description and purpose
- Usage syntax
- All available flags with descriptions
- Examples
- Exit codes
- Links to documentation

#### Interaction Matrix

**With All Other Flags:**

| Combination | Behavior | Rationale |
|-------------|----------|-----------|
| `--help` alone | Display help, exit 0 | Standard help mode |
| `--help <url>` | Display help, exit 0 | Help takes priority, URL ignored |
| `--help --version` | Display help, exit 0 | Help takes priority over version |
| `--version --help` | Display help, exit 0 | Help takes priority (regardless of order) |
| `--help` + any flags | Display help, exit 0 | Help ignores all other flags |

**Priority Rules:**
1. `--help` detected → Display help
2. Ignore all other flags completely
3. Exit with code 0
4. `--help` takes priority over `--version`

#### Examples

**Valid (All display help and exit):**
```bash
snag --help                                         # Basic help
snag -h                                             # Short form
snag --help https://example.com                     # Help (URL ignored)
snag --help --version                               # Help (version ignored)
snag --help -o file.md --format pdf --verbose       # Help (everything ignored)
```

**No Invalid Combinations:**
- Help flag ignores all other input
- Always succeeds (exit 0)

#### Implementation Details

**Location:**
- Flag definition: Built into `github.com/urfave/cli/v2` framework
- Auto-generated help by CLI framework

**Processing:**
1. CLI framework checks for `--help` or `-h`
2. If present, displays auto-generated help
3. Exits with code 0
4. No custom validation code runs

**Help Display:**
- Auto-generated from flag definitions
- Includes usage patterns
- Shows all flags with descriptions
- Framework handles everything

---

### `--version` / `-v` ✅

**Status:** Complete (2025-10-23)

#### Validation Rules

**Boolean Flag:**
- No value required (presence = enabled)
- No validation errors possible

**Priority Behavior:**
- Displays version and exits immediately
- Exit code 0 (success)
- Lower priority than `--help`

#### Behavior

**Basic Usage:**
```bash
snag --version
snag -v
```
- Displays version number
- Exits with code 0
- No other operations performed

**Version Format:**
- Format: `snag version {version}` (e.g., `snag version 0.0.3`)
- Version set at build time via `-ldflags`

#### Interaction Matrix

**With All Other Flags:**

| Combination | Behavior | Rationale |
|-------------|----------|-----------|
| `--version` alone | Display version, exit 0 | Standard version mode |
| `--version <url>` | Display version, exit 0 | Version takes priority, URL ignored |
| `--version --help` | Display help, exit 0 | **Help takes priority** |
| `--help --version` | Display help, exit 0 | **Help takes priority** |
| `--version` + any flags | Display version, exit 0 | Version ignores all other flags (except --help) |

**Priority Rules:**
1. `--help` detected → Display help (higher priority)
2. Otherwise, `--version` detected → Display version
3. Ignore all other flags
4. Exit with code 0

#### Examples

**Valid (Display version and exit):**
```bash
snag --version                                      # Basic version
snag -v                                             # Short form
snag --version https://example.com                  # Version (URL ignored)
snag --version -o file.md --format pdf              # Version (everything ignored)
```

**Help Takes Priority:**
```bash
snag --version --help                               # Shows HELP (not version)
snag --help --version                               # Shows HELP (not version)
```

**No Invalid Combinations:**
- Version flag ignores all other input (except --help)
- Always succeeds (exit 0)

#### Implementation Details

**Location:**
- Flag definition: Built into `github.com/urfave/cli/v2` framework
- Version set in `main.go:26` via `app.Version`

**Processing:**
1. CLI framework checks for `--help` first
2. If no help, checks for `--version` or `-v`
3. If present, displays version string
4. Exits with code 0
5. No custom validation code runs

**Version String:**
- Default: `"dev"` (development builds)
- Release: Set via build flag `-ldflags "-X main.version=0.0.3"`
- Format: Controlled by CLI framework

---

### `--close-tab` / `-c` ✅

**Status:** Complete (2025-10-23)

#### Validation Rules

**Boolean Flag:**
- No value required (presence = enabled)
- No validation errors possible

#### Behavior

**Default Behavior (No Flag):**
- Headless mode: Tabs closed automatically after fetch
- Visible browser: Tabs remain open after fetch

**With `--close-tab`:**
- Visible browser: Close tab after fetch (primary use case)
- Headless browser: Warning issued ("--close-tab has no effect in headless mode (tabs close automatically)"), proceeds normally
- Last tab: Closes tab AND browser with message: "Closing last tab, browser will close"
- Close failure: Warning issued but fetch considered successful (content already retrieved)

**Browser Close Behavior:**
- Closing the last tab will also close the browser
- Message displayed when this occurs
- Browser processes cleanly terminated

#### Interaction Matrix

**Content Source Interactions:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--close-tab` + single `<url>` | Works normally | Open tab, fetch, close tab (or warn if headless) |
| `--close-tab` + multiple `<url>` | Works normally | Close each tab after fetching |
| `--close-tab` + `--url-file` | Works normally | Close each tab after fetching |
| `--close-tab` + `--tab` | Works normally | Fetch from existing tab, then close it; if last tab, close browser |
| `--close-tab` + `--all-tabs` | Works normally | Fetch from all tabs, close all tabs, close browser |
| `--close-tab` + `--list-tabs` | **Error** | `"Cannot use --close-tab with --list-tabs. --list-tabs is standalone"` |
| `--close-tab` + `--open-browser` (no URL) | **Warning** | `"Warning: --close-tab has no effect with --open-browser (no content to close)"`, browser stays open |
| `--close-tab` + `--open-browser` + URL | Works | Open browser, fetch URL in tab, close tab/browser |

**Browser Mode Interactions:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--close-tab` + `--force-headless` | **Warning** | `"Warning: --close-tab has no effect in headless mode (tabs close automatically)"`, proceeds normally |

**Output Control (All work normally):**

| Combination | Behavior |
|-------------|----------|
| `--close-tab` + `--output` | Works - fetch content, save to file, close tab |
| `--close-tab` + `--output-dir` | Works - fetch content, save to directory, close tabs |
| `--close-tab` + `--format` (any) | Works - fetch in any format, close tab |

**Other Compatible Flags (All work normally):**

| Combination | Behavior |
|-------------|----------|
| `--close-tab` + `--wait-for` | Works - wait for selector, fetch, close tab |
| `--close-tab` + `--timeout` | Works - apply timeout to fetch, then close tab |
| `--close-tab` + `--port` | Works - connect to specific port, close tabs after fetch |
| `--close-tab` + `--user-agent` | Works - set user agent for new tabs, close after fetch |
| `--close-tab` + `--verbose`/`--quiet`/`--debug` | Works - logging applies to close operations |

#### Examples

**Valid:**
```bash
snag https://example.com --close-tab              # Visible: fetch and close tab
snag https://example.com --close-tab              # Headless: warn, close anyway
snag url1 url2 url3 --close-tab                   # Close each tab after fetch
snag --url-file urls.txt --close-tab              # Close each tab after fetch
snag --tab 1 --close-tab                          # Fetch from existing tab, close it
snag --tab "dashboard" --close-tab -o report.md   # Fetch, save, close tab
snag --all-tabs --close-tab -d ./output           # Fetch all, save all, close all, close browser
```

**Invalid:**
```bash
snag --list-tabs --close-tab                      # ERROR: list-tabs standalone
```

**With Warnings:**
```bash
snag --close-tab --open-browser                   # ⚠️ Warning: no effect (no content to close)
snag --force-headless https://example.com --close-tab  # ⚠️ Warning: headless closes anyway
```

**Special Cases:**
```bash
snag --tab 1 --close-tab                          # Last tab → closes browser too
snag --all-tabs --close-tab                       # Closes all tabs → closes browser
```

#### Implementation Details

**Location:**
- Flag definition: `main.go` (in CLI flag definitions)
- Close logic: `browser.go` (tab/browser management)

**How it works:**
1. After content fetch completes successfully
2. If `--close-tab` enabled and visible browser mode:
   - Attempt to close the tab via CDP
   - If last tab, message user and close browser
3. If `--close-tab` enabled and headless mode:
   - Warning message (tabs close automatically anyway)
4. If close operation fails:
   - Warning logged but operation considered successful
   - Content was already fetched before close attempted

**Error Handling:**
- Tab close failures don't fail the overall operation
- Content retrieval success is independent of tab close success
- Browser close failures logged but ignored (process cleanup)

**Parallel Processing Note:**
- Multiple URL processing strategy (sequential vs parallel) is tracked in TODO
- Affects whether tabs are closed one-by-one or in batch
- Current behavior: Sequential processing

---

### `--force-headless` ✅

**Status:** Complete (2025-10-23)

#### Validation Rules

**Boolean Flag:**
- No value required (presence = enabled)
- No validation errors possible

**Multiple Flags:**
- Multiple `--force-headless` flags → **Error**: `"Only one --force-headless flag allowed"`

#### Behavior

**Primary Purpose:**
- Override auto-detection to force launching a headless browser
- Useful for automation that requires consistent headless behavior

**When No Existing Browser:**
- Flag is **silently ignored** (headless is already default behavior)
- Browser launches in headless mode on default or specified port

**When Existing Browser Running:**
- Default port (9222): Connection attempt fails with port conflict error
- Custom port via `--port`: Works normally (launches new headless browser on custom port, ignoring existing browser)

**Browser Mode Conflicts:**
- Multiple `--force-headless`: **Error** - `"Only one --force-headless flag allowed"`
- With `--open-browser`: **Error** - `"Cannot use both --force-headless and --open-browser (conflicting modes)"`

#### Interaction Matrix

**Content Source Interactions:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--force-headless` + single `<url>` | **Silently ignore** flag | Headless is default, not needed |
| `--force-headless` + multiple `<url>` | **Silently ignore** flag | Headless is default, not needed |
| `--force-headless` + `--url-file` | **Silently ignore** flag | Headless is default, not needed |
| `--force-headless` + `--tab` | **Error** | `"Cannot use --force-headless with --tab (--tab requires existing browser connection)"` |
| `--force-headless` + `--all-tabs` | **Error** | `"Cannot use --force-headless with --all-tabs (--all-tabs requires existing browser connection)"` |
| `--force-headless` + `--list-tabs` | **Error** | `"Cannot use --force-headless with --list-tabs (--list-tabs requires existing browser connection)"` |

**Rationale for Tab Errors:**
- `--force-headless` implies launching a new browser
- Tab operations (`--tab`, `--all-tabs`, `--list-tabs`) require existing browser with tabs
- These are fundamentally incompatible operations

**Browser Mode Interactions:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--force-headless` + `--open-browser` | **Error** | Conflicting modes (open-browser implies visible) |
| `--force-headless` + `--user-data-dir` | Works normally | Launch headless with custom profile |

**Other Flag Interactions:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--force-headless` + `--close-tab` | **Warning** | `"Warning: --close-tab has no effect in headless mode (tabs close automatically)"` |
| `--force-headless` + `--port` | Works normally | Launch headless on specified port |
| `--force-headless` + `--output` / `--output-dir` | Works normally | Output control unaffected by browser mode |
| `--force-headless` + `--format` (any) | Works normally | Format conversion unaffected by browser mode |
| `--force-headless` + `--timeout` | Works normally | Navigation timeout applies |
| `--force-headless` + `--wait-for` | Works normally | Selector wait applies |
| `--force-headless` + `--user-agent` | Works normally | User agent set for headless browser |
| `--force-headless` + `--verbose`/`--quiet`/`--debug` | Works normally | Logging levels apply |

#### Examples

**Valid:**
```bash
# Force headless when browser might be open (silently ignored if none open)
snag --force-headless https://example.com

# Force headless with custom port (avoids conflict with existing browser)
snag --force-headless --port 9223 https://example.com

# Force headless with custom profile
snag --force-headless --user-data-dir /tmp/snag-profile https://example.com

# Force headless with output options
snag --force-headless https://example.com -o output.md
snag --force-headless https://example.com --format pdf
```

**Invalid (Errors):**
```bash
# ERROR: Conflicting modes
snag --force-headless --open-browser

# ERROR: Tab operations require existing browser
snag --force-headless --tab 1
snag --force-headless --all-tabs
snag --force-headless --list-tabs

# ERROR: Multiple flags
snag --force-headless --force-headless https://example.com
```

**With Warnings:**
```bash
# ⚠️ Warning: redundant in headless mode
snag --force-headless --close-tab https://example.com
```

**Silently Ignored (Not Needed):**
```bash
# Headless is default behavior when no browser open
snag --force-headless https://example.com          # Flag ignored (no browser)
snag --force-headless url1 url2                    # Flag ignored (no browser)
snag --force-headless --url-file urls.txt          # Flag ignored (no browser)
```

#### Implementation Details

**Location:**
- Flag definition: `main.go` (in CLI flag definitions)
- Browser launch logic: `browser.go` (browser mode detection and launch)

**How it works:**
1. Check if `--force-headless` is set
2. If set with conflicting flags (`--open-browser`, tab operations) → Error
3. If set with `--close-tab` → Warning (redundant)
4. If no existing browser running → Silently ignore (default is headless)
5. If existing browser on default port → Let connection fail (port conflict)
6. If existing browser + custom `--port` → Launch new headless on custom port

**Error Messages:**
- Tab operation conflicts: `"Cannot use --force-headless with --tab (--tab requires existing browser connection)"`
- Multiple flags: `"Only one --force-headless flag allowed"`

**Warning Messages:**
- With `--close-tab`: `"Warning: --close-tab has no effect in headless mode (tabs close automatically)"`

---

## All Arguments and Flags

### Positional Arguments

| Argument | Type   | Description  | Current         | Planned                    |
| -------- | ------ | ------------ | --------------- | -------------------------- |
| `<url>`  | String | URL to fetch | Single URL only | Multiple URLs supported 🚧 |

### Output Control Flags

| Flag           | Aliases | Type   | Default | Description                                       |
| -------------- | ------- | ------ | ------- | ------------------------------------------------- |
| `--output`     | `-o`    | String | -       | Save to specific file path                        |
| `--output-dir` | `-d`    | String | -       | Save with auto-generated name to directory        |
| `--format`     | `-f`    | String | `md`    | Output format: `md`, `html`, `text`, `pdf`, `png` |

### Page Loading Flags

| Flag         | Aliases | Type   | Default | Description                             |
| ------------ | ------- | ------ | ------- | --------------------------------------- |
| `--timeout`  | -       | Int    | `30`    | Page load timeout in seconds            |
| `--wait-for` | `-w`    | String | -       | Wait for CSS selector before extraction |

### Browser Control Flags

| Flag               | Aliases | Type   | Default | Description                                |
| ------------------ | ------- | ------ | ------- | ------------------------------------------ |
| `--port`           | `-p`    | Int    | `9222`  | Chrome remote debugging port               |
| `--close-tab`      | `-c`    | Bool   | `false` | Close browser tab after fetching           |
| `--force-headless` | -       | Bool   | `false` | Force headless mode                        |
| `--open-browser`   | `-b`    | Bool   | `false` | Open browser in visible state              |
| `--list-tabs`      | `-l`    | Bool   | `false` | List all open tabs                         |
| `--tab`            | `-t`    | String | -       | Fetch from existing tab (index or pattern) |
| `--all-tabs`       | `-a`    | Bool   | `false` | Process all open tabs                      |

### URL Input Flags (Planned 🚧)

| Flag         | Aliases | Type   | Default | Description                           |
| ------------ | ------- | ------ | ------- | ------------------------------------- |
| `--url-file` | `-f`    | String | -       | Read URLs from file (one per line) 🚧 |

### Logging Flags

| Flag        | Aliases | Type | Default | Description                |
| ----------- | ------- | ---- | ------- | -------------------------- |
| `--verbose` | -       | Bool | `false` | Enable verbose logging     |
| `--quiet`   | `-q`    | Bool | `false` | Suppress all except errors |
| `--debug`   | -       | Bool | `false` | Enable debug output        |

### Request Control Flags

| Flag           | Aliases | Type   | Default | Description              |
| -------------- | ------- | ------ | ------- | ------------------------ |
| `--user-agent` | -       | String | -       | Custom user agent string |

---

## Quick Reference Matrix

### Mode Selection (Mutually Exclusive Groups)

These determine the primary operation mode:

| Combination                          | Behavior                          | Status             |
| ------------------------------------ | --------------------------------- | ------------------ |
| No flags, no URL                     | ❌ ERROR: URL required            | ✅ Current         |
| `<url>` only                         | Fetch URL to stdout               | ✅ Current         |
| `--list-tabs` only                   | List tabs, exit                   | ✅ Current         |
| `--list-tabs` + anything else        | ❌ ERROR: standalone only         | ✅ Current         |
| `--tab <pattern>`                    | Fetch from existing tab           | ✅ Current         |
| `--tab` + `<url>`                    | ❌ ERROR: cannot mix              | ✅ Current         |
| `--all-tabs`                         | Fetch all tabs to files           | ✅ Current         |
| `--all-tabs` + `<url>`               | ❌ ERROR: cannot mix              | ✅ Current         |
| `--open-browser` only                | Open browser, keep open, no fetch | ✅ Current         |
| `--open-browser` + `<url>`           | Open browser AND fetch content    | ✅ Current         |
| `--open-browser` + `<url>` (planned) | 🚧 Open in tab, NO fetch          | 🚧 Breaking change |
| `<url> <url> <url>`                  | 🚧 Batch fetch multiple URLs      | 🚧 Planned         |
| `--url-file urls.txt`                | 🚧 Fetch URLs from file           | 🚧 Planned         |

### Output Destination (Mutually Exclusive)

| Combination                        | Behavior                                | Status     |
| ---------------------------------- | --------------------------------------- | ---------- |
| No output flags                    | Content to stdout                       | ✅ Current |
| `-o file.md`                       | Content to specific file                | ✅ Current |
| `-d ./dir`                         | Content to auto-generated file in dir   | ✅ Current |
| `-o` + `-d`                        | ❌ ERROR: cannot use both               | ✅ Current |
| Multiple URLs + `-o`               | 🚧 ❌ ERROR: use `-d` instead           | 🚧 Planned |
| Multiple URLs + `-d`               | 🚧 ✅ Each URL gets auto-generated name | 🚧 Planned |
| Multiple URLs, no output flags     | 🚧 ✅ Auto-save to current dir          | 🚧 Planned |
| Binary format (PDF/PNG), no output | Auto-generates filename in current dir  | ✅ Current |

### Browser Mode (Mutually Exclusive)

| Combination                            | Behavior                       | Status     |
| -------------------------------------- | ------------------------------ | ---------- |
| No browser flags                       | Auto-detect or launch headless | ✅ Current |
| `--force-headless`                     | Always headless                | ✅ Current |
| `--open-browser`                       | Open visible browser           | ✅ Current |
| `--open-browser` + `--force-headless`  | ⚠️ UNDEFINED                   | ❓ Unknown |

### Logging Level (Last Flag Wins) ✅

| Combination             | Effective Level         | Status     |
| ----------------------- | ----------------------- | ---------- |
| No logging flags        | Normal                  | ✅ Current |
| `--quiet`               | Quiet                   | ✅ Current |
| `--verbose`             | Verbose                 | ✅ Current |
| `--debug`               | Debug                   | ✅ Current |
| `--quiet` + `--verbose` | Verbose (last wins)     | ✅ Defined |
| `--debug` + `--verbose` | Verbose (last wins)     | ✅ Defined |
| `--quiet` + `--debug`   | Debug (last wins)       | ✅ Defined |
| `--verbose` + `--quiet` | Quiet (last wins)       | ✅ Defined |

---

## Mutually Exclusive Combinations

### Group 1: Operation Mode

**Only ONE of these can be active:**

1. Fetch URL: `<url>`
2. Fetch multiple URLs: `<url> <url>` or `--url-file` 🚧
3. Fetch from tab: `--tab <pattern>`
4. Fetch all tabs: `--all-tabs`
5. List tabs: `--list-tabs`
6. Open browser only: `--open-browser` (no URL)

**Conflict Matrix:**

|                             | URL | Multi-URL 🚧 | --tab | --all-tabs | --list-tabs | --open-browser (no URL) |
| --------------------------- | --- | ------------ | ----- | ---------- | ----------- | ----------------------- |
| **URL**                     | ✅  | 🚧 N/A       | ❌    | ❌         | ❌          | N/A                     |
| **Multi-URL** 🚧            | N/A | ✅           | 🚧 ❌ | 🚧 ❌      | 🚧 ❌       | N/A                     |
| **--tab**                   | ❌  | 🚧 ❌        | ✅    | ❌         | ❌          | N/A                     |
| **--all-tabs**              | ❌  | 🚧 ❌        | ❌    | ✅         | ❌          | N/A                     |
| **--list-tabs**             | ❌  | 🚧 ❌        | ❌    | ❌         | ✅          | N/A                     |
| **--open-browser (no URL)** | N/A | N/A          | N/A   | N/A        | N/A         | ✅                      |

**Error Messages:**

- `--tab` + URL: `"Cannot use --tab with URL argument. Use either --tab to fetch from existing tab OR provide URL to fetch new page"`
- `--all-tabs` + URL: `"Cannot use --all-tabs with URL argument. Use --all-tabs alone to process all existing tabs"`
- `--list-tabs` + URL: 🚧 `"Cannot use --list-tabs with URL argument. --list-tabs is standalone"`
- `--list-tabs` + `--tab`: 🚧 `"Cannot use --list-tabs with --tab. --list-tabs is standalone"`
- `--list-tabs` + `--all-tabs`: 🚧 `"Cannot use --list-tabs with --all-tabs. --list-tabs is standalone"`

### Group 2: Output Destination

**Only ONE of these can be active:**

1. Stdout (default, no flags)
2. Specific file: `-o file.md`
3. Auto-generated in directory: `-d ./dir`

**Conflict:**

- `-o` + `-d`: ❌ ERROR: `"Cannot use --output and --output-dir together"`

### Group 3: Browser Mode

**Only ONE of these can be active:**

1. Auto-detect (default, no flags)
2. Force headless: `--force-headless`

---

## Mode-Based Behavior

### Mode 1: Fetch Single URL (Current)

**Invocation:** `snag <url> [flags]`

**Compatible Flags:**

- ✅ `-o, -d` - Output control
- ✅ `--format` - Format selection
- ✅ `--timeout` - Load timeout
- ✅ `--wait-for` - Wait for selector
- ✅ `--port` - Remote debugging port
- ✅ `--close-tab` - Close after fetch
- ✅ `--force-headless` - Browser mode
- ✅ `--open-browser` - Open in visible browser
- ✅ `--user-agent` - Custom UA
- ✅ Logging flags

**Incompatible Flags:**

- ❌ `--tab` - Conflicts with URL
- ❌ `--all-tabs` - Conflicts with URL
- ❌ `--list-tabs` - Standalone only

**Output Behavior:**

- No `-o` or `-d`: → stdout (unless PDF/PNG)
- `-o file.md`: → file
- `-d ./dir`: → auto-generated filename in dir
- PDF/PNG without output flag: → auto-generated filename in current dir

### Mode 2: Fetch Multiple URLs (Planned 🚧)

**Invocation:** `snag <url1> <url2> <url3>` or `snag --url-file urls.txt [<url4> ...]`

**Compatible Flags:**

- ✅ `-d` - Output directory (default: current dir)
- ✅ `--format` - Applied to all URLs
- ✅ `--timeout` - Applied to each URL
- ✅ `--wait-for` - Applied to each page
- ✅ `--port` - Remote debugging port
- ✅ `--force-headless` - Browser mode
- ✅ `--open-browser` - 🚧 Opens all URLs in tabs, NO fetch
- ✅ `--user-agent` - Applied to all
- ✅ Logging flags

**Incompatible Flags:**

- ❌ `-o` - Ambiguous for multiple outputs
- ❌ `--close-tab` - Ambiguous for batch
- ❌ `--tab` - Conflicts with URLs
- ❌ `--all-tabs` - Conflicts with URLs
- ❌ `--list-tabs` - Standalone only

**Output Behavior:**

- Always saves to files (never stdout)
- No `-d`: → auto-generated names in current dir (`.`)
- `-d ./dir`: → auto-generated names in specified dir

**Error Behavior:**

- Continue-on-error (process all URLs)
- Summary: "X succeeded, Y failed"
- Exit 0 if all succeed, exit 1 if any fail

### Mode 3: Fetch from Tab

**Invocation:** `snag --tab <pattern>`

**Compatible Flags:**

- ✅ `-o, -d` - Output control
- ✅ `--format` - Format selection
- ✅ `--timeout` - Applies to `--wait-for` if present (warns if no --wait-for)
- ✅ `--wait-for` - Wait for selector (supports automation with persistent browser)
- ✅ `--port` - Remote debugging port
- ✅ `--user-agent` - Ignored (tab already open with its own user agent)
- ✅ `--close-tab` - Honored (closes the tab after fetching)
- ✅ Logging flags

**Incompatible Flags:**

- ❌ `<url>` - Conflicts with tab (mutually exclusive content sources)
- ❌ `--all-tabs` - Use one or the other (mutually exclusive)
- ❌ `--list-tabs` - Standalone only
- ❌ `--open-browser` - ⚠️ UNDEFINED
- ❌ `--force-headless` - Error (tab requires visible browser)

**Special Behavior:**

- Requires existing **visible browser** with remote debugging enabled
- If no browser running → Error: `"No browser running. Open browser first: snag --open-browser"`
- Tab pattern: index (1-based), exact URL, substring, or regex
- Tab remains open after fetch unless `--close-tab` specified

### Mode 4: Fetch All Tabs

**Invocation:** `snag --all-tabs`

**Compatible Flags:**

- ✅ `-d` - Output directory (REQUIRED or defaults to current dir)
- ✅ `--format` - Applied to all tabs
- ✅ `--timeout` - Applies to `--wait-for` if present (warns if no --wait-for)
- ✅ `--wait-for` - Wait for same selector in each tab before fetching
- ✅ `--port` - Remote debugging port
- ✅ Logging flags

**Incompatible Flags:**

- ❌ `<url>` - Conflicts with all-tabs (mutually exclusive content sources)
- ❌ `-o` - Multiple outputs (use `-d`)
- ❌ `--tab` - Use one or the other (mutually exclusive)
- ❌ `--list-tabs` - Standalone only
- ❌ `--open-browser` - ⚠️ UNDEFINED
- ❌ `--close-tab` - Error (ambiguous for batch operations)
- ❌ `--force-headless` - Error (tabs require visible browser)
- ❌ `--user-agent` - Ignored (tabs already open with their own user agents)

**Special Behavior:**

- Requires existing **visible browser** with remote debugging enabled
- If no browser running → Error: `"No browser running. Open browser first: snag --open-browser"`
- Requires `-d` or defaults to current directory
- All tabs get auto-generated filenames
- Continue-on-error (process all tabs)
- Summary: "X succeeded, Y failed"

### Mode 5: List Tabs

**Invocation:** `snag --list-tabs`

**Compatible Flags:**

- ✅ `--port` - Remote debugging port
- ✅ Logging flags

**Incompatible Flags:**

- ❌ Everything else (standalone mode)

**Special Behavior:**

- Lists tabs to stdout
- Exits after listing
- No content fetching

### Mode 6: Open Browser Only

**Invocation:** `snag --open-browser` (no URL)

**Compatible Flags:**

- ✅ `--port` - Remote debugging port
- ✅ Logging flags

**Incompatible Flags:**

- ❌ Most flags are irrelevant (no fetching)

**Special Behavior:**

- Opens visible browser
- Keeps browser open
- Exits without fetching
- User can manually interact

### Mode 7: Open Browser with URL (Current)

**Invocation:** `snag --open-browser <url>`

**Current Behavior:** Opens browser AND fetches content

**Compatible Flags:** Same as Mode 1

**Planned Change 🚧:** This will become "Open URL in tab, NO fetch" (breaking change)

---

## Output Routing Rules

### Stdout vs File Output

| Scenario                      | Output Destination                     | Notes                       |
| ----------------------------- | -------------------------------------- | --------------------------- |
| Single URL, no flags          | stdout                                 | Default behavior            |
| Single URL, `-o file.md`      | `file.md`                              | Specified file              |
| Single URL, `-d ./dir`        | `./dir/{auto-generated}.md`            | Auto-generated name         |
| Single URL, PDF/PNG, no flags | `./{auto-generated}.pdf`               | Binary formats never stdout |
| Multiple URLs, no flags       | 🚧 `./{auto-generated}.md` (each)      | Batch auto-save             |
| Multiple URLs, `-d ./dir`     | 🚧 `./dir/{auto-generated}.md` (each)  | Custom directory            |
| `--tab`, no flags             | stdout                                 | Same as single URL          |
| `--tab`, `-o file.md`         | `file.md`                              | Specified file              |
| `--all-tabs`                  | `-d` or `./{auto-generated}.md` (each) | Always files                |
| `--list-tabs`                 | stdout (tab list only)                 | Informational output        |

### Filename Generation Format

**Pattern:** `yyyy-mm-dd-hhmmss-{title-slug}.{ext}`

**Examples:**

- `2025-10-22-124752-example-domain.md`
- `2025-10-22-124753-go-programming-language.html`
- `2025-10-22-124754-github-grantcarthew-snag.pdf`

**Extension Mapping:**

- `md` → `.md`
- `html` → `.html`
- `text` → `.txt`
- `pdf` → `.pdf`
- `png` → `.png`

**Conflict Resolution:**

- If file exists: append `-1`, `-2`, etc.
- Example: `file.md` → `file-1.md` → `file-2.md`

---

## Special Cases and Edge Cases

### Case 1: Binary Formats (PDF, PNG) Without Output Flag

**Behavior:** Auto-generate filename in current directory

```bash
snag https://example.com --format pdf
# Creates: ./2025-10-22-124752-example-domain.pdf
# Does NOT output to stdout (would corrupt terminal)
```

**Implementation:** `handlers.go:118-133`

### Case 2: --open-browser Behavior Change (Planned 🚧)

**Current:**

```bash
snag --open-browser https://example.com
# Opens browser AND outputs content to stdout
```

**Planned 🚧:**

```bash
snag --open-browser https://example.com
# ONLY opens URL in browser tab, NO content output
# To fetch: snag --tab 1
```

**Rationale:** Consistency with multiple URL behavior

### Case 3: Tab Features Require Running Browser

**All tab operations require existing browser:**

```bash
snag --list-tabs        # Connects to existing browser
snag --tab 1            # Connects to existing browser
snag --all-tabs         # Connects to existing browser
```

**If no browser running:** `ErrNoBrowserRunning`

**Error Message:** `"No browser instance running with remote debugging. Start Chrome with --remote-debugging-port=9222 or run: snag --open-browser"`

### Case 4: --close-tab with Tab Features

**Question:** What happens with `snag --tab 1 --close-tab`?

**Decision:** **Allow** - Close the tab after fetching (honor user's explicit request)

**Rationale:**
- User explicitly requested the tab to be closed
- Clear intent to clean up after fetching
- Works with `--tab` (single tab)
- Errors with `--all-tabs` (ambiguous for batch operations)

### Case 5: Browser Mode Flags with Tab Features

**Question:** What happens with `snag --tab 1 --force-headless`?

**Current Behavior:** ⚠️ UNDEFINED

**Decision:**
- `--force-headless` → **Error** (tabs require visible browser)

**Rationale:**
- Tabs require visible browser with remote debugging
- `--force-headless` conflicts with this requirement → Error

### Case 6: --user-agent with Tab Features

**Question:** What happens with `snag --tab 1 --user-agent "Custom"`?

**Decision:** **Ignore** - Tab already loaded with its own user agent

**Rationale:**
- Tab already open in browser with established user agent
- Cannot change user agent for existing page
- Silently ignore rather than error (flag has no effect but doesn't break operation)
- Applies to both `--tab` and `--all-tabs`

### Case 7: Multiple Logging Flags

**Question:** What happens with `--quiet --verbose`?

**Current Behavior:** ⚠️ UNDEFINED

**Recommendation:** Priority order: `--debug` > `--verbose` > `--quiet` > normal

**Implementation:** First match wins (main.go:181-187)

### Case 8: --all-tabs with -o

**Question:** What happens with `snag --all-tabs -o output.md`?

**Current Behavior:** ⚠️ UNDEFINED (probably allowed but wrong)

**Expected:** Should ERROR with "Use --output-dir for multiple outputs"

**Status:** Needs validation

### Case 9: Zero URLs with --url-file

**Question:** What if URL file has no valid URLs?

**Planned Behavior 🚧:**

```bash
snag --url-file empty.txt
# ERROR: "No valid URLs found in file"
```

### Case 10: --open-browser + --force-headless

**Question:** Conflicting browser modes - which wins?

**Current Behavior:** ⚠️ UNDEFINED

**Logical Behavior:** Should ERROR (conflicting intent)

**Recommendation:** Add validation for this conflict

---

## Validation Order

**Current implementation order (main.go:178-316):**

1. Initialize logger (`--quiet`, `--verbose`, `--debug`)
2. Handle `--open-browser` without URL (exit early)
3. Handle `--list-tabs` (exit early)
4. Handle `--all-tabs` (check for URL conflict, exit early)
5. Handle `--tab` (check for URL conflict, exit early)
6. Validate URL argument required
7. Validate URL format
8. Validate `-o` + `-d` conflict
10. Validate format
11. Validate timeout
12. Validate port
13. Validate output path (if `-o`)
14. Validate output directory (if `-d`)
15. Execute fetch operation

**Planned validation additions 🚧:**

- Check `--url-file` + URLs (allowed)
- Check multiple URLs + `-o` (error)
- Check multiple URLs + `--close-tab` (error)
- Check `--list-tabs` + any tab feature (error)
- Check `--open-browser` + `--force-headless` (error)

---

## Undefined Behaviors

These combinations need clarification and implementation decisions:

### Priority 1: Should Error

| Combination                       | Current      | Recommendation                 |
| --------------------------------- | ------------ | ------------------------------ |
| `--all-tabs -o file.md`           | ⚠️ Undefined | ❌ ERROR: "Use --output-dir"   |
| `--tab <pattern> --all-tabs`      | ⚠️ Undefined | ❌ ERROR: Mutually exclusive   |
| `--list-tabs --tab 1`             | ⚠️ Undefined | ❌ ERROR: list-tabs standalone |
| `--list-tabs --all-tabs`          | ⚠️ Undefined | ❌ ERROR: list-tabs standalone |
| `--open-browser --force-headless` | ⚠️ Undefined | ❌ ERROR: Conflicting modes    |
| `--tab --force-headless`          | ⚠️ Undefined | ❌ ERROR: Tabs require visible browser |
| `--all-tabs --force-headless`     | ⚠️ Undefined | ❌ ERROR: Tabs require visible browser |

### Priority 2: Should Be Defined

| Combination                     | Current      | Recommendation                     |
| ------------------------------- | ------------ | ---------------------------------- |
| `--tab 1 --close-tab`           | ⚠️ Undefined | ✅ Allow: Close the tab            |
| `--tab 1 --user-agent "Custom"` | ⚠️ Undefined | ✅ Ignore: Tab already open        |
| `--all-tabs --close-tab`        | ⚠️ Undefined | ❌ ERROR: Ambiguous                |
| `--all-tabs --user-agent`       | ⚠️ Undefined | ✅ Ignore: Tabs already open       |

### Priority 3: Edge Cases

| Combination                                | Current      | Recommendation           |
| ------------------------------------------ | ------------ | ------------------------ |
| `--url-file` pointing to non-existent file | 🚧 N/A       | ❌ ERROR: File not found |
| `--url-file` with all invalid URLs         | 🚧 N/A       | ❌ ERROR: No valid URLs  |
| Multiple URLs with all failures            | 🚧 N/A       | Exit 1, summary shown    |
| `<url>` that redirects                     | ✅ Works     | ✅ Follow redirects      |
| `file:///path` URL                         | ✅ Supported | ✅ Load local file       |

---

## Compatibility Table: Complete Flag Matrix

**Legend:**

- ✅ Compatible
- ❌ Error (mutually exclusive)
- ⚠️ Undefined (needs specification)
- 🚧 Planned
- `-` Not applicable
- `?` Unknown behavior

### Output Flags

|                | -o  | -d  | --format | --timeout | --wait-for |
| -------------- | --- | --- | -------- | --------- | ---------- |
| **-o**         | -   | ❌  | ✅       | ✅        | ✅         |
| **-d**         | ❌  | -   | ✅       | ✅        | ✅         |
| **--format**   | ✅  | ✅  | -        | ✅        | ✅         |
| **--timeout**  | ✅  | ✅  | ✅       | -         | ✅         |
| **--wait-for** | ✅  | ✅  | ✅       | ✅        | -          |

### Browser Control Flags

|                      | --port | --close-tab | --force-headless | --open-browser |
| -------------------- | ------ | ----------- | ---------------- | -------------- |
| **--port**           | -      | ✅          | ✅               | ✅             |
| **--close-tab**      | ✅     | -           | ✅               | ✅             |
| **--force-headless** | ✅     | ✅          | -                | ⚠️             |
| **--open-browser**   | ✅     | ✅          | ⚠️               | -              |

### Tab Feature Flags

|                 | --list-tabs | --tab | --all-tabs |
| --------------- | ----------- | ----- | ---------- |
| **--list-tabs** | -           | ⚠️ ❌ | ⚠️ ❌      |
| **--tab**       | ⚠️ ❌       | -     | ⚠️ ❌      |
| **--all-tabs**  | ⚠️ ❌       | ⚠️ ❌ | -          |

### Arguments with Tab Features

|                 | \<url\> | Multiple URLs 🚧 | --url-file 🚧 |
| --------------- | ------- | ---------------- | ------------- |
| **--list-tabs** | ❌      | 🚧 ❌            | 🚧 ❌         |
| **--tab**       | ❌      | 🚧 ❌            | 🚧 ❌         |
| **--all-tabs**  | ❌      | 🚧 ❌            | 🚧 ❌         |

### Arguments with Output Flags

|        | \<url\> | Multiple URLs 🚧 |
| ------ | ------- | ---------------- |
| **-o** | ✅      | 🚧 ❌            |
| **-d** | ✅      | 🚧 ✅            |

### Logging Flags ✅

All logging flag conflicts resolved using "last flag wins" approach (Unix standard):

|               | --verbose | --quiet | --debug |
| ------------- | --------- | ------- | ------- |
| **--verbose** | -         | ✅ Last wins | ✅ Last wins |
| **--quiet**   | ✅ Last wins | -       | ✅ Last wins |
| **--debug**   | ✅ Last wins | ✅ Last wins | -       |

---

## Implementation Checklist

### Existing Validations ✅

- [x] `-o` + `-d` → ERROR
- [x] `--tab` + URL → ERROR
- [x] `--all-tabs` + URL → ERROR
- [x] No URL (when required) → ERROR
- [x] Invalid URL format → ERROR
- [x] Invalid timeout → ERROR
- [x] Invalid port → ERROR
- [x] Invalid output path → ERROR

### Missing Validations (Current)

- [ ] `--list-tabs` + URL → Should ERROR
- [ ] `--list-tabs` + `--tab` → Should ERROR
- [ ] `--list-tabs` + `--all-tabs` → Should ERROR
- [ ] `--tab` + `--all-tabs` → Should ERROR
- [ ] `--all-tabs` + `-o` → Should ERROR
- [ ] `--open-browser` + `--force-headless` → Should ERROR
- [x] Multiple logging flags → Last flag wins (Unix standard) ✅

### Planned Validations 🚧

- [ ] Multiple URLs + `-o` → ERROR
- [ ] Multiple URLs + `--close-tab` → ERROR
- [ ] `--url-file` + `--tab` → ERROR
- [ ] `--url-file` + `--all-tabs` → ERROR
- [ ] `--url-file` + `--list-tabs` → ERROR
- [ ] `--url-file` file not found → ERROR
- [ ] `--url-file` no valid URLs → ERROR

---

## Examples by Use Case

### Basic Fetching

```bash
snag https://example.com                    # Stdout
snag https://example.com -o page.md         # To file
snag https://example.com -d ./docs          # Auto-generated name
snag https://example.com --format html      # HTML output
snag https://example.com --format pdf       # PDF (auto-saves)
```

### Batch Processing (Planned 🚧)

```bash
snag url1 url2 url3                         # Multiple URLs
snag --url-file urls.txt                    # From file
snag -f urls.txt url4 url5                  # Combined
snag -f urls.txt -d ./results               # Custom directory
```

### Tab Management

```bash
snag --list-tabs                            # List tabs
snag --tab 1                                # Fetch tab 1
snag --tab "github"                         # Pattern match
snag --all-tabs -d ./tabs                   # All tabs
```

### Browser Control

```bash
snag --open-browser                         # Open browser only
snag --open-browser https://example.com     # Open + fetch (current)
snag --force-headless https://example.com   # Force headless
```

### Advanced Options

```bash
snag https://example.com --wait-for ".content"  # Wait for selector
snag https://example.com --timeout 60           # Custom timeout
snag https://example.com --user-agent "Custom"  # Custom UA
snag https://example.com --port 9223            # Custom port
```

---

## Notes

1. **Exit Codes:**

   - `0`: Success (all operations succeeded)
   - `1`: Error (any operation failed)
   - `130`: SIGINT (Ctrl+C)
   - `143`: SIGTERM

2. **Binary Formats:**

   - PDF and PNG never output to stdout
   - Always auto-generate filename if no `-o` or `-d`

3. **Tab Features:**

   - Require existing browser with remote debugging
   - Tabs use 1-based indexing for user display
   - Tab patterns: integer, exact URL, substring, regex

4. **URL Files:**

   - One URL per line
   - Comments: `#` or `//` (full-line or inline)
   - Auto-prepend `https://` if missing
   - Blank lines ignored

5. **Conflict Resolution:**
   - Filename conflicts append `-1`, `-2`, etc.
   - Single timestamp used for batch operations

---

**End of Document**

_This document should be updated whenever new flags are added or behaviors change._
