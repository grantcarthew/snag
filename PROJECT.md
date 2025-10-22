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

### Task 7: `--wait-for SELECTOR` / `-w`

**Questions to answer:**
- What happens with wrong values? (invalid CSS selector, empty string)
- What happens when combined with:
  - `<url>` (single)
  - `<url>` (multiple)
  - `--url-file`
  - `--output` / `-o`
  - `--output-dir` / `-d`
  - `--format` / `-f`
  - `--timeout`
  - Another `--wait-for`
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

**Define:** Timeout interaction, behavior with existing tabs

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
- [ ] Task 5: `--format` / `-f`
- [ ] Task 6: `--timeout`
- [ ] Task 7: `--wait-for` / `-w`
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
