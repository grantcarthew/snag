# `--url-file FILE`

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
| `--url-file` + `--list-tabs` | Flag **ignored**, tabs listed | `--list-tabs` acts like `--help`, overrides other args |

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
snag --url-file urls.txt --list-tabs               # --url-file ignored, lists tabs from existing browser
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
