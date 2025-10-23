# `--user-data-dir DIRECTORY`

**Status:** Incomplete (Design in progress)

## Overview

Specify a custom user data directory for the browser instance, enabling profile isolation and multi-instance support.

## Design Questions

This argument's behavior is currently being designed. See PROJECT.md Task 21 for design decisions in progress.

## Planned Features

- Custom browser profile directory for session isolation
- Enable multiple authenticated sessions (personal vs work accounts)
- Session isolation per project/client
- Privacy (separate from personal browsing)
- Enable true multi-instance browsers with different ports
- Directory creation behavior needs definition
- Validation for wrong values (doesn't exist, not a directory, permission denied, invalid path, empty string)
