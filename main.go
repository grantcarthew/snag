// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

// version can be set via ldflags at build time
var version = "dev"

const (
	FormatMarkdown = "md"
	FormatHTML     = "html"
	FormatText     = "text"
	FormatPDF      = "pdf"
	FormatPNG      = "png"
)

var (
	logger         *Logger
	browserManager *BrowserManager
)

var (
	// Flag variables
	urlFile     string
	output      string
	outputDir   string
	format      string
	timeout     int
	waitFor     string
	port        int
	closeTab    bool
	forceHead   bool
	openBrowser bool
	listTabs    bool
	tab         string
	allTabs     bool
	verbose     bool
	quiet       bool
	debug       bool
	userAgent   string
	userDataDir string
)

// Custom help template matching the original urfave/cli template
var cobraHelpTemplate = `USAGE:
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

var rootCmd = &cobra.Command{
	Use:     "snag [options] URL...",
	Short:   "Intelligently fetch web page content using a browser engine",
	Version: version,
	Args:    cobra.ArbitraryArgs, // Allow any number of arguments (URLs)
	RunE:    runCobra,
}

func init() {
	// String flags
	rootCmd.Flags().StringVar(&urlFile, "url-file", "", "Read URLs from FILE (one per line, supports comments)")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "Save output to FILE instead of stdout")
	rootCmd.Flags().StringVarP(&outputDir, "output-dir", "d", "", "Save files with auto-generated names to DIRECTORY")
	rootCmd.Flags().StringVarP(&format, "format", "f", FormatMarkdown, "Output FORMAT: md | html | text | pdf | png")
	rootCmd.Flags().StringVarP(&waitFor, "wait-for", "w", "", "Wait for CSS SELECTOR before extracting content")
	rootCmd.Flags().StringVarP(&tab, "tab", "t", "", "Fetch from existing tab by PATTERN (tab number or string)")
	rootCmd.Flags().StringVar(&userAgent, "user-agent", "", "Custom user agent STRING (bypass headless detection)")
	rootCmd.Flags().StringVar(&userDataDir, "user-data-dir", "", "Custom Chromium/Chrome user data DIRECTORY (for session isolation)")

	// Int flags
	rootCmd.Flags().IntVar(&timeout, "timeout", 30, "Page load timeout in SECONDS")
	rootCmd.Flags().IntVarP(&port, "port", "p", 9222, "Chrome remote debugging PORT")

	// Bool flags
	rootCmd.Flags().BoolVarP(&closeTab, "close-tab", "c", false, "Close the browser tab after fetching content")
	rootCmd.Flags().BoolVar(&forceHead, "force-headless", false, "Force headless mode even if Chrome is running")
	rootCmd.Flags().BoolVarP(&openBrowser, "open-browser", "b", false, "Open browser visibly with remote debugging enabled (no URL required)")
	rootCmd.Flags().BoolVarP(&listTabs, "list-tabs", "l", false, "List all open tabs in the browser")
	rootCmd.Flags().BoolVarP(&allTabs, "all-tabs", "a", false, "Process all open browser tabs (saves with auto-generated filenames)")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging output")
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress all output except errors and content")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug output")

	// Hide default values for boolean flags in help output
	// Cobra doesn't show "(default: false)" by default, so we're good!

	// Set custom help template
	rootCmd.SetHelpTemplate(cobraHelpTemplate)
}

func main() {
	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Fprintf(os.Stderr, "\nReceived %v, cleaning up...\n", sig)

		if browserManager != nil {
			browserManager.Close()
		}

		if sig == os.Interrupt {
			os.Exit(130)
		}
		os.Exit(143)
	}()

	// Execute Cobra command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runCobra(cmd *cobra.Command, args []string) error {
	// Determine logging level (same logic as original)
	level := LevelNormal

	lastLogFlag := ""
	lastLogIndex := -1
	for i, arg := range os.Args {
		if arg == "--quiet" || arg == "-q" {
			if i > lastLogIndex {
				lastLogIndex = i
				lastLogFlag = "quiet"
			}
		} else if arg == "--verbose" {
			if i > lastLogIndex {
				lastLogIndex = i
				lastLogFlag = "verbose"
			}
		} else if arg == "--debug" {
			if i > lastLogIndex {
				lastLogIndex = i
				lastLogFlag = "debug"
			}
		}
	}

	switch lastLogFlag {
	case "quiet":
		level = LevelQuiet
	case "debug":
		level = LevelDebug
	case "verbose":
		level = LevelVerbose
	}

	logger = NewLogger(level)

	var urls []string

	outputFile := strings.TrimSpace(output)
	outDir := strings.TrimSpace(outputDir)

	// Load URLs from file if specified
	if urlFile != "" {
		fileURLs, err := loadURLsFromFile(strings.TrimSpace(urlFile))
		if err != nil {
			return err
		}
		urls = append(urls, fileURLs...)
	}

	// Add command line URL arguments
	for _, arg := range args {
		trimmedArg := strings.TrimSpace(arg)
		if trimmedArg != "" {
			urls = append(urls, trimmedArg)
		}
	}

	// Cobra handles flag parsing automatically - flags can appear anywhere

	// Handle --list-tabs
	if listTabs {
		if len(urls) > 0 {
			logger.Verbose("--list-tabs overrides URL arguments (URLs will be ignored)")
		}
		return handleListTabs(cmd)
	}

	// Handle --all-tabs
	if allTabs {
		if len(urls) > 0 {
			logger.Error("Cannot use both --all-tabs and URL arguments (mutually exclusive content sources)")
			return fmt.Errorf("conflicting flags: --all-tabs and URL arguments")
		}
		if forceHead {
			logger.Error("Cannot use --force-headless with --all-tabs (--all-tabs requires existing browser connection)")
			return fmt.Errorf("conflicting flags: --force-headless and --all-tabs")
		}
		if outputFile != "" {
			logger.Error("Cannot use --output with --all-tabs (multiple outputs). Use --output-dir instead")
			return ErrOutputFlagConflict
		}
		if openBrowser {
			logger.Warning("--all-tabs ignored with --open-browser (no content fetching)")
		}
		return handleAllTabs(cmd)
	}

	// Handle --tab
	if cmd.Flags().Changed("tab") {
		if len(urls) > 0 {
			logger.Error("Cannot use --tab with URL argument (mutually exclusive content sources)")
			return ErrTabURLConflict
		}
		if allTabs {
			logger.Error("Cannot use both --tab and --all-tabs (mutually exclusive content sources)")
			return fmt.Errorf("conflicting flags: --tab and --all-tabs")
		}
		if forceHead {
			logger.Error("Cannot use --force-headless with --tab (--tab requires existing browser connection)")
			return fmt.Errorf("conflicting flags: --force-headless and --tab")
		}
		if openBrowser {
			logger.Warning("--tab ignored with --open-browser (no content fetching)")
		}
		return handleTabFetch(cmd)
	}

	// Check for conflicting browser modes
	if openBrowser && forceHead {
		logger.Error("Cannot use both --force-headless and --open-browser (conflicting modes)")
		return fmt.Errorf("conflicting flags: --force-headless and --open-browser")
	}

	// Handle --open-browser without URLs
	if openBrowser && len(urls) == 0 {
		if cmd.Flags().Changed("format") {
			logger.Warning("--format ignored with --open-browser (no content fetching)")
		}
		if cmd.Flags().Changed("output") {
			logger.Warning("--output ignored with --open-browser (no content fetching)")
		}
		if cmd.Flags().Changed("output-dir") {
			logger.Warning("--output-dir ignored with --open-browser (no content fetching)")
		}
		if cmd.Flags().Changed("timeout") {
			logger.Warning("--timeout ignored with --open-browser (no content fetching)")
		}
		if cmd.Flags().Changed("wait-for") {
			logger.Warning("--wait-for ignored with --open-browser (no content fetching)")
		}
		if cmd.Flags().Changed("user-agent") {
			logger.Warning("--user-agent ignored with --open-browser (no navigation)")
		}
		if closeTab {
			logger.Warning("--close-tab ignored with --open-browser (no content fetching)")
		}

		validatedUserDataDir := ""
		if cmd.Flags().Changed("user-data-dir") {
			validatedDir, err := validateUserDataDir(userDataDir)
			if err != nil {
				return err
			}
			validatedUserDataDir = validatedDir
		}

		logger.Info("Opening browser...")
		bm := NewBrowserManager(BrowserOptions{
			Port:        port,
			OpenBrowser: true,
			UserDataDir: validatedUserDataDir,
		})
		return bm.OpenBrowserOnly()
	}

	// Require at least one URL
	if len(urls) == 0 {
		logger.Error("No URLs provided")
		logger.ErrorWithSuggestion("Provide URLs as arguments or use --url-file", "snag <url> or snag --url-file urls.txt")
		return ErrNoValidURLs
	}

	// Handle --open-browser with URLs
	if openBrowser && len(urls) > 0 {
		return handleOpenURLsInBrowser(cmd, urls)
	}

	// Handle single URL
	if len(urls) == 1 {
		urlStr := urls[0]

		validatedURL, err := validateURL(urlStr)
		if err != nil {
			return err
		}

		logger.Verbose("Target URL: %s", validatedURL)

		outputFormat := normalizeFormat(format)

		validatedUserDataDir := ""
		if cmd.Flags().Changed("user-data-dir") {
			validatedDir, err := validateUserDataDir(userDataDir)
			if err != nil {
				return err
			}
			validatedUserDataDir = validatedDir
		}

		validatedUserAgent := ""
		if cmd.Flags().Changed("user-agent") {
			validatedUserAgent = validateUserAgent(userAgent)
		}

		validatedWaitFor := ""
		if cmd.Flags().Changed("wait-for") {
			validatedWaitFor = validateWaitFor(waitFor)
		}

		config := &Config{
			URL:           validatedURL,
			OutputFile:    outputFile,
			OutputDir:     outDir,
			Format:        outputFormat,
			Timeout:       timeout,
			WaitFor:       validatedWaitFor,
			Port:          port,
			CloseTab:      closeTab,
			ForceHeadless: forceHead,
			OpenBrowser:   openBrowser,
			UserAgent:     validatedUserAgent,
			UserDataDir:   validatedUserDataDir,
		}

		logger.Debug("Config: format=%s, timeout=%d, port=%d", config.Format, config.Timeout, config.Port)

		if config.CloseTab && config.ForceHeadless {
			logger.Warning("--close-tab is ignored in headless mode (tabs close automatically)")
		}

		if config.OutputFile != "" && config.OutputDir != "" {
			logger.Error("Cannot use --output and --output-dir together")
			logger.Info("Use --output for specific filename OR --output-dir for auto-generated filename")
			return fmt.Errorf("conflicting flags: --output and --output-dir")
		}

		if err := validateFormat(config.Format); err != nil {
			return err
		}

		if err := validateTimeout(config.Timeout); err != nil {
			return err
		}

		if err := validatePort(config.Port); err != nil {
			return err
		}

		if cmd.Flags().Changed("output") || config.OutputFile != "" {
			if err := validateOutputPath(config.OutputFile); err != nil {
				return err
			}
			checkExtensionMismatch(config.OutputFile, config.Format)
		}

		if cmd.Flags().Changed("output-dir") && config.OutputDir == "" {
			config.OutputDir = "."
		}

		if config.OutputDir != "" {
			if err := validateDirectory(config.OutputDir); err != nil {
				return err
			}
		}

		logger.Verbose("Configuration: format=%s, timeout=%ds, port=%d", config.Format, config.Timeout, config.Port)

		return snag(config)
	}

	// Handle multiple URLs
	return handleMultipleURLs(cmd, urls)
}
