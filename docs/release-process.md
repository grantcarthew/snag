# snag - Release Process

> **Purpose**: Step-by-step instructions for releasing a new version of snag
> **Audience**: AI agents and maintainers performing releases
> **Last Updated**: 2025-10-20

This document provides complete instructions for releasing snag. Follow each section in order.

---

## Overview

**Release Steps Summary**:

1. Pre-release checks
2. Update version numbers
3. Update CHANGELOG.md
4. Commit and tag release
5. Build multi-platform binaries
6. Create GitHub release
7. Update Homebrew tap
8. Verify installation
9. Post-release tasks

**Estimated Time**: 30-45 minutes

---

## Prerequisites

Before starting a release, ensure you have:

- [ ] Write access to `grantcarthew/snag` repository
- [ ] Write access to `grantcarthew/homebrew-tap` repository
- [ ] Go 1.25.3+ installed
- [ ] Git configured with proper credentials
- [ ] All planned features/fixes merged to main branch

---

## Step 1: Pre-Release Checklist

Run through this checklist before starting the release:

```bash
# Ensure you're on main branch
git checkout main
git pull origin main

# Verify all tests pass
go test -v ./...

# Verify build works
go build -o snag
./snag --version

# Clean up test binary
rm snag
```

**Checklist**:

- [ ] All tests passing (71 tests expected)
- [ ] Build completes without errors
- [ ] No uncommitted changes: `git status` is clean
- [ ] All intended features/fixes are merged
- [ ] Documentation is up to date

**If any checks fail, stop and fix issues before proceeding.**

---

## Step 2: Determine Version Number

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (1.0.0 → 2.0.0): Breaking API changes
- **MINOR** (1.0.0 → 1.1.0): New features, backward compatible
- **PATCH** (1.0.0 → 1.0.1): Bug fixes only

**Current version**: Check with `git tag -l | tail -1`

**Next version**: `v<MAJOR>.<MINOR>.<PATCH>`

**Example**: If current is `v0.0.3` and adding features → `v0.1.0`

**Decision**: ******\_\_\_****** (fill in new version number)

---

## Step 3: Update CHANGELOG.md

Create or update `CHANGELOG.md` with release notes.

**If CHANGELOG.md doesn't exist, create it**:

```bash
cat > CHANGELOG.md << 'EOF'
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [VERSION] - YYYY-MM-DD

### Added
- List new features here

### Changed
- List changes to existing functionality here

### Fixed
- List bug fixes here

### Deprecated
- List deprecated features here

### Removed
- List removed features here

### Security
- List security fixes here

[Unreleased]: https://github.com/grantcarthew/snag/compare/vVERSION...HEAD
[VERSION]: https://github.com/grantcarthew/snag/releases/tag/vVERSION
EOF
```

**If CHANGELOG.md exists, update it**:

1. Add new version section above previous releases
2. Move items from `[Unreleased]` to new version section
3. Update the date
4. Update the comparison links at bottom

**Example entry**:

```markdown
## [0.1.0] - 2025-10-20

### Added

- New feature X for better Y
- Support for Z

### Fixed

- Bug in component A
- Issue with B under certain conditions

[Unreleased]: https://github.com/grantcarthew/snag/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/grantcarthew/snag/releases/tag/v0.1.0
[0.0.3]: https://github.com/grantcarthew/snag/releases/tag/v0.0.3
```

**Verification**:

```bash
# Review the CHANGELOG
cat CHANGELOG.md
```

---

## Step 4: Commit Changes

Commit the CHANGELOG and any other pre-release updates:

```bash
# Set version variable (REPLACE with your version)
export VERSION="0.1.0"

# Stage changes
git add CHANGELOG.md

# Commit
git commit -m "chore: prepare for v${VERSION} release"

# Push to main
git push origin main
```

**Verification**: Check `git log -1` shows your commit

---

## Step 5: Create Git Tag

Create a summary of changes and an annotated tag for the release:

```bash
# Get the previous version tag
PREV_VERSION=$(git tag -l | tail -1)
echo "Previous version: $PREV_VERSION"

# Show changes since previous version
git log ${PREV_VERSION}..HEAD --oneline

# Create a one-line summary of changes
# Review the git log output and write a concise summary
# Examples:
#   "Add release process documentation and third-party licenses"
#   "Fix authentication handling and improve error messages"
#   "Add tab management features"

# Set your summary
SUMMARY="Your summary here"

# Create annotated tag with summary
git tag -a "v${VERSION}" -m "Release v${VERSION} - ${SUMMARY}"

# Verify tag was created
git tag -l | grep "v${VERSION}"

# View the tag message
git tag -l -n9 "v${VERSION}"

# Push tag to GitHub
git push origin "v${VERSION}"
```

**Verification**:

- `git tag -l` shows the new tag
- `git tag -l -n9 v${VERSION}` shows the tag with summary
- Check https://github.com/grantcarthew/snag/tags

---

## Step 6: Verify GitHub Source Tarball

GitHub automatically creates a source tarball when you push a tag. Verify it's accessible:

```bash
# Construct tarball URL
TARBALL_URL="https://github.com/grantcarthew/snag/archive/refs/tags/v${VERSION}.tar.gz"
echo "Tarball URL: $TARBALL_URL"

# Verify tarball exists and get SHA256
curl -sL "$TARBALL_URL" | sha256sum

# Or use this one-liner
TARBALL_SHA256=$(curl -sL "https://github.com/grantcarthew/snag/archive/refs/tags/v${VERSION}.tar.gz" | sha256sum | cut -d' ' -f1)
echo "Tarball SHA256: $TARBALL_SHA256"

# Save for next step
echo "$TARBALL_SHA256" > /tmp/snag-tarball-sha256.txt
```

**Verification**:

```bash
# Check tarball downloads successfully
curl -I "https://github.com/grantcarthew/snag/archive/refs/tags/v${VERSION}.tar.gz"

# Expected: HTTP/2 200 OK
```

**Note**: Homebrew builds from this source tarball during `brew install`. No pre-built binaries needed!

---

## Step 7: Update Homebrew Tap

Update the Homebrew formula in `grantcarthew/homebrew-tap`:

```bash
# Navigate to homebrew-tap directory
cd reference/homebrew-tap

# Ensure it's up to date
git pull origin main

# Get the tarball SHA256
TARBALL_URL="https://github.com/grantcarthew/snag/archive/refs/tags/v${VERSION}.tar.gz"
TARBALL_SHA256=$(curl -sL "$TARBALL_URL" | sha256sum | cut -d' ' -f1)

echo "Tarball URL: $TARBALL_URL"
echo "Tarball SHA256: $TARBALL_SHA256"
```

**Update Formula/snag.rb**:

Open `Formula/snag.rb` and update:

1. **url** line: Update to new version
2. **sha256** line: Update with new tarball SHA256
3. **ldflags** line: Update version number
4. **test** line: Update expected version

**Example**:

```ruby
class Snag < Formula
  desc "Intelligently fetch web pages using Chrome via CDP"
  homepage "https://github.com/grantcarthew/snag"
  url "https://github.com/grantcarthew/snag/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "NEW_SHA256_HERE"
  license "MPL-2.0"

  depends_on "go" => :build

  def install
    ENV["CGO_ENABLED"] = "0"
    system "go", "build", *std_go_args(ldflags: "-X main.version=0.1.0", output: bin/"snag")
  end

  test do
    assert_match "0.1.0", shell_output("#{bin}/snag --version")
  end
end
```

**Commit and push**:

```bash
# Stage formula changes
git add Formula/snag.rb

# Commit
git commit -m "snag: update to ${VERSION}"

# Push
git push origin main

# Return to snag directory
cd -
```

**Verification**:

```bash
# Check the commit went through
cd reference/homebrew-tap
git log -1
cd -
```

---

## Step 8: Test Homebrew Installation

Test the updated formula works:

```bash
# Update Homebrew
brew update

# Reinstall snag from tap
brew reinstall grantcarthew/tap/snag

# Verify version
snag --version

# Test basic functionality
snag --quiet example.com

# Expected output: Markdown content from example.com
```

**Verification**:

- [ ] `snag --version` shows new version number
- [ ] `snag` executes without errors
- [ ] Basic page fetch works

**If installation fails**:

1. Check formula syntax: `brew audit --strict grantcarthew/tap/snag`
2. Review formula: `brew cat grantcarthew/tap/snag`
3. Fix issues and push updated formula

---

## Step 9: Post-Release Tasks

Complete these final steps:

### Update README (if needed)

If version-specific information exists in README.md, update it:

```bash
# Check for version references
grep -n "v0\." README.md

# Update any hardcoded version numbers
# Then commit if changes made
git add README.md
git commit -m "docs: update README for v${VERSION}"
git push origin main
```

### Clean Up

```bash
# Remove dist directory
rm -rf dist/

# Verify clean state
git status
```

### Announce Release

Consider announcing the release:

- [ ] GitHub Discussions (if enabled)
- [ ] Twitter/X or other social media
- [ ] Project website or blog
- [ ] Relevant communities or forums

### Monitor for Issues

After release:

- [ ] Watch GitHub issues for bug reports
- [ ] Monitor Homebrew installation feedback
- [ ] Be ready to release a patch if critical bugs found

---

## Rollback Procedure

If critical issues are discovered after release:

### Option 1: Quick Patch Release

```bash
# Fix the issue in code
# Then release a patch version (e.g., v0.1.1)
# Follow this same process
```

### Option 2: Delete Release (Last Resort)

```bash
# Delete GitHub release
gh release delete "v${VERSION}" --yes

# Delete remote tag
git push origin --delete "v${VERSION}"

# Delete local tag
git tag -d "v${VERSION}"

# Revert homebrew-tap
cd reference/homebrew-tap
git revert HEAD
git push origin main
cd -
```

**Only use Option 2 for severe security issues or broken releases.**

---

## Quick Reference

### Version Commands

```bash
# Check current version
git tag -l | tail -1

# Check what version binary reports
./snag --version
```

### Build Commands

```bash
# Quick local build
go build -o snag

# Build with version
VERSION="0.1.0"
go build -ldflags "-X main.version=${VERSION}" -o snag
```

### Release Commands

```bash
# Complete release in one go
export VERSION="0.1.0"

# 1. Update CHANGELOG.md (manually)
# 2. Commit and tag
git add CHANGELOG.md
git commit -m "chore: prepare for v${VERSION} release"
git push origin main

# Create summary from changes
PREV_VERSION=$(git tag -l | tail -1)
git log ${PREV_VERSION}..HEAD --oneline
SUMMARY="Your one-line summary here"
git tag -a "v${VERSION}" -m "Release v${VERSION} - ${SUMMARY}"
git push origin "v${VERSION}"

# 3. Verify GitHub tarball and get SHA256
TARBALL_URL="https://github.com/grantcarthew/snag/archive/refs/tags/v${VERSION}.tar.gz"
TARBALL_SHA256=$(curl -sL "$TARBALL_URL" | sha256sum | cut -d' ' -f1)
echo "Tarball SHA256: $TARBALL_SHA256"

# 4. Update Homebrew tap (manually edit Formula/snag.rb)
cd reference/homebrew-tap
git pull
# Edit Formula/snag.rb
git add Formula/snag.rb
git commit -m "snag: update to ${VERSION}"
git push origin main
cd -

# 5. Test installation
brew update
brew reinstall grantcarthew/tap/snag
snag --version
```

---

## Troubleshooting

### "Tests failing"

- Review test output: `go test -v ./...`
- Fix failing tests before proceeding
- Never release with failing tests

### "Tarball download fails"

- Verify tag is pushed: `git ls-remote --tags origin`
- Check tag exists on GitHub: https://github.com/grantcarthew/snag/tags
- Wait a minute for GitHub to generate tarball

### "Homebrew formula fails audit"

- Run: `brew audit --strict grantcarthew/tap/snag`
- Fix reported issues
- Common issues:
  - Incorrect SHA256
  - Wrong URL format
  - Ruby syntax errors

### "Homebrew installation fails"

- Check formula: `brew cat grantcarthew/tap/snag`
- Try verbose install: `brew install --verbose grantcarthew/tap/snag`
- Verify tarball exists: `curl -I <tarball-url>`

---

## Checklist Template

Copy this for each release:

```markdown
## Release v**\_** Checklist

- [ ] Determined version number: **\_**
- [ ] All tests passing
- [ ] CHANGELOG.md updated (if it exists)
- [ ] Changes committed to main
- [ ] Git tag created with summary
- [ ] Tag pushed to GitHub
- [ ] GitHub tarball verified
- [ ] Tarball SHA256 obtained
- [ ] Homebrew formula updated
- [ ] Formula committed and pushed
- [ ] Homebrew installation tested
- [ ] snag --version shows correct version
- [ ] Basic functionality verified
- [ ] Release announced (optional)
```

---

## Related Documents

- `docs/design-record.md` - Design decisions and rationale
- `docs/projects/PROJECT-release.md` - Original release planning (may be outdated)
- `README.md` - User-facing documentation
- `AGENTS.md` - Repository context for AI agents
- `CHANGELOG.md` - Version history

---

**End of Release Process**
