// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"os"

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
	logger *Logger

	// Valid output formats
	validFormats = map[string]bool{
		FormatMarkdown: true,
		FormatHTML:     true,
	}
)

func main() {
	app := &cli.App{
		Name:            "snag",
		Usage:           "Intelligently fetch web page content with browser engine",
		UsageText:       "snag [options] <url>",
		Version:         version,
		HideVersion:     true,
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
				Name:  "format",
				Usage: "Output format: markdown | html",
				Value: FormatMarkdown,
			},

			// Page Loading
			&cli.IntFlag{
				Name:    "timeout",
				Aliases: []string{"t"},
				Usage:   "Page load timeout in `SECONDS`",
				Value:   30,
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
		return err
	}

	// Ensure browser cleanup
	defer func() {
		if config.CloseTab {
			logger.Verbose("Cleanup: closing tab and browser if needed")
		}
		bm.Close()
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
