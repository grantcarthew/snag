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
		logger.Error("No browser found. Try running 'snag --open-browser' first")
		return nil, ErrNoBrowserRunning
	}

	// Assign browser to manager
	bm.browser = browser

	return bm, nil
}

// stripURLParams removes query parameters and hash fragments from a URL
// Returns clean URL with only scheme, domain, and path
func stripURLParams(url string) string {
	// Find position of query params
	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}

	// Find position of hash fragment
	if idx := strings.Index(url, "#"); idx != -1 {
		url = url[:idx]
	}

	return url
}

// formatTabLine formats a single tab line for display with optional truncation
// Normal mode: "  [N] URL (Title)" with 120 char limit
// Verbose mode: "  [N] full-url - Title" with no truncation
func formatTabLine(index int, title, url string, maxLength int, verbose bool) string {
	if verbose {
		// Verbose mode: show full URL and title, no truncation
		if title == "" {
			return fmt.Sprintf("  [%d] %s", index, url)
		}
		return fmt.Sprintf("  [%d] %s - %s", index, url, title)
	}

	// Normal mode: clean URL and apply truncation
	cleanURL := stripURLParams(url)

	// Calculate available space for URL and title
	// Format: "  [NNN] URL (Title)"
	// Prefix is approximately: "  " + "[" + index + "] " = ~8 chars for index < 1000
	prefix := fmt.Sprintf("  [%d] ", index)
	prefixLen := len(prefix)

	// Maximum URL length: 80 chars
	const maxURLLen = 80
	displayURL := cleanURL
	if len(displayURL) > maxURLLen {
		// Truncate URL if too long
		displayURL = cleanURL[:maxURLLen-3] + "..."
	}

	// Calculate space available for title (in parentheses)
	// Total budget: maxLength, minus prefix, minus URL, minus space, minus parentheses
	titleBudget := maxLength - prefixLen - len(displayURL)
	if title != "" {
		titleBudget -= 3 // Account for space and parentheses: " (" and ")"
	}

	// Format the line
	if title == "" {
		// No title: just show URL
		return fmt.Sprintf("%s%s", prefix, displayURL)
	}

	// Truncate title if needed
	if len(title) > titleBudget && titleBudget > 3 {
		title = title[:titleBudget-3] + "..."
	}

	return fmt.Sprintf("%s%s (%s)", prefix, displayURL, title)
}

// displayTabList formats and displays a list of tabs to the specified writer
// verbose controls whether to show full URLs with query params (true) or clean, truncated display (false)
func displayTabList(tabs []TabInfo, w io.Writer, verbose bool) {
	if len(tabs) == 0 {
		fmt.Fprintf(w, "No tabs open in browser\n")
		return
	}

	fmt.Fprintf(w, "Available tabs in browser (%d tabs, sorted by URL):\n", len(tabs))
	for _, tab := range tabs {
		line := formatTabLine(tab.Index, tab.Title, tab.URL, 120, verbose)
		fmt.Fprintf(w, "%s\n", line)
	}
}

// displayTabListOnError displays available tabs to stderr as helpful error context
// Always uses non-verbose mode (clean, truncated display)
func displayTabListOnError(bm *BrowserManager) {
	if tabs, listErr := bm.ListTabs(); listErr == nil {
		fmt.Fprintln(os.Stderr, "")
		displayTabList(tabs, os.Stderr, false) // Always non-verbose for error context
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

	// Check verbose flag for full URL display
	verbose := c.Bool("verbose")

	// Display tabs to stdout
	displayTabList(tabs, os.Stdout, verbose)

	return nil
}

// handleTabRange processes a range of tabs with auto-generated filenames
func handleTabRange(c *cli.Context, bm *BrowserManager, start, end int) error {
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

	// Get tabs in range
	pages, err := bm.GetTabsByRange(start, end)
	if err != nil {
		// Display available tabs on error
		logger.Error("Failed to get tab range: %v", err)
		logger.Info("Run 'snag --list-tabs' to see available tabs")
		displayTabListOnError(bm)
		return err
	}

	// Create single timestamp for entire batch
	timestamp := time.Now()

	logger.Info("Processing %d tabs from range [%d-%d]...", len(pages), start, end)

	// Track success/failure counts
	successCount := 0
	failureCount := 0

	// Process each tab in range
	for i, page := range pages {
		tabNum := start + i // Actual tab number (1-based)
		totalTabs := len(pages)
		current := i + 1 // Current position in range (1-based)

		// Get page info
		info, err := page.Info()
		if err != nil {
			logger.Error("[%d/%d] Failed to get tab info for tab [%d]: %v", current, totalTabs, tabNum, err)
			failureCount++
			continue
		}

		logger.Info("[%d/%d] Processing tab [%d]: %s", current, totalTabs, tabNum, info.URL)

		// Wait for selector if specified
		if waitFor != "" {
			err := waitForSelector(page, waitFor, time.Duration(timeout)*time.Second)
			if err != nil {
				logger.Error("[%d/%d] Wait failed for tab [%d]: %v", current, totalTabs, tabNum, err)
				failureCount++
				continue
			}
		}

		// Generate output filename with conflict resolution
		outputPath, err := generateOutputFilename(
			info.Title, info.URL, format,
			timestamp, outputDir,
		)
		if err != nil {
			logger.Error("[%d/%d] Failed to generate filename for tab [%d]: %v", current, totalTabs, tabNum, err)
			failureCount++
			continue
		}

		// Process page content and output in requested format
		if err := processPageContent(page, format, outputPath); err != nil {
			logger.Error("[%d/%d] Failed to process content for tab [%d]: %v", current, totalTabs, tabNum, err)
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

// handleTabPatternBatch processes multiple tabs matching a pattern with auto-generated filenames
func handleTabPatternBatch(c *cli.Context, pages []*rod.Page, pattern string) error {
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

	// Create single timestamp for entire batch
	timestamp := time.Now()

	logger.Info("Processing %d tabs matching pattern '%s'...", len(pages), pattern)

	// Track success/failure counts
	successCount := 0
	failureCount := 0

	// Process each matched tab
	for i, page := range pages {
		totalTabs := len(pages)
		current := i + 1 // Current position (1-based)

		// Get page info
		info, err := page.Info()
		if err != nil {
			logger.Error("[%d/%d] Failed to get tab info: %v", current, totalTabs, err)
			failureCount++
			continue
		}

		logger.Info("[%d/%d] Processing: %s", current, totalTabs, info.URL)

		// Wait for selector if specified
		if waitFor != "" {
			err := waitForSelector(page, waitFor, time.Duration(timeout)*time.Second)
			if err != nil {
				logger.Error("[%d/%d] Wait failed: %v", current, totalTabs, err)
				failureCount++
				continue
			}
		}

		// Generate output filename with conflict resolution
		outputPath, err := generateOutputFilename(
			info.Title, info.URL, format,
			timestamp, outputDir,
		)
		if err != nil {
			logger.Error("[%d/%d] Failed to generate filename: %v", current, totalTabs, err)
			failureCount++
			continue
		}

		// Process page content and output in requested format
		if err := processPageContent(page, format, outputPath); err != nil {
			logger.Error("[%d/%d] Failed to process content: %v", current, totalTabs, err)
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

// handleTabFetch fetches content from an existing tab by index
func handleTabFetch(c *cli.Context) error {
	// Connect to existing browser first (needed for displaying tabs on error)
	bm, err := connectToExistingBrowser(c.Int("port"))
	if err != nil {
		return err
	}
	defer func() { browserManager = nil }()

	// Get the tab value
	tabValue := strings.TrimSpace(c.String("tab"))
	if tabValue == "" {
		logger.Error("Tab pattern cannot be empty")
		logger.Info("Run 'snag --list-tabs' to see available tabs")
		displayTabListOnError(bm)
		return fmt.Errorf("tab pattern cannot be empty")
	}

	// Warn about ignored flags
	if c.IsSet("user-agent") {
		logger.Warning("--user-agent is ignored with --tab (cannot change existing tab's user agent)")
	}
	if c.IsSet("user-data-dir") {
		logger.Warning("--user-data-dir ignored when connecting to existing browser")
	}
	if c.IsSet("timeout") && !c.IsSet("wait-for") {
		logger.Warning("--timeout is ignored without --wait-for when using --tab")
	}

	// Check if tab value is a range (N-M format)
	// Only treat as range if both parts are valid positive integers
	if strings.Contains(tabValue, "-") {
		parts := strings.SplitN(tabValue, "-", 2)
		if len(parts) == 2 {
			start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))

			// If both parts parse as positive integers, it's a valid range
			if err1 == nil && err2 == nil && start > 0 && end > 0 {
				// Valid range detected - validate --output flag
				if c.IsSet("output") {
					logger.Error("Cannot use --output with multiple tabs. Use --output-dir instead")
					return ErrOutputFlagConflict
				}

				// Process as range
				return handleTabRange(c, bm, start, end)
			}
			// If one or both parts don't parse as integers, treat as pattern (fall through)
			// This allows patterns like "my-url-pattern" to work correctly
		}
	}

	// Determine if tab value is an integer index or a pattern
	var page *rod.Page
	var multipleMatches bool
	var matchedPages []*rod.Page

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
		logger.Success("Connected to tab [%d] from sorted order (by URL)", tabIndex)
	} else {
		// Not an integer - treat as pattern
		logger.Verbose("Fetching from tab matching pattern: %s", tabValue)
		matchedPages, err = bm.GetTabsByPattern(tabValue)
		if err != nil {
			if errors.Is(err, ErrNoTabMatch) {
				logger.Error("No tab matches pattern '%s'", tabValue)
				logger.Info("Run 'snag --list-tabs' to see available tabs")
				displayTabListOnError(bm)
			}
			return err
		}

		// Check if we have single or multiple matches
		if len(matchedPages) == 1 {
			// Single match - use current single-page flow
			page = matchedPages[0]
			logger.Success("Connected to tab matching pattern: %s", tabValue)
		} else {
			// Multiple matches - validate --output flag and use batch processing
			multipleMatches = true
			if c.IsSet("output") {
				logger.Error("Cannot use --output with multiple tabs. Use --output-dir instead")
				logger.Info("Pattern '%s' matched %d tabs", tabValue, len(matchedPages))
				return ErrOutputFlagConflict
			}
			logger.Info("Pattern '%s' matched %d tabs", tabValue, len(matchedPages))
		}
	}

	// If multiple matches, use batch processing
	if multipleMatches {
		return handleTabPatternBatch(c, matchedPages, tabValue)
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
		// Check for extension mismatch and warn (non-blocking)
		checkExtensionMismatch(outputFile, format)
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
		logger.Error("Cannot use --output with multiple content sources. Use --output-dir instead")
		return ErrOutputFlagConflict
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

		// Close page in headless mode or if --close-tab is set
		if bm.launchedHeadless || c.Bool("close-tab") {
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
