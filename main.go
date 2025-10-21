// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/go-rod/rod"
	"github.com/urfave/cli/v2"
)

// version can be set via ldflags at build time
var version = "dev"

const (
	// Output format constants for --format flag
	FormatMarkdown = "markdown"
	FormatHTML     = "html"
	FormatText     = "text"
	FormatPDF      = "pdf"
)

var (
	logger         *Logger
	browserManager *BrowserManager
)

func main() {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Fprintf(os.Stderr, "\nReceived %v, cleaning up...\n", sig)

		// Clean up browser if it exists (only closes headless browsers)
		if browserManager != nil {
			browserManager.Close()
		}

		// Exit with standard signal codes
		if sig == os.Interrupt {
			os.Exit(130) // 128 + 2 (SIGINT)
		}
		os.Exit(143) // 128 + 15 (SIGTERM)
	}()

	app := &cli.App{
		Name:            "snag",
		Usage:           "Intelligently fetch web page content with browser engine",
		UsageText:       "snag [options] <url>",
		Version:         version,
		HideVersion:     false,
		HideHelpCommand: true,
		Authors: []*cli.Author{
			{
				Name:  "Grant Carthew",
				Email: "grant@carthew.net",
			},
		},
		Description: `snag fetches web page content using Chrome/Chromium via the Chrome DevTools Protocol.
   It can connect to existing browser sessions, launch headless browsers, or open
   visible browsers for authenticated sessions. Output can be Markdown or HTML.

   The perfect companion for AI agents to gain context from web pages.`,
		ArgsUsage: "<url>",
		Flags: []cli.Flag{
			// Output Control
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Save output to `FILE` instead of stdout",
			},
			&cli.StringFlag{
				Name:    "output-dir",
				Aliases: []string{"d"},
				Usage:   "Save files with auto-generated names to `DIRECTORY`",
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "Output `FORMAT`: markdown | html | text | pdf",
				Value:   FormatMarkdown,
			},
			&cli.BoolFlag{
				Name:    "screenshot",
				Aliases: []string{"s"},
				Usage:   "Capture full-page screenshot (PNG format)",
			},

			// Page Loading
			&cli.IntFlag{
				Name:  "timeout",
				Usage: "Page load timeout in `SECONDS`",
				Value: 30,
			},
			&cli.StringFlag{
				Name:    "wait-for",
				Aliases: []string{"w"},
				Usage:   "Wait for CSS `SELECTOR` before extracting content",
			},

			// Browser Control
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "Chrome remote debugging `PORT`",
				Value:   9222,
			},
			&cli.BoolFlag{
				Name:    "close-tab",
				Aliases: []string{"c"},
				Usage:   "Close the browser tab after fetching content",
			},
			&cli.BoolFlag{
				Name:  "force-headless",
				Usage: "Force headless mode even if Chrome is running",
			},
			&cli.BoolFlag{
				Name:  "force-visible",
				Usage: "Force visible mode for authentication",
			},
			&cli.BoolFlag{
				Name:    "open-browser",
				Aliases: []string{"b"},
				Usage:   "Open Chrome browser in visible state (no URL required)",
			},
			&cli.BoolFlag{
				Name:    "list-tabs",
				Aliases: []string{"l"},
				Usage:   "List all open tabs in the browser",
			},
			&cli.StringFlag{
				Name:    "tab",
				Aliases: []string{"t"},
				Usage:   "Fetch from existing tab by `PATTERN` (tab number or string)",
			},
			&cli.BoolFlag{
				Name:    "all-tabs",
				Aliases: []string{"a"},
				Usage:   "Process all open browser tabs (saves with auto-generated filenames)",
			},

			// Logging/Debugging
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Enable verbose logging output",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Suppress all output except errors and content",
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Enable debug output",
			},

			// Request Control
			&cli.StringFlag{
				Name:  "user-agent",
				Usage: "Custom user agent `STRING` (bypass headless detection)",
			},
		},
		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// run is the main application action
func run(c *cli.Context) error {
	// Initialize logger based on flags
	level := LevelNormal
	if c.Bool("quiet") {
		level = LevelQuiet
	} else if c.Bool("debug") {
		level = LevelDebug
	} else if c.Bool("verbose") {
		level = LevelVerbose
	}
	logger = NewLogger(level)

	// Handle --open-browser flag WITHOUT URL (just open browser)
	if c.Bool("open-browser") && c.NArg() == 0 {
		logger.Info("Opening browser...")
		bm := NewBrowserManager(BrowserOptions{
			Port:         c.Int("port"),
			OpenBrowser:  true,
			ForceVisible: true,
		})
		return bm.OpenBrowserOnly()
	}

	// Handle --list-tabs flag (list tabs and exit)
	if c.Bool("list-tabs") {
		return handleListTabs(c)
	}

	// Handle --all-tabs flag (process all tabs)
	if c.Bool("all-tabs") {
		// Check for conflicting URL argument
		if c.NArg() > 0 {
			logger.Error("Cannot use --all-tabs with URL argument")
			logger.Info("Use --all-tabs alone to process all existing tabs")
			return fmt.Errorf("conflicting flags: --all-tabs and URL argument")
		}
		return handleAllTabs(c)
	}

	// Handle --tab flag (fetch from existing tab)
	if c.IsSet("tab") {
		// Check for conflicting URL argument
		if c.NArg() > 0 {
			logger.Error("Cannot use --tab with URL argument")
			logger.Info("Use either --tab to fetch from existing tab OR provide URL to fetch new page")
			return ErrTabURLConflict
		}
		return handleTabFetch(c)
	}

	// Validate URL argument
	if c.NArg() == 0 {
		logger.Error("URL argument is required")
		logger.ErrorWithSuggestion("Missing URL", "snag <url>")
		return ErrInvalidURL
	}

	urlStr := c.Args().First()

	// Validate and normalize URL
	validatedURL, err := validateURL(urlStr)
	if err != nil {
		return err
	}

	logger.Verbose("Target URL: %s", validatedURL)

	// Validate conflicting flags
	if c.Bool("force-headless") && c.Bool("force-visible") {
		logger.Error("Conflicting flags: --force-headless and --force-visible cannot be used together")
		return fmt.Errorf("conflicting flags: --force-headless and --force-visible")
	}

	// Extract configuration from flags
	config := &Config{
		URL:           validatedURL,
		OutputFile:    c.String("output"),
		OutputDir:     c.String("output-dir"),
		Format:        c.String("format"),
		Screenshot:    c.Bool("screenshot"),
		Timeout:       c.Int("timeout"),
		WaitFor:       c.String("wait-for"),
		Port:          c.Int("port"),
		CloseTab:      c.Bool("close-tab"),
		ForceHeadless: c.Bool("force-headless"),
		ForceVisible:  c.Bool("force-visible"),
		OpenBrowser:   c.Bool("open-browser"),
		UserAgent:     c.String("user-agent"),
	}

	logger.Debug("Config: format=%s, timeout=%d, port=%d", config.Format, config.Timeout, config.Port)

	// Validate conflicting output flags
	if config.OutputFile != "" && config.OutputDir != "" {
		logger.Error("Cannot use --output and --output-dir together")
		logger.Info("Use --output for specific filename OR --output-dir for auto-generated filename")
		return fmt.Errorf("conflicting flags: --output and --output-dir")
	}

	// Validate format
	if err := validateFormat(config.Format); err != nil {
		return err
	}

	// Validate timeout
	if err := validateTimeout(config.Timeout); err != nil {
		return err
	}

	// Validate port
	if err := validatePort(config.Port); err != nil {
		return err
	}

	// Validate output file path if provided
	if config.OutputFile != "" {
		if err := validateOutputPath(config.OutputFile); err != nil {
			return err
		}
	}

	// Validate output directory if provided
	if config.OutputDir != "" {
		if err := validateDirectory(config.OutputDir); err != nil {
			return err
		}
	}

	logger.Verbose("Configuration: format=%s, timeout=%ds, port=%d", config.Format, config.Timeout, config.Port)

	// Handle --output-dir: Generate filename after page is fetched
	// Note: For single URL fetches with -d, we need to fetch first to get the title
	// This will be handled in snag() function after page load

	// Execute the snag operation
	return snag(config)
}

// snag is the main function that orchestrates the web page fetching
func snag(config *Config) error {
	// Create browser manager
	bm := NewBrowserManager(BrowserOptions{
		Port:          config.Port,
		ForceHeadless: config.ForceHeadless,
		ForceVisible:  config.ForceVisible,
		OpenBrowser:   config.OpenBrowser,
		UserAgent:     config.UserAgent,
	})

	// Assign to global for signal handler access
	browserManager = bm

	// Connect to browser
	_, err := bm.Connect()
	if err != nil {
		if err == ErrBrowserNotFound {
			logger.Error("No Chromium-based browser found")
			logger.ErrorWithSuggestion(
				"Install Chrome, Chromium, Edge, or Brave to use snag",
				"brew install --cask google-chrome",
			)
		}
		browserManager = nil // Clear global on error
		return err
	}

	// Ensure browser cleanup
	defer func() {
		if config.CloseTab {
			logger.Verbose("Cleanup: closing tab and browser if needed")
		}
		bm.Close()
		browserManager = nil // Clear global after cleanup
	}()

	// Create new page
	page, err := bm.NewPage()
	if err != nil {
		return err
	}

	// Ensure page cleanup if requested
	if config.CloseTab {
		defer bm.ClosePage(page)
	}

	// Create page fetcher
	fetcher := NewPageFetcher(page, config.Timeout)

	// Fetch the page (navigates and loads content)
	html, err := fetcher.Fetch(FetchOptions{
		URL:     config.URL,
		Timeout: config.Timeout,
		WaitFor: config.WaitFor,
	})
	if err != nil {
		return err
	}

	// Handle --output-dir: Generate filename from page title
	if config.OutputDir != "" {
		// Get page info for title
		info, err := page.Info()
		if err != nil {
			return fmt.Errorf("failed to get page info: %w", err)
		}

		// Determine format for filename generation
		filenameFormat := config.Format
		if config.Screenshot {
			filenameFormat = "png"
		}

		// Generate filename
		timestamp := time.Now()
		filename := GenerateFilename(info.Title, filenameFormat, timestamp, config.URL)

		// Resolve conflicts
		finalFilename, err := ResolveConflict(config.OutputDir, filename)
		if err != nil {
			return fmt.Errorf("failed to resolve filename conflict: %w", err)
		}

		// Set OutputFile to full path
		config.OutputFile = filepath.Join(config.OutputDir, finalFilename)
	}

	// For binary formats without -o or -d: auto-generate filename in current directory
	// Binary formats (PDF, screenshot) should NEVER output to stdout (corrupts terminal)
	if config.OutputFile == "" && (config.Format == FormatPDF || config.Screenshot) {
		// Get page info for title
		info, err := page.Info()
		if err != nil {
			return fmt.Errorf("failed to get page info: %w", err)
		}

		// Determine format for filename generation
		filenameFormat := config.Format
		if config.Screenshot {
			filenameFormat = "png"
		}

		// Generate filename
		timestamp := time.Now()
		filename := GenerateFilename(info.Title, filenameFormat, timestamp, config.URL)

		// Resolve conflicts in current directory
		finalFilename, err := ResolveConflict(".", filename)
		if err != nil {
			return fmt.Errorf("failed to resolve filename conflict: %w", err)
		}

		config.OutputFile = finalFilename
		logger.Info("Auto-generated filename: %s", finalFilename)
	}

	// Handle screenshot capture (binary format)
	if config.Screenshot {
		converter := NewContentConverter("png")
		if err := converter.ProcessPage(page, config.OutputFile); err != nil {
			return err
		}
		return nil
	}

	// Create content converter
	converter := NewContentConverter(config.Format)

	// Handle binary formats (PDF) that need the page object
	if config.Format == FormatPDF {
		// Generate PDF from the already-loaded page
		if err := converter.ProcessPage(page, config.OutputFile); err != nil {
			return err
		}
		return nil
	}

	// For text formats, process the HTML
	if err := converter.Process(html, config.OutputFile); err != nil {
		return err
	}

	return nil
}

// Config holds the application configuration
type Config struct {
	URL           string
	OutputFile    string
	OutputDir     string
	Format        string
	Screenshot    bool
	AllTabs       bool
	Timeout       int
	WaitFor       string
	Port          int
	CloseTab      bool
	ForceHeadless bool
	ForceVisible  bool
	OpenBrowser   bool
	UserAgent     string
}

// displayTabList formats and displays a list of tabs to the specified writer
func displayTabList(tabs []TabInfo, w io.Writer) {
	if len(tabs) == 0 {
		fmt.Fprintf(w, "No tabs open in browser\n")
		return
	}

	fmt.Fprintf(w, "Available tabs in browser (%d tabs):\n", len(tabs))
	for _, tab := range tabs {
		fmt.Fprintf(w, "  [%d] %s - %s\n", tab.Index, tab.URL, tab.Title)
	}
}

// handleAllTabs processes all open browser tabs with auto-generated filenames
func handleAllTabs(c *cli.Context) error {
	// Extract configuration from flags
	format := c.String("format")
	screenshot := c.Bool("screenshot")
	timeout := c.Int("timeout")
	waitFor := c.String("wait-for")
	outputDir := c.String("output-dir")
	if outputDir == "" {
		outputDir = "." // Default to current working directory
	}

	// Validate format
	if err := validateFormat(format); err != nil {
		return err
	}

	// Validate timeout
	if err := validateTimeout(timeout); err != nil {
		return err
	}

	// Validate output directory
	if err := validateDirectory(outputDir); err != nil {
		return err
	}

	// Create browser manager in connect-only mode
	bm := NewBrowserManager(BrowserOptions{
		Port: c.Int("port"),
	})

	// Assign to global for signal handler access
	browserManager = bm
	defer func() { browserManager = nil }()

	// Connect to existing browser
	browser, err := bm.connectToExisting()
	if err != nil {
		logger.Error("No browser instance running with remote debugging")
		logger.ErrorWithSuggestion(
			"Start Chrome with remote debugging enabled",
			fmt.Sprintf("chrome --remote-debugging-port=%d", c.Int("port")),
		)
		logger.Info("Or run: snag --open-browser")
		return ErrNoBrowserRunning
	}

	// Assign browser to manager
	bm.browser = browser

	// Get list of all tabs
	tabs, err := bm.ListTabs()
	if err != nil {
		return err
	}

	if len(tabs) == 0 {
		logger.Info("No tabs open in browser")
		return nil
	}

	// Create single timestamp for entire batch
	timestamp := time.Now()

	logger.Info("Processing %d tabs...", len(tabs))

	// Track success/failure counts
	successCount := 0
	failureCount := 0

	// Process each tab
	for i, tab := range tabs {
		tabNum := i + 1
		logger.Info("[%d/%d] Processing: %s", tabNum, len(tabs), tab.URL)

		// Get page object for this tab
		page, err := bm.GetTabByIndex(tabNum)
		if err != nil {
			logger.Error("[%d/%d] Failed to get tab: %v", tabNum, len(tabs), err)
			failureCount++
			continue
		}

		// Wait for selector if specified
		if waitFor != "" {
			err := waitForSelector(page, waitFor, time.Duration(timeout)*time.Second)
			if err != nil {
				logger.Error("[%d/%d] Wait failed: %v", tabNum, len(tabs), err)
				failureCount++
				continue
			}
		}

		// Determine format for filename generation
		filenameFormat := format
		if screenshot {
			filenameFormat = "png"
		}

		// Generate filename from page title and URL
		filename := GenerateFilename(tab.Title, filenameFormat, timestamp, tab.URL)

		// Resolve filename conflicts
		finalFilename, err := ResolveConflict(outputDir, filename)
		if err != nil {
			logger.Error("[%d/%d] Failed to resolve filename conflict: %v", tabNum, len(tabs), err)
			failureCount++
			continue
		}

		outputPath := filepath.Join(outputDir, finalFilename)

		// Process content (screenshot or format)
		if screenshot {
			converter := NewContentConverter("png")
			if err := converter.ProcessPage(page, outputPath); err != nil {
				logger.Error("[%d/%d] Failed to capture screenshot: %v", tabNum, len(tabs), err)
				failureCount++
				continue
			}
		} else if format == FormatPDF {
			converter := NewContentConverter(format)
			if err := converter.ProcessPage(page, outputPath); err != nil {
				logger.Error("[%d/%d] Failed to generate PDF: %v", tabNum, len(tabs), err)
				failureCount++
				continue
			}
		} else {
			// Text formats (markdown, html, text)
			html, err := page.HTML()
			if err != nil {
				logger.Error("[%d/%d] Failed to extract HTML: %v", tabNum, len(tabs), err)
				failureCount++
				continue
			}

			converter := NewContentConverter(format)
			if err := converter.Process(html, outputPath); err != nil {
				logger.Error("[%d/%d] Failed to process content: %v", tabNum, len(tabs), err)
				failureCount++
				continue
			}
		}

		successCount++
	}

	// Summary
	logger.Success("Batch complete: %d succeeded, %d failed", successCount, failureCount)

	if failureCount > 0 {
		return fmt.Errorf("batch processing completed with %d failures", failureCount)
	}

	return nil
}

// handleListTabs lists all open tabs in the browser
func handleListTabs(c *cli.Context) error {
	// Create browser manager in connect-only mode
	bm := NewBrowserManager(BrowserOptions{
		Port: c.Int("port"),
	})

	// Try to connect to existing browser
	browser, err := bm.connectToExisting()
	if err != nil {
		logger.Error("No browser instance running with remote debugging")
		logger.ErrorWithSuggestion(
			"Start Chrome with remote debugging enabled",
			fmt.Sprintf("chrome --remote-debugging-port=%d", c.Int("port")),
		)
		logger.Info("Or run: snag --open-browser")
		return ErrNoBrowserRunning
	}

	// Assign browser to manager so ListTabs() can use it
	bm.browser = browser

	// Get list of tabs
	tabs, err := bm.ListTabs()
	if err != nil {
		return err
	}

	// Display tabs to stdout
	displayTabList(tabs, os.Stdout)

	return nil
}

// handleTabFetch fetches content from an existing tab by index
func handleTabFetch(c *cli.Context) error {
	// Get the tab value
	tabValue := c.String("tab")
	if tabValue == "" {
		return fmt.Errorf("--tab flag requires a value")
	}

	// Create browser manager in connect-only mode
	bm := NewBrowserManager(BrowserOptions{
		Port: c.Int("port"),
	})

	// Assign to global for signal handler access
	browserManager = bm
	// Ensure cleanup on all exit paths
	defer func() { browserManager = nil }()

	// Connect to existing browser
	browser, err := bm.connectToExisting()
	if err != nil {
		logger.Error("No browser instance running with remote debugging")
		logger.ErrorWithSuggestion(
			"Start Chrome with remote debugging enabled",
			fmt.Sprintf("chrome --remote-debugging-port=%d", c.Int("port")),
		)
		logger.Info("Or run: snag --open-browser")
		return ErrNoBrowserRunning
	}

	// Assign browser to manager
	bm.browser = browser

	// Determine if tab value is an integer index or a pattern
	var page *rod.Page
	if tabIndex, err := strconv.Atoi(tabValue); err == nil {
		// Integer - use as tab index
		logger.Verbose("Fetching from tab index: %d", tabIndex)
		page, err = bm.GetTabByIndex(tabIndex)
		if err != nil {
			if errors.Is(err, ErrTabIndexInvalid) {
				logger.Error("Tab index out of range")
				logger.Info("Run 'snag --list-tabs' to see available tabs")

				// Display available tabs to help the user
				if tabs, listErr := bm.ListTabs(); listErr == nil {
					fmt.Fprintln(os.Stderr, "")
					displayTabList(tabs, os.Stderr)
					fmt.Fprintln(os.Stderr, "")
				}
			}
			return err
		}
		logger.Success("Connected to tab [%d]", tabIndex)
	} else {
		// Not an integer - treat as pattern
		logger.Verbose("Fetching from tab matching pattern: %s", tabValue)
		page, err = bm.GetTabByPattern(tabValue)
		if err != nil {
			if errors.Is(err, ErrNoTabMatch) {
				logger.Error("No tab matches pattern '%s'", tabValue)
				logger.Info("Run 'snag --list-tabs' to see available tabs")

				// Display available tabs to help the user
				if tabs, listErr := bm.ListTabs(); listErr == nil {
					fmt.Fprintln(os.Stderr, "")
					displayTabList(tabs, os.Stderr)
					fmt.Fprintln(os.Stderr, "")
				}
			}
			return err
		}
		logger.Success("Connected to tab matching pattern: %s", tabValue)
	}

	// Extract configuration from flags
	format := c.String("format")
	screenshot := c.Bool("screenshot")
	timeout := c.Int("timeout")
	waitFor := c.String("wait-for")
	outputFile := c.String("output")

	// Validate format
	if err := validateFormat(format); err != nil {
		return err
	}

	// Validate timeout
	if err := validateTimeout(timeout); err != nil {
		return err
	}

	// Validate output file path if provided
	if outputFile != "" {
		if err := validateOutputPath(outputFile); err != nil {
			return err
		}
	}

	// Get current page info
	info, err := page.Info()
	if err != nil {
		return fmt.Errorf("failed to get page info: %w", err)
	}

	logger.Info("Fetching content from: %s", info.URL)

	// Wait for selector if specified
	if waitFor != "" {
		err := waitForSelector(page, waitFor, time.Duration(timeout)*time.Second)
		if err != nil {
			return err
		}
	}

	// For binary formats without -o or -d: auto-generate filename in current directory
	// Binary formats (PDF, screenshot) should NEVER output to stdout (corrupts terminal)
	if outputFile == "" && (format == FormatPDF || screenshot) {
		// Determine format for filename generation
		filenameFormat := format
		if screenshot {
			filenameFormat = "png"
		}

		// Generate filename
		timestamp := time.Now()
		filename := GenerateFilename(info.Title, filenameFormat, timestamp, info.URL)

		// Resolve conflicts in current directory
		finalFilename, err := ResolveConflict(".", filename)
		if err != nil {
			return fmt.Errorf("failed to resolve filename conflict: %w", err)
		}

		outputFile = finalFilename
		logger.Info("Auto-generated filename: %s", finalFilename)
	}

	// Handle screenshot capture (binary format)
	if screenshot {
		converter := NewContentConverter("png")
		if err := converter.ProcessPage(page, outputFile); err != nil {
			return err
		}
		return nil
	}

	// Create content converter
	converter := NewContentConverter(format)

	// Handle binary formats (PDF) that need the page object
	if format == FormatPDF {
		// Generate PDF from the page
		if err := converter.ProcessPage(page, outputFile); err != nil {
			return err
		}
		return nil
	}

	// For text formats, extract HTML and process
	html, err := page.HTML()
	if err != nil {
		return fmt.Errorf("failed to extract HTML: %w", err)
	}

	// Process and output content
	if err := converter.Process(html, outputFile); err != nil {
		return err
	}

	return nil
}
