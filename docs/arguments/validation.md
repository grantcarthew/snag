# Validation Rules and Order

**Last Updated:** 2025-10-26

This document describes the validation order and cross-cutting validation rules that apply to multiple arguments.

---

## Cross-Cutting Validation Rules

### String Argument Trimming

All string arguments are trimmed using `strings.TrimSpace()` after reading from CLI framework:

- Removes leading and trailing whitespace (spaces, tabs, newlines)
- Applied to: `--output`, `--output-dir`, `--format`, `--wait-for`, `--user-agent`, `--user-data-dir`, `--tab`, `--url-file`, and `<url>` positional arguments
- Empty strings after trimming are handled per-argument (usually warning + ignored or error)
- Standard behavior in most CLI tools (git, docker, etc.)

### Multiple Flag Behavior

**Last Flag Wins (Standard CLI Behavior):**

- When the same flag is specified multiple times, the last value is used
- No error, no warning - silent override
- Applies to all flags:
  - **String flags**: `--output`, `--output-dir`, `--format`, `--wait-for`, `--user-agent`, `--user-data-dir`, `--tab`, `--url-file`
  - **Integer flags**: `--timeout`, `--port`
  - **Boolean flags**: `--close-tab`, `--force-headless`, `--open-browser`, `--list-tabs`, `--all-tabs`
  - **Logging flags**: `--verbose`, `--quiet`, `--debug`

**Examples:**

```bash
snag -o file1.md -o file2.md https://example.com  # Uses file2.md
snag --port 9222 --port 9223 https://example.com  # Uses port 9223
snag --quiet --verbose https://example.com        # Verbose wins (last flag)
```

### Priority Order for Special Flags

Certain flags override all others and exit immediately:

1. `--help` (highest priority) → Display help, exit 0
2. `--version` → Display version, exit 0
3. `--list-tabs` → List tabs, exit 0, ignore all flags except `--port` and logging flags

---

## Validation Order

**Current implementation order (main.go:178-316):**

1. Initialize logger (`--quiet`, `--verbose`, `--debug`) - last flag wins
2. Handle `--help` → exit early (handled by CLI framework)
3. Handle `--version` → exit early (handled by CLI framework)
4. Handle `--open-browser` without URL → exit early
5. Handle `--list-tabs` → extract `--port` and logging flags, ignore all others, list tabs, exit early
6. Handle `--all-tabs` → check for URL conflict, exit early
7. Handle `--tab` → check for URL conflict, exit early
8. Validate URL argument required (if not in special modes above)
9. Validate URL format
10. Validate `-o` + `-d` conflict
11. Validate format
12. Validate timeout
13. Validate port
14. Validate output path (if `-o`)
15. Validate output directory (if `-d`)
16. Execute fetch operation

**Key Patterns:**

- Early exits for standalone modes (help, version, list-tabs, open-browser)
- Content source validation before output validation
- Mutually exclusive flag checks before individual flag validation
- Path/filesystem validation happens last (just before operation)

---

## Notes

For argument-specific validation rules, error messages, and interaction matrices, see the individual argument documentation files in this directory.
