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
	"sync"
	"syscall"

	"github.com/spf13/cobra"
)

var version = "dev"

const (
	FormatMarkdown = "md"
	FormatHTML     = "html"
	FormatText     = "text"
	FormatPDF      = "pdf"
	FormatPNG      = "png"
)

const (
	ExitCodeSuccess   = 0
	ExitCodeError     = 1
	ExitCodeInterrupt = 130 // 128 + SIGINT (2)
	ExitCodeSIGTERM   = 143 // 128 + SIGTERM (15)
)

const (
	MaxDisplayURLLength = 80
	MaxTabLineLength    = 120
	MaxSlugLength       = 80
)

const (
	DefaultTimeout = 30
)

type Config struct {
	URL           string
	OutputFile    string
	OutputDir     string
	Format        string
	Timeout       int
	WaitFor       string
	Port          int
	CloseTab      bool
	ForceHeadless bool
	OpenBrowser   bool
	UserAgent     string
	UserDataDir   string
}

func (c *Config) BrowserOptions() BrowserOptions {
	return BrowserOptions{
		Port:          c.Port,
		ForceHeadless: c.ForceHeadless,
		OpenBrowser:   c.OpenBrowser,
		UserAgent:     c.UserAgent,
		UserDataDir:   c.UserDataDir,
	}
}

var (
	logger         *Logger
	browserManager *BrowserManager
	browserMutex   sync.Mutex
)

var (
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
	killBrowser bool
	doctor      bool
	showVersion bool
	verbose     bool
	quiet       bool
	debug       bool
	userAgent   string
	userDataDir string
)

const helpTemplate = `USAGE:
  snag [options] URL...

DESCRIPTION:
  snag fetches web page content using Chromium/Chrome automation.
  It can connect to existing browser sessions, launch headless browsers, or open
  visible browsers for authenticated sessions.

  Output formats:  Markdown (md), HTML, text (txt), PDF, or PNG.
  Filename format: yyyy-mm-dd-hhmmss-<title>-<n>.<ext>

  The perfect companion for AI agents to gain context from web pages.

EXAMPLES:
  # Fetch a single page (Markdown to stdout)
  snag example.com
  snag https://github.com/grantcarthew/snag

  # Different output formats
  snag -f html example.com
  snag -f text example.com > page.txt
  snag -f pdf -o doc.pdf example.com

  # Save to file
  snag -o page.md example.com
  snag -d output/ example.com          # Auto-generated filename

  # Fetch multiple pages
  snag example.com github.com          # Auto-generated filenames to pwd
  snag -d output/ url1 url2 url3
  snag --url-file urls.txt -d ./pages/

  # Work with browser tabs (index and listed in alphabetical order)
  snag --list-tabs                     # List all open tabs
  snag -t 1                            # Fetch first tab
  snag -t "github"                     # Match tab by URL pattern
  snag -t 2-5 -d tabs/                 # Fetch tabs 2 through 5
  snag --all-tabs -d output/           # Fetch all open tabs

  # Authenticated sessions
  snag --open-browser                  # Open browser, login manually
  snag -t "dashboard" -o data.md       # Fetch authenticated page

  # Advanced options
  snag --wait-for ".content" example.com
  snag --timeout 60 slow-site.com
  snag --user-agent "Bot/1.0" example.com

OPTIONS:
  -l, --list-tabs              List all open tabs in the browser
  -t, --tab int|string         Fetch from existing tab by pattern (tab number or string)
  -a, --all-tabs               Process all open browser tabs (saves with auto-generated filenames)
      --url-file string        Read URLs from file (one per line, supports comments)

  -f, --format string          Output format: md | html | text | pdf | png (default md)
  -o, --output string          Save output to file instead of stdout
  -d, --output-dir string      Save files with auto-generated names to directory

  -b, --open-browser           Open browser visibly with remote debugging enabled (no URL required)
  -c, --close-tab              Close the browser tab after fetching content
      --force-headless         Force headless mode even if the browser is running
  -p, --port int               Chromium/Chrome remote debugging port (default 9222)
      --user-agent string      Custom user agent (bypass headless detection)
      --user-data-dir string   Custom Chromium/Chrome user data directory (for session isolation)

      --timeout int            Page load timeout in seconds (default 30)
  -w, --wait-for string        Wait for CSS selector before extracting content

      --doctor                 Display comprehensive diagnostic information
  -k, --kill-browser           Kill browser processes with remote debugging enabled

      --debug                  Enable debug output
  -q, --quiet                  Suppress all output except errors and content
      --verbose                Enable verbose logging output

  -h, --help                   help for snag
  -v, --version                version for snag
`

var rootCmd = &cobra.Command{
	Use:          "snag [options] URL...",
	Short:        "Intelligently fetch web page content using a browser engine",
	Args:         cobra.ArbitraryArgs,
	RunE:         runCobra,
	SilenceUsage: true,
}

func init() {
	rootCmd.Flags().StringVar(&urlFile, "url-file", "", "Read URLs from file (one per line, supports comments)")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "Save output to file instead of stdout")
	rootCmd.Flags().StringVarP(&outputDir, "output-dir", "d", "", "Save files with auto-generated names to directory")
	rootCmd.Flags().StringVarP(&format, "format", "f", FormatMarkdown, "Output format: md | html | text | pdf | png")
	rootCmd.Flags().StringVarP(&waitFor, "wait-for", "w", "", "Wait for CSS selector before extracting content")
	rootCmd.Flags().StringVarP(&tab, "tab", "t", "", "Fetch from existing tab by pattern (tab number or string)")
	rootCmd.Flags().StringVar(&userAgent, "user-agent", "", "Custom user agent (bypass headless detection)")
	rootCmd.Flags().StringVar(&userDataDir, "user-data-dir", "", "Custom Chromium/Chrome user data directory (for session isolation)")

	rootCmd.Flags().IntVar(&timeout, "timeout", 30, "Page load timeout in seconds")
	rootCmd.Flags().IntVarP(&port, "port", "p", 9222, "Chromium/Chrome remote debugging port")

	rootCmd.Flags().BoolVarP(&closeTab, "close-tab", "c", false, "Close the browser tab after fetching content")
	rootCmd.Flags().BoolVar(&forceHead, "force-headless", false, "Force headless mode even if the browser is running")
	rootCmd.Flags().BoolVarP(&openBrowser, "open-browser", "b", false, "Open browser visibly with remote debugging enabled (no URL required)")
	rootCmd.Flags().BoolVarP(&listTabs, "list-tabs", "l", false, "List all open tabs in the browser")
	rootCmd.Flags().BoolVarP(&allTabs, "all-tabs", "a", false, "Process all open browser tabs (saves with auto-generated filenames)")
	rootCmd.Flags().BoolVarP(&killBrowser, "kill-browser", "k", false, "Kill browser processes with remote debugging enabled")
	rootCmd.Flags().BoolVar(&doctor, "doctor", false, "Display comprehensive diagnostic information")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Display version information")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging output")
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress all output except errors and content")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug output")

	rootCmd.MarkFlagsMutuallyExclusive("quiet", "verbose", "debug")

	rootCmd.SetHelpTemplate(helpTemplate)
}

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Fprintf(os.Stderr, "\nReceived %v, cleaning up...\n", sig)

		browserMutex.Lock()
		if browserManager != nil {
			browserManager.Close()
		}
		browserMutex.Unlock()

		if sig == os.Interrupt {
			os.Exit(ExitCodeInterrupt)
		}
		os.Exit(ExitCodeSIGTERM)
	}()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(ExitCodeError)
	}
}

func validateFlagCombinations(cmd *cobra.Command, hasURLs bool, hasMultipleURLs bool) error {
	contentSources := 0
	if hasURLs {
		contentSources++
	}
	if cmd.Flags().Changed("tab") {
		contentSources++
	}
	if allTabs {
		contentSources++
	}

	if cmd.Flags().Changed("tab") && hasURLs {
		logger.Error("Cannot use both --tab and URL arguments (mutually exclusive content sources)")
		return ErrTabURLConflict
	}

	if allTabs && hasURLs {
		logger.Error("Cannot use both --all-tabs and URL arguments (mutually exclusive content sources)")
		return fmt.Errorf("conflicting flags: --all-tabs and URL arguments")
	}

	if cmd.Flags().Changed("tab") && allTabs {
		logger.Error("Cannot use both --tab and --all-tabs (mutually exclusive content sources)")
		return fmt.Errorf("conflicting flags: --tab and --all-tabs")
	}

	if openBrowser && forceHead {
		logger.Error("Cannot use both --force-headless and --open-browser (conflicting modes)")
		return fmt.Errorf("conflicting flags: --force-headless and --open-browser")
	}

	if forceHead && cmd.Flags().Changed("tab") {
		logger.Error("Cannot use --force-headless with --tab (--tab requires existing browser connection)")
		return fmt.Errorf("conflicting flags: --force-headless and --tab")
	}

	if forceHead && allTabs {
		logger.Error("Cannot use --force-headless with --all-tabs (--all-tabs requires existing browser connection)")
		return fmt.Errorf("conflicting flags: --force-headless and --all-tabs")
	}

	outputFile := strings.TrimSpace(output)
	outDir := strings.TrimSpace(outputDir)

	if outputFile != "" && outDir != "" {
		logger.Error("Cannot use both --output and --output-dir")
		logger.Info("Use --output for specific filename OR --output-dir for auto-generated filename")
		return fmt.Errorf("conflicting flags: --output and --output-dir")
	}

	if hasMultipleURLs && outputFile != "" {
		logger.Error("Cannot use --output with multiple content sources. Use --output-dir instead")
		return ErrOutputFlagConflict
	}

	if allTabs && outputFile != "" {
		logger.Error("Cannot use --output with multiple content sources. Use --output-dir instead")
		return ErrOutputFlagConflict
	}

	if closeTab && forceHead {
		logger.Warning("--close-tab is ignored in headless mode (tabs close automatically)")
	}

	if openBrowser && cmd.Flags().Changed("tab") {
		logger.Warning("--tab ignored with --open-browser (no content fetching)")
	}

	if openBrowser && allTabs {
		logger.Warning("--all-tabs ignored with --open-browser (no content fetching)")
	}

	return nil
}

func runCobra(cmd *cobra.Command, args []string) error {
	level := LevelNormal
	if debug {
		level = LevelDebug
	} else if verbose {
		level = LevelVerbose
	} else if quiet {
		level = LevelQuiet
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

	for _, arg := range args {
		trimmedArg := strings.TrimSpace(arg)
		if trimmedArg != "" {
			urls = append(urls, trimmedArg)
		}
	}

	if doctor {
		return handleDoctor(cmd)
	}

	if showVersion {
		fmt.Printf("snag version %s\n", version)
		return nil
	}

	if killBrowser {
		if len(urls) > 0 {
			logger.Error("Cannot use --kill-browser with URL arguments (conflicting operations)")
			return fmt.Errorf("conflicting flags: --kill-browser and URL arguments")
		}
		if listTabs {
			logger.Error("Cannot use --kill-browser with --list-tabs (conflicting operations)")
			return fmt.Errorf("conflicting flags: --kill-browser and --list-tabs")
		}
		if allTabs {
			logger.Error("Cannot use --kill-browser with --all-tabs (conflicting operations)")
			return fmt.Errorf("conflicting flags: --kill-browser and --all-tabs")
		}
		if cmd.Flags().Changed("tab") {
			logger.Error("Cannot use --kill-browser with --tab (conflicting operations)")
			return fmt.Errorf("conflicting flags: --kill-browser and --tab")
		}
		if openBrowser {
			logger.Error("Cannot use --kill-browser with --open-browser (conflicting operations)")
			return fmt.Errorf("conflicting flags: --kill-browser and --open-browser")
		}
		if urlFile != "" {
			logger.Error("Cannot use --kill-browser with --url-file (conflicting operations)")
			return fmt.Errorf("conflicting flags: --kill-browser and --url-file")
		}
		return handleKillBrowser(cmd)
	}

	if listTabs {
		if len(urls) > 0 {
			logger.Verbose("--list-tabs overrides URL arguments (URLs will be ignored)")
		}
		return handleListTabs(cmd)
	}

	hasURLs := len(urls) > 0
	hasMultipleURLs := len(urls) > 1
	if err := validateFlagCombinations(cmd, hasURLs, hasMultipleURLs); err != nil {
		return err
	}

	if allTabs {
		return handleAllTabs(cmd)
	}

	if cmd.Flags().Changed("tab") {
		return handleTabFetch(cmd)
	}

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

	if len(urls) == 0 {
		logger.Error("No URLs provided")
		logger.ErrorWithSuggestion("Provide URLs as arguments or use --url-file", "snag <url> or snag --url-file urls.txt")
		return ErrNoValidURLs
	}

	if openBrowser && len(urls) > 0 {
		return handleOpenURLsInBrowser(cmd, urls)
	}

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

		validatedUserAgent := validateUserAgent(userAgent, cmd.Flags().Changed("user-agent"))
		validatedWaitFor := validateWaitFor(waitFor, cmd.Flags().Changed("wait-for"))

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

	return handleMultipleURLs(cmd, urls)
}
