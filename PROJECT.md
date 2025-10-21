# Snag Phase 3: Output Management & Batch Operations

**Status**: Design Complete - Ready for Implementation

This document specifies Phase 3 features for Snag: enhanced file output options, additional format support (text/PDF), screenshot capture, and batch tab operations.

## Overview

Phase 3 adds five major capabilities:

1. **Output Directory Management** (`--output-dir` / `-d`): Save files with auto-generated names
2. **Text Format Support** (`--format text` or `--format txt`): Extract plain text from pages
3. **PDF Export** (`--format pdf`): Generate PDF using Chrome's print-to-PDF
4. **Screenshot Capture** (`--screenshot` / `-s`): Capture full-page PNG screenshots
5. **Batch Tab Operations** (`--all-tabs` / `-a`): Process all open browser tabs at once

These features maintain Snag's core philosophy: simple, focused, pipe-friendly content fetching, while adding practical file management and format options for AI agents and automation workflows.

---

## Feature 1: Output Directory (`--output-dir` / `-d`)

### Specification

**Flag**: `--output-dir <directory>` or `-d <directory>`

**Purpose**: Save fetched content to specified directory with auto-generated filename.

**Behavior**:
- Validates directory exists and is writable (errors if not)
- Does NOT create directory if missing (user must `mkdir` first)
- Suppresses stdout (like `--output` flag)
- Auto-generates filename using timestamp + page title slug

**Examples**:
```bash
snag -d ./docs https://example.com
# → ./docs/2025-10-21-104430-example-domain.md

snag -d ./output --format html https://go.dev
# → ./output/2025-10-21-104532-documentation-the-go-programming-language.html

snag -t 1 -d ./archive
# → ./archive/2025-10-21-104615-github-project-page.md
```

### Interaction with `--output` Flag

The `-d` and `-o` flags can be **combined**:

- `-o` provides the **filename** (or relative subpath)
- `-d` provides the **directory**
- Combined path: `{directory}/{filename}`

**Examples**:
```bash
snag -o custom.md -d ./docs https://example.com
# → ./docs/custom.md

snag -o project/notes.md -d ./archive https://github.com
# → ./archive/project/notes.md (creates ./archive/project/ with mkdir -p)

snag -o /absolute/path.md https://example.com
# → /absolute/path.md (ignores -d if -o is absolute path)
```

### Security: Path Escape Validation

When combining `-o` and `-d`, validate the resulting path doesn't escape the output directory:

```bash
snag -o ../../etc/passwd -d ./docs https://example.com
# → ERROR: output path escapes directory
```

**Implementation**:
- Use `filepath.Clean()` and `filepath.Join()`
- Verify cleaned path has output directory as prefix
- Reject paths containing `..` that escape bounds

---

## Feature 2: Text Format Support (`--format text` / `--format txt`)

### Specification

**Flag**: `--format text` or `--format txt` (both accepted as aliases)

**Purpose**: Extract plain text content from web pages, stripping all HTML markup.

**Behavior**:
- Fetches page via browser (JavaScript executed, dynamic content loaded)
- Strips all HTML tags, scripts, styles
- Preserves basic text structure (paragraphs, spacing)
- Outputs clean, readable plain text
- Works with all output options (`-o`, `-d`, stdout)

**Examples**:
```bash
snag --format text https://example.com
# Plain text to stdout

snag -f txt -o content.txt https://example.com
# → ./content.txt

snag -f text -d ./docs https://example.com
# → ./docs/2025-10-21-104430-example-domain.txt

snag -a -f txt -d ./archive
# → All tabs as .txt files
```

**Use Cases**:
- Extract article content for text analysis
- Feed content to AI/LLM systems (plain text preferred)
- Create searchable text archives
- Remove formatting for readability

**File Extension**: `.txt`

---

## Feature 3: PDF Export (`--format pdf`)

### Specification

**Flag**: `--format pdf`

**Purpose**: Generate PDF using Chrome's native print-to-PDF rendering engine.

**Behavior**:
- Uses Chrome DevTools Protocol `Page.printToPDF` command
- Renders page exactly as browser displays it
- Preserves styles, fonts, images, layout
- Full page capture (not paginated by default)
- Binary output (can be piped or saved to file)

**Examples**:
```bash
snag --format pdf https://example.com > page.pdf
# Binary PDF to stdout, redirect to file

snag -f pdf -o report.pdf https://example.com
# → ./report.pdf

snag -f pdf -d ./docs https://example.com
# → ./docs/2025-10-21-104430-example-domain.pdf

snag -a -f pdf -d ./archive
# → All tabs as PDF files
```

**Use Cases**:
- Archive web pages for offline viewing
- Print-ready document generation
- Visual documentation with preserved formatting
- Legal/compliance archiving

**File Extension**: `.pdf`

**PDF Options** (using Chrome defaults):
- Print background graphics: enabled
- Format: Letter (8.5" × 11")
- Margins: default browser margins
- Scale: 1.0 (100%)

**Future Enhancement**: Add flags for page size, margins, orientation if needed.

---

## Feature 4: Screenshot Capture (`--screenshot` / `-s`)

### Specification

**Flag**: `--screenshot` or `-s`

**Purpose**: Capture full-page screenshot instead of fetching content.

**Format**: PNG only (lossless, best for text/UI)
- Based on Rod's `Page.Screenshot()` method
- CDP calls it `Page.captureScreenshot`
- Full page capture (not viewport only)
- No format options (KISS principle)

**Behavior**:
- Captures entire page (full page scroll)
- Saves as PNG with auto-generated filename
- Without `-d` or `-o`: saves to current working directory
- With `-d`: saves to specified directory
- With `-o`: saves to specified filename

**Examples**:
```bash
snag -s https://example.com
# → ./2025-10-21-104430-example-domain.png (CWD)

snag -s -d ./screenshots https://github.com
# → ./screenshots/2025-10-21-104532-github-project.png

snag -s -o capture.png https://go.dev
# → ./capture.png

snag -t 1 -s -d ./images
# → ./images/2025-10-21-104615-current-tab-title.png
```

**User Feedback**:

When saving to CWD without explicit output path, log helpful message:
```bash
$ snag -s https://example.com
Saving screenshot to ./2025-10-21-104430-example-domain.png
```

### Screenshot vs Content

`-s` flag saves **screenshot only** (not content). To get both:

```bash
snag -d ./out https://example.com    # Content
snag -s -d ./out https://example.com # Screenshot
# Results in two files:
#   ./out/2025-10-21-104430-example-domain.md
#   ./out/2025-10-21-104532-example-domain.png
```

---

## Feature 5: Batch Tab Operations (`--all-tabs` / `-a`)

### Specification

**Flag**: `--all-tabs` or `-a`

**Purpose**: Fetch content or screenshots from all open browser tabs.

**Behavior**:
- Requires existing browser connection (errors if none found)
- Processes all tabs in browser sequentially
- Continues on error (best effort approach)
- All files share same timestamp (generated at start)
- Defaults to current working directory
- Can specify output directory with `-d`

**Examples**:
```bash
snag -a
# Saves all tabs as Markdown to CWD

snag -a -d ./archive
# Saves all tabs as Markdown to ./archive/

snag -a -s
# Screenshots all tabs to CWD as PNG

snag -a -s -d ./screenshots
# Screenshots all tabs to ./screenshots/

snag -a --format html -d ./backup
# Saves all tabs as HTML to ./backup/

snag -a -f pdf -d ./docs
# Saves all tabs as PDF to ./docs/

snag -a -f txt -d ./text
# Saves all tabs as plain text to ./text/
```

### Progress Output

Display progress to stderr (stdout reserved for content in normal mode):

```bash
$ snag -a -d ./docs
Fetching 5 tabs to ./docs/...
[1/5] example-domain.md
[2/5] github-project.md
[3/5] timeout-error (error: page load timeout)
[4/5] go-documentation.md
[5/5] stackoverflow-question.md
Completed: 4 saved, 1 failed
```

### Error Handling

**Continue on error** (best effort):
- If a tab fails to load/render, log error and continue
- Success count reported at end
- Exit code `1` if any failures occurred
- Exit code `0` if all succeeded

### Timestamp & Conflict Resolution

1. Generate **one timestamp** at start: `2025-10-21-104430`
2. Use for **all files** in batch
3. If duplicate titles exist, append counter: `-1`, `-2`, etc.

**Example with duplicate titles**:
```bash
# Browser has 3 tabs:
# - Tab 1: "Example Domain" at https://example.com
# - Tab 2: "Example Domain" at https://example.org (same title!)
# - Tab 3: "GitHub"

$ snag -a -d ./output
Fetching 3 tabs to ./output/...
[1/3] example-domain.md
[2/3] example-domain-1.md
[3/3] github.md
Completed: 3 saved, 0 failed

# Result:
#   2025-10-21-104430-example-domain.md    (from example.com)
#   2025-10-21-104430-example-domain-1.md  (from example.org)
#   2025-10-21-104430-github.md
```

### Validation

**Requires browser connection**:
```bash
$ snag -a
Error: no browser found. Open browser first with --open-browser or connect to existing browser.
```

**Conflicts with other flags**:
```bash
snag -a -t 1
# ERROR: --all-tabs conflicts with --tab

snag -a https://example.com
# ERROR: --all-tabs conflicts with URL argument
```

---

## Format Support Summary

**Phase 3 Format Options:**

| Format | Flag | Extension | Output Type | Use Case |
|--------|------|-----------|-------------|----------|
| Markdown | `--format markdown` (default) | `.md` | Text | Documentation, readable content |
| HTML | `--format html` | `.html` | Text | Raw page source, full markup |
| Text | `--format text` or `--format txt` | `.txt` | Text | Plain text extraction, AI input |
| PDF | `--format pdf` | `.pdf` | Binary | Archival, print-ready documents |
| Screenshot | `--screenshot` or `-s` | `.png` | Binary | Visual capture, full page image |

**Format Aliases**:
- `text` and `txt` are equivalent (both accepted)
- Screenshot is a separate flag (not part of `--format`)

**Examples**:
```bash
snag -f markdown https://example.com  # Default
snag -f html https://example.com      # Raw HTML
snag -f text https://example.com      # Plain text
snag -f txt https://example.com       # Same as text
snag -f pdf https://example.com       # PDF document
snag -s https://example.com           # PNG screenshot
```

---

## File Naming Rules

### Auto-Generated Filename Format

**Pattern**: `yyyy-mm-dd-hhmmss-{title-slug}.{ext}`

**Components**:
- **Timestamp**: `2025-10-21-104430` (local time, sortable)
- **Separator**: Single hyphen `-`
- **Title slug**: Slugified page title (max 80 chars)
- **Extension**: `.md`, `.html`, `.txt`, `.pdf`, or `.png`

**Example**: `2025-10-21-104430-example-domain-website.md`

### Title Slugification Rules

1. Get page title from `page.Info().Title`
2. Convert to lowercase
3. Replace all non-alphanumeric characters with hyphen
4. Collapse multiple consecutive hyphens to single hyphen
5. Trim leading and trailing hyphens
6. Truncate to **80 characters maximum**

**Examples**:
```
"Example Domain"              → "example-domain"
"GitHub - Project Page"       → "github-project-page"
"Docs   -   The Go Language"  → "docs-the-go-language"
"!!!Test???"                  → "test"
"Very Long Title That Exceeds The Maximum Character Limit And Needs To Be Truncated Here"
  → "very-long-title-that-exceeds-the-maximum-character-limit-and-needs-to-be-t"
```

### Fallback: No Title Available

If page has no title or title is empty after slugification:
- Use URL hostname as slug
- Apply same slugification rules
- Max 80 characters

**Examples**:
```
https://example.com         → "example-com"
https://subdomain.site.org  → "subdomain-site-org"
https://192.168.1.1         → "192-168-1-1"
```

### File Extension

Determined by operation:
- Content with `--format markdown`: `.md` (default)
- Content with `--format html`: `.html`
- Content with `--format text` or `--format txt`: `.txt`
- Content with `--format pdf`: `.pdf`
- Screenshot: `.png` (always)

### Conflict Resolution

If generated filename already exists in target directory:
1. Append counter before extension: `-1`, `-2`, `-3`, etc.
2. Check again until unique filename found

**Examples**:
```bash
# First save
snag -d ./out https://example.com
# → ./out/2025-10-21-104430-example-domain.md

# Second save (same second, same title)
snag -d ./out https://example.com
# → ./out/2025-10-21-104430-example-domain-1.md

# Third save
snag -d ./out https://example.com
# → ./out/2025-10-21-104430-example-domain-2.md
```

**Note**: Conflicts rare due to timestamp precision (1 second), but handled for safety.

---

## Complete Flag Combinations Matrix

**Syntax**: `snag [options] <url>`
**Note**: Options MUST come before URL argument.

### Basic Content Fetching

| Command | Output |
|---------|--------|
| `snag https://example.com` | Markdown to stdout |
| `snag --format html https://example.com` | HTML to stdout |
| `snag --format text https://example.com` | Plain text to stdout |
| `snag --format txt https://example.com` | Plain text to stdout (same as text) |
| `snag --format pdf https://example.com > page.pdf` | PDF to file (redirected) |
| `snag -o file.md https://example.com` | Markdown to `./file.md` |
| `snag -d ./docs https://example.com` | Markdown to `./docs/2025-10-21-104430-example-domain.md` |
| `snag -o custom.md -d ./docs https://example.com` | Markdown to `./docs/custom.md` |
| `snag -o sub/file.md -d ./docs https://example.com` | Markdown to `./docs/sub/file.md` (mkdir -p) |

### Format Variations with Output Directory

| Command | Output |
|---------|--------|
| `snag -f html -d ./docs https://example.com` | HTML to `./docs/2025-10-21-104430-example-domain.html` |
| `snag -f text -d ./docs https://example.com` | Text to `./docs/2025-10-21-104430-example-domain.txt` |
| `snag -f txt -d ./docs https://example.com` | Text to `./docs/2025-10-21-104430-example-domain.txt` |
| `snag -f pdf -d ./docs https://example.com` | PDF to `./docs/2025-10-21-104430-example-domain.pdf` |
| `snag -f pdf -o report.pdf https://example.com` | PDF to `./report.pdf` |

### Screenshot Capture

| Command | Output |
|---------|--------|
| `snag -s https://example.com` | PNG to `./2025-10-21-104430-example-domain.png` (CWD) |
| `snag -s -o shot.png https://example.com` | PNG to `./shot.png` |
| `snag -s -d ./images https://example.com` | PNG to `./images/2025-10-21-104430-example-domain.png` |
| `snag -s -o capture.png -d ./images https://example.com` | PNG to `./images/capture.png` |

### Tab Selection

| Command | Output |
|---------|--------|
| `snag -t 1` | Tab 1 content to stdout (markdown) |
| `snag -t "github"` | Matching tab content to stdout |
| `snag -t 1 -d ./docs` | Tab 1 to `./docs/2025-10-21-104430-title.md` |
| `snag -t 1 -f pdf -d ./docs` | Tab 1 to `./docs/2025-10-21-104430-title.pdf` |
| `snag -t 1 -s` | Tab 1 screenshot to `./2025-10-21-104430-title.png` |
| `snag -t 1 -s -d ./images` | Tab 1 screenshot to `./images/2025-10-21-104430-title.png` |

### Batch Tab Operations

| Command | Output |
|---------|--------|
| `snag -a` | All tabs to CWD as `./2025-10-21-104430-*.md` |
| `snag -a -d ./archive` | All tabs to `./archive/2025-10-21-104430-*.md` |
| `snag -a -f html -d ./backup` | All tabs to `./backup/2025-10-21-104430-*.html` |
| `snag -a -f text -d ./text` | All tabs to `./text/2025-10-21-104430-*.txt` |
| `snag -a -f txt -d ./text` | All tabs to `./text/2025-10-21-104430-*.txt` (same) |
| `snag -a -f pdf -d ./docs` | All tabs to `./docs/2025-10-21-104430-*.pdf` |
| `snag -a -s` | All screenshots to CWD as `./2025-10-21-104430-*.png` |
| `snag -a -s -d ./screenshots` | All screenshots to `./screenshots/2025-10-21-104430-*.png` |

### Invalid Combinations (Errors)

| Command | Error |
|---------|-------|
| `snag -a -t 1` | `--all-tabs conflicts with --tab` |
| `snag -a https://example.com` | `--all-tabs conflicts with URL argument` |
| `snag -a` (no browser) | `no browser connection found` |
| `snag -o ../../escape.md -d ./docs https://example.com` | `output path escapes directory` |
| `snag -d ./nonexistent https://example.com` | `directory does not exist: ./nonexistent` |
| `snag -d /read-only https://example.com` | `directory not writable: /read-only` |
| `snag -s -f pdf https://example.com` | `--screenshot conflicts with --format` |

---

## Validation Rules

### Directory Validation (`-d` flag)

When `--output-dir` is specified:

1. **Exists**: Directory must exist (error if not)
2. **Writable**: Directory must be writable (error if not)
3. **Not Created**: Do NOT auto-create missing directories
4. **Relative/Absolute**: Support both path types

**Error messages**:
```
Error: directory does not exist: ./missing
Error: directory not writable: /read-only
```

### Path Escape Validation

When combining `-o` and `-d`:

1. Join paths: `filepath.Join(outputDir, filename)`
2. Clean result: `filepath.Clean(joined)`
3. Clean directory: `filepath.Clean(outputDir)`
4. Validate: cleaned path must start with `cleanDir + separator`

**Example validation**:
```go
func validateOutputPath(outputDir, filename string) error {
    if !filepath.IsAbs(filename) {
        fullPath := filepath.Join(outputDir, filename)
        cleanPath := filepath.Clean(fullPath)
        cleanDir := filepath.Clean(outputDir)

        // Ensure path doesn't escape directory
        if !strings.HasPrefix(cleanPath, cleanDir+string(filepath.Separator)) {
            return fmt.Errorf("output path escapes directory: %s", filename)
        }
    }
    return nil
}
```

### Format Validation

**Valid format values**:
- `markdown` (default)
- `html`
- `text` (canonical)
- `txt` (alias for `text`)
- `pdf`

**Implementation**:
```go
// Normalize format aliases
func normalizeFormat(format string) (string, error) {
    switch strings.ToLower(format) {
    case "markdown", "md":
        return "markdown", nil
    case "html":
        return "html", nil
    case "text", "txt":
        return "text", nil
    case "pdf":
        return "pdf", nil
    default:
        return "", fmt.Errorf("invalid format: %s (valid: markdown, html, text, txt, pdf)", format)
    }
}
```

### Flag Conflict Validation

**`--all-tabs` conflicts**:
- Cannot use with `--tab` (select all vs select one)
- Cannot use with URL argument (process tabs vs fetch URL)

**`--screenshot` conflicts**:
- Cannot use with `--format` (screenshot is separate from format)

**Browser requirement** (validation at runtime):
- `--all-tabs` requires existing browser connection
- `--tab` requires existing browser connection
- `--list-tabs` requires existing browser connection

### Filename Character Validation

Auto-generated filenames must be filesystem-safe:
- Only lowercase `a-z`, digits `0-9`, hyphens `-`, and dots `.`
- No special characters: `/ \ : * ? " < > |`
- Max length: 255 bytes (filesystem limit)
- Our format: ~102 characters typical (safe)

---

## Implementation Notes

### Module Organization

Add or modify these files:

- **`naming.go`** (new): Auto-filename generation, slugification, conflict resolution
- **`screenshot.go`** (new): Screenshot capture logic using Rod
- **`pdf.go`** (new): PDF generation using Chrome print-to-PDF
- **`text.go`** (new): Plain text extraction from HTML
- **`batch.go`** (new): Batch tab processing for `--all-tabs`
- **`validate.go`**: Add directory validation, path escape checks, format normalization
- **`main.go`**: Add new flags, handlers, validation orchestration
- **`browser.go`**: May need helper for tab iteration

### Key Functions to Implement

**naming.go**:
```go
// GenerateFilename creates auto filename from page title and timestamp
func GenerateFilename(title string, format string, timestamp time.Time) string

// SlugifyTitle converts page title to URL-safe slug
func SlugifyTitle(title string, maxLen int) string

// ResolveConflict appends counter if file exists
func ResolveConflict(dir, filename string) string

// GenerateURLSlug creates slug from URL hostname as fallback
func GenerateURLSlug(url string) string
```

**screenshot.go**:
```go
// CaptureScreenshot takes full-page PNG screenshot
func CaptureScreenshot(page *rod.Page) ([]byte, error)

// SaveScreenshot captures and saves to file
func SaveScreenshot(page *rod.Page, filepath string) error
```

**pdf.go**:
```go
// GeneratePDF creates PDF from page using Chrome's print-to-PDF
func GeneratePDF(page *rod.Page) ([]byte, error)

// SavePDF generates and saves PDF to file
func SavePDF(page *rod.Page, filepath string) error
```

**text.go**:
```go
// ExtractPlainText strips HTML and returns clean text
func ExtractPlainText(html string) (string, error)

// Uses goquery or similar to:
// - Remove script, style, nav, footer tags
// - Strip all HTML tags
// - Preserve paragraph breaks
// - Trim excessive whitespace
```

**batch.go**:
```go
// ProcessAllTabs fetches or screenshots all browser tabs
func ProcessAllTabs(browser *rod.Browser, opts BatchOptions) error

// BatchOptions configures batch operation
type BatchOptions struct {
    OutputDir   string
    Format      string // "markdown", "html", "text", "pdf"
    Screenshot  bool
    Timestamp   time.Time
}
```

**validate.go** (additions):
```go
// ValidateDirectory checks directory exists and is writable
func ValidateDirectory(dir string) error

// ValidateOutputPath prevents directory escape attacks
func ValidateOutputPath(outputDir, filename string) error

// NormalizeFormat converts format aliases to canonical form
func NormalizeFormat(format string) (string, error)
```

### Rod APIs Used

**Screenshot API**:
```go
screenshot, err := page.Screenshot(true, &proto.PageCaptureScreenshot{
    Format: proto.PageCaptureScreenshotFormatPng,
})
```

**Parameters**:
- First arg `true`: Full page capture
- Format: Always PNG (lossless, best for text/UI)
- Returns: `[]byte` of PNG data

**PDF API**:
```go
stream, err := page.PDF(&proto.PagePrintToPDF{
    PrintBackground: true,
    // Additional options as needed
})
```

**Returns**: `*StreamReader` that needs to be read into bytes

### Text Extraction Approach

**Option 1: Use existing html-to-markdown, then strip markdown**
- Pro: Reuses existing dependency
- Con: Inefficient (convert to markdown, then strip formatting)

**Option 2: Use goquery to extract text directly**
- Pro: Clean, direct approach
- Con: Adds new dependency
- Recommended: `github.com/PuerkitoBio/goquery`

**Option 3: Simple regex/string replacement**
- Pro: No dependencies
- Con: Fragile, misses edge cases

**Recommendation**: Option 2 (goquery) - clean, reliable, minimal overhead

### Testing Considerations

**New test cases needed**:

1. **Naming tests**:
   - Slugification with various titles
   - Conflict resolution with existing files
   - URL fallback when no title
   - 80 character truncation
   - Extension selection for all formats

2. **Format tests**:
   - Text extraction (verify clean output)
   - PDF generation (verify valid PDF bytes)
   - Format alias normalization (`txt` → `text`)

3. **Screenshot tests**:
   - Full page capture
   - Save to file
   - Auto-naming with timestamp

4. **Batch tests**:
   - Process multiple tabs
   - Error handling (continue on failure)
   - Conflict resolution in batch
   - Progress output
   - All format types in batch mode

5. **Validation tests**:
   - Directory existence/writability
   - Path escape attempts
   - Flag conflict detection
   - Format validation (valid and invalid)

6. **Integration tests**:
   - Combined flags (`-o` + `-d`)
   - Screenshot + directory
   - Batch with various options
   - All format types with all output methods

---

## Future Considerations (Not Phase 3)

### Phase 4 Ideas

1. **Multiple URL Arguments**:
   ```bash
   snag -d ./out https://site1.com https://site2.com https://site3.com
   ```
   - Batch fetch multiple URLs (not tabs)
   - Opens new tabs or reuses one?
   - Similar to `--all-tabs` but for fresh fetches

2. **URL Input from File**:
   ```bash
   snag -d ./out --url-file urls.txt
   ```
   - Read URLs from file (one per line)
   - Useful for large batch operations

3. **Custom Naming Templates**:
   ```bash
   snag -d ./out --name-template "{domain}-{title}" https://example.com
   ```
   - User-defined filename patterns
   - Adds complexity, evaluate need first

4. **PDF Customization**:
   ```bash
   snag -f pdf --pdf-size A4 --pdf-landscape https://example.com
   ```
   - Page size (Letter, A4, Legal)
   - Orientation (portrait, landscape)
   - Margins control

### Out of Scope

- **WebP/JPEG screenshots**: PNG only (KISS)
- **Viewport-only screenshots**: Full page only
- **Browser automation**: Snag remains passive observer
- **JavaScript execution**: Not a testing tool
- **Multiple screenshot formats**: PNG only
- **PDF advanced features**: Keep defaults simple

---

## Summary

**Phase 3 delivers**:

1. ✅ Auto-generated filenames with timestamps and title slugs
2. ✅ Output directory management with security validation
3. ✅ Text format support (plain text extraction with `text`/`txt` aliases)
4. ✅ PDF export using Chrome's print-to-PDF
5. ✅ Full-page PNG screenshot capture
6. ✅ Batch operations for all browser tabs (all formats supported)
7. ✅ Robust conflict resolution and error handling

**Maintains Snag's philosophy**:
- Simple, focused tool
- Passive observer (no automation)
- Pipe-friendly (stdout for content, stderr for logs)
- Single binary, no config files
- Clear, actionable error messages

**Format flexibility**:
- 4 content formats: markdown, html, text (txt), pdf
- 1 screenshot format: png
- All formats work with stdout, `-o`, `-d`, `-a`

**Ready for implementation**: All design decisions finalized, validation rules specified, implementation approach documented.
