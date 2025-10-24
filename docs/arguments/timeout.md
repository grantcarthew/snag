# `--timeout SECONDS`

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
- Multiple `--timeout` flags → Error: `"Only one --timeout option allowed"`

**Error Messages:**
- Negative/zero: `"Timeout must be positive"`
- Non-integer: `"Timeout must be an integer"`
- Non-numeric: `"Invalid timeout value: {value}"`
- Multiple flags: `"Only one --timeout option allowed"`

#### Behavior

**Scope:**
- Timeout applies to **page navigation** and **`--wait-for` selector waiting**
- Does not apply to format conversion or PDF/PNG generation
- Default timeout: 30 seconds (if flag not specified)

**Basic Usage:**
```bash
snag https://example.com --timeout 60
```
- Sets 60-second timeout for initial page navigation
- If page doesn't load within 60s → Error: `"Page load timeout"`
- Also applies to `--wait-for` selector waiting (if used)
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
- `--tab` (no --wait-for): `"Warning: --timeout is ignored without --wait-for when using --tab"`
- `--all-tabs` (no --wait-for): `"Warning: --timeout is ignored without --wait-for when using --all-tabs"`

**Special Operation Modes:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--timeout` + `--list-tabs` | Flag **ignored**, tabs listed | `--list-tabs` overrides all other args |
| `--timeout` + `--open-browser` (no URL) | **Warning**, flag ignored | `"Warning: --timeout ignored with --open-browser (no content fetching)"` |

**Timing-Related Interactions:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--timeout` + `--wait-for` | Works normally | Timeout applies to both navigation and selector wait |

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
| `--timeout` + `--user-data-dir` | Works normally |

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
snag --list-tabs --timeout 30                       # --timeout ignored, lists tabs from existing browser
```

**With Warnings:**
```bash
snag --tab 1 --timeout 30                           # ⚠️  Warning: --timeout is ignored without --wait-for when using --tab
snag --all-tabs --timeout 30                        # ⚠️  Warning: --timeout is ignored without --wait-for when using --all-tabs
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
- Selector wait timeout: Time for `--wait-for` selector to appear (uses same timeout value)
- Does NOT include:
  - Format conversion time (HTML → Markdown, Text)
  - PDF/PNG generation time (handled separately by Chrome)
  - Network retry delays
  - Browser launch time

**Default Behavior:**
- No `--timeout` flag → Default 30 seconds
- Configurable per-operation via flag
