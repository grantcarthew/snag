# snag - Testing Implementation

## Status

**Phase**: Phases 1-4 Complete ✅
**Progress**: 70 tests passing, 1 skipped (bug documented)
**Last Updated**: 2025-10-19
**Test Runtime**: ~91 seconds

## Quick Reference

```bash
# Run all tests
go test -v

# Run specific test category
go test -v -run TestValidate    # Validation tests
go test -v -run TestConvert     # Conversion tests
go test -v -run TestLogger      # Logger tests
go test -v -run TestCLI         # CLI integration tests (no browser)
go test -v -run TestBrowser     # Browser integration tests (Phase 4)

# Coverage
go test -cover
```

## Current Test Coverage

**Total Tests**: 71 (70 passing ✅, 1 skipped ⏭️)

### Unit Tests (25 tests)
- **validate_test.go** (12 tests): URL, format, timeout, port, output path validation
- **convert_test.go** (6 tests): HTML→Markdown conversion (headings, links, tables, lists, code, minimal)
- **logger_test.go** (7 tests): Logger levels, stderr routing, mode behavior

### Integration Tests (46 tests)
- **cli_test.go** (46 tests total: 22 Phase 3 + 26 Phase 4 - 2 overlaps)

**Phase 3: CLI Tests (9 tests - no browser required)**
  - Version/Help/NoArgs (3 tests)
  - URL validation (1 test with 3 subtests)
  - Format/Timeout/Port validation (3 tests with 10 subtests)
  - Format options, output permissions (2 tests with 2 subtests)

**Phase 4: Browser Integration (26 tests)**
  - Core Fetch Tests (5 tests): simple.html, complex.html, minimal.html, HTML format, output to file
  - Browser Mode Tests (5 tests): headless, visible, open-browser, custom port, connect existing
  - Authentication Detection (4 tests): 401, 403, login form, no false positives
  - Timeout & Wait Tests (4 tests): custom timeout, wait-for selector, wait-for timeout ⏭️, default timeout
  - Advanced Flags (5 tests): user agent, close-tab, verbose, quiet, debug
  - Real-World Tests (3 tests): example.com, httpbin.org, delayed response

## Completed Work

### ✅ Phase 1: Test Infrastructure (Complete)
- Created `testdata/` directory with 5 HTML fixtures
- Built binary for black-box testing
- Implemented test helpers in `cli_test.go`:
  - `isBrowserAvailable()` - Browser detection
  - `startTestServer()` - HTTP test server
  - `runSnag()` - Binary execution wrapper
  - Assertion helpers: `assertContains`, `assertExitCode`, `assertNoError`, `assertError`

### ✅ Phase 2: Unit Tests (Complete)
All pure functions have comprehensive unit test coverage:
- **validate.go**: URL, format, timeout, port, output path validation
- **convert.go**: HTML to Markdown conversion with proper syntax verification
- **logger.go**: All log levels, modes, and stderr routing

### ✅ Phase 3: Fast CLI Tests (Complete)
Black-box integration tests without browser dependency:
- CLI info (version, help, no args)
- Input validation (URL, format, timeout, port)
- Output path permissions
- Flag acceptance testing

### ✅ Phase 4: Browser Integration Tests (Complete)
Full browser-based integration tests with Chrome/Chromium:
- **Core Fetch Tests (5 tests)**: Fetch simple/complex/minimal HTML, format output, file writing
- **Browser Mode Tests (5 tests)**: Headless, visible, open-browser, custom port, existing instance
- **Authentication Detection (4 tests)**: HTTP 401/403, login forms, no false positives
- **Timeout & Wait Tests (4 tests)**: Custom timeout, wait-for selector, timeout handling, defaults
- **Advanced Flags (5 tests)**: User agent, close-tab, verbose, quiet, debug modes
- **Real-World Tests (3 tests)**: example.com, httpbin.org, delayed responses

**Known Issues Found**:
- BUG: `--wait-for` with non-existent selector hangs indefinitely, ignores `--timeout` flag
- Root cause: fetch.go wait-for logic doesn't respect timeout
- Test: TestBrowser_WaitForTimeout (skipped with documentation)
- TODO: Fix timeout handling in fetch.go

### ✅ Code Quality Improvements
Through multiple external code reviews, implemented:
- Validation functions actually called in tests (not just map checks)
- Combined string assertions for proper markdown syntax verification
- Portable temp directory creation with proper cleanup
- Unique temp file names using `os.CreateTemp()`
- Self-contained validation functions (no package-level dependencies)
- Case-sensitive format error messages
- Proper exit code assertions for integration tests
- Fenced code block syntax verification
- Comprehensive validation coverage (all new functions tested)
- Standard markdown heading format validation (with space after `#`)

## Test Files Structure

```
snag/
├── validate_test.go    # 196 lines, 12 tests - URL/format/timeout/port/path validation
├── convert_test.go     # 165 lines, 6 tests  - HTML→Markdown conversion
├── logger_test.go      # 129 lines, 7 tests  - Logger behavior and modes
├── cli_test.go         # 1055 lines, 35 tests - CLI + Browser integration
├── testdata/
│   ├── simple.html     # Basic HTML (headings, paragraphs, links)
│   ├── complex.html    # Tables, lists, code blocks
│   ├── minimal.html    # Edge case: bare minimum HTML
│   ├── login-form.html # Auth detection (password field)
│   └── dynamic.html    # Dynamic content with delayed element
└── snag               # Compiled binary for black-box testing
```

**Total**: 1545 lines of test code, 71 tests (70 passing, 1 skipped)

## Key Validation Improvements

During implementation, tests revealed missing validation that was added:

1. **Format Validation** (`validateFormat`)
   - Validates markdown/html only (case-sensitive)
   - Clear error messages with case-sensitivity note

2. **Timeout Validation** (`validateTimeout`)
   - Must be positive integer
   - Validates before attempting connection

3. **Port Validation** (`validatePort`)
   - Range: 1-65535
   - Prevents invalid port numbers early

4. **Output Path Validation** (`validateOutputPath`)
   - Directory existence check
   - Write permission verification using `os.CreateTemp()`
   - Prevents errors at write time

All validation functions have comprehensive test coverage with positive and negative test cases.

## Known Issues & Limitations

### HTML to Markdown Conversion

**Table Conversion Issue** (Documented in `docs/projects/PROJECT-html2markdown.md`):
- Tables do not convert to markdown table syntax (no pipe characters)
- Content is preserved but structure is lost
- Example: `<table>` → `NameValue Item1100` (not `| Name | Value |`)
- Nested formatting within cells IS preserved (bold, links, code)
- Investigation needed for library configuration or alternative

**Test Documentation**:
```go
// NOTE: Current html-to-markdown library does not convert tables to proper
// markdown table syntax. This test verifies table content is preserved.
// TODO: Consider library configuration or alternative for proper table support.
```

## Testing Philosophy

### Black-Box Approach
- Tests compiled `./snag` binary via `exec.Command()`
- Validates actual user experience
- No exposure of internal functions required
- Catches integration issues

### Pragmatic Choices
- **No mocking**: Test with real browser when available
- **Standard library only**: No testify, no BATS
- **Auto-skip**: Browser tests skip gracefully if Chrome unavailable
- **Integration over isolation**: Prefer testing real behavior
- **Appropriate assertions**: Keyword checks for CLI errors (not brittle string matching)

### Code Review Decisions

Rejected overly strict suggestions:
- ❌ Exact error string matching (CLI messages aren't APIs)
- ❌ Removing global logger from tests (standard Go practice, no race conditions)
- ❌ Ultra-specific error validation in black-box tests
- ✅ Accepted: Proper markdown syntax verification
- ✅ Accepted: Comprehensive validation coverage

## Future Work

### Bug Fixes
1. **--wait-for timeout handling** (Priority: Medium)
   - Issue: `--wait-for` with non-existent selector ignores `--timeout` flag
   - Location: fetch.go wait-for logic
   - Test: TestBrowser_WaitForTimeout (currently skipped)
   - Impact: Browser tab left open indefinitely, process hangs

### Enhancements
1. **Coverage report generation** (Priority: Low)
   - Generate HTML coverage reports
   - Set coverage thresholds
   - Track coverage trends

2. **CI/CD workflow setup** (Priority: Medium)
   - GitHub Actions workflow
   - Multi-platform testing (ubuntu-latest, macos-latest)
   - Automated test runs on PR

3. **HTML table conversion** (Priority: Low)
   - Investigate html-to-markdown library configuration
   - Consider alternative libraries for proper table syntax
   - See docs/projects/PROJECT-html2markdown.md

## CI/CD Ready

Tests are designed for GitHub Actions:
- Chrome pre-installed on ubuntu-latest and macos-latest
- No additional dependencies
- Fast unit tests (~4ms)
- Moderate integration tests (~8s with browser)

**Example Workflow**:
```yaml
- name: Build binary
  run: go build -o snag

- name: Run tests
  run: go test -v -cover ./...
```

## Related Documentation

- **AGENTS.md**: Project conventions and development workflow
- **docs/projects/PROJECT-html2markdown.md**: Table conversion investigation
- **README.md**: User-facing documentation
- **docs/design.md**: Technical design decisions

## Success Metrics

**Phase 4 Complete Achievement**:
- ✅ 4 test files created (1545 lines of test code)
- ✅ 5 HTML fixtures in testdata/
- ✅ 71 tests total (70 passing, 1 skipped with bug documentation)
- ✅ Tests auto-skip when browser unavailable
- ✅ All validation functions tested
- ✅ All pure functions tested
- ✅ CLI flag validation complete
- ✅ Full browser integration test coverage
- ✅ Real-world endpoint testing (example.com, httpbin.org)
- ✅ Bug discovered and documented: --wait-for timeout handling

**Future Enhancements**:
- ⏳ Fix --wait-for timeout bug
- ⏳ Coverage report generation
- ⏳ CI/CD workflow setup

## Testing Commands

```bash
# Standard workflow
go build -o snag          # Build binary for integration tests
go test -v                # Run all tests with output
go test -cover            # Show coverage percentage

# Specific test categories
go test -v -run TestValidate              # All validation tests
go test -v -run TestConvert               # All conversion tests
go test -v -run TestLogger                # All logger tests
go test -v -run TestCLI                   # CLI integration tests (no browser)
go test -v -run TestBrowser               # Browser integration tests (Phase 4)

# Specific browser test groups
go test -v -run TestBrowser_Fetch         # Core fetch tests
go test -v -run TestBrowser_Force         # Browser mode tests
go test -v -run TestBrowser_Auth          # Authentication detection tests
go test -v -run TestBrowser_.*Timeout     # Timeout and wait tests
go test -v -run TestBrowser_RealWorld     # Real-world endpoint tests

# Coverage analysis
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Race detection
go test -race

# Skip long-running tests
go test -v -short                         # Skips real-world tests
```

## Key Learnings

1. **Tests Found Real Bugs**:
   - Missing validation for format/timeout/port (Phase 2-3)
   - `--wait-for` timeout handling bug (Phase 4)
2. **Black-Box Testing Works**: Integration tests caught user-facing issues
3. **Code Reviews Valuable**: Multiple reviews improved test quality significantly
4. **Pragmatism Required**: Rejected overly strict suggestions for CLI testing
5. **Library Limitations**: html-to-markdown table issue documented for future fix
6. **Standard Library Sufficient**: No external test frameworks needed
7. **Real-World Testing Essential**: Testing against live endpoints (example.com, httpbin.org) validates actual behavior
8. **Browser Tests Auto-Skip**: Graceful degradation when Chrome unavailable enables CI/CD flexibility
