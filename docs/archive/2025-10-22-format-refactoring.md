# Snag: Format Refactoring Project (Phase 3)

**Project**: Format system refactoring for consistency and UX improvements
**Status**: ✅ Implementation Complete - Testing & Documentation in Progress
**Started**: 2025-10-22
**Completed**: 2025-10-22
**Goal**: Unify format handling, eliminate code smells, and improve CLI consistency

---

## Project Summary

Successfully completed Phase 3 format refactoring, building on Phase 1 (main.go refactoring) and Phase 2 (tab management).

**Key Achievements:**

- Removed `--screenshot` flag, unified as `--format png`
- Normalized format names: `markdown` → `md` for consistency
- Added format alias support with case-insensitive matching
- Eliminated parameter interdependency code smell
- Updated comprehensive documentation (README, design-record, CLI help)

**Code Impact:**

- 5 files modified (main.go, validate.go, handlers.go, formats.go, output.go)
- 19 lines eliminated through parameter removal
- 127 lines added to README.md for format documentation
- 201 lines added to design-record.md documenting 4 new decisions

---

## Development Status

**Snag is under active development (pre-v1.0)**

- Current version: v0.0.4
- Status: Format refactoring complete, tests need updates
- **Backward compatibility**: NOT guaranteed until v1.0.0
- Breaking changes acceptable for better UX (ideal time for improvements)
- Next milestone: v0.1.0 with format refactoring complete

---

## Phase 3 Completed Work

### ✅ Step 1: Format Constants & Screenshot Flag Removal

**Completed**: 2025-10-22

**Changes Made:**

1. Updated format constants in main.go:
   - Changed `FormatMarkdown = "markdown"` → `FormatMarkdown = "md"`
   - Added `FormatPNG = "png"`
2. Removed `--screenshot` CLI flag entirely (main.go:93-96)
3. Removed `Screenshot bool` from Config struct (handlers.go:28)
4. Updated CLI description to mention all 5 formats
5. Updated `--format` flag help text: `md | html | text | pdf | png`

**Results:**

- Build: ✅ Successful
- Cleaner CLI interface (no special-case flags)
- All formats treated consistently

**Key Files Modified:**

- main.go:21-28 (format constants)
- main.go:87-92 (removed --screenshot flag)
- main.go:69-73 (updated CLI description)
- handlers.go:22-36 (removed Screenshot from Config)

---

### ✅ Step 2: Format Normalization & Alias Support

**Completed**: 2025-10-22

**Solution:** Created `normalizeFormat()` function for smart format handling

```go
func normalizeFormat(format string) string {
    format = strings.ToLower(format)  // Case-insensitive

    aliases := map[string]string{
        "markdown": FormatMarkdown,  // "markdown" → "md"
        "txt":      FormatText,      // "txt" → "text"
    }

    if normalized, ok := aliases[format]; ok {
        return normalized
    }
    return format
}
```

**Implementation:**

- Called in 3 locations: main.go:252, handlers.go:231, handlers.go:401
- Placed before format validation in all handlers
- Enables backward compatibility while using canonical names internally

**User Benefits:**

- Case-insensitive: `MD`, `MARKDOWN`, `Png` all work
- Backward compatible: `--format markdown` still works
- Intuitive aliases: `--format txt` works for text format

**Results:**

- Zero breaking change pain (aliases provide smooth migration)
- Consistent internal format representation
- Better UX (users don't worry about capitalization)

---

### ✅ Step 3: Parameter Interdependency Elimination

**Completed**: 2025-10-22

**Problem Identified:**
Both helper functions had interdependent parameters:

```go
// Before - TWO parameters for format:
func processPageContent(page *rod.Page, format string, screenshot bool, outputFile string)
func generateOutputFilename(title, url, format string, screenshot bool, timestamp, outputDir)
```

When `screenshot=true`, the `format` parameter was ignored. This is a code smell.

**Solution:**

```go
// After - ONE parameter for format:
func processPageContent(page *rod.Page, format string, outputFile string)
func generateOutputFilename(title, url, format string, timestamp, outputDir)
```

**Updated Call Sites:**

1. handlers.go:136 - `snag()` function
2. handlers.go:107-110 - `snag()` with --output-dir
3. handlers.go:125-128 - `snag()` auto-generate binary
4. handlers.go:305-308 - `handleAllTabs()` filename generation
5. handlers.go:316 - `handleAllTabs()` content processing
6. handlers.go:442-445 - `handleTabFetch()` auto-generate binary
7. handlers.go:453 - `handleTabFetch()` content processing

**Results:**

- handlers.go: 19 lines eliminated
- Cleaner function signatures (fewer parameters)
- No special-case logic needed
- Single source of truth for format

---

### ✅ Step 4: Binary Format Handling Updates

**Completed**: 2025-10-22

**Changes:**

1. Updated `processPageContent()` to check `format == FormatPNG`:

   ```go
   // formats.go:146
   if format == FormatPDF || format == FormatPNG {
       return converter.ProcessPage(page, outputFile)
   }
   ```

2. Updated binary format auto-detection:

   ```go
   // handlers.go:118, 441
   if config.OutputFile == "" && (config.Format == FormatPDF || config.Format == FormatPNG)
   ```

3. Updated formats.go PNG handling:

   - Case statement uses `FormatPNG` constant
   - Updated comments: "PNG screenshot" instead of generic "screenshot"

4. Updated output.go extension mapping:
   - `FormatPNG` constant instead of hardcoded "png"

**Results:**

- Consistent constant usage throughout codebase
- No hardcoded "png" strings in logic
- Clear semantic meaning (PNG is a format, not special case)

---

### ✅ Step 5: Format Validation Updates

**Completed**: 2025-10-22

**Updated Functions:**

1. `validateFormat()` - Now accepts PNG in valid formats map
2. Added `normalizeFormat()` for preprocessing
3. Updated error messages to show all 5 formats

**Validation Flow:**

```
User input → normalizeFormat() → validateFormat() → Handler
   "MARKDOWN" → "markdown" → "md" → ✅ Valid
   "txt" → "txt" → "text" → ✅ Valid
   "PNG" → "png" → "png" → ✅ Valid
```

**Results:**

- Case-insensitive validation
- Alias support working
- Clear error messages listing all valid formats

---

### ✅ Step 6: Documentation Updates

**Completed**: 2025-10-22

#### README.md Updates (127 new lines)

- Added comprehensive "Output Formats" section (lines 131-235)
- Documented all 5 formats with examples
- Explained binary format auto-filename behavior
- Added `--output-dir` and `--all-tabs` to CLI Reference
- Updated CLI Reference with format aliases and case-insensitivity
- Enhanced examples showing format flexibility

#### docs/design-record.md Updates (201 new lines)

- Added Phase 3 decision summary table (4 new decisions: 23-26)
- Decision 23: Format Name Normalization
- Decision 24: Format Alias Support
- Decision 25: Screenshot → PNG Format Refactor
- Decision 26: Binary Format Safety
- Updated feature set sections (moved text/pdf/png from planned to implemented)
- Added Phase 3 implementation notes with full details

#### main.go CLI Help Updates

- Updated app description to mention all formats
- Updated `--format` flag usage text

**Results:**

- Complete documentation coverage
- Clear migration guidance
- User-friendly examples
- Design decisions well-documented for future reference

---

## Breaking Changes (Pre-v1.0 Acceptable)

### 1. Screenshot Flag Removed

**Old**: `snag --screenshot https://example.com`
**New**: `snag --format png https://example.com`

**Migration**: No backward compatibility - clean break for better consistency

### 2. Format Name Changed

**Old**: `snag --format markdown https://example.com`
**New**: `snag --format md https://example.com`

**Migration**: Backward compatible via alias - old syntax still works!

---

## Code Quality Improvements

### Eliminated Code Smells

1. ✅ **Parameter interdependency** - No more `screenshot bool` + `format string` confusion
2. ✅ **Inconsistent naming** - All formats now match file extensions (md, html, text, pdf, png)
3. ✅ **Special-case logic** - PNG treated as regular format, not special case
4. ✅ **Hardcoded strings** - Using `FormatPNG` constant instead of "png" literals

### Design Patterns Applied

1. **Single Responsibility** - `normalizeFormat()` handles one thing (normalization)
2. **DRY Principle** - Format normalization in one place, called consistently
3. **Fail Fast** - Format validation happens early (after normalization)
4. **Least Surprise** - Case-insensitive matching matches user expectations
5. **Backward Compatibility** - Aliases provide smooth migration path

---

## File Structure (Updated)

```
main.go (318 lines) - CLI Framework
├── Format constants: FormatMarkdown="md", FormatHTML, FormatText, FormatPDF, FormatPNG
├── Removed --screenshot flag (was lines 93-96)
├── Updated --format help text
├── Calls normalizeFormat() before creating Config
└── Updated CLI description text

validate.go (222 lines) - Input Validation
├── normalizeFormat() - NEW function (lines 124-141)
│   ├── Lowercase conversion
│   ├── Alias mapping (markdown→md, txt→text)
│   └── Returns canonical format name
└── validateFormat() - Updated for PNG support

handlers.go (469 lines) - Business Logic
├── Config struct - Removed Screenshot bool field
├── processPageContent() - Removed screenshot parameter
│   └── Now checks: format == FormatPDF || format == FormatPNG
├── generateOutputFilename() - Removed screenshot parameter
│   └── Simpler signature, format is already canonical
├── 3 handlers call normalizeFormat() before validation
└── All binary format checks updated to use format constants

formats.go (306 lines) - Format Conversion
├── ProcessPage() - PNG case uses FormatPNG constant
├── Updated comments (PNG screenshot → PNG capture)
└── No special-case screenshot handling

output.go (160 lines) - File Naming
└── GetFileExtension() - Uses FormatPNG constant
```

---

## Test Status

### Build Status

- ✅ Compiles successfully: `go build -o snag`
- ✅ No syntax errors
- ✅ All imports resolved

### Test Updates Needed

1. ⏳ Update test files to use `--format md` instead of `--format markdown`
2. ⏳ Update tests to use `--format png` instead of `--screenshot`
3. ⏳ Fix `TestValidateFormat_Invalid` - pdf/text are now VALID, not invalid
4. ⏳ Fix `TestCLI_InvalidFormat` - expecting wrong invalid formats
5. ⏳ Add tests for case-insensitive format input
6. ⏳ Add tests for format aliases (markdown→md, txt→text)

### Test Plan

- Update cli_test.go references
- Update validate_test.go valid/invalid format lists
- Add new test cases for normalization
- Run full test suite: `go test -v ./...`
- Verify all 56+ tests pass

---

## Design Decisions (Phase 3)

### Decision 1: Format Name Normalization

**Choice**: Use `md` instead of `markdown`

**Rationale**:

- Consistency with other formats (all 2-4 chars, matching extensions)
- `markdown` was the only outlier (8 chars vs 2-4 for others)
- Matches file extension (.md files → md format)
- Less typing for most commonly used format
- Predictable (users can guess: .pdf → pdf, .png → png, .md → md)

### Decision 2: Alias Support + Case Insensitivity

**Choice**: Support backward-compatible aliases with case-insensitive input

**Rationale**:

- **Smooth migration**: Existing scripts using "markdown" continue working
- **Better UX**: Users don't worry about capitalization
- **Common expectations**: "txt" is intuitive for text files
- **No complexity penalty**: Simple map lookup, negligible performance
- **Future-proof**: Easy to add more aliases if needed

**Implementation**: Single `normalizeFormat()` function called before validation

### Decision 3: Screenshot → PNG Format

**Choice**: Remove `--screenshot` flag, use `--format png`

**Rationale**:

- **Eliminates code smell**: No more parameter interdependency
- **Consistency**: All visual outputs (PDF, PNG) are formats
- **Semantic clarity**: PNG and PDF both "visual captures" (not content extraction)
- **Simpler codebase**: No special-case handling needed
- **One way to do it**: Reduces confusion about screenshot vs format png

**Breaking Change**: Acceptable pre-v1.0 for long-term benefits

### Decision 4: Binary Format Auto-Filename

**Choice**: Auto-generate filenames for PDF/PNG when no output specified

**Rationale**:

- **Terminal safety**: Binary data corrupts terminal if output to stdout
- **Better UX**: Users don't need to remember `-o` for binary formats
- **Sensible default**: Users expect binary files to be saved
- **Consistent behavior**: All binary formats behave the same way

**Format**: `yyyy-mm-dd-hhmmss-{page-title-slug}.{ext}`

---

## Implementation Learnings

### What Went Well ✅

1. **Step-by-step approach** - Methodical implementation reduced errors
2. **Comprehensive code review** - Caught issues before testing
3. **User collaboration** - Discussed design decisions, improved solution (no screenshot alias, case-insensitive)
4. **Documentation-first** - Updated docs alongside code changes
5. **Build verification** - Checked compilation after each major change

### Key Insights 💡

1. **Alias support crucial** - Makes breaking changes painless for users
2. **Case-insensitivity expected** - Users don't think about format capitalization
3. **Binary format safety** - Auto-filename prevents terminal corruption issues
4. **Constant usage** - Using `FormatPNG` instead of "png" improves maintainability
5. **Single normalization point** - `normalizeFormat()` makes all handlers consistent

### Technical Patterns Used 🔧

1. **Normalization + Validation** - Two-stage approach (normalize → validate)
2. **Alias mapping** - Simple map lookup for backward compatibility
3. **Constant-based comparisons** - Avoid hardcoded strings in logic
4. **Early validation** - Fail fast with clear error messages
5. **Transparent aliases** - Internal code only sees canonical names

---

## Next Steps

### Immediate (This Session)

- [ ] Update test files (cli_test.go, validate_test.go, formats_test.go)
- [ ] Fix 2 outdated format validation tests
- [ ] Run full test suite and verify all pass
- [ ] Update AGENTS.md documentation
- [ ] Manual testing with common use cases
- [ ] Git commit with comprehensive message

### Short Term (Next Session)

- [ ] Consider additional format aliases if user feedback requests
- [ ] Monitor for any edge cases in format normalization
- [ ] Add integration tests for case-insensitive formats
- [ ] Add integration tests for alias support

### Long Term (Post-v1.0)

- [ ] Evaluate adding more binary formats (JPEG, WebP, etc.)
- [ ] Consider format-specific options (e.g., PDF page size)
- [ ] Explore format auto-detection from URLs
- [ ] Consider batch format conversion

---

## Related Documentation

### Design Documentation

- **docs/design-record.md** - Complete design decisions (26 total, 4 added in Phase 3)
- **README.md** - User-facing documentation (updated with all formats)
- **AGENTS.md** - AI agent reference (needs Phase 3 updates)

### Code Documentation

- **main.go** - Format constants and CLI setup
- **validate.go** - Format normalization and validation
- **handlers.go** - Format usage in business logic
- **formats.go** - Format conversion implementation
- **output.go** - Format-to-extension mapping

---

## Historical Context

### Phase 1: main.go Refactoring (2025-10-21)

- Split 868-line main.go into main.go (318 lines) + handlers.go (469 lines)
- Extracted 4 helper functions
- Fixed 5 code smells
- Reduced codebase by 399 lines

### Phase 2: Tab Management (2025-10-21)

- Added `--list-tabs`, `--tab`, `--all-tabs` features
- 1-based tab indexing
- Progressive pattern matching (exact → contains → regex)
- Performance optimization (3x improvement with caching)

### Phase 3: Format Refactoring (2025-10-22) ← Current

- Unified format handling (removed `--screenshot`)
- Normalized format names (markdown → md)
- Added alias support with case-insensitivity
- Comprehensive documentation updates
- 4 new design decisions documented

---

## Session Continuity Notes

### Current State (2025-10-22)

- ✅ Phase 3 implementation complete
- ✅ Code builds successfully
- ✅ Documentation updated (README.md, design-record.md)
- ⏳ Tests need updates for new format names
- ⏳ AGENTS.md needs Phase 3 updates
- ⏳ Manual testing pending

### To Resume Work

**Priority 1: Testing**

1. Update cli_test.go: Change "markdown" → "md", "--screenshot" → "--format png"
2. Update validate_test.go: Fix valid/invalid format lists
3. Add new tests for case-insensitivity
4. Add new tests for aliases (markdown→md, txt→text)
5. Run `go test -v ./...` and verify all pass

**Priority 2: Documentation**

1. Update AGENTS.md with Phase 3 changes
2. Verify all format examples use canonical names
3. Check for any remaining "screenshot" references

**Priority 3: Finalization**

1. Manual testing: Try all 5 formats with various options
2. Test case-insensitive input: `--format MD`, `--format MARKDOWN`
3. Test aliases: `--format markdown`, `--format txt`
4. Create comprehensive git commit

### Key Files to Review

- cli_test.go - Most test updates needed here
- validate_test.go - Format validation test fixes
- AGENTS.md - Needs Phase 3 format examples

---

**Document Version**: 2025-10-22 - Phase 3 format refactoring complete
