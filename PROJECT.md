# Test Coverage Enhancement Project

**Project Goal**: Enhance test coverage for the snag CLI tool by adding 45-50 new unit tests without modifying any production `.go` code.

**Status**: Planning Complete - Ready for Implementation
**Estimated Effort**: 8-12 hours
**Priority**: High
**Created**: 2025-10-31

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Current State Analysis](#current-state-analysis)
3. [Project Constraints](#project-constraints)
4. [Project Phases](#project-phases)
5. [Detailed Task Specifications](#detailed-task-specifications)
6. [Test Templates and Examples](#test-templates-and-examples)
7. [Success Criteria](#success-criteria)
8. [Testing Guidelines](#testing-guidelines)

---

## Project Overview

### Background

The snag CLI tool is a Go application that fetches web content via Chrome DevTools Protocol. While the project has good integration test coverage via `cli_test.go`, many pure functions and utility methods lack dedicated unit tests.

### Objectives

1. **Create 2 new test files** for untested modules (`fetch_test.go`, `errors_test.go`)
2. **Add ~26 tests** to existing test files to cover edge cases and security scenarios
3. **Achieve this without modifying any production `.go` code** - only pure function testing
4. **Maintain Go testing best practices** - clear test names, table-driven tests, proper assertions

### Key Metrics

- **Target**: 45-50 new tests
- **New files**: 2 (`fetch_test.go`, `errors_test.go`)
- **Enhanced files**: 4 (`formats_test.go`, `handlers_test.go`, `validate_test.go`, `output_test.go`)
- **Code changes**: 0 production files modified

---

## Current State Analysis

### Test Coverage Summary

| File | Test File | Status | Missing Tests |
|------|-----------|--------|---------------|
| `main.go` | `cli_test.go` | ✅ Good (integration) | N/A (requires DI) |
| `browser.go` | `browser_test.go` | ✅ Good | N/A (requires mocking) |
| **`fetch.go`** | ❌ **NONE** | ⚠️ Critical Gap | **~18 tests** |
| `formats.go` | `formats_test.go` | ⚠️ Partial | **~16 tests** |
| `handlers.go` | `handlers_test.go` | ✅ Good | **~3 tests** |
| `output.go` | `output_test.go` | ✅ Excellent | **~3 tests** |
| `logger.go` | `logger_test.go` | ✅ Excellent | 0 |
| `doctor.go` | `doctor_test.go` | ✅ Excellent | 0 |
| `validate.go` | `validate_test.go` | ✅ Very Good | **~4 tests** |
| **`errors.go`** | ❌ **NONE** | ⚠️ Critical Gap | **~7 tests** |

### Testable vs Non-Testable Functions

**Testable (Pure Functions)**:
- `hasLoginIndicators()` in fetch.go
- `hasPasswordField()` in fetch.go
- `wrapError()` in errors.go
- `convertToMarkdown()` in formats.go (with mock HTML)
- `extractPlainText()` in formats.go (with mock HTML)
- All validation functions in validate.go
- All formatting functions in output.go and handlers.go

**Non-Testable (Without Code Changes)**:
- Functions requiring browser connection (CDP)
- Functions with hard-coded file I/O
- Functions with direct HTTP calls
- Functions with global state dependencies

---

## Project Constraints

### Must Follow

1. ✅ **Zero Production Code Changes** - Only add/modify `_test.go` files
2. ✅ **No External Test Dependencies** - Use only Go standard library + existing project dependencies
3. ✅ **Follow Existing Test Patterns** - Match coding style of current test files
4. ✅ **All Tests Must Pass** - Run `go test ./...` successfully
5. ✅ **Table-Driven Tests** - Use table-driven test pattern where applicable

### Testing Standards

- Use `t.Helper()` for assertion helper functions
- Use descriptive test names: `TestFunctionName_Scenario`
- Use table-driven tests for multiple cases
- Include both positive and negative test cases
- Test edge cases: empty strings, nil values, boundary conditions
- Include security test cases where applicable

---

## Project Phases

### Phase 1: Create Missing Test Files (High Priority)

**Estimated Time**: 3-4 hours

#### Task 1.1: Create `fetch_test.go`
- Create new test file with package header and copyright
- Add ~18 tests for pure functions
- Focus on `hasLoginIndicators()`, `hasPasswordField()`

#### Task 1.2: Create `errors_test.go`
- Create new test file with package header and copyright
- Add ~7 tests for error handling
- Focus on `wrapError()` and error type checking

---

### Phase 2: Enhance formats_test.go (Medium Priority)

**Estimated Time**: 2-3 hours

#### Task 2.1: Add Markdown Conversion Edge Cases
- Add 16 new test cases for complex HTML scenarios
- Cover nested lists, tables, images, blockquotes, definition lists
- Test malformed HTML and special characters

---

### Phase 3: Enhance Other Test Files (Medium Priority)

**Estimated Time**: 2-3 hours

#### Task 3.1: Enhance `handlers_test.go`
- Add 3 tests for large lists and Unicode handling

#### Task 3.2: Enhance `validate_test.go`
- Add 4 security-focused tests (IDN attacks, extremely long URLs)

#### Task 3.3: Enhance `output_test.go`
- Add 3 tests for Unicode normalization and high conflict counts

---

### Phase 4: Verification and Cleanup (Low Priority)

**Estimated Time**: 1 hour

#### Task 4.1: Run Complete Test Suite
- Execute `go test ./...`
- Verify all tests pass
- Check test coverage with `go test -cover ./...`

#### Task 4.2: Code Review and Formatting
- Run `gofmt` on all test files
- Verify consistent style
- Add any missing documentation

---

## Detailed Task Specifications

### Phase 1, Task 1.1: Create fetch_test.go

**Location**: `/Users/gcarthew/Projects/snag/fetch_test.go`

**Required Tests**:

#### Test 1.1.1: `TestHasLoginIndicators`
Test the `hasLoginIndicators()` function which detects login-related keywords in HTML.

**Test Cases** (10 tests):
1. HTML with "sign in" (lowercase)
2. HTML with "Sign In" (title case)
3. HTML with "LOG IN" (uppercase)
4. HTML with "login" (single word)
5. HTML with "signin" (no space)
6. HTML with "authentication"
7. HTML with "Log In to Your Account" (in context)
8. HTML without login keywords - should return false
9. Empty HTML string - should return false
10. Very long HTML with login keyword at end - should return true

**Example Test Structure**:
```go
func TestHasLoginIndicators(t *testing.T) {
    tests := []struct {
        name     string
        html     string
        expected bool
    }{
        {
            name:     "lowercase sign in",
            html:     "<html><body><h1>sign in</h1></body></html>",
            expected: true,
        },
        {
            name:     "no login keywords",
            html:     "<html><body><h1>Welcome to Example</h1></body></html>",
            expected: false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := hasLoginIndicators(tt.html)
            if result != tt.expected {
                t.Errorf("hasLoginIndicators() = %v, expected %v", result, tt.expected)
            }
        })
    }
}
```

#### Test 1.1.2: `TestHasPasswordField`
Test the `hasPasswordField()` function which detects password input fields in HTML.

**Test Cases** (8 tests):
1. HTML with `<input type="password">`
2. HTML with `<input type="PASSWORD">` (uppercase)
3. HTML with `<input TYPE="password">` (uppercase attribute)
4. HTML with multiple password fields
5. HTML with `<input type="text">` only - should return false
6. HTML with "password" in text but no input field - should return false
7. Empty HTML - should return false
8. Malformed HTML with unclosed tags but valid password input

**Example Test Structure**:
```go
func TestHasPasswordField(t *testing.T) {
    tests := []struct {
        name     string
        html     string
        expected bool
    }{
        {
            name:     "standard password input",
            html:     `<input type="password" name="pwd">`,
            expected: true,
        },
        {
            name:     "uppercase type attribute",
            html:     `<input type="PASSWORD">`,
            expected: true,
        },
        {
            name:     "multiple password fields",
            html:     `<input type="password"><input type="password">`,
            expected: true,
        },
        {
            name:     "no password field",
            html:     `<input type="text" name="username">`,
            expected: false,
        },
        {
            name:     "password in text only",
            html:     `<p>Please enter your password</p>`,
            expected: false,
        },
        {
            name:     "empty html",
            html:     "",
            expected: false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := hasPasswordField(tt.html)
            if result != tt.expected {
                t.Errorf("hasPasswordField() = %v, expected %v for HTML: %s", result, tt.expected, tt.html)
            }
        })
    }
}
```

**Note**: The function `detectLoginForm()` requires CDP mocking, so skip it for this project.

---

### Phase 1, Task 1.2: Create errors_test.go

**Location**: `/Users/gcarthew/Projects/snag/errors_test.go`

**Required Tests**:

#### Test 1.2.1: `TestWrapError`
Test the `wrapError()` function which wraps errors with additional context.

**Test Cases** (5 tests):
1. Wrap a standard error with context message
2. Wrap a nil error (should still create error with context)
3. Wrap an already-wrapped error (nested wrapping)
4. Wrap with empty context message
5. Wrap complex error chain and verify full message

**Example Test Structure**:
```go
func TestWrapError(t *testing.T) {
    tests := []struct {
        name        string
        err         error
        msg         string
        shouldBeNil bool
        contains    []string
    }{
        {
            name:        "wrap standard error",
            err:         errors.New("original error"),
            msg:         "context message",
            shouldBeNil: false,
            contains:    []string{"context message", "original error"},
        },
        {
            name:        "wrap nil error",
            err:         nil,
            msg:         "context message",
            shouldBeNil: false,
            contains:    []string{"context message"},
        },
        {
            name:        "wrap wrapped error",
            err:         wrapError(errors.New("base"), "first wrap"),
            msg:         "second wrap",
            shouldBeNil: false,
            contains:    []string{"second wrap", "first wrap", "base"},
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := wrapError(tt.err, tt.msg)

            if tt.shouldBeNil && result != nil {
                t.Errorf("expected nil error, got: %v", result)
            }

            if !tt.shouldBeNil && result == nil {
                t.Error("expected non-nil error, got nil")
            }

            if result != nil {
                errStr := result.Error()
                for _, substr := range tt.contains {
                    if !strings.Contains(errStr, substr) {
                        t.Errorf("error message %q should contain %q", errStr, substr)
                    }
                }
            }
        })
    }
}
```

#### Test 1.2.2: `TestErrorConstants`
Test that all error constants are properly defined.

**Test Cases** (1 test with 8 checks):
```go
func TestErrorConstants(t *testing.T) {
    errors := []struct {
        name string
        err  error
    }{
        {"ErrBrowserNotFound", ErrBrowserNotFound},
        {"ErrBrowserNotRunning", ErrBrowserNotRunning},
        {"ErrPortInUse", ErrPortInUse},
        {"ErrTimeout", ErrTimeout},
        {"ErrInvalidURL", ErrInvalidURL},
        {"ErrInvalidFormat", ErrInvalidFormat},
        {"ErrConnectionFailed", ErrConnectionFailed},
        {"ErrAuthRequired", ErrAuthRequired},
    }

    for _, tc := range errors {
        t.Run(tc.name, func(t *testing.T) {
            if tc.err == nil {
                t.Errorf("%s is nil", tc.name)
            }
            if tc.err.Error() == "" {
                t.Errorf("%s has empty error message", tc.name)
            }
        })
    }
}
```

#### Test 1.2.3: `TestErrorTypeChecking`
Test that error type checking works with `errors.Is()`.

**Test Cases** (1 test):
```go
func TestErrorTypeChecking(t *testing.T) {
    // Test that errors.Is works correctly
    if !errors.Is(ErrBrowserNotFound, ErrBrowserNotFound) {
        t.Error("errors.Is should return true for same error")
    }

    if errors.Is(ErrBrowserNotFound, ErrBrowserNotRunning) {
        t.Error("errors.Is should return false for different errors")
    }

    // Test wrapped error
    wrapped := wrapError(ErrTimeout, "additional context")
    if !errors.Is(wrapped, ErrTimeout) {
        t.Error("errors.Is should work with wrapped errors")
    }
}
```

---

### Phase 2, Task 2.1: Enhance formats_test.go

**Location**: `/Users/gcarthew/Projects/snag/formats_test.go`

**Required Tests** (16 new tests):

#### Test 2.1.1: `TestConvertToMarkdown_NestedLists`
**Test Cases** (2 tests):
1. Nested unordered lists (ul > li > ul)
2. Nested ordered lists (ol > li > ol)

```go
func TestConvertToMarkdown_NestedLists(t *testing.T) {
    html := `<html><body>
        <ul>
            <li>Item 1
                <ul>
                    <li>Nested 1.1</li>
                    <li>Nested 1.2</li>
                </ul>
            </li>
            <li>Item 2</li>
        </ul>
    </body></html>`

    converter := NewContentConverter(FormatMarkdown)
    md, err := converter.convertToMarkdown(html)
    if err != nil {
        t.Fatalf("convertToMarkdown failed: %v", err)
    }

    // Verify nested structure is preserved
    if !strings.Contains(md, "Item 1") {
        t.Errorf("expected 'Item 1' in output, got:\n%s", md)
    }
    if !strings.Contains(md, "Nested 1.1") {
        t.Errorf("expected nested items in output, got:\n%s", md)
    }
}
```

#### Test 2.1.2: `TestConvertToMarkdown_ComplexTables`
**Test Cases** (2 tests):
1. Table with colspan/rowspan
2. Table with header and footer

```go
func TestConvertToMarkdown_ComplexTables(t *testing.T) {
    html := `<html><body>
        <table>
            <thead>
                <tr><th colspan="2">Header</th></tr>
            </thead>
            <tbody>
                <tr><td>Cell 1</td><td>Cell 2</td></tr>
            </tbody>
            <tfoot>
                <tr><td colspan="2">Footer</td></tr>
            </tfoot>
        </table>
    </body></html>`

    converter := NewContentConverter(FormatMarkdown)
    md, err := converter.convertToMarkdown(html)
    if err != nil {
        t.Fatalf("convertToMarkdown failed: %v", err)
    }

    // Verify table content is preserved
    if !strings.Contains(md, "Header") {
        t.Errorf("expected 'Header' in output, got:\n%s", md)
    }
    if !strings.Contains(md, "Cell 1") || !strings.Contains(md, "Cell 2") {
        t.Errorf("expected cell content in output, got:\n%s", md)
    }
}
```

#### Test 2.1.3: `TestConvertToMarkdown_Images`
**Test Cases** (2 tests):
1. Image with alt text
2. Multiple images

```go
func TestConvertToMarkdown_Images(t *testing.T) {
    html := `<html><body>
        <img src="https://example.com/image.png" alt="Example Image">
        <p>Some text</p>
        <img src="/local/image.jpg" alt="Local Image">
    </body></html>`

    converter := NewContentConverter(FormatMarkdown)
    md, err := converter.convertToMarkdown(html)
    if err != nil {
        t.Fatalf("convertToMarkdown failed: %v", err)
    }

    // Check for markdown image syntax: ![alt](url)
    if !strings.Contains(md, "![Example Image]") {
        t.Errorf("expected markdown image syntax in output, got:\n%s", md)
    }
    if !strings.Contains(md, "example.com/image.png") {
        t.Errorf("expected image URL in output, got:\n%s", md)
    }
}
```

#### Test 2.1.4: `TestConvertToMarkdown_Blockquotes`
**Test Cases** (2 tests):
1. Simple blockquote
2. Nested blockquotes

```go
func TestConvertToMarkdown_Blockquotes(t *testing.T) {
    html := `<html><body>
        <blockquote>
            <p>This is a quote.</p>
            <cite>Author Name</cite>
        </blockquote>
    </body></html>`

    converter := NewContentConverter(FormatMarkdown)
    md, err := converter.convertToMarkdown(html)
    if err != nil {
        t.Fatalf("convertToMarkdown failed: %v", err)
    }

    // Check for blockquote syntax (> )
    if !strings.Contains(md, ">") {
        t.Errorf("expected blockquote marker (>) in output, got:\n%s", md)
    }
    if !strings.Contains(md, "This is a quote") {
        t.Errorf("expected quote content in output, got:\n%s", md)
    }
}
```

#### Test 2.1.5: `TestConvertToMarkdown_DefinitionLists`
**Test Cases** (2 tests):
1. Simple definition list
2. Multiple definitions per term

```go
func TestConvertToMarkdown_DefinitionLists(t *testing.T) {
    html := `<html><body>
        <dl>
            <dt>Term 1</dt>
            <dd>Definition 1</dd>
            <dt>Term 2</dt>
            <dd>Definition 2a</dd>
            <dd>Definition 2b</dd>
        </dl>
    </body></html>`

    converter := NewContentConverter(FormatMarkdown)
    md, err := converter.convertToMarkdown(html)
    if err != nil {
        t.Fatalf("convertToMarkdown failed: %v", err)
    }

    // Verify terms and definitions are preserved
    if !strings.Contains(md, "Term 1") {
        t.Errorf("expected 'Term 1' in output, got:\n%s", md)
    }
    if !strings.Contains(md, "Definition 1") {
        t.Errorf("expected 'Definition 1' in output, got:\n%s", md)
    }
}
```

#### Test 2.1.6: `TestConvertToMarkdown_MalformedHTML`
**Test Cases** (3 tests):
1. Unclosed tags
2. Mismatched tags
3. Tags in wrong order

```go
func TestConvertToMarkdown_MalformedHTML(t *testing.T) {
    tests := []struct {
        name string
        html string
    }{
        {
            name: "unclosed tags",
            html: "<html><body><p>Unclosed paragraph<p>Another paragraph</body></html>",
        },
        {
            name: "mismatched tags",
            html: "<html><body><strong>Bold<em>BoldItalic</strong></em></body></html>",
        },
        {
            name: "wrong order",
            html: "<html><body><ul><li>Item</ul></li></body></html>",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            converter := NewContentConverter(FormatMarkdown)
            md, err := converter.convertToMarkdown(tt.html)

            // Should not panic or error, just handle gracefully
            if err != nil {
                t.Logf("Warning: malformed HTML caused error: %v", err)
            }

            // Should still produce some output
            if len(strings.TrimSpace(md)) == 0 && err == nil {
                t.Error("expected some output even for malformed HTML")
            }
        })
    }
}
```

#### Test 2.1.7: `TestConvertToMarkdown_SpecialCharacters`
**Test Cases** (3 tests):
1. HTML entities (`&nbsp;`, `&copy;`, `&trade;`)
2. Unicode characters
3. Special symbols

```go
func TestConvertToMarkdown_SpecialCharacters(t *testing.T) {
    html := `<html><body>
        <p>Copyright &copy; 2025</p>
        <p>Trademark&trade; Symbol</p>
        <p>Non-breaking&nbsp;space</p>
        <p>Unicode: 你好 مرحبا</p>
        <p>Symbols: © ® ™ € £ ¥</p>
    </body></html>`

    converter := NewContentConverter(FormatMarkdown)
    md, err := converter.convertToMarkdown(html)
    if err != nil {
        t.Fatalf("convertToMarkdown failed: %v", err)
    }

    // Verify special characters are handled
    if !strings.Contains(md, "Copyright") {
        t.Errorf("expected 'Copyright' in output, got:\n%s", md)
    }
    // Note: HTML entities may be converted to actual symbols
    // Just verify content is preserved in some form
}
```

#### Test 2.1.8: `TestExtractPlainText_HiddenElements`
**Test Cases** (1 test):
```go
func TestExtractPlainText_HiddenElements(t *testing.T) {
    html := `<html><head>
        <style>body { color: red; }</style>
    </head><body>
        <p>Visible content</p>
        <noscript>No JavaScript content</noscript>
        <script>console.log("hidden");</script>
    </body></html>`

    converter := NewContentConverter(FormatText)
    text := converter.extractPlainText(html)

    // Should contain visible content
    if !strings.Contains(text, "Visible content") {
        t.Errorf("expected 'Visible content', got:\n%s", text)
    }

    // Should NOT contain style, script, or noscript content
    if strings.Contains(text, "color: red") {
        t.Error("should not contain CSS")
    }
    if strings.Contains(text, "console.log") {
        t.Error("should not contain JavaScript")
    }
}
```

---

### Phase 3, Task 3.1: Enhance handlers_test.go

**Required Tests** (3 tests):

#### Test 3.1.1: `TestDisplayTabList_LargeLists`
```go
func TestDisplayTabList_LargeLists(t *testing.T) {
    // Create 100 tabs
    tabs := make([]TabInfo, 100)
    for i := 0; i < 100; i++ {
        tabs[i] = TabInfo{
            Index: i + 1,
            URL:   fmt.Sprintf("https://example.com/page%d", i),
            Title: fmt.Sprintf("Page %d", i),
        }
    }

    var buf strings.Builder
    displayTabList(tabs, &buf, false)
    output := buf.String()

    // Verify header shows correct count
    if !strings.Contains(output, "100 tabs") {
        t.Error("should show correct tab count")
    }

    // Verify first and last tab are included
    if !strings.Contains(output, "[1]") {
        t.Error("should show first tab")
    }
    if !strings.Contains(output, "[100]") {
        t.Error("should show last tab")
    }
}
```

#### Test 3.1.2: `TestFormatTabLine_Unicode`
```go
func TestFormatTabLine_Unicode(t *testing.T) {
    tests := []struct {
        name  string
        title string
        url   string
    }{
        {
            name:  "emoji in title",
            title: "🚀 Rocket Launch 🌟",
            url:   "https://example.com",
        },
        {
            name:  "chinese characters",
            title: "中文标题",
            url:   "https://example.cn/页面",
        },
        {
            name:  "arabic text",
            title: "عنوان عربي",
            url:   "https://example.sa",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := formatTabLine(1, tt.title, tt.url, 120, false)

            // Should not crash with Unicode
            if len(result) == 0 {
                t.Error("should produce output for Unicode input")
            }

            // Should contain title (at least partially)
            // Note: May be truncated, but should have some content
        })
    }
}
```

#### Test 3.1.3: `TestStripURLParams_EdgeCases`
```go
func TestStripURLParams_EdgeCases(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "encoded characters",
            input:    "https://example.com/path?q=%20space%20",
            expected: "https://example.com/path",
        },
        {
            name:     "multiple question marks",
            input:    "https://example.com/path?query=value?extra",
            expected: "https://example.com/path",
        },
        {
            name:     "multiple hashes",
            input:    "https://example.com/path#section#subsection",
            expected: "https://example.com/path",
        },
        {
            name:     "very long query string",
            input:    "https://example.com/path?" + strings.Repeat("a=b&", 100),
            expected: "https://example.com/path",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := stripURLParams(tt.input)
            if result != tt.expected {
                t.Errorf("stripURLParams() = %q, expected %q", result, tt.expected)
            }
        })
    }
}
```

---

### Phase 3, Task 3.2: Enhance validate_test.go

**Required Tests** (4 tests):

#### Test 3.2.1: `TestValidateURL_IDNHomograph`
```go
func TestValidateURL_IDNHomograph(t *testing.T) {
    // IDN (Internationalized Domain Names) homograph attack tests
    tests := []struct {
        name        string
        url         string
        shouldAllow bool
    }{
        {
            name:        "normal domain",
            url:         "https://example.com",
            shouldAllow: true,
        },
        {
            name:        "punycode domain",
            url:         "https://xn--e1afmkfd.xn--p1ai", // пример.рф
            shouldAllow: true,
        },
        {
            name:        "mixed script (potential homograph)",
            url:         "https://раypal.com", // Note: contains Cyrillic 'а'
            shouldAllow: true, // validateURL doesn't block these
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := validateURL(tt.url)
            hasError := err != nil

            if tt.shouldAllow && hasError {
                t.Errorf("should allow URL %q, got error: %v", tt.url, err)
            }
            if !tt.shouldAllow && !hasError {
                t.Errorf("should reject URL %q", tt.url)
            }
        })
    }
}
```

#### Test 3.2.2: `TestValidateURL_ExtremelyLong`
```go
func TestValidateURL_ExtremelyLong(t *testing.T) {
    tests := []struct {
        name      string
        urlLength int
        shouldErr bool
    }{
        {"normal length", 100, false},
        {"2000 chars", 2000, false},
        {"10000 chars", 10000, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Create URL of specified length
            longPath := strings.Repeat("a", tt.urlLength-20)
            url := "https://example.com/" + longPath

            _, err := validateURL(url)

            if tt.shouldErr && err == nil {
                t.Errorf("expected error for %d char URL", tt.urlLength)
            }
            if !tt.shouldErr && err != nil {
                t.Errorf("unexpected error for %d char URL: %v", tt.urlLength, err)
            }
        })
    }
}
```

#### Test 3.2.3: `TestValidateUserAgent_ExtremeLength`
```go
func TestValidateUserAgent_ExtremeLength(t *testing.T) {
    tests := []struct {
        name   string
        length int
    }{
        {"normal length", 100},
        {"very long", 1000},
        {"extremely long", 10000},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            longUA := strings.Repeat("A", tt.length)
            result := validateUserAgent(longUA, true)

            // Should not panic, should return trimmed string
            if len(result) == 0 {
                t.Error("should return non-empty result")
            }
        })
    }
}
```

#### Test 3.2.4: `TestValidateWaitFor_Injection`
```go
func TestValidateWaitFor_Injection(t *testing.T) {
    // Test that CSS selectors don't allow script injection
    tests := []struct {
        name     string
        selector string
        expected string
    }{
        {
            name:     "normal selector",
            selector: ".content",
            expected: ".content",
        },
        {
            name:     "selector with quotes",
            selector: "div[data-test='value']",
            expected: "div[data-test='value']",
        },
        // Note: validateWaitFor just trims, doesn't sanitize
        // These tests document current behavior
        {
            name:     "selector with angle brackets",
            selector: "<script>alert()</script>",
            expected: "<script>alert()</script>", // Not sanitized
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := validateWaitFor(tt.selector, true)
            if result != tt.expected {
                t.Errorf("validateWaitFor() = %q, expected %q", result, tt.expected)
            }
        })
    }
}
```

---

### Phase 3, Task 3.3: Enhance output_test.go

**Required Tests** (3 tests):

#### Test 3.3.1: `TestSlugifyTitle_UnicodeNormalization`
```go
func TestSlugifyTitle_UnicodeNormalization(t *testing.T) {
    tests := []struct {
        name     string
        title    string
        expected string
    }{
        {
            name:     "emoji",
            title:    "Hello 🚀 World",
            expected: "hello-world",
        },
        {
            name:     "chinese characters",
            title:    "中文标题 English",
            expected: "english",
        },
        {
            name:     "arabic text",
            title:    "عربي Arabic Text",
            expected: "arabic-text",
        },
        {
            name:     "mixed unicode",
            title:    "Café ☕ 2025",
            expected: "caf-2025",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := SlugifyTitle(tt.title, 80)
            if result != tt.expected {
                t.Errorf("SlugifyTitle() = %q, expected %q", result, tt.expected)
            }
        })
    }
}
```

#### Test 3.3.2: `TestGenerateFilename_InvalidChars`
```go
func TestGenerateFilename_InvalidChars(t *testing.T) {
    // Test filesystem-restricted characters
    tests := []struct {
        name  string
        title string
    }{
        {"less than", "File<Name"},
        {"greater than", "File>Name"},
        {"colon", "File:Name"},
        {"quote", "File\"Name"},
        {"pipe", "File|Name"},
        {"question mark", "File?Name"},
        {"asterisk", "File*Name"},
    }

    timestamp := time.Date(2025, 10, 21, 14, 30, 45, 0, time.UTC)

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := GenerateFilename(tt.title, FormatMarkdown, timestamp, "https://example.com")

            // Result should not contain filesystem-restricted chars
            restricted := []string{"<", ">", ":", "\"", "|", "?", "*"}
            for _, char := range restricted {
                if strings.Contains(result, char) {
                    t.Errorf("filename %q should not contain restricted char %q", result, char)
                }
            }
        })
    }
}
```

#### Test 3.3.3: `TestResolveConflict_HighCount`
```go
func TestResolveConflict_HighCount(t *testing.T) {
    tmpDir := t.TempDir()

    // Create test.md and test-1.md through test-50.md
    baseFile := filepath.Join(tmpDir, "test.md")
    os.WriteFile(baseFile, []byte("test"), 0644)

    for i := 1; i <= 50; i++ {
        conflictFile := filepath.Join(tmpDir, fmt.Sprintf("test-%d.md", i))
        os.WriteFile(conflictFile, []byte("test"), 0644)
    }

    // Should return test-51.md
    filename, err := ResolveConflict(tmpDir, "test.md")
    if err != nil {
        t.Fatalf("ResolveConflict failed: %v", err)
    }

    expected := "test-51.md"
    if filename != expected {
        t.Errorf("expected %q, got %q", expected, filename)
    }
}
```

---

## Test Templates and Examples

### Standard Test File Header

All new test files should start with this header:

```go
// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
    "errors"
    "strings"
    "testing"
)

// Add init() if logger needs to be initialized for tests
func init() {
    // Initialize global logger for tests (discard output)
    logger = &Logger{
        level:  LevelQuiet,
        color:  false,
        writer: io.Discard,
    }
}
```

### Table-Driven Test Template

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "descriptive test case name",
            input:    "input value",
            expected: "expected output",
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := functionName(tt.input)
            if result != tt.expected {
                t.Errorf("functionName(%q) = %q, expected %q", tt.input, result, tt.expected)
            }
        })
    }
}
```

### Helper Function Template

```go
// assertContains checks if output contains expected substring
func assertContains(t *testing.T, output, expected string) {
    t.Helper()
    if !strings.Contains(output, expected) {
        t.Errorf("expected output to contain %q, got:\n%s", expected, output)
    }
}
```

---

## Success Criteria

### Phase Completion Criteria

**Phase 1 Complete When**:
- ✅ `fetch_test.go` created with ~18 tests
- ✅ `errors_test.go` created with ~7 tests
- ✅ All tests pass: `go test ./... -v`
- ✅ Code formatted: `gofmt -w *_test.go`

**Phase 2 Complete When**:
- ✅ `formats_test.go` has 16 new tests
- ✅ All edge cases covered (nested lists, tables, images, blockquotes, etc.)
- ✅ All tests pass

**Phase 3 Complete When**:
- ✅ `handlers_test.go` has 3 new tests
- ✅ `validate_test.go` has 4 new tests
- ✅ `output_test.go` has 3 new tests
- ✅ All tests pass

**Phase 4 Complete When**:
- ✅ `go test ./...` passes with 0 failures
- ✅ `go test -cover ./...` shows improved coverage
- ✅ All test files properly formatted
- ✅ No production `.go` files modified

### Overall Project Success

✅ **45-50 new tests added**
✅ **2 new test files created**
✅ **0 production code changes**
✅ **All tests passing**
✅ **Code follows existing patterns and style**

---

## Testing Guidelines

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests with coverage
go test -cover ./...

# Run specific test file
go test -v -run TestHasLoginIndicators

# Run tests in specific package
go test -v ./

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Writing Best Practices

1. **Clear Test Names**
   - Use format: `TestFunctionName_Scenario`
   - Example: `TestValidateURL_EmptyString`

2. **Table-Driven Tests**
   - Use for multiple similar test cases
   - Include `name` field for clear subtest identification

3. **Helper Functions**
   - Mark with `t.Helper()` for better error reporting
   - Reuse across test files when appropriate

4. **Error Messages**
   - Include actual and expected values
   - Provide context for debugging

5. **Test Independence**
   - Each test should run independently
   - Don't rely on test execution order
   - Clean up resources in `t.Cleanup()`

6. **Edge Cases**
   - Always test: empty strings, nil values, zero values
   - Test boundary conditions
   - Test unexpected input

---

## Project Timeline

| Phase | Tasks | Estimated Time | Priority |
|-------|-------|----------------|----------|
| Phase 1 | Create fetch_test.go, errors_test.go | 3-4 hours | High |
| Phase 2 | Enhance formats_test.go | 2-3 hours | Medium |
| Phase 3 | Enhance handlers/validate/output tests | 2-3 hours | Medium |
| Phase 4 | Verification and cleanup | 1 hour | Low |
| **Total** | | **8-12 hours** | |

---

## Notes

- All code must follow Mozilla Public License 2.0
- Maintain existing code style and conventions
- Use existing test patterns as reference
- Focus on pure function testing only
- Document any assumptions or limitations
- If a test cannot be written without code changes, document it and skip

---

## References

- Existing test files: `*_test.go`
- Go testing documentation: https://pkg.go.dev/testing
- Table-driven tests: https://go.dev/wiki/TableDrivenTests
- Test coverage: https://go.dev/blog/cover

---

**End of Project Documentation**
