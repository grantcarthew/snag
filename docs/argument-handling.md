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
