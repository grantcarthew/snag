# PROJECT: Kill Browser Feature

Add a `--kill-browser` (`-k`) flag to forcefully terminate browser processes with remote debugging enabled.

## Overview

Provide a convenient way to kill browser processes launched by snag or lingering from previous runs. Targets only browsers with `--remote-debugging-port` enabled (development/debugging browsers), not regular user browsing sessions. Useful for cleanup, troubleshooting, and scripting.

## Design Decisions

### Core Behavior

1. **Flag name**: `-k | --kill-browser` (short and long form)
2. **Scope - Without `--port`**: Kill only browsers with `--remote-debugging-port` flag enabled
   - Detect browser type using `findBrowserPath()` logic
   - Filter processes by browser name + `--remote-debugging-port` pattern
   - Kill all matching processes (development/debugging browsers only)
   - **Safety**: Won't kill regular user browsing sessions
3. **Scope - With `--port`**: Kill only the browser on that specific port
   - Connect via rod to verify it's a browser
   - Extract PID from browser connection
   - Kill parent process + all children
4. **Confirmation**: No confirmation prompt (immediate execution)
5. **Exit codes**:
   - Exit 0: Success (processes killed OR none found - idempotent)
   - Exit 1: Errors only (permission denied, browser detection failed, etc.)
6. **Child processes**: For port-specific killing, kill parent + children explicitly
7. **Error handling**: Follow existing patterns (`logger.Error()`, `logger.ErrorWithSuggestion()`)
8. **Terminology**: Use generic "browser" in output (not "Chrome", "Brave", etc.)

### Flag Compatibility

**Works with `--kill-browser`:**
- `--port` (target specific port)
- `--verbose` (show detailed process info)
- `--quiet` (minimal output)
- `--debug` (debug logging)
- `--help`, `--version` (help/version take priority)

**Errors with `--kill-browser` (conflicting operations):**
- URL arguments
- `--list-tabs`
- `--all-tabs`
- `--tab`
- `--open-browser`
- `--url-file`

**Ignored with `--kill-browser` (not applicable):**
- `--output`, `--output-dir`, `--format`, `--wait-for`, `--timeout`
- `--user-agent`, `--user-data-dir`, `--close-tab`, `--force-headless`

### Flag Priority Chain

```
--help > --version > --kill-browser > --list-tabs > --open-browser > normal operation
```

## Command-Line Interface

```bash
# Kill all browser processes with remote debugging enabled
snag --kill-browser
snag -k

# Kill only browser on specific port
snag --kill-browser --port 9223
snag -k --port 9223

# With logging levels
snag --kill-browser --verbose
snag --kill-browser --quiet
snag --kill-browser --debug
```

## Behavior Specification

### Without `--port` flag

1. Detect browser type using `findBrowserPath()` logic
2. Get browser executable name (e.g., "chrome", "brave-browser", "chromium")
3. Find processes: `ps aux | grep "$browserExe.*--remote-debugging-port"`
4. Kill matched processes with `kill -9 <PID>`
5. Report count of killed processes

**Scope**: Only kills browsers with `--remote-debugging-port` enabled (development browsers)

### With `--port` flag

1. Try to connect to browser on specified port using rod
2. If connection fails: Report "No browser running on port N" (exit 0)
3. If connection succeeds: Extract PID from browser metadata
4. Kill parent process: `kill -9 <parent_pid>`
5. Kill child processes: `pkill -9 -P <parent_pid>`
6. Report killed process PID

**Scope**: Only kills the specific browser on that port

## Output Examples

### Normal Mode

**Success (no port specified):**
```
Killing all browser processes with remote debugging...
✓ Killed 3 process(es)
```

**Success (port specified, browser running):**
```
Checking port 9222...
✓ Killed browser process (PID 12345)
```

**Success (port specified, no browser):**
```
Checking port 9222...
No browser running on port 9222
```

**No browsers found (no port specified):**
```
Killing all browser processes with remote debugging...
No browser processes found
```

### Verbose Mode

```bash
snag --kill-browser --verbose
```

```
Detecting browser...
✓ Found browser at /Applications/Google Chrome.app/Contents/MacOS/Google Chrome
Searching for browser processes with --remote-debugging-port...
  Found PID 12345: chrome --remote-debugging-port=9222 ...
  Found PID 12346: chrome --type=renderer --parent=12345 ...
  Found PID 12347: chrome --type=gpu-process --parent=12345 ...
Killing 3 process(es)...
✓ Killed 3 process(es)
```

### Debug Mode

All verbose output plus:
- Exact commands executed (`ps aux | grep ...`, `kill -9 ...`)
- Full process output and parsing details
- Each kill command result
- Error details with context

### Quiet Mode

Completely silent on success (errors still printed to stderr).

### Error Messages

**No browser found:**
```
✗ No Chromium-based browser found
  Install Chrome, Chromium, Brave, or Edge
  Example: brew install chromium
```

**Permission denied:**
```
✗ Permission denied killing browser processes
  Try with elevated privileges: sudo snag --kill-browser
```

**Port-specific, cannot connect:**
```
Checking port 9222...
No browser running on port 9222
```
(Exit 0 - not an error)

**Conflicting operation:**
```
✗ Cannot use --kill-browser with URL arguments (conflicting operations)
```

## Implementation Plan

### Phase 1: Core Implementation

**Tasks:**
1. ✅ Add `--kill-browser` (`-k`) flag definition to main.go (main.go:103, main.go:221)
2. ✅ Update help template with --kill-browser in new section (main.go:183)
3. ✅ Implement `KillBrowser(port int)` in browser.go (browser.go:616-628)
   - Uses exit-early pattern: if port > 0, kill specific port; else kill all
   - Implemented as single public method with two private helpers:
     - `killBrowserOnPort(port int)` (browser.go:630-697)
     - `killAllBrowsers()` (browser.go:699-777)
   - Uses `lsof -ti :port` to find PID for specific port
   - Uses `ps aux` to find all browsers with --remote-debugging-port
   - Returns (int, error) - count of killed processes
4. ✅ Implement `handleKillBrowser()` in handlers.go (handlers.go:974-993)
   - Checks if `--port` flag changed using `cmd.Flags().Changed("port")`
   - Routes to `bm.KillBrowser(port)` or `bm.KillBrowser(0)`
   - All output/error handling done in browser methods (minimal handler)
   - Follows existing handler pattern

**Files modified:**
- ✅ `main.go`: Flag definition, help template
- ✅ `handlers.go`: Handler function (handlers.go:974-993)
- ✅ `browser.go`: Kill methods (browser.go:616-785)

### Phase 2: Integration & Validation

**Tasks:**
5. ✅ Add flag conflict validation in main.go (main.go:398-422)
   - Validates inline in priority chain (before execution)
   - Error if `--kill-browser` + URL arguments
   - Error if `--kill-browser` + `--tab`, `--all-tabs`, `--list-tabs`, `--open-browser`, `--url-file`
   - Allows `--port`, `--verbose`, `--quiet`, `--debug`, `--help`, `--version`
   - **Bug fix**: Moved validation from validateFlagCombinations (dead code) to priority chain
6. ✅ Add `--kill-browser` to flag priority chain (main.go:396-424)
   - Priority: `help > version > kill-browser > list-tabs > open-browser > normal`
   - Validates conflicts inline before executing
   - Routes to handleKillBrowser(cmd)
7. ✅ Sentinel errors reviewed (none needed)
   - Reuses `ErrBrowserNotFound` appropriately
   - All other errors are contextual `fmt.Errorf()` with proper wrapping
   - Follows existing error handling patterns

**Files modified:**
- ✅ `main.go`: Priority chain (main.go:396-424), validation inline, removed dead code from validateFlagCombinations (main.go:266-267)
- ✅ `errors.go`: No changes needed (reuses existing sentinel errors)

### Phase 3: Testing

**Tasks:**
8. Manual testing on macOS
   - Test all scenarios (see "Manual Test Cases" below)
   - Verify exit codes
   - Test verbose/quiet/debug output
   - Test error conditions
9. Add automated tests
   - Test flag parsing and validation
   - Test conflicting flag detection
   - Integration tests (if feasible with real browser)
   - Mock tests for error conditions
10. Full regression testing
    - Run entire test suite: `go test -v ./...`
    - Verify no regressions in existing functionality
    - Check code coverage

**Files modified:**
- `kill_browser_test.go` (new file)
- `cli_test.go` (additional test cases)
- `validate_test.go` (flag validation tests)

### Phase 4: Documentation

**Tasks:**
11. Create `docs/arguments/kill-browser.md`
    - Follow existing argument documentation pattern
    - Validation rules
    - Interaction matrix
    - Examples (valid, invalid, warnings)
    - Implementation details
    - Use cases
12. Update `docs/design-record.md`
    - Add kill-browser to arguments section
    - Add design decision entry (DD-XX)
    - Document rationale and implementation approach
13. Update all argument documentation files
    - Add `--kill-browser` row to interaction matrix in ALL 20+ `docs/arguments/*.md` files:
      - `all-tabs.md`, `close-tab.md`, `debug.md`, `force-headless.md`, `format.md`
      - `help.md`, `list-tabs.md`, `open-browser.md`, `output-dir.md`, `output.md`
      - `port.md`, `quiet.md`, `tab.md`, `timeout.md`, `url-file.md`
      - `user-agent.md`, `user-data-dir.md`, `verbose.md`, `version.md`, `wait-for.md`
14. Update `README.md`
    - Add "Killing Browser Processes" section to Troubleshooting or new "Browser Management" section
    - Usage examples with warnings
    - Exit code behavior
15. Update `AGENTS.md`
    - Add to troubleshooting examples
    - Document usage in scripting scenarios

**Files modified:**
- `docs/arguments/kill-browser.md` (new)
- `docs/design-record.md`
- `docs/arguments/*.md` (20+ files)
- `README.md`
- `AGENTS.md`

## Implementation Details

### Process Detection (Without --port)

**Strategy**: Detect browser type, then filter by `--remote-debugging-port` flag

```go
// 1. Detect browser path
browserPath, err := findBrowserPath()
if err != nil {
    return 0, ErrBrowserNotFound
}

// 2. Get executable name
browserExe := filepath.Base(browserPath)

// 3. Find processes with remote debugging enabled
cmd := exec.Command("ps", "aux")
output, err := cmd.CombinedOutput()

// 4. Filter: browser name + --remote-debugging-port pattern
// grep "$browserExe.*--remote-debugging-port" | grep -v grep

// 5. Parse PIDs from output
// 6. Kill each PID: exec.Command("kill", "-9", pid)
// 7. Return count
```

### Process Killing (With --port)

**Strategy**: Connect via rod, extract PID, kill parent + children

```go
// 1. Try to connect
browser, err := rod.New().ControlURL(fmt.Sprintf("127.0.0.1:%d", port)).Connect()
if err != nil {
    // No browser on this port - not an error
    logger.Info("No browser running on port %d", port)
    return nil
}

// 2. Extract PID from browser metadata
// Try browser.GetPID() or parse from browser info

// 3. Kill parent process
exec.Command("kill", "-9", fmt.Sprintf("%d", pid)).Run()

// 4. Kill children
exec.Command("pkill", "-9", "-P", fmt.Sprintf("%d", pid)).Run()

// 5. Return PID
```

### File Structure

**main.go:**
- Flag definition: `rootCmd.Flags().BoolP("kill-browser", "k", false, "Kill browser processes with remote debugging enabled")`
- Priority chain in RunE function
- Route to handler: `return handleKillBrowser(cmd)`

**handlers.go:**
```go
func handleKillBrowser(cmd *cobra.Command) error {
    // Check if --port specified
    port, _ := cmd.Flags().GetInt("port")
    portChanged := cmd.Flags().Changed("port")

    bm := NewBrowserManager(BrowserOptions{Port: port})

    if portChanged {
        // Kill specific port
        count, err := bm.KillBrowser(port)
        // Handle output and errors
        return err
    } else {
        // Kill all with remote debugging
        count, err := bm.KillBrowser(0)
        // Handle output and errors
        return err
    }
}
```

**browser.go:**
```go
// Single public method with exit-early pattern
func (bm *BrowserManager) KillBrowser(port int) (int, error) {
    if port > 0 {
        return bm.killBrowserOnPort(port)
    }
    return bm.killAllBrowsers()
}

// Private helper: kills browser on specific port using lsof
func (bm *BrowserManager) killBrowserOnPort(port int) (int, error) {
    // Implementation at browser.go:630-697
}

// Private helper: kills all browsers with --remote-debugging-port
func (bm *BrowserManager) killAllBrowsers() (int, error) {
    // Implementation at browser.go:699-777
}
```

## Testing

### Manual Test Cases

```bash
# Test 1: Kill all (no browsers with remote debugging)
snag --kill-browser
# Expected: "No browser processes found" (exit 0)

# Test 2: Kill all (browsers running with remote debugging)
snag --open-browser
ps aux | grep chrome | grep remote-debugging-port  # Note PIDs
snag --kill-browser
ps aux | grep chrome | grep remote-debugging-port  # Verify killed
# Expected: "✓ Killed N process(es)" (exit 0)

# Test 3: Kill specific port (browser running)
snag --open-browser --port 9223
snag --kill-browser --port 9223
# Expected: "✓ Killed browser process (PID XXXXX)" (exit 0)

# Test 4: Kill specific port (not running)
snag --kill-browser --port 9999
# Expected: "No browser running on port 9999" (exit 0)

# Test 5: Verbose output
snag --open-browser
snag --kill-browser --verbose
# Expected: Detailed PID list and process info

# Test 6: Quiet output
snag --open-browser
snag --kill-browser --quiet
# Expected: Silent (or minimal success message)

# Test 7: Debug output
snag --open-browser
snag --kill-browser --debug
# Expected: All verbose + command execution details

# Test 8: Conflicting flags - URL
snag --kill-browser https://example.com
# Expected: "✗ Cannot use --kill-browser with URL arguments (conflicting operations)" (exit 1)

# Test 9: Conflicting flags - list-tabs
snag --kill-browser --list-tabs
# Expected: "✗ Cannot use --kill-browser with --list-tabs (conflicting operations)" (exit 1)

# Test 10: Conflicting flags - tab
snag --kill-browser --tab 1
# Expected: "✗ Cannot use --kill-browser with --tab (conflicting operations)" (exit 1)

# Test 11: Multiple browsers on different ports
snag --open-browser --port 9222
snag --open-browser --port 9223
snag --kill-browser --port 9222
ps aux | grep chrome | grep remote-debugging-port
# Expected: Only 9222 killed, 9223 still running

# Test 12: Permission denied (if applicable)
# Run as non-root, kill process owned by different user
# Expected: "✗ Permission denied killing browser processes" (exit 1)

# Test 13: Help priority
snag --kill-browser --help
# Expected: Show help, ignore kill-browser

# Test 14: Version priority
snag --kill-browser --version
# Expected: Show version, ignore kill-browser

# Test 15: Regular browser unaffected
# Open regular Chrome window (not via snag, no remote debugging)
snag --kill-browser
# Expected: Regular Chrome window remains open
```

### Automated Test Cases

**`kill_browser_test.go`:**
- `TestKillBrowserFlagParsing()` - flag defined and accessible
- `TestKillBrowserConflictingFlags()` - errors on URL, --tab, --list-tabs, etc.
- `TestKillBrowserWithPort()` - port flag integration
- `TestKillBrowserPriority()` - help/version take priority

**Integration tests (if feasible):**
- `TestKillAllBrowsers()` - launch browser, kill, verify
- `TestKillBrowserOnPort()` - launch on specific port, kill, verify
- `TestKillBrowserNoProcesses()` - exit 0 when nothing to kill

**Validation tests (`validate_test.go`):**
- Test flag combination validation logic
- Test conflicting operation detection

## Success Criteria

**Phase 1 - Core Implementation:**
- [x] `--kill-browser` (`-k`) flag defined in main.go (main.go:103, main.go:221)
- [x] Help template updated with --kill-browser in new section (main.go:183)
- [x] `KillBrowser(port int)` implemented in browser.go with exit-early pattern (browser.go:616-785)
  - [x] `killAllBrowsers()` private helper (filters by --remote-debugging-port using ps aux)
  - [x] `killBrowserOnPort(port int)` private helper (uses lsof to find PID, kills with kill -9)
- [x] `handleKillBrowser()` implemented in handlers.go (handlers.go:974-993)
- [x] Basic error handling and output formatting (all in browser methods, handler minimal)

**Phase 2 - Integration & Validation:**
- [x] Flag conflict validation (errors on URL, --tab, --all-tabs, --list-tabs, --open-browser, --url-file) (main.go:398-422)
- [x] Flag priority chain (help > version > kill-browser > list-tabs > open-browser > normal) (main.go:396-424)
- [x] Sentinel errors reviewed (none needed - reuses ErrBrowserNotFound, others are contextual)
- [x] **Bug fix**: Fixed unreachable dead code (external review items #6 & #7) - validation moved from validateFlagCombinations to inline priority chain

**Phase 3 - Testing:**
- [x] Manual testing completed on macOS (all 15 test cases)
- [x] Automated tests written and passing (kill_browser_test.go with 2 tests)
- [x] Full regression testing (126/126 tests pass - fixed 2 assertion mismatches)
- [x] Exit codes verified (0 for success/nothing to kill, 1 for errors)
- [x] Bug fixed: browser.Close() was killing browser before lsof could find PID

**Phase 4 - Documentation:**
- [ ] `docs/arguments/kill-browser.md` created
- [ ] `docs/design-record.md` updated with design decision
- [ ] ALL 20+ `docs/arguments/*.md` files updated with interaction rows
- [ ] `README.md` updated with usage examples and warnings
- [ ] `AGENTS.md` updated with troubleshooting examples

## Platform Support

**Supported:**
- macOS (current development platform)
- Linux (tested commands compatible)

**Process commands used:**
- `ps aux` (available on both)
- `grep` (available on both)
- `kill -9` (available on both)
- `pkill -9 -P` (available on both)

**Future consideration:**
- Windows support would require `taskkill /F /IM chrome.exe` approach
- Document as macOS/Linux only for now

## Edge Cases & Error Handling

| Scenario | Behavior | Exit Code |
|----------|----------|-----------|
| No Chromium-based browser installed | Error: "No Chromium-based browser found" | 1 |
| Browser installed, no processes with remote debugging | Info: "No browser processes found" | 0 |
| Port specified, no browser on that port | Info: "No browser running on port N" | 0 |
| Permission denied (cannot kill process) | Error: "Permission denied..." | 1 |
| Multiple browsers on different ports (kill one port) | Kill only specified port, others untouched | 0 |
| Regular Chrome window open (no remote debugging) | Unaffected, not killed | 0 |
| Conflicting flags (--kill-browser + URL) | Error: "Cannot use ... (conflicting operations)" | 1 |
| Help/version flag also present | Show help/version, ignore kill-browser | 0 |

## Future Enhancements

Not in scope for initial implementation:

1. **Confirmation prompt**: Add `--confirm` flag for interactive mode
2. **Dry run**: Add `--dry-run` to show what would be killed
3. **Kill all variants**: Add `--all` to kill Chrome, Brave, Edge, etc. simultaneously
4. **Windows support**: Implement `taskkill` approach for Windows
5. **PID tracking**: Maintain `~/.snag/browser.pid` file to track snag-launched browsers
6. **Process age filter**: Only kill processes older than N seconds
7. **Pattern matching**: Kill browsers matching port range (e.g., `--port-range 9222-9230`)

## Notes

- Feature primarily for development/troubleshooting, not production use
- Scope limited to browsers with `--remote-debugging-port` flag (safe default)
- No external dependencies added (uses standard Unix commands)
- Exit 0 for "nothing to kill" enables idempotent scripting
- Generic "browser" terminology in output (not browser-specific)
- Follows existing snag patterns for errors, logging, and validation

## External Code Review Findings

**Addressed:**
- ✅ **Items #6 & #7 (Unreachable validation code)**: Fixed - moved validation from validateFlagCombinations to inline priority chain
  - Validation now executes before handleKillBrowser()
  - Removed dead code from validateFlagCombinations

**Known Limitations (Documented):**
- **Items #1, #3 (lsof, ps aux not cross-platform)**: macOS/Linux only (documented in Platform Support section)
  - Windows support deferred to Future Enhancements

**Design Decisions (Accepted):**
- **Items #2, #5 (kill -9 immediate vs graceful SIGTERM)**: Using immediate SIGKILL by design
  - Rationale: Tool is for cleanup/troubleshooting stuck browsers, not graceful shutdown
  - User intent is clear: force-kill debug browsers
  - Adds complexity + delay for minimal benefit in this use case
- **Item #4 (ps aux parsing brittle)**: Standard practice, low risk
  - PID in fields[1] is stable across Unix-like systems
