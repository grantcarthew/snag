// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// Timeout constants
const (
	ConnectTimeout   = 10 * time.Second // Browser connection timeout (existing or newly launched)
	StabilizeTimeout = 3 * time.Second  // Page stabilization wait time
)

// BrowserManager handles browser lifecycle and connection
type BrowserManager struct {
	browser          *rod.Browser
	launcher         *launcher.Launcher
	port             int
	wasLaunched      bool
	launchedHeadless bool // True if we launched the browser in headless mode
	userAgent        string
	forceHeadless    bool
	openBrowser      bool
	browserName      string // Detected browser name (Chrome, Chromium, Edge, Brave, etc.)
}

// BrowserOptions contains options for browser management
type BrowserOptions struct {
	Port          int
	ForceHeadless bool
	OpenBrowser   bool
	UserAgent     string
}

// TabInfo represents information about a browser tab
type TabInfo struct {
	Index int    // Tab index (1-based for display)
	URL   string // Current URL of the tab
	Title string // Page title
	ID    string // Internal target ID (for rod)
}

// findBrowserPath locates the browser executable and detects its name
// Returns the path or an error if not found
func (bm *BrowserManager) findBrowserPath() (string, error) {
	// Find browser executable
	// SECURITY: We trust the system-installed browser binary found by launcher.LookPath().
	// Binary integrity is the responsibility of the OS package manager. If an attacker
	// can replace the browser binary, they already have system-level access.
	path, exists := launcher.LookPath()
	if !exists {
		return "", ErrBrowserNotFound
	}

	// Detect browser name from path
	bm.browserName = detectBrowserName(path)

	logger.Debug("Found browser at: %s", path)

	return path, nil
}

// detectBrowserName extracts the browser name from the executable path
func detectBrowserName(path string) string {
	// Extract just the executable name to avoid false positives from directory/user names
	base := filepath.Base(path)

	// Remove common extensions for matching
	baseName := strings.TrimSuffix(base, ".exe")
	baseName = strings.TrimSuffix(baseName, ".app")

	// Convert to lowercase for case-insensitive matching
	lowerName := strings.ToLower(baseName)

	// Check for specific browsers in order of specificity
	// Note: Order matters - check specific names before generic ones
	// (e.g., "chrome" before "chromium", "ungoogled-chromium" before "chromium")

	// Ungoogled Chromium (check before regular Chromium)
	if strings.Contains(lowerName, "ungoogled") {
		return "Ungoogled-Chromium"
	}

	// Chrome (check before Chromium to avoid false matches)
	if strings.Contains(lowerName, "chrome") && !strings.Contains(lowerName, "chromium") {
		return "Chrome"
	}

	// Chromium
	if strings.Contains(lowerName, "chromium") {
		return "Chromium"
	}

	// Microsoft Edge
	if strings.Contains(lowerName, "edge") || strings.Contains(lowerName, "msedge") {
		return "Edge"
	}

	// Brave
	if strings.Contains(lowerName, "brave") {
		return "Brave"
	}

	// Opera
	if strings.Contains(lowerName, "opera") {
		return "Opera"
	}

	// Vivaldi
	if strings.Contains(lowerName, "vivaldi") {
		return "Vivaldi"
	}

	// Arc
	if strings.Contains(lowerName, "arc") {
		return "Arc"
	}

	// Yandex
	if strings.Contains(lowerName, "yandex") {
		return "Yandex"
	}

	// Thorium
	if strings.Contains(lowerName, "thorium") {
		return "Thorium"
	}

	// Slimjet
	if strings.Contains(lowerName, "slimjet") {
		return "Slimjet"
	}

	// Cent Browser
	if strings.Contains(lowerName, "cent") {
		return "Cent"
	}

	// Fallback: Capitalize first letter of base name
	if len(baseName) > 0 {
		return strings.ToUpper(baseName[:1]) + baseName[1:]
	}

	return "Browser"
}

// NewBrowserManager creates a new browser manager
func NewBrowserManager(opts BrowserOptions) *BrowserManager {
	return &BrowserManager{
		port:          opts.Port,
		userAgent:     opts.UserAgent,
		forceHeadless: opts.ForceHeadless,
		openBrowser:   opts.OpenBrowser,
	}
}

// Connect attempts to connect to an existing browser or launch a new one
func (bm *BrowserManager) Connect() (*rod.Browser, error) {
	// Strategy 1: Try to connect to existing browser instance (unless forced)
	if !bm.forceHeadless && !bm.openBrowser {
		logger.Verbose("Checking for existing browser instance on port %d...", bm.port)
		if browser, err := bm.connectToExisting(); err == nil {
			logger.Success("Connected to existing browser instance")
			bm.browser = browser
			bm.wasLaunched = false
			return browser, nil
		}
		logger.Verbose("No existing browser instance found")
	}

	// Strategy 2: Launch new browser instance
	// Priority: forceHeadless takes precedence over openBrowser
	headless := bm.forceHeadless || !bm.openBrowser

	if headless {
		logger.Verbose("Launching browser in headless mode...")
	} else {
		logger.Info("Launching browser in visible mode...")
	}

	browser, err := bm.launchBrowser(headless)
	if err != nil {
		return nil, err
	}

	if headless {
		logger.Success("%s launched in headless mode", bm.browserName)
	} else {
		logger.Success("%s launched in visible mode", bm.browserName)
	}

	bm.browser = browser
	bm.wasLaunched = true
	bm.launchedHeadless = headless
	return browser, nil
}

// connectToExisting attempts to connect to an existing browser instance
func (bm *BrowserManager) connectToExisting() (*rod.Browser, error) {
	// Query the browser for its WebSocket debugger URL
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", bm.port)

	// Use launcher's ResolveURL to get the WebSocket URL from the HTTP endpoint
	wsURL, err := launcher.ResolveURL(baseURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBrowserConnection, err)
	}

	// Create browser instance and connect with timeout
	// We need to assign the result because Timeout() creates a new instance
	browser := rod.New().ControlURL(wsURL).Timeout(ConnectTimeout)

	// Try to connect
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBrowserConnection, err)
	}

	// Return the browser but without timeout for future operations
	// CancelTimeout() removes the timeout context from subsequent operations
	return browser.CancelTimeout(), nil
}

// launchBrowser launches a new browser instance
func (bm *BrowserManager) launchBrowser(headless bool) (*rod.Browser, error) {
	// Find browser executable and detect its name
	path, err := bm.findBrowserPath()
	if err != nil {
		return nil, err
	}

	// Create launcher with options
	l := launcher.New().
		Bin(path).
		Headless(headless).
		Leakless(headless). // Only kill browser on exit if headless; visible browsers persist
		Set("disable-blink-features", "AutomationControlled")

	// Set custom user agent if provided
	if bm.userAgent != "" {
		l = l.Set("user-agent", bm.userAgent)
		logger.Verbose("Using custom user agent: %s", bm.userAgent)
	}

	// Always set remote debugging port explicitly (don't rely on launcher's default)
	l = l.Set("remote-debugging-port", fmt.Sprintf("%d", bm.port))

	// Launch browser
	controlURL, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	bm.launcher = l

	// Create browser instance and connect with timeout
	// We need to assign the result because Timeout() creates a new instance
	browser := rod.New().ControlURL(controlURL).Timeout(ConnectTimeout)

	// Try to connect
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBrowserConnection, err)
	}

	// Return the browser but without timeout for future operations
	// CancelTimeout() removes the timeout context from subsequent operations
	return browser.CancelTimeout(), nil
}

// OpenBrowserOnly opens a browser without navigating to any page
// The browser is left running with CDP debugging enabled, and snag exits
func (bm *BrowserManager) OpenBrowserOnly() error {
	// Check if a browser is already running on this port
	logger.Verbose("Checking for existing browser instance on port %d...", bm.port)
	if _, err := bm.connectToExisting(); err == nil {
		// Browser is already running - nothing to do
		// Don't call Close() - that would close the browser itself
		// Just let the connection be garbage collected
		logger.Success("Browser already running on port %d", bm.port)
		logger.Info("You can connect to it using: snag <url>")
		return nil
	}

	// Find browser executable and detect its name
	path, err := bm.findBrowserPath()
	if err != nil {
		return err
	}

	// Create launcher with options
	// Leakless(false) allows the browser to persist after this process exits
	l := launcher.New().
		Bin(path).
		Leakless(false). // Browser persists after snag exits
		Headless(false). // Always visible
		Set("disable-blink-features", "AutomationControlled").
		Set("remote-debugging-port", fmt.Sprintf("%d", bm.port))

	// Launch browser and let it run independently
	controlURL, err := l.Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}

	// Connect to browser briefly to open a visible tab, then disconnect
	browser := rod.New().ControlURL(controlURL).Timeout(ConnectTimeout)
	if err := browser.Connect(); err != nil {
		return fmt.Errorf("%w: %w", ErrBrowserConnection, err)
	}

	// Create a new page so the browser window is visible
	_, err = browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		browser.Close()
		return fmt.Errorf("failed to create page: %w", err)
	}

	// Disconnect from browser but leave it running
	// Don't call Close() - we want the browser to persist

	logger.Success("Browser opened on port %d", bm.port)
	logger.Info("Browser is running with remote debugging enabled")
	logger.Info("You can now connect to it using: snag <url>")

	// Don't store launcher or browser - let it run independently
	// Don't call cleanup - the browser stays running after we exit
	return nil
}

// NewPage creates a new page in the browser
func (bm *BrowserManager) NewPage() (*rod.Page, error) {
	if bm.browser == nil {
		return nil, fmt.Errorf("browser not connected")
	}

	page, err := bm.browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	return page, nil
}

// Close closes the browser if it was launched by us in headless mode
// Cleanup is best-effort and does not return errors
func (bm *BrowserManager) Close() {
	if bm.browser == nil {
		return
	}

	// Only close browser if we launched it AND it's headless
	// Visible browsers are left running for user convenience (e.g., authenticated sessions)
	if bm.wasLaunched && bm.launchedHeadless {
		logger.Verbose("Closing headless browser...")
		if err := bm.browser.Close(); err != nil {
			logger.Warning("Failed to close browser: %v", err)
		}

		// Cleanup launcher
		if bm.launcher != nil {
			bm.launcher.Cleanup()
		}
	} else if bm.wasLaunched && !bm.launchedHeadless {
		logger.Verbose("Leaving visible browser running")
	} else {
		logger.Verbose("Leaving existing browser instance running")
	}
}

// ClosePage closes a specific page
// Cleanup is best-effort and does not return errors
func (bm *BrowserManager) ClosePage(page *rod.Page) {
	if page == nil {
		return
	}

	logger.Verbose("Closing page...")
	if err := page.Close(); err != nil {
		logger.Warning("Failed to close page: %v", err)
	}
}

// WasLaunched returns true if the browser was launched by this manager
func (bm *BrowserManager) WasLaunched() bool {
	return bm.wasLaunched
}

// pageWithInfo holds a page reference with its cached info for efficient sorting
type pageWithInfo struct {
	page  *rod.Page
	url   string
	title string
	id    string
}

// getSortedPagesWithInfo returns all pages sorted by URL→Title→ID with info cached
// This fetches page.Info() exactly once per page, then sorts in-memory
// Sorting provides predictable, stable ordering (CDP's internal order is unpredictable)
func (bm *BrowserManager) getSortedPagesWithInfo() ([]pageWithInfo, error) {
	if bm.browser == nil {
		return nil, ErrNoBrowserRunning
	}

	pages, err := bm.browser.Pages()
	if err != nil {
		return nil, fmt.Errorf("failed to get pages: %w", err)
	}

	// Build list with cached info (single page.Info() call per page)
	pagesWithInfo := make([]pageWithInfo, 0, len(pages))
	for _, page := range pages {
		info, err := page.Info()
		if err != nil {
			// Skip pages we can't get info for, but log a warning
			logger.Warning("Failed to get info for page: %v", err)
			continue
		}
		pagesWithInfo = append(pagesWithInfo, pageWithInfo{
			page:  page,
			url:   info.URL,
			title: info.Title,
			id:    string(page.TargetID),
		})
	}

	// Sort by URL (primary) → Title (secondary) → ID (tertiary)
	sort.Slice(pagesWithInfo, func(i, j int) bool {
		if pagesWithInfo[i].url != pagesWithInfo[j].url {
			return pagesWithInfo[i].url < pagesWithInfo[j].url
		}
		if pagesWithInfo[i].title != pagesWithInfo[j].title {
			return pagesWithInfo[i].title < pagesWithInfo[j].title
		}
		return pagesWithInfo[i].id < pagesWithInfo[j].id
	})

	return pagesWithInfo, nil
}

// ListTabs returns information about all open tabs in the browser
// Tabs are sorted by URL (primary), Title (secondary), and ID (tertiary) for predictable ordering
// This requires an existing browser connection and will not launch a new browser
func (bm *BrowserManager) ListTabs() ([]TabInfo, error) {
	pagesWithInfo, err := bm.getSortedPagesWithInfo()
	if err != nil {
		return nil, err
	}

	// Convert to TabInfo slice with 1-based indices
	tabs := make([]TabInfo, len(pagesWithInfo))
	for i, pwi := range pagesWithInfo {
		tabs[i] = TabInfo{
			Index: i + 1, // 1-based indexing for display
			URL:   pwi.url,
			Title: pwi.title,
			ID:    pwi.id,
		}
	}

	return tabs, nil
}

// GetTabByIndex returns a specific tab by its index (1-based) from the sorted tab list
// Index 1 = first tab (by URL sort order), Index 2 = second tab, etc.
// Returns ErrTabIndexInvalid if index is out of range
func (bm *BrowserManager) GetTabByIndex(index int) (*rod.Page, error) {
	pagesWithInfo, err := bm.getSortedPagesWithInfo()
	if err != nil {
		return nil, err
	}

	// Validate index (1-based, so valid range is 1 to len(pagesWithInfo))
	if index < 1 || index > len(pagesWithInfo) {
		return nil, fmt.Errorf("%w: tab index %d (valid range: 1-%d)", ErrTabIndexInvalid, index, len(pagesWithInfo))
	}

	// Convert 1-based index to 0-based
	arrayIndex := index - 1

	// Log tab selection (info already cached, no extra network call)
	logger.Verbose("Selected tab [%d] from sorted order: %s", index, pagesWithInfo[arrayIndex].url)

	return pagesWithInfo[arrayIndex].page, nil
}

// GetTabByPattern returns the first tab matching the given pattern
// Pattern matching uses progressive fallthrough:
// 1. Try exact URL match (case-insensitive)
// 2. Try substring/contains match (case-insensitive)
// 3. Try regex match (case-insensitive)
// 4. Return error if no matches
// Returns ErrNoTabMatch if no tab matches the pattern
func (bm *BrowserManager) GetTabByPattern(pattern string) (*rod.Page, error) {
	pages, err := bm.GetTabsByPattern(pattern)
	if err != nil {
		return nil, err
	}
	// Return first match only (backward compatibility)
	return pages[0], nil
}

// GetTabsByPattern returns all tabs matching the given pattern
// Pattern matching uses progressive fallthrough:
// 1. Try exact URL match (case-insensitive) - returns ALL exact matches
// 2. Try substring/contains match (case-insensitive) - returns ALL substring matches
// 3. Try regex match (case-insensitive) - returns ALL regex matches
// 4. Return error if no matches
// Returns ErrNoTabMatch if no tab matches the pattern
func (bm *BrowserManager) GetTabsByPattern(pattern string) ([]*rod.Page, error) {
	if bm.browser == nil {
		return nil, ErrNoBrowserRunning
	}

	pages, err := bm.browser.Pages()
	if err != nil {
		return nil, fmt.Errorf("failed to get pages: %w", err)
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("%w: '%s' (no tabs open)", ErrNoTabMatch, pattern)
	}

	// Fetch page info once for all pages (avoid repeated network calls)
	type pageCache struct {
		page  *rod.Page
		url   string
		index int
	}

	var cached []pageCache
	for i, page := range pages {
		info, err := page.Info()
		if err != nil {
			// Skip pages we can't get info for
			logger.Warning("Failed to get info for page %d: %v", i+1, err)
			continue
		}
		cached = append(cached, pageCache{
			page:  page,
			url:   info.URL,
			index: i + 1, // 1-based for logging
		})
	}

	if len(cached) == 0 {
		return nil, fmt.Errorf("%w: '%s' (no accessible tabs)", ErrNoTabMatch, pattern)
	}

	patternLower := strings.ToLower(pattern)

	// Step 1: Try exact match (case-insensitive) - collect ALL exact matches
	var exactMatches []*rod.Page
	for _, pc := range cached {
		if strings.EqualFold(pc.url, pattern) {
			logger.Verbose("Matched tab [%d] via exact URL: %s", pc.index, pc.url)
			exactMatches = append(exactMatches, pc.page)
		}
	}
	if len(exactMatches) > 0 {
		return exactMatches, nil
	}

	// Step 2: Try contains/substring match (case-insensitive) - collect ALL substring matches
	var substringMatches []*rod.Page
	for _, pc := range cached {
		if strings.Contains(strings.ToLower(pc.url), patternLower) {
			logger.Verbose("Matched tab [%d] via substring: %s", pc.index, pc.url)
			substringMatches = append(substringMatches, pc.page)
		}
	}
	if len(substringMatches) > 0 {
		return substringMatches, nil
	}

	// Step 3: Try regex match (case-insensitive) - collect ALL regex matches
	re, err := regexp.Compile("(?i)" + pattern) // (?i) = case-insensitive
	if err != nil {
		// Invalid regex pattern - return specific error
		return nil, fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
	}

	var regexMatches []*rod.Page
	for _, pc := range cached {
		if re.MatchString(pc.url) {
			logger.Verbose("Matched tab [%d] via regex: %s", pc.index, pc.url)
			regexMatches = append(regexMatches, pc.page)
		}
	}
	if len(regexMatches) > 0 {
		return regexMatches, nil
	}

	// Step 4: No matches found
	return nil, fmt.Errorf("%w: '%s'", ErrNoTabMatch, pattern)
}

// GetTabsByRange returns tabs within the specified 1-based index range (inclusive) from the sorted tab list
// Range format: "N-M" where N and M are positive integers >= 1
// Examples: "1-3" returns tabs 1, 2, and 3 (by URL sort order)
// Returns error if range is invalid or indices are out of bounds
func (bm *BrowserManager) GetTabsByRange(start, end int) ([]*rod.Page, error) {
	pagesWithInfo, err := bm.getSortedPagesWithInfo()
	if err != nil {
		return nil, err
	}

	// Validate range
	if start < 1 {
		return nil, fmt.Errorf("tab range must start from 1 (got %d)", start)
	}
	if start > end {
		return nil, fmt.Errorf("invalid range: start must be <= end (got %d-%d)", start, end)
	}

	// Validate that both start and end indices exist
	if start > len(pagesWithInfo) {
		return nil, fmt.Errorf("tab index %d out of range in range %d-%d (only %d tabs open)", start, start, end, len(pagesWithInfo))
	}
	if end > len(pagesWithInfo) {
		return nil, fmt.Errorf("tab index %d out of range in range %d-%d (only %d tabs open)", end, start, end, len(pagesWithInfo))
	}

	// Extract the range of pages (convert from 1-based to 0-based indexing)
	rangeWithInfo := pagesWithInfo[start-1 : end]

	// Extract just the pages from the range
	rangeTabs := make([]*rod.Page, len(rangeWithInfo))
	for i, pwi := range rangeWithInfo {
		rangeTabs[i] = pwi.page
	}

	logger.Verbose("Selected %d tabs from sorted range [%d-%d]", len(rangeTabs), start, end)
	return rangeTabs, nil
}
