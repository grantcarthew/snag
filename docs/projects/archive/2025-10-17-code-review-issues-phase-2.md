# snag - Code Review Issues (Phase 2 - Completed)

**Review Date**: 2025-10-17
**Reviewer**: Claude Code (Comprehensive Go Code Review)
**Phase 2 Results**: 6 issues resolved (2 fixed, 4 documented/won't do)
**Previous Phase**: See docs/projects/2025-10-17-code-review-issues-phase-1.md (7 completed, 2 deferred)

## Status Legend

- ⏳ **Pending** - Not yet started
- 🔧 **In Progress** - Currently being worked on
- ⏭️ **Deferred** - Postponed to post-v1.0

---

## ⚠️ Important Issues (3)

### 2. Fragile Error Detection

**Status**: ✅ Fixed

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

### 3. No Build-time Version Injection

**Status**: ⏭️ Deferred to GitHub workflows

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

### 4. Memory Concerns for Large Pages

**Status**: ❌ Won't do - unrealistic use case

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

## 📋 Best Practice Violations (6)

### 5. No Context Usage

**Status**: ❌ Won't do - unnecessary for CLI tool

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

### 6. Inconsistent Error Wrapping

**Status**: ✅ Fixed

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

### 7. No Structured Logging

**Status**: ❌ Won't do - custom logger better for CLI UX

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

### 8. Magic Numbers

**Status**: ✅ Fixed

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

### 9. Logger Should Be Interface

**Status**: ❌ Won't do - unnecessary for main package

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

### 10. Close() Should Not Log And Return Error

**Status**: ✅ Fixed

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

## 🐛 Potential Bugs (2)

### 11. Race Condition in Logger

**Status**: ❌ Won't do - single-threaded CLI tool

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
- Pass logger as parameter (related to deferred Issue #4)
- Or use sync.Mutex if keeping global

**Complexity**: See deferred Issue #4 in docs/2025-10-17-code-review-issues-one.md

**Priority**: LOW for CLI, HIGH for library

---

### 12. WaitFor Element Timeout

**Status**: ⏭️ Deferred to UX improvements

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

## 🎯 Missing Functionality (3)

### 13. No Progress Indicators

**Status**: ⏭️ Deferred - see UX.md

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

### 14. No Retry Logic

**Status**: ❌ Won't do - not applicable for interactive CLI

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

### 15. No Cookie/Session Management

**Status**: ⏭️ Deferred - on backlog

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

## 📁 File-Specific Issues (4)

### 16. Format Validation Should Use Enum

**Status**: ✅ Fixed

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

### 17. Config Struct Should Have Validation Method

**Status**: 📄 Moved to CONFIG.md

**Location**: main.go:250-261

**Resolution**: Design decision documented in CONFIG.md for future consideration. Current single-field validation is adequate for v1.0.

---

### 18. Consider Adding Error Codes

**Status**: ❌ Won't do - overkill for CLI tool

**Location**: errors.go

**Resolution**: Numeric error codes are unnecessary for a simple CLI tool. Current sentinel errors with `errors.Is()` are sufficient.

---

### 19. disable-blink-features May Not Work

**Status**: ❌ Won't do - flag fails gracefully

**Location**: browser.go:114

**Resolution**: If unsupported by a browser variant, the flag is safely ignored. No action needed.

---

## 🔒 Security Concerns (3)

### 20. JavaScript Evaluation

**Status**: ✅ Fixed

**Location**: fetch.go:105-109

**Resolution**: Added security comment documenting that JavaScript must remain hardcoded and never accept user input to prevent XSS vulnerabilities.

---

### 21. Path Traversal in File Output

**Status**: ❌ Won't do - user has shell access

**Location**: convert.go:89

**Resolution**: Not protecting users from themselves. Users have shell access and can already write files anywhere they have permissions. Validation would be security theater.

---

### 22. Browser Binary Execution

**Status**: ✅ Fixed

**Location**: browser.go:116-118

**Resolution**: Added security comment documenting that snag trusts system-installed browser binaries. Binary integrity is the responsibility of the OS package manager.

---

## ✅ Non-Issues (Verified)

### Browser Cleanup Race

**Location**: main.go:207-212

**Status**: ✅ Not a bug - already handled correctly

**Analysis**:
The `Close()` method handles nil browser safely:
```go
// browser.go:182
func (bm *BrowserManager) Close() error {
    if bm.browser == nil {
        return nil  // Safe!
    }
    // ...
}
```

Good defensive programming already in place.

---

### OpenBrowserOnly Browser Persistence

**Location**: browser.go:147-164, main.go:138

**Status**: ✅ Not a bug - works correctly

**Analysis**:
- `--open-browser` takes early return (line 138)
- No defers executed before this point
- Browser stays open (correct behavior)

---

## Summary by Priority

### Critical (Fix Before v1.0)
1. No Tests (Phase 7 planned)

### High Priority (Should Fix Soon)
2. Fragile Error Detection
3. No Build-time Version Injection
4. Memory Concerns for Large Pages (document limitation)

### Medium Priority (Consider for v1.0)
9. Logger Should Be Interface
21. Path Traversal in File Output

### Low Priority (Post-v1.0)
5-8, 10-20, 22: Various improvements and enhancements

### Deferred (Design Decisions)
15. Cookie/Session Management (post-MVP feature)

See docs/2025-10-17-code-review-issues-one.md for completed and deferred issues.

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
9. URL validation with auto-scheme addition (user-friendly)
10. File operations follow Unix conventions (silent by default, verbose feedback available)
11. Proper timeout isolation using Rod's CancelTimeout() pattern

---

**Document Version**: 2.0
**Last Updated**: 2025-10-17
**Archive**: See docs/2025-10-17-code-review-issues-one.md for completed issues
