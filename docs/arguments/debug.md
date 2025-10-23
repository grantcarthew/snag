# `--debug`

**Status:** Complete (2025-10-23)

#### Validation Rules

**Boolean Flag:**
- No value required (presence = enabled)
- No validation errors possible

**Multiple Flag Conflicts:**
- Multiple `--debug` flags → Last flag honored (standard Unix behavior)
- `--debug` + `--verbose` → Last flag wins
- `--debug` + `--quiet` → Last flag wins

#### Behavior

**Logging Level:**
- Enables maximum logging output to stderr
- Shows all verbose information plus:
  - Chrome DevTools Protocol (CDP) messages
  - Browser connection debugging
  - Internal state information
  - Detailed error traces
- For troubleshooting and development

**Basic Usage:**
```bash
snag https://example.com --debug
```
- Outputs page content to stdout (as normal)
- Logs extensive debug information to stderr
- Includes CDP protocol messages

#### Interaction Matrix

**Logging Level Priority (Last Flag Wins):**

| Combination | Effective Level | Rationale |
|-------------|----------------|-----------|
| `--debug` | Debug | Maximum logging |
| `--debug --verbose` | Verbose | Last flag wins (Unix standard) |
| `--verbose --debug` | Debug | Last flag wins |
| `--debug --quiet` | Quiet | Last flag wins |
| `--quiet --debug` | Debug | Last flag wins |
| `--debug --quiet --verbose` | Verbose | Last flag wins |

**All Other Flags:**
- Works normally with all other flags
- Simply controls stderr logging verbosity
- No conflicts or special behaviors

#### Examples

**Valid:**
```bash
snag https://example.com --debug                    # Debug logging
snag https://example.com --debug -o page.md         # Debug + file output
snag --url-file urls.txt --debug                    # Debug batch processing
snag --tab 1 --debug                                # Debug tab fetch
snag --list-tabs --debug                            # Debug tab listing
snag https://example.com --verbose --debug          # Debug wins (last flag)
```

**No Invalid Combinations:**
- Boolean flag, no invalid values
- Works with everything

#### Implementation Details

**Location:**
- Flag definition: `main.go` (CLI framework)
- Logger initialization: `main.go:181-187`
- Logging level: `logger.go`

**Processing:**
1. Check if `--debug` flag is present
2. If multiple logging flags, last one wins
3. Initialize logger with debug level
4. All subsequent operations use debug logging
5. CDP messages logged via rod's debug capabilities

**Logging Behavior:**
- Debug output: Everything (verbose + CDP messages + internals)
- Extremely detailed for troubleshooting
- Logs go to stderr (stdout reserved for content)

---
