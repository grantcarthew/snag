# snag - Testing & Validation Project

## Overview

This document outlines the testing strategy and implementation plan for the `snag` CLI tool. The testing phase ensures all features work correctly, error handling is robust, and the tool behaves predictably across different scenarios.

**Status**: Phase 7 - Not Started (0%)
**Last Updated**: 2025-10-17

## About snag

`snag` is a CLI tool for intelligently fetching web page content using a browser engine, with smart session management and format conversion capabilities. Built in Go, it provides a single binary solution for retrieving web content as Markdown or HTML, with seamless authentication support through Chrome/Chromium browsers.

## Technology Stack

- **Language**: Go 1.21+
- **CLI Framework**: github.com/urfave/cli/v2
- **Browser Control**: github.com/go-rod/rod
- **HTML Conversion**: github.com/JohannesKaufmann/html-to-markdown/v2
- **Testing**: Go's built-in testing package

## Project Structure

```
snag/
├── main.go          # CLI framework & orchestration (180 lines)
├── browser.go       # Browser management (206 lines)
├── fetch.go         # Page fetching & auth detection (165 lines)
├── convert.go       # HTML→Markdown conversion (107 lines)
├── logger.go        # Custom 4-level logger (143 lines)
├── errors.go        # Sentinel errors (33 lines)
├── testdata/        # Test fixtures (TO BE CREATED)
├── *_test.go        # Test files (TO BE CREATED)
├── go.mod
└── go.sum
```

## Current Implementation Status

### ✅ Working Features (Tested Manually)

- **Core Functionality**: URL fetching, browser management, format conversion, file output
- **CLI Features**: All 16 flags working (--format, -o, --quiet, --verbose, --debug, etc.)
- **Logging System**: 4 levels, color output, emoji indicators, stdout/stderr separation
- **Browser Compatibility**: Chrome/Chromium with headless mode

### ⏳ Implemented But Untested

- Existing session detection
- Visible browser mode
- Authentication detection (401/403, login forms)
- Timeout handling edge cases
- User-agent customization
- Wait-for selector logic
- Edge/Brave browser support

## Testing Strategy

### Test Categories

1. **Unit Tests**: Test individual functions in isolation
2. **Integration Tests**: Test component interactions
3. **End-to-End Tests**: Test full CLI workflows
4. **Error Handling Tests**: Validate error scenarios and messages

### Test Environment

- Local test HTTP server for controlled testing
- HTML fixtures in `testdata/` directory
- Mock scenarios for auth, timeouts, and errors
- Real-world testing with example.com and other public sites

## Implementation Plan

### Task 1: Test Infrastructure Setup

**Description**: Create the foundation for testing

**Subtasks**:
- Create `testdata/` directory
- Create test HTTP server helper (serves local HTML files)
- Set up test fixtures structure
- Create helper functions for common test operations

**Deliverables**:
- `testdata/` directory with organized structure
- `test_server.go` or similar helper file
- Reusable test utilities

### Task 2: HTML Test Fixtures

**Description**: Create HTML test files for various scenarios

**Test Fixtures to Create**:

1. **simple.html**: Basic HTML page with standard elements
   - Title, headings, paragraphs, links, images
   - Use for basic fetch and conversion tests

2. **auth-401.html**: Page that returns HTTP 401
   - Test auth detection via HTTP status code

3. **auth-403.html**: Page that returns HTTP 403
   - Test auth detection via HTTP status code

4. **login-form.html**: Page with login form
   - Password input field
   - Login/signin text
   - Test auth detection via DOM patterns

5. **complex.html**: Rich HTML with various elements
   - Tables, lists, code blocks, nested structures
   - Test Markdown conversion quality

6. **minimal.html**: Bare minimum HTML
   - Just `<html><body>Hello</body></html>`
   - Test edge cases

7. **slow-load.html**: Page with delayed content
   - JavaScript that loads content after delay
   - Test timeout and wait-for scenarios

8. **dynamic.html**: Page with dynamic selector
   - Element that appears after load
   - Test `--wait-for` flag

**Deliverables**:
- 8 HTML fixture files in `testdata/`
- Documentation of each fixture's purpose

### Task 3: Integration Tests - Core Functionality

**Description**: Test basic page fetching and conversion

**Test Cases**:

1. **TestSimpleFetch**: Fetch simple.html and verify output
   - Start test server
   - Fetch page with default settings
   - Verify Markdown output contains expected content
   - Check exit code is 0

2. **TestHTMLOutput**: Fetch with `--format html`
   - Fetch simple.html with HTML format
   - Verify output is raw HTML
   - Confirm no Markdown conversion occurred

3. **TestMarkdownConversion**: Verify conversion quality
   - Fetch complex.html
   - Verify headings converted to `#` syntax
   - Verify links converted to `[text](url)` syntax
   - Verify tables converted correctly

4. **TestFileOutput**: Test `-o` flag
   - Fetch page with output file specified
   - Verify file created at correct path
   - Verify file contents match expected output
   - Verify stdout is empty (content went to file)

5. **TestMinimalPage**: Edge case with bare HTML
   - Fetch minimal.html
   - Verify basic conversion works
   - Check no errors with simple structure

**Deliverables**:
- `fetch_test.go` or `integration_test.go` with 5 passing tests
- Test coverage for fetch.go and convert.go

### Task 4: Integration Tests - Authentication Detection

**Description**: Test authentication detection logic

**Test Cases**:

1. **TestAuth401Detection**: HTTP 401 status
   - Configure test server to return 401
   - Verify auth detection triggers
   - Check error message suggests visible browser

2. **TestAuth403Detection**: HTTP 403 status
   - Configure test server to return 403
   - Verify auth detection triggers
   - Check appropriate error message

3. **TestLoginFormDetection**: DOM-based auth detection
   - Fetch login-form.html
   - Verify password input triggers auth detection
   - Confirm "login" text pattern recognition

4. **TestNoAuthDetection**: Verify no false positives
   - Fetch simple.html
   - Confirm no auth detection
   - Verify normal flow completes

**Deliverables**:
- `auth_test.go` with 4 passing tests
- Test coverage for auth detection in fetch.go

### Task 5: Integration Tests - Timeout Handling

**Description**: Test timeout and wait scenarios

**Test Cases**:

1. **TestPageLoadTimeout**: Test `--timeout` flag
   - Configure slow server response (> timeout)
   - Set short timeout (e.g., 2 seconds)
   - Verify timeout error occurs
   - Check error message is clear

2. **TestWaitForSelector**: Test `--wait-for` flag
   - Fetch dynamic.html with selector
   - Verify waits for element to appear
   - Confirm successful completion

3. **TestWaitForTimeout**: Selector never appears
   - Fetch page with `--wait-for` for non-existent selector
   - Verify timeout error
   - Check error mentions selector

4. **TestDefaultTimeout**: Verify default timeout works
   - Fetch slow-load.html without explicit timeout
   - Confirm uses default (30 seconds)
   - Verify completes or times out appropriately

**Deliverables**:
- `timeout_test.go` with 4 passing tests
- Test coverage for timeout logic in fetch.go

### Task 6: Integration Tests - Browser Connection Modes

**Description**: Test browser management logic

**Test Cases**:

1. **TestLaunchHeadless**: Default headless mode
   - Ensure no existing browser
   - Run snag without special flags
   - Verify headless browser launches
   - Confirm browser closes after completion

2. **TestForceHeadless**: `--force-headless` flag
   - Run with `--force-headless`
   - Verify headless mode used
   - Check no visible window appears

3. **TestForceVisible**: `--force-visible` flag
   - Run with `--force-visible`
   - Verify visible browser launches
   - Note: May need headless environment handling

4. **TestExistingBrowser**: Connect to running instance
   - Start Chrome with remote debugging
   - Run snag to connect to existing instance
   - Verify connection successful
   - Confirm browser NOT closed after

5. **TestBrowserNotFound**: No Chrome/Chromium installed
   - Mock launcher.LookPath() to return not found
   - Verify appropriate error message
   - Check error suggests installation

6. **TestCustomPort**: `--port` flag
   - Launch with custom port (e.g., 9333)
   - Verify browser uses correct port
   - Test connection on that port

**Deliverables**:
- `browser_test.go` with 6 passing tests
- Test coverage for browser.go

### Task 7: Integration Tests - CLI Flag Combinations

**Description**: Test various flag combinations

**Test Cases**:

1. **TestVerboseLogging**: `--verbose` flag
   - Run with --verbose
   - Verify detailed logs to stderr
   - Confirm stdout still has content only

2. **TestQuietMode**: `--quiet` flag
   - Run with --quiet
   - Verify no stderr output (except errors)
   - Confirm stdout has content

3. **TestDebugMode**: `--debug` flag
   - Run with --debug
   - Verify extensive debug logs
   - Check includes browser messages

4. **TestUserAgent**: `--user-agent` flag
   - Set custom user agent
   - Verify server receives custom UA
   - Check request headers in test

5. **TestCloseTab**: `--close-tab` flag
   - Run with --close-tab
   - Verify tab closes after fetch
   - Confirm works with existing browser

6. **TestOpenBrowser**: `--open-browser` flag
   - Run with --open-browser and URL
   - Verify browser opens
   - Confirm no content fetch occurs

7. **TestCombinations**: Multiple flags together
   - Test: `--verbose --format html -o output.html`
   - Test: `--quiet --timeout 10`
   - Test: `--force-headless --user-agent "CustomBot"`
   - Verify flags work together correctly

**Deliverables**:
- `cli_test.go` with 7+ passing tests
- Test coverage for main.go CLI logic

### Task 8: Error Handling and Edge Cases

**Description**: Validate error messages and edge cases

**Test Cases**:

1. **TestInvalidURL**: Malformed URL
   - Pass invalid URL (e.g., "not-a-url")
   - Verify clear error message
   - Check exit code is 1

2. **TestMissingURL**: No URL argument
   - Run snag without URL
   - Verify error message
   - Check suggests usage

3. **TestInvalidFormat**: Unknown format value
   - Use `--format invalid`
   - Verify validation error
   - Check lists valid formats

4. **TestOutputFileError**: Can't write output file
   - Specify output to read-only location
   - Verify file write error
   - Check error message is helpful

5. **TestNetworkError**: Unreachable host
   - Fetch `http://localhost:99999` (no server)
   - Verify network error handling
   - Check error message suggests checking URL

6. **TestConflictingFlags**: Incompatible flags
   - Test: `--force-headless --force-visible`
   - Test: `--quiet --verbose`
   - Verify validation errors

7. **TestErrorSuggestions**: Verify helpful error messages
   - Check all errors include suggestions
   - Verify error format is consistent
   - Confirm actionable guidance provided

**Deliverables**:
- `errors_test.go` with 7 passing tests
- Improved error messages if gaps found

### Task 9: Real-World Testing

**Description**: Test against actual websites

**Test Sites**:

1. **example.com**: Basic static site
2. **httpbin.org/html**: HTML endpoint
3. **httpbin.org/delay/5**: Timeout testing
4. **httpbin.org/status/401**: Auth detection
5. **httpbin.org/status/403**: Auth detection

**Test Cases**:

1. **TestRealWebsite**: Fetch example.com
   - Verify successful fetch
   - Check Markdown output quality
   - Confirm all features work

2. **TestHTTPBin**: Various httpbin endpoints
   - Test different response codes
   - Test delayed responses
   - Verify auth detection on 401/403

**Deliverables**:
- `realworld_test.go` with tests against live sites
- Note: These tests require internet connectivity

### Task 10: Test Documentation and Coverage

**Description**: Document tests and measure coverage

**Subtasks**:

1. **Run Test Coverage**:
   ```bash
   go test -cover ./...
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out -o coverage.html
   ```

2. **Document Test Cases**:
   - Add comments to all test functions
   - Explain what each test validates
   - Document any test setup requirements

3. **Create Test README** (optional):
   - Document how to run tests
   - Explain test fixtures
   - List any prerequisites (Chrome installed, etc.)

4. **Identify Coverage Gaps**:
   - Review coverage report
   - Identify untested code paths
   - Add tests for gaps if critical

**Target Coverage**: 70%+ for core packages (browser.go, fetch.go, convert.go)

**Deliverables**:
- Test coverage report
- Well-documented test files
- Optional: `TESTING.md` guide

## Testing Guidelines

### Best Practices

1. **Test Isolation**: Each test should be independent
2. **Cleanup**: Always cleanup resources (browsers, files, servers)
3. **Clear Names**: Use descriptive test function names
4. **Table-Driven Tests**: Use for similar test cases with different inputs
5. **Error Checking**: Always check errors in tests
6. **Parallel Tests**: Use `t.Parallel()` where safe

### Test Structure Pattern

```go
func TestFeatureName(t *testing.T) {
    // Setup
    server := startTestServer(t)
    defer server.Close()

    // Execute
    result, err := functionUnderTest(args)

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }

    // Cleanup (if not using defer)
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run TestSimpleFetch

# Run with coverage
go test -cover ./...

# Run with race detector
go test -race ./...
```

## Dependencies for Testing

```go
// Likely additions to go.mod for testing
require (
    // No additional dependencies needed
    // Using Go's built-in testing package
)
```

## Success Criteria

- ✅ All test files created with comprehensive coverage
- ✅ Minimum 70% code coverage for core packages
- ✅ All tests passing consistently
- ✅ Error scenarios properly tested
- ✅ Real-world websites tested successfully
- ✅ Test documentation complete
- ✅ Browser connection modes validated
- ✅ Authentication detection verified
- ✅ Timeout handling confirmed
- ✅ All CLI flags tested individually and in combinations

## Known Challenges

1. **Browser Testing**: Requires Chrome/Chromium installed
2. **Headless Environment**: CI/CD may need special setup for browser tests
3. **Timing Issues**: Network and browser operations may be non-deterministic
4. **Platform Differences**: Browser paths differ on macOS/Linux
5. **Auth Testing**: Difficult to test full auth flows without real services

## Next Steps After Testing

Once testing is complete (Phase 7), proceed to:
- **Phase 8**: Documentation - Add troubleshooting section, create LICENSES/ directory
- **Phase 9**: Distribution - GitHub Actions, multi-platform builds, Homebrew formula

## Related Documents

- `PROJECT.md`: Main project implementation plan
- `README.md`: User-facing documentation
- `docs/design.md`: Design decisions and rationale
- `docs/notes.md`: Development notes
