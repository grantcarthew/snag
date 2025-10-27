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

func snag(config *Config) error {
	bm := NewBrowserManager(BrowserOptions{
		Port:          config.Port,
		ForceHeadless: config.ForceHeadless,
		OpenBrowser:   config.OpenBrowser,
		UserAgent:     config.UserAgent,
		UserDataDir:   config.UserDataDir,
	})

	browserManager = bm

	_, err := bm.Connect()
	if err != nil {
		if errors.Is(err, ErrBrowserNotFound) {
			logger.Error("No Chromium-based browser found")
			logger.ErrorWithSuggestion(
				"Install Chrome, Chromium, Edge, or Brave to use snag",
				"brew install --cask google-chrome",
			)
		}
		browserManager = nil
		return err
	}

	// Ensure browser cleanup
	defer func() {
		if config.CloseTab {
			logger.Verbose("Cleanup: closing tab and browser if needed")
		}
		bm.Close()
		browserManager = nil
	}()

	page, err := bm.NewPage()
	if err != nil {
		return err
	}

	if config.CloseTab {
		defer bm.ClosePage(page)
	}

	fetcher := NewPageFetcher(page, config.Timeout)

	_, err = fetcher.Fetch(FetchOptions{
		URL:     config.URL,
		Timeout: config.Timeout,
		WaitFor: config.WaitFor,
	})
	if err != nil {
		return err
	}

	if config.OutputDir != "" {
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

	return processPageContent(page, config.Format, config.OutputFile)
}

// processPageContent handles format conversion for all output types
func processPageContent(page *rod.Page, format string, outputFile string) error {
	converter := NewContentConverter(format)

	// Handle binary formats (PDF, PNG) that need the page object
	if format == FormatPDF || format == FormatPNG {
		return converter.ProcessPage(page, outputFile)
	}

	html, err := page.HTML()
	if err != nil {
		return fmt.Errorf("failed to extract HTML: %w", err)
	}

	return converter.Process(html, outputFile)
}

func generateOutputFilename(title, url, format string,
	timestamp time.Time, outputDir string) (string, error) {
	filename := GenerateFilename(title, format, timestamp, url)

	finalFilename, err := ResolveConflict(outputDir, filename)
	if err != nil {
		return "", fmt.Errorf("failed to resolve filename conflict: %w", err)
	}

	return filepath.Join(outputDir, finalFilename), nil
}

func connectToExistingBrowser(port int) (*BrowserManager, error) {
	bm := NewBrowserManager(BrowserOptions{
		Port: port,
	})

	browserManager = bm

	browser, err := bm.connectToExisting()
	if err != nil {
		browserManager = nil
		logger.Error("No browser found. Try running 'snag --open-browser' first")
		return nil, ErrNoBrowserRunning
	}

	bm.browser = browser

	return bm, nil
}

func stripURLParams(url string) string {
	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}

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
		if title == "" {
			return fmt.Sprintf("  [%d] %s", index, url)
		}
		return fmt.Sprintf("  [%d] %s - %s", index, url, title)
	}

	cleanURL := stripURLParams(url)

	prefix := fmt.Sprintf("  [%d] ", index)
	prefixLen := len(prefix)

	const maxURLLen = 80
	displayURL := cleanURL
	if len(displayURL) > maxURLLen {
		displayURL = cleanURL[:maxURLLen-3] + "..."
	}

	titleBudget := maxLength - prefixLen - len(displayURL)
	if title != "" {
		titleBudget -= 3
	}

	if title == "" {
		return fmt.Sprintf("%s%s", prefix, displayURL)
	}

	if len(title) > titleBudget && titleBudget > 3 {
		title = title[:titleBudget-3] + "..."
	}

	return fmt.Sprintf("%s%s (%s)", prefix, displayURL, title)
}

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

func displayTabListOnError(bm *BrowserManager) {
	if tabs, listErr := bm.ListTabs(); listErr == nil {
		fmt.Fprintln(os.Stderr, "")
		displayTabList(tabs, os.Stderr, false)
		fmt.Fprintln(os.Stderr, "")
	}
}

func handleAllTabs(c *cli.Context) error {
	outputDir := strings.TrimSpace(c.String("output-dir"))

	format := normalizeFormat(c.String("format"))
	timeout := c.Int("timeout")
	waitFor := validateWaitFor(c.String("wait-for"))
	closeTab := c.Bool("close-tab")
	if outputDir == "" {
		outputDir = "."
	}

	if c.IsSet("user-agent") {
		logger.Warning("--user-agent is ignored with --all-tabs (cannot change existing tabs' user agents)")
	}
	if c.IsSet("user-data-dir") {
		logger.Warning("--user-data-dir ignored when connecting to existing browser")
	}
	if c.IsSet("timeout") && waitFor == "" {
		logger.Warning("--timeout is ignored without --wait-for when using --all-tabs")
	}

	if err := validateFormat(format); err != nil {
		return err
	}

	if err := validateTimeout(timeout); err != nil {
		return err
	}

	if err := validateDirectory(outputDir); err != nil {
		return err
	}

	bm, err := connectToExistingBrowser(c.Int("port"))
	if err != nil {
		return err
	}
	defer func() { browserManager = nil }()

	tabs, err := bm.ListTabs()
	if err != nil {
		return err
	}

	if len(tabs) == 0 {
		logger.Info("No tabs open in browser")
		return nil
	}

	timestamp := time.Now()

	logger.Info("Processing %d tabs...", len(tabs))

	successCount := 0
	failureCount := 0

	for i, tab := range tabs {
		tabNum := i + 1

		if isNonFetchableURL(tab.URL) {
			logger.Warning("[%d/%d] Skipping tab: %s (not fetchable)", tabNum, len(tabs), tab.URL)
			continue
		}

		logger.Info("[%d/%d] Processing: %s", tabNum, len(tabs), tab.URL)

		page, err := bm.GetTabByIndex(tabNum)
		if err != nil {
			logger.Error("[%d/%d] Failed to get tab: %v", tabNum, len(tabs), err)
			failureCount++
			continue
		}

		if waitFor != "" {
			err := waitForSelector(page, waitFor, time.Duration(timeout)*time.Second)
			if err != nil {
				logger.Error("[%d/%d] Wait failed: %v", tabNum, len(tabs), err)
				failureCount++
				continue
			}
		}

		outputPath, err := generateOutputFilename(
			tab.Title, tab.URL, format,
			timestamp, outputDir,
		)
		if err != nil {
			logger.Error("[%d/%d] Failed to generate filename: %v", tabNum, len(tabs), err)
			failureCount++
			continue
		}

		if err := processPageContent(page, format, outputPath); err != nil {
			logger.Error("[%d/%d] Failed to process content: %v", tabNum, len(tabs), err)
			failureCount++
			if closeTab {
				if err := page.Close(); err != nil {
					logger.Verbose("[%d/%d] Failed to close tab: %v", tabNum, len(tabs), err)
				}
			}
			continue
		}

		successCount++

		if closeTab {
			if tabNum == len(tabs) {
				logger.Info("Closing last tab, browser will close")
			}
			if err := page.Close(); err != nil {
				logger.Verbose("[%d/%d] Failed to close tab: %v", tabNum, len(tabs), err)
			}
		}
	}

	logger.Success("Batch complete: %d succeeded, %d failed", successCount, failureCount)

	if failureCount > 0 {
		return fmt.Errorf("batch processing completed with %d failures", failureCount)
	}

	return nil
}

func handleListTabs(c *cli.Context) error {
	bm, err := connectToExistingBrowser(c.Int("port"))
	if err != nil {
		return err
	}

	tabs, err := bm.ListTabs()
	if err != nil {
		return err
	}

	verbose := c.Bool("verbose")

	displayTabList(tabs, os.Stdout, verbose)

	return nil
}

func handleTabRange(c *cli.Context, bm *BrowserManager, start, end int) error {
	outputDir := strings.TrimSpace(c.String("output-dir"))

	format := normalizeFormat(c.String("format"))
	timeout := c.Int("timeout")
	waitFor := validateWaitFor(c.String("wait-for"))
	if outputDir == "" {
		outputDir = "." // Default to current working directory
	}

	if err := validateFormat(format); err != nil {
		return err
	}

	if err := validateTimeout(timeout); err != nil {
		return err
	}

	if err := validateDirectory(outputDir); err != nil {
		return err
	}

	pages, err := bm.GetTabsByRange(start, end)
	if err != nil {
		logger.Error("Failed to get tab range: %v", err)
		logger.Info("Run 'snag --list-tabs' to see available tabs")
		displayTabListOnError(bm)
		return err
	}

	// Create single timestamp for entire batch
	timestamp := time.Now()

	logger.Info("Processing %d tabs from range [%d-%d]...", len(pages), start, end)

	successCount := 0
	failureCount := 0

	for i, page := range pages {
		tabNum := start + i
		totalTabs := len(pages)
		current := i + 1

		info, err := page.Info()
		if err != nil {
			logger.Error("[%d/%d] Failed to get tab info for tab [%d]: %v", current, totalTabs, tabNum, err)
			failureCount++
			continue
		}

		logger.Info("[%d/%d] Processing tab [%d]: %s", current, totalTabs, tabNum, info.URL)

		if waitFor != "" {
			err := waitForSelector(page, waitFor, time.Duration(timeout)*time.Second)
			if err != nil {
				logger.Error("[%d/%d] Wait failed for tab [%d]: %v", current, totalTabs, tabNum, err)
				failureCount++
				continue
			}
		}

		outputPath, err := generateOutputFilename(
			info.Title, info.URL, format,
			timestamp, outputDir,
		)
		if err != nil {
			logger.Error("[%d/%d] Failed to generate filename for tab [%d]: %v", current, totalTabs, tabNum, err)
			failureCount++
			continue
		}

		if err := processPageContent(page, format, outputPath); err != nil {
			logger.Error("[%d/%d] Failed to process content for tab [%d]: %v", current, totalTabs, tabNum, err)
			failureCount++
			continue
		}

		successCount++
	}

	logger.Success("Batch complete: %d succeeded, %d failed", successCount, failureCount)

	if failureCount > 0 {
		return fmt.Errorf("batch processing completed with %d failures", failureCount)
	}

	return nil
}

func handleTabPatternBatch(c *cli.Context, pages []*rod.Page, pattern string) error {
	outputDir := strings.TrimSpace(c.String("output-dir"))

	format := normalizeFormat(c.String("format"))
	timeout := c.Int("timeout")
	waitFor := validateWaitFor(c.String("wait-for"))
	if outputDir == "" {
		outputDir = "." // Default to current working directory
	}

	if err := validateFormat(format); err != nil {
		return err
	}

	if err := validateTimeout(timeout); err != nil {
		return err
	}

	if err := validateDirectory(outputDir); err != nil {
		return err
	}

	// Create single timestamp for entire batch
	timestamp := time.Now()

	logger.Info("Processing %d tabs matching pattern '%s'...", len(pages), pattern)

	successCount := 0
	failureCount := 0

	for i, page := range pages {
		totalTabs := len(pages)
		current := i + 1

		info, err := page.Info()
		if err != nil {
			logger.Error("[%d/%d] Failed to get tab info: %v", current, totalTabs, err)
			failureCount++
			continue
		}

		logger.Info("[%d/%d] Processing: %s", current, totalTabs, info.URL)

		if waitFor != "" {
			err := waitForSelector(page, waitFor, time.Duration(timeout)*time.Second)
			if err != nil {
				logger.Error("[%d/%d] Wait failed: %v", current, totalTabs, err)
				failureCount++
				continue
			}
		}

		outputPath, err := generateOutputFilename(
			info.Title, info.URL, format,
			timestamp, outputDir,
		)
		if err != nil {
			logger.Error("[%d/%d] Failed to generate filename: %v", current, totalTabs, err)
			failureCount++
			continue
		}

		if err := processPageContent(page, format, outputPath); err != nil {
			logger.Error("[%d/%d] Failed to process content: %v", current, totalTabs, err)
			failureCount++
			continue
		}

		successCount++
	}

	logger.Success("Batch complete: %d succeeded, %d failed", successCount, failureCount)

	if failureCount > 0 {
		return fmt.Errorf("batch processing completed with %d failures", failureCount)
	}

	return nil
}

func handleTabFetch(c *cli.Context) error {
	bm, err := connectToExistingBrowser(c.Int("port"))
	if err != nil {
		return err
	}
	defer func() { browserManager = nil }()

	tabValue := strings.TrimSpace(c.String("tab"))
	if tabValue == "" {
		logger.Error("Tab pattern cannot be empty")
		logger.Info("Run 'snag --list-tabs' to see available tabs")
		displayTabListOnError(bm)
		return fmt.Errorf("tab pattern cannot be empty")
	}

	if c.IsSet("user-agent") {
		logger.Warning("--user-agent is ignored with --tab (cannot change existing tab's user agent)")
	}
	if c.IsSet("user-data-dir") {
		logger.Warning("--user-data-dir ignored when connecting to existing browser")
	}
	if c.IsSet("timeout") && !c.IsSet("wait-for") {
		logger.Warning("--timeout is ignored without --wait-for when using --tab")
	}

	if strings.Contains(tabValue, "-") {
		parts := strings.SplitN(tabValue, "-", 2)
		if len(parts) == 2 {
			start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))

			if err1 == nil && err2 == nil && start > 0 && end > 0 {
				if c.IsSet("output") {
					logger.Error("Cannot use --output with multiple tabs. Use --output-dir instead")
					return ErrOutputFlagConflict
				}

				return handleTabRange(c, bm, start, end)
			}
		}
	}

	var page *rod.Page
	var multipleMatches bool
	var matchedPages []*rod.Page

	if tabIndex, err := strconv.Atoi(tabValue); err == nil {
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

		if len(matchedPages) == 1 {
			page = matchedPages[0]
			logger.Success("Connected to tab matching pattern: %s", tabValue)
		} else {
			multipleMatches = true
			if c.IsSet("output") {
				logger.Error("Cannot use --output with multiple tabs. Use --output-dir instead")
				logger.Info("Pattern '%s' matched %d tabs", tabValue, len(matchedPages))
				return ErrOutputFlagConflict
			}
			logger.Info("Pattern '%s' matched %d tabs", tabValue, len(matchedPages))
		}
	}

	if multipleMatches {
		return handleTabPatternBatch(c, matchedPages, tabValue)
	}

	format := normalizeFormat(c.String("format"))
	timeout := c.Int("timeout")
	waitFor := validateWaitFor(c.String("wait-for"))
	outputFile := strings.TrimSpace(c.String("output"))

	if err := validateFormat(format); err != nil {
		return err
	}

	if err := validateTimeout(timeout); err != nil {
		return err
	}

	if outputFile != "" {
		if err := validateOutputPath(outputFile); err != nil {
			return err
		}
		checkExtensionMismatch(outputFile, format)
	}

	info, err := page.Info()
	if err != nil {
		return fmt.Errorf("failed to get page info: %w", err)
	}

	logger.Info("Fetching content from: %s", info.URL)

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

	return processPageContent(page, format, outputFile)
}

// handleOpenURLsInBrowser opens multiple URLs in browser tabs without fetching content
// This implements the --open-browser behavior with URLs (just opens, no output)
func handleOpenURLsInBrowser(c *cli.Context, urls []string) error {
	// Warn about ignored flags (no content fetching with URLs)
	if c.IsSet("output") {
		logger.Warning("--output ignored with --open-browser (no content fetching)")
	}
	if c.IsSet("output-dir") {
		logger.Warning("--output-dir ignored with --open-browser (no content fetching)")
	}
	if c.IsSet("format") {
		logger.Warning("--format ignored with --open-browser (no content fetching)")
	}
	if c.IsSet("timeout") {
		logger.Warning("--timeout ignored with --open-browser (no content fetching)")
	}
	if c.IsSet("wait-for") {
		logger.Warning("--wait-for ignored with --open-browser (no content fetching)")
	}
	if c.Bool("close-tab") {
		logger.Warning("--close-tab ignored with --open-browser (no content fetching)")
	}

	logger.Info("Opening %d URLs in browser...", len(urls))

	userDataDir := ""
	if c.IsSet("user-data-dir") {
		validatedDir, err := validateUserDataDir(c.String("user-data-dir"))
		if err != nil {
			return err
		}
		userDataDir = validatedDir
	}

	userAgent := validateUserAgent(c.String("user-agent"))

	bm := NewBrowserManager(BrowserOptions{
		Port:          c.Int("port"),
		OpenBrowser:   true,
		ForceHeadless: false,
		UserAgent:     userAgent,
		UserDataDir:   userDataDir,
	})

	browserManager = bm

	_, err := bm.Connect()
	if err != nil {
		browserManager = nil
		return err
	}

	for i, urlStr := range urls {
		current := i + 1
		logger.Info("[%d/%d] Opening: %s", current, len(urls), urlStr)

		validatedURL, err := validateURL(urlStr)
		if err != nil {
			logger.Warning("[%d/%d] Invalid URL - skipping: %s", current, len(urls), urlStr)
			continue
		}

		page, err := bm.NewPage()
		if err != nil {
			logger.Error("[%d/%d] Failed to create page: %v", current, len(urls), err)
			continue
		}

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
	if strings.TrimSpace(c.String("output")) != "" {
		logger.Error("Cannot use --output with multiple content sources. Use --output-dir instead")
		return ErrOutputFlagConflict
	}

	format := normalizeFormat(c.String("format"))
	timeout := c.Int("timeout")
	waitFor := validateWaitFor(c.String("wait-for"))
	outputDir := c.String("output-dir")
	if outputDir == "" {
		outputDir = "." // Default to current working directory
	}

	if err := validateFormat(format); err != nil {
		return err
	}

	if err := validateTimeout(timeout); err != nil {
		return err
	}

	if err := validateDirectory(outputDir); err != nil {
		return err
	}

	// Warn if --close-tab is used with --force-headless (tabs close automatically)
	if c.Bool("close-tab") && c.Bool("force-headless") {
		logger.Warning("--close-tab is ignored in headless mode (tabs close automatically)")
	}

	userDataDir := ""
	if c.IsSet("user-data-dir") {
		validatedDir, err := validateUserDataDir(c.String("user-data-dir"))
		if err != nil {
			return err
		}
		userDataDir = validatedDir
	}

	userAgent := validateUserAgent(c.String("user-agent"))

	bm := NewBrowserManager(BrowserOptions{
		Port:          c.Int("port"),
		ForceHeadless: c.Bool("force-headless"),
		UserAgent:     userAgent,
		UserDataDir:   userDataDir,
	})

	browserManager = bm

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

	successCount := 0
	failureCount := 0

	for i, urlStr := range urls {
		current := i + 1
		total := len(urls)

		logger.Info("[%d/%d] Fetching: %s", current, total, urlStr)

		validatedURL, err := validateURL(urlStr)
		if err != nil {
			logger.Error("[%d/%d] Invalid URL - skipping: %v", current, total, err)
			failureCount++
			continue
		}

		page, err := bm.NewPage()
		if err != nil {
			logger.Error("[%d/%d] Failed to create page: %v", current, total, err)
			failureCount++
			continue
		}

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

		info, err := page.Info()
		if err != nil {
			logger.Error("[%d/%d] Failed to get page info: %v", current, total, err)
			bm.ClosePage(page)
			failureCount++
			continue
		}

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

		if err := processPageContent(page, format, outputPath); err != nil {
			logger.Error("[%d/%d] Failed to save content: %v", current, total, err)
			bm.ClosePage(page)
			failureCount++
			continue
		}

		if bm.launchedHeadless || c.Bool("close-tab") {
			bm.ClosePage(page)
		}

		successCount++
	}

	logger.Success("Batch complete: %d succeeded, %d failed", successCount, failureCount)

	if failureCount > 0 {
		return fmt.Errorf("batch processing completed with %d failures", failureCount)
	}

	return nil
}
