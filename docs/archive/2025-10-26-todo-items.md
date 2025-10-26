# Project TODO Items

**Status:** Complete
**Last Updated:** 2025-10-26

This document tracks outstanding implementation tasks identified during the argument handling analysis phase.

---

## Implementation Tasks

### 1. Argument Trimming

**Priority:** Medium
**Effort:** Low
**Status:** ✅ Completed (2025-10-24)

Apply `strings.TrimSpace()` to all string arguments after reading from CLI framework.

**Affected Arguments:**
- `--wait-for` ✅
- `--user-agent` ✅
- `--output` ✅
- `--output-dir` ✅
- `--tab` ✅
- `--url-file` ✅
- `<url>` positional arguments ✅
- `--user-data-dir` (not yet implemented in code)

**Rationale:**
- Standard behavior in most CLI tools (git, docker, etc.)
- Handles copy-paste trailing spaces gracefully
- Prevents user confusion from invisible whitespace

**Implementation Notes:**
- Applied `strings.TrimSpace()` immediately after reading flag value
- Trimming occurs before any validation or processing
- Added to both `main.go` and `handlers.go`
- URLs from files already trimmed in `loadURLsFromFile()` (validate.go:265)
- Positional URL arguments now trimmed in loop (main.go:206-208)

**Testing:**
- ✅ Verified `--output` with leading/trailing spaces
- ✅ Verified `--output-dir` with leading/trailing spaces
- ✅ Verified `--url-file` path trimming
- ✅ Verified URL arguments trimming
- ✅ All tests pass

---

### 2. Parallel Processing Strategy

**Priority:** High
**Effort:** High
**Status:** ❌ Not Required (2025-10-26)

Define and implement strategy for processing multiple URLs from various sources.

**Decision:** Sequential processing performance is adequate for current use cases. Tested with 30 URLs and performance is acceptable. Future optimization can be reconsidered if needed.

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

**Status:** ✅ Completed (2025-10-26)

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

**Implementation Notes:**
- Reviewed all 21 argument documentation files
- Found that only wait-for.md was missing `--user-data-dir` interaction
- Added `--user-data-dir` to wait-for.md's "Other Flag Interactions" section
- All other files (Tasks 1-6, 9, 15-17) already had proper documentation
- All files now consistently document `--user-data-dir` interactions per user-data-dir.md specification

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
**Status:** ✅ Completed (2025-10-24)

From docs/archive/2025-10-24-argument-documentation-cross-review.md:

### 5. Resolve Documentation Inconsistencies

**Phase 1: Critical Contradictions (7 issues)** - ✅ COMPLETE
1. `--list-tabs` + `--wait-for`: Changed to silently ignore ✅
2. `--close-tab` + `--url-file`: Changed to works normally ✅
3. `--tab` + `--open-browser`: Changed to warning/ignore ✅
4. `--all-tabs` + `--open-browser`: Changed to warning/ignore ✅
5. `--output` + `--open-browser`: Changed to warning/ignore ✅
6. `--tab` + `--user-data-dir`: Changed to warning/ignore ✅
7. `--all-tabs` + `--user-data-dir`: Changed to warning/ignore ✅

**Phase 2: Warning Message Standardization (7 issues)** - ✅ COMPLETE
8. `--user-agent` + `--tab`: Standardized ✅
9. `--user-agent` + `--all-tabs`: Standardized ✅
10. `--close-tab` + `--open-browser`: Standardized ✅
11. `--format` + `--open-browser`: Added warning ✅
12. `--timeout` + `--open-browser`: Added warning ✅
13. `--output-dir` + `--open-browser`: Added warning ✅
14. `--wait-for` + `--open-browser`: Standardized ✅

**Phase 3: Verification** - ✅ COMPLETE
- Re-verified all contradictions resolved ✅
- Verified examples sections match documented behavior ✅
- Updated validation.md compatibility matrices ✅
- Updated README.md compatibility matrices ✅

**Files Modified:** 11 files in `docs/arguments/`

See docs/archive/2025-10-24-argument-documentation-cross-review.md for complete details.

---

## Notes

- Tasks are independent unless noted
- Priority levels: High (affects functionality), Medium (affects UX/docs), Low (cleanup)
- Cross-reference with PROJECT-review.md for documentation fixes
- Update this file as tasks are completed or new tasks identified
