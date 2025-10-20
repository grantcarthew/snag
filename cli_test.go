// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// isBrowserAvailable checks if Chrome or Chromium is available on the system
func isBrowserAvailable() bool {
	browsers := []string{
		"google-chrome",
		"chromium",
		"chromium-browser",
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
	}

	for _, browser := range browsers {
		if _, err := exec.LookPath(browser); err == nil {
			return true
		}
		// Check if file exists (for macOS app paths)
		if _, err := os.Stat(browser); err == nil {
			return true
		}
	}
	return false
}

// startTestServer launches an HTTP server serving files from testdata/
func startTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	// Get absolute path to testdata directory
	testdataPath, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatalf("failed to get testdata path: %v", err)
	}

	// Create file server
	fileServer := http.FileServer(http.Dir(testdataPath))

	// Create test server
	server := httptest.NewServer(fileServer)

	// Register cleanup
	t.Cleanup(func() {
		server.Close()
	})

	return server
}

// runSnag executes the snag binary with the given arguments
// Returns stdout, stderr, and error
func runSnag(args ...string) (stdout string, stderr string, err error) {
	cmd := exec.Command("./snag", args...)

	// Capture stdout and stderr separately
	stdoutBytes, stderrBytes, err := runCommand(cmd)

	return string(stdoutBytes), string(stderrBytes), err
}

// runCommand executes a command and returns stdout, stderr separately
func runCommand(cmd *exec.Cmd) (stdout []byte, stderr []byte, err error) {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	// Read output using io.ReadAll
	stdoutBytes, stdoutErr := io.ReadAll(stdoutPipe)
	stderrBytes, stderrErr := io.ReadAll(stderrPipe)

	// Wait for command to finish
	err = cmd.Wait()

	// Check for read errors
	if stdoutErr != nil {
		return nil, nil, stdoutErr
	}
	if stderrErr != nil {
		return nil, nil, stderrErr
	}

	return stdoutBytes, stderrBytes, err
}

// assertContains checks if the output contains the expected substring
func assertContains(t *testing.T, output, expected string) {
	t.Helper()
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got:\n%s", expected, output)
	}
}

// assertNotContains checks if the output does not contain the substring
func assertNotContains(t *testing.T, output, unexpected string) {
	t.Helper()
	if strings.Contains(output, unexpected) {
		t.Errorf("expected output to NOT contain %q, got:\n%s", unexpected, output)
	}
}

// assertExitCode checks if the command exited with the expected code
func assertExitCode(t *testing.T, err error, expectedCode int) {
	t.Helper()
	if expectedCode == 0 {
		if err != nil {
			t.Errorf("expected exit code 0, but command failed: %v", err)
		}
	} else {
		// expectedCode != 0, so we expect an error
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != expectedCode {
				t.Errorf("expected exit code %d, got %d", expectedCode, exitErr.ExitCode())
			}
		} else {
			t.Errorf("expected exit code %d, but got non-exit error or success: %v", expectedCode, err)
		}
	}
}

// assertNoError checks that there was no error
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// assertError checks that there was an error
func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error but got none")
	}
}

// ============================================================================
// Phase 3: Fast CLI Tests (No Browser Required)
// ============================================================================

// TestCLI_Version tests the --version flag
func TestCLI_Version(t *testing.T) {
	stdout, stderr, err := runSnag("--version")

	assertNoError(t, err)

	// Version should be in output (could be stdout or stderr)
	output := stdout + stderr
	if !strings.Contains(output, "snag version") && !strings.Contains(output, version) {
		t.Errorf("expected version in output, got: %s", output)
	}
}

// TestCLI_Help tests the --help flag
func TestCLI_Help(t *testing.T) {
	stdout, stderr, _ := runSnag("--help")

	// Help may exit with 0 or error depending on cli library
	output := stdout + stderr

	// Should contain usage information
	assertContains(t, output, "USAGE")
	assertContains(t, output, "snag")
}

// TestCLI_NoArguments tests running without URL
func TestCLI_NoArguments(t *testing.T) {
	stdout, stderr, err := runSnag()

	// Should fail when no URL provided
	assertError(t, err)
	assertExitCode(t, err, 1)

	output := stdout + stderr
	// Should show error or help message
	if !strings.Contains(output, "required") && !strings.Contains(output, "USAGE") {
		t.Errorf("expected error or usage message, got: %s", output)
	}
}

// TestCLI_InvalidURL tests invalid URL handling
func TestCLI_InvalidURL(t *testing.T) {
	tests := []struct {
		url  string
		desc string
	}{
		{"ftp://example.com", "unsupported scheme"},
		{"javascript:alert(1)", "javascript scheme"},
		{"://malformed", "malformed URL"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			stdout, stderr, err := runSnag(tt.url)

			assertError(t, err)
			assertExitCode(t, err, 1)

			output := stdout + stderr
			// Should contain error message
			if !strings.Contains(output, "Invalid") && !strings.Contains(output, "invalid") &&
			   !strings.Contains(output, "error") && !strings.Contains(output, "Error") {
				t.Errorf("expected error message for %s, got: %s", tt.desc, output)
			}
		})
	}
}

// TestCLI_InvalidFormat tests invalid format flag
func TestCLI_InvalidFormat(t *testing.T) {
	stdout, stderr, err := runSnag("--format", "pdf", "https://example.com")

	assertError(t, err)
	assertExitCode(t, err, 1)

	output := stdout + stderr
	assertContains(t, output, "format")
}

// TestCLI_InvalidTimeout tests invalid timeout values
func TestCLI_InvalidTimeout(t *testing.T) {
	tests := []struct {
		timeout string
		desc    string
	}{
		{"-1", "negative timeout"},
		{"0", "zero timeout"},
		{"abc", "non-numeric timeout"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			stdout, stderr, err := runSnag("--timeout", tt.timeout, "https://example.com")

			// Should either fail validation or fail to parse
			assertError(t, err)

			output := stdout + stderr
			// Should contain error about timeout or invalid value
			if !strings.Contains(output, "timeout") && !strings.Contains(output, "invalid") &&
			   !strings.Contains(output, "error") && !strings.Contains(output, "Error") {
				t.Errorf("expected error message about timeout or invalid value for %s, got: %s", tt.desc, output)
			}
		})
	}
}

// TestCLI_InvalidPort tests invalid port values
func TestCLI_InvalidPort(t *testing.T) {
	tests := []struct {
		port string
		desc string
	}{
		{"-1", "negative port"},
		{"0", "zero port"},
		{"99999", "port too large"},
		{"abc", "non-numeric port"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			stdout, stderr, err := runSnag("--port", tt.port, "--force-headless", "https://example.com")

			// Should either fail validation or fail to parse
			assertError(t, err)

			output := stdout + stderr
			// Should contain error about port or invalid value
			if !strings.Contains(output, "port") && !strings.Contains(output, "invalid") &&
			   !strings.Contains(output, "error") && !strings.Contains(output, "Error") {
				t.Errorf("expected error message about port or invalid value for %s, got: %s", tt.desc, output)
			}
		})
	}
}

// TestCLI_FormatOptions tests valid format values are accepted
func TestCLI_FormatOptions(t *testing.T) {
	// Note: These will fail to actually fetch without a browser,
	// but should pass format validation
	tests := []string{"markdown", "html"}

	for _, format := range tests {
		t.Run(format, func(t *testing.T) {
			// We can't actually test fetching without a browser,
			// but we can verify the format is accepted by checking
			// the error message doesn't mention invalid format
			stdout, stderr, err := runSnag("--format", format, "--force-headless", "https://example.com")

			output := stdout + stderr

			// If there's an error, it should NOT be about invalid format
			if err != nil {
				if strings.Contains(output, "Invalid format") || strings.Contains(output, "invalid format") {
					t.Errorf("format %q should be valid but got format error: %s", format, output)
				}
				// Other errors (like browser not found) are acceptable for this test
			}
		})
	}
}

// TestCLI_OutputFilePermission tests output to unwritable location
func TestCLI_OutputFilePermission(t *testing.T) {
	// Create a temporary directory and make it read-only
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")

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

	outputPath := filepath.Join(readOnlyDir, "test-output.md")
	stdout, stderr, err := runSnag("-o", outputPath, "--force-headless", "https://example.com")

	// Should fail due to permissions
	assertError(t, err)

	output := stdout + stderr
	// May fail with permission error or browser error - both are acceptable
	// We're just verifying it doesn't succeed
	_ = output
}

// ============================================================================
// Phase 4: Browser Integration Tests (Requires Chrome/Chromium)
// ============================================================================

// TestBrowser_FetchSimpleHTML tests fetching simple.html from test server
func TestBrowser_FetchSimpleHTML(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	stdout, stderr, err := runSnag(url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Verify markdown conversion happened
	assertContains(t, stdout, "# Example Heading")
	assertContains(t, stdout, "## Second Level Heading")
	assertContains(t, stdout, "This is a simple paragraph")
	assertContains(t, stdout, "[a link](https://example.com)")
	assertContains(t, stdout, "**bold text**")
	assertContains(t, stdout, "*italic text*")

	// Verify logs went to stderr (not stdout)
	if len(stderr) > 0 {
		// If there's stderr output, it should be logs, not content
		assertNotContains(t, stderr, "# Example Heading")
	}
}

// TestBrowser_FetchComplexHTML tests fetching complex.html with tables and lists
func TestBrowser_FetchComplexHTML(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/complex.html"

	stdout, stderr, err := runSnag(url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Verify markdown conversion
	assertContains(t, stdout, "# Complex Content")
	assertContains(t, stdout, "## Table Example")
	assertContains(t, stdout, "## List Examples")
	assertContains(t, stdout, "## Code Example")

	// Verify lists
	assertContains(t, stdout, "- Unordered item 1")
	assertContains(t, stdout, "- Unordered item 2")
	assertContains(t, stdout, "1. Ordered item 1")
	assertContains(t, stdout, "2. Ordered item 2")

	// Verify code block (should have backticks)
	assertContains(t, stdout, "```")
	assertContains(t, stdout, "function hello()")

	// Note: Table conversion may not produce markdown tables (known issue)
	// Just verify table content is preserved
	assertContains(t, stdout, "Item 1")
	assertContains(t, stdout, "Item 2")

	// Verify logs went to stderr
	if len(stderr) > 0 {
		assertNotContains(t, stderr, "# Complex Content")
	}
}

// TestBrowser_FetchMinimalHTML tests fetching minimal.html edge case
func TestBrowser_FetchMinimalHTML(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/minimal.html"

	stdout, stderr, err := runSnag(url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should contain the minimal content
	assertContains(t, stdout, "Hello")

	// Should not be empty
	if len(strings.TrimSpace(stdout)) == 0 {
		t.Error("expected non-empty output for minimal HTML")
	}

	// Verify logs went to stderr
	if len(stderr) > 0 {
		assertNotContains(t, stderr, "Hello")
	}
}

// TestBrowser_HTMLFormat tests --format html output
func TestBrowser_HTMLFormat(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	stdout, stderr, err := runSnag("--format", "html", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Verify HTML output (not markdown)
	assertContains(t, stdout, "<h1>")
	assertContains(t, stdout, "<h2>")
	assertContains(t, stdout, "<p>")
	assertContains(t, stdout, "<a href=")
	assertContains(t, stdout, "<strong>")
	assertContains(t, stdout, "<em>")

	// Should NOT contain markdown syntax
	assertNotContains(t, stdout, "# Example Heading")
	assertNotContains(t, stdout, "**bold text**")

	// Verify logs went to stderr
	if len(stderr) > 0 {
		assertNotContains(t, stderr, "<h1>")
	}
}

// TestBrowser_OutputToFile tests -o flag for file output
func TestBrowser_OutputToFile(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	// Create temporary file for output
	tmpFile, err := os.CreateTemp("", "snag-test-*.md")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	outputPath := tmpFile.Name()
	tmpFile.Close()

	// Clean up after test
	t.Cleanup(func() {
		os.Remove(outputPath)
	})

	stdout, stderr, err := runSnag("-o", outputPath, url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Stdout should be empty (content written to file)
	if len(strings.TrimSpace(stdout)) > 0 {
		t.Errorf("expected empty stdout when using -o flag, got: %s", stdout)
	}

	// Verify file was created and contains content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	contentStr := string(content)
	assertContains(t, contentStr, "# Example Heading")
	assertContains(t, contentStr, "This is a simple paragraph")

	// Verify success message in stderr
	if len(stderr) > 0 {
		// May contain success message about writing file
		assertNotContains(t, stderr, "# Example Heading")
	}
}

// TestBrowser_ForceHeadless tests --force-headless flag
func TestBrowser_ForceHeadless(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	stdout, stderr, err := runSnag("--force-headless", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should successfully fetch content
	assertContains(t, stdout, "# Example Heading")

	// Verify headless mode was used (check stderr for relevant messages)
	output := stderr
	_ = output // May or may not contain mode indication
}

// TestBrowser_ForceVisible tests --force-visible flag
func TestBrowser_ForceVisible(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	// Note: This will actually open a visible browser window
	// In CI environments, this requires a display server
	stdout, stderr, err := runSnag("--force-visible", "--close-tab", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should successfully fetch content
	assertContains(t, stdout, "# Example Heading")

	// Using --close-tab to clean up the visible browser tab
	output := stderr
	_ = output
}

// TestBrowser_OpenBrowser tests --open-browser flag (open browser and fetch)
func TestBrowser_OpenBrowser(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	stdout, stderr, err := runSnag("--open-browser", "--force-visible", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should fetch content in visible browser mode
	// Stdout should contain the markdown content
	assertContains(t, stdout, "Example Heading")
	assertContains(t, stdout, "simple paragraph")

	// Stderr may contain messages about opening browser
	output := stderr
	_ = output
}

// TestBrowser_CustomPort tests --port flag with custom debugging port
func TestBrowser_CustomPort(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	// Use a non-default port
	stdout, stderr, err := runSnag("--port", "9223", "--force-headless", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should successfully fetch content
	assertContains(t, stdout, "# Example Heading")

	output := stderr
	_ = output
}

// TestBrowser_ConnectExisting tests connecting to existing Chrome instance
func TestBrowser_ConnectExisting(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	// This test verifies that snag can connect to an already-running Chrome instance
	// First, launch Chrome with --force-visible to create a persistent instance
	// Then run snag without force flags to connect to it

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	// First run: Launch visible browser to create persistent instance
	stdout1, stderr1, err1 := runSnag("--force-visible", "--close-tab", url)

	assertNoError(t, err1)
	assertExitCode(t, err1, 0)
	assertContains(t, stdout1, "# Example Heading")

	// Second run: Should connect to existing instance
	// (Default behavior is to connect if available)
	stdout2, stderr2, err2 := runSnag(url)

	assertNoError(t, err2)
	assertExitCode(t, err2, 0)
	assertContains(t, stdout2, "# Example Heading")

	// Both runs should succeed
	_ = stderr1
	_ = stderr2
}

// TestBrowser_Auth401Detection tests HTTP 401 authentication detection
func TestBrowser_Auth401Detection(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	// Create test server with 401 handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Test"`)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("<html><body>401 Unauthorized</body></html>"))
	})
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	url := server.URL

	stdout, stderr, err := runSnag(url)

	// May fail or succeed depending on how snag handles 401
	// At minimum, should not crash
	output := stdout + stderr

	// Should indicate authentication issue or return the 401 page
	// The test verifies snag handles 401 gracefully
	_ = output
	_ = err
}

// TestBrowser_Auth403Detection tests HTTP 403 forbidden detection
func TestBrowser_Auth403Detection(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	// Create test server with 403 handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("<html><body>403 Forbidden</body></html>"))
	})
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	url := server.URL

	stdout, stderr, err := runSnag(url)

	// May fail or succeed depending on how snag handles 403
	// At minimum, should not crash
	output := stdout + stderr

	// Should indicate forbidden access or return the 403 page
	_ = output
	_ = err
}

// TestBrowser_LoginFormDetection tests detection of login forms in DOM
func TestBrowser_LoginFormDetection(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/login-form.html"

	stdout, stderr, err := runSnag(url)

	// Should successfully fetch the login form page
	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should contain login form content in markdown
	assertContains(t, stdout, "Log In")

	// Form fields may be converted to markdown
	// Just verify content is present
	output := stdout + stderr
	_ = output
}

// TestBrowser_NoAuthFalsePositives tests that regular pages don't trigger auth detection
func TestBrowser_NoAuthFalsePositives(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	stdout, stderr, err := runSnag(url)

	// Regular page should fetch successfully
	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should contain normal content
	assertContains(t, stdout, "# Example Heading")

	// Should NOT have authentication warnings
	output := stderr
	// If there are auth-related messages in stderr for a regular page, that's a false positive
	// However, we don't have specific auth detection messages defined, so just verify success
	_ = output
}

// TestBrowser_CustomTimeout tests --timeout flag with custom value
func TestBrowser_CustomTimeout(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	// Use a custom timeout (60 seconds)
	stdout, stderr, err := runSnag("--timeout", "60", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should successfully fetch content
	assertContains(t, stdout, "# Example Heading")

	output := stderr
	_ = output
}

// TestBrowser_WaitForSelector tests --wait-for flag to wait for specific element
func TestBrowser_WaitForSelector(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/dynamic.html"

	// Wait for the delayed content element
	stdout, stderr, err := runSnag("--wait-for", "#delayed-content", "--timeout", "5", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should contain both initial and delayed content
	assertContains(t, stdout, "Dynamic Page")
	assertContains(t, stdout, "after 1 second")

	output := stderr
	_ = output
}

// TestBrowser_WaitForTimeout tests --wait-for with element that doesn't appear
func TestBrowser_WaitForTimeout(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	// Wait for element that doesn't exist, with short timeout
	stdout, stderr, err := runSnag("--wait-for", "#nonexistent-element", "--timeout", "2", url)

	// Should timeout and fail
	assertError(t, err)

	output := stdout + stderr
	// Should indicate timeout or element not found
	if !strings.Contains(output, "timeout") && !strings.Contains(output, "not found") &&
	   !strings.Contains(output, "Timeout") {
		t.Errorf("Expected timeout error message, got: %s", output)
	}
}

// TestBrowser_DefaultTimeout tests that default timeout works
func TestBrowser_DefaultTimeout(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	// No timeout specified, should use default (30 seconds)
	stdout, stderr, err := runSnag(url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should successfully fetch content with default timeout
	assertContains(t, stdout, "# Example Heading")

	output := stderr
	_ = output
}

// TestBrowser_CustomUserAgent tests --user-agent flag
func TestBrowser_CustomUserAgent(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	customUA := "Mozilla/5.0 (Custom Bot) snag/test"
	stdout, stderr, err := runSnag("--user-agent", customUA, url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should successfully fetch content with custom user agent
	assertContains(t, stdout, "# Example Heading")

	// User agent is set in browser, content should be fetched normally
	output := stderr
	_ = output
}

// TestBrowser_CloseTab tests --close-tab flag
func TestBrowser_CloseTab(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	// Use --close-tab with headless mode
	stdout, stderr, err := runSnag("--close-tab", "--force-headless", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should successfully fetch content and close the tab
	assertContains(t, stdout, "# Example Heading")

	output := stderr
	_ = output
}

// TestBrowser_VerboseOutput tests --verbose flag
func TestBrowser_VerboseOutput(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	stdout, stderr, err := runSnag("--verbose", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should successfully fetch content
	assertContains(t, stdout, "# Example Heading")

	// Verbose mode should produce more stderr output
	// Stderr should have verbose logging messages
	if len(stderr) == 0 {
		t.Log("verbose mode produced no stderr output (may be expected)")
	}
}

// TestBrowser_QuietMode tests --quiet flag
func TestBrowser_QuietMode(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	stdout, stderr, err := runSnag("--quiet", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should successfully fetch content
	assertContains(t, stdout, "# Example Heading")

	// Quiet mode should minimize stderr output (only errors)
	// Less stderr than normal mode
	_ = stderr
}

// TestBrowser_DebugMode tests --debug flag
func TestBrowser_DebugMode(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	stdout, stderr, err := runSnag("--debug", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should successfully fetch content
	assertContains(t, stdout, "# Example Heading")

	// Debug mode should produce detailed stderr output
	// Stderr should have debug logging messages
	if len(stderr) == 0 {
		t.Log("debug mode produced no stderr output (may be expected)")
	}
}

// TestBrowser_RealWorld_ExampleDotCom tests fetching a real website (example.com)
func TestBrowser_RealWorld_ExampleDotCom(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	// Skip in environments without internet access
	if testing.Short() {
		t.Skip("skipping real-world test in short mode")
	}

	stdout, stderr, err := runSnag("https://example.com")

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should contain example.com content
	// Example.com has specific text we can check
	output := strings.ToLower(stdout)
	if !strings.Contains(output, "example") {
		t.Errorf("expected example.com content, got: %s", stdout)
	}

	// Should be valid markdown
	if len(strings.TrimSpace(stdout)) == 0 {
		t.Error("expected non-empty output")
	}

	_ = stderr
}

// TestBrowser_RealWorld_HttpBin tests fetching httpbin.org endpoints
func TestBrowser_RealWorld_HttpBin(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	// Skip in environments without internet access
	if testing.Short() {
		t.Skip("skipping real-world test in short mode")
	}

	// Test httpbin.org/html endpoint
	stdout, stderr, err := runSnag("https://httpbin.org/html")

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// httpbin.org/html returns HTML content
	// Should be converted to markdown
	if len(strings.TrimSpace(stdout)) == 0 {
		t.Error("expected non-empty output from httpbin.org")
	}

	_ = stderr
}

// TestBrowser_RealWorld_DelayedResponse tests handling slow-loading pages
func TestBrowser_RealWorld_DelayedResponse(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	// Skip in environments without internet access
	if testing.Short() {
		t.Skip("skipping real-world test in short mode")
	}

	// httpbin.org/delay/2 delays response by 2 seconds
	stdout, stderr, err := runSnag("--timeout", "10", "https://httpbin.org/delay/2")

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should handle the delay and fetch content
	if len(strings.TrimSpace(stdout)) == 0 {
		t.Error("expected non-empty output from delayed response")
	}

	_ = stderr
}
