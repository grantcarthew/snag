# snag - Testing Implementation

## Status

**Phase**: Phases 1-3 Complete ✅
**Progress**: 46 tests implemented and passing (Phase 4 browser integration pending)
**Last Updated**: 2025-10-18
**Test Runtime**: ~8 seconds

## Quick Reference

```bash
# Run all tests
go test -v

# Run specific test category
go test -v -run TestValidate    # Validation tests
go test -v -run TestConvert     # Conversion tests
go test -v -run TestLogger      # Logger tests
go test -v -run TestCLI         # CLI integration tests

# Coverage
go test -cover
```

## Current Test Coverage

**Total Tests**: 46 passing ✅

### Unit Tests (24 tests)
- **validate_test.go** (12 tests): URL, format, timeout, port, output path validation
- **convert_test.go** (6 tests): HTML→Markdown conversion (headings, links, tables, lists, code, minimal)
- **logger_test.go** (7 tests): Logger levels, stderr routing, mode behavior

### Integration Tests (22 tests)
- **cli_test.go** (22 tests): CLI flags, error handling, validation (no browser required)
  - Version/Help/NoArgs (3 tests)
  - URL validation (1 test with 3 subtests)
  - Format/Timeout/Port validation (3 tests with 10 subtests)
  - Format options, output permissions (2 tests with 2 subtests)

### Phase 4: Browser Integration (NOT YET IMPLEMENTED)
Planned: ~20-25 tests for browser-based functionality

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
├── cli_test.go         # 362 lines, 22 tests - CLI integration (no browser)
├── testdata/
│   ├── simple.html     # Basic HTML (headings, paragraphs, links)
│   ├── complex.html    # Tables, lists, code blocks
│   ├── minimal.html    # Edge case: bare minimum HTML
│   ├── login-form.html # Auth detection (password field)
│   └── dynamic.html    # Dynamic content with delayed element
└── snag               # Compiled binary for black-box testing
```

**Total**: 852 lines of test code, 46 tests

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

## Pending Work

### Phase 4: Browser Integration Tests (Not Started)

Planned tests requiring Chrome/Chromium (~20-25 tests):

1. **Core Fetch Tests** (5 tests)
   - Fetch simple.html from test server
   - Fetch complex.html with tables/lists
   - Minimal HTML edge case
   - HTML format output
   - Output to file

2. **Browser Mode Tests** (5 tests)
   - Force headless mode
   - Force visible mode
   - Open browser without fetching
   - Custom port
   - Connect to existing instance

3. **Authentication Detection** (4 tests)
   - HTTP 401 detection
   - HTTP 403 detection
   - Login form detection (DOM-based)
   - No false positives

4. **Timeout & Wait Tests** (4 tests)
   - Custom timeout flag
   - Wait-for selector
   - Wait-for timeout
   - Default timeout

5. **Advanced Flags** (5 tests)
   - Custom user agent
   - Close tab flag
   - Verbose output
   - Quiet mode
   - Debug mode

6. **Real-World Tests** (3 tests)
   - example.com fetch
   - httpbin.org endpoints
   - Delayed response handling

**Estimated Effort**: 4-8 hours
**Requirement**: Chrome/Chromium installed
**Auto-skip**: Yes, if browser unavailable

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

**Current Achievement**:
- ✅ 4 test files created
- ✅ 5 HTML fixtures in testdata/
- ✅ 46 tests passing consistently
- ✅ Tests auto-skip when browser unavailable
- ✅ All validation functions tested
- ✅ All pure functions tested
- ✅ CLI flag validation complete

**Remaining for Full Completion**:
- ⏳ Browser integration tests (Phase 4)
- ⏳ Coverage report generation
- ⏳ CI/CD workflow setup

## Testing Commands

```bash
# Standard workflow
go build -o snag          # Build binary for integration tests
go test -v                # Run all tests with output
go test -cover            # Show coverage percentage

# Specific test categories
go test -v -run TestValidate.*_Valid     # All validation success cases
go test -v -run TestValidate.*_Invalid   # All validation failure cases
go test -v -run TestConvertToMarkdown    # All conversion tests
go test -v -run TestLogger               # All logger tests
go test -v -run TestCLI                  # All CLI integration tests

# Coverage analysis
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Race detection
go test -race

# Specific test
go test -v -run TestCLI_InvalidFormat
```

## Key Learnings

1. **Tests Found Real Bugs**: Missing validation for format/timeout/port
2. **Black-Box Testing Works**: Integration tests caught user-facing issues
3. **Code Reviews Valuable**: Multiple reviews improved test quality significantly
4. **Pragmatism Required**: Rejected overly strict suggestions for CLI testing
5. **Library Limitations**: html-to-markdown table issue documented for future fix
6. **Standard Library Sufficient**: No external test frameworks needed
