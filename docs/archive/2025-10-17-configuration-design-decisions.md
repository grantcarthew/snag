# snag - Configuration Design Decisions

This document tracks configuration-related design decisions and potential improvements.

## Active Considerations

### Config Validation Pattern

**Status**: Under consideration

**Current Implementation**:

- `Config` struct is a simple data holder (main.go:268-279)
- Validation logic inline in `run()` function (main.go:185-193)
- URL validation in separate `validate.go` module

**Question**: Where should config validation live?

**Options**:

1. **Keep current approach** (validation inline)

   - Pro: Simple, clear, logging/errors at call site
   - Pro: Only one field currently validated
   - Con: Mixed concerns in `run()` function

2. **Add to validate.go** (functional approach)

   ```go
   func validateConfig(config *Config) error {
       if !validFormats[config.Format] {
           return ErrInvalidFormat
       }
       return nil
   }
   ```

   - Pro: Consistent with existing `validateURL()` pattern
   - Pro: All validation in one module
   - Con: Separates validation from error logging/suggestions

3. **Method on Config** (OOP approach)
   ```go
   func (c *Config) Validate() error {
       if !validFormats[c.Format] {
           return ErrInvalidFormat
       }
       return nil
   }
   ```
   - Pro: Self-validating struct
   - Pro: Easily testable
   - Con: Inconsistent with current validate.go pattern
   - Con: Logging/error suggestions still needed at call site

**Consideration**: Currently only format field is validated. URL is validated separately before Config creation. May be premature abstraction.

**Decision**: Deferred - document for future when more config validation is needed.

---

## Related Files

- main.go:268-279 - Config struct definition
- main.go:185-193 - Current validation logic
- validate.go - URL validation module
- errors.go - Error definitions

---

**Document Created**: 2025-10-17
