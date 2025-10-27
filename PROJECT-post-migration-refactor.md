# PROJECT: Post-Migration Refactoring (Cobra CLI Cleanup)

**Status:** Not Started
**Priority:** High
**Effort:** 8-12 hours
**Start Date:** TBD
**Completion Date:** TBD

## Overview

Following the successful migration from urfave/cli to Cobra, a comprehensive code review identified several technical debt items and improvement opportunities. This project addresses critical issues, architectural improvements, code quality enhancements, and documentation gaps introduced or exposed during the migration.

While the migration is functional (all 124 tests passing), these refactorings will improve maintainability, testability, and adherence to Go best practices.

## Success Criteria

- ✅ All critical issues (Priority 1) resolved
- ✅ All high-priority issues (Priority 2) resolved
- ✅ All 124 tests continue to pass
- ✅ No behavioral changes for users
- ✅ Code coverage remains ≥ current level
- ✅ Go vet and golint pass with no warnings
- ✅ Binary size remains ≤ 20 MB

## Phase 1: Critical Fixes (Priority 1) - BLOCKING

**Goal:** Resolve critical architectural issues that could cause bugs or maintenance problems.

### Task 1.1: Replace Global Mutable State with Dependency Injection

**Location:** `main.go:31-33`

**Current Code:**
```go
var (
	logger         *Logger
	browserManager *BrowserManager
)
```

**Problem:** Global mutable state creates:
- Race conditions in concurrent scenarios
- Difficult-to-test code
- Hidden dependencies
- State leakage between operations

**Solution:** Pass logger and browserManager as parameters through call chain.

**Implementation Steps:**

1. Create a `Runtime` struct to hold shared state:
   ```go
   type Runtime struct {
       logger         *Logger
       browserManager *BrowserManager
   }
   ```

2. Update all handler signatures to accept `*Runtime`:
   ```go
   func handleListTabs(rt *Runtime, cmd *cobra.Command) error
   func handleAllTabs(rt *Runtime, cmd *cobra.Command) error
   func handleTabFetch(rt *Runtime, cmd *cobra.Command) error
   // ... etc
   ```

3. Update `runCobra` to create and pass Runtime:
   ```go
   func runCobra(cmd *cobra.Command, args []string) error {
       rt := &Runtime{
           logger: NewLogger(level),
       }
       // Pass rt to all handlers
   }
   ```

4. Update `snag()` function signature:
   ```go
   func snag(rt *Runtime, config *Config) error
   ```

5. Update signal handler to use captured runtime:
   ```go
   go func() {
       sig := <-sigChan
       fmt.Fprintf(os.Stderr, "\nReceived %v, cleaning up...\n", sig)
       if rt.browserManager != nil {
           rt.browserManager.Close()
       }
       // ...
   }()
   ```

**Testing:**
- Verify all 124 tests still pass
- Add test for concurrent snag operations (if applicable)

**Files Modified:**
- `main.go`
- `handlers.go`

---

### Task 1.2: Remove Manual os.Args Parsing for Log Level

**Location:** `main.go:174-193`

**Current Code:**
```go
lastLogFlag := ""
lastLogIndex := -1
for i, arg := range os.Args {
	if arg == "--quiet" || arg == "-q" {
		if i > lastLogIndex {
			lastLogIndex = i
			lastLogFlag = "quiet"
		}
	}
	// ... more manual parsing
}
```

**Problem:**
- Bypasses Cobra's flag parsing mechanism
- Doesn't handle edge cases (e.g., `--quiet=true`, combined flags)
- May not respect Cobra's flag precedence rules
- Duplicates work that Cobra already does

**Solution:** Use Cobra's `cmd.Flags().Changed()` mechanism.

**Implementation Steps:**

1. Add a custom flag priority tracker:
   ```go
   type LogLevel int

   const (
       LevelNormal LogLevel = iota
       LevelQuiet
       LevelVerbose
       LevelDebug
   )

   func determineLogLevel(cmd *cobra.Command) LogLevel {
       // Check which flags were explicitly set
       quietSet := cmd.Flags().Changed("quiet")
       verboseSet := cmd.Flags().Changed("verbose")
       debugSet := cmd.Flags().Changed("debug")

       // Priority: debug > verbose > quiet > normal
       if debugSet && debug {
           return LevelDebug
       }
       if verboseSet && verbose {
           return LevelVerbose
       }
       if quietSet && quiet {
           return LevelQuiet
       }
       return LevelNormal
   }
   ```

2. Update `runCobra` to use new function:
   ```go
   func runCobra(cmd *cobra.Command, args []string) error {
       level := determineLogLevel(cmd)
       logger := NewLogger(level)
       // ...
   }
   ```

3. Add mutually exclusive flag group (Cobra 1.8+):
   ```go
   rootCmd.MarkFlagsMutuallyExclusive("quiet", "verbose", "debug")
   ```

**Testing:**
- Test `--quiet`
- Test `--verbose`
- Test `--debug`
- Test `--quiet --verbose` (should error or use last one)
- Test `-q -v` (short forms)

**Files Modified:**
- `main.go`

---

## Phase 2: Architecture Improvements (Priority 2)

**Goal:** Improve error handling, validation ordering, and add context support.

### Task 2.1: Move Validation Before Expensive Operations

**Location:** Multiple handlers in `handlers.go`

**Problem:** Browser connections happen before validating format, timeout, etc.

**Example (handlers.go:281-291):**
```go
bm, err := connectToExistingBrowser(port)  // Expensive operation
if err != nil {
    return err
}

if err := validateFormat(outputFormat); err != nil {  // Should be first!
    return err
}
```

**Solution:** Move all validation to the top of each handler function.

**Implementation Steps:**

1. Create a validation helper for each handler:
   ```go
   func validateAllTabsConfig(format string, timeout int, outDir string) error {
       if err := validateFormat(format); err != nil {
           return err
       }
       if err := validateTimeout(timeout); err != nil {
           return err
       }
       if err := validateDirectory(outDir); err != nil {
           return err
       }
       return nil
   }
   ```

2. Update each handler to validate first:
   ```go
   func handleAllTabs(rt *Runtime, cmd *cobra.Command) error {
       outputFormat := normalizeFormat(format)
       outDir := strings.TrimSpace(outputDir)
       if outDir == "" {
           outDir = "."
       }

       // VALIDATE FIRST - before any expensive operations
       if err := validateAllTabsConfig(outputFormat, timeout, outDir); err != nil {
           return err
       }

       // Now connect to browser
       bm, err := connectToExistingBrowser(port)
       // ...
   }
   ```

3. Apply pattern to all handlers:
   - `handleAllTabs`
   - `handleTabFetch`
   - `handleTabRange`
   - `handleTabPatternBatch`
   - `handleMultipleURLs`
   - `handleOpenURLsInBrowser`

**Testing:**
- Test that validation errors don't create browser connections
- Verify all validation tests still pass

**Files Modified:**
- `handlers.go`

---

### Task 2.2: Add Context Support for Cancellation

**Location:** All handler functions

**Problem:** No context support means:
- Can't cancel long-running operations
- No timeout control beyond page-level timeouts
- No graceful shutdown for batch operations

**Solution:** Add context.Context throughout the call chain.

**Implementation Steps:**

1. Update Cobra command to use context:
   ```go
   func runCobra(cmd *cobra.Command, args []string) error {
       ctx := cmd.Context()
       rt := &Runtime{
           ctx:    ctx,
           logger: NewLogger(level),
       }
       // ...
   }
   ```

2. Add context to Runtime struct:
   ```go
   type Runtime struct {
       ctx            context.Context
       logger         *Logger
       browserManager *BrowserManager
   }
   ```

3. Update handler signatures:
   ```go
   func handleAllTabs(ctx context.Context, rt *Runtime, cmd *cobra.Command) error
   ```

4. Pass context to browser operations:
   ```go
   page, err := bm.NewPageWithContext(ctx)
   ```

5. Add context checks in batch operations:
   ```go
   for i, tab := range tabs {
       select {
       case <-ctx.Done():
           return ctx.Err()
       default:
           // Process tab
       }
   }
   ```

6. Setup context with signal handling:
   ```go
   func main() {
       ctx, cancel := context.WithCancel(context.Background())
       defer cancel()

       sigChan := make(chan os.Signal, 1)
       signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

       go func() {
           <-sigChan
           cancel() // Cancel context on signal
       }()

       rootCmd.SetContext(ctx)
       if err := rootCmd.Execute(); err != nil {
           // ...
       }
   }
   ```

**Testing:**
- Test Ctrl+C during batch operations
- Test context timeout
- Verify graceful cleanup on cancellation

**Files Modified:**
- `main.go`
- `handlers.go`
- `browser.go` (if adding context support to browser operations)

---

### Task 2.3: Consistent Error Handling Pattern

**Location:** Throughout `handlers.go`

**Problem:** Inconsistent error handling - some errors logged before return, others not.

**Solution:** Establish and enforce consistent pattern.

**Implementation Steps:**

1. Create error handling helpers:
   ```go
   // logAndReturn logs an error and returns it
   func logAndReturn(logger *Logger, format string, args ...interface{}) error {
       err := fmt.Errorf(format, args...)
       logger.Error("%v", err)
       return err
   }

   // wrapAndLog wraps an error with context and logs it
   func wrapAndLog(logger *Logger, err error, context string) error {
       wrapped := fmt.Errorf("%s: %w", context, err)
       logger.Error("%v", wrapped)
       return wrapped
   }
   ```

2. Establish pattern rule:
   - **Browser/network errors**: Log with details, return wrapped error
   - **Validation errors**: Return directly (caller logs if needed)
   - **Configuration errors**: Log with suggestion, return error

3. Apply pattern consistently across all handlers:
   ```go
   bm, err := connectToExistingBrowser(port)
   if err != nil {
       return wrapAndLog(logger, err, "failed to connect to browser")
   }

   if err := validateFormat(outputFormat); err != nil {
       return err // Validation errors don't need logging here
   }
   ```

4. Document the pattern in handlers.go header comment:
   ```go
   // Error Handling Pattern:
   // - Browser/network errors: Log details and return wrapped error
   // - Validation errors: Return directly without logging
   // - Configuration errors: Log with user-facing suggestion
   ```

**Testing:**
- Verify error messages appear consistently in tests
- Check that stderr output makes sense

**Files Modified:**
- `handlers.go`

---

### Task 2.4: Fix Inconsistent Cleanup Pattern

**Location:** `handlers.go:838-848`

**Problem:** Cleanup is manual on error path, deferred on success path.

**Current Code:**
```go
_, err := bm.Connect()
if err != nil {
    browserManager = nil  // Manual cleanup
    return err
}
defer func() {  // Deferred cleanup
    bm.Close()
    browserManager = nil
}()
```

**Solution:** Always use deferred cleanup.

**Implementation Steps:**

1. Update pattern to always defer:
   ```go
   bm := NewBrowserManager(...)
   rt.browserManager = bm

   defer func() {
       if bm != nil {
           bm.Close()
       }
       rt.browserManager = nil
   }()

   _, err := bm.Connect()
   if err != nil {
       return err  // Defer will still run
   }
   ```

2. Apply to all handlers with browser connections:
   - `snag()`
   - `handleAllTabs()`
   - `handleTabFetch()`
   - `handleOpenURLsInBrowser()`
   - `handleMultipleURLs()`

**Testing:**
- Test cleanup happens on error paths
- Test cleanup happens on success paths
- Verify no orphaned browsers after errors

**Files Modified:**
- `handlers.go`

---

## Phase 3: Code Quality (Priority 3)

**Goal:** Reduce code duplication, improve maintainability, add comprehensive validation.

### Task 3.1: Extract Duplicate Batch Processing Code

**Location:** `handlers.go:522-606` (handleTabRange) and `handlers.go:608-682` (handleTabPatternBatch)

**Problem:** 80+ lines of nearly identical batch processing logic duplicated.

**Solution:** Extract common batch processing function.

**Implementation Steps:**

1. Create batch configuration struct:
   ```go
   type BatchConfig struct {
       Format      string
       WaitFor     string
       Timeout     int
       OutputDir   string
       CloseTab    bool
       Timestamp   time.Time
   }
   ```

2. Create batch processor function:
   ```go
   // processBatchTabs processes multiple tabs with common logic
   func processBatchTabs(
       rt *Runtime,
       pages []*rod.Page,
       config BatchConfig,
       logPrefix func(current, total int) string,
   ) error {
       successCount := 0
       failureCount := 0

       for i, page := range pages {
           current := i + 1
           total := len(pages)

           prefix := logPrefix(current, total)

           info, err := page.Info()
           if err != nil {
               rt.logger.Error("%s Failed to get tab info: %v", prefix, err)
               failureCount++
               continue
           }

           rt.logger.Info("%s Processing: %s", prefix, info.URL)

           if config.WaitFor != "" {
               err := waitForSelector(page, config.WaitFor,
                   time.Duration(config.Timeout)*time.Second)
               if err != nil {
                   rt.logger.Error("%s Wait failed: %v", prefix, err)
                   failureCount++
                   continue
               }
           }

           outputPath, err := generateOutputFilename(
               info.Title, info.URL, config.Format,
               config.Timestamp, config.OutputDir,
           )
           if err != nil {
               rt.logger.Error("%s Failed to generate filename: %v", prefix, err)
               failureCount++
               continue
           }

           if err := processPageContent(page, config.Format, outputPath); err != nil {
               rt.logger.Error("%s Failed to process content: %v", prefix, err)
               failureCount++
               continue
           }

           successCount++
       }

       rt.logger.Success("Batch complete: %d succeeded, %d failed",
           successCount, failureCount)

       if failureCount > 0 {
           return fmt.Errorf("batch processing completed with %d failures",
               failureCount)
       }

       return nil
   }
   ```

3. Update `handleTabRange` to use common function:
   ```go
   func handleTabRange(rt *Runtime, cmd *cobra.Command, bm *BrowserManager, start, end int) error {
       // ... validation ...

       pages, err := bm.GetTabsByRange(start, end)
       if err != nil {
           // ... error handling ...
       }

       rt.logger.Info("Processing %d tabs from range [%d-%d]...", len(pages), start, end)

       config := BatchConfig{
           Format:    outputFormat,
           WaitFor:   validatedWaitFor,
           Timeout:   timeout,
           OutputDir: outDir,
           Timestamp: time.Now(),
       }

       logPrefix := func(current, total int) string {
           tabNum := start + current - 1
           return fmt.Sprintf("[%d/%d] Tab [%d]:", current, total, tabNum)
       }

       return processBatchTabs(rt, pages, config, logPrefix)
   }
   ```

4. Update `handleTabPatternBatch` similarly

5. Update `handleAllTabs` to use common function

**Testing:**
- Verify all batch operations produce same output as before
- Test error handling in batch operations

**Files Modified:**
- `handlers.go`

---

### Task 3.2: Replace Magic Numbers with Constants

**Location:** Multiple files

**Problem:** Magic numbers scattered throughout code reduce maintainability.

**Implementation Steps:**

1. Define constants at package level:
   ```go
   // Exit codes
   const (
       ExitCodeSuccess   = 0
       ExitCodeError     = 1
       ExitCodeInterrupt = 130  // 128 + SIGINT (2)
       ExitCodeSIGTERM   = 143  // 128 + SIGTERM (15)
   )

   // Display formatting
   const (
       MaxDisplayURLLength = 80
       MaxTabLineLength    = 120
   )

   // Batch processing
   const (
       DefaultTimeout = 30 // seconds
   )
   ```

2. Replace magic numbers throughout codebase:

   **main.go:**
   ```go
   // Before:
   os.Exit(130)
   os.Exit(143)

   // After:
   os.Exit(ExitCodeInterrupt)
   os.Exit(ExitCodeSIGTERM)
   ```

   **handlers.go:**
   ```go
   // Before:
   const maxURLLen = 80
   // ... line 233 ...
   line := formatTabLine(tab.Index, tab.Title, tab.URL, 120, verbose)

   // After:
   const maxURLLen = MaxDisplayURLLength
   // ... line 233 ...
   line := formatTabLine(tab.Index, tab.Title, tab.URL, MaxTabLineLength, verbose)
   ```

3. Document constants with comments:
   ```go
   // ExitCodeInterrupt is returned when process receives SIGINT (Ctrl+C)
   ExitCodeInterrupt = 130
   ```

**Testing:**
- Verify exit codes unchanged
- Verify display output unchanged

**Files Modified:**
- `main.go`
- `handlers.go`

---

### Task 3.3: Comprehensive Flag Combination Validation

**Location:** `main.go:runCobra`

**Problem:** Some invalid flag combinations caught, others only warned about.

**Solution:** Create comprehensive validation function.

**Implementation Steps:**

1. Create validation function:
   ```go
   // validateFlagCombinations checks for invalid flag combinations
   func validateFlagCombinations(cmd *cobra.Command) error {
       // Conflicting browser modes
       if openBrowser && forceHead {
           return fmt.Errorf("cannot use both --force-headless and --open-browser")
       }

       // Content source conflicts
       var contentSources []string
       if cmd.Flags().Changed("tab") {
           contentSources = append(contentSources, "--tab")
       }
       if allTabs {
           contentSources = append(contentSources, "--all-tabs")
       }
       if listTabs {
           contentSources = append(contentSources, "--list-tabs")
       }

       if len(contentSources) > 1 {
           return fmt.Errorf("cannot use multiple content sources: %s",
               strings.Join(contentSources, ", "))
       }

       // Tab features require existing browser (not --force-headless)
       if forceHead && (cmd.Flags().Changed("tab") || allTabs) {
           return fmt.Errorf("cannot use --force-headless with tab features")
       }

       // Output conflicts
       if output != "" && outputDir != "" {
           return fmt.Errorf("cannot use both --output and --output-dir")
       }

       // Multiple URL conflicts
       if output != "" && (allTabs || cmd.Flags().Changed("tab")) {
           return fmt.Errorf("cannot use --output with multiple content sources, use --output-dir")
       }

       return nil
   }
   ```

2. Call early in `runCobra`:
   ```go
   func runCobra(cmd *cobra.Command, args []string) error {
       // Validate flag combinations first
       if err := validateFlagCombinations(cmd); err != nil {
           return err
       }

       // ... rest of function
   }
   ```

3. Remove duplicate validation checks from throughout the code

**Testing:**
- Test all invalid flag combinations
- Verify helpful error messages

**Files Modified:**
- `main.go`

---

### Task 3.4: Move rootCmd Initialization to init()

**Location:** `main.go:136-142`

**Problem:** Package-level variable defined before init() function - Go style prefers init() for setup.

**Solution:** Move command definition into init().

**Implementation Steps:**

1. Declare variable without initialization:
   ```go
   var rootCmd *cobra.Command
   ```

2. Move initialization into init():
   ```go
   func init() {
       rootCmd = &cobra.Command{
           Use:     "snag [options] URL...",
           Short:   "Intelligently fetch web page content using a browser engine",
           Version: version,
           Args:    cobra.ArbitraryArgs,
           RunE:    runCobra,
       }

       // Set custom help template
       rootCmd.SetHelpTemplate(cobraHelpTemplate)

       // String flags
       rootCmd.Flags().StringVar(&urlFile, "url-file", "", "...")
       // ... all other flags ...
   }
   ```

**Testing:**
- Verify --help output unchanged
- Verify all flags work

**Files Modified:**
- `main.go`

---

### Task 3.5: Move Help Template to Separate Function

**Location:** `main.go:58-102`

**Problem:** 45-line string literal reduces readability.

**Solution:** Extract to function or variable.

**Implementation Steps:**

1. Create function:
   ```go
   // getHelpTemplate returns the custom Cobra help template
   func getHelpTemplate() string {
       return `USAGE:
     {{.UseLine}}{{if .HasAvailableSubCommands}}
     {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

   ALIASES:
     {{.NameAndAliases}}{{end}}{{if .HasExample}}

   EXAMPLES:
   {{.Example}}{{end}}

   DESCRIPTION:
     snag fetches web page content using Chromium/Chrome automation.
     It can connect to existing browser sessions, launch headless browsers, or open
     visible browsers for authenticated sessions.

     Output formats: Markdown, HTML, text, PDF, or PNG.

     The perfect companion for AI agents to gain context from web pages.

   AGENT USAGE:
     Common workflows:
     • Single page: snag example.com
     • Multiple pages: snag -d output/ url1 url2 url3
     • Authenticated pages: snag --open-browser (authenticate), then snag -t <pattern>
     • All browser tabs: snag --all-tabs -d output/

     Integration:
     • Content → stdout, logs → stderr (pipe-safe)
     • Non-zero exit on errors
     • Auto-names files with timestamps

     Performance: 2-5 seconds per page. Tab reuse is faster.
   {{if .HasAvailableLocalFlags}}

   GLOBAL OPTIONS:
   {{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

   GLOBAL OPTIONS:
   {{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

   ADDITIONAL HELP TOPICS:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
     {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

   Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
   `
   }
   ```

2. Use in init():
   ```go
   func init() {
       rootCmd = &cobra.Command{...}
       rootCmd.SetHelpTemplate(getHelpTemplate())
       // ...
   }
   ```

**Testing:**
- Verify --help output unchanged

**Files Modified:**
- `main.go`

---

### Task 3.6: Standardize Error Message Format

**Location:** Throughout `handlers.go`

**Problem:** Inconsistent error message formats.

**Solution:** Define and enforce standard formats.

**Implementation Steps:**

1. Create error message templates:
   ```go
   // Error message templates
   const (
       errCannotUseBoth     = "cannot use both %s and %s"
       errCannotUseWith     = "cannot use %s with %s"
       errRequiresFlag      = "%s requires %s flag"
       errMutuallyExclusive = "%s and %s are mutually exclusive"
   )
   ```

2. Create error helper functions:
   ```go
   func errFlagConflict(flag1, flag2 string) error {
       return fmt.Errorf(errCannotUseBoth, flag1, flag2)
   }

   func errFlagRequires(flag, required string) error {
       return fmt.Errorf(errRequiresFlag, flag, required)
   }
   ```

3. Update all error messages to use helpers:
   ```go
   // Before:
   return fmt.Errorf("conflicting flags: --force-headless and --open-browser")

   // After:
   return errFlagConflict("--force-headless", "--open-browser")
   ```

**Testing:**
- Verify error messages are clear and consistent
- Test all error paths

**Files Modified:**
- `handlers.go`
- `main.go`

---

### Task 3.7: Use cmd.Flags().Changed() Consistently

**Location:** Throughout `main.go` and `handlers.go`

**Problem:** Some flags check `.Changed()`, others check value directly.

**Solution:** Establish consistent pattern.

**Implementation Steps:**

1. Document the pattern:
   ```go
   // Flag checking pattern:
   // - Use cmd.Flags().Changed() when distinguishing between:
   //   1. User explicitly set flag to default value
   //   2. User didn't set flag at all (using default)
   // - Direct value check when only caring about final value
   ```

2. Apply pattern consistently:
   - **Configuration flags**: Use `.Changed()` (format, timeout, etc.)
   - **Boolean mode flags**: Direct check is fine (closeTab, verbose, etc.)
   - **Feature selection flags**: Use `.Changed()` (tab, output, etc.)

3. Update all flag checks:
   ```go
   // Configuration flags - use Changed()
   if cmd.Flags().Changed("format") {
       logger.Warning("--format ignored with --open-browser")
   }

   // Boolean mode flags - direct check OK
   if closeTab {
       logger.Warning("--close-tab ignored with --open-browser")
   }

   // Feature selection - use Changed()
   if cmd.Flags().Changed("tab") {
       // User explicitly used --tab flag
   }
   ```

**Testing:**
- Verify warnings appear correctly
- Test default values vs explicit values

**Files Modified:**
- `main.go`
- `handlers.go`

---

## Phase 4: Documentation & Polish (Priority 4)

**Goal:** Improve code documentation and clean up minor issues.

### Task 4.1: Add Godoc Comments to Public Functions

**Location:** `handlers.go`, `main.go`

**Problem:** Public functions lack documentation.

**Implementation Steps:**

1. Add godoc comments to all exported functions:
   ```go
   // snag fetches content from a single URL according to the provided configuration.
   // It manages the browser lifecycle, handles all output formats, and ensures proper
   // cleanup even on errors.
   //
   // The function will:
   //   - Connect to existing browser or launch new one
   //   - Navigate to the URL and wait for content
   //   - Convert content to requested format
   //   - Write output to file or stdout
   //   - Clean up browser resources
   func snag(rt *Runtime, config *Config) error {
   ```

2. Add godoc to helper functions:
   ```go
   // processPageContent handles format conversion for all output types.
   // For binary formats (PDF, PNG), it uses the page object directly.
   // For text formats, it extracts HTML first then converts.
   func processPageContent(page *rod.Page, format string, outputFile string) error {
   ```

3. Add godoc to configuration structs:
   ```go
   // Config holds all configuration for a single page fetch operation.
   type Config struct {
       URL           string  // Target URL to fetch
       OutputFile    string  // Optional output file path
       OutputDir     string  // Optional output directory for auto-naming
       Format        string  // Output format: md, html, text, pdf, png
       Timeout       int     // Page load timeout in seconds
       WaitFor       string  // Optional CSS selector to wait for
       Port          int     // Chrome remote debugging port
       CloseTab      bool    // Close tab after fetching
       ForceHeadless bool    // Force headless mode
       OpenBrowser   bool    // Open visible browser
       UserAgent     string  // Custom user agent
       UserDataDir   string  // Custom user data directory
   }
   ```

**Testing:**
- Run `go doc` to verify documentation
- Verify godoc.org rendering (if applicable)

**Files Modified:**
- `handlers.go`
- `main.go`

---

### Task 4.2: Remove Redundant Comments

**Location:** `main.go:129-130`

**Problem:** Comment states the obvious about Cobra's default behavior.

**Implementation Steps:**

1. Remove or update comment:
   ```go
   // Before:
   // Hide default values for boolean flags in help output
   // Cobra doesn't show "(default: false)" by default, so we're good!

   // After:
   // (removed)
   ```

**Files Modified:**
- `main.go`

---

### Task 4.3: Consider Renaming Flag Variables

**Location:** `main.go:36-55`

**Problem:** Flag variables shadow common terms (tab, output, format).

**Solution:** Prefix with `flag` or use more specific names.

**Implementation Steps:**

1. Decide on naming convention:
   - Option A: Prefix all with `flag` (flagTab, flagOutput)
   - Option B: Use more specific names (tabPattern, outputFile)
   - Option C: Keep current (acceptable in this context)

2. If renaming, update all references:
   ```go
   // Option A:
   var (
       flagURLFile     string
       flagOutput      string
       flagOutputDir   string
       flagFormat      string
       // ...
   )

   rootCmd.Flags().StringVar(&flagOutput, "output", "o", "", "...")
   ```

**Decision Point:** This is optional - current naming is acceptable since these are package-level in main.

**Files Modified:**
- `main.go` (if implementing)

---

### Task 4.4: Inline Trivial plural() Function

**Location:** `handlers.go:930-935`

**Problem:** Single-use function for trivial logic.

**Implementation Steps:**

1. Option A: Inline at use site:
   ```go
   // Before:
   logger.Info("Processing %d URL%s...", len(validatedURLs), plural(len(validatedURLs)))

   // After:
   urlWord := "URL"
   if len(validatedURLs) != 1 {
       urlWord = "URLs"
   }
   logger.Info("Processing %d %s...", len(validatedURLs), urlWord)
   ```

2. Option B: Use more sophisticated pluralization:
   ```go
   func pluralize(word string, count int) string {
       if count == 1 {
           return word
       }
       // Handle common cases
       if strings.HasSuffix(word, "s") {
           return word + "es"
       }
       return word + "s"
   }

   logger.Info("Processing %d %s...", count, pluralize("URL", count))
   ```

3. Option C: Keep as-is (it's fine for this simple case)

**Decision Point:** Current implementation is acceptable. Only refactor if adding more plural cases.

**Files Modified:**
- `handlers.go` (if implementing)

---

## Testing Strategy

### Unit Tests

After each task:
- Run `go test -v ./...`
- Verify all 124 tests pass
- Check test coverage: `go test -cover ./...`

### Integration Tests

After each phase:
- Run CLI integration tests
- Test common workflows:
  ```bash
  ./snag https://example.com
  ./snag --list-tabs
  ./snag --all-tabs -d output/
  ./snag -t 1
  ```

### Code Quality

Before completion:
- Run `go vet ./...` (must pass)
- Run `golint ./...` (review warnings)
- Run `gofmt -s -w .` (apply formatting)
- Check binary size: `ls -lh snag`

### Performance

Before completion:
- Benchmark single URL fetch
- Benchmark multiple URL fetch
- Compare with pre-refactor performance

---

## Documentation Updates

### AGENTS.md

Update after Phase 1 completion:
- Document new Runtime approach
- Update architecture section
- Note removal of global state

### README.md

No changes needed (user-facing behavior unchanged)

### PROJECT.md

Update main PROJECT.md after completion:
- Mark post-migration refactor as complete
- Update "Recently Completed" section

---

## Migration Path

### Recommended Order

1. **Phase 1 (Critical)** - Do first, test thoroughly
2. **Phase 2 (Architecture)** - Do second, enables better testing
3. **Phase 3 (Code Quality)** - Do third, reduces duplication
4. **Phase 4 (Documentation)** - Do last, polish

### Can Skip/Defer

- Task 4.3 (Renaming flag variables) - optional
- Task 4.4 (Inline plural function) - optional
- Task 3.5 (Move help template) - nice to have

### Risk Areas

- **Task 1.1 (Global state removal)** - Largest refactor, test extensively
- **Task 2.2 (Context support)** - Changes many function signatures
- **Task 3.1 (Extract batch code)** - Complex refactor, verify behavior unchanged

---

## Success Metrics

### Code Quality Metrics

- **Cyclomatic Complexity:** Reduced by extraction of common code
- **Function Length:** No function > 100 lines
- **Code Duplication:** < 5% duplicate code
- **Test Coverage:** Maintain ≥ current coverage

### Behavioral Metrics

- **All 124 tests pass** ✅
- **Binary size ≤ 20 MB** ✅
- **Performance unchanged** ✅
- **Help output unchanged** ✅

### Maintainability Metrics

- **No global mutable state** ✅
- **Consistent error handling** ✅
- **Comprehensive validation** ✅
- **Full godoc coverage** ✅

---

## Notes

### Why This Refactoring Matters

1. **Testability:** Removing global state makes code much easier to unit test
2. **Concurrency:** Context support enables future concurrent operations
3. **Maintainability:** Less duplication means easier bug fixes
4. **Correctness:** Better validation catches errors earlier

### Original Migration Review

This project was created from a comprehensive code review of the Cobra migration (commit ebd544c). The migration itself was successful, but exposed several areas for improvement that are standard Go best practices.

### Timeline Estimate

- **Phase 1:** 3-4 hours
- **Phase 2:** 3-4 hours
- **Phase 3:** 2-3 hours
- **Phase 4:** 1-2 hours
- **Testing:** 1-2 hours

**Total:** 10-15 hours

---

## Completion Checklist

- [ ] Phase 1: All critical issues resolved
- [ ] Phase 2: All architecture improvements implemented
- [ ] Phase 3: All code quality improvements implemented
- [ ] Phase 4: All documentation added
- [ ] All 124 tests passing
- [ ] Code coverage ≥ current
- [ ] `go vet` passes
- [ ] `gofmt` applied
- [ ] Binary size ≤ 20 MB
- [ ] Performance benchmarks comparable
- [ ] AGENTS.md updated
- [ ] PROJECT.md updated
