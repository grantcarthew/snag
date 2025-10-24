// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import "errors"

// Sentinel errors for internal logic and testing
var (
	// ErrBrowserNotFound indicates no Chromium-based browser was found on the system
	ErrBrowserNotFound = errors.New("no Chromium-based browser found")

	// ErrPageLoadTimeout indicates the page took too long to load
	ErrPageLoadTimeout = errors.New("page load timeout exceeded")

	// ErrAuthRequired indicates authentication is needed to access the page
	ErrAuthRequired = errors.New("authentication required")

	// ErrInvalidURL indicates the provided URL is invalid or malformed
	ErrInvalidURL = errors.New("invalid URL")

	// ErrConversionFailed indicates HTML to Markdown conversion failed
	ErrConversionFailed = errors.New("HTML to Markdown conversion failed")

	// ErrBrowserConnection indicates failure to connect to the browser
	ErrBrowserConnection = errors.New("failed to connect to browser")

	// ErrNavigationFailed indicates page navigation failed
	ErrNavigationFailed = errors.New("page navigation failed")

	// ErrNoBrowserRunning indicates no browser instance is running with remote debugging
	ErrNoBrowserRunning = errors.New("no browser instance running with remote debugging")

	// ErrTabIndexInvalid indicates the tab index is out of range
	ErrTabIndexInvalid = errors.New("tab index out of range")

	// ErrTabURLConflict indicates both --tab flag and URL argument were provided
	ErrTabURLConflict = errors.New("cannot use --tab with URL argument")

	// ErrNoTabMatch indicates no tab matches the provided pattern
	ErrNoTabMatch = errors.New("no tab matches pattern")

	// ErrNoValidURLs indicates no valid URLs were provided or found in URL file
	ErrNoValidURLs = errors.New("no valid URLs provided")

	// ErrOutputFlagConflict indicates --output cannot be used with multiple URLs
	ErrOutputFlagConflict = errors.New("--output cannot be used with multiple URLs, use --output-dir instead")
)
