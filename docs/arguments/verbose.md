# `--verbose`

**Status:** Complete (2025-10-23)

#### Validation Rules

**Boolean Flag:**

- No value required (presence = enabled)
- No validation errors possible

**Multiple Flag Conflicts:**

- Multiple `--verbose` flags → Last flag honored (standard Unix behavior)
- `--verbose` + `--quiet` → Last flag wins
- `--verbose` + `--debug` → Last flag wins

#### Behavior

**Logging Level:**

- Enables verbose logging output to stderr
- Shows additional information about operations:
  - Browser connection details
  - Page navigation steps
  - Content conversion progress
  - File writing confirmations
- Does not affect stdout content output

**Basic Usage:**

```bash
snag https://example.com --verbose
```

- Outputs page content to stdout (as normal)
- Logs verbose messages to stderr:
  - "Connecting to Chrome on port 9222..."
  - "Navigating to https://example.com..."
  - "Waiting for page load..."
  - "Converting HTML to Markdown..."
  - "Content written to stdout"

#### Interaction Matrix

**Logging Level Priority (Last Flag Wins):**

| Combination                 | Effective Level | Rationale                      |
| --------------------------- | --------------- | ------------------------------ |
| `--verbose`                 | Verbose         | Standard verbose mode          |
| `--verbose --quiet`         | Quiet           | Last flag wins (Unix standard) |
| `--quiet --verbose`         | Verbose         | Last flag wins                 |
| `--verbose --debug`         | Debug           | Last flag wins                 |
| `--debug --verbose`         | Verbose         | Last flag wins                 |
| `--verbose --quiet --debug` | Debug           | Last flag wins                 |

**All Other Flags:**

- Works normally with all other flags
- Simply controls stderr logging verbosity
- No conflicts or special behaviors

**Examples with Other Flags:**

- `--verbose` + `--user-data-dir` - Works normally (verbose logs with custom profile)
- `--verbose` + `--user-agent` - Works normally (verbose logs with custom UA)
- `--verbose` + all browser/output/timing flags - Works normally

#### Examples

**Valid:**

```bash
snag https://example.com --verbose                  # Verbose logging
snag https://example.com --verbose -o page.md       # Verbose + file output
snag --url-file urls.txt --verbose                  # Verbose batch processing
snag --tab 1 --verbose                              # Verbose tab fetch
snag --list-tabs --verbose                          # Verbose tab listing
snag https://example.com --quiet --verbose          # Verbose wins (last flag)
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

1. Check if `--verbose` flag is present
2. If multiple logging flags, last one wins
3. Initialize logger with verbose level
4. All subsequent operations use verbose logging

**Logging Behavior:**

- Normal output: Important messages only
- Verbose output: All operational details
- Logs go to stderr (stdout reserved for content)

---
