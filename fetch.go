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

type PageFetcher struct {
	page    *rod.Page
	timeout time.Duration
}

type FetchOptions struct {
	URL     string
	Timeout int
	WaitFor string
}

func NewPageFetcher(page *rod.Page, timeout int) *PageFetcher {
	return &PageFetcher{
		page:    page,
		timeout: time.Duration(timeout) * time.Second,
	}
}

func (pf *PageFetcher) Fetch(opts FetchOptions) (string, error) {
	logger.Info("Fetching %s...", opts.URL)

	logger.Verbose("Navigating to %s (timeout: %ds)...", opts.URL, opts.Timeout)

	// Apply timeout to long-running operations (navigation, wait-for) using inline .Timeout()
	// This creates temporary timeout clones that don't affect subsequent fast operations
	// (HTML extraction, auth detection), preventing cumulative timeout issues
	err := pf.page.Timeout(pf.timeout).Navigate(opts.URL)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Error("Page load timeout exceeded (%ds)", opts.Timeout)
			logger.ErrorWithSuggestion(
				fmt.Sprintf("The page took too long to load"),
				fmt.Sprintf("snag %s --timeout 60", opts.URL),
			)
			return "", ErrPageLoadTimeout
		}
		return "", fmt.Errorf("%w: %w", ErrNavigationFailed, err)
	}

	logger.Verbose("Waiting for page to stabilize...")
	err = pf.page.WaitStable(StabilizeTimeout)
	if err != nil {
		logger.Warning("Page did not stabilize: %v", err)
	}

	if opts.WaitFor != "" {
		err := waitForSelector(pf.page, opts.WaitFor, pf.timeout)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				logger.ErrorWithSuggestion(
					fmt.Sprintf("Selector not found within %ds", opts.Timeout),
					fmt.Sprintf("snag --wait-for '%s' --timeout 60 %s", opts.WaitFor, opts.URL),
				)
			}
			return "", err
		}
	}

	if authErr := pf.detectAuth(); authErr != nil {
		return "", authErr
	}

	logger.Verbose("Extracting HTML content...")
	html, err := pf.page.HTML()
	if err != nil {
		return "", fmt.Errorf("failed to extract HTML: %w", err)
	}

	logger.Debug("Extracted %d bytes of HTML", len(html))
	logger.Success("Fetched successfully")

	return html, nil
}

func (pf *PageFetcher) detectAuth() error {
	// SECURITY: This JavaScript is hardcoded and safe. Never accept user-provided
	// JavaScript for evaluation as it would create XSS vulnerabilities.
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
				"snag --open-browser "+pf.getURL(),
			)
			return ErrAuthRequired
		}
	}

	hasLogin, _, err := pf.page.Has("input[type='password']")
	if err == nil && hasLogin {
		hasUsername, _, _ := pf.page.Has("input[type='text'], input[type='email'], input[name*='user'], input[name*='login']")
		hasSubmit, _, _ := pf.page.Has("button[type='submit'], input[type='submit']")

		if hasUsername && hasSubmit {
			logger.Debug("Detected login form on page")

			title, _ := pf.page.Info()
			if title != nil && (strings.Contains(strings.ToLower(title.Title), "login") ||
				strings.Contains(strings.ToLower(title.Title), "sign in") ||
				strings.Contains(strings.ToLower(title.URL), "/login") ||
				strings.Contains(strings.ToLower(title.URL), "/signin") ||
				strings.Contains(strings.ToLower(title.URL), "/auth")) {

				logger.Warning("This appears to be a login page")
				logger.ErrorWithSuggestion(
					"Authentication may be required",
					"snag --open-browser "+pf.getURL(),
				)
			}
		}
	}

	return nil
}

func (pf *PageFetcher) getURL() string {
	info, err := pf.page.Info()
	if err != nil {
		return ""
	}
	return info.URL
}

// waitForSelector waits for a CSS selector to appear and be visible on the page
// This is a shared helper function to avoid code duplication between Fetch and tab operations
func waitForSelector(page *rod.Page, selector string, timeout time.Duration) error {
	logger.Verbose("Waiting for selector: %s", selector)

	// Apply timeout to Element - it inherits to WaitVisible
	elem, err := page.Timeout(timeout).Element(selector)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Error("Timeout waiting for selector: %s", selector)
			return fmt.Errorf("timeout waiting for selector %s: %w", selector, err)
		}
		return fmt.Errorf("failed to find selector %s: %w", selector, err)
	}

	err = elem.WaitVisible()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Error("Timeout waiting for selector to be visible: %s", selector)
			return fmt.Errorf("timeout waiting for selector %s to be visible: %w", selector, err)
		}
		return fmt.Errorf("selector %s not visible: %w", selector, err)
	}

	logger.Verbose("Selector found: %s", selector)
	return nil
}
