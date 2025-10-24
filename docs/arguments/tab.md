# `--tab PATTERN` / `-t`

**Status:** Complete (2025-10-23)

#### Validation Rules

**Pattern Types:**
- Integer: Tab index (1-based, e.g., "1" = first tab)
- String: URL pattern for matching (exact/substring/regex)
- Empty/whitespace: **Error + list tabs**

**Multiple Flags:**
- Multiple `--tab` flags → **Error**: `"Only one --tab option allowed"`

**Pattern Matching Priority (Progressive Fallthrough):**
1. **Integer** → Tab index (1-based)
2. **Exact match** → Case-insensitive URL match (`strings.EqualFold`)
3. **Contains match** → Case-insensitive substring (`strings.Contains`)
4. **Regex match** → Case-insensitive regex (`(?i)` flag)
5. **No match** → Error + list tabs

**Error Messages:**
- Empty/whitespace pattern: `"Tab pattern cannot be empty"` + list available tabs
- No matching tabs: `"No tab matches pattern '{pattern}'"` + list available tabs
- Index out of range: `"Tab index {n} out of range (only {count} tabs open)"` + list available tabs
- Invalid regex: `"Invalid regex pattern: {error}"`
- Multiple flags: `"Only one --tab option allowed"`

#### Behavior

**Primary Purpose:**
- Fetch content from an existing browser tab without creating a new one
- Requires existing browser connection (won't auto-launch)
- Mutually exclusive with all other content sources

**Pattern Matching Examples:**

```bash
# By index (1-based)
snag --tab 1                                    # First tab
snag -t 3                                       # Third tab

# By exact URL (case-insensitive)
snag -t "https://github.com/grantcarthew/snag"  # Exact match
snag -t "EXAMPLE.COM"                           # Case-insensitive

# By substring/contains
snag -t "github"                                # Contains "github"
snag -t "dashboard"                             # Contains "dashboard"

# By regex pattern
snag -t "https://.*\.com"                       # Regex: https:// + anything + .com
snag -t ".*/dashboard"                          # Regex: any URL ending with /dashboard
snag -t "(github|gitlab)\.com"                  # Regex: github.com or gitlab.com
```

**Multiple Matches:**
- First matching tab wins (no error)
- Consider using more specific pattern if multiple tabs match

**No Browser Connection:**
- Error: `"No browser found. Try running 'snag --open-browser' first"`
- No special error handling (same as other browser operations)

#### Interaction Matrix

**Content Source Conflicts (All ERROR - Mutually Exclusive):**

| Combination | Behavior | Error Message |
|-------------|----------|---------------|
| `--tab` + `<url>` (single) | **Error** | `"Cannot use both --tab and URL arguments (mutually exclusive content sources)"` |
| `--tab` + `<url>` (multiple) | **Error** | `"Cannot use both --tab and URL arguments (mutually exclusive content sources)"` |
| `--tab` + `--url-file` | **Error** | `"Cannot use both --tab and --url-file (mutually exclusive content sources)"` |
| `--tab` + `--all-tabs` | **Error** | `"Cannot use both --tab and --all-tabs (mutually exclusive content sources)"` |
| `--tab` + `--list-tabs` | `--list-tabs` overrides | `--list-tabs` overrides all other options (no error) |

**Browser Mode Conflicts (All ERROR):**

| Combination | Behavior | Error Message |
|-------------|----------|---------------|
| `--tab` + `--force-headless` | **Error** | `"Cannot use --force-headless with --tab (--tab requires existing browser connection)"` |
| `--tab` + `--open-browser` | **Warning**, flag ignored | `"Warning: --tab ignored with --open-browser (no content fetching)"` |

**Output Control (All Work Normally):**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--tab` + `--output` | Works normally | Fetch from tab, save to file |
| `--tab` + `--output-dir` | Works normally | Fetch from tab, auto-generate filename |
| `--tab` + `--format` (all) | Works normally | All formats supported (md/html/text/pdf/png) |
| `--tab` + no output flag | Works normally | Output to stdout |

**Timing & Selector:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--tab` + `--timeout` (no `--wait-for`) | Works with **warning** | `"Warning: --timeout is ignored without --wait-for when using --tab"` |
| `--tab` + `--timeout` + `--wait-for` | Works normally | Timeout applies to selector wait |
| `--tab` + `--wait-for` | Works normally | Wait for selector in existing tab (automation use case) |

**Browser Configuration:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--tab` + `--port` | Works normally | Connect to browser on specific port |
| `--tab` + `--close-tab` | Works normally | Close tab after fetching (covered in Task 9) |
| `--tab` + `--user-agent` | **Warning**, ignored | `"Warning: --user-agent is ignored with --tab (cannot change existing tab's user agent)"` |
| `--tab` + `--user-data-dir` | **Warning**, ignored | `"Warning: --user-data-dir ignored when connecting to existing browser"` |

**Logging Flags (All Work Normally):**

| Combination | Behavior |
|-------------|----------|
| `--tab` + `--verbose` | Works normally - verbose logging of tab selection and fetch |
| `--tab` + `--quiet` | Works normally - suppress non-error messages |
| `--tab` + `--debug` | Works normally - debug logging of tab operations |

#### Examples

**Valid:**
```bash
# Fetch from tab by index
snag --tab 1                                    # First tab
snag -t 2 -o output.md                          # Second tab, save to file

# Fetch by URL pattern
snag -t "github.com" --format html              # Match substring
snag -t "https://example.com" -o page.md        # Exact URL match

# With selectors and timeouts (automation)
snag -t "dashboard" --wait-for ".loaded"        # Wait for selector in existing tab
snag -t 1 --wait-for ".content" --timeout 60    # Custom timeout for selector

# With output options
snag -t "docs" -o reference.md                  # Save to specific file
snag -t 2 --format pdf                          # Save as PDF (auto-generates filename)
snag -t 1 -d ./output/                          # Auto-generate filename in directory

# Close tab after fetching
snag -t 1 --close-tab                           # Fetch and close tab
```

**Invalid (Errors):**
```bash
snag --tab ""                                   # ERROR: Empty pattern + list tabs
snag --tab "   "                                # ERROR: Whitespace only + list tabs
snag --tab 99                                   # ERROR: Index out of range + list tabs
snag --tab "nonexistent"                        # ERROR: No match + list tabs
snag --tab "([invalid"                          # ERROR: Invalid regex
snag --tab 1 --tab 2                            # ERROR: Multiple flags
snag --tab 1 https://example.com                # ERROR: Mutually exclusive with URL
snag --tab 1 --url-file urls.txt                # ERROR: Mutually exclusive with --url-file
snag --tab 1 --all-tabs                         # ERROR: Mutually exclusive with --all-tabs
snag --tab 1 --force-headless                   # ERROR: Tab requires existing browser
```

**With Warnings:**
```bash
snag --tab 1 --open-browser                     # ⚠️  Warning: --tab ignored (no content fetching)
snag --tab 1 --timeout 30                       # ⚠️  Warning: --timeout is ignored without --wait-for when using --tab
snag --tab 1 --user-agent "Bot/1.0"             # ⚠️  Warning: --user-agent is ignored with --tab (cannot change existing tab's user agent)
```

**Overridden (No Error):**
```bash
snag --list-tabs --tab 1                        # --list-tabs overrides, lists all tabs
```

#### Implementation Details

**Location:**
- Flag definition: `main.go` (CLI flag definitions)
- Handler: `main.go:412-534` (`handleTabFetch()`)
- Tab selection: `browser.go:434-463` (`GetTabByIndex()`), `browser.go:473-544` (`GetTabByPattern()`)
- Pattern matching: Progressive fallthrough in `GetTabByPattern()`

**How it works:**
1. Validate pattern is not empty/whitespace
2. Check for mutually exclusive flags (URL, --url-file, --all-tabs, --open-browser, --force-headless)
3. Connect to existing browser (error if none found)
4. Parse pattern:
   - If integer: Convert to tab index (1-based → 0-based)
   - If string: Try exact match → substring → regex
5. If no match: Error and list available tabs
6. Fetch content from selected tab
7. Apply output options (--format, --output, --output-dir)
8. Close tab if `--close-tab` is set

**Pattern Matching Algorithm:**
```go
// Pseudocode
func GetTabByPattern(pattern string) (*Tab, error) {
    // 1. Try integer (tab index)
    if isInteger(pattern) {
        return GetTabByIndex(index)
    }

    // 2. Cache page.Info() for all tabs (single pass, optimization)
    tabInfos := cacheAllTabInfo()

    // 3. Try exact match (case-insensitive)
    for tab, info := range tabInfos {
        if strings.EqualFold(info.URL, pattern) {
            return tab
        }
    }

    // 4. Try substring/contains match (case-insensitive)
    for tab, info := range tabInfos {
        if strings.Contains(strings.ToLower(info.URL), strings.ToLower(pattern)) {
            return tab
        }
    }

    // 5. Try regex match (case-insensitive)
    regex := regexp.MustCompile("(?i)" + pattern)
    for tab, info := range tabInfos {
        if regex.MatchString(info.URL) {
            return tab
        }
    }

    // 6. No match found
    return nil, ErrNoTabMatch
}
```

**Performance Optimization:**
- Single-pass `page.Info()` caching (browser.go:487-507)
- Reduces network calls from 3N to N (3x improvement for 10 tabs)
- Do not modify pattern matching without preserving this optimization

**Design Note:**
- User-facing indexes are 1-based (natural for humans)
- Internal indexes are 0-based (natural for Go)
- Conversion happens in `TabInfo` struct and `GetTabByIndex()`
