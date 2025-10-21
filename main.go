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
	"strconv"
	"syscall"
	"time"

	"github.com/go-rod/rod"
	"github.com/urfave/cli/v2"
)

// version can be set via ldflags at build time
var version = "dev"

const (
	// Output format constants
	FormatMarkdown = "markdown"
	FormatHTML     = "html"
)

var (
	logger         *Logger
	browserManager *BrowserManager

	// Valid output formats
	validFormats = map[string]bool{
		FormatMarkdown: true,
		FormatHTML:     true,
	}
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
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "Output `FORMAT`: markdown | html",
				Value:   FormatMarkdown,
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
		Format:        c.String("format"),
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

	logger.Verbose("Configuration: format=%s, timeout=%ds, port=%d", config.Format, config.Timeout, config.Port)

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

	// Fetch the page
	html, err := fetcher.Fetch(FetchOptions{
		URL:     config.URL,
		Timeout: config.Timeout,
		WaitFor: config.WaitFor,
	})
	if err != nil {
		return err
	}

	// Create content converter
	converter := NewContentConverter(config.Format)

	// Process and output content
	if err := converter.Process(html, config.OutputFile); err != nil {
		return err
	}

	return nil
}

// Config holds the application configuration
type Config struct {
	URL           string
	OutputFile    string
	Format        string
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

	// Extract HTML content from the current page state
	html, err := page.HTML()
	if err != nil {
		return fmt.Errorf("failed to extract HTML: %w", err)
	}

	// Create content converter
	converter := NewContentConverter(format)

	// Process and output content
	if err := converter.Process(html, outputFile); err != nil {
		return err
	}

	return nil
}
