# P-002: CLI Info Flag

- Status: Completed
- Started: 2025-02-04
- Completed: 2025-02-04

## Overview

Add an `--info` flag that outputs page metadata as JSON. This enables automation scripts to extract page title, URL, and other metadata without parsing HTML or markdown output.

Primary use case: job application automation scripts that need the page title to generate directory names.

## Goals

1. Add `--info` / `-i` flag to output page metadata as JSON
2. Support fetching info from URLs and existing tabs
3. Allow saving JSON output to file with `-o`

## Scope

In Scope:

- New `--info` flag with `-i` shorthand
- JSON output with defined fields
- Works with URL argument
- Works with `--tab` flag
- Works with `-o` output flag
- Works with `--quiet` flag

Out of Scope:

- Multiple URL support (single URL only for now)
- Custom field selection
- Non-JSON output formats

## Success Criteria

- [x] `snag --info <url>` outputs JSON to stdout
- [x] `snag -i <url>` works as shorthand
- [x] `snag --info --tab 1` gets info from existing tab
- [x] `snag --info -o info.json <url>` saves to file
- [x] `snag --info --quiet <url>` suppresses logs, outputs only JSON (note: --info is quiet by default)
- [x] Comprehensive tests covering success and error cases
- [x] README updated with `--info` documentation

## Deliverables

- Modified `main.go` - Add flag definition
- Modified `handlers.go` or new `info.go` - Handle info output
- Modified `README.md` - Document the flag
- `info_test.go` - Comprehensive tests:
  - Valid URL returns correct JSON fields
  - Tab mode works (`--info --tab 1`)
  - Output to file works (`--info -o`)
  - Quiet mode works (`--info --quiet`)
  - Error cases (invalid URL, no browser, no tab match)
  - Flag conflicts (e.g., `--info --format html`)

## Technical Approach

JSON output fields:

```json
{
  "title": "Senior DevOps Engineer - SEEK",
  "url": "https://www.seek.com.au/job/82097817",
  "domain": "seek.com.au",
  "slug": "senior-devops-engineer-seek",
  "timestamp": "2025-02-04T14:30:22+10:00"
}
```

Flag behaviour:

- Mutually exclusive with `--format` (info always outputs JSON)
- Compatible with `--tab` for existing browser tabs
- Compatible with `-o` for file output
- Compatible with `--quiet` for clean output
- Requires exactly one URL (or `--tab`)

Implementation notes:

- Reuse existing `SlugifyTitle()` from output.go
- Reuse page title extraction from fetch logic
- Add new handler function for `--info` mode
