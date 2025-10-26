// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"bufio"
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
	// Check for empty string
	if path == "" {
		logger.Error("Output file path cannot be empty")
		logger.ErrorWithSuggestion(
			"Provide a valid file path",
			"snag <url> -o /path/to/output.md",
		)
		return fmt.Errorf("output file path cannot be empty")
	}

	// Check if path is a directory
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		logger.Error("Output path is a directory, not a file: %s", path)
		logger.ErrorWithSuggestion(
			"Specify a file path, not a directory",
			"snag <url> -o /path/to/file.md",
		)
		return fmt.Errorf("output path is a directory, not a file: %s", path)
	}

	// Check if existing file is read-only
	if info, err := os.Stat(path); err == nil {
		// File exists, check if it's writable
		if info.Mode().Perm()&0200 == 0 {
			logger.Error("Cannot write to read-only file: %s", path)
			logger.ErrorWithSuggestion(
				"Make file writable or choose different path",
				"chmod u+w "+path,
			)
			return fmt.Errorf("cannot write to read-only file: %s", path)
		}
	}

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

// normalizeFormat converts format to lowercase and handles aliases
// Aliases: "markdown" → "md", "txt" → "text"
func normalizeFormat(format string) string {
	// Trim whitespace first
	format = strings.TrimSpace(format)

	// Convert to lowercase for case-insensitive matching
	format = strings.ToLower(format)

	// Handle aliases
	aliases := map[string]string{
		"markdown": FormatMarkdown, // "markdown" → "md"
		"txt":      FormatText,     // "txt" → "text"
	}

	if normalized, ok := aliases[format]; ok {
		return normalized
	}

	return format
}

// validateFormat checks if format is valid (md, html, text, pdf, or png)
func validateFormat(format string) error {
	// Check for empty format
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

	// Check for mismatch
	if ext != expectedExt {
		// Special case: no extension at all
		if ext == "" {
			logger.Warning("Writing %s format to file with no extension: %s", format, outputFile)
			return true
		}
		// Extension mismatch
		logger.Warning("Writing %s format to file with %s extension: %s", format, ext, outputFile)
		return true
	}

	return false
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

// validateWaitFor validates the wait-for CSS selector string.
// Returns the trimmed selector, or empty string with warning if input is empty.
func validateWaitFor(selector string) string {
	// Trim whitespace
	selector = strings.TrimSpace(selector)

	// Empty selector means don't wait (not an error, just a warning if explicitly set)
	if selector == "" {
		logger.Warning("--wait-for is empty, ignoring")
		return ""
	}

	return selector
}

// validateUserAgent validates and sanitizes the user agent string.
// Returns the sanitized user agent, or empty string with warning if input is empty.
// Sanitizes newlines for security (prevents HTTP header injection).
func validateUserAgent(ua string) string {
	// Trim whitespace
	ua = strings.TrimSpace(ua)

	// Empty user agent means use browser default (not an error, just a warning)
	if ua == "" {
		logger.Warning("--user-agent is empty, using default user agent")
		return ""
	}

	// Sanitize newlines for security (prevents HTTP header injection)
	// Replace \n and \r with spaces
	ua = strings.ReplaceAll(ua, "\n", " ")
	ua = strings.ReplaceAll(ua, "\r", " ")

	return ua
}

// validateUserDataDir validates and expands the user data directory path.
// Returns the expanded path, or empty string with warning if input is empty.
// Performs tilde expansion, directory existence check, and permission check.
func validateUserDataDir(path string) (string, error) {
	// Trim whitespace
	path = strings.TrimSpace(path)

	// Empty path means use browser default (not an error, just a warning)
	if path == "" {
		logger.Warning("--user-data-dir is empty, using default profile")
		return "", nil
	}

	// Expand tilde to home directory
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

	// Check if path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		logger.Error("User data directory does not exist: %s", path)
		logger.ErrorWithSuggestion(
			fmt.Sprintf("Directory '%s' not found", path),
			fmt.Sprintf("mkdir -p %s && snag --user-data-dir %s <url>", path, path),
		)
		return "", fmt.Errorf("user data directory does not exist: %s", path)
	}
	if err != nil {
		logger.Error("Error accessing user data directory: %s", path)
		return "", fmt.Errorf("error accessing directory: %w", err)
	}

	// Check if it's a directory (not a file)
	if !info.IsDir() {
		logger.Error("Path is not a directory: %s", path)
		logger.ErrorWithSuggestion(
			"User data directory must be a directory, not a file",
			"snag --user-data-dir /path/to/directory <url>",
		)
		return "", fmt.Errorf("path is not a directory: %s", path)
	}

	// Check if directory is readable and writable
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

// loadURLsFromFile reads and parses a URL file, returning a list of valid URLs.
// File format supports:
//   - Full-line comments starting with # or //
//   - Inline comments with " #" or " //"
//   - Blank lines (ignored)
//   - Auto-prepends https:// if no scheme present
//   - Invalid URLs are logged as warnings and skipped
func loadURLsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		logger.Error("Failed to open URL file: %s", filename)
		return nil, fmt.Errorf("failed to open URL file: %w", err)
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip full-line comments
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		// Handle inline comments
		hasComment := false
		for _, marker := range []string{" #", " //"} {
			if idx := strings.Index(line, marker); idx != -1 {
				line = strings.TrimSpace(line[:idx])
				hasComment = true
				break
			}
		}

		// Check for space without comment marker (formatting error)
		if !hasComment && strings.Contains(line, " ") {
			logger.Warning("Line %d: URL contains space without comment marker - skipping: %s", lineNum, line)
			continue
		}

		// Auto-prepend https:// if missing protocol
		if !strings.HasPrefix(line, "http://") && !strings.HasPrefix(line, "https://") && !strings.HasPrefix(line, "file://") {
			line = "https://" + line
		}

		// Validate URL
		if _, err := validateURL(line); err != nil {
			logger.Warning("Line %d: Invalid URL - skipping: %s", lineNum, scanner.Text())
			continue
		}

		urls = append(urls, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	if len(urls) == 0 {
		return nil, ErrNoValidURLs
	}

	logger.Verbose("Loaded %d URLs from %s", len(urls), filename)
	return urls, nil
}
