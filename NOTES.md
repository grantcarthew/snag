# snag - Development Notes

Minor items, considerations, and reminders that don't fit in the main design document.

## Testing Requirements

- **Test suite needed** - Not currently part of design document
  - Unit tests for core functions
  - Integration tests with real browser
  - Mock vs real browser considerations
  - CI/CD pipeline integration
  - Test fixtures and example pages
  - Add to PROJECT.md design decisions

## Argument Parsing

- **Position independence required** - Unlike bash version, Go version should accept flags in any position:
  - `snag -v example.com` ✓
  - `snag example.com -v` ✓
  - Both should work identically
  - This is standard behavior with most Go CLI frameworks (cobra, urfave/cli)

## License & Copyright

- **License Headers**: Add MPL 2.0 header to all Go source files when creating them

  ```go
  // Copyright (c) 2025 Grant Carthew
  //
  // This Source Code Form is subject to the terms of the Mozilla Public
  // License, v. 2.0. If a copy of the MPL was not distributed with this
  // file, You can obtain one at https://mozilla.org/MPL/2.0/.

  package main
  ```

- Apply this header to every new `.go` file created
- See `LICENSE` file Exhibit A (lines 355-367) for reference

## Open Questions

- [ ] Add items here as they come up during development

## Future Considerations

- [ ] Add items for post-MVP features here

---

**Created**: 2025-10-16
