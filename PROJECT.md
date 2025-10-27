# PROJECT: Migrate from urfave/cli to Cobra

**Status:** Planning
**Priority:** Medium
**Effort:** 2-4 hours
**Start Date:** TBD
**Target Completion:** TBD

## Overview

Migrate snag's CLI framework from `github.com/urfave/cli/v2` to `github.com/spf13/cobra` to gain better flag handling, cleaner help output (no `default: false` spam), and align with industry-standard CLI tooling.

## Goals

### Primary Goals
1. Remove `(default: false)` clutter from boolean flags in help output
2. Maintain 100% functional compatibility with current CLI interface
3. Preserve custom "AGENT USAGE" help section
4. Improve maintainability with industry-standard framework

### Success Criteria
- ✅ All existing flags work identically
- ✅ Help output is cleaner (no boolean default spam)
- ✅ AGENT USAGE section preserved in help
- ✅ All tests pass
- ✅ Binary size ≤ current size + 10%
- ✅ No behavioral changes for users

## Current State Assessment

### Current Framework: urfave/cli v2.27.7

**Flags (19 total):**
- String flags: `url-file`, `output`, `output-dir`, `format`, `wait-for`, `tab`, `user-agent`, `user-data-dir`
- Int flags: `timeout`, `port`
- Bool flags (9): `close-tab`, `force-headless`, `open-browser`, `list-tabs`, `all-tabs`, `verbose`, `quiet`, `debug`

**Custom Features:**
- Custom help template with AGENT USAGE section
- Custom USAGE pattern (`URL...`)
- Removed NAME and AUTHOR sections

**Dependencies:**
```
github.com/urfave/cli/v2 v2.27.7
```

### Target Framework: Cobra

**Version:** Latest stable (github.com/spf13/cobra)

**Known Advantages:**
- `MarkFlagRequired()` for required flags
- `MarkHidden()` for hiding defaults
- Better help template customization
- Industry standard (kubectl, docker, gh, hugo)
- Active maintenance

## Migration Phases

### Phase 1: Preparation (30 minutes)

**Tasks:**

1. **Create migration branch**
   ```bash
   git checkout -b migrate-to-cobra
   ```

2. **Document current behavior**
   - [ ] Run `./snag --help > docs/help-before-migration.txt`
   - [ ] Document all flag combinations in test matrix
   - [ ] Capture current binary size: `ls -lh snag`

3. **Add Cobra dependency**
   - [ ] Run `go get github.com/spf13/cobra@latest`
   - [ ] Run `go mod tidy`

4. **Create parallel implementation file**
   - [ ] Create `main_cobra.go` (implement alongside existing `main.go`)
   - [ ] Allows easy comparison and rollback

### Phase 2: Core Migration (1-2 hours)

**Tasks:**

5. **Create root command structure**
   - [ ] Define `rootCmd` with cobra.Command
   - [ ] Set `Use: "snag"`
   - [ ] Set `Short` description
   - [ ] Set `Long` description (from current Description field)
   - [ ] Set `Version` from version variable

6. **Migrate all flags**
   - [ ] String flags: `rootCmd.Flags().StringP()` for each
   - [ ] Int flags: `rootCmd.Flags().IntP()` for each
   - [ ] Bool flags: `rootCmd.Flags().BoolP()` for each
   - [ ] Preserve all aliases (-o, -d, -f, etc.)
   - [ ] Set default values (format="md", timeout=30, port=9222)

7. **Migrate flag validation**
   - [ ] Port `validateURL()` logic
   - [ ] Port `validateFormat()` logic
   - [ ] Port `validateTimeout()` logic
   - [ ] Port `validatePort()` logic
   - [ ] Port `validateOutputPath()` logic
   - [ ] Port `validateDirectory()` logic
   - [ ] Port `validateUserDataDir()` logic
   - [ ] Port `validateUserAgent()` logic
   - [ ] Port `validateWaitFor()` logic

8. **Migrate main action handler**
   - [ ] Create `Run` function for rootCmd
   - [ ] Port `run()` function logic
   - [ ] Port URL argument parsing
   - [ ] Port flag conflict detection
   - [ ] Port logging level detection

9. **Migrate handler functions**
   - [ ] Port `handleListTabs()`
   - [ ] Port `handleAllTabs()`
   - [ ] Port `handleTabFetch()`
   - [ ] Port `handleOpenURLsInBrowser()`
   - [ ] Port `handleMultipleURLs()`
   - [ ] Port `snag()` core function

### Phase 3: Custom Help Template (1 hour)

**Tasks:**

10. **Create custom help template**
    - [ ] Create help template with USAGE section
    - [ ] Add DESCRIPTION section
    - [ ] Add AGENT USAGE section (preserve current content)
    - [ ] Add GLOBAL OPTIONS section
    - [ ] Verify no NAME section appears
    - [ ] Verify no AUTHOR section appears
    - [ ] Verify VERSION section is removed

11. **Configure template in Cobra**
    - [ ] Use `rootCmd.SetUsageTemplate()`
    - [ ] Test template rendering: `./snag --help`
    - [ ] Compare with `docs/help-before-migration.txt`

12. **Verify AGENT USAGE content**
    - [ ] Common workflows section intact
    - [ ] Integration tips section intact
    - [ ] Performance note intact
    - [ ] Concrete examples preserved (example.com, not <url>)

### Phase 4: Testing (30 minutes)

**Tasks:**

13. **Unit test verification**
    - [ ] Run `go test -v ./...`
    - [ ] Fix any broken tests
    - [ ] Verify all tests pass

14. **Manual testing - Basic operations**
    - [ ] `snag https://example.com` (basic fetch)
    - [ ] `snag --format html https://example.com`
    - [ ] `snag -o output.md https://example.com`
    - [ ] `snag -d output/ https://example.com https://google.com`

15. **Manual testing - Browser operations**
    - [ ] `snag --open-browser`
    - [ ] `snag --list-tabs`
    - [ ] `snag --tab 1`
    - [ ] `snag --tab "example"`
    - [ ] `snag --all-tabs -d output/`

16. **Manual testing - Edge cases**
    - [ ] `snag --force-headless https://example.com`
    - [ ] `snag --timeout 60 https://example.com`
    - [ ] `snag --wait-for ".content" https://example.com`
    - [ ] `snag --user-agent "Custom Agent" https://example.com`
    - [ ] `snag --verbose https://example.com`
    - [ ] `snag --quiet https://example.com`
    - [ ] `snag --debug https://example.com`

17. **Manual testing - Flag conflicts**
    - [ ] `snag --tab 1 https://example.com` (should error)
    - [ ] `snag --all-tabs https://example.com` (should error)
    - [ ] `snag --output out.md --output-dir ./` (should error)
    - [ ] `snag --force-headless --open-browser` (should error)

18. **Help output verification**
    - [ ] Run `./snag --help > docs/help-after-migration.txt`
    - [ ] Verify no `(default: false)` appears
    - [ ] Verify AGENT USAGE section present
    - [ ] Verify concrete examples (example.com)
    - [ ] Compare line-by-line with expected output

### Phase 5: Cleanup & Finalization (30 minutes)

**Tasks:**

19. **Remove old framework code**
    - [ ] Delete or rename `main.go` to `main_urfave.go.bak`
    - [ ] Rename `main_cobra.go` to `main.go`
    - [ ] Remove urfave/cli dependency: `go mod tidy`

20. **Build verification**
    - [ ] `go build -o snag`
    - [ ] Verify binary builds without errors
    - [ ] Check binary size: `ls -lh snag`
    - [ ] Compare with pre-migration size

21. **Update documentation**
    - [ ] Update AGENTS.md if CLI interface changed
    - [ ] Update README.md examples if needed
    - [ ] Document any breaking changes (should be none)
    - [ ] Update this PROJECT.md with completion status

22. **Final commit**
    - [ ] Run `go test -v ./...` one final time
    - [ ] Commit with message: `feat: migrate from urfave/cli to cobra`
    - [ ] Push branch: `git push -u origin migrate-to-cobra`

### Phase 6: Review & Merge (Optional)

**Tasks:**

23. **Create PR for review** (if desired)
    - [ ] Create pull request on GitHub
    - [ ] Add before/after help output comparison
    - [ ] Document migration rationale
    - [ ] Self-review changes

24. **Merge to main**
    - [ ] Merge PR or merge branch directly
    - [ ] Delete migration branch
    - [ ] Tag release if appropriate

## Risk Mitigation

### Rollback Plan

If migration fails or introduces bugs:

1. **Immediate rollback:**
   ```bash
   git checkout main
   git branch -D migrate-to-cobra
   ```

2. **Keep old implementation:**
   - Keep `main_urfave.go.bak` until migration verified in production
   - Can quickly revert if issues discovered

### Known Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Help template formatting different | Low | Extensive testing in Phase 3 |
| Flag parsing behavior differs | Medium | Comprehensive test suite in Phase 4 |
| Binary size increases significantly | Low | Monitor in Phase 5, acceptable up to +10% |
| Undiscovered flag conflicts | Medium | Test all combinations in Phase 4 |
| Custom AGENT USAGE section breaks | Low | Template testing in Phase 3 |

## Testing Strategy

### Test Categories

1. **Unit tests** - All existing tests must pass
2. **Manual CLI tests** - All flag combinations
3. **Help output tests** - Visual comparison
4. **Integration tests** - Real browser operations
5. **Edge case tests** - Error conditions

### Test Matrix

| Category | Test Cases | Pass Criteria |
|----------|-----------|---------------|
| Basic fetch | Single URL, multiple URLs | Content fetched correctly |
| Output formats | md, html, text, pdf, png | All formats work |
| File output | -o, -d flags | Files created correctly |
| Browser modes | headless, visible, existing | All modes work |
| Tab operations | list, select, all | Tab features work |
| Logging | verbose, quiet, debug | Log levels work |
| Validation | Invalid URLs, formats, timeouts | Proper errors shown |
| Flag conflicts | Mutually exclusive flags | Proper errors shown |

## Dependencies

### New Dependencies
- `github.com/spf13/cobra` (latest)

### Removed Dependencies
- `github.com/urfave/cli/v2` v2.27.7

### Unchanged Dependencies
- `github.com/go-rod/rod` v0.116.2
- `github.com/JohannesKaufmann/html-to-markdown/v2` v2.4.0
- `github.com/k3a/html2text` v1.2.1

## Post-Migration

### Validation Checklist

After merging to main:

- [ ] Run full test suite on clean checkout
- [ ] Verify `--help` output meets requirements
- [ ] Build for all target platforms
- [ ] Test on macOS and Linux
- [ ] Update release documentation

### Future Enhancements Enabled

With Cobra, these become easier:

1. Subcommands (if needed): `snag browser`, `snag tabs`, etc.
2. Better shell completion
3. More sophisticated flag validation
4. Plugin architecture (if desired)

## Notes

- Migration should be done in one continuous session to avoid context switching
- Keep `docs/help-before-migration.txt` and `docs/help-after-migration.txt` for comparison
- The AGENT USAGE section is the most critical custom feature to preserve
- Boolean flags should NOT show `(default: false)` in new implementation

## References

- Cobra documentation: https://github.com/spf13/cobra
- Cobra user guide: https://cobra.dev/
- Current urfave/cli code: `main.go`
- Help template reference: `customAppHelpTemplate` variable in main.go
