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
			&cli.StringFlag{
				Name:  "user-data-dir",
				Usage: "Custom Chrome user data `DIRECTORY` (for session isolation)",
			},
		},
		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(c *cli.Context) error {
	// Initialize logger based on flags (last flag wins)
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

	// Trim string flags once at the source
	outputFile := strings.TrimSpace(c.String("output"))
	outputDir := strings.TrimSpace(c.String("output-dir"))

	if urlFile := strings.TrimSpace(c.String("url-file")); urlFile != "" {
		fileURLs, err := loadURLsFromFile(urlFile)
		if err != nil {
			return err
		}
		urls = append(urls, fileURLs...)
	}

	for _, arg := range c.Args().Slice() {
		trimmedArg := strings.TrimSpace(arg)
		if trimmedArg != "" {
			urls = append(urls, trimmedArg)
		}
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

	// --list-tabs overrides URL arguments if both are present
	if c.Bool("list-tabs") {
		if len(urls) > 0 {
			logger.Verbose("--list-tabs overrides URL arguments (URLs will be ignored)")
		}
		return handleListTabs(c)
	}

	if c.Bool("all-tabs") {
		if len(urls) > 0 {
			logger.Error("Cannot use both --all-tabs and URL arguments (mutually exclusive content sources)")
			return fmt.Errorf("conflicting flags: --all-tabs and URL arguments")
		}
		if c.Bool("force-headless") {
			logger.Error("Cannot use --force-headless with --all-tabs (--all-tabs requires existing browser connection)")
			return fmt.Errorf("conflicting flags: --force-headless and --all-tabs")
		}
		if outputFile != "" {
			logger.Error("Cannot use --output with --all-tabs (multiple outputs). Use --output-dir instead")
			return ErrOutputFlagConflict
		}
		if c.Bool("open-browser") {
			logger.Warning("--all-tabs ignored with --open-browser (no content fetching)")
		}
		return handleAllTabs(c)
	}

	if c.IsSet("tab") {
		if len(urls) > 0 {
			logger.Error("Cannot use --tab with URL argument (mutually exclusive content sources)")
			return ErrTabURLConflict
		}
		if c.Bool("all-tabs") {
			logger.Error("Cannot use both --tab and --all-tabs (mutually exclusive content sources)")
			return fmt.Errorf("conflicting flags: --tab and --all-tabs")
		}
		if c.Bool("force-headless") {
			logger.Error("Cannot use --force-headless with --tab (--tab requires existing browser connection)")
			return fmt.Errorf("conflicting flags: --force-headless and --tab")
		}
		if c.Bool("open-browser") {
			logger.Warning("--tab ignored with --open-browser (no content fetching)")
		}
		return handleTabFetch(c)
	}

	if c.Bool("open-browser") && c.Bool("force-headless") {
		logger.Error("Cannot use both --force-headless and --open-browser (conflicting modes)")
		return fmt.Errorf("conflicting flags: --force-headless and --open-browser")
	}

	if c.Bool("open-browser") && len(urls) == 0 {
		if c.IsSet("format") {
			logger.Warning("--format ignored with --open-browser (no content fetching)")
		}
		if c.IsSet("output") {
			logger.Warning("--output ignored with --open-browser (no content fetching)")
		}
		if c.IsSet("output-dir") {
			logger.Warning("--output-dir ignored with --open-browser (no content fetching)")
		}
		if c.IsSet("timeout") {
			logger.Warning("--timeout ignored with --open-browser (no content fetching)")
		}
		if c.IsSet("wait-for") {
			logger.Warning("--wait-for ignored with --open-browser (no content fetching)")
		}
		if c.IsSet("user-agent") {
			logger.Warning("--user-agent ignored with --open-browser (no navigation)")
		}
		if c.Bool("close-tab") {
			logger.Warning("--close-tab ignored with --open-browser (no content fetching)")
		}

		userDataDir := ""
		if c.IsSet("user-data-dir") {
			validatedDir, err := validateUserDataDir(c.String("user-data-dir"))
			if err != nil {
				return err
			}
			userDataDir = validatedDir
		}

		logger.Info("Opening browser...")
		bm := NewBrowserManager(BrowserOptions{
			Port:        c.Int("port"),
			OpenBrowser: true,
			UserDataDir: userDataDir,
		})
		return bm.OpenBrowserOnly()
	}

	if len(urls) == 0 {
		logger.Error("No URLs provided")
		logger.ErrorWithSuggestion("Provide URLs as arguments or use --url-file", "snag <url> or snag --url-file urls.txt")
		return ErrNoValidURLs
	}

	if c.Bool("open-browser") && len(urls) > 0 {
		return handleOpenURLsInBrowser(c, urls)
	}

	if len(urls) == 1 {
		urlStr := urls[0]

		validatedURL, err := validateURL(urlStr)
		if err != nil {
			return err
		}

		logger.Verbose("Target URL: %s", validatedURL)

		format := normalizeFormat(c.String("format"))

		userDataDir := ""
		if c.IsSet("user-data-dir") {
			validatedDir, err := validateUserDataDir(c.String("user-data-dir"))
			if err != nil {
				return err
			}
			userDataDir = validatedDir
		}

		userAgent := ""
		if c.IsSet("user-agent") {
			userAgent = validateUserAgent(c.String("user-agent"))
		}

		waitFor := ""
		if c.IsSet("wait-for") {
			waitFor = validateWaitFor(c.String("wait-for"))
		}

		config := &Config{
			URL:           validatedURL,
			OutputFile:    outputFile,
			OutputDir:     outputDir,
			Format:        format,
			Timeout:       c.Int("timeout"),
			WaitFor:       waitFor,
			Port:          c.Int("port"),
			CloseTab:      c.Bool("close-tab"),
			ForceHeadless: c.Bool("force-headless"),
			OpenBrowser:   c.Bool("open-browser"),
			UserAgent:     userAgent,
			UserDataDir:   userDataDir,
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

		if c.IsSet("output") || config.OutputFile != "" {
			if err := validateOutputPath(config.OutputFile); err != nil {
				return err
			}
			checkExtensionMismatch(config.OutputFile, config.Format)
		}

		if c.IsSet("output-dir") && config.OutputDir == "" {
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

	return handleMultipleURLs(c, urls)
}
