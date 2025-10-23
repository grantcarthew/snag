# `--all-tabs` / `-a`

**Status:** Incomplete (Design in progress)

## Overview

Fetch content from all open tabs in an existing browser connection.

## Design Questions

This argument's behavior is currently being designed. See PROJECT.md Task 14 for design decisions in progress.

## Planned Features

- Fetch from all open tabs
- Requires existing browser connection
- Mutually exclusive with `<url>`, `--url-file`, `--tab`
- Works with `--output-dir` for auto-generated filenames
- Works with `--wait-for` to apply selector to all tabs
