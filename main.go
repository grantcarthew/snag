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

const (
	version = "1.0.0"

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
		Name:    "snag",
		Usage:   "Intelligently fetch web page content with browser engine",
		Version: version,
		Authors: []*cli.Author{
			{
				Name:  "Grant Carthew",
				Email: "grant@carthew.net",
			},
		},
		Description: `snag fetches web page content using Chrome/Chromium via the Chrome DevTools Protocol.
   It can connect to existing browser sessions, launch headless browsers, or open
   visible browsers for authenticated sessions. Output can be Markdown or HTML.`,
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
				Usage: "Output format: markdown (default) | html",
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
				Aliases: []string{"ob"},
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

	// Handle --open-browser flag (no URL required)
	if c.Bool("open-browser") {
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
		UserAgent:     c.String("user-agent"),
	}

	// Validate format
	if !validFormats[config.Format] {
		logger.Error("Invalid format: %s", config.Format)
		logger.ErrorWithSuggestion(
			fmt.Sprintf("Format must be '%s' or '%s'", FormatMarkdown, FormatHTML),
			fmt.Sprintf("snag <url> --format %s", FormatMarkdown),
		)
		return fmt.Errorf("invalid format: %s", config.Format)
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
	UserAgent     string
}
