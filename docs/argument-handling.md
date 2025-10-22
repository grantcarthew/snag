# Argument Handling Reference

**Purpose:** Complete specification of all argument/flag combinations and their interactions.

**Status:** Current implementation + Planned features (marked with 🚧)

**Last Updated:** 2025-10-22

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
| `<url>` + `--force-visible` | Open visible browser, navigate, **fetch content**, leave browser open | Useful for authenticated sessions |

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
snag https://example.com --force-visible           # Visible browser + fetch
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
| `--url-file` + `--force-visible` | Works normally, auto-save | Force visible, fetch all URLs |
| `--url-file` + `--force-headless --force-visible` | **Error** | Conflicting browser modes |

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
snag --url-file urls.txt --force-headless --force-visible  # ERROR: Conflicting modes
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
- ✅ `--force-visible` - Force visible browser mode
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
| `--force-visible`  | -       | Bool   | `false` | Force visible browser mode                 |
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
| `--force-visible`                      | Always visible                 | ✅ Current |
| `--force-headless` + `--force-visible` | ❌ ERROR: conflicting flags    | ✅ Current |
| `--open-browser`                       | Open visible browser           | ✅ Current |
| `--open-browser` + `--force-headless`  | ⚠️ UNDEFINED                   | ❓ Unknown |
| `--open-browser` + `--force-visible`   | ✅ Redundant but allowed       | ✅ Current |

### Logging Level (Mutually Exclusive Priority)

| Combination             | Effective Level         | Status     |
| ----------------------- | ----------------------- | ---------- |
| No logging flags        | Normal                  | ✅ Current |
| `--quiet`               | Quiet                   | ✅ Current |
| `--verbose`             | Verbose                 | ✅ Current |
| `--debug`               | Debug                   | ✅ Current |
| `--quiet` + `--verbose` | ⚠️ UNDEFINED            | ❓ Unknown |
| `--debug` + `--verbose` | Debug (higher priority) | ✅ Current |
| `--quiet` + `--debug`   | ⚠️ UNDEFINED            | ❓ Unknown |

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
3. Force visible: `--force-visible`

**Conflict:**

- `--force-headless` + `--force-visible`: ❌ ERROR: `"Conflicting flags: --force-headless and --force-visible cannot be used together"`

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
- ✅ `--force-headless, --force-visible` - Browser mode
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
- ✅ `--force-headless, --force-visible` - Browser mode
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
- ✅ `--timeout` - Wait timeout (for --wait-for)
- ✅ `--wait-for` - Wait for selector
- ✅ `--port` - Remote debugging port
- ✅ `--user-agent` - ⚠️ UNDEFINED (tab already open)
- ✅ Logging flags

**Incompatible Flags:**

- ❌ `<url>` - Conflicts with tab
- ❌ `--all-tabs` - Use one or the other
- ❌ `--list-tabs` - Standalone only
- ❌ `--open-browser` - ⚠️ UNDEFINED
- ❌ `--close-tab` - ⚠️ Tab persists (not created by snag)
- ❌ `--force-headless, --force-visible` - ⚠️ UNDEFINED (browser already running)

**Special Behavior:**

- Requires existing browser with remote debugging
- Tab pattern: index (1-based), exact URL, substring, or regex
- Tab remains open after fetch (not created by this invocation)

### Mode 4: Fetch All Tabs

**Invocation:** `snag --all-tabs`

**Compatible Flags:**

- ✅ `-d` - Output directory (REQUIRED or defaults to current dir)
- ✅ `--format` - Applied to all tabs
- ✅ `--timeout` - Applied to each tab
- ✅ `--wait-for` - Applied to each page
- ✅ `--port` - Remote debugging port
- ✅ Logging flags

**Incompatible Flags:**

- ❌ `<url>` - Conflicts with all-tabs
- ❌ `-o` - Multiple outputs (use `-d`)
- ❌ `--tab` - Use one or the other
- ❌ `--list-tabs` - Standalone only
- ❌ `--open-browser` - ⚠️ UNDEFINED
- ❌ `--close-tab` - ⚠️ Tabs persist
- ❌ `--force-headless, --force-visible` - ⚠️ UNDEFINED (browser already running)
- ❌ `--user-agent` - ⚠️ UNDEFINED (tabs already open)

**Special Behavior:**

- Requires existing browser with remote debugging
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
- ✅ `--force-visible` - Redundant but allowed
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

**Current Behavior:** ⚠️ UNDEFINED

**Possible Behaviors:**

1. ❌ Error: "Cannot close tab not created by snag"
2. ✅ Ignore flag (tab persists)
3. ✅ Close the tab (user requested it)

**Recommendation:** Option 3 - Honor user's explicit request

### Case 5: Browser Mode Flags with Tab Features

**Question:** What happens with `snag --tab 1 --force-headless`?

**Current Behavior:** ⚠️ UNDEFINED

**Rationale:** Browser already running, mode flags are irrelevant

**Recommendation:** Ignore browser mode flags when using tab features (browser already connected)

### Case 6: --user-agent with Tab Features

**Question:** What happens with `snag --tab 1 --user-agent "Custom"`?

**Current Behavior:** ⚠️ UNDEFINED

**Rationale:** Tab already open with its own user agent

**Recommendation:** Ignore `--user-agent` when using tab features (tab already loaded)

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
8. Validate `--force-headless` + `--force-visible` conflict
9. Validate `-o` + `-d` conflict
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
| `--quiet --verbose`               | ⚠️ Undefined | ❌ ERROR or priority order?    |

### Priority 2: Should Be Defined

| Combination                     | Current      | Recommendation                     |
| ------------------------------- | ------------ | ---------------------------------- |
| `--tab 1 --close-tab`           | ⚠️ Undefined | ✅ Allow: Close the tab            |
| `--tab 1 --force-headless`      | ⚠️ Undefined | ✅ Ignore: Browser already running |
| `--tab 1 --user-agent "Custom"` | ⚠️ Undefined | ✅ Ignore: Tab already open        |
| `--all-tabs --force-headless`   | ⚠️ Undefined | ✅ Ignore: Browser already running |
| `--all-tabs --close-tab`        | ⚠️ Undefined | ❌ ERROR: Ambiguous                |

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

|                      | --port | --close-tab | --force-headless | --force-visible | --open-browser |
| -------------------- | ------ | ----------- | ---------------- | --------------- | -------------- |
| **--port**           | -      | ✅          | ✅               | ✅              | ✅             |
| **--close-tab**      | ✅     | -           | ✅               | ✅              | ✅             |
| **--force-headless** | ✅     | ✅          | -                | ❌              | ⚠️             |
| **--force-visible**  | ✅     | ✅          | ❌               | -               | ✅             |
| **--open-browser**   | ✅     | ✅          | ⚠️               | ✅              | -              |

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

### Logging Flags

|               | --verbose | --quiet | --debug |
| ------------- | --------- | ------- | ------- |
| **--verbose** | -         | ⚠️      | ?       |
| **--quiet**   | ⚠️        | -       | ⚠️      |
| **--debug**   | ?         | ⚠️      | -       |

---

## Implementation Checklist

### Existing Validations ✅

- [x] `--force-headless` + `--force-visible` → ERROR
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
- [ ] Multiple logging flags → Define priority

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
snag --force-visible https://example.com    # Force visible
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
