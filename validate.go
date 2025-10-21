// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
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

// validateTimeout checks if timeout value is valid (must be positive)
func validateTimeout(timeout int) error {
	if timeout <= 0 {
		logger.Error("Invalid timeout: %d", timeout)
		logger.ErrorWithSuggestion(
			"Timeout must be a positive number of seconds",
			"snag <url> --timeout 30",
		)
		return fmt.Errorf("invalid timeout: %d", timeout)
	}
	return nil
}

// validatePort checks if port is in valid range (1-65535)
func validatePort(port int) error {
	if port < 1 || port > 65535 {
		logger.Error("Invalid port: %d", port)
		logger.ErrorWithSuggestion(
			"Port must be between 1 and 65535",
			"snag <url> --port 9222",
		)
		return fmt.Errorf("invalid port: %d", port)
	}
	return nil
}

// validateOutputPath checks if the output file path is writable
func validateOutputPath(path string) error {
	// Get the directory path
	dir := filepath.Dir(path)

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.Error("Output directory does not exist: %s", dir)
		logger.ErrorWithSuggestion(
			fmt.Sprintf("Directory '%s' not found", dir),
			"snag <url> -o /path/to/existing/dir/output.md",
		)
		return fmt.Errorf("output directory does not exist: %s", dir)
	}

	// Check if directory is writable by attempting to create a temp file
	f, err := os.CreateTemp(dir, ".snag-write-test-*")
	if err != nil {
		logger.Error("Cannot write to output directory: %s", dir)
		logger.ErrorWithSuggestion(
			"Permission denied or directory not writable",
			"snag <url> -o /path/to/writable/dir/output.md",
		)
		return fmt.Errorf("cannot write to output directory: %s", dir)
	}
	testFile := f.Name()
	f.Close()
	os.Remove(testFile)

	return nil
}

// validateFormat checks if format is valid (markdown, html, text, or pdf)
func validateFormat(format string) error {
	// Define valid formats locally for better testability and self-containment
	validFormats := map[string]bool{
		FormatMarkdown: true,
		FormatHTML:     true,
		FormatText:     true,
		FormatPDF:      true,
	}

	if !validFormats[format] {
		logger.Error("Invalid format: %s", format)
		logger.ErrorWithSuggestion(
			fmt.Sprintf("Format must be '%s', '%s', '%s', or '%s'", FormatMarkdown, FormatHTML, FormatText, FormatPDF),
			fmt.Sprintf("snag <url> --format %s", FormatMarkdown),
		)
		return fmt.Errorf("invalid format: %s", format)
	}

	return nil
}

// validateDirectory checks if a directory exists and is writable
func validateDirectory(dir string) error {
	// Check if directory exists
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		logger.Error("Directory does not exist: %s", dir)
		logger.ErrorWithSuggestion(
			fmt.Sprintf("Directory '%s' not found", dir),
			"mkdir -p /path/to/dir && snag <url> -d /path/to/dir",
		)
		return fmt.Errorf("directory does not exist: %s", dir)
	}
	if err != nil {
		return fmt.Errorf("error accessing directory: %w", err)
	}

	// Check if it's actually a directory
	if !info.IsDir() {
		logger.Error("Path is not a directory: %s", dir)
		return fmt.Errorf("not a directory: %s", dir)
	}

	// Check if directory is writable
	testFile, err := os.CreateTemp(dir, ".snag-write-test-*")
	if err != nil {
		logger.Error("Directory not writable: %s", dir)
		logger.ErrorWithSuggestion(
			"Permission denied or directory not writable",
			"chmod u+w /path/to/dir",
		)
		return fmt.Errorf("directory not writable: %s", dir)
	}
	testFilePath := testFile.Name()
	testFile.Close()
	os.Remove(testFilePath)

	return nil
}

// validateOutputPathEscape prevents directory escape attacks when combining -o and -d flags.
// Ensures that the resulting path doesn't escape the output directory using .. or similar.
func validateOutputPathEscape(outputDir, filename string) error {
	// Skip validation if filename is an absolute path (it ignores outputDir)
	if filepath.IsAbs(filename) {
		return nil
	}

	// Join the paths and clean them
	fullPath := filepath.Join(outputDir, filename)
	cleanPath := filepath.Clean(fullPath)
	cleanDir := filepath.Clean(outputDir)

	// Make cleanDir absolute for comparison
	absDir, err := filepath.Abs(cleanDir)
	if err != nil {
		return fmt.Errorf("failed to resolve directory path: %w", err)
	}

	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to resolve output path: %w", err)
	}

	// Ensure the cleaned path starts with the cleaned directory
	// Add separator to prevent partial matching (e.g., /tmp vs /tmp2)
	if !strings.HasPrefix(absPath+string(filepath.Separator), absDir+string(filepath.Separator)) {
		logger.Error("Output path escapes directory: %s", filename)
		logger.ErrorWithSuggestion(
			"Path contains '..' or similar that escapes the output directory",
			"snag <url> -o output.md -d /path/to/dir",
		)
		return fmt.Errorf("output path escapes directory: %s", filename)
	}

	return nil
}
