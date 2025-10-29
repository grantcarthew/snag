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
	if !strings.Contains(urlStr, "://") {
		urlStr = "https://" + urlStr
		logger.Verbose("No scheme provided, using: %s", urlStr)
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		logger.Error("Invalid URL: %s", urlStr)
		logger.Debug("URL parsing error: %v", err)
		logger.ErrorWithSuggestion(
			fmt.Sprintf("URL parsing failed: %v", err),
			"snag https://example.com",
		)
		return "", ErrInvalidURL
	}
	logger.Debug("Parsed URL - scheme: %s, host: %s, path: %s", parsedURL.Scheme, parsedURL.Host, parsedURL.Path)

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

// isNonFetchableURL checks if a URL is a browser-internal URL that cannot be fetched.
// These URLs (chrome://, about:, devtools://, etc.) should be skipped when processing tabs.
func isNonFetchableURL(urlStr string) bool {
	nonFetchablePrefixes := []string{
		"chrome://",
		"about:",
		"devtools://",
		"chrome-extension://",
		"edge://",
		"brave://",
	}

	urlLower := strings.ToLower(urlStr)
	for _, prefix := range nonFetchablePrefixes {
		if strings.HasPrefix(urlLower, prefix) {
			return true
		}
	}
	return false
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

// validatePort checks if port is in valid range (1024-65535)
// Excludes privileged ports 1-1023 which require root/admin access
func validatePort(port int) error {
	if port < 1024 || port > 65535 {
		logger.Error("Invalid port: %d", port)
		logger.ErrorWithSuggestion(
			"Port must be between 1024 and 65535",
			"snag <url> --port 9222",
		)
		return fmt.Errorf("invalid port: %d", port)
	}
	return nil
}

// validateOutputPath checks if the output file path is writable
func validateOutputPath(path string) error {
	if path == "" {
		logger.Error("Output file path cannot be empty")
		logger.ErrorWithSuggestion(
			"Provide a valid file path",
			"snag <url> -o /path/to/output.md",
		)
		return fmt.Errorf("output file path cannot be empty")
	}

	if info, err := os.Stat(path); err == nil && info.IsDir() {
		logger.Error("Output path is a directory, not a file: %s", path)
		logger.ErrorWithSuggestion(
			"Specify a file path, not a directory",
			"snag <url> -o /path/to/file.md",
		)
		return fmt.Errorf("output path is a directory, not a file: %s", path)
	}

	if info, err := os.Stat(path); err == nil {
		if info.Mode().Perm()&0200 == 0 {
			logger.Error("Cannot write to read-only file: %s", path)
			logger.ErrorWithSuggestion(
				"Make file writable or choose different path",
				"chmod u+w "+path,
			)
			return fmt.Errorf("cannot write to read-only file: %s", path)
		}
	}

	dir := filepath.Dir(path)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.Error("Output directory does not exist: %s", dir)
		logger.ErrorWithSuggestion(
			fmt.Sprintf("Directory '%s' not found", dir),
			"snag <url> -o /path/to/existing/dir/output.md",
		)
		return fmt.Errorf("output directory does not exist: %s", dir)
	}

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

// normalizeFormat converts format to lowercase and handles aliases
// Aliases: "markdown" → "md", "txt" → "text"
func normalizeFormat(format string) string {
	format = strings.TrimSpace(format)
	format = strings.ToLower(format)

	// Normalize common aliases
	switch format {
	case "markdown":
		return FormatMarkdown
	case "txt":
		return FormatText
	default:
		return format
	}
}

// validateFormat checks if format is valid (md, html, text, pdf, or png)
func validateFormat(format string) error {
	if format == "" {
		logger.Error("Format cannot be empty")
		logger.ErrorWithSuggestion(
			"Format must be specified",
			fmt.Sprintf("snag <url> --format %s", FormatMarkdown),
		)
		return fmt.Errorf("format cannot be empty")
	}

	// Define valid formats locally for better testability and self-containment
	validFormats := map[string]bool{
		FormatMarkdown: true,
		FormatHTML:     true,
		FormatText:     true,
		FormatPDF:      true,
		FormatPNG:      true,
	}

	if !validFormats[format] {
		logger.Error("Invalid format '%s'. Supported: md, html, text, pdf, png", format)
		logger.ErrorWithSuggestion(
			"Choose a valid format",
			fmt.Sprintf("snag <url> --format %s", FormatMarkdown),
		)
		return fmt.Errorf("invalid format: %s", format)
	}

	return nil
}

// checkExtensionMismatch warns if output file extension doesn't match format
// Returns true if there's a mismatch (for testing), but doesn't return an error
func checkExtensionMismatch(outputFile string, format string) bool {
	if outputFile == "" {
		return false
	}

	ext := strings.ToLower(filepath.Ext(outputFile))
	expectedExt := strings.ToLower(GetFileExtension(format))

	if ext != expectedExt {
		if ext == "" {
			logger.Warning("Writing %s format to file with no extension: %s", format, outputFile)
			return true
		}
		logger.Warning("Writing %s format to file with %s extension: %s", format, ext, outputFile)
		return true
	}

	return false
}

// validateDirectory checks if a directory exists and is writable
func validateDirectory(dir string) error {
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

	if !info.IsDir() {
		logger.Error("Path is not a directory: %s", dir)
		return fmt.Errorf("not a directory: %s", dir)
	}

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

	fullPath := filepath.Join(outputDir, filename)
	cleanPath := filepath.Clean(fullPath)
	cleanDir := filepath.Clean(outputDir)

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

// validateWaitFor validates the wait-for CSS selector string.
// Returns the trimmed selector, or empty string with warning if input is empty.
// flagSet indicates whether the user explicitly provided the flag (prevents spurious warnings).
func validateWaitFor(selector string, flagSet bool) string {
	selector = strings.TrimSpace(selector)

	if selector == "" {
		if flagSet {
			logger.Warning("--wait-for is empty, ignoring")
		}
		return ""
	}

	return selector
}

// validateUserAgent validates and sanitizes the user agent string.
// Returns the sanitized user agent, or empty string with warning if input is empty.
// Sanitizes newlines for security (prevents HTTP header injection).
// flagSet indicates whether the user explicitly provided the flag (prevents spurious warnings).
func validateUserAgent(ua string, flagSet bool) string {
	ua = strings.TrimSpace(ua)

	if ua == "" {
		if flagSet {
			logger.Warning("--user-agent is empty, using default user agent")
		}
		return ""
	}

	// SECURITY: Sanitize newlines to prevent HTTP header injection
	ua = strings.ReplaceAll(ua, "\n", " ")
	ua = strings.ReplaceAll(ua, "\r", " ")

	return ua
}

// validateUserDataDir validates and expands the user data directory path.
// Returns the expanded path, or empty string with warning if input is empty.
// Performs tilde expansion, directory existence check, and permission check.
func validateUserDataDir(path string) (string, error) {
	path = strings.TrimSpace(path)

	if path == "" {
		logger.Warning("--user-data-dir is empty, using default profile")
		return "", nil
	}

	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		if path == "~" {
			path = homeDir
		} else if strings.HasPrefix(path, "~/") {
			path = filepath.Join(homeDir, path[2:])
		}
	}

	// Validate only if directory exists - let Chrome create it if it doesn't
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		// Directory doesn't exist - Chrome will create it
		logger.Verbose("User data directory will be created by browser: %s", path)
		return path, nil
	}
	if err != nil {
		logger.Error("Error accessing user data directory: %s", path)
		return "", fmt.Errorf("error accessing directory: %w", err)
	}

	// Directory exists - validate it's actually a directory
	if !info.IsDir() {
		logger.Error("Path is not a directory: %s", path)
		logger.ErrorWithSuggestion(
			"User data directory must be a directory, not a file",
			"snag --user-data-dir /path/to/directory <url>",
		)
		return "", fmt.Errorf("path is not a directory: %s", path)
	}

	// Test write permissions on existing directory
	testFile, err := os.CreateTemp(path, ".snag-permission-test-*")
	if err != nil {
		logger.Error("Permission denied accessing user data directory: %s", path)
		logger.ErrorWithSuggestion(
			"Cannot read/write to directory",
			fmt.Sprintf("chmod u+rw %s", path),
		)
		return "", fmt.Errorf("permission denied accessing user data directory: %s", path)
	}
	testFilePath := testFile.Name()
	testFile.Close()
	os.Remove(testFilePath)

	return path, nil
}

