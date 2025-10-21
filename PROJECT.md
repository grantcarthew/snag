# Snag Phase 3: Output Management & Batch Operations

**Status**: Implementation In Progress (Step 3 of 15 Complete)

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
  - Removed duplicate validFormats map from main.go

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

**Step 3: PDF Generation Support** ✅
- Implemented PDF generation using Chrome's print-to-PDF
- Added `ProcessPage()` method for binary formats needing Rod page object
- Implemented `generatePDF()` using `page.PDF()` and CDP `Page.printToPDF`
- Added binary output methods:
  - `writeBinaryToStdout()` - Binary data to stdout
  - `writeBinaryToFile()` - Binary data to file
- Updated main.go handlers:
  - `run()` function detects PDF format and calls `ProcessPage()`
  - `handleTabFetch()` similarly updated for tab-based PDF
- **PDF settings**:
  - Uses Chrome's default (locale-aware paper size: A4 in AU, Letter in US)
  - Print background graphics enabled
  - Default margins and scaling
- **Code review fix**: Removed duplicate `fetcher.Fetch()` call
  - Single fetch before format branching
  - Better performance and cleaner code
- **Testing**: Generated valid PDF (29 KB, version 1.4)
  - Works with file output (`-o test.pdf`)
  - Works with stdout redirect (`> page.pdf`)

**Files Created/Modified**:
- `output.go` (new, 165 lines)
- `formats.go` (renamed from convert.go, +95 lines for PDF + text)
- `formats_test.go` (renamed from convert_test.go)
- `validate.go` (+99 lines)
- `main.go` (+4 format constants, +18 lines for PDF handling)
- `go.mod` (+1 dependency: k3a/html2text)

**Testing**:
- All 30+ existing tests pass
- Build successful (20 MB binary)
- Manual testing verified:
  - Text extraction works
  - PDF generation produces valid PDFs
  - Binary stdout works correctly

---

### 🚧 Pending Implementation

**Step 4-15: Remaining Features**
1. ⏳ **Next**: Implement screenshot capture (`--screenshot` / `-s`)
2. Implement batch tab operations (`--all-tabs` / `-a`)
3. Add CLI flags and handlers for all new features
4. Add `--output-dir` / `-d` flag implementation
5. Integrate all features into main CLI flow
6. Write comprehensive tests for all functionality

---

## Key Design Decisions & Learnings

### Module Organization (Finalized)

**Actual implementation**:
- `output.go` - File naming, slugification, conflict resolution
- `formats.go` - Format conversion (markdown, html, text, pdf, screenshot)
- `validate.go` - Input/directory/path validation
- Screenshot & batch modules TBD (may integrate into formats.go)

**Rationale**: Grouped by functionality rather than narrow files. Keeps related operations together.

### Binary vs Text Format Architecture

**Design pattern established**:
- `Process(html, outputFile)` - For text formats (markdown, html, text)
- `ProcessPage(page, outputFile)` - For binary formats (pdf, screenshot)

**Why separate methods**:
- Text formats only need HTML string
- Binary formats need live Rod page object
- Cleaner separation of concerns
- Avoids unnecessary HTML extraction for binary formats

**Implementation**:
```go
// Main flow pattern
html, err := fetcher.Fetch(...)  // Navigate and load page once

if format == FormatPDF {
    converter.ProcessPage(page, outputFile)  // Use page object
} else {
    converter.Process(html, outputFile)       // Use HTML string
}
```

### Text Extraction Library Choice

**Selected**: `github.com/k3a/html2text` (154 ⭐)

**Why**:
- Zero non-standard dependencies (aligns with single binary philosophy)
- Outputs actual plain text (not markdown-flavored)
- Lightweight (334 lines vs 549+ lines)
- Feature-complete and API stable
- We already have markdown conversion (no duplication)

**Rejected**: `jaytaylor/html2text` - Too heavyweight (3 dependencies), outputs markdown-flavored text

### PDF Paper Size Decision

**Question**: Letter vs A4 default?

**Solution**: Use Chrome's default (locale-aware)
- Chrome respects system locale automatically
- A4 in Australia, Europe, Asia, most of world
- Letter in US, Canada, Mexico
- No hardcoding needed
- Future: Add `--pdf-size` flag if customization needed

**Implementation**: Call `page.PDF()` without PaperWidth/PaperHeight parameters

### Code Quality Improvements

**External code reviews caught**:
1. ✅ Regex compilation inefficiency - Fixed with package-level variables
2. ✅ Constant consistency issues - All formats now use constants
3. ✅ Infinite loop risk in `ResolveConflict()` - Added proper error handling
4. ✅ DRY violation with `validFormats` map - Removed from main.go
5. ✅ Logging side effects in utilities - Removed for clean separation
6. ✅ Duplicate fetch calls - Refactored to single fetch before branching
7. ❌ Duplicate format branching logic - Acknowledged but kept (simple, clear)

**Principle established**: Accept minor duplication when abstraction adds more complexity than value

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
- CLI flag: ⏳ Pending (Step 7)

**Security**: Path escape validation prevents `../../etc/passwd` attacks

### Feature 2: Text Format Support (`--format text`)

**Status**: ✅ Complete

**Implementation**:
- Format constant: ✅ `FormatText` in main.go
- Extraction function: ✅ `extractPlainText()` in formats.go
- Validation: ✅ `validateFormat()` updated
- File extension: ✅ `.txt` via `GetFileExtension()`
- Integration: ✅ Works with `Process()` method

**Testing**: Manual verification successful

**Example**:
```bash
$ snag --format text https://example.com
Test Title

This is bold text.
```

### Feature 3: PDF Export (`--format pdf`)

**Status**: ✅ Complete

**Implementation**:
- Format constant: ✅ `FormatPDF` in main.go
- PDF generation: ✅ `generatePDF()` in formats.go
- Binary output: ✅ `ProcessPage()` method with binary I/O
- Validation: ✅ `validateFormat()` updated
- File extension: ✅ `.pdf` via `GetFileExtension()`
- Integration: ✅ Works in both `run()` and `handleTabFetch()`

**Technical details**:
- Uses Rod's `page.PDF()` method
- Chrome DevTools Protocol `Page.printToPDF`
- Locale-aware paper size (A4/Letter)
- Print background graphics enabled
- Returns StreamReader, read with `io.ReadAll()`

**Testing**:
- ✅ Generates valid PDF (version 1.4)
- ✅ File output works (`-o test.pdf`)
- ✅ Stdout redirect works (`> page.pdf`)
- ✅ Binary magic bytes correct (`%PDF-1.4`)

**Example**:
```bash
$ snag --format pdf https://example.com > page.pdf
$ snag -f pdf -o report.pdf https://example.com
```

### Feature 4: Screenshot Capture (`--screenshot` / `-s`)

**Status**: ⏳ Next step (Step 4)

**Planned approach**:
- Add screenshot case to `ProcessPage()` method
- Use Rod's `page.Screenshot()` method
- Full page PNG capture
- Reuse binary output methods
- CLI flag integration in Step 7

### Feature 5: Batch Tab Operations (`--all-tabs` / `-a`)

**Status**: ⏳ Pending (Step 5+)

**Planned approach**:
- Iterate all browser tabs
- Single timestamp for all files
- Continue-on-error strategy
- Progress output to stderr
- Support all formats (markdown, html, text, pdf, screenshot)

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

**Examples**:
```
"Example Domain"              → "example-domain"
"GitHub - Project Page"       → "github-project-page"
"Docs   -   The Go Language"  → "docs-the-go-language"
"!!!Test???"                  → "test"
```

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

| Format | Flag | Extension | Status | Output Type |
|--------|------|-----------|--------|-------------|
| Markdown | `--format markdown` | `.md` | ✅ Existing | Text |
| HTML | `--format html` | `.html` | ✅ Existing | Text |
| Text | `--format text` | `.txt` | ✅ Complete | Text |
| PDF | `--format pdf` | `.pdf` | ✅ Complete | Binary |
| Screenshot | `--screenshot` | `.png` | ⏳ Step 4 | Binary |

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
- ⏳ Add `png` when screenshot implemented

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

**Immediate**: Step 4 - Implement screenshot capture
**Following**: Steps 5-8 per original plan

**Ready to continue with step-by-step implementation.**

---

## Architecture Patterns Established

### Format Processing Pattern

**Text formats** (markdown, html, text):
```go
html, err := fetcher.Fetch(opts)
converter.Process(html, outputFile)
```

**Binary formats** (pdf, screenshot):
```go
html, err := fetcher.Fetch(opts)  // Still need to load page
converter.ProcessPage(page, outputFile)
```

### Code Organization

**formats.go structure**:
- `Process()` - Text format conversion (string → string)
- `ProcessPage()` - Binary format generation (page → []byte)
- Individual converters: `convertToMarkdown()`, `extractPlainText()`, `generatePDF()`
- Output methods: text I/O + binary I/O

**Benefit**: Clear separation between format types, reusable output methods

---

## Complete Feature Specifications (Reference)

<details>
<summary>Click to expand original Phase 3 specifications</summary>

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
- Locale-aware default paper size

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

**Document Version**: Updated 2025-10-21 during Phase 3 implementation (Step 3 complete)
