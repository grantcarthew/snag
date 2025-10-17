// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

// PageFetcher handles fetching page content
type PageFetcher struct {
	page    *rod.Page
	timeout time.Duration
}

// FetchOptions contains options for fetching a page
type FetchOptions struct {
	URL     string
	Timeout int
	WaitFor string
}

// NewPageFetcher creates a new page fetcher
func NewPageFetcher(page *rod.Page, timeout int) *PageFetcher {
	return &PageFetcher{
		page:    page,
		timeout: time.Duration(timeout) * time.Second,
	}
}

// Fetch navigates to a URL and returns the HTML content
func (pf *PageFetcher) Fetch(opts FetchOptions) (string, error) {
	logger.Progress("Fetching %s...", opts.URL)

	// Navigate to the URL with timeout
	logger.Verbose("Navigating to %s (timeout: %ds)...", opts.URL, opts.Timeout)

	// Apply timeout only to navigation - use original page for other operations
	// This prevents "context deadline exceeded" on HTML extraction if total time > timeout
	err := pf.page.Timeout(pf.timeout).Navigate(opts.URL)
	if err != nil {
		// Check if it's a timeout using proper error type checking
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Error("Page load timeout exceeded (%ds)", opts.Timeout)
			logger.ErrorWithSuggestion(
				fmt.Sprintf("The page took too long to load"),
				fmt.Sprintf("snag %s --timeout 60", opts.URL),
			)
			return "", ErrPageLoadTimeout
		}
		return "", fmt.Errorf("%w: %v", ErrNavigationFailed, err)
	}

	// Wait for page to be stable (use original page - no timeout constraint)
	logger.Verbose("Waiting for page to stabilize...")
	err = pf.page.WaitStable(StabilizeTimeout)
	if err != nil {
		logger.Warning("Page did not stabilize: %v", err)
	}

	// If WaitFor selector is specified, wait for it
	if opts.WaitFor != "" {
		logger.Verbose("Waiting for selector: %s", opts.WaitFor)
		elem, err := pf.page.Element(opts.WaitFor)
		if err != nil {
			return "", fmt.Errorf("failed to find selector %s: %w", opts.WaitFor, err)
		}
		err = elem.WaitVisible()
		if err != nil {
			return "", fmt.Errorf("selector %s not visible: %w", opts.WaitFor, err)
		}
		logger.Verbose("Selector found: %s", opts.WaitFor)
	}

	// Check for authentication requirements
	if authErr := pf.detectAuth(); authErr != nil {
		return "", authErr
	}

	// Extract HTML content (use original page - should be fast, no timeout needed)
	logger.Verbose("Extracting HTML content...")
	html, err := pf.page.HTML()
	if err != nil {
		return "", fmt.Errorf("failed to extract HTML: %w", err)
	}

	logger.Debug("Extracted %d bytes of HTML", len(html))
	logger.Success("Fetched successfully")

	return html, nil
}

// detectAuth checks if the page requires authentication
func (pf *PageFetcher) detectAuth() error {
	// Check HTTP status code by evaluating JavaScript
	statusCode, err := pf.page.Eval(`() => {
		return window.performance?.getEntriesByType?.('navigation')?.[0]?.responseStatus || 0;
	}`)

	if err == nil && statusCode.Value.Int() > 0 {
		status := statusCode.Value.Int()
		logger.Debug("HTTP status code: %d", status)

		if status == 401 || status == 403 {
			logger.Error("Authentication required (HTTP %d)", status)
			logger.ErrorWithSuggestion(
				"This page requires authentication",
				"snag --force-visible "+pf.getURL(),
			)
			return ErrAuthRequired
		}
	}

	// Check for common login page indicators in the page content
	hasLogin, _, err := pf.page.Has("input[type='password']")
	if err == nil && hasLogin {
		// Check if we also have username field and submit button (likely a login form)
		hasUsername, _, _ := pf.page.Has("input[type='text'], input[type='email'], input[name*='user'], input[name*='login']")
		hasSubmit, _, _ := pf.page.Has("button[type='submit'], input[type='submit']")

		if hasUsername && hasSubmit {
			logger.Debug("Detected login form on page")

			// Also check the title or URL for login keywords
			title, _ := pf.page.Info()
			if title != nil && (strings.Contains(strings.ToLower(title.Title), "login") ||
				strings.Contains(strings.ToLower(title.Title), "sign in") ||
				strings.Contains(strings.ToLower(title.URL), "/login") ||
				strings.Contains(strings.ToLower(title.URL), "/signin") ||
				strings.Contains(strings.ToLower(title.URL), "/auth")) {

				logger.Warning("This appears to be a login page")
				logger.ErrorWithSuggestion(
					"Authentication may be required",
					"snag --force-visible "+pf.getURL(),
				)
				// Don't return error yet - might be a page that just happens to have a login form
			}
		}
	}

	return nil
}

// getURL gets the current page URL
func (pf *PageFetcher) getURL() string {
	info, err := pf.page.Info()
	if err != nil {
		return ""
	}
	return info.URL
}
