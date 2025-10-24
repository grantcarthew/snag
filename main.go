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

	"github.com/urfave/cli/v2"
)

// version can be set via ldflags at build time
var version = "dev"

const (
	// Output format constants for --format flag
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
   visible browsers for authenticated sessions. Output formats: Markdown, HTML, text, PDF, or PNG.

   The perfect companion for AI agents to gain context from web pages.`,
		ArgsUsage: "[url...]",
		Flags: []cli.Flag{
			// Input Control
			&cli.StringFlag{
				Name:  "url-file",
				Usage: "Read URLs from `FILE` (one per line, supports comments)",
			},

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
				Usage:   "Output `FORMAT`: md | html | text | pdf | png",
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

	// Collect URLs from --url-file and command line arguments
	var urls []string

	// Load URLs from file if --url-file is provided
	if urlFile := strings.TrimSpace(c.String("url-file")); urlFile != "" {
		fileURLs, err := loadURLsFromFile(urlFile)
		if err != nil {
			return err
		}
		urls = append(urls, fileURLs...)
	}

	// Add command-line URL arguments (trim whitespace)
	for _, arg := range c.Args().Slice() {
		urls = append(urls, strings.TrimSpace(arg))
	}

	// Validate that no flags are mixed with URLs (common user error)
	for _, arg := range urls {
		if strings.HasPrefix(arg, "-") {
			logger.Error("Flags must come before URL arguments")
			logger.ErrorWithSuggestion(
				fmt.Sprintf("Found '%s' after URLs - flags must be specified before URLs", arg),
				"snag --force-headless -d ./output example.com go.dev",
			)
			return fmt.Errorf("invalid argument order: flags must come before URLs")
		}
	}

	// Handle --list-tabs flag (list tabs and exit)
	// Note: --list-tabs overrides URL arguments if both are present (URLs are ignored)
	if c.Bool("list-tabs") {
		if len(urls) > 0 {
			logger.Verbose("--list-tabs overrides URL arguments (URLs will be ignored)")
		}
		return handleListTabs(c)
	}

	// Handle --all-tabs flag (process all tabs)
	if c.Bool("all-tabs") {
		// Check for conflicting URL arguments
		if len(urls) > 0 {
			logger.Error("Cannot use both --all-tabs and URL arguments (mutually exclusive content sources)")
			return fmt.Errorf("conflicting flags: --all-tabs and URL arguments")
		}
		return handleAllTabs(c)
	}

	// Handle --tab flag (fetch from existing tab)
	if c.IsSet("tab") {
		// Check for conflicting URL arguments
		if len(urls) > 0 {
			logger.Error("Cannot use both --tab and URL arguments (mutually exclusive content sources)")
			return ErrTabURLConflict
		}
		return handleTabFetch(c)
	}

	// Handle --open-browser flag WITHOUT URLs (just open browser)
	if c.Bool("open-browser") && len(urls) == 0 {
		// Warn if --format was specified but will be ignored
		if c.IsSet("format") {
			logger.Warning("--format ignored with --open-browser (no content fetching)")
		}
		logger.Info("Opening browser...")
		bm := NewBrowserManager(BrowserOptions{
			Port:        c.Int("port"),
			OpenBrowser: true,
		})
		return bm.OpenBrowserOnly()
	}

	// Validate that at least one URL was provided
	if len(urls) == 0 {
		logger.Error("No URLs provided")
		logger.ErrorWithSuggestion("Provide URLs as arguments or use --url-file", "snag <url> or snag --url-file urls.txt")
		return ErrNoValidURLs
	}

	// Handle --open-browser WITH URLs (open URLs in tabs, no output)
	if c.Bool("open-browser") && len(urls) > 0 {
		return handleOpenURLsInBrowser(c, urls)
	}

	// Route based on URL count
	if len(urls) == 1 {
		// Single URL - use existing single URL flow
		urlStr := urls[0]

		// Validate and normalize URL
		validatedURL, err := validateURL(urlStr)
		if err != nil {
			return err
		}

		logger.Verbose("Target URL: %s", validatedURL)

		// Normalize format (handles case-insensitive input and aliases)
		format := normalizeFormat(c.String("format"))

		// Extract configuration from flags
		config := &Config{
			URL:           validatedURL,
			OutputFile:    strings.TrimSpace(c.String("output")),
			OutputDir:     strings.TrimSpace(c.String("output-dir")),
			Format:        format,
			Timeout:       c.Int("timeout"),
			WaitFor:       strings.TrimSpace(c.String("wait-for")),
			Port:          c.Int("port"),
			CloseTab:      c.Bool("close-tab"),
			ForceHeadless: c.Bool("force-headless"),
			OpenBrowser:   c.Bool("open-browser"),
			UserAgent:     strings.TrimSpace(c.String("user-agent")),
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
		if c.IsSet("output") || config.OutputFile != "" {
			if err := validateOutputPath(config.OutputFile); err != nil {
				return err
			}
			// Check for extension mismatch and warn (non-blocking)
			checkExtensionMismatch(config.OutputFile, config.Format)
		}

		// Handle --output-dir empty string: default to current directory
		if c.IsSet("output-dir") && config.OutputDir == "" {
			config.OutputDir = "."
		}

		// Validate output directory if provided
		if config.OutputDir != "" {
			if err := validateDirectory(config.OutputDir); err != nil {
				return err
			}
		}

		logger.Verbose("Configuration: format=%s, timeout=%ds, port=%d", config.Format, config.Timeout, config.Port)

		// Execute the snag operation
		return snag(config)
	}

	// Multiple URLs (2+) - use batch processing
	return handleMultipleURLs(c, urls)
}
