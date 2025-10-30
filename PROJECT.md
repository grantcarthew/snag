# PROJECT: Doctor/Diagnostics Feature

Add a `--doctor` flag to display comprehensive diagnostic information about snag's environment, browser detection, and system status.

## Overview

Provide a diagnostic command that helps users troubleshoot issues and provides maintainers with essential information when debugging user problems. Similar to `brew doctor`, `npm doctor`, etc.

## Decisions Made

1. **Flag name**: `--doctor` (preferred over `--info`, `--diagnose`, `--system-info`)
2. **Override behavior**: Priority order is `help > doctor > version > everything else`
3. **Works with**: `--port` (check custom port), `--verbose` (verbose logging), `--debug` (debug logging)
4. **Ignores**: All other flags (`--quiet`, URLs, `--list-tabs`, etc.)
5. **Output destination**: stdout (diagnostic report), stderr (verbose/debug logs)
6. **Exit code**: Always 0 (unless exception)
7. **Information to report**:
   - Version info (snag, latest from GitHub, Go, OS/Arch)
   - Working directory
   - Browser detection (name, path, version - raw output)
   - Connection status (default port + custom if specified)
   - Profile path (detected browser only)
   - Environment variables (CHROME_PATH, CHROMIUM_PATH)
8. **GitHub release check**: Included in doctor output (10 second timeout), not in `--version`
9. **Output format**: Human-readable with repo link in header, Unicode/emoji formatting
10. **Browser profile mapping**: Extend existing `browserDetectionRule` struct with OS-specific profile paths
11. **No JSON output**: Human-readable format is grep-friendly

## Requirements

### Command-Line Interface

```bash
# Run diagnostics with default settings
snag --doctor

# Run diagnostics checking specific port
snag --doctor --port 9223
```

### Information to Display

**Version Information:**

- snag version (from build ldflags or version constant)
- Go version (from `runtime.Version()`)
- OS/Architecture (from `runtime.GOOS` / `runtime.GOARCH`)

**Working Directory:**

- Current working directory (from `os.Getwd()`)

**Browser Detection:**

- Detected browser name (Chrome, Chromium, Brave, Edge, etc.)
- Browser executable path
- Browser version (from `--version` flag if available)

**Default Profile Paths:**

- List all common browser profile locations for current OS
- Indicate which ones exist vs don't exist
- Show paths for: Chrome, Chromium, Brave, Edge, Vivaldi, Arc, Opera

**Connection Status:**

- Default port 9222: Running/Not running
- If running: Number of tabs open
- Custom port (if `--port` specified): Running/Not running, tab count

**Environment Variables:**

- `CHROME_PATH` (if set)
- `CHROMIUM_PATH` (if set)
- Any other browser-related env vars

### Output Example

```
snag Doctor Report
==================
https://github.com/grantcarthew/snag

Version Information
───────────────────
  snag version:    0.0.4
  Latest version:  0.0.5 (update available)
  Go version:      go1.25.3
  OS/Arch:         darwin/arm64

Working Directory
─────────────────
  /home/user/projects/myproject

Browser Detection
─────────────────
  Detected:        Chrome
  Path:            /usr/bin/google-chrome
  Version:         131.0.6778.85

Connection Status
─────────────────
  Port 9222:       ✓ Running (7 tabs open)
  Port 9223:       Not running

Profile Location
─────────────────
  Chrome:          ✓ ~/Library/Application Support/Google/Chrome

Environment Variables
─────────────────────
  CHROME_PATH:     (not set)
  CHROMIUM_PATH:   (not set)
```

**With custom port:**

```bash
snag --doctor --port 9223

# Connection Status section shows:
Connection Status
─────────────────
  Port 9222:       ✓ Running (7 tabs open)
  Port 9223:       ✓ Running (3 tabs open)
```

**When no browser found:**

```
Browser Detection
─────────────────
  Detected:        ✗ No Chromium-based browser found
  Path:            (none)
  Version:         (none)
```

## Implementation Details

### File Changes

**main.go:**

- Add `--doctor` boolean flag definition
- Add to flag priority logic: `help > doctor > version > everything else`
- When `--doctor` is set, works with `--port`, `--verbose`, `--debug`; ignores all other flags
- Route to handler

**handlers.go:**

- Add `handleDoctor(cmd *cobra.Command) error` function
- Route to doctor.go implementation
- Return nil (always succeeds, even if some checks fail)

**browser.go:**

- Extend `browserDetectionRule` struct with `profilePathMac` and `profilePathLinux` fields
- Update `browserDetectionRules` table with profile path mappings for all browsers
- Add `func (bm *BrowserManager) GetBrowserVersion() (string, error)` - executes `<browser> --version`, returns raw output
- Add `func (bm *BrowserManager) GetProfilePath() (string, bool)` - returns profile path and whether it exists

**New file: `doctor.go`:**

- `type DoctorReport struct` - holds all diagnostic data
- `func CollectDoctorInfo(port int) (*DoctorReport, error)` - gathers all info
- `func (dr *DoctorReport) Print()` - formats and outputs to stdout
- `func checkLatestVersion() (string, error)` - queries GitHub API with 10s timeout

### Data Structures

```go
type DoctorReport struct {
    SnagVersion    string
    LatestVersion  string
    GoVersion      string
    OS             string
    Arch           string
    WorkingDir     string

    BrowserName    string
    BrowserPath    string
    BrowserVersion string
    BrowserError   error

    ProfilePath    string
    ProfileExists  bool

    DefaultPortStatus  *PortStatus
    CustomPortStatus   *PortStatus  // nil if --port not specified

    EnvVars        map[string]string
}

type PortStatus struct {
    Port      int
    Running   bool
    TabCount  int
    Error     error
}
```

### Implementation Steps

**Version Information:**

```go
// Set via ldflags at build time
var version = "dev"

// In code
report.SnagVersion = version
report.GoVersion = runtime.Version()
report.OS = runtime.GOOS
report.Arch = runtime.GOARCH
```

**Browser Detection:**

```go
// Reuse existing logic
path, err := findBrowserPath()
if err == nil {
    report.BrowserPath = path
    report.BrowserName = detectBrowserName(path)

    // Get version - return raw output directly
    cmd := exec.Command(path, "--version")
    output, err := cmd.Output()
    if err == nil {
        report.BrowserVersion = strings.TrimSpace(string(output))
    }
}
```

**GitHub Version Check:**

```go
// Query GitHub API for latest release
client := &http.Client{Timeout: 10 * time.Second}
resp, err := client.Get("https://api.github.com/repos/grantcarthew/snag/releases/latest")
if err == nil {
    defer resp.Body.Close()
    var release struct {
        TagName string `json:"tag_name"`
    }
    json.NewDecoder(resp.Body).Decode(&release)
    report.LatestVersion = strings.TrimPrefix(release.TagName, "v")
}
// On error, LatestVersion remains empty (handled in output)
```

**Connection Status:**

```go
// Try to connect to port
browser, err := connectToExisting(port)
if err == nil {
    // Connected - get tab count
    pages, err := browser.Pages()
    status.Running = true
    status.TabCount = len(pages)
}
```

**Profile Path:**

```go
// Get profile path from the matched browserDetectionRule
// The rule that matched during detectBrowserName() contains profilePathMac and profilePathLinux fields
profilePath, exists := bm.GetProfilePath()
if profilePath != "" {
    report.ProfilePath = profilePath
    report.ProfileExists = exists
}

// In browser.go:
func (bm *BrowserManager) GetProfilePath() (string, bool) {
    home, _ := os.UserHomeDir()

    // Find the rule that matched this browser
    rule := findMatchingRule(bm.browserPath)
    if rule == nil {
        return "", false
    }

    var profileSubdir string
    if runtime.GOOS == "darwin" {
        profileSubdir = rule.profilePathMac
        profilePath := filepath.Join(home, "Library/Application Support", profileSubdir)
        _, err := os.Stat(profilePath)
        return profilePath, err == nil
    } else {
        profileSubdir = rule.profilePathLinux
        profilePath := filepath.Join(home, ".config", profileSubdir)
        _, err := os.Stat(profilePath)
        return profilePath, err == nil
    }
}
```

**Environment Variables:**

```go
envVars := map[string]string{
    "CHROME_PATH":    os.Getenv("CHROME_PATH"),
    "CHROMIUM_PATH":  os.Getenv("CHROMIUM_PATH"),
}
```

### Output Formatting

Use consistent formatting:

- Section headers with underlines (Unicode box drawing characters)
- Checkmarks (✓) for success/exists
- X marks (✗) for failure/not found
- Indentation for readability
- Clear labels and values aligned

Example formatting helper:

```go
func printSection(title string) {
    fmt.Fprintf(os.Stderr, "\n%s\n", title)
    fmt.Fprintf(os.Stderr, "%s\n", strings.Repeat("─", len(title)))
}

func printItem(label, value string) {
    fmt.Fprintf(os.Stderr, "  %-20s %s\n", label+":", value)
}

func printCheck(label, value string, ok bool) {
    mark := "✗"
    if ok {
        mark = "✓"
    }
    fmt.Fprintf(os.Stderr, "  %-20s %s %s\n", label+":", mark, value)
}
```

## Flag Interactions

**Priority order:** `help > doctor > version > everything else`

| Combination                     | Behavior       | Notes                                             |
| ------------------------------- | -------------- | ------------------------------------------------- |
| `--doctor` + `--help`           | Help wins      | Show help, ignore doctor                          |
| `--doctor` + `--version`        | Doctor wins    | Doctor includes version info                      |
| `--doctor` + `--port`           | Works together | Shows status for both default and custom port     |
| `--doctor` + `--verbose`        | Works together | Verbose logging to stderr during diagnostics      |
| `--doctor` + `--debug`          | Works together | Debug logging to stderr during diagnostics        |
| `--doctor` + `--quiet`          | Doctor wins    | Ignored (doctor output cannot be suppressed)      |
| `--doctor` + URL                | Doctor wins    | Ignore URL, show diagnostics                      |
| `--doctor` + `--list-tabs`      | Doctor wins    | Ignore list-tabs, show diagnostics                |
| `--doctor` + `--all-tabs`       | Doctor wins    | Ignore all-tabs, show diagnostics                 |
| `--doctor` + `--tab`            | Doctor wins    | Ignore tab, show diagnostics                      |
| `--doctor` + `--open-browser`   | Doctor wins    | Ignore open-browser, show diagnostics             |
| `--doctor` + `--url-file`       | Doctor wins    | Ignore url-file, show diagnostics                 |
| `--doctor` + `--output`         | Doctor wins    | Ignore output, show diagnostics                   |
| `--doctor` + `--output-dir`     | Doctor wins    | Ignore output-dir, show diagnostics               |
| `--doctor` + `--format`         | Doctor wins    | Ignore format, show diagnostics                   |
| `--doctor` + `--wait-for`       | Doctor wins    | Ignore wait-for, show diagnostics                 |
| `--doctor` + `--timeout`        | Doctor wins    | Ignore timeout, show diagnostics                  |
| `--doctor` + `--user-agent`     | Doctor wins    | Ignore user-agent, show diagnostics               |
| `--doctor` + `--user-data-dir`  | Doctor wins    | Ignore user-data-dir, show diagnostics            |
| `--doctor` + `--close-tab`      | Doctor wins    | Ignore close-tab, show diagnostics                |
| `--doctor` + `--force-headless` | Doctor wins    | Ignore force-headless, show diagnostics           |
| `--doctor` + `--kill-browser`   | Doctor wins    | Ignore kill-browser, show diagnostics             |

## Documentation Updates Required

### 1. Create `docs/arguments/doctor.md`

Follow existing argument documentation pattern:

- Description and purpose
- Output sections explained
- Interaction matrix (table above)
- Examples
- Use cases (troubleshooting, bug reports)

### 2. Update `docs/design-record.md`

Add to the "Arguments" section:

```markdown
- **Diagnostics**: [Doctor](arguments/doctor.md)
```

Add design decision entry:

```markdown
### DD-XX: Doctor Diagnostic Flag

**Decision:** Add `--doctor` flag for comprehensive diagnostic output

**Rationale:**

- Helps users troubleshoot their environment
- Provides maintainers essential debug info for issue reports
- Single command to gather all relevant system information
- Similar to other tools (brew doctor, npm doctor, etc.)

**Implementation:**

- Overrides all flags except `--help`
- Works with `--port` to check custom ports
- Always outputs to stderr (never stdout)
- Never returns error (diagnostic information always useful)
- Formatted for readability with sections and checkmarks
```

### 3. Review and update ALL `docs/arguments/*.md` files

Every argument document needs interaction rules with `--doctor`:

Files to update:

- `docs/arguments/all-tabs.md`
- `docs/arguments/close-tab.md`
- `docs/arguments/debug.md`
- `docs/arguments/force-headless.md`
- `docs/arguments/format.md`
- `docs/arguments/help.md`
- `docs/arguments/kill-browser.md` (new)
- `docs/arguments/list-tabs.md`
- `docs/arguments/open-browser.md`
- `docs/arguments/output-dir.md`
- `docs/arguments/output.md`
- `docs/arguments/port.md`
- `docs/arguments/quiet.md`
- `docs/arguments/tab.md`
- `docs/arguments/timeout.md`
- `docs/arguments/url-file.md`
- `docs/arguments/user-agent.md`
- `docs/arguments/user-data-dir.md`
- `docs/arguments/verbose.md`
- `docs/arguments/version.md`
- `docs/arguments/wait-for.md`

For most arguments, add to their interaction matrix:

```markdown
| `--<flag>` + `--doctor` | Doctor wins | Diagnostics override normal operation |
```

### 4. Update `README.md`

Add to Troubleshooting section:

````markdown
### Diagnostic Information

Get comprehensive diagnostic information about your snag environment:

```bash
# Run diagnostics
snag --doctor

# Check specific port
snag --doctor --port 9223
```
````

This shows:

- snag and Go versions
- Detected browser and version
- Browser connection status
- Profile locations for all common browsers
- Environment variables
- Working directory

**Use this when:**

- Troubleshooting issues
- Reporting bugs (include doctor output)
- Checking if browser is running
- Finding profile paths

````

### 5. Update `AGENTS.md`

Add to troubleshooting section:
```bash
# Get diagnostic information
snag --doctor

# Include in bug reports - redirect to file
snag --doctor > diagnostics.txt 2>&1
````

Add to "Troubleshooting" section in AGENTS.md:

````markdown
**Get diagnostic information:**

```bash
snag --doctor
```
````

Include this output when reporting issues.

````

## Testing

### Manual Test Cases

```bash
# Test 1: Basic doctor output
snag --doctor
# Expected: Full diagnostic report

# Test 2: Doctor with custom port
snag --open-browser --port 9223
snag --doctor --port 9223
# Expected: Shows both port 9222 and 9223 status

# Test 3: Doctor overrides URL
snag --doctor https://example.com
# Expected: Shows diagnostics, ignores URL

# Test 4: Doctor overrides list-tabs
snag --doctor --list-tabs
# Expected: Shows diagnostics, ignores list-tabs

# Test 5: Help overrides doctor
snag --doctor --help
# Expected: Shows help, ignores doctor

# Test 6: Version overrides doctor
snag --doctor --version
# Expected: Shows version, ignores doctor (Actually, help > version > doctor, so version should win)

# Test 7: Doctor on system with no browser
# (Need test environment without chrome/chromium)
# Expected: Reports "No Chromium-based browser found"

# Test 8: Doctor with running browser
snag --open-browser
snag --doctor
# Expected: Shows browser running with tab count

# Test 9: Doctor with multiple browsers installed
# (System with Chrome, Brave, etc.)
# Expected: Shows detected browser + all profile paths with existence

# Test 10: Doctor with environment variables
export CHROME_PATH=/opt/google/chrome/chrome
snag --doctor
# Expected: Shows CHROME_PATH in environment section
````

### Automated Tests

Add to `cli_test.go` or new `doctor_test.go`:

- `TestDoctorBasicOutput()` - runs successfully, contains expected sections
- `TestDoctorOverridesURL()` - doctor wins over URL argument
- `TestDoctorWithPort()` - shows custom port status
- `TestDoctorHelpWins()` - help flag overrides doctor
- `TestDoctorVersionWins()` - version flag overrides doctor

### Output Validation Tests

Verify output contains:

- Version section with snag version
- Browser section with path
- Connection section with port status
- Profile section with at least one browser
- Environment section

## Implementation Phases

### Phase 1: Core Diagnostic Collection

1. Add flag definition to `main.go`
2. Create `handleDoctor()` in `handlers.go`
3. Implement version info collection
4. Implement browser detection reuse
5. Implement working directory display
6. Basic formatted output

### Phase 2: Connection & Profile Status

1. Implement port connection checking
2. Add tab counting when connected
3. Implement profile path detection (OS-specific)
4. Add existence checking for profiles

### Phase 3: Enhanced Information

1. Implement browser version detection
2. Add environment variable collection
3. Add custom port support (`--port` flag)
4. Refine output formatting

### Phase 4: Documentation

1. Create `docs/arguments/doctor.md`
2. Update `docs/design-record.md`
3. Review and update ALL argument docs for interactions
4. Update README.md with troubleshooting section
5. Update AGENTS.md

### Phase 5: Testing & Polish

1. Manual testing on Linux and macOS
2. Test with various browser installations
3. Test with/without browsers running
4. Automated test cases
5. Output formatting refinement

## Success Criteria

- [ ] CLI flag `--doctor` defined and working
- [ ] Overrides all flags except `--help`
- [ ] Works with `--port` to check custom ports
- [ ] Displays all required information sections
- [ ] Formatted output with sections, checkmarks, alignment
- [ ] Works on Linux and macOS
- [ ] Handles missing browser gracefully
- [ ] Handles no browser running gracefully
- [ ] `docs/arguments/doctor.md` created
- [ ] `docs/design-record.md` updated with design decision
- [ ] ALL `docs/arguments/*.md` files reviewed and updated
- [ ] README.md updated with troubleshooting section
- [ ] AGENTS.md updated with diagnostic examples
- [ ] Manual testing completed on both platforms
- [ ] Automated tests passing
- [ ] No regressions in existing functionality

## Implementation Decisions

1. **Browser version parsing:**
   - ✓ **Decision:** Display raw output from `--version` (no parsing)

2. **Tab counting timeout:**
   - **Open:** Should connection attempts have a timeout? Suggest 5 seconds.

3. **Additional information:**
   - ✓ **Decision:** Keep it browser-focused for v1, expand later if needed

4. **Output destination:**
   - ✓ **Decision:** stdout (diagnostic report), stderr (verbose/debug logs)

5. **Machine-readable output:**
   - ✓ **Decision:** No JSON output - human-readable is grep-friendly

6. **Profile path organization:**
   - ✓ **Decision:** Show only detected browser's profile path (simplified)

## Notes

- Doctor should NEVER modify state (read-only diagnostic)
- Doctor always exits with code 0 (unless exception occurs)
- Always provides information even if partial (no failures)
- Output goes to stdout for easy piping/redirection
- Verbose/debug logs go to stderr during diagnostic operations
- Output includes repo link in header for easy bug reporting
- Output is copy-paste friendly for bug reports (plain text, Unicode/emoji ok)
- Will be used by future `--report-issue` flag to auto-populate GitHub issues
- Future enhancement: Include test fetching a known URL to validate full pipeline
- Future enhancement: More detailed network/system diagnostics if needed
