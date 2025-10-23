# `--user-agent STRING`

**Status:** Incomplete (Design in progress)

## Overview

Set a custom user agent string for browser requests.

## Design Questions

This argument's behavior is currently being designed. See PROJECT.md Task 18 for design decisions in progress.

## Planned Features

- Set custom user agent for new page navigation
- Should be ignored/warned when used with `--tab` (can't change existing tab's user agent)
- Works normally with URL fetching modes
- May be ignored when used with `--list-tabs` (no navigation)
