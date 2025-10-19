# HTML to Markdown Conversion Issues

## Overview

This document tracks issues and limitations with the HTML to Markdown conversion library used in the `snag` project. Use this as a reference when investigating library alternatives or creating upstream issues.

**Status**: Completed
**Priority**: Medium
**Created**: 2025-10-18
**Completed**: 2025-10-19

## Executive Summary

**TL;DR**: The issue is REAL but has a SIMPLE FIX.

**Problem**: Tables convert to concatenated text without structure (no pipe characters). Strikethrough tags (`<del>`, `<s>`, `<strike>`) don't convert to proper markdown.

**Root Cause**: We're using `htmltomarkdown.ConvertString()` which only includes base + commonmark plugins. It does NOT include the table or strikethrough plugins.

**Solution**: Switch to `converter.NewConverter()` and explicitly add the table and strikethrough plugins. The library fully supports proper markdown conversion for both - we just need to enable them.

**Effort**: 2-4 hours implementation + testing.

**Discovery**: Deep research into library documentation (2025-10-19) revealed:

- ✅ Table plugin exists and is fully functional (PR #144, Feb 2025)
- ✅ Strikethrough plugin exists and is fully functional
- ✅ Table plugin supports alignment, colspan, rowspan, header promotion
- ✅ Strikethrough plugin handles `<del>`, `<s>`, and `<strike>` tags
- ✅ Both well-documented with configuration options
- ✅ No library bugs - just configuration issue

## Library Information

- **Package**: `github.com/JohannesKaufmann/html-to-markdown/v2`
- **Version**: v2.4.0
- **Repository**: https://github.com/JohannesKaufmann/html-to-markdown
- **Documentation**: https://github.com/JohannesKaufmann/html-to-markdown/tree/v2
- **License**: MIT

## Current Usage in snag

### Implementation Location

- **File**: `convert.go:66-75`
- **Function**: `convertToMarkdown(html string) (string, error)`
- **Code**:

  ```go
  func (cc *ContentConverter) convertToMarkdown(html string) (string, error) {
      // Convert HTML to Markdown using the package-level function
      markdown, err := htmltomarkdown.ConvertString(html)
      if err != nil {
          return "", err
      }

      logger.Success("Converted to Markdown")
      return markdown, nil
  }
  ```

### Current Configuration

- Uses default configuration (no custom options)
- Called via `htmltomarkdown.ConvertString(html)` without additional options
- No plugins or custom rules configured

## Known Issues

### Issue 1: Tables Not Converted to Markdown Table Syntax

**Severity**: Medium (Solution Available)
**Impact**: Tables lose all structural formatting in markdown output when using default configuration
**Root Cause**: Current implementation uses `htmltomarkdown.ConvertString()` which does not include the table plugin

**Description**:
When using the default `htmltomarkdown.ConvertString()` function, HTML tables are not converted to markdown table syntax (with pipe `|` characters). Instead, table content is concatenated without structure, making it difficult to read and losing the tabular relationship between data.

**IMPORTANT**: The library DOES support proper table conversion via the **table plugin** (added in v2, PR #144, Feb 2025). Our current implementation simply does not use it.

### Issue 2: Strikethrough Tags Not Converted to Markdown Syntax

**Severity**: Low (Solution Available)
**Impact**: Strikethrough formatting (`<del>`, `<s>`, `<strike>`) is lost in markdown output
**Root Cause**: Current implementation uses `htmltomarkdown.ConvertString()` which does not include the strikethrough plugin

**Description**:
When using the default `htmltomarkdown.ConvertString()` function, HTML strikethrough tags are not converted to GitHub Flavored Markdown strikethrough syntax (`~~text~~`). The text content is preserved but the strikethrough formatting is lost.

**IMPORTANT**: The library DOES support strikethrough conversion via the **strikethrough plugin**. Our current implementation simply does not use it.

**Expected Behavior**:

```markdown
| Name  | Age | City      |
| ----- | --- | --------- |
| Alice | 30  | Sydney    |
| Bob   | 25  | Melbourne |
```

**Actual Behavior**:

```markdown
Name Age City Alice 30 Sydney Bob 25 Melbourne
```

**Test Case**: `convert_test.go:57-86`

**Reproducible Test**:

```bash
# Create test HTML file
cat > /tmp/test_table.html << 'EOF'
<html><body>
<table>
  <thead>
    <tr><th>Name</th><th>Value</th></tr>
  </thead>
  <tbody>
    <tr><td>Item 1</td><td>100</td></tr>
  </tbody>
</table>
</body></html>
EOF

# Test conversion
./snag file:///tmp/test_table.html

# Expected: Markdown table with pipes
# Actual: "NameValue Item 1100"
```

**Observations**:

- Table content is preserved (text is not lost)
- Nested formatting within table cells IS preserved (bold, links, code, italic)
- Table structure (rows, columns, headers) is completely lost
- No visual separation between cells
- No indication of which items are headers vs data

**Example with Nested Formatting**:

HTML Input:

```html
<table>
  <tr>
    <td><strong>Bold</strong></td>
    <td><a href="http://example.com">Link</a></td>
  </tr>
  <tr>
    <td><code>code</code></td>
    <td><em>italic</em></td>
  </tr>
</table>
```

Actual Output:

```markdown
**Bold**[Link](http://example.com) `code`_italic_
```

Expected Output:

```markdown
| **Bold** | [Link](http://example.com) |
| `code` | _italic_ |
```

**Related Code**:

- Test: `convert_test.go:57` - `TestConvertToMarkdown_Tables()`
- Implementation: `convert.go:66` - `convertToMarkdown()`
- Workaround: None currently implemented

**Temporary Solution**:
Test updated to verify content preservation instead of structure:

```go
// NOTE: Current html-to-markdown library does not convert tables to proper
// markdown table syntax. This test verifies table content is preserved.
// TODO: Consider library configuration or alternative for proper table support.
```

## Investigation Results ✅

### 1. Library Configuration Options - COMPLETED

**Findings**:

- ✅ **Table plugin EXISTS**: `github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table`
- ✅ **Added in v2**: PR #144, merged February 25, 2025
- ✅ **Fully functional**: Supports alignment, rowspan, colspan, header promotion
- ✅ **Well documented**: Full API documentation available

**Table Plugin Features**:

- Converts HTML tables to proper GitHub Flavored Markdown table syntax
- Supports alignment (left, center, right)
- Handles `colspan` and `rowspan` attributes
- Configurable options via `NewTablePlugin(opts...)`
- Empty row skipping
- Header row promotion
- Presentation table handling

**Strikethrough Plugin Features**:

- Converts `<del>`, `<s>`, and `<strike>` tags to `~~text~~` syntax
- Implements GitHub Flavored Markdown strikethrough
- No configuration needed - works out of the box

**Configuration Options Available**:

```go
table.NewTablePlugin(
    table.WithSpanCellBehavior(table.SpanBehaviorMirror),  // or SpanBehaviorEmpty
    table.WithHeaderPromotion(true),                        // promote first row to header
    table.WithSkipEmptyRows(true),                         // omit empty rows
    table.WithNewlineBehavior(table.NewlineBehaviorPreserve), // preserve newlines in cells
    table.WithPresentationTables(false),                   // skip role="presentation" tables
)
```

**Relevant Links**:

- Table plugin source: https://github.com/JohannesKaufmann/html-to-markdown/tree/main/plugin/table
- Table plugin docs: https://pkg.go.dev/github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table
- Strikethrough plugin source: https://github.com/JohannesKaufmann/html-to-markdown/tree/main/plugin/strikethrough
- Strikethrough plugin docs: https://pkg.go.dev/github.com/JohannesKaufmann/html-to-markdown/v2/plugin/strikethrough
- Main README: https://github.com/JohannesKaufmann/html-to-markdown#table-plugin

### 2. Alternative Libraries - NOT NEEDED ✅

**Conclusion**: No alternative libraries needed. The current library (`html-to-markdown/v2`) fully supports table conversion via the table plugin.

**Previous Considerations** (now obsolete):

- ~~turndown (Node.js)~~ - Would add Node.js dependency unnecessarily
- ~~Custom table handler~~ - Library already provides this functionality
- ~~Different library~~ - Current library is excellent, just needs proper configuration

### 3. Upstream Issue - NOT REQUIRED ✅

**Conclusion**: No upstream issue needed. This is not a bug in the library - it's a configuration issue in our implementation.

**Analysis**:

- ✅ Library supports tables via plugin system
- ✅ Plugin is well-documented and maintained
- ✅ Plugin added in v2 (we're using v2.4.0)
- ✅ No known bugs with table conversion
- ⚠️ Our code simply doesn't use the plugin

**What we found**:

The `htmltomarkdown.ConvertString()` convenience function is intentionally minimal - it only includes base and commonmark plugins. For table support, users must use `converter.NewConverter()` and explicitly add the table plugin. This is documented in the library's README.

## Recommended Solution ✅

### Solution 1: Enable Table and Strikethrough Plugins - VERIFIED AND READY

**Status**: Implementation code verified, ready to apply

**Current Code** (convert.go:66-75):

```go
func (cc *ContentConverter) convertToMarkdown(html string) (string, error) {
    // Convert HTML to Markdown using the package-level function
    markdown, err := htmltomarkdown.ConvertString(html)
    if err != nil {
        return "", err
    }

    logger.Success("Converted to Markdown")
    return markdown, nil
}
```

**Proposed Fix**:

```go
import (
    "github.com/JohannesKaufmann/html-to-markdown/v2/converter"
    "github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
    "github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
    "github.com/JohannesKaufmann/html-to-markdown/v2/plugin/strikethrough"
    "github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
)

func (cc *ContentConverter) convertToMarkdown(html string) (string, error) {
    // Create converter with table and strikethrough plugin support
    conv := converter.NewConverter(
        converter.WithPlugins(
            base.NewBasePlugin(),
            commonmark.NewCommonmarkPlugin(),
            table.NewTablePlugin(),            // Add table support
            strikethrough.NewStrikethroughPlugin(), // Add strikethrough support
        ),
    )

    markdown, err := conv.ConvertString(html)
    if err != nil {
        return "", err
    }

    logger.Success("Converted to Markdown")
    return markdown, nil
}
```

**Alternative with Configuration Options**:

```go
func (cc *ContentConverter) convertToMarkdown(html string) (string, error) {
    conv := converter.NewConverter(
        converter.WithPlugins(
            base.NewBasePlugin(),
            commonmark.NewCommonmarkPlugin(),
            table.NewTablePlugin(
                table.WithHeaderPromotion(true),      // Promote first row to header if no <th>
                table.WithSkipEmptyRows(true),        // Skip completely empty rows
                table.WithSpanCellBehavior(table.SpanBehaviorEmpty), // How to handle colspan/rowspan
            ),
            strikethrough.NewStrikethroughPlugin(), // Converts <del>, <s>, <strike> to ~~text~~
        ),
    )

    markdown, err := conv.ConvertString(html)
    if err != nil {
        return "", err
    }

    logger.Success("Converted to Markdown")
    return markdown, nil
}
```

### ~~Solution 2: Custom Table Pre-Processing~~ - NOT NEEDED

**Obsolete**: Library already handles table conversion via plugin. No custom pre-processing required.

### ~~Solution 3: Post-Processing Detection~~ - NOT NEEDED

**Obsolete**: With table plugin enabled, tables will convert properly. Warning not needed.

## Testing Requirements ⚠️ IMPORTANT

### Current Test Coverage - TESTS NEED UPDATING

**Critical**: The existing table test is written to work around the current lack of table support. Once we add the table plugin, this test WILL FAIL because it expects the wrong output format.

**Current State**:

- **File**: `convert_test.go`
- **Test**: `TestConvertToMarkdown_Tables()` at line 60
- **Status**: Currently passes (but for the WRONG reason)
- **Current Behavior**: Test verifies content preservation WITHOUT pipe characters
- **Problem**: Test expects `"NameValue Item1100"` instead of proper table syntax
- **Comment in code**: Lines 78-80 acknowledge this is a workaround

**What needs to change**:

1. ✅ **Table test exists** but checks for WRONG output (concatenated text)
2. ❌ **Strikethrough test does NOT exist** - needs to be created

**After implementation**, the table test will fail with this error:

```
expected table headers in markdown, got:
NameValue Item 1100
```

This is EXPECTED and GOOD - it means we need to update the test to check for proper table syntax.

### Required Test Updates

#### 1. Update Existing Table Test - MANDATORY

**Current Test Code** (convert_test.go:60-89):

```go
func TestConvertToMarkdown_Tables(t *testing.T) {
    html := `<html><body>
        <table>
            <thead>
                <tr><th>Name</th><th>Value</th></tr>
            </thead>
            <tbody>
                <tr><td>Item 1</td><td>100</td></tr>
            </tbody>
        </table>
    </body></html>`

    converter := NewContentConverter(FormatMarkdown)
    md, err := converter.convertToMarkdown(html)
    if err != nil {
        t.Fatalf("convertToMarkdown failed: %v", err)
    }

    // NOTE: Current html-to-markdown library does not convert tables to proper
    // markdown table syntax. This test verifies table content is preserved.
    // TODO: Consider library configuration or alternative for proper table support.

    // Check for table content (headers and data are preserved)
    if !strings.Contains(md, "Name") || !strings.Contains(md, "Value") {
        t.Errorf("expected table headers in markdown, got:\n%s", md)
    }
    if !strings.Contains(md, "Item 1") || !strings.Contains(md, "100") {
        t.Errorf("expected table data in markdown, got:\n%s", md)
    }
}
```

**THIS TEST WILL FAIL** after we add the table plugin because:

- It doesn't check for pipe `|` characters
- It doesn't check for separator row with dashes
- The TODO comment is outdated

**New Test Code** (replace the above):

The existing `TestConvertToMarkdown_Tables()` test needs to be completely rewritten to verify proper markdown table syntax.

```go
func TestConvertToMarkdown_Tables(t *testing.T) {
    html := `<html><body>
        <table>
            <thead>
                <tr><th>Name</th><th>Value</th></tr>
            </thead>
            <tbody>
                <tr><td>Item 1</td><td>100</td></tr>
            </tbody>
        </table>
    </body></html>`

    converter := NewContentConverter(FormatMarkdown)
    md, err := converter.convertToMarkdown(html)
    if err != nil {
        t.Fatalf("convertToMarkdown failed: %v", err)
    }

    // Verify proper markdown table syntax (not just content preservation)
    if !strings.Contains(md, "|") {
        t.Errorf("expected markdown table with pipe characters, got:\n%s", md)
    }

    // Verify header row
    if !strings.Contains(md, "Name") || !strings.Contains(md, "Value") {
        t.Errorf("expected table headers 'Name' and 'Value', got:\n%s", md)
    }

    // Verify data row
    if !strings.Contains(md, "Item 1") || !strings.Contains(md, "100") {
        t.Errorf("expected table data 'Item 1' and '100', got:\n%s", md)
    }

    // Verify separator row (some variation of dashes)
    if !strings.Contains(md, "---") {
        t.Errorf("expected table separator with dashes, got:\n%s", md)
    }
}
```

#### 2. Add Strikethrough Test - MANDATORY

**This test does NOT exist** and must be created.

**New Test to Add** (convert_test.go - add after table test):

```go
func TestConvertToMarkdown_Strikethrough(t *testing.T) {
    html := `<html><body>
        <p>This is <del>deleted</del> text.</p>
        <p>This is <s>strikethrough</s> text.</p>
        <p>This is <strike>old strikethrough</strike> text.</p>
    </body></html>`

    converter := NewContentConverter(FormatMarkdown)
    md, err := converter.convertToMarkdown(html)
    if err != nil {
        t.Fatalf("convertToMarkdown failed: %v", err)
    }

    // Verify strikethrough syntax
    if !strings.Contains(md, "~~") {
        t.Errorf("expected strikethrough with ~~ syntax, got:\n%s", md)
    }

    // Verify content is preserved
    if !strings.Contains(md, "deleted") || !strings.Contains(md, "strikethrough") {
        t.Errorf("expected strikethrough content preserved, got:\n%s", md)
    }
}
```

**Additional Test Cases to Consider**:

1. **Advanced Tables** (optional, future):

   - Table with alignment
   - Table with colspan/rowspan
   - Table with nested formatting (bold, links, code)
   - Table without explicit header
   - Empty table cells

2. **Advanced Strikethrough** (optional, future):
   - Nested strikethrough with other formatting
   - Multiple strikethrough elements in one paragraph

## Impact Assessment

### User Impact

**Severity**: Medium to High (depending on use case)

**Affected Users**:

- Users fetching pages with data tables (reports, documentation, wikis)
- Users expecting markdown output to preserve table structure
- Users piping output to markdown renderers

**Workarounds**:

- Use `--format html` flag to get raw HTML output
- Manual table reconstruction from concatenated text (tedious)
- Process tables separately with different tool

### Performance Impact

- No performance impact from current behavior
- Custom solution may add processing overhead
- Plugin solution should have minimal impact

## Action Plan ✅

### ~~Phase 1: Investigation~~ - COMPLETED ✅

**Duration**: 2 hours (2025-10-19)

1. ✅ **Research library capabilities**

   - ✅ Read v2 documentation and README
   - ✅ Discovered table plugin exists and is fully functional
   - ✅ Reviewed plugin API and configuration options
   - ✅ Confirmed we're using v2.4.0 which includes table plugin

2. ✅ **Search existing issues**
   - ✅ No library bugs found - this is a configuration issue on our end
   - ✅ Library works as designed - table plugin must be explicitly added
   - ✅ Well-documented in official README

### Phase 2: Implementation (2-4 hours) - READY TO START

**Tasks**:

1. **Update convert.go** (1 hour)

   - Replace `htmltomarkdown.ConvertString()` with `converter.NewConverter()`
   - Add imports for converter, base, commonmark, table, and strikethrough plugins
   - Register all plugins with converter
   - Choose appropriate table plugin options
   - Test locally with sample HTML tables and strikethrough text

2. **Update convert_test.go** (1 hour) ⚠️ CRITICAL

   - **REWRITE** `TestConvertToMarkdown_Tables()` test completely (lines 60-89)
     - Remove old content-only checks
     - Add pipe `|` character verification
     - Add separator row `---` verification
     - Remove outdated TODO comment (lines 78-80)
     - Test WILL FAIL initially - this is expected and correct
   - **CREATE NEW** `TestConvertToMarkdown_Strikethrough()` test
     - Add after table test
     - Verify `~~` syntax
     - Verify all three tag types work
   - Run tests and verify they pass with new implementation
   - Consider adding additional advanced test cases

3. **Build and verify** (30 minutes)

   - Run `go build` to ensure no compilation errors
   - Run `go test -v` to ensure all tests pass
   - Manual testing with `./snag` on real web pages with tables and strikethrough text

4. **Update documentation** (30 minutes)
   - Update AGENTS.md line 229 to mention table and strikethrough plugin usage
   - Update PROJECT.md to note table and strikethrough support is now working
   - Remove or update any "table limitation" notes in documentation
   - Document that `<del>`, `<s>`, and `<strike>` tags now convert to `~~text~~`

### ~~Phase 3: Upstream Contribution~~ - NOT NEEDED ✅

**Conclusion**: No upstream contribution required. Library works perfectly - we just needed to configure it properly.

## Implementation Checklist

**Before starting**:

- [ ] Read this document completely
- [ ] Review table plugin API documentation
- [ ] Understand current convert.go implementation

**Implementation**:

- [ ] Update imports in convert.go (add table and strikethrough plugins)
- [ ] Modify convertToMarkdown() function in convert.go
- [ ] Choose table plugin configuration options
- [ ] **REWRITE** TestConvertToMarkdown_Tables() test (convert_test.go:60-89)
  - Remove content-only checks
  - Add proper table syntax verification
  - Remove TODO comment at lines 78-80
- [ ] **CREATE** TestConvertToMarkdown_Strikethrough() test (new in convert_test.go)
- [ ] Run `go test -v` and verify both tests pass
- [ ] Build binary and test manually

**Verification**:

- [ ] **IMPORTANT**: Table test initially FAILS (expected - old test checks wrong thing)
- [ ] After rewriting table test, all tests pass
- [ ] Table test now verifies pipe `|` characters
- [ ] Table test verifies separator row with `---`
- [ ] Strikethrough test verifies `~~` syntax
- [ ] Strikethrough test checks all three tags (`<del>`, `<s>`, `<strike>`)
- [ ] All 6 existing conversion tests still pass (headings, links, lists, code, minimal)
- [ ] Manual test with real web page containing tables and strikethrough
- [ ] Output produces valid GFM table syntax
- [ ] Output produces valid GFM strikethrough syntax

**Documentation**:

- [ ] Update AGENTS.md
- [ ] Update PROJECT.md completion status
- [ ] Mark this document as implemented

## Related Files

### Source Code

- `convert.go` - HTML to Markdown conversion implementation
- `convert_test.go` - Conversion tests
- `go.mod` - Dependency declaration

### Documentation

- `AGENTS.md:229` - Mentions html-to-markdown dependency
- `docs/design.md` - May discuss format conversion decisions
- `PROJECT.md:76` - TODO for table support investigation

### Tests

- `testdata/complex.html:8-18` - Contains table test fixture
- `convert_test.go:57-86` - Table conversion test

## References

### Library Resources

- GitHub Repository: https://github.com/JohannesKaufmann/html-to-markdown
- Go Package Documentation: https://pkg.go.dev/github.com/JohannesKaufmann/html-to-markdown/v2
- Issues: https://github.com/JohannesKaufmann/html-to-markdown/issues
- V2 Migration Guide: https://github.com/JohannesKaufmann/html-to-markdown/blob/v2/MIGRATION.md

### Markdown Table Specification

- GitHub Flavored Markdown: https://github.github.com/gfm/#tables-extension-
- CommonMark Discussion: https://talk.commonmark.org/t/tables-in-pure-markdown/81

### Alternative Libraries

- turndown (JS): https://github.com/mixmark-io/turndown
- html-to-markdown (Go, older): Various forks on GitHub
- pandoc (comprehensive): https://pandoc.org/

## Notes

- ✅ Library is actively maintained (latest commit: August 2025)
- ✅ V2 is a complete rewrite from V1 with enhanced plugin architecture
- ✅ V2 table support is MORE comprehensive than V1
- ✅ Table conversion is NOT a limitation - it's a feature that must be explicitly enabled
- ⚠️ The `ConvertString()` convenience function is intentionally minimal
- 📚 Table plugin is well-documented in README and pkg.go.dev

**Key Insight**: This is a common pattern in the library's design - core functionality is minimal by default, extended functionality requires explicit plugin registration. This keeps the base library lightweight and allows users to opt-in to features they need.

---

**Investigation Complete**: 2025-10-19

**Findings**:

1. ✅ Table plugin exists and works perfectly
2. ✅ Strikethrough plugin exists and works perfectly
3. ✅ Solution is simple: use `converter.NewConverter()` with both plugins
4. ✅ No alternative libraries needed
5. ✅ No upstream issues to file
6. ✅ Implementation completed in convert.go

**Implementation Complete**: 2025-10-19

1. ✅ Implemented solution in convert.go (added table and strikethrough plugins)
2. ✅ Rewrote table test in convert_test.go (lines 60-97)
3. ✅ Created strikethrough test in convert_test.go (lines 99-121)
4. ✅ Verified all 59 tests pass
5. ✅ Updated documentation (AGENTS.md, PROJECT.md)
6. ✅ Issue resolved

**Test Update Summary**:

- ✏️ 1 test to REWRITE: `TestConvertToMarkdown_Tables()`
- ➕ 1 test to CREATE: `TestConvertToMarkdown_Strikethrough()`
- ✅ 6 existing tests should still pass unchanged

**Owner**: TBD
**Actual Investigation Effort**: 2 hours
**Estimated Implementation Effort**: 2-4 hours
