# PROJECT: Shell Completion Support

**Status:** Not Started
**Priority:** Low (Nice-to-have enhancement)
**Effort:** 2-3 hours
**Start Date:** TBD
**Target Completion:** TBD

## Overview

Add intelligent shell completion support to snag using Cobra's built-in completion framework. This will enable users to tab-complete commands, flags, and values across Bash, Zsh, Fish, and PowerShell.

**Motivation:**
- Improve CLI user experience with tab completion
- Leverage Cobra's built-in completion framework (already available, just needs configuration)
- Provide intelligent completions for flag values (formats, tab patterns)
- Enable dynamic completions (e.g., live browser tabs)
- Zero runtime overhead (completions generated once, cached by shell)

## Success Criteria

### Must Have ✅

1. **Built-in completion command** - `snag completion <shell>` generates completion scripts
2. **Format flag completion** - Tab-complete `--format` with: `md`, `html`, `text`, `pdf`, `png`
3. **File/directory completion** - `--output` completes filenames, `--output-dir` completes directories
4. **Documentation** - README.md has installation instructions for all 4 shells
5. **Testing** - Manual testing in Bash and Zsh confirms completions work
6. **Help text** - `snag completion --help` explains usage for each shell

### Should Have 🎯

1. **Port validation** - `--port` suggests common ports: `9222`, `9223`, `9224`
2. **Boolean flag hints** - Flags like `--format` show descriptions in completion menu
3. **Installation script** - Optional `install-completion.sh` helper script
4. **Homebrew integration** - Completions auto-install with `brew install` (future)

### Could Have 💡

1. **Dynamic tab completion** - `--tab` queries running browser and suggests actual tab numbers/URLs
2. **URL history** - Complete URLs from shell history
3. **Subcommands** - Future-proof for potential subcommand structure (`snag browser`, `snag tabs`)

### Won't Have ❌

1. **Cross-shell compatibility layer** - Each shell uses native completion format
2. **Completion for URL arguments** - Too complex, shells already complete file paths
3. **Custom completions for third-party tools** - Only snag's own flags

## Current State Analysis

### What Cobra Provides Out of the Box

Cobra **automatically** provides:
- `snag completion bash` - Generates Bash completion script
- `snag completion zsh` - Generates Zsh completion script
- `snag completion fish` - Generates Fish completion script
- `snag completion powershell` - Generates PowerShell completion script

**Zero code required** - this already works in current implementation!

### What Needs Custom Implementation

**Flag value completions** (main.go:108-138 flag definitions):
```go
// These flags need intelligent completions:
--format      → md, html, text, pdf, png
--output      → <filename completion>
--output-dir  → <directory completion>
--port        → 9222, 9223, 9224, 9225
--tab         → <dynamic: query browser tabs>
```

**Why this matters:**
```bash
# Without completion:
$ snag --format <TAB>
# (no suggestions)

# With completion:
$ snag --format <TAB>
md    html    text    pdf    png
```

## Technical Approach

### Implementation Strategy

**Phase 1: Basic Completions** (1 hour)
- Mark file/directory flags for path completion
- Add static completions for `--format` flag
- Add static completions for `--port` flag
- Test in Bash and Zsh

**Phase 2: Documentation** (1 hour)
- Add completion section to README.md
- Document installation for all 4 shells
- Add examples and troubleshooting
- Update help text

**Phase 3: Advanced Features** (Optional, 2-3 hours)
- Dynamic tab completion (query browser)
- Installation helper script
- Homebrew formula integration

### Code Changes Required

**File:** `main.go`

**Location:** Add new `initCompletion()` function after `init()` (line ~138)

```go
// initCompletion configures shell completion behavior
func initCompletion() {
	// File/directory completions
	rootCmd.MarkFlagFilename("output", "md", "html", "txt", "pdf", "png")
	rootCmd.MarkFlagDirname("output-dir")
	rootCmd.MarkFlagFilename("url-file", "txt")
	rootCmd.MarkFlagDirname("user-data-dir")

	// Format flag completion (static list with descriptions)
	rootCmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"md\tMarkdown format (default)",
			"html\tHTML format",
			"text\tPlain text format",
			"pdf\tPDF document",
			"png\tPNG screenshot",
		}, cobra.ShellCompDirectiveNoFileComp
	})

	// Port flag completion (common debugging ports)
	rootCmd.RegisterFlagCompletionFunc("port", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"9222\tDefault Chrome DevTools port",
			"9223\tAlternate port 1",
			"9224\tAlternate port 2",
			"9225\tAlternate port 3",
		}, cobra.ShellCompDirectiveNoFileComp
	})

	// Tab flag completion (basic - Phase 3 could make dynamic)
	rootCmd.RegisterFlagCompletionFunc("tab", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Phase 1: Static suggestions
		return []string{
			"1\tFirst tab",
			"2\tSecond tab",
			"3\tThird tab",
		}, cobra.ShellCompDirectiveNoFileComp

		// Phase 3: Dynamic (query actual browser tabs)
		// bm, err := connectToExistingBrowser(port)
		// if err == nil {
		//     tabs, _ := bm.ListTabs()
		//     suggestions := make([]string, len(tabs))
		//     for i, tab := range tabs {
		//         suggestions[i] = fmt.Sprintf("%d\t%s", tab.Index, tab.URL)
		//     }
		//     return suggestions, cobra.ShellCompDirectiveNoFileComp
		// }
	})

	// Disable file completion for boolean flags
	rootCmd.Flags().SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		return pflag.NormalizedName(name)
	})
}
```

**Call from `init()`:**
```go
func init() {
	// ... existing flag definitions (lines 110-138) ...

	// Set custom help template
	rootCmd.SetHelpTemplate(cobraHelpTemplate)

	// Configure shell completions
	initCompletion()  // ← ADD THIS LINE
}
```

**Customize completion command help:**
```go
// Add to rootCmd definition (around line 140)
var rootCmd = &cobra.Command{
	Use:     "snag [options] URL...",
	Short:   "Intelligently fetch web page content using a browser engine",
	Version: version,
	Args:    cobra.ArbitraryArgs,
	RunE:    runCobra,

	// Custom completion message
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd:   false,
		DisableNoDescFlag:   false,
		DisableDescriptions: false,
		HiddenDefaultCmd:    false,
	},
}
```

### Files to Modify

1. **main.go** (~50 lines added)
   - Add `initCompletion()` function
   - Call from `init()`
   - Add completion examples to help template (optional)

2. **README.md** (~100 lines added)
   - New section: "Shell Completion"
   - Installation instructions for 4 shells
   - Examples and troubleshooting

3. **AGENTS.md** (~20 lines added)
   - Update "Build and Test Commands" section
   - Add completion testing commands

4. **docs/shell-completion.md** (new file, ~200 lines)
   - Detailed completion guide
   - Advanced usage examples
   - Dynamic tab completion explanation

## Installation Instructions (for README.md)

### Bash

```bash
# Generate completion script
snag completion bash > /usr/local/etc/bash_completion.d/snag

# Or for user-only install
mkdir -p ~/.local/share/bash-completion/completions
snag completion bash > ~/.local/share/bash-completion/completions/snag

# Reload shell
source ~/.bashrc
```

### Zsh

```bash
# Generate completion script
snag completion zsh > "${fpath[1]}/_snag"

# Or add to a fpath directory
mkdir -p ~/.zsh/completions
snag completion zsh > ~/.zsh/completions/_snag
echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
echo 'autoload -Uz compinit && compinit' >> ~/.zshrc

# Reload shell
source ~/.zshrc
```

### Fish

```bash
# Generate completion script
snag completion fish > ~/.config/fish/completions/snag.fish

# Reload completions
fish_update_completions
```

### PowerShell

```powershell
# Generate completion script
snag completion powershell | Out-String | Invoke-Expression

# Or save to profile
snag completion powershell >> $PROFILE
```

### Homebrew (Future)

When snag is distributed via Homebrew, completions will auto-install:

```bash
brew install snag
# Completions automatically installed to:
# - Bash: $(brew --prefix)/etc/bash_completion.d/snag
# - Zsh: $(brew --prefix)/share/zsh/site-functions/_snag
# - Fish: $(brew --prefix)/share/fish/vendor_completions.d/snag.fish
```

## Testing Strategy

### Manual Testing Checklist

**Bash:**
```bash
# 1. Install completions
snag completion bash > /tmp/snag-completion.bash
source /tmp/snag-completion.bash

# 2. Test flag completion
snag --format <TAB>        # Should show: md html text pdf png
snag --port <TAB>          # Should show: 9222 9223 9224 9225
snag --output <TAB>        # Should complete filenames
snag --output-dir <TAB>    # Should complete directories

# 3. Test flag name completion
snag --fo<TAB>             # Should complete to --format
snag --ver<TAB>            # Should complete to --verbose or --version

# 4. Test subcommand completion
snag completion <TAB>      # Should show: bash zsh fish powershell
```

**Zsh:**
```zsh
# Same tests as Bash
# Verify descriptions appear in completion menu
snag --format <TAB>        # Should show descriptions like "Markdown format (default)"
```

**Fish:**
```fish
# Same tests as Bash
# Fish has inline descriptions
snag --format <TAB>        # Descriptions appear inline
```

### Automated Testing

**No automated tests required** - shell completion is user-facing and shell-dependent.

**Verification:** Manual testing in 2+ shells (Bash + Zsh) is sufficient.

## Documentation Updates

### README.md Changes

**Add new section after "Installation":**

```markdown
## Shell Completion

snag supports tab completion for Bash, Zsh, Fish, and PowerShell.

### Quick Install

#### Bash
```bash
snag completion bash > /usr/local/etc/bash_completion.d/snag
source ~/.bashrc
```

#### Zsh
```bash
snag completion zsh > "${fpath[1]}/_snag"
source ~/.zshrc
```

#### Fish
```bash
snag completion fish > ~/.config/fish/completions/snag.fish
```

### What Gets Completed

- **Flags**: `--format`, `--port`, `--output`, etc.
- **Values**: Format options (md/html/text/pdf/png), common ports
- **Files**: `--output` completes filenames, `--output-dir` completes directories

### Examples

```bash
snag --format <TAB>        # → md  html  text  pdf  png
snag --port <TAB>          # → 9222  9223  9224  9225
snag --output report.<TAB> # → report.md  report.html  report.pdf
```

See `snag completion --help` for detailed installation instructions.
```

### AGENTS.md Changes

**Add to "Build and Test Commands" section:**

```bash
# Test shell completion
snag completion bash > /tmp/snag-completion.bash
source /tmp/snag-completion.bash
snag --format <TAB>  # Test completion works

# Generate all completion scripts
snag completion bash > completions/snag.bash
snag completion zsh > completions/_snag
snag completion fish > completions/snag.fish
snag completion powershell > completions/snag.ps1
```

## Advanced Features (Phase 3 - Optional)

### Dynamic Tab Completion

Query the actual browser tabs and provide real suggestions:

```go
rootCmd.RegisterFlagCompletionFunc("tab", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Try to connect to existing browser
	bm, err := connectToExistingBrowser(port)
	if err != nil {
		// Fallback to static suggestions
		return []string{"1", "2", "3"}, cobra.ShellCompDirectiveNoFileComp
	}
	defer bm.Close()

	tabs, err := bm.ListTabs()
	if err != nil || len(tabs) == 0 {
		return []string{"1", "2", "3"}, cobra.ShellCompDirectiveNoFileComp
	}

	// Return actual tab suggestions
	suggestions := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		// Format: "1\thttps://github.com - GitHub"
		desc := fmt.Sprintf("%s", tab.URL)
		if tab.Title != "" && tab.Title != "New Tab" {
			desc = fmt.Sprintf("%s - %s", tab.URL, tab.Title)
		}
		suggestions = append(suggestions, fmt.Sprintf("%d\t%s", tab.Index, desc))
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
})
```

**Example:**
```bash
$ snag --open-browser
$ snag --tab <TAB>
1  https://github.com - GitHub
2  https://google.com - Google Search
3  https://example.com - Example Domain
```

**Challenges:**
- Performance: Each tab completion triggers browser connection (~200ms)
- Error handling: Browser might not be running
- Security: Exposing browser tabs in shell history

**Recommendation:** Implement in Phase 3 if users request it.

### Installation Helper Script

Create `install-completion.sh`:

```bash
#!/usr/bin/env bash
set -e

# Detect shell
SHELL_NAME=$(basename "$SHELL")

case "$SHELL_NAME" in
  bash)
    echo "Installing Bash completion..."
    snag completion bash > "${BASH_COMPLETION_USER_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/bash-completion}/completions/snag"
    echo "✓ Installed. Restart your shell or run: source ~/.bashrc"
    ;;
  zsh)
    echo "Installing Zsh completion..."
    mkdir -p ~/.zsh/completions
    snag completion zsh > ~/.zsh/completions/_snag
    echo "✓ Installed. Add to ~/.zshrc:"
    echo "  fpath=(~/.zsh/completions \$fpath)"
    echo "  autoload -Uz compinit && compinit"
    ;;
  fish)
    echo "Installing Fish completion..."
    mkdir -p ~/.config/fish/completions
    snag completion fish > ~/.config/fish/completions/snag.fish
    echo "✓ Installed. Restart your shell."
    ;;
  *)
    echo "Unknown shell: $SHELL_NAME"
    echo "Run manually:"
    echo "  snag completion bash|zsh|fish|powershell"
    exit 1
    ;;
esac
```

## Risks and Considerations

### Risks

1. **Shell compatibility** - Different shell versions behave differently
   - **Mitigation:** Test on macOS Bash 3.2, Bash 5.x, Zsh 5.8+

2. **Performance of dynamic completions** - Browser queries slow down tab completion
   - **Mitigation:** Cache completions, timeout after 100ms

3. **Installation complexity** - Users might not understand completion setup
   - **Mitigation:** Provide install script, clear docs

### Breaking Changes

**None** - This is purely additive functionality.

Existing users won't be affected. Completion is opt-in (users must install it).

## Implementation Timeline

### Phase 1: Basic Completions (1-2 hours)

**Tasks:**
1. Add `initCompletion()` function to main.go (30 min)
2. Implement static completions for `--format`, `--port` (30 min)
3. Mark file/directory flags (10 min)
4. Test in Bash and Zsh (30 min)

**Deliverable:** Working completions for format and port flags

### Phase 2: Documentation (1 hour)

**Tasks:**
1. Add "Shell Completion" section to README.md (30 min)
2. Update AGENTS.md with completion testing (10 min)
3. Test installation instructions on clean system (20 min)

**Deliverable:** Complete user-facing documentation

### Phase 3: Advanced Features (Optional, 2-3 hours)

**Tasks:**
1. Implement dynamic tab completion (1 hour)
2. Create `install-completion.sh` helper (30 min)
3. Performance testing and optimization (1 hour)

**Deliverable:** Dynamic tab suggestions, one-command installation

## Success Metrics

### Quantitative

- ✅ All 4 shell completion scripts generate without errors
- ✅ `--format` completion shows 5 options (md/html/text/pdf/png)
- ✅ `--port` completion shows 4 common ports
- ✅ `--output` completes filenames, `--output-dir` completes directories
- ✅ Completion works in at least 2 shells (Bash + Zsh)

### Qualitative

- ✅ Users can install completions in < 2 minutes
- ✅ Tab completion feels responsive (< 100ms)
- ✅ Documentation is clear and includes examples
- ✅ No breaking changes to existing behavior

## Dependencies

**Code:**
- `github.com/spf13/cobra` v1.10.1 (already installed ✓)
- `github.com/spf13/pflag` v1.0.9 (already installed ✓)

**Testing:**
- macOS with Bash 3.2+ or Bash 5.x
- Zsh 5.8+ (included in macOS)
- Fish 3.x (optional: `brew install fish`)

**Documentation:**
- None (standard markdown)

## Future Enhancements

1. **Homebrew integration** - Auto-install completions with `brew install snag`
2. **Completion caching** - Cache browser tab list for faster completions
3. **URL history completion** - Complete URLs from `.snag_history` file
4. **Subcommand support** - Future-proof for `snag browser`, `snag tabs`, etc.
5. **Man page generation** - Cobra can generate man pages from commands

## References

- [Cobra Completion Documentation](https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md)
- [Bash Completion Guide](https://github.com/scop/bash-completion)
- [Zsh Completion System](https://zsh.sourceforge.io/Doc/Release/Completion-System.html)
- [Fish Completion Tutorial](https://fishshell.com/docs/current/completions.html)

## Notes

- Cobra's completion framework is mature and well-tested (used by kubectl, gh, docker)
- Shell completion is a **quality-of-life feature**, not critical functionality
- Installation is one-time per user, zero runtime cost
- Dynamic tab completion is impressive but optional (Phase 3)

---

**Project Status:** Ready for implementation whenever desired. Low priority, high user satisfaction impact.
