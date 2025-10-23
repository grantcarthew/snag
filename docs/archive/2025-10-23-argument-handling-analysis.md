# Argument Handling Analysis - Design Decision Document

**Project:** Systematically analyze each snag CLI argument's behavior and interactions

**Goal:** Create comprehensive design decisions for every argument combination

**Deliverable:** Complete `docs/arguments/` directory with all behaviors defined

---

## Design Process

This document tracks a systematic design session for the snag CLI tool's argument handling behavior.

### Why This Process?

Before implementing new features or fixing edge cases, we need crystal-clear design decisions for every argument interaction. This prevents:

- Inconsistent behavior across similar scenarios
- Undocumented edge cases discovered by users
- Implementation decisions made without considering all implications
- Technical debt from "we'll figure it out later" approaches

### Strict Process

This is a **design-first, implementation-later** methodology:

1. **Question Phase**: For each argument, I will ask you design questions about:
   - What happens with invalid/wrong values?
   - How it interacts with every other argument
   - Whether combinations should error, work together, modify behavior, or be ignored

2. **Discussion Phase**: We discuss and decide together on the correct behavior

3. **Documentation Phase**: Only after your explicit permission, I will:
   - Update this PROJECT.md with progress tracking
   - Update `docs/arguments/[argument].md` with the design decisions

4. **No Implementation**: This is design-only. No code changes, only documentation.

### Question Format Rules

To enable structured responses, questions will be asked ONE CATEGORY AT A TIME with clear numbering:

```
Question 1: Invalid Values
1. Scenario description
2. Scenario description
3. Scenario description

Question 2: Key Combinations
1. Scenario description
2. Scenario description
```

You can respond with structured answers like:
- `1.1` - answer for Question 1, item 1
- `1.2` - answer for Question 1, item 2
- `2.1` - answer for Question 2, item 1

### Rules

- I will NOT make assumptions about behavior without asking
- I will NOT update documentation without your permission
- I will ask about every combination systematically
- I will ask ONE category at a time with clear numbering
- We decide together, document completely, implement later

---

## Analysis Structure

For each argument, answer:
1. **What happens if we supply wrong values?** (validation, error messages)
2. **What happens when combined with every other argument?** (compatibility matrix)
3. **Define behavior:** Error, ignore, modify behavior, work together?

---

## Tasks by Argument

### Task 1: `<url>` (Positional Argument)

**Questions to answer:**
- What happens with invalid URL values? (malformed, missing protocol, etc.)
- What happens with multiple URLs? (current vs planned)
- What happens when combined with:
  - `--url-file`
  - `--output` / `-o`
  - `--output-dir` / `-d`
  - `--format` / `-f`
  - `--timeout`
  - `--wait-for` / `-w`
  - `--port` / `-p`
  - `--close-tab` / `-c`
  - `--force-headless`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`
  - `--user-data-dir`

**Define:** Which combinations error, which work together, which modify behavior?

---

### Task 2: `--url-file FILE`

**Questions to answer:**
- What happens with wrong values? (file doesn't exist, empty file, all invalid URLs, permission denied)
- What happens when combined with:
  - `<url>`
  - Another `--url-file`
  - `--output` / `-o`
  - `--output-dir` / `-d`
  - `--format` / `-f`
  - `--timeout`
  - `--wait-for` / `-w`
  - `--port` / `-p`
  - `--close-tab` / `-c`
  - `--force-headless`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`
  - `--user-data-dir`

**Define:** Error conditions, valid combinations, output behavior

---

### Task 3: `--output FILE` / `-o`

**Questions to answer:**
- What happens with wrong values? (invalid path, permission denied, directory instead of file)
- What happens when combined with:
  - `<url>` (single)
  - `<url>` (multiple)
  - `--url-file`
  - Another `--output`
  - `--output-dir` / `-d`
  - `--format` / `-f`
  - `--timeout`
  - `--wait-for` / `-w`
  - `--port` / `-p`
  - `--close-tab` / `-c`
  - `--force-headless`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`
  - `--user-data-dir`

**Define:** Conflicts with `--output-dir`, behavior with multiple sources

---

### Task 4: `--output-dir DIRECTORY` / `-d`

**Questions to answer:**
- What happens with wrong values? (doesn't exist, not a directory, permission denied)
- What happens when combined with:
  - `<url>` (single)
  - `<url>` (multiple)
  - `--url-file`
  - `--output` / `-o`
  - Another `--output-dir`
  - `--format` / `-f`
  - `--timeout`
  - `--wait-for` / `-w`
  - `--port` / `-p`
  - `--close-tab` / `-c`
  - `--force-headless`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`
  - `--user-data-dir`

**Define:** Filename generation behavior, conflicts with `--output`

---

### Task 5: `--format FORMAT` / `-f`

**Questions to answer:**
- What happens with wrong values? (invalid format, empty string, case sensitivity)
- What happens when combined with:
  - `<url>` (single)
  - `<url>` (multiple)
  - `--url-file`
  - `--output` / `-o`
  - `--output-dir` / `-d`
  - Another `--format`
  - `--timeout`
  - `--wait-for` / `-w`
  - `--port` / `-p`
  - `--close-tab` / `-c`
  - `--force-headless`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`
  - `--user-data-dir`

**Define:** Binary format auto-save behavior, stdout vs file interaction

---

### Task 6: `--timeout SECONDS`

**Questions to answer:**
- What happens with wrong values? (negative, zero, non-integer, extremely large)
- What happens when combined with:
  - `<url>` (single)
  - `<url>` (multiple)
  - `--url-file`
  - `--output` / `-o`
  - `--output-dir` / `-d`
  - `--format` / `-f`
  - Another `--timeout`
  - `--wait-for` / `-w`
  - `--port` / `-p`
  - `--close-tab` / `-c`
  - `--force-headless`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`
  - `--user-data-dir`

**Define:** Timeout behavior with `--wait-for`, batch operations

---

### Task 7: `--wait-for SELECTOR` / `-w` ✅

**Status:** Complete (2025-10-23)

#### Invalid Values

**Empty string:**
- Behavior: **Ignored** (silently skipped, no error or warning)
- Rationale: Empty string is effectively "don't use wait-for", which is default behavior
- Many CLI tools silently ignore empty string values for optional flags

**Whitespace-only string:**
- Behavior: **Ignored** after trimming
- All string arguments should be trimmed using `strings.TrimSpace()` after reading from CLI framework
- This is standard behavior in most CLI tools (git, docker, etc.)

**Invalid CSS selector syntax:**
- Behavior: **Error at runtime** (caught by rod's `Element()` method)
- No upfront validation - selector validation happens when rod tries to use it
- Error message from rod will indicate invalid selector

**Valid selector that never appears:**
- Behavior: **Timeout error** after `--timeout` duration expires
- Error message: "Timeout waiting for selector {selector}" with suggestion to increase timeout
- This is normal/expected behavior tested in test suite

**Multiple `--wait-for` flags:**
- Behavior: **Error**
- Error message: "Only one --wait-for flag allowed"

**Extremely complex selector:**
- Behavior: **Allow** if valid CSS syntax
- User's responsibility to provide working selectors

#### Content Source Interactions

**With URL arguments:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--wait-for` + single `<url>` | Works normally | Wait for selector after navigation |
| `--wait-for` + multiple `<url>` | Works normally | Same selector applied to all URLs |
| `--wait-for` + `--url-file` | Works normally | Same selector applied to all URLs from file |

**With tab operations:**

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--wait-for` + `--tab` | **Works** | Supports automation with persistent browser - wait for selector in existing tab |
| `--wait-for` + `--all-tabs` | **Works** | Same selector applied to all tabs before fetching |
| `--wait-for` + `--list-tabs` | **Error** | List-tabs doesn't fetch content, standalone only |

**Error messages:**
- `--list-tabs`: "Cannot use --wait-for with --list-tabs (no content fetching)"

**Use case for tabs + wait-for:**
- Persistent visible browser with authenticated sessions
- Automated script runs periodically (e.g., cron job)
- Dynamic content loads asynchronously in existing tabs
- `--wait-for` ensures script waits for content before extracting
- Example: `snag --tab "dashboard" --wait-for ".data-loaded" -o daily-report.md`

#### Output Control Interactions

All output flags work normally with `--wait-for`:

| Combination | Behavior |
|-------------|----------|
| `--wait-for` + `--output` | Works normally - wait before writing to file |
| `--wait-for` + `--output-dir` | Works normally - wait before auto-saving |
| `--wait-for` + `--format` (all) | Works normally - wait before format conversion |

#### Browser Mode Interactions

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--wait-for` + `--force-headless` | Works normally | Wait in headless mode |
| `--wait-for` + `--open-browser` (no URL) | **Warning** | Flag ignored, no content to fetch |
| `--wait-for` + `--open-browser` + URL | **Warning** | Flag ignored, `--open-browser` doesn't fetch content |

**Warning message:**
- "Warning: --wait-for has no effect with --open-browser (no content fetching)"

#### Timing Interactions

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--wait-for` + `--timeout` | Works normally | `--timeout` applies to navigation; `--wait-for` has separate timeout logic |
| `--wait-for` + `--timeout` + `--tab` | Works normally | `--timeout` applies to `--wait-for` selector wait (warns if no --wait-for) |

**Current implementation:**
- Navigation timeout: Controlled by `--timeout` flag (default 30s)
- Selector wait timeout: Uses same `--timeout` duration
- Both use rod's timeout mechanism via `page.Timeout(duration)`

#### Other Flag Interactions

**Compatible flags (work normally):**
- `--port` - Remote debugging port
- `--close-tab` - Close tab after fetching (works with `--tab`)
- `--verbose` / `--quiet` / `--debug` - Logging levels
- `--user-agent` - Set user agent for new pages (ignored for existing tabs)

#### Examples

**Valid:**
```bash
snag https://example.com --wait-for ".content"              # Basic usage
snag https://example.com --wait-for "#main-content"         # ID selector
snag https://example.com --wait-for "div > .loaded"         # Complex selector
snag https://example.com -w ".content" --timeout 60         # Custom timeout
snag url1 url2 --wait-for ".content"                        # Multiple URLs
snag --url-file urls.txt --wait-for ".content"              # URL file
snag --tab 1 --wait-for ".content" -o output.md             # Existing tab (automation)
snag --all-tabs --wait-for ".content" -d ./output           # All tabs
snag https://example.com --wait-for ".content" --format pdf # With PDF output
```

**Invalid:**
```bash
snag --list-tabs --wait-for ".content"                      # ERROR: List-tabs standalone
snag --wait-for ".content" --wait-for ".other"              # ERROR: Multiple flags
```

**With Warnings:**
```bash
snag --open-browser https://example.com --wait-for ".content"  # ⚠️  No effect (no fetch)
snag --wait-for "   "                                          # Ignored after trim
snag --wait-for ""                                             # Ignored (empty)
```

#### Implementation Details

**Location:**
- Flag definition: `main.go:108-111`
- Wait logic: `fetch.go:167-193` (`waitForSelector()` helper function)
- Usage in fetch: `fetch.go:71-84`
- Usage in tab handlers: `handlers.go` (various locations)

**How it works:**
1. Takes CSS selector string value
2. If empty string after trim, silently skip (no validation error)
3. Wait for element to appear using `page.Element(selector)` with timeout
4. Wait for element to be visible using `elem.WaitVisible()`
5. Return error if selector never appears or never becomes visible
6. Timeout errors include helpful suggestion to increase `--timeout`

**Validation:**
- No upfront CSS syntax validation
- All validation happens at runtime through rod's CDP implementation
- Invalid selectors caught by rod's `Element()` method
- Timeout handled by rod's context deadline mechanism

**Timeout behavior:**
- Uses `--timeout` flag value (default 30 seconds)
- Applied to both `Element()` and `WaitVisible()` calls
- Errors include context about which operation timed out

---

### Task 8: `--port PORT` / `-p`

**Questions to answer:**
- What happens with wrong values? (negative, zero, > 65535, non-integer, in-use)
- What happens when combined with:
  - `<url>` (single)
  - `<url>` (multiple)
  - `--url-file`
  - `--output` / `-o`
  - `--output-dir` / `-d`
  - `--format` / `-f`
  - `--timeout`
  - `--wait-for` / `-w`
  - Another `--port`
  - `--close-tab` / `-c`
  - `--force-headless`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`
  - `--user-data-dir`

**Define:** Port availability checking, all operation modes

---

### Task 9: `--close-tab` / `-c` ✅

**Status:** Complete (2025-10-23)

#### Behavior

**Primary use case:**
- Visible browser mode: Close tab after fetch
- Headless mode: Warning (tabs close automatically anyway)

**Default behavior (no flag):**
- Headless: Tabs closed automatically
- Visible: Tabs remain open

**Last tab handling:**
- Closing last tab also closes browser
- Message: "Closing last tab, browser will close"

**Close failure:**
- Warning issued but fetch considered successful
- Content already retrieved before close attempted

#### Content Source Interactions

- Single/multiple URLs: Works - close each tab after fetch ✅
- `--url-file`: Works - close each tab after fetch ✅
- `--tab`: Works - close existing tab; if last, close browser ✅
- `--all-tabs`: Works - close all tabs and browser ✅
- `--list-tabs`: ❌ ERROR (standalone)
- `--open-browser` (no URL): ⚠️ Warning, ignored
- `--open-browser` + URL: Works - close tab/browser ✅

#### Browser Mode

- `--force-headless`: ⚠️ Warning (redundant), proceeds

#### All Other Flags

Output control, wait-for, timeout, port, user-agent, logging: All work normally ✅

**Note:** Parallel processing strategy for multiple URLs tracked in TODO.

---

### Task 10: `--force-headless` ✅

**Status:** Complete (2025-10-23)

#### Core Behavior & Browser Connection

**No existing browser:**
- Flag **silently ignored** (headless is default behavior)

**Existing browser on default port:**
- Connection attempt fails with port conflict error

**Existing browser + custom `--port`:**
- Works normally (launches new headless on custom port)

**Multiple flags:**
- Multiple `--force-headless` → **Error**: `"Only one --force-headless flag allowed"`

#### Browser Mode Conflicts

- `--force-headless` + `--open-browser` → **Error**: `"Cannot use both --force-headless and --open-browser (conflicting modes)"`

#### Content Source Interactions

- Single/multiple URLs, `--url-file` → **Silently ignore** (headless is default)
- `--tab` → **Error**: `"Cannot use --force-headless with --tab (--tab requires existing browser connection)"`
- `--all-tabs` → **Error**: `"Cannot use --force-headless with --all-tabs (--all-tabs requires existing browser connection)"`
- `--list-tabs` → **Error**: `"Cannot use --force-headless with --list-tabs (--list-tabs requires existing browser connection)"`

#### Other Flag Interactions

- `--close-tab` → **Warning**: `"--close-tab has no effect in headless mode (tabs close automatically)"`
- `--user-data-dir` → Works normally (launch headless with custom profile)
- All others (output, format, timeout, wait-for, port, user-agent, logging) → Work normally

---

### Task 11: `--open-browser` / `-b` ✅

**Status:** Complete (2025-10-23)

#### Core Behavior

**Primary purpose:**
- Launch persistent visible browser and exit snag (browser stays open)
- "Launch and exit" means **exit snag**, not the browser
- **Does not fetch content** - purely a browser launcher

**Multiple flags:**
- Multiple `--open-browser` → **Silently ignore** (duplicate boolean)

**Modes:**
1. Standalone (no URLs): Open browser, exit snag
2. With URLs: Open browser, navigate to URLs in tabs, exit snag (no fetch)

#### Browser Mode Conflicts

- `--open-browser` + `--force-headless` → **Error**: `"Cannot use both --force-headless and --open-browser (conflicting modes)"`
- `--open-browser` + `--list-tabs` → **Error**: `"Cannot use --list-tabs with --open-browser (--list-tabs is standalone)"`

#### Content Source Interactions

- Single/multiple `<url>`, `--url-file` → Navigate to URLs in tabs, exit (no fetch)
- `--tab`, `--all-tabs` → **Warning**: Flags ignored (no content fetching)

#### Output & Timing Flags

All **warned and ignored** (no content fetching):
- `--output`, `--output-dir`, `--format` → Warning
- `--timeout`, `--wait-for` → Warning

#### Browser Configuration

- `--port`, `--user-data-dir` → Work normally
- `--close-tab` → **Warning**: Ignored (no fetching)
- `--user-agent` (no URLs) → **Warning**: Ignored (no navigation)
- `--user-agent` + URLs → Works normally (applied during navigation)

#### Logging

- `--verbose`, `--quiet`, `--debug` → Work normally

---

### Task 12: `--list-tabs` / `-l`

**Questions to answer:**
- What happens with wrong values? (N/A - boolean flag)
- What happens when combined with:
  - `<url>` (single)
  - `<url>` (multiple)
  - `--url-file`
  - `--output` / `-o`
  - `--output-dir` / `-d`
  - `--format` / `-f`
  - `--timeout`
  - `--wait-for` / `-w`
  - `--port` / `-p`
  - `--close-tab` / `-c`
  - `--force-headless`
  - `--open-browser` / `-b`
  - Another `--list-tabs`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`
  - `--user-data-dir`

**Define:** Standalone mode requirement, all conflicts

---

### Task 13: `--tab PATTERN` / `-t`

**Questions to answer:**
- What happens with wrong values? (no match found, invalid regex, empty pattern)
- What happens when combined with:
  - `<url>` (single)
  - `<url>` (multiple)
  - `--url-file`
  - `--output` / `-o`
  - `--output-dir` / `-d`
  - `--format` / `-f`
  - `--timeout`
  - `--wait-for` / `-w`
  - `--port` / `-p`
  - `--close-tab` / `-c`
  - `--force-headless`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - Another `--tab`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`
  - `--user-data-dir`

**Define:** Pattern matching priority, conflicts, should-ignore flags

---

### Task 14: `--all-tabs` / `-a`

**Questions to answer:**
- What happens with wrong values? (N/A - boolean flag)
- What happens when combined with:
  - `<url>` (single)
  - `<url>` (multiple)
  - `--url-file`
  - `--output` / `-o`
  - `--output-dir` / `-d`
  - `--format` / `-f`
  - `--timeout`
  - `--wait-for` / `-w`
  - `--port` / `-p`
  - `--close-tab` / `-c`
  - `--force-headless`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - Another `--all-tabs`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`
  - `--user-data-dir`

**Define:** Output requirement, conflicts, error handling behavior

---

### Task 15: `--verbose`

**Questions to answer:**
- What happens with wrong values? (N/A - boolean flag)
- What happens when combined with:
  - `<url>` (single)
  - `<url>` (multiple)
  - `--url-file`
  - `--output` / `-o`
  - `--output-dir` / `-d`
  - `--format` / `-f`
  - `--timeout`
  - `--wait-for` / `-w`
  - `--port` / `-p`
  - `--close-tab` / `-c`
  - `--force-headless`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - Another `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`
  - `--user-data-dir`

**Define:** Logging level priority order, conflicts with `--quiet` and `--debug`

---

### Task 16: `--quiet` / `-q`

**Questions to answer:**
- What happens with wrong values? (N/A - boolean flag)
- What happens when combined with:
  - `<url>` (single)
  - `<url>` (multiple)
  - `--url-file`
  - `--output` / `-o`
  - `--output-dir` / `-d`
  - `--format` / `-f`
  - `--timeout`
  - `--wait-for` / `-w`
  - `--port` / `-p`
  - `--close-tab` / `-c`
  - `--force-headless`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - Another `--quiet`
  - `--debug`
  - `--user-agent`
  - `--user-data-dir`

**Define:** Logging level priority order, conflicts with `--verbose` and `--debug`

---

### Task 17: `--debug`

**Questions to answer:**
- What happens with wrong values? (N/A - boolean flag)
- What happens when combined with:
  - `<url>` (single)
  - `<url>` (multiple)
  - `--url-file`
  - `--output` / `-o`
  - `--output-dir` / `-d`
  - `--format` / `-f`
  - `--timeout`
  - `--wait-for` / `-w`
  - `--port` / `-p`
  - `--close-tab` / `-c`
  - `--force-headless`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - Another `--debug`
  - `--user-agent`
  - `--user-data-dir`

**Define:** Logging level priority order, output format, conflicts

---

### Task 18: `--user-agent STRING`

**Questions to answer:**
- What happens with wrong values? (empty string, extremely long string)
- What happens when combined with:
  - `<url>` (single)
  - `<url>` (multiple)
  - `--url-file`
  - `--output` / `-o`
  - `--output-dir` / `-d`
  - `--format` / `-f`
  - `--timeout`
  - `--wait-for` / `-w`
  - `--port` / `-p`
  - `--close-tab` / `-c`
  - `--force-headless`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - Another `--user-agent`
  - `--user-data-dir`

**Define:** Behavior with new pages vs existing tabs (should ignore for existing tabs)

---

### Task 19: `--help` / `-h`

**Questions to answer:**
- What happens with wrong values? (N/A - boolean flag)
- What happens when combined with any other argument (including `--user-data-dir`)?

**Define:** Should display help and exit, ignoring all other flags

---

### Task 20: `--version` / `-v`

**Questions to answer:**
- What happens with wrong values? (N/A - boolean flag)
- What happens when combined with any other argument (including `--user-data-dir`)?

**Define:** Should display version and exit, ignoring all other flags

---

### Task 21: `--user-data-dir DIRECTORY` ✅

**Status:** Complete (2025-10-23)

#### Invalid Values & Basic Behavior

**Empty/whitespace:**
- Warn and use browser default profile

**Directory doesn't exist:**
- Error

**Path is file not directory:**
- Error

**Permission denied:**
- Error

**Invalid path characters:**
- Error

**Relative/absolute paths:**
- Both supported, browser handles as needed

**Multiple flags:**
- Last wins

**Tilde expansion:**
- Snag expands `~` to home directory

#### Browser Mode Interactions

- Works with `--force-headless` and `--open-browser`
- Existing browser connection → Warn and ignore flag
- Multiple instances same profile → Let browser error
- Profile persists between invocations (headless and visible)

#### Content Source Interactions

- Works with URLs, `--url-file`
- Tab operations (`--tab`, `--all-tabs`) → Warn and ignore
- `--list-tabs` → List-tabs wins (standalone like `--help`)

#### Other Flag Interactions

- Output control, timing, wait, close-tab → All work normally
- `--user-agent` → Custom profile WITH custom UA (both applied)
- Logging flags → Work normally
- `--help` / `--version` → These win, ignore profile

---

## Completion Status

Track completion of each task:

- [x] Task 1: `<url>` - **COMPLETE** (2025-10-22)
- [x] Task 2: `--url-file` - **COMPLETE** (2025-10-22)
- [x] Task 3: `--output` / `-o` - **COMPLETE** (2025-10-22)
- [x] Task 4: `--output-dir` / `-d` - **COMPLETE** (2025-10-22)
- [x] Task 5: `--format` / `-f` - **COMPLETE** (2025-10-22)
- [x] Task 6: `--timeout` - **COMPLETE** (2025-10-22)
- [x] Task 7: `--wait-for` / `-w` - **COMPLETE** (2025-10-23)
- [x] Task 8: `--port` / `-p` - **COMPLETE** (2025-10-23)
- [x] Task 9: `--close-tab` / `-c` - **COMPLETE** (2025-10-23)
- [x] Task 10: `--force-headless` - **COMPLETE** (2025-10-23)
- [x] Task 11: `--open-browser` / `-b` - **COMPLETE** (2025-10-23)
- [x] Task 12: `--list-tabs` / `-l` - **COMPLETE** (2025-10-23)
- [x] Task 13: `--tab` / `-t` - **COMPLETE** (2025-10-23)
- [x] Task 14: `--all-tabs` / `-a` - **COMPLETE** (2025-10-23)
- [x] Task 15: `--verbose` - **COMPLETE** (2025-10-23)
- [x] Task 16: `--quiet` / `-q` - **COMPLETE** (2025-10-23)
- [x] Task 17: `--debug` - **COMPLETE** (2025-10-23)
- [x] Task 18: `--user-agent` - **COMPLETE** (2025-10-23)
- [x] Task 19: `--help` / `-h` - **COMPLETE** (2025-10-23)
- [x] Task 20: `--version` / `-v` - **COMPLETE** (2025-10-23)
- [x] Task 21: `--user-data-dir` - **COMPLETE** (2025-10-23)

---

## Output Document

All findings are documented in `docs/arguments/` directory with:
- Individual files for each argument's behavior definitions
- README.md with compatibility matrices and quick reference
- validation.md with error conditions, validation requirements, and edge cases
- Organized by argument for easy navigation

---

## Project Status: COMPLETE ✅

**Completion Date:** 2025-10-23

All 21 arguments have been systematically analyzed and documented. This design-first approach successfully defined the behavior for every argument combination before implementation.

### What Was Accomplished

**Design Methodology:**
- Systematic design-first, implementation-later approach
- Each argument analyzed for invalid values and all possible combinations
- Comprehensive documentation created before writing code
- Cross-review performed to ensure consistency

**Deliverables:**
- 21 individual argument specification files in `docs/arguments/`
- Complete compatibility matrices in `docs/arguments/README.md`
- Validation rules and edge cases in `docs/arguments/validation.md`
- Cross-review findings in `PROJECT-review.md`
- Implementation tasks in `PROJECT-todo.md`

**Arguments Completed:**
1. `<url>` - Positional URL argument
2. `--url-file` - Read URLs from file
3. `--output` / `-o` - Save to specific file
4. `--output-dir` / `-d` - Save with auto-generated name
5. `--format` / `-f` - Output format selection
6. `--timeout` - Page load timeout
7. `--wait-for` / `-w` - Wait for CSS selector
8. `--port` / `-p` - Remote debugging port
9. `--close-tab` / `-c` - Close tab after fetching
10. `--force-headless` - Force headless mode
11. `--open-browser` / `-b` - Open visible browser
12. `--list-tabs` / `-l` - List all open tabs
13. `--tab` / `-t` - Fetch from existing tab
14. `--all-tabs` / `-a` - Process all open tabs
15. `--verbose` - Verbose logging
16. `--quiet` / `-q` - Quiet mode
17. `--debug` - Debug logging
18. `--user-agent` - Custom user agent
19. `--help` / `-h` - Show help
20. `--version` / `-v` - Show version
21. `--user-data-dir` - Custom browser profile

### Quality Assurance

**Cross-Review (2025-10-23):**
- Reviewed all 21 argument files for bidirectional consistency
- Found 7 critical contradictions and 7 minor inconsistencies
- All findings documented in PROJECT-review.md
- Recommendations provided for resolution

### Next Steps

**Implementation Phase:**
1. Resolve documentation inconsistencies (see PROJECT-review.md)
2. Implement validation logic based on documented behavior
3. Complete outstanding tasks (see PROJECT-todo.md)
4. Write integration tests covering all argument combinations

**Documentation:**
- Outstanding tasks tracked in PROJECT-todo.md
- Cross-review findings in PROJECT-review.md
- All argument specifications in docs/arguments/

---

## Process Reflection

This design-first approach prevented:
- Inconsistent behavior across similar scenarios
- Undocumented edge cases discovered by users
- Implementation decisions made without considering all implications
- Technical debt from "we'll figure it out later" approaches

The systematic question-and-answer process ensured every combination was explicitly designed, not assumed.

---

**Project Archive:** This PROJECT.md document represents the completed design phase. All active work items have been moved to PROJECT-todo.md.
