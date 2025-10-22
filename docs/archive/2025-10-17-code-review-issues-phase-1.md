# snag - Code Review Issues (Completed - Phase 1)

**Review Date**: 2025-10-17
**Reviewer**: Claude Code (Comprehensive Go Code Review)
**Status**: Archive of completed and deferred issues from initial review
**Issues Resolved**: 7 completed, 2 deferred

## Status Legend

- ✅ **Completed** - Issue resolved
- ⏭️ **Deferred** - Postponed to post-v1.0
- ⏭️ **Not Doing** - Skipped by design decision

---

## ✅ Completed Issues (7)

### 1. Duplicate BrowserOptions Passing

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

### 2. Unused Variable in Auth Detection

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

### 3. Lost Error Messages in main()

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

### 6. Context Deadline Exceeded on Page Operations

**Status**: ✅ Fixed (2025-10-17)

**Location**: browser.go:89-104, 139-153; fetch.go:47

**Problem**:
Operations were getting "context deadline exceeded" errors during HTML extraction and browser cleanup, even though navigation succeeded:

```bash
$ snag https://gitlab.com
✓ Chrome launched in headless mode
Fetching https://gitlab.com...
⚠ Failed to close browser: context deadline exceeded
Error: failed to extract HTML: context deadline exceeded
```

**Root Cause**:
Rod's `.Timeout()` method creates a shallow clone with a timeout context that affects ALL subsequent operations:

```go
// WRONG: Timeout applies to ALL operations on this browser/page
browser := rod.New().ControlURL(url).Timeout(30 * time.Second)
browser.Connect()  // Has timeout
// Later...
page.HTML()  // STILL has timeout! Fails if Navigate + HTML > 30s
browser.Close()  // STILL has timeout! May fail
```

The timeout was being inherited by:

- All pages created from the browser
- HTML extraction, auth detection, and other operations after navigation
- Browser and page close operations

**Impact**:

- Slow-loading sites (like gitlab.com from certain networks) would fail
- Operations that succeeded would still error during cleanup
- Users saw confusing "context deadline exceeded" instead of clear timeout messages

**Resolution**:

**1. Browser timeout isolation (browser.go:95-104, 144-153)**:
Apply timeout only to `Connect()`, then cancel it for future operations:

```go
// Create browser and connect with timeout
browser := rod.New().ControlURL(controlURL).Timeout(5 * time.Second)

// Try to connect
if err := browser.Connect(); err != nil {
    return nil, fmt.Errorf("%w: %v", ErrBrowserConnection, err)
}

// Return browser WITHOUT timeout for future operations
// CancelTimeout() removes the timeout context
return browser.CancelTimeout(), nil
```

**2. Page timeout isolation (fetch.go:47)**:
Apply timeout only to `Navigate()`, use original page for other operations:

```go
// Apply timeout only to navigation - creates a clone with timeout
err := pf.page.Timeout(pf.timeout).Navigate(opts.URL)

// Use original pf.page for subsequent operations (no timeout)
html, err := pf.page.HTML()  // Won't inherit navigation timeout
```

**Benefits**:

- ✅ Timeout applies only to connection/navigation (intended behavior)
- ✅ HTML extraction, auth detection work without timeout constraints
- ✅ Browser/page close operations don't timeout
- ✅ Slow sites like gitlab.com work correctly
- ✅ Clear error messages when actual timeouts occur

**Files Modified**:

- browser.go: Lines 95-104 (connectToExisting), 144-153 (launchBrowser)
- fetch.go: Line 47 (Navigate with isolated timeout)

**Reference**: [Rod documentation on Context and Timeout](https://github.com/go-rod/go-rod.github.io/blob/main/context-and-timeout.md)

---

### 7. No URL Validation

**Status**: ✅ Fixed (2025-10-17)

**Location**: main.go:149 (original), validate.go (new)

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

- `snag example.com` (no scheme - now auto-adds https://)
- `snag "bad url with spaces"` (invalid characters)
- `snag ftp://example.com` (unsupported scheme)

**Resolution**:

- Created new `validate.go` module with `validateURL()` function
- Auto-adds `https://` if no scheme is present (user-friendly)
- Validates URL using `net/url.Parse()`
- Supports `http://`, `https://`, and `file://` schemes
- Validates host exists (except for file:// URLs)
- Provides clear error messages with examples
- Keeps main.go clean and focused

**Implementation**:

```go
// validate.go
func validateURL(urlStr string) (string, error) {
    // Add https:// if no scheme present
    if !strings.Contains(urlStr, "://") {
        urlStr = "https://" + urlStr
    }

    // Parse and validate
    parsedURL, err := url.Parse(urlStr)
    if err != nil {
        return "", ErrInvalidURL
    }

    // Check supported schemes
    validSchemes := map[string]bool{
        "http":  true,
        "https": true,
        "file":  true,
    }

    if !validSchemes[parsedURL.Scheme] {
        return "", ErrInvalidURL
    }

    // Validate host (except file://)
    if parsedURL.Scheme != "file" && parsedURL.Host == "" {
        return "", ErrInvalidURL
    }

    return urlStr, nil
}
```

**Files Modified**:

- validate.go: New file (15-58)
- main.go: Updated to call validateURL() (149-157)

**User Experience**:

```bash
# Auto-adds https://
$ snag example.com
→ Uses: https://example.com ✓

# Invalid characters
$ snag "bad url with spaces"
✗ Invalid URL: https://bad url with spaces
✗ URL parsing failed: invalid character " " in host name
  Try: snag https://example.com

# Unsupported scheme
$ snag ftp://example.com
✗ Unsupported URL scheme: ftp
✗ URL must use http://, https://, or file://
  Try: snag https://example.com
```

---

### 8. Hard-coded Timeouts Throughout

**Status**: ✅ Fixed (2025-10-17)

**Location**: browser.go:19-22 (constants), browser.go:101, browser.go:150; fetch.go:63

**Problem**:
Magic numbers scattered throughout code:

```go
// browser.go - Hard-coded timeouts
browser := rod.New().ControlURL(controlURL).Timeout(5 * time.Second)   // Existing browser
browser := rod.New().ControlURL(controlURL).Timeout(30 * time.Second)  // New browser

// fetch.go - Hard-coded wait time
err = page.WaitStable(3 * time.Second)
```

**Why It's Bad**:

- Magic numbers not self-documenting
- Inconsistent timeout values (5s vs 30s for same operation)
- Hard to change consistently across codebase

**Resolution**:

- Added package-level constants in browser.go
- Unified connection timeout to 10 seconds (both existing and new browsers)
- All timeouts now use named constants

**Implementation**:

```go
// browser.go:19-22
const (
    ConnectTimeout   = 10 * time.Second // Browser connection timeout (existing or newly launched)
    StabilizeTimeout = 3 * time.Second  // Page stabilization wait time
)

// Usage throughout codebase
browser := rod.New().ControlURL(controlURL).Timeout(ConnectTimeout)
err = pf.page.WaitStable(StabilizeTimeout)
```

**Benefits**:

- ✅ Single source of truth for timeout values
- ✅ Self-documenting code
- ✅ Easy to adjust values globally
- ✅ Consistent 10s timeout for all connection operations
- ✅ 10s accommodates slower systems while remaining responsive

**Files Modified**:

- browser.go: Lines 19-22 (added constants), 101, 150 (use constants)
- fetch.go: Line 63 (use StabilizeTimeout constant)

**Note**: Timeouts are not yet CLI-configurable. Can add `--connect-timeout` and `--stabilize-timeout` flags post-v1.0 if needed.

---

### 10. File Overwrite Without Warning

**Status**: ✅ Fixed (2025-10-17)

**Location**: convert.go:93

**Problem**:

```go
err := os.WriteFile(filename, []byte(content), 0644)
// Silently overwrites existing files!
```

**Why It Was Flagged**:

- Initially appeared to violate principle of least surprise
- Concern about user data loss without warning
- No indication when overwriting files

**Analysis**:
Unix tools like `mv`, `cp`, `curl -o`, and `wget` all overwrite files silently by default. This is expected behavior.

**Resolution**:
Added verbose-mode warning when overwriting existing files, following Unix conventions:

```go
// Check if file exists and warn in verbose mode
if _, err := os.Stat(filename); err == nil {
    logger.Verbose("Overwriting existing file: %s", filename)
}
err := os.WriteFile(filename, []byte(content), 0644)
```

**Benefits**:

- ✅ Maintains Unix tool conventions (silent by default)
- ✅ Provides feedback in verbose mode for users who want it
- ✅ No breaking changes to CLI interface
- ✅ Helpful for debugging file operations

**Files Modified**:

- convert.go: Lines 92-95 (added existence check and verbose warning)

---

## ⏭️ Deferred Issues (2)

### 4. Global Mutable Logger

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

### 5. No Signal Handling

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

## Summary

### Completed (7)

1. ✅ Duplicate BrowserOptions Passing
2. ✅ Unused Variable in Auth Detection
3. ✅ Lost Error Messages in main()
4. ✅ Context Deadline Exceeded on Page Operations
5. ✅ No URL Validation
6. ✅ Hard-coded Timeouts Throughout
7. ✅ File Overwrite Without Warning

### Deferred (2)

4. ⏭️ Global Mutable Logger (standard CLI practice, post-v1.0)
5. ⏭️ No Signal Handling (post-v1.0, see SIGINT.md)

---

**Document Version**: 1.0
**Created**: 2025-10-17
**Archive Date**: 2025-10-17
