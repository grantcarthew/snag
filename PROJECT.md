# Argument Documentation Cross-Review Results

**Review Date:** 2025-10-23
**Reviewer:** Claude Code (Sonnet 4.5)
**Task:** Cross-review all argument documentation in `docs/arguments/` for inconsistencies

---

## Executive Summary

Completed a comprehensive cross-review of all 21 argument documentation files plus README.md and validation.md. Found **7 critical contradictions** where the same flag combination is documented differently in the two relevant files, and **7 minor inconsistencies** in warning message wording.

**Critical Issues:** These are cases where one document says "error" and the other says "works" or "warning" - fundamentally incompatible behaviors that must be resolved.

**Minor Issues:** These are cases where both documents agree on the behavior but use different wording for warning messages or descriptions.

---

## Critical Contradictions (Require Resolution)

### 1. `--list-tabs` + `--wait-for`: Error vs Silently Ignored

**Contradiction:**
- **wait-for.md (line 51)** says: **Error** with message "Cannot use --wait-for with --list-tabs (no content fetching)"
- **list-tabs.md (line 66)** says: **Flag ignored** - Lists tabs (no content fetch)

**Context:**
- `--list-tabs` is documented as a standalone mode like `--help` and `--version`
- list-tabs.md lines 18-19 explicitly say it "acts like `--help` or `--version`: Overrides all other flags except those needed for its operation"
- list-tabs.md lines 54-73 show comprehensive list of flags that are "SILENTLY IGNORED"

**Recommendation:**
- **Silently ignore** (no error)
- Rationale: `--list-tabs` is standalone mode, should ignore all flags except `--port` and logging flags
- Update wait-for.md line 51 to change from "Error" to "Silently ignored"
- Update wait-for.md line 54 to remove the error message

**Files to Update:**
- `docs/arguments/wait-for.md` (lines 51-54)

---

### 2. `--close-tab` + `--url-file`: Works Normally vs Error

**Contradiction:**
- **close-tab.md (line 36)** says: **Works normally** - "Close each tab after fetching"
- **url-file.md (line 123)** says: **Error** - "Ambiguous for batch operations"

**Context:**
- close-tab.md says it works with `--url-file` normally
- url-file.md explicitly lists it as an error in the "Special Behaviors" section
- Both documents agree that `--close-tab` + multiple `<url>` arguments works normally (close-tab.md line 35, url-file.md doesn't cover this since multiple URLs are planned)

**Recommendation:**
- **Works normally** (no error)
- Rationale: Consistent with multiple URLs behavior; close each tab after fetch in batch operations
- Note: PROJECT.md TODO line 902-907 mentions "parallel processing strategy" needs definition for multiple URLs, which affects close-tab behavior
- Update url-file.md line 123 to change from "Error" to "Works normally"

**Files to Update:**
- `docs/arguments/url-file.md` (lines 123-124)

---

### 3. `--tab` + `--open-browser`: Warning vs Error

**Contradiction:**
- **open-browser.md (line 60)** says: **Warning** with "Warning: --tab ignored with --open-browser (no content fetching)"
- **tab.md (line 82)** says: **Error** with "Cannot use both --tab and --open-browser (conflicting purposes)"

**Context:**
- open-browser.md shows tab operations warned and ignored
- tab.md shows it as a mutually exclusive error in the Browser Mode Conflicts section
- This is a fundamental design decision: does `--open-browser` override everything, or do they conflict?

**Recommendation:**
- **Error** (mutually exclusive)
- Rationale: `--tab` and `--open-browser` have fundamentally different purposes (fetch from existing tab vs launch browser); user intent is unclear when both specified
- This aligns with the pattern that conflicting content sources error rather than warn
- Update open-browser.md lines 60-61 to change from "Warning" to "Error"
- Add error message: "Cannot use both --tab and --open-browser (conflicting purposes)"

**Files to Update:**
- `docs/arguments/open-browser.md` (lines 60-61, and line 160 in examples)

---

### 4. `--all-tabs` + `--open-browser`: Warning vs Error

**Contradiction:**
- **open-browser.md (line 61)** says: **Warning** with "Warning: --all-tabs ignored with --open-browser (no content fetching)"
- **all-tabs.md (line 75)** says: **Error** with "Cannot use both --all-tabs and --open-browser (conflicting purposes)"

**Context:**
- Same issue as #3 above, but for `--all-tabs`
- Same design decision needed

**Recommendation:**
- **Error** (mutually exclusive)
- Rationale: Same as #3 - conflicting purposes, unclear user intent
- Update open-browser.md line 61 to change from "Warning" to "Error"
- Add error message: "Cannot use both --all-tabs and --open-browser (conflicting purposes)"

**Files to Update:**
- `docs/arguments/open-browser.md` (lines 61-62, and line 161 in examples)

---

### 5. `--output` + `--open-browser` (no URL): Error vs Warning

**Contradiction:**
- **output.md (line 95)** says: **Error** with "Cannot use --output without content source (URL or --tab)"
- **open-browser.md (line 75)** says: **Warning**, flag ignored with "Warning: --output ignored with --open-browser (no content fetching)"

**Context:**
- output.md treats lack of content source as an error
- open-browser.md treats all output flags as warnings (they're ignored because no fetching)
- This is about whether `--open-browser` (no URL) is considered "having no content source" or is a valid mode that ignores output flags

**Recommendation:**
- **Warning** (flag ignored)
- Rationale: `--open-browser` is a valid operation mode (launch browser), just doesn't fetch content; better UX to warn than error
- Consistent with open-browser.md's treatment of all output/timing flags as warnings
- Update output.md line 95 to change from "Error" to "Flag ignored"
- Update output.md line 100 to change error message to warning: "Warning: --output ignored with --open-browser (no content fetching)"

**Files to Update:**
- `docs/arguments/output.md` (lines 95-100)

---

### 6. `--tab` + `--user-data-dir`: Warning/Ignore vs Works Normally

**Contradiction:**
- **user-data-dir.md (line 97)** says: **Warning**, ignore flag with "Warning: Ignoring --user-data-dir (connecting to existing browser for tab operations)"
- **tab.md (line 108)** says: **Works normally** - "Connects to browser using specified profile"

**Context:**
- user-data-dir.md says tab operations connect to existing browser, so profile flag is ignored
- tab.md says it works normally to connect to browser using specified profile
- This is about whether you can specify which profile to connect to when using `--tab`

**Recommendation:**
- **Warning + Ignore**
- Rationale: When you use `--tab`, you're connecting to an ALREADY RUNNING browser instance that already has its profile loaded; you can't change the profile of a running browser
- However, if you're using `--tab` with `--port` to connect to a specific browser instance, the user might reasonably specify `--user-data-dir` to document which profile that browser is using
- Best UX: Warn that the flag has no effect (browser already running with its profile)
- Update tab.md line 108 to change from "Works normally" to "Warning, ignored"
- Add note: "Warning: --user-data-dir ignored when connecting to existing browser"

**Files to Update:**
- `docs/arguments/tab.md` (line 108)

---

### 7. `--all-tabs` + `--user-data-dir`: Warning/Ignore vs Works Normally

**Contradiction:**
- **user-data-dir.md (line 98)** says: **Warning**, ignore flag with "Warning: Ignoring --user-data-dir (connecting to existing browser for tab operations)"
- **all-tabs.md (line 101)** says: **Works normally** - "Connects to browser using specified profile"

**Context:**
- Same issue as #6 above, but for `--all-tabs`
- Same reasoning applies

**Recommendation:**
- **Warning + Ignore**
- Rationale: Same as #6 - connecting to existing browser with existing profile
- Update all-tabs.md line 101 to change from "Works normally" to "Warning, ignored"
- Add note: "Warning: --user-data-dir ignored when connecting to existing browser"

**Files to Update:**
- `docs/arguments/all-tabs.md` (line 101)

---

## Minor Inconsistencies (Warning Message Wording)

These are cases where both documents agree on the behavior (warning/ignore) but use different wording in the warning messages or descriptions.

### 8. `--user-agent` + `--tab`: Different Warning Message Wording

**Inconsistency:**
- **user-agent.md (line 59)**: "Warning: --user-agent has no effect with --tab (tab already has its user agent)"
- **tab.md (line 107)**: "Warning: --user-agent has no effect with --tab (cannot change existing tab's user agent)"

**Recommendation:**
- Standardize on one message
- Suggested: "Warning: --user-agent has no effect with --tab (cannot change existing tab's user agent)"
- Rationale: More explicit about WHY (can't change) vs just stating the fact (already has)

**Files to Update:**
- `docs/arguments/user-agent.md` (line 59)

---

### 9. `--user-agent` + `--all-tabs`: Different Warning Message Wording

**Inconsistency:**
- **user-agent.md (line 60)**: "tabs already have their user agents"
- **all-tabs.md (line 100)**: "cannot change existing tabs' user agents"

**Recommendation:**
- Standardize on one message
- Suggested: "Warning: --user-agent has no effect with --all-tabs (cannot change existing tabs' user agents)"
- Rationale: Consistent with #8 recommendation

**Files to Update:**
- `docs/arguments/user-agent.md` (line 60)

---

### 10. `--close-tab` + `--open-browser`: Different Warning Reason Wording

**Inconsistency:**
- **close-tab.md (line 40)**: "Warning: --close-tab has no effect with --open-browser (no content to close)"
- **open-browser.md (line 91)**: "Warning: --close-tab ignored with --open-browser (no content fetching)"

**Recommendation:**
- Standardize on one reason
- Suggested: "Warning: --close-tab ignored with --open-browser (no content fetching)"
- Rationale: Consistent with other open-browser warnings about "no content fetching"

**Files to Update:**
- `docs/arguments/close-tab.md` (line 40)

---

### 11. `--format` + `--open-browser`: Missing Warning Mention

**Inconsistency:**
- **format.md (line 106)**: Says format is "ignored" but doesn't mention warning
- **open-browser.md (line 77)**: Says "Warning, flag ignored" with explicit warning message

**Recommendation:**
- Add warning to format.md
- Suggested: "**Warning**, flag ignored" with message "Warning: --format ignored with --open-browser (no content fetching)"

**Files to Update:**
- `docs/arguments/format.md` (line 106-107)

---

### 12. `--timeout` + `--open-browser`: Missing Warning Mention

**Inconsistency:**
- **timeout.md (line 76)**: Says timeout is "ignored" but doesn't mention warning
- **open-browser.md (line 83)**: Says "Warning, flag ignored" with explicit warning message

**Recommendation:**
- Add warning to timeout.md
- Suggested: "**Warning**, flag ignored" with message "Warning: --timeout ignored with --open-browser (no content fetching)"

**Files to Update:**
- `docs/arguments/timeout.md` (line 76-77)

---

### 13. `--output-dir` + `--open-browser`: Missing Warning Mention

**Inconsistency:**
- **output-dir.md (line 92)**: Says `-d` is "ignored" but doesn't mention warning
- **open-browser.md (line 76)**: Says "Warning, flag ignored" with explicit warning message

**Recommendation:**
- Add warning to output-dir.md
- Suggested: "**Warning**, flag ignored" with message "Warning: --output-dir ignored with --open-browser (no content fetching)"

**Files to Update:**
- `docs/arguments/output-dir.md` (line 92-93)

---

### 14. `--wait-for` + `--open-browser`: Phrasing Difference

**Inconsistency:**
- **wait-for.md (line 82)**: "has no effect with --open-browser"
- **open-browser.md (line 84)**: "ignored with --open-browser"

**Recommendation:**
- Standardize on "ignored with" phrasing
- Suggested: "Warning, flag ignored" (consistent with other open-browser warnings)

**Files to Update:**
- `docs/arguments/wait-for.md` (lines 78-82)

---

## Files Requiring Updates

Summary of all files that need changes:

1. **docs/arguments/wait-for.md** - Issues #1, #14
2. **docs/arguments/url-file.md** - Issue #2
3. **docs/arguments/open-browser.md** - Issues #3, #4
4. **docs/arguments/output.md** - Issue #5
5. **docs/arguments/tab.md** - Issues #3, #6
6. **docs/arguments/all-tabs.md** - Issues #4, #7
7. **docs/arguments/user-agent.md** - Issues #8, #9
8. **docs/arguments/close-tab.md** - Issue #10
9. **docs/arguments/format.md** - Issue #11
10. **docs/arguments/timeout.md** - Issue #12
11. **docs/arguments/output-dir.md** - Issue #13

---

## Recommended Resolution Approach

### Phase 1: Resolve Critical Contradictions (Issues #1-7)

For each issue, make a decision on the correct behavior:

1. **Issue #1** (`--list-tabs` + `--wait-for`): Change to **silently ignore** ✓
2. **Issue #2** (`--close-tab` + `--url-file`): Change to **works normally** ✓
3. **Issue #3** (`--tab` + `--open-browser`): Change to **error** ✓
4. **Issue #4** (`--all-tabs` + `--open-browser`): Change to **error** ✓
5. **Issue #5** (`--output` + `--open-browser`): Change to **warning** ✓
6. **Issue #6** (`--tab` + `--user-data-dir`): Change to **warning/ignore** ✓
7. **Issue #7** (`--all-tabs` + `--user-data-dir`): Change to **warning/ignore** ✓

### Phase 2: Standardize Warning Messages (Issues #8-14)

Standardize all warning message wording across documents.

### Phase 3: Verification

After updates:
1. Re-run cross-review to verify all contradictions resolved
2. Verify examples sections match the documented behavior
3. Check that validation.md is updated if any validation rules changed
4. Check that README.md quick reference matrices are updated

---

## Additional Notes

### Related TODO Items from PROJECT.md

- Line 900: "Argument trimming" - Apply `strings.TrimSpace()` to all string arguments
- Lines 902-907: "Parallel processing strategy" - Affects `--close-tab` behavior with multiple URLs
- Lines 909-915: "Review `--user-data-dir` interactions" - This review partially addresses this TODO
- Lines 917-923: "Remove `--force-visible` from code" - Not related to documentation issues

### Pattern Observations

1. **Standalone flags** (`--help`, `--version`, `--list-tabs`) should silently ignore all other flags except those needed for operation
2. **Conflicting content sources** (URL vs `--tab` vs `--all-tabs` vs `--url-file`) should **error**, not warn
3. **Browser mode conflicts** (`--force-headless` vs `--open-browser`) should **error**
4. **Ignored flags with no effect** (like `--user-agent` with existing tabs) should **warn**, not error
5. **All `--open-browser` ignored flags** should use consistent wording: "Warning: {flag} ignored with --open-browser (no content fetching)"

### Cross-Reference Check Methodology

For future reviews, the systematic approach:
1. For each flag in `docs/arguments/{flag}.md`
2. Read the "Interaction Matrix" section
3. For each interaction listed (e.g., "flag X + flag Y")
4. Open `docs/arguments/{flag-y}.md`
5. Find the reverse interaction ("flag Y + flag X")
6. Verify both documents say the same thing (behavior AND error/warning messages)
7. If mismatch found, document as inconsistency

---

## Next Steps

1. Review this document and approve recommendations for critical contradictions (#1-7)
2. Implement updates to all 11 files listed above
3. Run verification cross-review
4. Update validation.md if any validation rules changed
5. Update README.md compatibility matrices if needed
6. Consider updating PROJECT.md to mark the user-data-dir review TODO as complete

---

## Questions for Grant

Before implementing changes, confirm:

1. **Issue #2** (`--close-tab` + `--url-file`): Confirm it should work normally (close each tab after fetch). Note that the parallel processing TODO may affect implementation.

2. **Issues #3 & #4** (`--tab`/`--all-tabs` + `--open-browser`): Confirm these should error rather than warn. This is a stricter approach but clearer UX.

3. **Issue #5** (`--output` + `--open-browser` no URL): Confirm warning is better than error. This treats `--open-browser` as a valid mode that just doesn't fetch content.

4. **Issues #6 & #7** (`--tab`/`--all-tabs` + `--user-data-dir`): Confirm the flag should be ignored with warning. User might specify it for documentation purposes (knowing which profile the running browser uses), but it has no effect.

5. Are there any other flag combinations I should specifically verify in a second pass?

---

**Review Completed:** 2025-10-23 23:00 AEST
**Status:** Ready for Grant's review and approval before implementation
