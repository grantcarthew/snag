# Project TODO Items

**Status:** Active
**Last Updated:** 2025-10-23

This document tracks outstanding implementation tasks identified during the argument handling analysis phase.

---

## Implementation Tasks

### 1. Argument Trimming

**Priority:** Medium
**Effort:** Low

Apply `strings.TrimSpace()` to all string arguments after reading from CLI framework.

**Affected Arguments:**
- `--wait-for`
- `--user-agent`
- `--output`
- `--output-dir`
- `--tab`
- `--url-file`
- `--user-data-dir`

**Rationale:**
- Standard behavior in most CLI tools (git, docker, etc.)
- Handles copy-paste trailing spaces gracefully
- Prevents user confusion from invisible whitespace

**Implementation Notes:**
- Apply trimming immediately after reading flag value
- Before any validation or processing
- Documented in multiple argument files as expected behavior

---

### 2. Parallel Processing Strategy

**Priority:** High
**Effort:** High

Define and implement strategy for processing multiple URLs from various sources.

**Affected Operations:**
- Multiple `<url>` arguments
- `--url-file` with multiple URLs
- `--all-tabs` processing

**Options to Evaluate:**

1. **Sequential** (current behavior):
   - Process URLs one-by-one
   - Predictable order
   - Lower resource usage
   - Slower for large batches

2. **Parallel**:
   - Process URLs concurrently with goroutines
   - Faster for large batches
   - Higher resource usage
   - May overwhelm browser

3. **Hybrid**:
   - Parallel with configurable concurrency limit
   - Balance between speed and resource usage
   - Example: Process 5 URLs at a time

**Considerations:**
- Browser resource usage and stability
- Tab creation/closing order
- Error handling across goroutines
- Output file ordering/naming
- `--close-tab` behavior with concurrent processing
- Progress logging with concurrent operations

**Related Documentation:**
- close-tab.md line 120: "Parallel Processing Note"
- Multiple files assume sequential processing currently

---

### 3. Review `--user-data-dir` Interactions

**Priority:** Medium
**Effort:** Medium

**Status:** Partially complete (reviewed during cross-review)

Review all 21 completed argument analysis tasks to ensure `--user-data-dir` flag behavior is properly documented in each `docs/arguments/[argument].md` file.

**Background:**
- `--user-data-dir` flag was added after initial tasks (1-9) were designed
- Needs to be retroactively documented for consistency

**Tasks to Review:**
- [x] Task 1: `<url>` - Not covered (should add)
- [x] Task 2: `--url-file` - Not covered (should add)
- [x] Task 3: `--output` / `-o` - Not covered (should add)
- [x] Task 4: `--output-dir` / `-d` - Not covered (should add)
- [x] Task 5: `--format` / `-f` - Not covered (should add)
- [x] Task 6: `--timeout` - Not covered (should add)
- [x] Task 7: `--wait-for` / `-w` - Not covered (should add)
- [x] Task 8: `--port` / `-p` - Covered ✓
- [x] Task 9: `--close-tab` / `-c` - Not covered (should add)
- [x] Task 10: `--force-headless` - Covered ✓
- [x] Task 11: `--open-browser` / `-b` - Covered ✓
- [x] Task 12: `--list-tabs` / `-l` - Covered ✓
- [x] Task 13: `--tab` / `-t` - Covered ✓ (inconsistency found, see PROJECT-review.md)
- [x] Task 14: `--all-tabs` / `-a` - Covered ✓ (inconsistency found, see PROJECT-review.md)
- [x] Task 15: `--verbose` - Not covered (should add)
- [x] Task 16: `--quiet` / `-q` - Not covered (should add)
- [x] Task 17: `--debug` - Not covered (should add)
- [x] Task 18: `--user-agent` - Covered ✓
- [x] Task 19: `--help` / `-h` - Covered ✓
- [x] Task 20: `--version` / `-v` - Covered ✓
- [x] Task 21: `--user-data-dir` - N/A (self)

**Action Required:**
- Add `--user-data-dir` interaction sections to Tasks 1-7, 9, 15-17
- Standard interaction: "Works normally" for most flags
- See user-data-dir.md for expected interactions

---

### 4. Remove `--force-visible` Flag

**Priority:** Low
**Effort:** Low
**Status:** ✅ Completed (2025-10-24)

Remove the deprecated `--force-visible` flag from the codebase.

**Tasks:**
- [x] Remove flag definition from `main.go`
- [x] Remove all references and logic in `browser.go`
- [x] Remove validation logic and error messages
- [x] Update any tests that reference this flag
- [x] Ensure browser mode logic works correctly without this flag
- [x] Verify no documentation references remain

**Background:**
- This flag was mentioned in early design discussions
- Was replaced by `--open-browser` flag
- Successfully removed from codebase

**Implementation Notes:**
- Removed `ForceVisible` field from `Config` and `BrowserOptions` structs
- Updated browser mode logic to use `!openBrowser` instead of `!forceVisible`
- Updated error messages in `fetch.go` to suggest `--open-browser` instead
- Renamed test `TestBrowser_ForceVisible` to `TestBrowser_OpenBrowserWithCloseTab`
- All browser mode tests pass successfully

---

## Cross-Review Follow-up Tasks

**Priority:** High
**Effort:** Medium

From PROJECT-review.md (2025-10-23 review):

### 5. Resolve Documentation Inconsistencies

**Critical Contradictions (7 issues):**
1. `--list-tabs` + `--wait-for`: Change to silently ignore
2. `--close-tab` + `--url-file`: Change to works normally
3. `--tab` + `--open-browser`: Change to error
4. `--all-tabs` + `--open-browser`: Change to error
5. `--output` + `--open-browser`: Change to warning
6. `--tab` + `--user-data-dir`: Change to warning/ignore
7. `--all-tabs` + `--user-data-dir`: Change to warning/ignore

**Minor Inconsistencies (7 issues):**
8. Standardize warning message wording across all affected files

**Files to Update:** 11 files in `docs/arguments/`

See PROJECT-review.md for complete details and recommendations.

---

## Notes

- Tasks are independent unless noted
- Priority levels: High (affects functionality), Medium (affects UX/docs), Low (cleanup)
- Cross-reference with PROJECT-review.md for documentation fixes
- Update this file as tasks are completed or new tasks identified
