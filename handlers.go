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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/urfave/cli/v2"
)

// Config holds the application configuration
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
}

// snag is the main function that orchestrates the web page fetching
func snag(config *Config) error {
	// Create browser manager
	bm := NewBrowserManager(BrowserOptions{
		Port:          config.Port,
		ForceHeadless: config.ForceHeadless,
		OpenBrowser:   config.OpenBrowser,
		UserAgent:     config.UserAgent,
	})

	// Assign to global for signal handler access
	browserManager = bm

	// Connect to browser
	_, err := bm.Connect()
	if err != nil {
		if errors.Is(err, ErrBrowserNotFound) {
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
	_, err = fetcher.Fetch(FetchOptions{
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

		config.OutputFile, err = generateOutputFilename(
			info.Title, config.URL, config.Format,
			time.Now(), config.OutputDir,
		)
		if err != nil {
			return err
		}
	}

	// For binary formats without -o or -d: auto-generate filename in current directory
	// Binary formats (PDF, PNG) should NEVER output to stdout (corrupts terminal)
	if config.OutputFile == "" && (config.Format == FormatPDF || config.Format == FormatPNG) {
		// Get page info for title
		info, err := page.Info()
		if err != nil {
			return fmt.Errorf("failed to get page info: %w", err)
		}

		config.OutputFile, err = generateOutputFilename(
			info.Title, config.URL, config.Format,
			time.Now(), ".",
		)
		if err != nil {
			return err
		}
		logger.Info("Auto-generated filename: %s", config.OutputFile)
	}

	// Process page content and output in requested format
	return processPageContent(page, config.Format, config.OutputFile)
}

// processPageContent handles format conversion for all output types
// Returns error if processing fails
func processPageContent(page *rod.Page, format string, outputFile string) error {
	// Create content converter for specified format
	converter := NewContentConverter(format)

	// Handle binary formats (PDF, PNG) that need the page object
	if format == FormatPDF || format == FormatPNG {
		return converter.ProcessPage(page, outputFile)
	}

	// For text formats, extract HTML and process
	html, err := page.HTML()
	if err != nil {
		return fmt.Errorf("failed to extract HTML: %w", err)
	}

	return converter.Process(html, outputFile)
}

// generateOutputFilename creates an auto-generated filename for binary formats
// Takes title, URL, format info and returns full path with conflict resolution
func generateOutputFilename(title, url, format string,
	timestamp time.Time, outputDir string) (string, error) {
	// Generate filename
	filename := GenerateFilename(title, format, timestamp, url)

	// Resolve conflicts in directory
	finalFilename, err := ResolveConflict(outputDir, filename)
	if err != nil {
		return "", fmt.Errorf("failed to resolve filename conflict: %w", err)
	}

	// Return full path
	return filepath.Join(outputDir, finalFilename), nil
}

// connectToExistingBrowser creates a browser manager and connects to existing browser
// Sets global browserManager for signal handling and returns the manager
func connectToExistingBrowser(port int) (*BrowserManager, error) {
	// Create browser manager in connect-only mode
	bm := NewBrowserManager(BrowserOptions{
		Port: port,
	})

	// Assign to global for signal handler access
	browserManager = bm

	// Connect to existing browser
	browser, err := bm.connectToExisting()
	if err != nil {
		browserManager = nil // Clear global on error
		logger.Error("No browser instance running with remote debugging")
		logger.ErrorWithSuggestion(
			"Start Chrome with remote debugging enabled",
			fmt.Sprintf("chrome --remote-debugging-port=%d", port),
		)
		logger.Info("Or run: snag --open-browser")
		return nil, ErrNoBrowserRunning
	}

	// Assign browser to manager
	bm.browser = browser

	return bm, nil
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

// displayTabListOnError displays available tabs to stderr as helpful error context
func displayTabListOnError(bm *BrowserManager) {
	if tabs, listErr := bm.ListTabs(); listErr == nil {
		fmt.Fprintln(os.Stderr, "")
		displayTabList(tabs, os.Stderr)
		fmt.Fprintln(os.Stderr, "")
	}
}

// handleAllTabs processes all open browser tabs with auto-generated filenames
func handleAllTabs(c *cli.Context) error {
	// Extract configuration from flags
	format := normalizeFormat(c.String("format"))
	timeout := c.Int("timeout")
	waitFor := strings.TrimSpace(c.String("wait-for"))
	outputDir := strings.TrimSpace(c.String("output-dir"))
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

	// Connect to existing browser with remote debugging enabled
	bm, err := connectToExistingBrowser(c.Int("port"))
	if err != nil {
		return err
	}
	defer func() { browserManager = nil }()

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

		// Generate output filename with conflict resolution
		outputPath, err := generateOutputFilename(
			tab.Title, tab.URL, format,
			timestamp, outputDir,
		)
		if err != nil {
			logger.Error("[%d/%d] Failed to generate filename: %v", tabNum, len(tabs), err)
			failureCount++
			continue
		}

		// Process page content and output in requested format
		if err := processPageContent(page, format, outputPath); err != nil {
			logger.Error("[%d/%d] Failed to process content: %v", tabNum, len(tabs), err)
			failureCount++
			continue
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
	// Connect to existing browser with remote debugging enabled
	bm, err := connectToExistingBrowser(c.Int("port"))
	if err != nil {
		return err
	}

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
	tabValue := strings.TrimSpace(c.String("tab"))
	if tabValue == "" {
		return fmt.Errorf("--tab flag requires a value")
	}

	// Connect to existing browser with remote debugging enabled
	bm, err := connectToExistingBrowser(c.Int("port"))
	if err != nil {
		return err
	}
	defer func() { browserManager = nil }()

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
				displayTabListOnError(bm)
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
				displayTabListOnError(bm)
			}
			return err
		}
		logger.Success("Connected to tab matching pattern: %s", tabValue)
	}

	// Extract configuration from flags
	format := normalizeFormat(c.String("format"))
	timeout := c.Int("timeout")
	waitFor := strings.TrimSpace(c.String("wait-for"))
	outputFile := strings.TrimSpace(c.String("output"))

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
	// Binary formats (PDF, PNG) should NEVER output to stdout (corrupts terminal)
	if outputFile == "" && (format == FormatPDF || format == FormatPNG) {
		outputFile, err = generateOutputFilename(
			info.Title, info.URL, format,
			time.Now(), ".",
		)
		if err != nil {
			return err
		}
		logger.Info("Auto-generated filename: %s", outputFile)
	}

	// Process page content and output in requested format
	return processPageContent(page, format, outputFile)
}

// handleOpenURLsInBrowser opens multiple URLs in browser tabs without fetching content
// This implements the --open-browser behavior with URLs (just opens, no output)
func handleOpenURLsInBrowser(c *cli.Context, urls []string) error {
	logger.Info("Opening %d URLs in browser...", len(urls))

	// Create browser manager in visible mode
	bm := NewBrowserManager(BrowserOptions{
		Port:          c.Int("port"),
		OpenBrowser:   true,
		ForceHeadless: false,
		UserAgent:     strings.TrimSpace(c.String("user-agent")),
	})

	// Assign to global for signal handler access
	browserManager = bm

	// Connect to browser
	_, err := bm.Connect()
	if err != nil {
		browserManager = nil
		return err
	}

	// Open each URL in a new tab
	for i, urlStr := range urls {
		current := i + 1
		logger.Info("[%d/%d] Opening: %s", current, len(urls), urlStr)

		// Validate URL
		validatedURL, err := validateURL(urlStr)
		if err != nil {
			logger.Warning("[%d/%d] Invalid URL - skipping: %s", current, len(urls), urlStr)
			continue
		}

		// Create new page
		page, err := bm.NewPage()
		if err != nil {
			logger.Error("[%d/%d] Failed to create page: %v", current, len(urls), err)
			continue
		}

		// Navigate to URL (with timeout)
		timeout := c.Int("timeout")
		err = page.Timeout(time.Duration(timeout) * time.Second).Navigate(validatedURL)
		if err != nil {
			logger.Error("[%d/%d] Failed to navigate: %v", current, len(urls), err)
			continue
		}

		logger.Success("[%d/%d] Opened: %s", current, len(urls), validatedURL)
	}

	logger.Success("Browser will remain open with %d tabs", len(urls))
	logger.Info("Use 'snag --list-tabs' to see opened tabs")
	logger.Info("Use 'snag --tab <index>' to fetch content from a tab")

	// Don't close browser - leave it running for user
	return nil
}

// handleMultipleURLs processes multiple URLs with batch fetching
// Follows the same pattern as handleAllTabs but for URL arguments
func handleMultipleURLs(c *cli.Context, urls []string) error {
	// Validate conflicting flags for multiple URLs
	if strings.TrimSpace(c.String("output")) != "" {
		logger.Error("Cannot use --output with multiple URLs")
		logger.Info("Use --output-dir instead for auto-generated filenames")
		return ErrOutputFlagConflict
	}

	if c.Bool("close-tab") {
		logger.Error("Cannot use --close-tab with multiple URLs")
		logger.Info("--close-tab is only supported for single URL fetching")
		return ErrCloseTabMultipleURLs
	}

	// Extract configuration from flags
	format := normalizeFormat(c.String("format"))
	timeout := c.Int("timeout")
	waitFor := strings.TrimSpace(c.String("wait-for"))
	outputDir := strings.TrimSpace(c.String("output-dir"))
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

	// Create browser manager
	bm := NewBrowserManager(BrowserOptions{
		Port:          c.Int("port"),
		ForceHeadless: c.Bool("force-headless"),
		UserAgent:     strings.TrimSpace(c.String("user-agent")),
	})

	// Assign to global for signal handler access
	browserManager = bm

	// Connect to browser
	_, err := bm.Connect()
	if err != nil {
		browserManager = nil
		return err
	}

	// Ensure browser cleanup
	defer func() {
		bm.Close()
		browserManager = nil
	}()

	// Create single timestamp for entire batch
	timestamp := time.Now()

	logger.Info("Processing %d URLs...", len(urls))

	// Track success/failure counts
	successCount := 0
	failureCount := 0

	// Process each URL
	for i, urlStr := range urls {
		current := i + 1
		total := len(urls)

		logger.Info("[%d/%d] Fetching: %s", current, total, urlStr)

		// Validate URL
		validatedURL, err := validateURL(urlStr)
		if err != nil {
			logger.Error("[%d/%d] Invalid URL - skipping: %v", current, total, err)
			failureCount++
			continue
		}

		// Create new page
		page, err := bm.NewPage()
		if err != nil {
			logger.Error("[%d/%d] Failed to create page: %v", current, total, err)
			failureCount++
			continue
		}

		// Fetch page content
		fetcher := NewPageFetcher(page, timeout)
		_, err = fetcher.Fetch(FetchOptions{
			URL:     validatedURL,
			Timeout: timeout,
			WaitFor: waitFor,
		})
		if err != nil {
			logger.Error("[%d/%d] Failed to fetch: %v", current, total, err)
			bm.ClosePage(page)
			failureCount++
			continue
		}

		// Get page info for filename generation
		info, err := page.Info()
		if err != nil {
			logger.Error("[%d/%d] Failed to get page info: %v", current, total, err)
			bm.ClosePage(page)
			failureCount++
			continue
		}

		// Generate output filename with conflict resolution
		outputPath, err := generateOutputFilename(
			info.Title, validatedURL, format,
			timestamp, outputDir,
		)
		if err != nil {
			logger.Error("[%d/%d] Failed to generate filename: %v", current, total, err)
			bm.ClosePage(page)
			failureCount++
			continue
		}

		// Process page content and output in requested format
		if err := processPageContent(page, format, outputPath); err != nil {
			logger.Error("[%d/%d] Failed to save content: %v", current, total, err)
			bm.ClosePage(page)
			failureCount++
			continue
		}

		// Close page in headless mode
		if bm.launchedHeadless {
			bm.ClosePage(page)
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
