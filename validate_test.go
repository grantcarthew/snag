// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func init() {
	// Initialize global logger for tests (discard output)
	logger = &Logger{
		level:  LevelQuiet,
		color:  false,
		writer: io.Discard,
	}
}

func TestValidateURL_Valid(t *testing.T) {
	tests := []string{
		"https://example.com",
		"http://example.com",
		"https://example.com/path",
		"https://example.com/path?query=value",
		"https://subdomain.example.com",
		"https://example.com:8080",
	}

	for _, url := range tests {
		_, err := validateURL(url)
		if err != nil {
			t.Errorf("expected valid URL %q to pass, got error: %v", url, err)
		}
	}
}

func TestValidateURL_Invalid(t *testing.T) {
	tests := []struct {
		url  string
		desc string
	}{
		{"ftp://example.com", "invalid scheme"},
		{"javascript:alert(1)", "javascript scheme"},
		{"://example.com", "malformed URL"},
	}

	for _, tt := range tests {
		_, err := validateURL(tt.url)
		if err == nil {
			t.Errorf("expected invalid URL %q (%s) to fail validation", tt.url, tt.desc)
		}
	}
}

func TestValidateURL_MissingScheme(t *testing.T) {
	// validateURL actually adds https:// if no scheme is present
	tests := []string{
		"example.com",
		"www.example.com",
		"example.com/path",
	}

	for _, url := range tests {
		normalized, err := validateURL(url)
		if err != nil {
			t.Errorf("expected URL without scheme %q to be normalized, got error: %v", url, err)
		}
		if !strings.HasPrefix(normalized, "https://") {
			t.Errorf("expected normalized URL to start with https://, got: %s", normalized)
		}
	}
}

func TestValidateFormat_Valid(t *testing.T) {
	// Test with normalized format values (as they would be after normalizeFormat)
	validFormats := []string{
		FormatMarkdown, // "md"
		FormatHTML,     // "html"
		FormatText,     // "text"
		FormatPDF,      // "pdf"
		FormatPNG,      // "png"
	}

	for _, format := range validFormats {
		err := validateFormat(format)
		if err != nil {
			t.Errorf("expected valid format %q to pass validation, got error: %v", format, err)
		}
	}
}

func TestValidateFormat_Invalid(t *testing.T) {
	// Test with truly invalid formats (not supported by snag)
	// Note: validateFormat expects already-normalized input
	invalidFormats := []string{
		"json",
		"xml",
		"yaml",
		"txt", // Should be normalized to "text" before validation
		"",
		"invalid",
		"markdown", // Should be normalized to "md" before validation
	}

	for _, format := range invalidFormats {
		err := validateFormat(format)
		if err == nil {
			t.Errorf("expected invalid format %q to fail validation", format)
		}
	}
}

func TestNormalizeFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Case normalization
		{"MD", "md"},
		{"HTML", "html"},
		{"Text", "text"},
		{"PDF", "pdf"},
		{"PNG", "png"},
		// Aliases
		{"markdown", "md"},
		{"Markdown", "md"},
		{"MARKDOWN", "md"},
		{"txt", "text"},
		{"TXT", "text"},
		{"Txt", "text"},
		// Already normalized
		{"md", "md"},
		{"html", "html"},
		{"text", "text"},
		{"pdf", "pdf"},
		{"png", "png"},
		// Invalid formats (returned as-is after lowercase conversion)
		{"json", "json"},
		{"xml", "xml"},
	}

	for _, tt := range tests {
		result := normalizeFormat(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeFormat(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestValidateTimeout_Valid(t *testing.T) {
	validTimeouts := []int{1, 30, 60, 120, 3600}

	for _, timeout := range validTimeouts {
		err := validateTimeout(timeout)
		if err != nil {
			t.Errorf("expected valid timeout %d to pass validation, got error: %v", timeout, err)
		}
	}
}

func TestValidateTimeout_Invalid(t *testing.T) {
	invalidTimeouts := []int{-1, 0, -100}

	for _, timeout := range invalidTimeouts {
		err := validateTimeout(timeout)
		if err == nil {
			t.Errorf("expected invalid timeout %d to fail validation", timeout)
		}
	}
}

func TestValidatePort_Valid(t *testing.T) {
	validPorts := []int{1, 80, 443, 8080, 9222, 65535}

	for _, port := range validPorts {
		err := validatePort(port)
		if err != nil {
			t.Errorf("expected valid port %d to pass validation, got error: %v", port, err)
		}
	}
}

func TestValidatePort_Invalid(t *testing.T) {
	invalidPorts := []int{-1, 0, -100, 65536, 99999}

	for _, port := range invalidPorts {
		err := validatePort(port)
		if err == nil {
			t.Errorf("expected invalid port %d to fail validation", port)
		}
	}
}

func TestValidateOutputPath_Valid(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Test with valid writable path
	validPath := tmpDir + "/output.md"
	err := validateOutputPath(validPath)
	if err != nil {
		t.Errorf("expected valid writable path %q to pass validation, got error: %v", validPath, err)
	}
}

func TestValidateOutputPath_NonexistentDirectory(t *testing.T) {
	// Test with path to non-existent directory
	invalidPath := "/nonexistent/directory/output.md"
	err := validateOutputPath(invalidPath)
	if err == nil {
		t.Errorf("expected path with non-existent directory %q to fail validation", invalidPath)
	}
}

func TestValidateOutputPath_ReadOnlyDirectory(t *testing.T) {
	// Create a temporary directory and make it read-only
	tmpDir := t.TempDir()
	readOnlyDir := tmpDir + "/readonly"

	err := os.Mkdir(readOnlyDir, 0755)
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Make directory read-only (no write permission)
	err = os.Chmod(readOnlyDir, 0555)
	if err != nil {
		t.Fatalf("failed to make directory read-only: %v", err)
	}

	// Ensure cleanup restores permissions so TempDir can clean up
	t.Cleanup(func() {
		os.Chmod(readOnlyDir, 0755)
	})

	// Test with path to read-only directory
	invalidPath := readOnlyDir + "/output.md"
	err = validateOutputPath(invalidPath)
	if err == nil {
		t.Errorf("expected path to read-only directory %q to fail validation", invalidPath)
	}
}

// Phase 3 validator tests

func TestValidateDirectory_Valid(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Should pass validation for existing writable directory
	err := validateDirectory(tmpDir)
	if err != nil {
		t.Errorf("expected valid directory %q to pass validation, got error: %v", tmpDir, err)
	}
}

func TestValidateDirectory_NonExistent(t *testing.T) {
	// Test with non-existent directory
	invalidDir := "/nonexistent/test/directory"
	err := validateDirectory(invalidDir)
	if err == nil {
		t.Errorf("expected non-existent directory %q to fail validation", invalidDir)
	}
}

func TestValidateDirectory_NotADirectory(t *testing.T) {
	// Create a temporary file (not a directory)
	tmpFile, err := os.CreateTemp("", "snag-test-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpFilePath)

	// Should fail because it's a file, not a directory
	err = validateDirectory(tmpFilePath)
	if err == nil {
		t.Errorf("expected file path %q to fail directory validation", tmpFilePath)
	}
}

func TestValidateDirectory_ReadOnly(t *testing.T) {
	// Create a temporary directory and make it read-only
	tmpDir := t.TempDir()
	readOnlyDir := tmpDir + "/readonly"

	err := os.Mkdir(readOnlyDir, 0755)
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Make directory read-only (no write permission)
	err = os.Chmod(readOnlyDir, 0555)
	if err != nil {
		t.Fatalf("failed to make directory read-only: %v", err)
	}

	// Ensure cleanup restores permissions so TempDir can clean up
	t.Cleanup(func() {
		os.Chmod(readOnlyDir, 0755)
	})

	// Should fail because directory is not writable
	err = validateDirectory(readOnlyDir)
	if err == nil {
		t.Errorf("expected read-only directory %q to fail validation", readOnlyDir)
	}
}

func TestValidateOutputPathEscape_Safe(t *testing.T) {
	tmpDir := t.TempDir()

	safeTests := []struct {
		outputDir string
		filename  string
		desc      string
	}{
		{tmpDir, "output.md", "simple filename"},
		{tmpDir, "subdir/output.md", "subdirectory"},
		{tmpDir, "./output.md", "current directory"},
		{"/tmp", "/absolute/path.md", "absolute path (ignores outputDir)"},
	}

	for _, tt := range safeTests {
		err := validateOutputPathEscape(tt.outputDir, tt.filename)
		if err != nil {
			t.Errorf("validateOutputPathEscape(%q, %q) [%s] should be safe, got error: %v",
				tt.outputDir, tt.filename, tt.desc, err)
		}
	}
}

func TestValidateOutputPathEscape_Dangerous(t *testing.T) {
	tmpDir := t.TempDir()

	dangerousTests := []struct {
		outputDir string
		filename  string
		desc      string
	}{
		{tmpDir, "../etc/passwd", "parent directory escape"},
		{tmpDir, "../../etc/passwd", "multiple parent escapes"},
		{tmpDir, "subdir/../../etc/passwd", "escape via subdirectory"},
	}

	for _, tt := range dangerousTests {
		err := validateOutputPathEscape(tt.outputDir, tt.filename)
		if err == nil {
			t.Errorf("validateOutputPathEscape(%q, %q) [%s] should be dangerous and fail validation",
				tt.outputDir, tt.filename, tt.desc)
		}
	}
}
