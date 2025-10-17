// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
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
	browser       *rod.Browser
	launcher      *launcher.Launcher
	port          int
	wasLaunched   bool
	userAgent     string
	forceHeadless bool
	forceVisible  bool
	openBrowser   bool
}

// BrowserOptions contains options for browser management
type BrowserOptions struct {
	Port          int
	ForceHeadless bool
	ForceVisible  bool
	OpenBrowser   bool
	UserAgent     string
}

// NewBrowserManager creates a new browser manager
func NewBrowserManager(opts BrowserOptions) *BrowserManager {
	return &BrowserManager{
		port:          opts.Port,
		userAgent:     opts.UserAgent,
		forceHeadless: opts.ForceHeadless,
		forceVisible:  opts.ForceVisible,
		openBrowser:   opts.OpenBrowser,
	}
}

// Connect attempts to connect to an existing browser or launch a new one
func (bm *BrowserManager) Connect() (*rod.Browser, error) {
	// Strategy 1: Try to connect to existing browser instance (unless forced)
	if !bm.forceHeadless && !bm.forceVisible {
		logger.Verbose("Checking for existing Chrome instance on port %d...", bm.port)
		if browser, err := bm.connectToExisting(); err == nil {
			logger.Success("Connected to existing Chrome instance")
			bm.browser = browser
			bm.wasLaunched = false
			return browser, nil
		}
		logger.Verbose("No existing Chrome instance found")
	}

	// Strategy 2: Launch new browser instance
	headless := !bm.forceVisible && !bm.openBrowser

	if headless {
		logger.Verbose("Launching Chrome in headless mode...")
	} else {
		logger.Info("Launching Chrome in visible mode...")
	}

	browser, err := bm.launchBrowser(headless)
	if err != nil {
		return nil, err
	}

	if headless {
		logger.Success("Chrome launched in headless mode")
	} else {
		logger.Success("Chrome launched in visible mode")
	}

	bm.browser = browser
	bm.wasLaunched = true
	return browser, nil
}

// connectToExisting attempts to connect to an existing Chrome instance
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
	// Find browser executable
	// SECURITY: We trust the system-installed browser binary found by launcher.LookPath().
	// Binary integrity is the responsibility of the OS package manager. If an attacker
	// can replace the browser binary, they already have system-level access.
	path, exists := launcher.LookPath()
	if !exists {
		return nil, ErrBrowserNotFound
	}

	logger.Debug("Found browser at: %s", path)

	// Create launcher with options
	l := launcher.New().
		Bin(path).
		Headless(headless).
		Set("disable-blink-features", "AutomationControlled")

	// Set custom user agent if provided
	if bm.userAgent != "" {
		l = l.Set("user-agent", bm.userAgent)
		logger.Verbose("Using custom user agent: %s", bm.userAgent)
	}

	// Set remote debugging port
	if bm.port != 9222 {
		l = l.Set("remote-debugging-port", fmt.Sprintf("%d", bm.port))
	}

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
	// Find browser executable
	path, exists := launcher.LookPath()
	if !exists {
		return ErrBrowserNotFound
	}

	logger.Debug("Found browser at: %s", path)

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

// Close closes the browser if it was launched by us
// Cleanup is best-effort and does not return errors
func (bm *BrowserManager) Close() {
	if bm.browser == nil {
		return
	}

	// Only close browser if we launched it
	if bm.wasLaunched {
		logger.Verbose("Closing browser...")
		if err := bm.browser.Close(); err != nil {
			logger.Warning("Failed to close browser: %v", err)
		}

		// Cleanup launcher
		if bm.launcher != nil {
			bm.launcher.Cleanup()
		}
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
