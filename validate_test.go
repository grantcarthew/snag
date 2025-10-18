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
	validFormats := []string{"markdown", "html"}

	for _, format := range validFormats {
		err := validateFormat(format)
		if err != nil {
			t.Errorf("expected valid format %q to pass validation, got error: %v", format, err)
		}
	}
}

func TestValidateFormat_Invalid(t *testing.T) {
	invalidFormats := []string{
		"pdf",
		"text",
		"json",
		"xml",
		"",
		"MARKDOWN",
		"HTML",
	}

	for _, format := range invalidFormats {
		err := validateFormat(format)
		if err == nil {
			t.Errorf("expected invalid format %q to fail validation", format)
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
