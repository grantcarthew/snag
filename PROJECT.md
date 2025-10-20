# PROJECT.md - Phase 2: Tab Management

> **Project**: snag - Phase 2 Enhancement
> **Feature**: Tab Management
> **Status**: Phase 2.1 Complete, Phase 2.2 In Progress
> **Created**: 2025-10-20
> **Last Updated**: 2025-10-20
> **Target Version**: v0.1.0 (Phase 2)

## Session Summary (2025-10-20)

### Completed: Phase 2.1 - Tab Listing Feature

**What was built**:
- `--list-tabs` (`-l`) flag to list all open browser tabs
- `TabInfo` struct to represent tab metadata (index, URL, title, ID)
- `ListTabs()` function to retrieve tabs from browser via rod
- `handleListTabs()` CLI handler with connect-only mode
- Integration tests validating both error and success cases

**Key Achievement**: First tab management feature working end-to-end with full test coverage.

**Critical Bug Discovered & Fixed**:
- **Issue**: When using default port 9222, rod's launcher wouldn't explicitly set `--remote-debugging-port`, causing it to pick a random port
- **Impact**: Browsers launched with `--force-visible` couldn't be reached by `--list-tabs`
- **Root Cause**: Code only set debugging port when `!= 9222` (browser.go:260-262)
- **Fix**: Always explicitly set `remote-debugging-port` flag regardless of value
- **Location**: browser.go:259-260
- **Lesson**: Never rely on framework defaults for critical connection parameters

### Key Learnings

1. **Remote Debugging Port Configuration**:
   - ALWAYS explicitly set `--remote-debugging-port` flag
   - Don't assume framework defaults match your expectations
   - Test with both default and custom ports

2. **Tab Management Prerequisites**:
   - Tab features (`--list-tabs`, `--tab`) require existing browser connection
   - Will NOT auto-launch browser (design decision)
   - Fail fast with clear error messages guiding user to `--open-browser`

3. **Testing Strategy**:
   - Integration tests more valuable than unit tests for browser operations
   - Test both error paths (no browser) and success paths (with browser)
   - Manual testing revealed the port configuration bug

4. **1-Based vs 0-Based Indexing**:
   - User-facing: 1-based (tabs [1], [2], [3]...)
   - Internal: 0-based (converted in TabInfo struct)
   - Better UX for CLI tool vs programming API

### Next Steps

**Phase 2.2** (Ready to implement):
- Add `--tab <index>` flag to fetch from specific tab by index
- Move `-t` alias from `--timeout` to `--tab`
- Implement `GetTabByIndex()` function
- Add `handleTabFetch()` handler
- Integration tests for tab fetching

**Phase 2.3** (Subsequent):
- Pattern matching with `--tab <pattern>`
- Progressive fallthrough (integer → regex → exact → substring)
- Full regex support with case-insensitive matching

## Overview

This document defines the implementation plan for Phase 2 of snag: **Tab Management**. This enhancement adds the ability to list, select, and fetch content from existing browser tabs, avoiding unnecessary tab creation and enabling efficient content retrieval from already-authenticated sessions.

**Current State**: snag can connect to existing Chrome instances and fetch URLs by creating new tabs/pages. ✅ Phase 2.1 adds `--list-tabs` to view existing tabs.

**Goal**: Enable snag to work with existing tabs in the browser without creating new ones.

## Objectives

1. **List existing tabs** - Display all open tabs with their index, URL, and title
2. **Select tab by index** - Fetch content from a specific tab using its index
3. **Select tab by URL pattern** - Match and fetch from a tab using URL pattern matching
4. **Preserve existing behavior** - Maintain backward compatibility with Phase 1 functionality
5. **Enable efficient workflows** - Reduce tab clutter and leverage existing authenticated sessions

## Use Cases

### Use Case 1: List All Open Tabs

**Scenario**: User has multiple tabs open and wants to see what's available.

```bash
$ snag --list-tabs

Available tabs in Chrome (10 tabs):
  [1] https://github.com/grantcarthew/snag - grantcarthew/snag: Intelligent web content fetcher
  [2] https://go.dev/doc/ - Documentation - The Go Programming Language
  [3] https://pkg.go.dev/github.com/go-rod/rod - rod package - github.com/go-rod/rod - Go Packages
  [4] https://example.com - Example Domain
  [5] https://news.ycombinator.com/ - Hacker News
  [6] https://app.internal.com/dashboard - Dashboard - Internal App (authenticated)
  [7] about:blank - New Tab
  [8] https://github.com/JohannesKaufmann/html-to-markdown - JohannesKaufmann/html-to-markdown
  [9] https://claude.ai/chat - Claude
  [10] https://docs.google.com/document/d/xxx - Meeting Notes - Google Docs
```

**Expected Output**: List of tabs with index, URL, and title to stderr (structured info, not content).

**Exit Code**: 0 (success)

### Use Case 2: Fetch from Specific Tab by Index

**Scenario**: User wants content from an already-open authenticated page without creating a new tab.

```bash
$ snag --tab 6
# or
$ snag -t 6
# Fetches content from tab [6] (app.internal.com/dashboard)
# Outputs Markdown to stdout
```

**Workflow**:
1. Connect to existing Chrome instance
2. Get list of all pages/tabs
3. Select tab at index 6 (internally converts to 0-based index 5)
4. Extract HTML content from that tab
5. Convert to Markdown (or HTML with --format html)
6. Output to stdout (or file with --output)

**Benefits**:
- No new tab created (cleaner browser)
- Uses existing authentication/cookies in that tab
- Instant access to current state of the page

### Use Case 3: Fetch from Tab by URL Pattern

**Scenario**: User knows the URL pattern but not the tab index.

```bash
# Exact match (case-insensitive)
$ snag -t https://app.internal.com/dashboard
$ snag -t GITHUB.COM/grantcarthew/snag  # matches github.com (case-insensitive)

# Regex pattern matching
$ snag -t "github\.com/.*"              # regex: github.com/ + anything
$ snag -t ".*/dashboard"                # regex: any URL ending with /dashboard
$ snag -t "(github|gitlab)\.com"        # regex: github.com or gitlab.com

# Substring/contains matching (fallback)
$ snag -t "dashboard"                   # contains "dashboard"
$ snag -t "github"                      # contains "github"
```

**Workflow**:
1. Connect to existing Chrome instance
2. Get list of all pages/tabs
3. Find first tab matching the URL pattern
4. Extract and convert content
5. Output result

**Pattern Matching Rules** (Progressive Fallthrough):
1. Integer → Tab index (1-based)
2. Has regex chars → Regex match (case-insensitive)
3. Exact URL match (case-insensitive)
4. Substring/contains match (case-insensitive)
5. Error if no tabs match

If multiple tabs match, first match wins.

### Use Case 4: List Tabs with Output Format

**Scenario**: User wants machine-readable tab list for scripting.

```bash
# JSON output (future enhancement)
$ snag --list-tabs --format json
[
  {"index": 1, "url": "https://github.com/grantcarthew/snag", "title": "grantcarthew/snag: Intelligent web content fetcher"},
  {"index": 2, "url": "https://go.dev/doc/", "title": "Documentation - The Go Programming Language"},
  ...
]
```

**Note**: JSON format for --list-tabs is a future enhancement, not MVP.

## Feature Specifications

### Feature 1: `--list-tabs` Flag

**Flag Definition**:

```go
&cli.BoolFlag{
    Name:    "list-tabs",
    Aliases: []string{"l"},
    Usage:   "List all open tabs in the browser",
}
```

**Behavior**:
- Connects to existing Chrome instance (requires Chrome running with remote debugging)
- Lists all pages/tabs with index, URL, and title
- Output goes to **stdout** (not stderr) for piping/scripting
- Does NOT require a URL argument
- Does NOT fetch any content
- Exit code 0 on success, 1 on error (e.g., browser not running)

**Error Handling**:
- If no Chrome instance is running: Error message with suggestion to start Chrome
- If connection fails: Clear error with debugging instructions

**Output Format** (MVP - Human-Readable):

```
Available tabs in Chrome (10 tabs):
  [1] https://github.com/grantcarthew/snag - grantcarthew/snag: Intelligent web content fetcher
  [2] https://go.dev/doc/ - Documentation - The Go Programming Language
  ...
```

**Note**: Tab indexes are 1-based for user display (more intuitive).

**Future**: Add `--format json` support for machine-readable output.

### Feature 2: `--tab <value>` Flag

**Flag Definition**:

```go
&cli.StringFlag{
    Name:    "tab",
    Aliases: []string{"t"},
    Usage:   "Fetch from existing tab by `INDEX` or URL pattern",
}
```

**Note**: The `-t` short alias is moved from `--timeout` to `--tab` (more frequently used).

**Behavior**:
- Connects to existing Chrome instance
- Accepts multiple value types with progressive matching:
  - **Integer**: Tab index 1-based (e.g., `--tab 1` for first tab)
  - **Regex pattern**: Full regex support (e.g., `--tab "github\.com/.*"`)
  - **Exact match**: Case-insensitive URL (e.g., `--tab "github.com"`)
  - **Substring**: Contains match fallback (e.g., `--tab "dashboard"`)
- Fetches content from the matched tab
- Converts to requested format (Markdown/HTML)
- Outputs to stdout or file (respects --output flag)
- Does NOT create a new tab
- Does NOT navigate the tab to a different URL
- Captures current state of the tab
- All matching is case-insensitive

**Tab Selection Logic** (Progressive Fallthrough):

1. **Integer check**: If value is valid integer → Use as tab index (1-based, converts to 0-based internally)
2. **Regex detection**: If contains regex chars (`* + ? [ ] { } ( ) | ^ $ \`) → Compile and match as regex (case-insensitive)
3. **Exact match**: Try case-insensitive exact URL match
4. **Contains match**: Try case-insensitive substring search
5. **Error**: If no tabs match at any stage

**First matching tab wins** if multiple tabs match the pattern.

**Interaction with URL Argument**:
- `--tab` and `<url>` argument are **mutually exclusive**
- If both provided: Error with clear message
- If `--tab` provided: Ignore URL argument, fetch from existing tab
- If neither provided: Error (URL required)

**Error Handling**:
- Browser not running: "No Chrome instance running with remote debugging"
- Tab index out of range: "Tab index 15 out of range (1-10)"
- Pattern no match: "No tab matches pattern 'github.com/foo'"
- Invalid regex: "Invalid regex pattern 'github\.com/[invalid': missing closing bracket"

**Examples**:

```bash
# Fetch from tab by index (1-based)
snag -t 1                                # First tab
snag -t 6                                # Sixth tab

# Fetch by exact URL (case-insensitive)
snag -t https://github.com/grantcarthew/snag
snag -t GITHUB.COM                       # Case-insensitive

# Fetch by regex pattern
snag -t "github\.com/.*"                 # Regex: github.com/ + anything
snag -t ".*/dashboard"                   # Regex: ends with /dashboard
snag -t "(github|gitlab)\.com/snag"      # Regex: alternation

# Fetch by substring (fallback)
snag -t "dashboard"                      # Contains "dashboard"
snag -t "internal"                       # Contains "internal"

# With output options
snag -t 6 --output content.md            # Save to file
snag -t 6 --format html                  # HTML format
snag -t "github" -o repo.md --verbose    # Verbose mode
```

### Feature 3: Integration with Existing Flags

**Compatible Flags** (work with --tab):
- `--format` (markdown/html)
- `--output` (save to file)
- `--verbose`, `--quiet`, `--debug` (logging levels)
- `--port` (specify remote debugging port)
- `--wait-for` (wait for selector before extracting - useful for dynamic content)

**Incompatible Flags** (ignored or error with --tab):
- `--close-tab` - **Ignored** when using --tab (we don't close existing tabs)
- `--force-headless` - **Ignored** (can't change mode of existing browser)
- `--force-visible` - **Ignored** (can't change mode of existing browser)
- `--open-browser` - **Error** (mutually exclusive)
- `--user-agent` - **Ignored** (tab already has its user agent)
- `<url>` argument - **Error** (mutually exclusive with --tab)

**Timeout Behavior**:
- `--timeout` applies to content extraction (wait-for selector, page stabilization)
- Does NOT apply to navigation (we're not navigating)

## Technical Design

### Architecture Changes

**No Major Architectural Changes Required**:
- Existing `BrowserManager` handles connection to Chrome
- New functions in `browser.go` for tab listing and selection
- Minimal changes to `main.go` for new flags and logic
- Reuse existing `fetch.go` and `convert.go` for content extraction

### New Components

#### 1. Tab Information Structure

**File**: `browser.go`

```go
// TabInfo represents information about a browser tab
type TabInfo struct {
    Index int    // Tab index (1-based for display, internally 0-based)
    URL   string // Current URL of the tab
    Title string // Page title
    ID    string // Internal target ID (for rod)
}
```

#### 2. Tab Management Functions

**File**: `browser.go`

**Function**: `ListTabs() ([]TabInfo, error)`

- Connects to existing browser
- Retrieves all pages using `browser.Pages()`
- Extracts URL, title, and creates index
- Returns slice of TabInfo
- Error if browser not connected

**Function**: `GetTabByIndex(index int) (*rod.Page, error)`

- Gets list of tabs
- Validates index is in range
- Returns the rod.Page at that index
- Error if index out of range

**Function**: `GetTabByPattern(pattern string) (*rod.Page, error)`

- Gets list of tabs
- First tries exact URL match
- Then tries wildcard pattern matching
- Returns first matching rod.Page
- Error if no match found

**Function**: `hasRegexChars(s string) bool`

- Helper to detect if string contains regex metacharacters
- Regex chars: `* + ? [ ] { } ( ) | ^ $ \`
- NOT regex (URL-safe): `. / - _ : ? & = # %`
- Returns true if string has regex special characters

**Function**: `matchURLPattern(url, pattern string) bool` _(deprecated/replaced)_

- Replaced by progressive fallthrough in `GetTabByPattern()`
- No longer needed as standalone helper

#### 3. CLI Flag Validation

**File**: `main.go`

**Logic Changes in `run()` function**:

```go
// Validate flag combinations
if c.Bool("list-tabs") {
    // Just list tabs and exit
    return handleListTabs(c)
}

if c.IsSet("tab") {
    // Fetch from existing tab
    if c.NArg() > 0 {
        return fmt.Errorf("cannot use --tab with URL argument")
    }
    return handleTabFetch(c)
}

// Existing URL fetch logic
if c.NArg() == 0 {
    return fmt.Errorf("URL argument required")
}
// ... rest of existing logic
```

**New Handler Functions**:

```go
// handleListTabs lists all open tabs
func handleListTabs(c *cli.Context) error {
    // Connect to existing browser (only)
    // Get tabs list
    // Format and print to stdout
    // Return error or nil
}

// handleTabFetch fetches content from an existing tab
func handleTabFetch(c *cli.Context) error {
    // Get --tab value
    // Determine if index or pattern
    // Connect to browser
    // Get tab (by index or pattern)
    // Fetch content (reuse existing fetch logic)
    // Convert and output (reuse existing logic)
}
```

### Rod API Usage

**Key Rod Methods**:

```go
// Get all pages/tabs
pages, err := browser.Pages()

// Each page has:
page.MustInfo().Title  // Page title
page.MustInfo().URL    // Current URL
page.TargetID          // Internal target ID
```

**Implementation Pattern**:

```go
func (bm *BrowserManager) ListTabs() ([]TabInfo, error) {
    if bm.browser == nil {
        return nil, fmt.Errorf("browser not connected")
    }

    pages, err := bm.browser.Pages()
    if err != nil {
        return nil, fmt.Errorf("failed to get pages: %w", err)
    }

    var tabs []TabInfo
    for i, page := range pages {
        info, err := page.Info()
        if err != nil {
            logger.Warning("Failed to get info for page %d: %v", i, err)
            continue
        }

        tabs = append(tabs, TabInfo{
            Index: i,
            URL:   info.URL,
            Title: info.Title,
            ID:    string(page.TargetID),
        })
    }

    return tabs, nil
}
```

### Pattern Matching Algorithm

**Progressive Fallthrough with Full Regex Support**:

```go
func (bm *BrowserManager) GetTabByPattern(pattern string) (*rod.Page, error) {
    pages, err := bm.browser.Pages()
    if err != nil {
        return nil, fmt.Errorf("failed to get pages: %w", err)
    }

    if len(pages) == 0 {
        return nil, ErrNoTabsOpen
    }

    patternLower := strings.ToLower(pattern)
    hasRegex := hasRegexChars(pattern)

    // Step 1: Try regex if pattern has regex chars
    if hasRegex {
        re, err := regexp.Compile("(?i)" + pattern) // (?i) = case-insensitive
        if err != nil {
            return nil, fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
        }

        for i, page := range pages {
            info, err := page.Info()
            if err != nil {
                continue
            }
            if re.MatchString(info.URL) {
                logger.Verbose("Matched tab [%d] via regex: %s", i+1, info.URL)
                return page, nil
            }
        }
        logger.Debug("No regex match for '%s', trying exact/contains...", pattern)
        // Fall through to exact/contains
    }

    // Step 2: Try exact match (case-insensitive)
    for i, page := range pages {
        info, err := page.Info()
        if err != nil {
            continue
        }
        if strings.EqualFold(info.URL, pattern) {
            logger.Verbose("Matched tab [%d] via exact URL: %s", i+1, info.URL)
            return page, nil
        }
    }

    // Step 3: Try contains/substring match (case-insensitive)
    for i, page := range pages {
        info, err := page.Info()
        if err != nil {
            continue
        }
        if strings.Contains(strings.ToLower(info.URL), patternLower) {
            logger.Verbose("Matched tab [%d] via substring: %s", i+1, info.URL)
            return page, nil
        }
    }

    // Step 4: No matches found
    return nil, fmt.Errorf("%w: '%s'", ErrNoTabMatch, pattern)
}

// hasRegexChars checks if string contains regex metacharacters
// (excluding characters commonly found in URLs)
func hasRegexChars(s string) bool {
    // Regex metacharacters that are NOT common in URLs:
    // * + ? [ ] { } ( ) | ^ $ \
    //
    // Note: We exclude . / - _ : ? & = # % which are URL-safe
    regexChars := `*+?[]{}()|^$\`
    return strings.ContainsAny(s, regexChars)
}
```

**Matching Examples**:

```go
// 1. Integer - Tab index (1-based)
"1"          → Tab 1 (first tab, internally index 0)
"10"         → Tab 10 (tenth tab, internally index 9)

// 2. Regex pattern (has regex chars)
"github\\.com/.*"              → Matches https://github.com/grantcarthew/snag
".*/dashboard"                 → Matches https://app.internal.com/dashboard
"(github|gitlab)\\.com"        → Matches github.com or gitlab.com
"docs\\.(md|html)$"            → Matches URLs ending with .md or .html

// 3. Exact match (case-insensitive)
"https://github.com/grantcarthew/snag"  → Exact match
"GITHUB.COM"                             → Matches github.com (case-insensitive)

// 4. Contains/substring match (fallback)
"dashboard"                    → Contains "dashboard" anywhere
"github"                       → Contains "github" anywhere
"internal.com"                 → Contains "internal.com"
```

**Note**: Regex fallthrough means even if regex chars are detected, if the regex doesn't match, we still try exact and contains matching. This provides maximum flexibility.

### Error Handling

**New Sentinel Errors**:

```go
// errors.go
var (
    ErrNoBrowserRunning = errors.New("no Chrome instance running with remote debugging")
    ErrTabNotFound      = errors.New("tab not found")
    ErrTabIndexInvalid  = errors.New("tab index out of range")
    ErrNoTabMatch       = errors.New("no tab matches pattern")
    ErrTabURLConflict   = errors.New("cannot use --tab with URL argument")
    ErrNoTabsOpen       = errors.New("no tabs open in browser")
)
```

**Error Messages**:

- Clear, actionable error messages
- Suggest solutions when possible
- Include context (which pattern failed, valid range, etc.)

**Examples**:

```
✗ No Chrome instance running with remote debugging
  Start Chrome with: chrome --remote-debugging-port=9222
  Or run: snag --open-browser

✗ Tab index 15 out of range (1-10)
  Run 'snag --list-tabs' to see available tabs

✗ No tab matches pattern 'github.com/foo'
  Run 'snag --list-tabs' to see available tabs

✗ Invalid regex pattern 'github\.com/[invalid': missing closing bracket

  Common patterns:
    github.com                  Substring match (contains)
    github\.com/.*              Regex: github.com/ + anything
    .*/dashboard                Regex: ends with /dashboard
    (github|gitlab)\.com        Regex: github.com or gitlab.com
```

## Implementation Plan

### Phase 2.1: Tab Listing (--list-tabs)

**Files to Modify**:
- `browser.go`: Add `ListTabs()`, `TabInfo` struct
- `main.go`: Add `--list-tabs` flag, `handleListTabs()` function
- `errors.go`: Add `ErrNoBrowserRunning`

**Note**: Phase 2.1 doesn't require changing the timeout flag yet.

**Implementation Steps**:

1. **Add TabInfo struct** to `browser.go`
2. **Implement ListTabs()** function in `browser.go`
   - Use `browser.Pages()` to get all pages
   - Extract URL, title for each page
   - Handle errors gracefully (skip pages with errors)
   - Return slice of TabInfo
3. **Add --list-tabs flag** to `main.go`
4. **Implement handleListTabs()** in `main.go`
   - Create BrowserManager with connect-only mode
   - Call `ListTabs()`
   - Format output (human-readable list)
   - Print to stdout
   - Handle errors
5. **Add validation logic** in `run()` to route to `handleListTabs()`
6. **Test thoroughly**:
   - With Chrome running (multiple tabs)
   - With no Chrome running
   - With single tab
   - With many tabs (20+)

**Success Criteria**:
- `snag --list-tabs` lists all open tabs
- Clean, readable output format
- Correct indexing (1-based for display)
- Handles errors gracefully

### Phase 2.2: Tab Selection by Index (--tab <index>)

**Files to Modify**:
- `browser.go`: Add `GetTabByIndex()`
- `main.go`: Add `--tab` flag with `-t` alias, `handleTabFetch()` function, **REMOVE** `-t` alias from `--timeout` flag
- `errors.go`: Add `ErrTabIndexInvalid`, `ErrTabURLConflict`

**Implementation Steps**:

1. **Add --tab flag** to `main.go` with `-t` alias
2. **Remove `-t` alias** from `--timeout` flag in `main.go`
3. **Implement GetTabByIndex()** in `browser.go`
   - Get all pages
   - Validate index in range
   - Return page at index
4. **Implement handleTabFetch()** in `main.go`
   - Parse --tab value (detect if integer)
   - Connect to browser
   - Get tab by index
   - **Reuse existing content fetch logic** from main `run()` function
   - Extract HTML from page
   - Convert to Markdown/HTML
   - Output to stdout/file
5. **Add validation** for --tab and URL conflict
6. **Test thoroughly**:
   - Fetch from tab 0
   - Fetch from last tab
   - Invalid index (negative, too large)
   - Output formats (markdown, html)
   - Save to file (--output)

**Success Criteria**:
- `snag --tab 5` fetches content from tab 5
- Works with all output flags (--format, --output)
- Clear error for invalid index
- Doesn't create new tabs

### Phase 2.3: Tab Selection by Pattern (--tab <pattern>)

**Files to Modify**:
- `browser.go`: Add `GetTabByPattern()`, `matchURLPattern()`
- `errors.go`: Add `ErrNoTabMatch`

**Implementation Steps**:

1. **Implement hasRegexChars()** helper in `browser.go`
   - Detect regex metacharacters: `* + ? [ ] { } ( ) | ^ $ \`
   - Exclude URL-safe chars: `. / - _ : ? & = # %`
2. **Implement GetTabByPattern()** in `browser.go`
   - Get all pages
   - Try regex match if pattern has regex chars
   - Try exact URL match (case-insensitive)
   - Try contains/substring match (case-insensitive)
   - Return first match
   - Error if no match
3. **Update handleTabFetch()** to support patterns
   - Detect if --tab value is integer or string
   - Route to GetTabByIndex or GetTabByPattern
4. **Test thoroughly**:
   - Exact URL match (case-insensitive)
   - Regex patterns (alternation, wildcards, anchors)
   - Substring/contains matching
   - Invalid regex (error handling)
   - Multiple matching tabs (use first)
   - No matching tabs (error)
   - Edge cases (special chars in URLs, case variations)

**Success Criteria**:
- `snag -t "github\.com/.*"` matches correct tab via regex
- `snag -t "github"` matches correct tab via substring
- Case-insensitive matching works
- Progressive fallthrough works correctly
- Clear error when no match
- First match used when multiple matches

### Phase 2.4: Integration Testing & Documentation

**Tasks**:

1. **Integration testing**:
   - Test all flag combinations
   - Test with existing flags (--format, --output, --verbose, etc.)
   - Test error cases
   - Test with different browser states
2. **Update AGENTS.md**:
   - Document new flags in "Build and Test Commands" section
   - Update "Code Style Guidelines" if needed
   - Add examples to relevant sections
3. **Update README.md**:
   - Add examples for --list-tabs
   - Add examples for --tab
   - Document use cases
   - Update flag reference
4. **Update docs/design-record.md**:
   - Mark Phase 2 as implemented
   - Document any design decisions made during implementation
   - Update feature comparison table

**Success Criteria**:
- All tests pass
- Documentation is complete and accurate
- Examples work as documented
- No regressions in Phase 1 functionality

## Testing Strategy

### Unit Tests

**New Test Files**: None needed (use existing test files)

**New Test Functions**:

**In `browser_test.go`**:

```go
func TestListTabs(t *testing.T)
func TestGetTabByIndex(t *testing.T)
func TestGetTabByPattern(t *testing.T)
func TestMatchURLPattern(t *testing.T)
```

**Test Coverage**:
- Happy path for each function
- Error cases (browser not connected, invalid index, no match)
- Edge cases (empty tab list, single tab, many tabs)
- Pattern matching edge cases

### Integration Tests

**In `cli_test.go`** (or new `tab_test.go`):

```go
func TestListTabsCommand(t *testing.T)
func TestTabByIndexCommand(t *testing.T)
func TestTabByPatternCommand(t *testing.T)
func TestTabFlagValidation(t *testing.T)
```

**Test Scenarios**:

1. **List tabs with browser running**:
   - Open 5 tabs with known URLs
   - Run `snag --list-tabs`
   - Verify output contains all 5 tabs
   - Verify correct indexing

2. **Fetch from tab by index**:
   - Open tab with known content
   - Run `snag --tab 0`
   - Verify correct content extracted
   - Verify Markdown conversion

3. **Fetch from tab by URL pattern**:
   - Open multiple tabs
   - Run `snag --tab "example.com"`
   - Verify correct tab matched

4. **Error cases**:
   - No browser running
   - Invalid index
   - No matching pattern
   - --tab with URL argument

### Manual Testing

**Test Scenarios**:

1. Real-world authenticated session:
   - Log into a website in Chrome
   - List tabs to find index
   - Fetch content using --tab
   - Verify authentication preserved

2. Multiple matching tabs:
   - Open 3 GitHub tabs
   - Use pattern `github.com/*`
   - Verify first tab matched

3. Complex URLs:
   - Tabs with query parameters
   - Tabs with fragments
   - Pattern matching with special chars

## Success Criteria

### Functionality

- ✅ **Phase 2.1 Complete**: `--list-tabs` lists all open tabs with correct information
- ⏳ **Phase 2.2 Next**: `--tab <index>` fetches content from specific tab
- ⏳ **Phase 2.3 Pending**: `--tab <pattern>` matches and fetches from tab by URL pattern
- ⏳ All existing flags work with --tab (--format, --output, etc.)
- ✅ Proper error handling with clear messages (Phase 2.1)
- ⏳ No new tabs created when using --tab

### Quality

- ✅ All unit tests pass (Phase 2.1)
- ✅ All integration tests pass (Phase 2.1)
- ✅ Code coverage maintained (Phase 2.1)
- ✅ No regressions in Phase 1 functionality
- ✅ Code follows project style guidelines (gofmt, vet)
- N/A MPL 2.0 headers on all new files (no new files created)

### Documentation

- ✅ AGENTS.md updated with new features (Phase 2.1)
- ⏳ README.md updated with examples (pending Phase 2 completion)
- ⏳ docs/design-record.md updated (pending Phase 2 completion)
- ✅ --help output includes new flags (Phase 2.1)
- ✅ Error messages are clear and actionable (Phase 2.1)

### Performance

- ✅ Listing tabs is fast (< 1 second) - Phase 2.1 validated
- ⏳ Tab selection is fast (< 100ms) - Phase 2.2/2.3
- ⏳ No memory leaks with repeated operations - Phase 2.2/2.3

## Risks and Mitigations

### Risk 1: Rod Pages() API Changes

**Risk**: The `browser.Pages()` API might behave differently than expected or change in future versions.

**Mitigation**:
- Thoroughly test with current rod version
- Add integration tests to catch API changes
- Pin rod version in go.mod
- Document rod version dependency in AGENTS.md

### Risk 2: Tab Ordering Changes

**Risk**: Browser might reorder tabs, making indexes unreliable.

**Mitigation**:
- Document that indexes are from current state
- Suggest using URL patterns for more reliable matching
- Show warning if tab count changes between list and fetch
- Consider adding tab ID support in future

### Risk 3: Pattern Matching Ambiguity

**Risk**: Wildcard patterns might match unexpected tabs.

**Mitigation**:
- Prefer exact matches over patterns
- Use first match consistently
- Clear documentation on pattern matching rules
- Verbose mode shows which tab matched
- Consider adding --tab-all for multiple matches in future

### Risk 4: Backward Compatibility

**Risk**: New flags might break existing workflows.

**Mitigation**:
- Make --tab and URL mutually exclusive (explicit error)
- Don't change behavior of existing flags
- Comprehensive testing of all flag combinations
- Document breaking changes if any

## Future Enhancements (Post Phase 2)

### Phase 2.5: Enhanced Tab Management

- `--tab-all <pattern>` - Fetch from all matching tabs
- `--list-tabs --format json` - JSON output for scripting
- Tab filtering: `--list-tabs --filter "github.com/*"`
- Sort options: `--list-tabs --sort-by title|url|index`

### Phase 3: Advanced Features

- `--close-tab <index>` - Close specific tab
- `--new-tab <url>` - Create new tab without fetching
- `--activate-tab <index>` - Bring tab to front
- `--tab-state` - Show tab loading state, memory usage

### Integration Ideas

- Export tabs to bookmarks
- Bulk operations on tabs
- Tab session management
- Screenshot specific tab

## Open Questions

### Q1: Should --list-tabs output go to stdout or stderr?

**Decision**: stdout

**Rationale**:
- Enables piping: `snag --list-tabs | grep github`
- Consistent with other listing commands (ls, ps, etc.)
- Logs still go to stderr as per design
- Content vs. information distinction is clear

### Q2: How to handle tab index changes between list and fetch?

**Options**:
1. Accept current state (indexes might shift)
2. Add warning if tab count changes
3. Use internal tab IDs instead of indexes

**Decision**: Option 1 (accept current state) for MVP

**Rationale**:
- Simpler implementation
- Document behavior clearly
- URL patterns available for more stable matching
- Can add warnings in future if needed

### Q3: Should --tab create a new tab if pattern doesn't match?

**Decision**: No - error if no match

**Rationale**:
- Explicit is better than implicit
- Avoids unexpected tab creation
- User can use regular `snag <url>` to create new tab
- Clear error message guides user

### Q4: Support regex patterns or just wildcards?

**Decision**: Full regex support (RESOLVED)

**Rationale**:
- Implementation complexity is the same (using regex internally anyway)
- Maximum flexibility for power users
- Simple users can still use basic patterns
- Document common patterns in README with examples
- Progressive fallthrough means simple patterns still work (substring match)
- One flag, simple mental model: "it's a pattern that tries everything"

## Timeline Estimate

**Phase 2.1** (Tab Listing): 4-6 hours
- Implementation: 2-3 hours
- Testing: 1-2 hours
- Documentation: 1 hour

**Phase 2.2** (Tab by Index): 4-6 hours
- Implementation: 2-3 hours
- Testing: 1-2 hours
- Documentation: 1 hour

**Phase 2.3** (Tab by Pattern): 6-8 hours
- Implementation: 3-4 hours
- Testing: 2-3 hours
- Documentation: 1 hour

**Phase 2.4** (Integration & Docs): 3-4 hours
- Integration testing: 1-2 hours
- Documentation updates: 1-2 hours
- Final review: 1 hour

**Total Estimate**: 17-24 hours

**Recommended Approach**: Implement in sequential phases (2.1 → 2.2 → 2.3 → 2.4) with testing and validation at each step.

## Implementation Checklist

### Phase 2.1: Tab Listing ✅ COMPLETE

- [x] Add `TabInfo` struct to `browser.go` (browser.go:49-55)
- [x] Implement `ListTabs()` function in `browser.go` (browser.go:404-434)
- [x] Add `--list-tabs` flag to `main.go` (main.go:130-134)
- [x] Implement `handleListTabs()` in `main.go` (main.go:345-383)
- [x] Add routing logic in `run()` function (main.go:190-193)
- [x] Add `ErrNoBrowserRunning` to `errors.go` (errors.go:34-35)
- [x] Write unit tests for `ListTabs()` - N/A (integration tests only)
- [x] Write integration test for `--list-tabs` command (cli_test.go:395-406, 1142-1170)
- [x] Manual testing with Chrome - PASSED
- [x] Update AGENTS.md with examples (AGENTS.md:58-62)

**Critical Bug Fix Applied**:
- Fixed browser.go:259-260 to **always set remote-debugging-port explicitly**
- Previously, when port was 9222 (default), rod launcher would pick random port
- This caused `--list-tabs` to fail connecting to browsers launched with `--force-visible`
- **Resolution**: Always set `remote-debugging-port` flag regardless of port value

**Test Results**:
- `TestCLI_ListTabsNoBrowser`: PASSED (validates error when no browser running)
- `TestBrowser_ListTabs`: PASSED (validates tab listing with real browser)

**Implementation Notes**:
- Tab indexes are 1-based for user display (converted from 0-based internally)
- Output goes to stdout for piping compatibility
- Requires existing browser with remote debugging enabled
- Will NOT launch new browser (connect-only mode)

### Phase 2.2: Tab Selection by Index

- [ ] Add `--tab` flag with `-t` alias to `main.go`
- [ ] Remove `-t` alias from `--timeout` flag in `main.go`
- [ ] Implement `GetTabByIndex()` in `browser.go` (1-based indexing)
- [ ] Implement `handleTabFetch()` in `main.go` (index support)
- [ ] Add flag validation for --tab and URL conflict
- [ ] Add `ErrTabIndexInvalid` and `ErrTabURLConflict` to `errors.go`
- [ ] Write unit tests for `GetTabByIndex()`
- [ ] Write integration tests for `--tab <index>`
- [ ] Test with all output formats and flags
- [ ] Test 1-based indexing (tab 1 = first tab)
- [ ] Manual testing with real tabs
- [ ] Update documentation

### Phase 2.3: Tab Selection by Pattern

- [ ] Implement `hasRegexChars()` helper in `browser.go`
- [ ] Implement `GetTabByPattern()` in `browser.go`
- [ ] Update `handleTabFetch()` to support patterns
- [ ] Add `ErrNoTabMatch` to `errors.go`
- [ ] Write unit tests for pattern matching
- [ ] Write integration tests for `--tab <pattern>`
- [ ] Test edge cases (regex patterns, case-insensitivity, fallthrough)
- [ ] Manual testing with various patterns
- [ ] Update documentation

### Phase 2.4: Integration & Documentation

- [ ] Run full test suite
- [ ] Test all flag combinations
- [ ] Verify no regressions in Phase 1
- [ ] Update AGENTS.md (comprehensive)
- [ ] Update README.md with examples
- [ ] Update docs/design-record.md
- [ ] Update --help output
- [ ] Run `go vet ./...`
- [ ] Run `gofmt -l .`
- [ ] Code review
- [ ] Final manual testing

## Notes

- **Browser Requirement**: --list-tabs and --tab require an existing Chrome instance with remote debugging enabled
- **No New Tabs**: Phase 2 features focus on working with existing tabs, not creating new ones
- **Backward Compatible**: All Phase 1 functionality remains unchanged
- **Future Extensible**: Design allows for future enhancements (multiple tabs, JSON output, etc.)

## References

- **Design Document**: docs/design-record.md (Phase 2 section)
- **Agent Documentation**: AGENTS.md
- **Rod Documentation**: github.com/go-rod/rod
- **Current Implementation**: browser.go, main.go, fetch.go

---

**Document Version**: 2.0
**Last Updated**: 2025-10-20
**Status**: Ready for Implementation

## Design Decisions

### Decision 1: Flag Assignment
**Move `-t` alias from `--timeout` to `--tab`**
- `--timeout` will have no short alias
- `--tab` gets `-t` (more frequently used flag deserves shorter alias)
- Rationale: Tab selection will be used far more often than custom timeouts

### Decision 2: Tab Indexing
**Use 1-based indexing (not 0-based)**
- First tab is `[1]`, not `[0]`
- More intuitive for end users (UI tool, not programming API)
- Internally convert to 0-based for arrays

### Decision 3: Pattern Matching
**Progressive fallthrough with full regex support**
1. Integer → Use as tab index (1-based)
2. Has regex chars (`* + ? [ ] { } ( ) | ^ $ \`) → Try regex match
3. Try exact URL match (case-insensitive)
4. Try substring/contains match (case-insensitive)
5. Error if no matches

### Decision 4: Case Sensitivity
**Case-insensitive matching for all modes**
- Regex: use `(?i)` flag
- Exact: use `strings.EqualFold()`
- Contains: convert both to lowercase
- Rationale: Better UX, URLs are typically lowercase but users might capitalize

### Decision 5: Regex Support
**Full regex support (not just wildcards)**
- Users can write full regex patterns
- Regex chars detected: `* + ? [ ] { } ( ) | ^ $ \`
- NOT treated as regex (URL-safe): `. / - _ : ? & = # %`
- Rationale: Same implementation complexity, maximum flexibility for power users

### Decision 6: Regex Fallthrough
**Always fall through to exact/contains even after trying regex**
- If regex chars detected but no match, try exact/contains anyway
- Catches edge cases, costs nothing (few string comparisons)
- More forgiving "try everything" approach

### Decision 7: Multiple Matches
**First match wins**
- Return first tab that matches pattern
- Users can use `--list-tabs` to see tab order
- Future: `--tab-all` for fetching from multiple tabs
