# snag - Testing Implementation Plan

## Overview

This document outlines the testing strategy and implementation plan for the `snag` CLI tool. The testing ensures all features work correctly, error handling is robust, and the tool behaves predictably across different scenarios.

**Status**: Not Started (0%)
**Last Updated**: 2025-10-18

## About snag

`snag` is a CLI tool for intelligently fetching web page content using a browser engine, with smart session management and format conversion capabilities. Built in Go, it provides a single binary solution for retrieving web content as Markdown or HTML, with seamless authentication support through Chrome/Chromium browsers.

## Technology Stack

- **Language**: Go 1.23+
- **CLI Framework**: github.com/urfave/cli/v2
- **Browser Control**: github.com/go-rod/rod
- **HTML Conversion**: github.com/JohannesKaufmann/html-to-markdown/v2
- **Testing**: Go's built-in testing package (no external test frameworks)
- **Assertions**: Standard library only (no testify or other assertion libraries)

## Project Structure

```
snag/
├── main.go              # CLI framework & orchestration
├── browser.go           # Browser management
├── fetch.go             # Page fetching & auth detection
├── convert.go           # HTML→Markdown conversion
├── logger.go            # Custom 4-level logger
├── errors.go            # Sentinel errors
├── validate.go          # Input validation
│
├── validate_test.go     # Unit tests: URL/format validation (TO BE CREATED)
├── convert_test.go      # Unit tests: HTML→Markdown conversion (TO BE CREATED)
├── logger_test.go       # Unit tests: Logger output (TO BE CREATED)
├── cli_test.go          # Black-box integration tests (~550-650 lines) (TO BE CREATED)
│
├── testdata/            # Test fixtures (TO BE CREATED)
│   ├── simple.html      # Basic HTML page
│   ├── complex.html     # Rich HTML with tables, lists, code
│   ├── minimal.html     # Bare minimum HTML
│   ├── login-form.html  # Auth detection test
│   └── dynamic.html     # Dynamic content test
│
├── go.mod
├── go.sum
├── README.md
├── AGENTS.md
├── PROJECT.md           # This file
└── docs/
    ├── design.md
    └── notes.md
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

## Testing Philosophy & Design Decisions

### Core Principles

1. **Go-Only Testing**: Use only Go's built-in `testing` package, no external frameworks (no BATS, no testify)
2. **Black-Box Approach**: Test the compiled `./snag` binary via `exec.Command()` for integration tests
3. **Flat Structure**: Test files (`*_test.go`) live alongside source files in root directory (Go convention)
4. **No Mocking**: No mocks or complex test doubles - test real browser behavior
5. **Single Test Command**: Always run `go test ./...` - tests auto-skip if browser unavailable
6. **Pragmatic Coverage**: Unit test pure functions only, integration test everything else

### Why Black-Box Testing?

- Tests actual user experience
- Simpler to write and maintain
- No need to expose internal functions
- Catches integration issues
- Works naturally with flat project structure

### Why No Mocking?

- Browser/CDP code (`browser.go`, `fetch.go`) is tightly coupled to rod library
- Mocking rod would be complex, brittle, and doesn't test real behavior
- GitHub Actions has Chrome pre-installed, so integration tests work in CI/CD
- Better to test with real browser and skip gracefully if unavailable

## Testing Strategy

### Two-Tier Approach

**Tier 1: Unit Tests** - Fast, no dependencies
- Test pure functions that don't require browser
- `validate_test.go`: URL validation, format validation
- `convert_test.go`: HTML to Markdown conversion
- `logger_test.go`: Logger output verification

**Tier 2: Black-Box Integration Tests** - Requires browser
- Single file: `cli_test.go` (~550-650 lines)
- Tests compiled `./snag` binary end-to-end
- Includes fast tests (--version, --help, validation) and browser tests
- Auto-skips browser tests if Chrome/Chromium not available

### Test Environment

- Local test HTTP server for controlled testing (httptest.Server)
- HTML fixtures in `testdata/` directory
- Real browser via rod/CDP (headless mode)
- Real-world testing with example.com and httpbin.org

## Implementation Plan

The testing implementation follows a four-phase approach, building from simple to complex.

### Phase 1: Test Infrastructure & Fixtures

**Goal**: Set up testing foundation

**Tasks**:

1. **Create `testdata/` directory**
   ```bash
   mkdir testdata
   ```

2. **Create HTML test fixtures** in `testdata/`:
   - `simple.html`: Basic HTML with headings, paragraphs, links
   - `complex.html`: Rich HTML with tables, lists, code blocks, nested structures
   - `minimal.html`: Bare minimum `<html><body>Hello</body></html>`
   - `login-form.html`: Page with password input field (auth detection)
   - `dynamic.html`: Page with delayed content appearance

3. **Build the binary for testing**
   ```bash
   go build -o snag
   ```

4. **Create test helpers** (in `cli_test.go`):
   - `isBrowserAvailable()`: Check if Chrome/Chromium exists
   - `startTestServer()`: Launch HTTP server serving testdata/
   - `runSnag(args...)`: Execute `./snag` with arguments
   - `assertContains()`, `assertExitCode()`: Simple assertion helpers

**Deliverables**:
- `testdata/` directory with 5 HTML fixtures
- Compiled `./snag` binary
- Basic test helper functions

### Phase 2: Unit Tests (Pure Functions)

**Goal**: Test functions that don't require browser or external dependencies

**Tasks**:

1. **Create `validate_test.go`**:
   - `TestValidateURL_Valid`: Test valid URLs pass validation
   - `TestValidateURL_Invalid`: Test invalid URLs are rejected
   - `TestValidateURL_MissingScheme`: Test URLs without http/https
   - `TestValidateFormat_Valid`: Test "markdown" and "html" formats
   - `TestValidateFormat_Invalid`: Test unknown format rejection

2. **Create `convert_test.go`**:
   - `TestConvertToMarkdown_Headings`: Verify `<h1>` → `#` conversion
   - `TestConvertToMarkdown_Links`: Verify `<a>` → `[text](url)` conversion
   - `TestConvertToMarkdown_Tables`: Verify table conversion
   - `TestConvertToMarkdown_Lists`: Verify ordered/unordered lists
   - `TestConvertToMarkdown_CodeBlocks`: Verify `<pre><code>` conversion
   - `TestConvertToMarkdown_Minimal`: Test with minimal HTML

3. **Create `logger_test.go`**:
   - `TestLogger_Success`: Verify success message format
   - `TestLogger_Error`: Verify error message format
   - `TestLogger_Info`: Verify info message format
   - `TestLogger_Debug`: Verify debug message format
   - `TestLogger_QuietMode`: Verify quiet mode suppresses non-errors
   - `TestLogger_VerboseMode`: Verify verbose mode shows detail
   - `TestLogger_StderrOnly`: Verify all logs go to stderr, not stdout

**Estimated Tests**: ~15-18 unit tests

**Deliverables**:
- `validate_test.go` with 5 tests
- `convert_test.go` with 6 tests
- `logger_test.go` with 7 tests
- All tests passing with `go test`

### Phase 3: Black-Box Tests - Fast CLI Tests (No Browser)

**Goal**: Test CLI flags and validation without requiring browser

**Tasks**:

Create `cli_test.go` with fast tests (use `./snag` binary):

1. **CLI Information Tests**:
   - `TestCLI_Version`: Test `--version` flag shows version
   - `TestCLI_Help`: Test `--help` flag shows usage
   - `TestCLI_NoArguments`: Test running without URL shows error

2. **Input Validation Tests**:
   - `TestCLI_InvalidURL`: Test invalid URL returns error and exit code 1
   - `TestCLI_InvalidFormat`: Test `--format invalid` shows error
   - `TestCLI_InvalidTimeout`: Test negative timeout shows error
   - `TestCLI_InvalidPort`: Test invalid port range shows error

3. **Flag Validation Tests**:
   - `TestCLI_ConflictingFlags`: Test conflicting flags (e.g., `--quiet --verbose`)
   - `TestCLI_OutputFilePermission`: Test output to unwritable location
   - `TestCLI_FormatOptions`: Test valid format values accepted

**Estimated Tests**: ~10 fast tests

**Deliverables**:
- `cli_test.go` started with ~10 fast tests (no browser needed)
- All tests passing with `go test`

### Phase 4: Black-Box Tests - Browser Integration Tests

**Goal**: Test full workflows with real browser (all tests in `cli_test.go`)

**Tasks**:

Add browser-based tests to `cli_test.go` (with browser availability check):

1. **Core Fetch Tests**:
   - `TestCLI_FetchSimple`: Fetch simple.html from test server, verify Markdown output
   - `TestCLI_FetchComplex`: Fetch complex.html, verify tables/lists converted
   - `TestCLI_FetchMinimal`: Fetch minimal.html edge case
   - `TestCLI_HTMLFormat`: Test `--format html` outputs raw HTML
   - `TestCLI_OutputFile`: Test `-o` flag writes to file, stdout empty

2. **Browser Mode Tests**:
   - `TestCLI_ForceHeadless`: Test `--force-headless` flag
   - `TestCLI_ForceVisible`: Test `--force-visible` flag (skip in headless CI)
   - `TestCLI_OpenBrowser`: Test `--open-browser` opens without fetching
   - `TestCLI_CustomPort`: Test `--port` flag with custom port
   - `TestCLI_ExistingBrowser`: Connect to existing Chrome instance

3. **Authentication Detection Tests**:
   - `TestCLI_Auth401`: Test HTTP 401 detection (httpbin.org/status/401)
   - `TestCLI_Auth403`: Test HTTP 403 detection (httpbin.org/status/403)
   - `TestCLI_LoginForm`: Test DOM-based auth detection (login-form.html)
   - `TestCLI_NoFalsePositives`: Verify simple.html doesn't trigger auth

4. **Timeout & Wait Tests**:
   - `TestCLI_TimeoutFlag`: Test `--timeout` with short timeout
   - `TestCLI_WaitForSelector`: Test `--wait-for` with dynamic.html
   - `TestCLI_WaitForTimeout`: Test `--wait-for` with non-existent selector
   - `TestCLI_DefaultTimeout`: Verify default 30s timeout

5. **Advanced Flag Tests**:
   - `TestCLI_UserAgent`: Test `--user-agent` custom user agent
   - `TestCLI_CloseTab`: Test `--close-tab` flag
   - `TestCLI_VerboseOutput`: Test `--verbose` stderr logging
   - `TestCLI_QuietMode`: Test `--quiet` minimal output
   - `TestCLI_DebugMode`: Test `--debug` detailed logging

6. **Real-World Tests** (require internet):
   - `TestCLI_ExampleDotCom`: Fetch example.com
   - `TestCLI_HTTPBin`: Test httpbin.org endpoints
   - `TestCLI_HTTPBinDelay`: Test delayed response handling

**Estimated Tests**: ~20-25 browser integration tests

**Browser Availability Pattern**:
```go
func TestCLI_FetchSimple(t *testing.T) {
    if !isBrowserAvailable() {
        t.Skip("Chrome/Chromium not available")
    }
    // Test implementation
}
```

**Deliverables**:
- `cli_test.go` complete with ~30-35 total tests (fast + browser)
- All tests passing with `go test`
- Tests gracefully skip if no browser found

## Test Organization

### Test File Structure

**Unit Test Files** (3 files):
- `validate_test.go` (~100-150 lines): URL and format validation
- `convert_test.go` (~150-200 lines): HTML to Markdown conversion
- `logger_test.go` (~150-200 lines): Logger output verification

**Integration Test File** (1 file):
- `cli_test.go` (~550-650 lines): All black-box CLI tests
  - Fast tests (no browser): ~10 tests
  - Browser tests: ~20-25 tests
  - Helper functions
  - Test server setup

**Total Expected**: ~4 test files, ~1000-1200 lines of test code

### Testing Guidelines

**Best Practices**:

1. **Test Isolation**: Each test should be independent and idempotent
2. **Cleanup**: Use `defer` for cleanup (browsers, files, servers)
3. **Clear Names**: Use `TestCLI_FeatureName` pattern for readability
4. **Browser Check**: Always check browser availability before browser tests
5. **Standard Library Only**: No external assertion libraries
6. **Build Binary First**: Run `go build -o snag` before integration tests

**Test Structure Pattern**:

```go
// Unit test pattern
func TestValidateURL_Valid(t *testing.T) {
    err := validateURL("https://example.com")
    if err != nil {
        t.Errorf("expected valid URL to pass, got error: %v", err)
    }
}

// Black-box CLI test pattern (fast)
func TestCLI_Version(t *testing.T) {
    cmd := exec.Command("./snag", "--version")
    output, err := cmd.CombinedOutput()

    if err != nil {
        t.Fatalf("command failed: %v", err)
    }
    if !strings.Contains(string(output), "snag version") {
        t.Errorf("expected version in output, got: %s", output)
    }
}

// Black-box CLI test pattern (browser)
func TestCLI_FetchSimple(t *testing.T) {
    if !isBrowserAvailable() {
        t.Skip("Chrome/Chromium not available")
    }

    server := startTestServer(t)
    defer server.Close()

    cmd := exec.Command("./snag", server.URL+"/simple.html")
    output, err := cmd.Output()

    if err != nil {
        t.Fatalf("fetch failed: %v", err)
    }
    if !strings.Contains(string(output), "# Example") {
        t.Errorf("expected heading in markdown output")
    }
}
```

### Running Tests

**Standard Commands**:

```bash
# Run all tests (unit + integration, auto-skips if no browser)
go test

# Run with verbose output
go test -v

# Run specific test
go test -v -run TestCLI_Version

# Run with coverage
go test -cover

# Generate coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Run with race detector
go test -race
```

**Before Running Tests**:

```bash
# Build the binary (required for black-box tests)
go build -o snag

# Verify Chrome/Chromium available (optional)
which google-chrome || which chromium
```

## CI/CD Considerations

### GitHub Actions

Chrome is pre-installed on GitHub-hosted runners (ubuntu-latest, macos-latest), so all tests including browser tests will run in CI/CD.

**Example Workflow**:

```yaml
name: Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Build binary
        run: go build -o snag

      - name: Run tests
        run: go test -v -cover ./...

      - name: Generate coverage
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
```

**Browser Availability**:
- ✅ **Ubuntu**: Chrome pre-installed
- ✅ **macOS**: Chrome pre-installed
- ✅ **Local Dev**: Tests skip if no browser
- ✅ **Docker**: Can install Chrome in image

## Dependencies for Testing

**No additional dependencies required**:
- Use Go's built-in `testing` package
- Use `net/http/httptest` for test server
- Use `os/exec` for CLI testing
- Use standard library for assertions

## Success Criteria

### Must Have
- ✅ **4 test files created**: validate_test.go, convert_test.go, logger_test.go, cli_test.go
- ✅ **5 HTML fixtures created** in testdata/ directory
- ✅ **All tests passing** consistently with `go test`
- ✅ **Tests auto-skip** gracefully when browser unavailable
- ✅ **~30-35 total tests**: ~15-18 unit, ~10 fast CLI, ~20-25 browser integration
- ✅ **Coverage target met**: 60%+ overall, 70%+ for pure functions (validate, convert)

### Should Have
- ✅ **All CLI flags tested** individually
- ✅ **Error scenarios covered** with helpful messages verified
- ✅ **Browser modes tested**: headless, visible, existing session, custom port
- ✅ **Auth detection verified**: HTTP 401/403, login form patterns
- ✅ **Timeout handling tested**: custom timeout, wait-for, defaults
- ✅ **Real-world tests**: example.com, httpbin.org endpoints

### Nice to Have
- ⭕ **Flag combinations tested**: Multiple flags working together
- ⭕ **Race detector clean**: No race conditions found
- ⭕ **Coverage report generated**: HTML coverage visualization
- ⭕ **Test documentation**: Comments explaining complex tests

## Known Challenges

### Expected Challenges

1. **Browser Availability**
   - Tests require Chrome/Chromium installed
   - Solution: Auto-skip with clear message if not found

2. **Timing Issues**
   - Browser/network operations can be non-deterministic
   - Solution: Use reasonable timeouts, retry flaky tests if needed

3. **Platform Differences**
   - Browser paths differ on macOS/Linux
   - Solution: rod's `launcher.LookPath()` handles this automatically

4. **CI/CD Environment**
   - Need headless browser in CI
   - Solution: GitHub Actions has Chrome pre-installed, tests work out of the box

5. **Test Server Cleanup**
   - Ensure test servers and browsers cleaned up
   - Solution: Use `defer` consistently for all resources

### Mitigation Strategies

- **Browser detection**: Check availability before tests, skip gracefully
- **Test isolation**: Each test cleans up its own resources
- **Generous timeouts**: Use longer timeouts in tests to avoid flakiness
- **Headless mode**: Default to headless for CI/CD compatibility
- **Port management**: Use dynamic ports for test servers to avoid conflicts

## Related Documents

- **AGENTS.md**: Comprehensive project documentation and conventions
- **README.md**: User-facing documentation and usage examples
- **docs/design.md**: Design decisions and technical rationale
- **docs/notes.md**: Development notes and implementation details
- **PROJECT.md**: This file - testing implementation plan
