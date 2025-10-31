// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import "errors"

var (
	ErrBrowserNotFound    = errors.New("no Chromium-based browser found")
	ErrPageLoadTimeout    = errors.New("page load timeout exceeded")
	ErrAuthRequired       = errors.New("authentication required")
	ErrInvalidURL         = errors.New("invalid URL")
	ErrConversionFailed   = errors.New("HTML to Markdown conversion failed")
	ErrBrowserConnection  = errors.New("failed to connect to browser")
	ErrNavigationFailed   = errors.New("page navigation failed")
	ErrNoBrowserRunning   = errors.New("no browser instance running with remote debugging")
	ErrTabIndexInvalid    = errors.New("tab index out of range")
	ErrTabURLConflict     = errors.New("cannot use both --tab and URL arguments")
	ErrNoTabMatch         = errors.New("no tab matches pattern")
	ErrNoValidURLs        = errors.New("no valid URLs provided")
	ErrOutputFlagConflict = errors.New("--output cannot be used with multiple content sources, use --output-dir instead")
)
