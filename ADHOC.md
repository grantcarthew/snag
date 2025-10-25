# ADHOC: Tab UX Improvements

## Overview

Two quality-of-life improvements for tab management:
1. **Improved --list-tabs display**: Clean, readable output with truncation and query param stripping
2. **Pattern matching enhancement**: --tab <pattern> saves all matches instead of first match only

## Task 1: Improve --list-tabs Display Format

### Current Problem

Tab listings show full URLs with query parameters, making output difficult to read:

```
Available tabs in browser (7 tabs, sorted by URL):
  [4] https://www.ato.gov.au/about-ato/contact-us?gclsrc=aw.ds&gad_source=1&gad_campaignid=22717615122&gbraid=0AAAAAolERwRiHRb8KrrNHk7GbTWG6FA0E&gclid=EAIaIQobChMImPuotYK-kAMVJ6hmAh10QhjZEAAYASAAEgKG6_D_BwE - Contact us | Australian Taxation Office
  [6] https://www.google.com/search?gs_ssp=eJzj4tTP1TewzEouKzZg9GKvzC8tKU1KBQA_-AaN&q=youtube&oq=tyou&gs_lcrp=EgZjaHJvbWUqDwgBEC4YChiDARixAxiABDIGCAAQRRg5Mg8IARAuGAoYgwEYsQMYgAQyDwgCEAAYChiDARixAxiABDIPCAMQABgKGIMBGLEDGIAEMhUIBBAuGAoYgwEYxwEYsQMY0QMYgAQyDwgFEAAYChiDARixAxiABDIMCAYQABgKGLEDGIAEMgYIBxAFGEDSAQgyMzYxajBqN6gCB7ACAfEF13V1Q3FsVZ4&sourceid=chrome&ie=UTF-8&sei=ow_8aIvBO4uMseMPmrudyAk - youtube - Google Search
```

### Proposed Solution

**Format**: `[N] Title (domain/path)`

**Rules**:
- Strip all query parameters (`?...`) and hash fragments (`#...`) from displayed URL
- Show title first (more distinctive than URL)
- Show domain/path in parentheses (clean, scannable)
- If title is empty/missing: omit title entirely → `[N] (domain/path)` (avoids multiple spaces)
- Truncate total line to **120 characters maximum** with `...`
  - **Layout**: `  [NNN] Title (domain/path)` = ~8 chars prefix + title + URL
  - **URL limit**: Maximum 80 chars (including parentheses)
  - **Title space**: Remainder (typically 30-70 chars depending on URL length)
  - **Truncation priority**: Keep URL complete up to 80 chars, title uses remaining space
- **Verbose mode**: `--list-tabs --verbose` shows full URLs and titles (no truncation, includes query params)

**Example Output**:

```
Available tabs in browser (7 tabs, sorted by URL):
  [1] New tab (chrome://newtab)
  [2] Example Domain (example.com)
  [3] Order in browser.pages() · Issue #7452 (github.com/puppeteer/puppeteer/issues/7452)
  [4] Contact us | Australian Taxation Office (ato.gov.au/about-ato/contact-us)
  [5] BIG W | How good's that (bigw.com.au)
  [6] youtube - Google Search (google.com/search)
  [7] X. It's what's happening / X (x.com)
```

**With --verbose**:

```
Available tabs in browser (7 tabs, sorted by URL):
  [4] https://www.ato.gov.au/about-ato/contact-us?gclsrc=aw.ds&gad_source=1&... - Contact us | Australian Taxation Office
  [6] https://www.google.com/search?gs_ssp=eJzj4tTP1TewzEouKzZg9... - youtube - Google Search
```

### Implementation Tasks

- [ ] Add helper function to strip query params and hash from URL (e.g., `stripURLParams(url string) string`)
- [ ] Add helper function to format tab display line with truncation (e.g., `formatTabLine(index, title, url, maxLength, verbose)`)
- [ ] Update `displayTabList()` in handlers.go to use new formatting
- [ ] Support `--verbose` flag to show full URLs
- [ ] Update `displayTabListOnError()` to use same formatting
- [ ] Test with various URL types (with/without params, long/short titles)

### Documentation Updates

- [ ] **AGENTS.md**: Update tab listing examples to show new format
- [ ] **README.md**: Update `--list-tabs` examples (2 locations)
- [ ] **docs/arguments/list-tabs.md**: Update output format section
- [ ] **docs/arguments/verbose.md**: Document that --verbose shows full URLs in --list-tabs
- [ ] **docs/design-record.md**: Add design decision for display format (truncation rules, format choice)

### Edge Cases to Handle

- URLs without paths: `https://example.com` → `(example.com)`
- Empty/missing titles: Omit title entirely (no double spaces)
- Very long titles: Truncate after URL takes its space (up to 80 chars)
- Very long URLs: Truncate at 80 chars with `...`
- Multiple tabs with same domain/path but different params: Title should distinguish them
- Chrome internal URLs: `chrome://newtab/`, `chrome://settings/`, etc. (typically short)

### Testing

- [ ] Test with 0 tabs, 1 tab, many tabs
- [ ] Test with long URLs (>120 chars)
- [ ] Test with short URLs (<120 chars)
- [ ] Test with URLs containing query params
- [ ] Test with URLs containing hash fragments
- [ ] Test with --verbose flag
- [ ] Test that pattern matching still works with full URLs (not truncated display)

---

## Task 2: Pattern Matching - Save All Matches

### Current Problem

`--tab <pattern>` only returns the **first** matching tab and exits. If multiple tabs match a pattern, the others are ignored.

**Example**:
```bash
# Browser has 3 GitHub tabs:
#   [2] github.com/user/repo1
#   [5] github.com/user/repo2
#   [8] github.com/org/repo3

snag --tab "github"
# Only fetches tab [2], ignores [5] and [8]
```

### Proposed Solution

When `--tab <pattern>` matches **multiple tabs**, save all matches (like `--all-tabs` but filtered).

**Behavior**:
- **Single match**: Fetch and output to stdout (current behavior, unchanged)
- **Multiple matches**: Fetch all, auto-save with generated filenames (like `--all-tabs`)
  - No confirmation prompt - auto-proceed with good logging (consistent with `--all-tabs`)
  - Process in same sort order as `--list-tabs` (alphabetically by URL)
- **No matches**: Error with tab listing (current behavior, unchanged)

**Examples**:

```bash
# Single match - outputs to stdout (unchanged)
snag --tab "example.com"
# Output: Markdown content to stdout

# Multiple matches - auto-saves all matches
snag --tab "github"
# Processing 3 tabs matching pattern 'github'...
# [1/3] Processing tab [2]: github.com/user/repo1
# ✓ Saved to 2025-10-25-123456-repo1.md
# [2/3] Processing tab [5]: github.com/user/repo2
# ✓ Saved to 2025-10-25-123456-repo2.md
# [3/3] Processing tab [8]: github.com/org/repo3
# ✓ Saved to 2025-10-25-123456-repo3.md
# ✓ Batch complete: 3 succeeded, 0 failed

# With output directory
snag --tab "github" -d ./repos/
# Saves all matches to ./repos/ directory

# Conflict: Cannot use --output with multiple matches
snag --tab "github" -o output.md
# ERROR: Cannot use --output with multiple tabs. Use --output-dir instead
```

### Implementation Tasks

- [ ] Modify `GetTabByPattern()` to return `[]*rod.Page` instead of single `*rod.Page`
  - Or create new `GetTabsByPattern()` function
  - Return all matching tabs, not just first
- [ ] Update `handleTabFetch()` to detect single vs. multiple matches
  - Single match: Current behavior (stdout or -o file)
  - Multiple matches: Call batch processing like `handleTabRange()`
- [ ] Add validation: Error if `--output` is set with multiple matches
- [ ] Update success logging to indicate "pattern matched N tabs"
- [ ] Ensure error messages list all matched tabs on errors

### Documentation Updates

- [ ] **AGENTS.md**: Update pattern matching section to document multi-match behavior
- [ ] **README.md**: Update `--tab <pattern>` examples to show multi-match scenarios
- [ ] **docs/arguments/tab.md**:
  - Update "Multiple Matches" section from "First match wins" to "All matches processed"
  - Add multi-match examples
  - Update output validation rules (cannot use -o with multi-match)
- [ ] **docs/design-record.md**: Update Design Decision #21 "Multiple Matches" with new behavior and rationale

### Breaking Change Considerations

**Current behavior**: First match wins (predictable, simple)
**New behavior**: All matches processed (more useful, intuitive)

**Impact**: Tool is not in production - no breaking change concerns

**Rationale for change**:
- **User expectation**: When using a pattern, users likely want all matching tabs
- **Consistency**: Aligns with `--all-tabs` behavior (batch processing)
- **Workflow improvement**: Batch download all tabs matching a pattern (e.g., all GitHub repos)
- **Still predictable**: Single match = stdout, multiple = auto-save (clear, documented)

**Migration path** (if needed in future):
- Users relying on "first match" behavior can use more specific patterns
- Or use `--tab 1` for exact index selection

### Edge Cases to Handle

- Pattern matches 1 tab with `--output`: Should work (single match)
- Pattern matches 2+ tabs with `--output`: Error (use `--output-dir`)
- Pattern matches 0 tabs: Error with tab listing (current behavior)
- Pattern matches all tabs: Same as `--all-tabs` (acceptable)
- Regex pattern matching multiple tabs: Process all matches

### Testing

- [ ] Test pattern matching 1 tab (stdout behavior)
- [ ] Test pattern matching multiple tabs (batch processing)
- [ ] Test pattern matching 0 tabs (error)
- [ ] Test with --output flag (error on multiple matches)
- [ ] Test with --output-dir flag (saves all matches)
- [ ] Test with various formats (md, html, pdf, png)
- [ ] Test with --wait-for selector (applies to all matches)

---

## Success Criteria

### Task 1 (Display Format)
- [ ] Tab listings are readable and scannable
- [ ] Query parameters stripped from display
- [ ] Lines truncated to 120 characters max
- [ ] `--verbose` shows full URLs and titles
- [ ] All documentation updated consistently

### Task 2 (Multi-match)
- [ ] Single pattern match outputs to stdout (backward compatible)
- [ ] Multiple pattern matches auto-save all tabs
- [ ] Clear error messages for `--output` conflicts
- [ ] All documentation updated with new behavior
- [ ] Design decision documented with rationale

## Notes

- Both tasks improve tab UX without adding new flags (opinionated design)
- Task 1 is pure display/UX (no behavioral changes)
- Task 2 changes behavior but tool is not in production (no breaking change concerns)
- Full documentation sweep required for both (AGENTS.md, README.md, docs/arguments/, docs/design-record.md)
- CommonMark compliance can be handled after implementation
- Test with real browser tabs to verify UX improvements
