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
