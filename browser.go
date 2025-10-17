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
	controlURL := fmt.Sprintf("ws://localhost:%d", bm.port)

	// Create browser instance and connect with timeout
	// We need to assign the result because Timeout() creates a new instance
	browser := rod.New().ControlURL(controlURL).Timeout(5 * time.Second)

	// Try to connect
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBrowserConnection, err)
	}

	// Return the browser but without timeout for future operations
	// CancelTimeout() removes the timeout context from subsequent operations
	return browser.CancelTimeout(), nil
}

// launchBrowser launches a new browser instance
func (bm *BrowserManager) launchBrowser(headless bool) (*rod.Browser, error) {
	// Find browser executable
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
	browser := rod.New().ControlURL(controlURL).Timeout(30 * time.Second)

	// Try to connect
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBrowserConnection, err)
	}

	// Return the browser but without timeout for future operations
	// CancelTimeout() removes the timeout context from subsequent operations
	return browser.CancelTimeout(), nil
}

// OpenBrowserOnly opens a browser without navigating to any page
func (bm *BrowserManager) OpenBrowserOnly() error {
	browser, err := bm.launchBrowser(false) // Always visible
	if err != nil {
		return err
	}

	bm.browser = browser
	bm.wasLaunched = true

	logger.Success("Browser opened on port %d", bm.port)
	logger.Info("Browser is running with remote debugging enabled")
	logger.Info("You can now connect to it using: snag <url>")

	// Keep the browser open (don't close it)
	// The process will exit but browser stays running
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
func (bm *BrowserManager) Close() error {
	if bm.browser == nil {
		return nil
	}

	// Only close browser if we launched it
	if bm.wasLaunched {
		logger.Verbose("Closing browser...")
		if err := bm.browser.Close(); err != nil {
			logger.Warning("Failed to close browser: %v", err)
			return err
		}

		// Cleanup launcher
		if bm.launcher != nil {
			bm.launcher.Cleanup()
		}
	} else {
		logger.Verbose("Leaving existing browser instance running")
	}

	return nil
}

// ClosePage closes a specific page
func (bm *BrowserManager) ClosePage(page *rod.Page) error {
	if page == nil {
		return nil
	}

	logger.Verbose("Closing page...")
	if err := page.Close(); err != nil {
		logger.Warning("Failed to close page: %v", err)
		return err
	}

	return nil
}

// WasLaunched returns true if the browser was launched by this manager
func (bm *BrowserManager) WasLaunched() bool {
	return bm.wasLaunched
}
