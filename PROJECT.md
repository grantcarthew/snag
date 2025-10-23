# Argument Handling Analysis - Design Decision Document

**Project:** Systematically analyze each snag CLI argument's behavior and interactions

**Goal:** Create comprehensive design decisions for every argument combination

**Deliverable:** Complete `docs/argument-handling.md` with all behaviors defined

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
   - Update `docs/argument-handling.md` with the design decisions

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
  - `--force-visible`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`

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
  - `--force-visible`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`

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
  - `--force-visible`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`

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
  - `--force-visible`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`

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
  - `--force-visible`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`

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
  - `--force-visible`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`

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
| `--wait-for` + `--force-visible` | Works normally | Wait in visible mode |
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
  - `--force-visible`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`

**Define:** Port availability checking, all operation modes

---

### Task 9: `--close-tab` / `-c`

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
  - Another `--close-tab`
  - `--force-headless`
  - `--force-visible`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`

**Define:** Behavior with existing tabs vs new tabs, batch operations

---

### Task 10: `--force-headless`

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
  - Another `--force-headless`
  - `--force-visible`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`

**Define:** Conflict with `--force-visible` and `--open-browser`, behavior with existing browser

---

### Task 11: `--force-visible`

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
  - Another `--force-visible`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`

**Define:** Conflict with `--force-headless`, redundancy with `--open-browser`

---

### Task 12: `--open-browser` / `-b`

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
  - `--force-visible`
  - Another `--open-browser`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`

**Define:** Behavior with vs without URL (current vs planned), conflicts

---

### Task 13: `--list-tabs` / `-l`

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
  - `--force-visible`
  - `--open-browser` / `-b`
  - Another `--list-tabs`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`

**Define:** Standalone mode requirement, all conflicts

---

### Task 14: `--tab PATTERN` / `-t`

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
  - `--force-visible`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - Another `--tab`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`

**Define:** Pattern matching priority, conflicts, should-ignore flags

---

### Task 15: `--all-tabs` / `-a`

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
  - `--force-visible`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - Another `--all-tabs`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`

**Define:** Output requirement, conflicts, error handling behavior

---

### Task 16: `--verbose`

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
  - `--force-visible`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - Another `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - `--user-agent`

**Define:** Logging level priority order, conflicts with `--quiet` and `--debug`

---

### Task 17: `--quiet` / `-q`

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
  - `--force-visible`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - Another `--quiet`
  - `--debug`
  - `--user-agent`

**Define:** Logging level priority order, conflicts with `--verbose` and `--debug`

---

### Task 18: `--debug`

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
  - `--force-visible`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - Another `--debug`
  - `--user-agent`

**Define:** Logging level priority order, output format, conflicts

---

### Task 19: `--user-agent STRING`

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
  - `--force-visible`
  - `--open-browser` / `-b`
  - `--list-tabs` / `-l`
  - `--tab` / `-t`
  - `--all-tabs` / `-a`
  - `--verbose`
  - `--quiet` / `-q`
  - `--debug`
  - Another `--user-agent`

**Define:** Behavior with new pages vs existing tabs (should ignore for existing tabs)

---

### Task 20: `--help` / `-h`

**Questions to answer:**
- What happens with wrong values? (N/A - boolean flag)
- What happens when combined with any other argument?

**Define:** Should display help and exit, ignoring all other flags

---

### Task 21: `--version` / `-v`

**Questions to answer:**
- What happens with wrong values? (N/A - boolean flag)
- What happens when combined with any other argument?

**Define:** Should display version and exit, ignoring all other flags

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
- [ ] Task 8: `--port` / `-p`
- [ ] Task 9: `--close-tab` / `-c`
- [ ] Task 10: `--force-headless`
- [ ] Task 11: `--force-visible`
- [ ] Task 12: `--open-browser` / `-b`
- [ ] Task 13: `--list-tabs` / `-l`
- [ ] Task 14: `--tab` / `-t`
- [ ] Task 15: `--all-tabs` / `-a`
- [ ] Task 16: `--verbose`
- [ ] Task 17: `--quiet` / `-q`
- [ ] Task 18: `--debug`
- [ ] Task 19: `--user-agent`
- [ ] Task 20: `--help` / `-h`
- [ ] Task 21: `--version` / `-v`

---

## Output Document

All findings will be documented in `docs/argument-handling.md` with:
- Behavior definitions for each argument
- Compatibility matrix
- Error conditions and messages
- Validation requirements
- Edge case handling

---

## TODO Items

- [ ] **Argument trimming**: Scan all string arguments (`--wait-for`, `--user-agent`, `--output`, `--output-dir`, `--tab`, etc.) and apply `strings.TrimSpace()` after reading from CLI framework. This is standard behavior in most CLI tools (git, docker, etc.) and handles copy-paste trailing spaces gracefully.
