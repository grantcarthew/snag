# Snag Phase 3: Output Management & Batch Operations

**Status**: Implementation In Progress (Step 2 of 15 Complete)

This document tracks Phase 3 implementation for Snag: enhanced file output options, additional format support (text/PDF), screenshot capture, and batch tab operations.

---

## Implementation Progress

### ✅ Completed Tasks

**Step 1: Output Management Module** ✅
- Created `output.go` with file naming functions
  - `SlugifyTitle()` - URL-safe slug generation (max 80 chars)
  - `GenerateURLSlug()` - Fallback slug from URL hostname
  - `GetFileExtension()` - Format to extension mapping
  - `GenerateFilename()` - Timestamp + slug + extension
  - `ResolveConflict()` - Append counter for file conflicts (with error handling)
- Updated `validate.go` with:
  - `validateDirectory()` - Check directory exists and is writable
  - `validateOutputPathEscape()` - Prevent path escape attacks
  - `validateFormat()` - Support for text and pdf formats
- Added format constants to `main.go`:
  - `FormatMarkdown`, `FormatHTML`, `FormatText`, `FormatPDF`
- **Code review fixes applied**:
  - Regex patterns compiled at package level (performance)
  - All format constants used consistently (no string literals)
  - Fixed infinite loop bug in `ResolveConflict()` (error handling)
  - Safety limit added (10,000 iterations max)

**Step 2: Text Format Support** ✅
- Renamed `convert.go` → `formats.go` (git history preserved)
- Renamed `convert_test.go` → `formats_test.go`
- Added dependency: `github.com/k3a/html2text` v1.2.1
  - **Zero non-standard dependencies** (pure stdlib)
  - Lightweight, feature-complete, stable API
- Implemented `extractPlainText()` method
  - Uses k3a/html2text with Unix line breaks
  - Strips all HTML tags and scripts
  - Preserves text structure
  - Handles HTML entities
- Added `FormatText` case to `Process()` method
- **Code review fix**: Removed logging side effects from utility functions
  - Removed `logger.Success()` from `convertToMarkdown()`
  - Removed `logger.Success()` from `extractPlainText()`
  - Clean separation: utilities transform data, `Process()` handles logging

**Files Created/Modified**:
- `output.go` (new, 165 lines)
- `formats.go` (renamed from convert.go, +11 lines)
- `formats_test.go` (renamed from convert_test.go)
- `validate.go` (+99 lines)
- `main.go` (+4 format constants)
- `go.mod` (+1 dependency)

**Testing**:
- All 30+ existing tests pass
- Build successful (20 MB binary)
- Manual testing verified for text extraction

---

### 🚧 Pending Implementation

**Step 3-15: Remaining Features**
1. Add PDF generation support (`--format pdf`)
2. Implement screenshot capture (`--screenshot` / `-s`)
3. Implement batch tab operations (`--all-tabs` / `-a`)
4. Add CLI flags and handlers for all new features
5. Add `--output-dir` / `-d` flag implementation
6. Integrate all features into main CLI flow
7. Write comprehensive tests for all functionality

---

## Key Design Decisions & Learnings

### Module Organization (Updated from Original Plan)

**Actual implementation**:
- `output.go` - File naming, slugification, conflict resolution
- `formats.go` - Format conversion (markdown, html, text, pdf)
- `validate.go` - Input/directory/path validation
- Screenshot & batch modules TBD

**Rationale**: Grouped by functionality rather than narrow files. Keeps related operations together.

### Text Extraction Library Choice

**Selected**: `github.com/k3a/html2text` (154 ⭐)

**Why**:
- Zero non-standard dependencies (aligns with single binary philosophy)
- Outputs actual plain text (not markdown-flavored)
- Lightweight (334 lines vs 549+ lines)
- Feature-complete and API stable
- We already have markdown conversion (no duplication)

**Rejected**: `jaytaylor/html2text` - Too heavyweight (3 dependencies), outputs markdown-flavored text

### Code Quality Improvements

**External code reviews caught**:
1. Regex compilation inefficiency - Fixed with package-level variables
2. Constant consistency issues - All formats now use constants
3. Infinite loop risk in `ResolveConflict()` - Added proper error handling
4. DRY violation with `validFormats` map - Removed from main.go
5. Logging side effects in utilities - Removed for clean separation

**All fixes applied before continuing implementation**.

### Format Constants Design

**Constants defined**:
```go
FormatMarkdown = "markdown"
FormatHTML     = "html"
FormatText     = "text"
FormatPDF      = "pdf"
```

**Not constant**: `"png"` for screenshots (internal use only, not a `--format` option)

**Rationale**: Screenshot is a separate flag (`--screenshot`), not part of format selection.

---

## Phase 3 Feature Specifications

### Feature 1: Output Directory (`--output-dir` / `-d`)

**Status**: Foundation complete (validation functions), CLI integration pending

**Implementation**:
- Directory validation: ✅ `validateDirectory()` in validate.go
- Path escape prevention: ✅ `validateOutputPathEscape()` in validate.go
- File naming: ✅ Functions in output.go
- CLI flag: ⏳ Pending

**Security**: Path escape validation prevents `../../etc/passwd` attacks

### Feature 2: Text Format Support (`--format text`)

**Status**: ✅ Complete

**Implementation**:
- Format constant: ✅ `FormatText` in main.go
- Extraction function: ✅ `extractPlainText()` in formats.go
- Validation: ✅ `validateFormat()` updated
- File extension: ✅ `.txt` via `GetFileExtension()`

**Testing**: Manual verification successful

### Feature 3: PDF Export (`--format pdf`)

**Status**: ⏳ Pending (Step 3)

**Planned approach**:
- Use Rod's `page.PDF()` method
- Chrome DevTools Protocol `Page.printToPDF`
- Binary output support
- Default Chrome PDF options

### Feature 4: Screenshot Capture (`--screenshot` / `-s`)

**Status**: ⏳ Pending (Step 4)

**Planned approach**:
- Use Rod's `page.Screenshot()` method
- Full page PNG capture
- Auto-generated filenames
- Save to CWD by default

### Feature 5: Batch Tab Operations (`--all-tabs` / `-a`)

**Status**: ⏳ Pending (Step 5)

**Planned approach**:
- Iterate all browser tabs
- Single timestamp for all files
- Continue-on-error strategy
- Progress output to stderr

---

## File Naming System (Implemented)

### Auto-Generated Filename Format

**Pattern**: `yyyy-mm-dd-hhmmss-{title-slug}.{ext}`

**Example**: `2025-10-21-104430-example-domain-website.md`

### Slugification Rules (Implemented)

1. Convert to lowercase
2. Replace non-alphanumeric with hyphen
3. Collapse multiple hyphens
4. Trim leading/trailing hyphens
5. Truncate to 80 characters

**Performance**: Regex patterns compiled once at package level

### Conflict Resolution (Implemented)

- Appends counter: `-1`, `-2`, `-3`, etc.
- Error handling for filesystem issues
- Safety limit: 10,000 iterations max
- Returns error signature: `(string, error)`

### URL Fallback (Implemented)

When page title is empty:
- Extract hostname from URL
- Apply same slugification rules
- Examples: `example.com` → `example-com`

---

## Format Support Summary

| Format | Flag | Extension | Status |
|--------|------|-----------|--------|
| Markdown | `--format markdown` | `.md` | ✅ Existing |
| HTML | `--format html` | `.html` | ✅ Existing |
| Text | `--format text` | `.txt` | ✅ Complete |
| PDF | `--format pdf` | `.pdf` | ⏳ Pending |
| Screenshot | `--screenshot` | `.png` | ⏳ Pending |

**Format aliases**: `txt` accepted as alias for `text` (via validation)

---

## Validation Rules (Implemented)

### Directory Validation
- ✅ Check directory exists
- ✅ Check directory is writable
- ✅ Do NOT auto-create directories
- ✅ Support relative and absolute paths

### Path Security
- ✅ Prevent `../` escape attacks
- ✅ Validate combined `-o` + `-d` paths
- ✅ Use `filepath.Clean()` and absolute path checks

### Format Validation
- ✅ Support: `markdown`, `html`, `text`, `pdf`
- ✅ Validate against constants
- ✅ Clear error messages

---

## Testing Strategy

**Current approach**: Implement all features first, comprehensive testing at end

**Existing tests**: All 30+ Phase 1/2 tests passing

**Planned tests** (Step 8):
1. Naming system tests (slugification, conflicts, fallbacks)
2. Format conversion tests (text, pdf)
3. Screenshot capture tests
4. Batch operation tests
5. Validation tests (directory, path escape, formats)
6. Integration tests (flag combinations)

---

## Dependencies

**Added in Phase 3**:
- `github.com/k3a/html2text` v1.2.1 - Plain text extraction

**Existing**:
- `github.com/urfave/cli/v2` - CLI framework
- `github.com/go-rod/rod` - Chrome DevTools Protocol
- `github.com/JohannesKaufmann/html-to-markdown/v2` - Markdown conversion

---

## Next Steps

**Immediate**: Step 3 - Add PDF generation support
**Following**: Steps 4-8 per original plan

**Ready to continue with step-by-step implementation.**

---

## Complete Feature Specifications (Reference)

<details>
<summary>Click to expand full Phase 3 specifications</summary>

[Original specifications retained for reference - available in git history or can be expanded as needed]

### Output Directory (`--output-dir` / `-d`)
- Save files with auto-generated names
- Validate directory exists and is writable
- Combine with `-o` flag for subdirectories
- Security: Path escape validation

### Text Format (`--format text` / `--format txt`)
- Extract plain text from HTML
- Strip all tags, scripts, styles
- Preserve basic text structure
- Unix line breaks for consistency

### PDF Export (`--format pdf`)
- Chrome print-to-PDF rendering
- Preserves styles, fonts, images
- Binary output support
- Default Chrome options

### Screenshot Capture (`--screenshot` / `-s`)
- Full-page PNG screenshots
- Auto-generated filenames
- Save to CWD or specified directory
- Rod's Page.Screenshot() method

### Batch Tab Operations (`--all-tabs` / `-a`)
- Process all browser tabs
- Single timestamp for batch
- Continue-on-error handling
- Progress output to stderr
- All formats supported

### File Naming Rules
- Pattern: `yyyy-mm-dd-hhmmss-{title-slug}.{ext}`
- 80 character slug limit
- URL hostname fallback
- Conflict resolution with counters

### Validation Rules
- Directory existence and writability
- Path escape prevention
- Format validation
- Flag conflict detection

</details>

---

**Document Version**: Updated 2025-10-21 during Phase 3 implementation
