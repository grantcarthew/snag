# `--quiet` / `-q`

**Status:** Complete (2025-10-23)

#### Validation Rules

**Boolean Flag:**
- No value required (presence = enabled)
- No validation errors possible

**Multiple Flag Conflicts:**
- Multiple `--quiet` flags → Last flag honored (standard Unix behavior)
- `--quiet` + `--verbose` → Last flag wins
- `--quiet` + `--debug` → Last flag wins

#### Behavior

**Logging Level:**
- Suppresses all logging output to stderr except errors
- Shows only:
  - Fatal errors
  - Critical validation failures
  - Operation completion status (success/failure)
- Ideal for scripting and automation

**Basic Usage:**
```bash
snag https://example.com --quiet
```
- Outputs page content to stdout (as normal)
- Suppresses all stderr messages except errors
- Silent operation on success

#### Interaction Matrix

**Logging Level Priority (Last Flag Wins):**

| Combination | Effective Level | Rationale |
|-------------|----------------|-----------|
| `--quiet` | Quiet | Suppress all but errors |
| `--quiet --verbose` | Verbose | Last flag wins (Unix standard) |
| `--verbose --quiet` | Quiet | Last flag wins |
| `--quiet --debug` | Debug | Last flag wins |
| `--debug --quiet` | Quiet | Last flag wins |
| `--quiet --verbose --debug` | Debug | Last flag wins |

**All Other Flags:**
- Works normally with all other flags
- Simply controls stderr logging verbosity
- No conflicts or special behaviors

**Examples with Other Flags:**
- `--quiet` + `--user-data-dir` - Works normally (quiet mode with custom profile)
- `--quiet` + `--user-agent` - Works normally (quiet mode with custom UA)
- `--quiet` + all browser/output/timing flags - Works normally

#### Examples

**Valid:**
```bash
snag https://example.com --quiet                    # Silent operation
snag https://example.com -q -o page.md              # Silent file save
snag --url-file urls.txt --quiet                    # Silent batch processing
snag --tab 1 --quiet                                # Silent tab fetch
snag --list-tabs --quiet                            # Silent tab listing (shows tabs only)
snag https://example.com --verbose --quiet          # Quiet wins (last flag)
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
1. Check if `--quiet` flag is present
2. If multiple logging flags, last one wins
3. Initialize logger with quiet level
4. All subsequent operations suppress non-error logs

**Logging Behavior:**
- Quiet output: Errors only
- Exit code 0 on success (silent)
- Exit code 1 on failure (error message shown)

---
