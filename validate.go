// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"net/url"
	"strings"
)

// validateURL checks and normalizes the URL, adding https:// if no scheme is present.
// Supported schemes: http, https, file
func validateURL(urlStr string) (string, error) {
	// Add https:// if no scheme present
	if !strings.Contains(urlStr, "://") {
		urlStr = "https://" + urlStr
		logger.Verbose("No scheme provided, using: %s", urlStr)
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		logger.Error("Invalid URL: %s", urlStr)
		logger.ErrorWithSuggestion(
			fmt.Sprintf("URL parsing failed: %v", err),
			"snag https://example.com",
		)
		return "", ErrInvalidURL
	}

	// Validate scheme
	validSchemes := map[string]bool{
		"http":  true,
		"https": true,
		"file":  true,
	}

	if !validSchemes[parsedURL.Scheme] {
		logger.Error("Unsupported URL scheme: %s", parsedURL.Scheme)
		logger.ErrorWithSuggestion(
			"URL must use http://, https://, or file://",
			"snag https://example.com",
		)
		return "", ErrInvalidURL
	}

	// Validate host (except for file://)
	if parsedURL.Scheme != "file" && parsedURL.Host == "" {
		logger.Error("Invalid URL: missing host")
		logger.ErrorWithSuggestion(
			"URL must include a hostname",
			"snag https://example.com",
		)
		return "", ErrInvalidURL
	}

	return urlStr, nil
}
