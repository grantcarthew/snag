# HTML to Markdown Conversion Issues

## Overview

This document tracks issues and limitations with the HTML to Markdown conversion library used in the `snag` project. Use this as a reference when investigating library alternatives or creating upstream issues.

**Status**: Investigation Required
**Priority**: Medium
**Created**: 2025-10-18
**Last Updated**: 2025-10-18

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

**Severity**: High
**Impact**: Tables lose all structural formatting in markdown output

**Description**:
HTML tables are not converted to markdown table syntax (with pipe `|` characters). Instead, table content is concatenated without structure, making it difficult to read and losing the tabular relationship between data.

**Expected Behavior**:
```markdown
| Name  | Age | City      |
|-------|-----|-----------|
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
  <tr><td><strong>Bold</strong></td><td><a href="http://example.com">Link</a></td></tr>
  <tr><td><code>code</code></td><td><em>italic</em></td></tr>
</table>
```

Actual Output:
```markdown
**Bold**[Link](http://example.com) `code`*italic*
```

Expected Output:
```markdown
| **Bold** | [Link](http://example.com) |
| `code`   | *italic*                   |
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

## Investigation Needed

### 1. Library Configuration Options

**Action Items**:
- [ ] Review library documentation for table conversion options
- [ ] Check if table plugin exists: `github.com/JohannesKaufmann/html-to-markdown/plugin`
- [ ] Investigate `Options` struct for table-related settings
- [ ] Test with `Domain` and `Rule` configurations

**Relevant Links**:
- Plugin documentation: https://github.com/JohannesKaufmann/html-to-markdown/tree/v2/plugin
- Options reference: https://pkg.go.dev/github.com/JohannesKaufmann/html-to-markdown/v2

### 2. Alternative Libraries

If configuration doesn't resolve the issue, evaluate alternatives:

**Option A: turndown (Node.js) via exec**
- Repository: https://github.com/mixmark-io/turndown
- Pros: Mature, excellent table support, widely used
- Cons: Requires Node.js dependency, subprocess overhead
- Impact: Would require system dependency

**Option B: gomarkdown/html-to-markdown**
- Repository: https://github.com/JohannesKaufmann/html-to-markdown (same library, check older versions)
- Action: Test v1 vs v2 for table support differences

**Option C: Custom Table Handler**
- Implement custom table conversion using `go-rod`'s DOM access
- Parse table structure before HTML conversion
- Convert tables separately, then insert into markdown
- Pros: Full control, no additional dependencies
- Cons: More code to maintain

### 3. Upstream Issue

**Before Filing**:
- [ ] Search existing issues: https://github.com/JohannesKaufmann/html-to-markdown/issues
- [ ] Test with latest version (check for updates beyond v2.4.0)
- [ ] Create minimal reproduction case
- [ ] Review CONTRIBUTING.md for issue guidelines

**Issue Template Draft**:
```markdown
### Description
Tables are not converted to markdown table syntax when using ConvertString()

### Version
v2.4.0

### Code Sample
[Provide minimal reproduction]

### Expected Output
[Markdown table with pipes]

### Actual Output
[Concatenated text without structure]

### Configuration
Using default configuration with no custom options or plugins.
```

## Potential Solutions

### Solution 1: Enable Table Plugin (If Available)

**Investigation Required**: Check if table plugin exists

```go
// Hypothetical example - needs verification
import (
    "github.com/JohannesKaufmann/html-to-markdown/v2"
    "github.com/JohannesKaufmann/html-to-markdown/v2/plugin"
)

func (cc *ContentConverter) convertToMarkdown(html string) (string, error) {
    converter := htmltomarkdown.NewConverter()
    converter.Use(plugin.Table()) // If this exists

    markdown, err := converter.ConvertString(html)
    if err != nil {
        return "", err
    }

    return markdown, nil
}
```

### Solution 2: Custom Table Pre-Processing

**Approach**: Extract tables before conversion, convert separately, reinsert

```go
// Pseudocode - not implemented
func (cc *ContentConverter) convertToMarkdown(html string) (string, error) {
    // 1. Parse HTML, find all <table> elements
    // 2. Replace tables with placeholders
    // 3. Convert remaining HTML to markdown
    // 4. Convert each table to markdown table syntax
    // 5. Replace placeholders with converted tables

    // Would require HTML parsing library like golang.org/x/net/html
}
```

### Solution 3: Post-Processing Detection

**Approach**: Detect unconverted table content and warn user

```go
// Add to convert.go after conversion
if strings.Contains(originalHTML, "<table") && !strings.Contains(markdown, "|") {
    logger.Warning("Table detected but may not be formatted properly in markdown")
}
```

## Testing Requirements

### Current Test Coverage

- **File**: `convert_test.go`
- **Test**: `TestConvertToMarkdown_Tables()`
- **Status**: Passes (verifies content preservation, not structure)

### Required Tests After Fix

```go
func TestConvertToMarkdown_Tables_ProperSyntax(t *testing.T) {
    html := `<table>
        <thead><tr><th>Name</th><th>Value</th></tr></thead>
        <tbody><tr><td>Item 1</td><td>100</td></tr></tbody>
    </table>`

    converter := NewContentConverter(FormatMarkdown)
    md, err := converter.convertToMarkdown(html)

    // Assert proper markdown table syntax
    assertContains(t, md, "|")
    assertContains(t, md, "| Name | Value |")
    assertContains(t, md, "|------|-------|")
    assertContains(t, md, "| Item 1 | 100 |")
}
```

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

## Action Plan

### Phase 1: Investigation (1-2 hours)

1. **Research library capabilities**
   - Read full v2 documentation
   - Check plugin ecosystem
   - Review changelog for table-related updates
   - Test different configuration options

2. **Search existing issues**
   - Check if this is known limitation
   - Look for workarounds from other users
   - Check if fix is planned

### Phase 2: Testing (2-4 hours)

3. **Test configuration options**
   - Try all available plugins
   - Test with custom rules/domains
   - Experiment with Options struct

4. **Benchmark alternatives**
   - Test alternative libraries (if needed)
   - Compare output quality
   - Measure performance impact

### Phase 3: Implementation (4-8 hours)

5. **Implement solution**
   - Based on investigation results
   - Update `convert.go`
   - Add proper test coverage

6. **Update documentation**
   - Update AGENTS.md if behavior changes
   - Update README.md with table support notes
   - Document any new dependencies

### Phase 4: Upstream Contribution (Optional)

7. **Contribute fix upstream**
   - If issue is in library
   - File detailed issue report
   - Consider submitting PR if feasible

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

- Library is actively maintained (last update: check repository)
- V2 is a complete rewrite from V1 - check if V1 had table support
- May be intentional limitation due to markdown table complexity
- Consider if HTML passthrough is acceptable alternative for tables

---

**Next Steps**:
1. Investigate library documentation for table plugin
2. Test with latest version
3. Search existing issues
4. Update this document with findings

**Owner**: TBD
**Estimated Effort**: 8-16 hours (investigation + implementation)
