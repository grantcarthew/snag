# snag - Code Review Issues

**Review Date**: 2025-10-17
**Reviewer**: Claude Code (Comprehensive Go Code Review)
**Total Issues**: 31
**Progress**: 3 completed, 2 skipped/deferred, 26 remaining

## Status Legend

- ✅ **Completed** - Issue resolved
- 🔧 **In Progress** - Currently being worked on
- ⏳ **Pending** - Not yet started
- ⏭️ **Deferred** - Postponed to post-v1.0
- ⏭️ **Not Doing** - Skipped by design decision

---

## 🔴 Critical Issues (5)

### 1. Duplicate BrowserOptions Passing ✅ **COMPLETED**

**Status**: ✅ Fixed (2025-10-17)

**Location**: main.go:181-194, browser.go:45

**Problem**:
Options were passed to both `NewBrowserManager()` constructor AND `Connect()` method, creating redundancy and confusion. The stored fields in BrowserManager were never actually used.

```go
// Before (redundant):
bm := NewBrowserManager(BrowserOptions{...})
_, err := bm.Connect(BrowserOptions{...})  // Same options again!
```

**Impact**: API design confusion, duplicate code, potential for inconsistent options.

**Resolution**:
- Added `forceHeadless`, `forceVisible`, `openBrowser` fields to BrowserManager struct
- Updated constructor to store all options
- Removed `opts` parameter from `Connect()` method
- Updated `Connect()` to use stored struct fields
- Updated main.go to only pass options once

```go
// After (clean):
bm := NewBrowserManager(BrowserOptions{...})
_, err := bm.Connect()  // Uses stored options
```

**Files Modified**:
- browser.go: Lines 19-28, 40-47, 51-65
- main.go: Line 189

---

### 2. Unused Variable in Auth Detection ✅ **COMPLETED**

**Status**: ✅ Fixed (2025-10-17)

**Location**: fetch.go:106, 159

**Problem**:
```go
// Line 106
info, err := proto.NetworkGetResponseBody{}.Call(pf.page)
if err != nil {
    logger.Debug("Could not get response body for auth detection: %v", err)
}
// ... never used ...
_ = info // Line 159: Suppress unused variable warning
```

**Why It's Bad**:
1. Variable is fetched but never used
2. Comment suggests incomplete implementation
3. Auth detection doesn't actually use this data
4. Dead code that serves no purpose

**Current Auth Detection Methods**:
- JavaScript eval for HTTP status codes (401, 403)
- Password input detection
- Login form pattern matching

**Resolution**:
- Removed unused `proto.NetworkGetResponseBody{}` call (lines 106-110)
- Removed unused variable suppression (line 159)
- Removed unused `proto` import from fetch.go
- Code compiles and auth detection remains fully functional

**Files Modified**:
- fetch.go: Lines 9-15 (removed import), 103-105 (removed dead code), 151-153 (removed suppression)

---

### 3. Lost Error Messages in main() ✅ **COMPLETED**

**Status**: ✅ Fixed (2025-10-17)

**Location**: main.go:112-114

**Problem**:
```go
if err := app.Run(os.Args); err != nil {
    os.Exit(1)  // Error is lost! Never logged.
}
```

**Why It's Bad**:
- Errors from `app.Run()` are silently swallowed
- User sees exit code 1 but no error message
- Makes debugging impossible
- Violates principle of clear error communication

**Impact**: User confusion, poor debugging experience.

**Resolution**:
- Added error message output to stderr before exit
- Users now see clear error messages when app fails
- Exit code 1 retained for proper shell integration

**Fix Applied**:
```go
if err := app.Run(os.Args); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```

**Files Modified**:
- main.go: Line 113 (added error output)

**Note**: Uses `fmt.Fprintf` instead of logger because logger is not initialized until `run()` is called.

---

### 4. Global Mutable Logger ⏭️ **NOT DOING**

**Status**: ⏭️ Skipped - Standard practice for CLI tools

**Location**: main.go:20

**Problem**:
```go
var logger *Logger  // Package-level global variable
```

**Why It's Bad**:
1. Makes testing difficult (can't inject mock logger)
2. Potential for race conditions if snag becomes a library
3. Violates dependency injection principles
4. Hidden dependency - functions use logger without declaring it

**Impact**:
- Hard to test individual functions
- Coupling between all components
- Can't use different loggers in different contexts

**Fix Options**:
1. Pass logger as parameter to functions that need it
2. Store logger in a context struct
3. Use dependency injection pattern

**Example Fix**:
```go
// Option 1: Pass as parameter
func snag(config *Config, logger *Logger) error {
    bm := NewBrowserManager(opts, logger)
    // ...
}

// Option 2: Context struct
type Snag struct {
    logger  *Logger
    browser *BrowserManager
}
```

**Complexity**: HIGH - Would require refactoring all files.

**Recommendation**: Defer to post-v1.0 (low priority for CLI tool).

---

### 5. No Signal Handling ⏭️ **DEFERRED**

**Status**: ⏭️ Deferred to post-v1.0 (documented in SIGINT.md)

**Location**: All files (missing feature)

**Problem**:
No handling for SIGINT (Ctrl+C) or SIGTERM signals.

**Why It's Bad**:
- User presses Ctrl+C → program exits immediately
- Browser process may be left running
- Page may not be closed
- Cleanup defers do NOT execute on signals

**Impact**: Orphaned browser processes, resource leaks.

**Recommendation**:
Use global `browserManager` variable (matching logger pattern) with signal handler. Full implementation plan documented in **SIGINT.md**.

**Complexity**: MEDIUM

**Priority**: HIGH - Should fix before stable release

**Reference**: See SIGINT.md for complete analysis and implementation options

---

## ⚠️ Important Issues (5)

### 6. No URL Validation ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: main.go:148

**Problem**:
```go
url := c.Args().First()  // No validation!
```

**Why It's Bad**:
- Accepts any string as URL
- No check for valid scheme (http://, https://)
- No check for malformed URLs
- May cause confusing errors later

**Examples of Bad Input**:
- `snag example.com` (no scheme)
- `snag not-a-url`
- `snag http://`

**Fix**:
```go
import "net/url"

urlStr := c.Args().First()
parsedURL, err := url.Parse(urlStr)
if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
    logger.Error("Invalid URL: %s", urlStr)
    logger.ErrorWithSuggestion(
        "URL must start with http:// or https://",
        "snag https://example.com",
    )
    return ErrInvalidURL
}
```

**Complexity**: LOW

**Priority**: MEDIUM

---

### 7. Hard-coded Timeouts Throughout ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Locations**:
- browser.go:90 - 5 second connection timeout
- browser.go:138 - 30 second browser timeout
- fetch.go:66 - 3 second stable wait

**Problem**:
```go
// browser.go:90
browser := rod.New().
    ControlURL(controlURL).
    Timeout(5 * time.Second)  // Hard-coded!

// browser.go:138
browser := rod.New().
    ControlURL(controlURL).
    Timeout(30 * time.Second)  // Hard-coded!

// fetch.go:66
err = page.WaitStable(3 * time.Second)  // Hard-coded!
```

**Why It's Bad**:
- Not configurable by users
- May be too short for slow networks
- May be too long for fast networks
- Magic numbers scattered in code

**Fix**:
```go
// Add constants at package level or config
const (
    DefaultConnectTimeout = 5 * time.Second
    DefaultBrowserTimeout = 30 * time.Second
    DefaultStableTimeout  = 3 * time.Second
)

// Or make them configurable via CLI flags
```

**Complexity**: LOW

**Priority**: LOW (current values are reasonable defaults)

**Recommendation**: Document as constants for now, make configurable post-v1.0.

---

### 8. Fragile Error Detection ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: fetch.go:53

**Problem**:
```go
if strings.Contains(err.Error(), "timeout") ||
   strings.Contains(err.Error(), "context deadline exceeded") {
    return "", ErrPageLoadTimeout
}
```

**Why It's Bad**:
- Checking error strings is brittle
- Error messages can change between library versions
- Not type-safe
- Go convention is to use `errors.Is()` or `errors.As()`

**Fix**:
```go
import "errors"

// Check for context deadline errors properly
if errors.Is(err, context.DeadlineExceeded) {
    return "", ErrPageLoadTimeout
}

// Or use errors.As for wrapped errors
var timeoutErr interface{ Timeout() bool }
if errors.As(err, &timeoutErr) && timeoutErr.Timeout() {
    return "", ErrPageLoadTimeout
}
```

**Complexity**: LOW

**Priority**: MEDIUM

**Note**: May need to check rod's error types to implement properly.

---

### 9. File Overwrite Without Warning ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: convert.go:93

**Problem**:
```go
err := os.WriteFile(filename, []byte(content), 0644)
// Silently overwrites existing files!
```

**Why It's Bad**:
- User loses data without warning
- No confirmation prompt
- No backup created
- Violates principle of least surprise

**Impact**: User data loss.

**Fix Options**:

**Option 1**: Check and warn (recommended for CLI):
```go
if _, err := os.Stat(filename); err == nil {
    logger.Warning("File %s already exists, overwriting", filename)
}
err := os.WriteFile(filename, []byte(content), 0644)
```

**Option 2**: Add --force flag:
```go
if _, err := os.Stat(filename); err == nil && !forceOverwrite {
    return fmt.Errorf("file %s already exists (use --force to overwrite)", filename)
}
```

**Option 3**: Prompt user (not ideal for CLI piping):
```go
if _, err := os.Stat(filename); err == nil {
    fmt.Fprintf(os.Stderr, "File exists. Overwrite? [y/N]: ")
    // ... read input
}
```

**Complexity**: LOW

**Priority**: MEDIUM

**Recommendation**: Option 1 (warn) or Option 2 (--force flag).

---

### 10. Memory Concerns for Large Pages ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Locations**:
- fetch.go:92 - HTML loaded entirely into memory
- convert.go:42 - Duplicate created during conversion

**Problem**:
```go
// fetch.go:92 - Entire page in memory
html, err := page.HTML()

// convert.go:42 - Creates duplicate
content, err = cc.convertToMarkdown(html)
```

**Why It's Bad**:
- Large pages (>100MB) could cause OOM
- No streaming support
- Two copies in memory (HTML + Markdown)

**Realistic Impact**:
- Most web pages: <5MB (fine)
- Large documentation sites: 10-50MB (probably fine)
- Extreme cases: 100MB+ (could fail)

**Fix**:
Implement streaming conversion (complex):
```go
// Would require streaming API from html-to-markdown
// Or write content in chunks
```

**Complexity**: HIGH

**Priority**: LOW (edge case)

**Recommendation**: Defer to post-v1.0. Document limitation in README if needed.

---

## 📋 Best Practice Violations (5)

### 11. No Context Usage ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: All files

**Problem**:
No `context.Context` used anywhere for cancellation or timeout propagation.

**Why It's Bad**:
- Can't cancel operations mid-flight
- Timeouts are handled at individual operation level
- No way to propagate cancellation down the stack
- Not idiomatic Go for network operations

**Example**:
```go
// Current
func (pf *PageFetcher) Fetch(opts FetchOptions) (string, error)

// Better
func (pf *PageFetcher) Fetch(ctx context.Context, opts FetchOptions) (string, error)
```

**Benefits of Context**:
- Proper cancellation (works with signal handling)
- Timeout propagation
- Request-scoped values
- Better resource cleanup

**Complexity**: HIGH - Would require refactoring all function signatures.

**Priority**: LOW for CLI tool, HIGH if this becomes a library.

**Recommendation**: Defer to post-v1.0 or library extraction.

---

### 12. Inconsistent Error Wrapping ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Locations**: Multiple files

**Problem**:
```go
// browser.go:94 - Mix of %w and %v
return nil, fmt.Errorf("%w: %v", ErrBrowserConnection, err)

// browser.go:130 - Only %w
return nil, fmt.Errorf("failed to launch browser: %w", err)
```

**Why It's Bad**:
- Mixing `%w` (wrapping) and `%v` (formatting) is confusing
- `%w` should be used for errors to enable `errors.Is()` and `errors.As()`
- `%v` breaks error chain

**Fix**:
```go
// Consistent wrapping
return nil, fmt.Errorf("%w: %w", ErrBrowserConnection, err)
// Or just:
return nil, fmt.Errorf("failed to connect to browser: %w", err)
```

**Best Practice**:
- Use `%w` for wrapping errors
- Use sentinel errors at package boundary
- Allow error chain inspection

**Complexity**: LOW

**Priority**: LOW

---

### 13. No Structured Logging ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: logger.go

**Problem**:
Using fmt-based logging instead of structured logging (like `log/slog`).

**Current Approach**:
```go
logger.Verbose("Target URL: %s", url)
logger.Debug("HTTP status code: %d", status)
```

**Structured Alternative**:
```go
slog.Info("fetching page", "url", url, "timeout", timeout)
slog.Debug("page loaded", "status", status, "duration", duration)
```

**Trade-offs**:

**Current (custom logger)**:
- ✅ Simple, focused
- ✅ Human-readable output
- ✅ Color support built-in
- ❌ No machine parsing
- ❌ No JSON output

**Structured (slog)**:
- ✅ Machine-parseable
- ✅ JSON output option
- ✅ Standard library (Go 1.21+)
- ❌ More verbose
- ❌ Less human-friendly by default

**Recommendation**: Keep custom logger for v1.0 (better UX), add structured option post-v1.0.

**Complexity**: MEDIUM

**Priority**: LOW

---

### 14. No Tests ⏳ **PENDING**

**Status**: ⏳ Not yet addressed (Phase 7 in PROJECT.md)

**Location**: Entire project

**Problem**:
Zero test files. No unit tests, no integration tests.

**Missing Coverage**:
- ❌ Browser connection logic
- ❌ Page fetching
- ❌ Auth detection
- ❌ Format conversion
- ❌ Error handling
- ❌ CLI flag parsing
- ❌ File output

**Should Have**:

**Unit Tests**:
```go
// browser_test.go
func TestBrowserConnection(t *testing.T)
func TestBrowserLaunch(t *testing.T)

// fetch_test.go
func TestAuthDetection(t *testing.T)
func TestPageFetch(t *testing.T)

// convert_test.go
func TestMarkdownConversion(t *testing.T)
func TestHTMLPassthrough(t *testing.go)
```

**Integration Tests**:
```go
// integration_test.go
func TestFetchRealPage(t *testing.T)
func TestAuthWorkflow(t *testing.T)
```

**Complexity**: MEDIUM

**Priority**: HIGH - Critical for v1.0

**Status**: Planned in Phase 7 (PROJECT.md:171-187)

---

### 15. Magic Numbers ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Locations**:
- convert.go:93 - File mode `0644`
- convert.go:99 - Division by `1024.0` for KB

**Problem**:
```go
err := os.WriteFile(filename, []byte(content), 0644)  // What is 0644?
sizeKB := float64(len(content)) / 1024.0  // Why 1024.0?
```

**Why It's Bad**:
- Unclear meaning
- Not self-documenting
- Hard to change consistently

**Fix**:
```go
const (
    DefaultFileMode = 0644  // Owner RW, Group R, Other R
    BytesPerKB = 1024.0
)

err := os.WriteFile(filename, []byte(content), DefaultFileMode)
sizeKB := float64(len(content)) / BytesPerKB
```

**Complexity**: TRIVIAL

**Priority**: LOW

---

## 🐛 Potential Bugs (4)

### 16. Race Condition in Logger ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: logger.go + main.go:20

**Problem**:
Global logger can be accessed concurrently without synchronization.

**Scenario**:
```go
// main.go
var logger *Logger  // Global

// If snag becomes a library and is used concurrently:
go snag.Fetch(url1)  // Sets logger
go snag.Fetch(url2)  // Overwrites logger!
```

**Current Risk**: LOW (single-threaded CLI)

**Future Risk**: HIGH (if becomes library)

**Fix**:
- Pass logger as parameter (Issue #4)
- Or use sync.Mutex if keeping global

**Complexity**: See Issue #4

**Priority**: LOW for CLI, HIGH for library

---

### 17. Browser Cleanup Race ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: main.go:207-212

**Problem**:
```go
defer func() {
    if config.CloseTab {
        logger.Verbose("Cleanup: closing tab and browser if needed")
    }
    bm.Close()  // Called even if bm.Connect() failed
}()
```

**Why It's Not Actually a Bug**:
The `Close()` method handles nil browser:
```go
// browser.go:182
func (bm *BrowserManager) Close() error {
    if bm.browser == nil {
        return nil  // Safe!
    }
    // ...
}
```

**Status**: False alarm - already handled correctly.

**Complexity**: N/A

**Priority**: N/A

**Resolution**: No fix needed. Good defensive programming already in place.

---

### 18. OpenBrowserOnly Might Not Keep Browser Open ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: browser.go:147-164, main.go:138

**Problem**:
```go
// main.go:130-138
if c.Bool("open-browser") {
    logger.Info("Opening browser...")
    bm := NewBrowserManager(...)
    return bm.OpenBrowserOnly()  // Returns immediately
}
// ... rest of run() has defers that might close browser
```

**Why It Might Be a Problem**:
- `OpenBrowserOnly()` returns nil
- Function exits
- Defers in main.go won't execute (different code path)
- Actually, this is fine!

**Analysis**:
Looking at the code flow:
1. `--open-browser` takes early return (line 138)
2. No defers executed before this point
3. Browser stays open (correct)

**Status**: False alarm - works correctly.

**Resolution**: No fix needed.

---

### 19. WaitFor Element Timeout ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: fetch.go:72-82

**Problem**:
```go
if opts.WaitFor != "" {
    logger.Verbose("Waiting for selector: %s", opts.WaitFor)
    elem, err := page.Element(opts.WaitFor)  // Uses page timeout!
    if err != nil {
        return "", fmt.Errorf("failed to find selector %s: %w", opts.WaitFor, err)
    }
    // ...
}
```

**Why It's Bad**:
- If selector never appears, user waits full page timeout (30s)
- No feedback during wait
- No separate timeout for element wait

**Fix**:
```go
// Add feedback
logger.Progress("Waiting for selector: %s (timeout: %ds)", opts.WaitFor, timeout)

// Or add separate timeout
elem, err := page.Timeout(5 * time.Second).Element(opts.WaitFor)
```

**Complexity**: LOW

**Priority**: LOW

---

## 🎯 Missing Functionality (5)

### 20. No Build-time Version Injection ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: main.go:16-18

**Problem**:
```go
const (
    version = "1.0.0"  // Hard-coded
)
```

**Why It's Bad**:
- Version must be manually updated
- No git commit hash
- No build date
- Can't tell development vs release builds

**Fix**:
Use `-ldflags` at build time:

```go
// main.go
var (
    version   = "dev"
    commit    = "unknown"
    buildDate = "unknown"
)
```

```bash
# Build command
go build -ldflags="-X main.version=1.0.0 -X main.commit=$(git rev-parse HEAD) -X main.buildDate=$(date -u +%Y-%m-%d)"
```

**Complexity**: LOW

**Priority**: MEDIUM (needed for releases)

---

### 21. No Progress Indicators ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: All files

**Problem**:
Long operations have no visual feedback:
- Browser launch: 2-5 seconds
- Page load: 5-30 seconds
- Conversion: <1 second (usually fine)

**User Experience**:
```bash
$ snag https://slow-site.com
# ... 10 seconds of silence ...
# User: "Is it working??"
```

**Fix**:
Add spinner or progress dots:
```go
logger.ProgressWithSpinner("Fetching page...")
// or
logger.ProgressDots("Loading", interval)
```

**Complexity**: MEDIUM

**Priority**: LOW (verbose mode helps)

---

### 22. No Retry Logic ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: fetch.go

**Problem**:
Network requests don't retry on transient failures:
- DNS resolution failures
- Connection refused
- Temporary network issues

**Fix**:
```go
func fetchWithRetry(maxRetries int, backoff time.Duration) error {
    for i := 0; i < maxRetries; i++ {
        err := fetch()
        if err == nil {
            return nil
        }
        if !isRetriable(err) {
            return err
        }
        time.Sleep(backoff * time.Duration(i+1))
    }
    return ErrMaxRetriesExceeded
}
```

**Complexity**: MEDIUM

**Priority**: LOW (most sites are reliable)

**Recommendation**: Defer to post-v1.0.

---

### 23. No Configuration File Support ⏭️ **DEFERRED**

**Status**: ⏭️ Intentionally deferred (Design Decision #5)

**Location**: N/A (missing feature)

**Problem**:
All options must be specified via CLI flags every time.

**Use Case**:
```bash
# User always wants:
snag --timeout 60 --verbose --user-agent "Custom/1.0" <url>

# Instead of ~/.snagrc:
# timeout = 60
# verbose = true
# user-agent = "Custom/1.0"
```

**Design Decision**:
- Not in MVP (see docs/design.md:521-531)
- Can use shell aliases as workaround
- Post-v1.0 consideration

**Complexity**: MEDIUM

**Priority**: Post-MVP

---

### 24. No Cookie/Session Management ⏭️ **DEFERRED**

**Status**: ⏭️ Post-MVP feature

**Location**: N/A (missing feature)

**Problem**:
Can't save/load cookies between runs. Auth sessions aren't persistent across invocations.

**Use Case**:
```bash
# First run: authenticate and save cookies
snag --save-cookies auth.json https://site.com

# Second run: reuse cookies
snag --load-cookies auth.json https://site.com/private
```

**Workaround**:
Use `--open-browser` to keep session in running browser.

**Complexity**: MEDIUM

**Priority**: Post-MVP

---

## 📁 File-Specific Issues (6)

### 25. main.go - Format Validation Should Use Enum ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: main.go:166-170

**Problem**:
```go
if config.Format != "markdown" && config.Format != "html" {
    logger.Error("Invalid format: %s", config.Format)
    return fmt.Errorf("invalid format: %s", config.Format)
}
```

**Why It's Bad**:
- Magic strings duplicated
- No single source of truth for valid formats
- Hard to extend (add PDF, text, etc.)

**Fix**:
```go
// At package level
const (
    FormatMarkdown = "markdown"
    FormatHTML     = "html"
)

var validFormats = map[string]bool{
    FormatMarkdown: true,
    FormatHTML:     true,
}

// In validation
if !validFormats[config.Format] {
    logger.Error("Invalid format: %s", config.Format)
    logger.ErrorWithSuggestion(
        "Format must be 'markdown' or 'html'",
        fmt.Sprintf("Valid formats: %s", strings.Join(getValidFormats(), ", ")),
    )
    return ErrInvalidFormat
}
```

**Complexity**: LOW

**Priority**: LOW

---

### 26. main.go - Config Struct Should Have Validation Method ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: main.go:250-261

**Problem**:
Validation logic scattered in `run()` function. Config struct has no self-validation.

**Current**:
```go
func run(c *cli.Context) error {
    // ... validation logic mixed with initialization
    if config.Format != "markdown" && config.Format != "html" {
        // ...
    }
    // ... more code
}
```

**Better**:
```go
type Config struct {
    // fields...
}

func (c *Config) Validate() error {
    if c.Format != "markdown" && c.Format != "html" {
        return ErrInvalidFormat
    }
    if c.Timeout < 1 {
        return ErrInvalidTimeout
    }
    // ... etc
    return nil
}

func run(c *cli.Context) error {
    config := &Config{...}
    if err := config.Validate(); err != nil {
        return err
    }
    // ...
}
```

**Benefits**:
- Separation of concerns
- Reusable validation
- Testable independently
- Cleaner `run()` function

**Complexity**: LOW

**Priority**: LOW

---

### 27. errors.go - Consider Adding Error Codes ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: errors.go

**Problem**:
Errors have no programmatic codes for handling.

**Current**:
```go
var ErrBrowserNotFound = errors.New("no Chromium-based browser found")
```

**Enhancement**:
```go
type ErrorCode int

const (
    ErrCodeBrowserNotFound ErrorCode = 1001
    ErrCodePageTimeout     ErrorCode = 1002
    ErrCodeAuthRequired    ErrorCode = 1003
    // ...
)

type SnagError struct {
    Code    ErrorCode
    Message string
    Err     error
}

func (e *SnagError) Error() string {
    return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
}
```

**Benefits**:
- Programmatic error handling
- Error codes in logs
- Easier debugging
- Better for automation

**Trade-offs**:
- More complex
- Might be overkill for CLI

**Recommendation**: Defer to post-v1.0 or library use.

**Complexity**: MEDIUM

**Priority**: LOW

---

### 28. logger.go - Should Be Interface ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: logger.go:38-43

**Problem**:
Logger is concrete struct, not interface.

**Current**:
```go
type Logger struct {
    level  LogLevel
    color  bool
    writer io.Writer
}
```

**Better**:
```go
type Logger interface {
    Success(format string, args ...interface{})
    Info(format string, args ...interface{})
    Verbose(format string, args ...interface{})
    Debug(format string, args ...interface{})
    Warning(format string, args ...interface{})
    Error(format string, args ...interface{})
    ErrorWithSuggestion(errMsg string, suggestion string)
    Progress(format string, args ...interface{})
}

type ConsoleLogger struct {
    level  LogLevel
    color  bool
    writer io.Writer
}

func (c *ConsoleLogger) Success(...) { ... }
// ... implement interface
```

**Benefits**:
- Testability (mock logger)
- Multiple implementations
- Standard Go practice

**Complexity**: LOW

**Priority**: MEDIUM

---

### 29. browser.go - Close() Should Not Log And Return Error ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: browser.go:181-203

**Problem**:
```go
func (bm *BrowserManager) Close() error {
    // ...
    if err := bm.browser.Close(); err != nil {
        logger.Warning("Failed to close browser: %v", err)  // Logs
        return err  // And returns error
    }
    // ...
}
```

**Why It's Bad**:
Go convention: Either log and handle, OR return error for caller to handle. Don't do both.

**Current**: Logs warning AND returns error (caller may log again = duplicate).

**Fix Options**:

**Option 1**: Don't return error (recommended for cleanup):
```go
func (bm *BrowserManager) Close() {
    if bm.browser == nil {
        return
    }
    if err := bm.browser.Close(); err != nil {
        logger.Warning("Failed to close browser: %v", err)
    }
    // ... cleanup launcher
}
```

**Option 2**: Return error, don't log:
```go
func (bm *BrowserManager) Close() error {
    if bm.browser == nil {
        return nil
    }
    if err := bm.browser.Close(); err != nil {
        return fmt.Errorf("failed to close browser: %w", err)
    }
    // ...
    return nil
}
```

**Recommendation**: Option 1 (cleanup should be best-effort).

**Complexity**: LOW

**Priority**: LOW

---

### 30. browser.go - disable-blink-features May Not Work ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: browser.go:114

**Problem**:
```go
l := launcher.New().
    Bin(path).
    Headless(headless).
    Set("disable-blink-features", "AutomationControlled")  // May not work everywhere
```

**Why It Might Be a Problem**:
- Flag may not be supported on all Chromium variants
- Brave/Edge might handle it differently
- Could fail silently

**Current Risk**: LOW (common flag, widely supported)

**Recommendation**:
- Test with Edge/Brave
- Add error handling if unsupported
- Document which browsers are tested

**Complexity**: LOW

**Priority**: LOW

---

## 🔒 Security Concerns (3)

### 31. JavaScript Evaluation ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: fetch.go:113

**Problem**:
```go
statusCode, err := pf.page.Eval(`() => {
    return window.performance?.getEntriesByType?.('navigation')?.[0]?.responseStatus || 0;
}`)
```

**Current Status**: SAFE (hardcoded JavaScript)

**Future Risk**:
If JavaScript ever becomes dynamic or user-provided → XSS risk.

**Recommendation**:
- Keep JS hardcoded (never accept user input)
- Add comment warning about security
- If needed in future, sanitize all inputs

**Complexity**: N/A (preventive)

**Priority**: LOW (document only)

---

### 32. Path Traversal in File Output ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: convert.go:89

**Problem**:
```go
func (cc *ContentConverter) writeToFile(content string, filename string) error {
    logger.Verbose("Writing to file: %s", filename)
    err := os.WriteFile(filename, []byte(content), 0644)
    // No validation of filename!
}
```

**Security Risk**:
```bash
# User could specify:
snag https://example.com --output ../../../../etc/cron.d/malicious

# Or:
snag https://example.com --output ~/.ssh/authorized_keys
```

**Impact**:
- File overwrite anywhere user has permissions
- Could overwrite system files
- Could overwrite SSH keys

**Fix**:
```go
import "path/filepath"

func validateOutputPath(filename string) error {
    // Clean path
    clean := filepath.Clean(filename)

    // Resolve to absolute path
    abs, err := filepath.Abs(clean)
    if err != nil {
        return fmt.Errorf("invalid output path: %w", err)
    }

    // Check if trying to escape current directory (optional)
    cwd, _ := os.Getwd()
    if !strings.HasPrefix(abs, cwd) {
        logger.Warning("Output file is outside current directory: %s", abs)
    }

    // Check directory is writable
    dir := filepath.Dir(abs)
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        return fmt.Errorf("output directory does not exist: %s", dir)
    }

    return nil
}
```

**Complexity**: MEDIUM

**Priority**: MEDIUM

**Note**: User can already write anywhere they have permissions via shell (`cat > file`), so this is not critical, but good practice to validate.

---

### 33. Browser Binary Execution ⏳ **PENDING**

**Status**: ⏳ Not yet addressed

**Location**: browser.go:100-106

**Problem**:
```go
path, exists := launcher.LookPath()
if !exists {
    return nil, ErrBrowserNotFound
}
// Executes found browser without verification
```

**Security Risk**:
- Browser binary could be replaced by malicious version
- No hash verification
- No signature check

**Realistic Risk**: VERY LOW
- Requires attacker to replace system Chrome binary
- If they can do that, they already own the system
- Not practical to verify (signatures vary by OS/vendor)

**Recommendation**:
- Document that snag trusts system browser
- Rely on OS/package manager for binary integrity
- Not worth implementing verification

**Complexity**: N/A

**Priority**: VERY LOW (document only)

---

## Summary by Priority

### Critical (Fix Before v1.0)

1. ✅ Duplicate BrowserOptions Passing - **FIXED**
2. ✅ Unused Variable in Auth Detection - **FIXED**
3. ✅ Lost Error Messages in main() - **FIXED**
4. ⏭️ Global Mutable Logger - **SKIPPED** (standard practice for CLI)
5. ⏭️ No Signal Handling - **DEFERRED** (see SIGINT.md)
14. No Tests (Phase 7 planned)

### High Priority (Should Fix Soon)

6. No URL Validation
7. File Overwrite Without Warning
8. Fragile Error Detection
9. No Build-time Version Injection

### Medium Priority (Consider for v1.0)

10. Hard-coded Timeouts
11. Logger Should Be Interface
12. Path Traversal in File Output

### Low Priority (Post-v1.0)

13-24. Various improvements and enhancements

### Deferred (Design Decisions)

25. Global Logger (defer to library phase)
26. No Context Usage (defer to library phase)
27. Config File Support (intentional for MVP)
28. Cookie Management (post-MVP feature)

---

## ✅ Things Done Well

1. Clean code structure - well organized
2. Good error messages with suggestions
3. Sentinel errors - proper error handling pattern
4. Clean separation of concerns
5. License headers on all files
6. Proper resource cleanup with defer
7. Verbose logging for debugging
8. User-friendly CLI design

---

**Document Version**: 1.0
**Last Updated**: 2025-10-17
