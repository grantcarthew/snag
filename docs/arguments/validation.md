# Validation Rules and Special Cases

**Last Updated:** 2025-10-24

This document describes validation order, special cases, edge cases, and the implementation checklist for argument handling.

---

## General Validation Rules

**String Argument Trimming:**

All string arguments are trimmed using `strings.TrimSpace()` after reading from CLI framework:
- Removes leading and trailing whitespace (spaces, tabs, newlines)
- Applied to: `--output`, `--output-dir`, `--format`, `--wait-for`, `--user-agent`, `--user-data-dir`, `--tab`, `--url-file`
- Empty strings after trimming are handled per-argument (usually warning + ignored or error)
- Standard behavior in most CLI tools (git, docker, etc.)

---

## Validation Order

**Current implementation order (main.go:178-316):**

1. Initialize logger (`--quiet`, `--verbose`, `--debug`)
2. Handle `--open-browser` without URL (exit early)
3. Handle `--list-tabs` (extract --port and logging flags, ignore all others, list tabs, exit early)
4. Handle `--all-tabs` (check for URL conflict, exit early)
5. Handle `--tab` (check for URL conflict, exit early)
6. Validate URL argument required
7. Validate URL format
8. Validate `-o` + `-d` conflict
9. Validate format
10. Validate timeout
11. Validate port
12. Validate output path (if `-o`)
13. Validate output directory (if `-d`)
14. Execute fetch operation

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

### Case 2: --open-browser Does Not Fetch Content

**Behavior:**

```bash
snag --open-browser https://example.com
# Opens URL in browser tab, NO content fetching
# Exits snag immediately, browser stays open
# To fetch content: snag --tab 1
```

**Rationale:**
- `--open-browser` is for launching and navigating browser only
- Consistent with "open and exit" behavior
- Content fetching requires separate operation

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

**Decision:**
- `--force-headless` → **Error** (tabs require existing browser)

**Rationale:**
- Tabs require visible browser with remote debugging
- `--force-headless` conflicts with this requirement → Error

### Case 6: --user-agent with Tab Features

**Question:** What happens with `snag --tab 1 --user-agent "Custom"`?

**Decision:** **Warn and ignore** - Tab already loaded with its own user agent

**Rationale:**
- Tab already open in browser with established user agent
- Cannot change user agent for existing page
- Warn rather than error (flag has no effect but doesn't break operation)
- Applies to both `--tab` and `--all-tabs`

### Case 7: Multiple Logging Flags

**Question:** What happens with `--quiet --verbose`?

**Decision:** Last flag wins (Unix standard)

**Priority order:** `--debug` > `--verbose` > `--quiet` > normal

**Implementation:** Last matching flag wins (main.go:181-187)

### Case 8: --all-tabs with -o

**Question:** What happens with `snag --all-tabs -o output.md`?

**Expected:** Should ERROR with "Use --output-dir for multiple outputs"

**Status:** Needs validation

### Case 9: Zero URLs with --url-file

**Behavior:** Error if no valid URLs found

```bash
snag --url-file empty.txt
# ERROR: "No valid URLs found in file"
```

**Implementation:** Validation in `validate.go:loadURLsFromFile()`

### Case 10: --open-browser + --force-headless

**Question:** Conflicting browser modes - which wins?

**Decision:** Should ERROR (conflicting intent)

**Error message:** `"Cannot use both --force-headless and --open-browser (conflicting modes)"`

### Case 11: --user-data-dir with Existing Browser

**Question:** What happens with `snag --tab 1 --user-data-dir ~/.snag/profile`?

**Decision:** **Warn and ignore** - Cannot change profile of already-running browser

**Rationale:**
- Tab operations connect to existing browser instance
- Browser profile is set at launch time, cannot be changed mid-session
- Warn rather than error (flag has no effect but doesn't break operation)
- Applies to `--tab`, `--all-tabs`, and `--list-tabs`

**Warning message:** `"Warning: --user-data-dir ignored when connecting to existing browser"`

---

## Compatibility Table: Complete Flag Matrix

**Legend:**

- ✅ Compatible
- ❌ Error (mutually exclusive)
- `-` Not applicable

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
| **--force-headless** | ✅     | ✅          | -                | ❌             |
| **--open-browser**   | ✅     | ✅          | ❌               | -              |

### Tab Feature Flags

|                 | --list-tabs | --tab   | --all-tabs |
| --------------- | ----------- | ------- | ---------- |
| **--list-tabs** | Ignores     | Ignores | Ignores    |
| **--tab**       | Ignored     | -       | ❌         |
| **--all-tabs**  | Ignored     | ❌      | -          |

### Arguments with Tab Features

|                 | \<url\> | Multiple URLs | --url-file |
| --------------- | ------- | ------------- | ---------- |
| **--list-tabs** | Ignores | Ignores       | Ignores    |
| **--tab**       | ❌      | ❌            | ❌         |
| **--all-tabs**  | ❌      | ❌            | ❌         |

### Arguments with Output Flags

|        | \<url\> | Multiple URLs |
| ------ | ------- | ------------- |
| **-o** | ✅      | ❌            |
| **-d** | ✅      | ✅            |

### Logging Flags ✅

All logging flag conflicts resolved using "last flag wins" approach (Unix standard):

|               | --verbose    | --quiet      | --debug      |
| ------------- | ------------ | ------------ | ------------ |
| **--verbose** | -            | ✅ Last wins | ✅ Last wins |
| **--quiet**   | ✅ Last wins | -            | ✅ Last wins |
| **--debug**   | ✅ Last wins | ✅ Last wins | -            |

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

- [x] `--tab` + `--all-tabs` → Should ERROR ✅
- [x] `--all-tabs` + `-o` → Should ERROR ✅
- [x] `--open-browser` + `--force-headless` → Should ERROR ✅
- [x] Multiple logging flags → Last flag wins (Unix standard) ✅

### Additional Validations

- [ ] Multiple URLs + `-o` → ERROR
- [ ] `--url-file` + `--tab` → ERROR
- [ ] `--url-file` + `--all-tabs` → ERROR
- [ ] `--url-file` file not found → ERROR
- [ ] `--url-file` no valid URLs → ERROR
