// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

// DoctorReport holds all diagnostic information for the doctor command.
type DoctorReport struct {
	SnagVersion   string
	LatestVersion string
	GoVersion     string
	OS            string
	Arch          string
	WorkingDir    string

	BrowserName    string
	BrowserPath    string
	BrowserVersion string
	BrowserError   error

	ProfilePath   string
	ProfileExists bool

	DefaultPortStatus *PortStatus
	CustomPortStatus  *PortStatus // nil if --port not specified

	EnvVars map[string]string
}

// PortStatus contains information about a browser debugging port.
type PortStatus struct {
	Port     int
	Running  bool
	TabCount int
	Error    error
}

// CollectDoctorInfo gathers all diagnostic information.
// Returns a DoctorReport even if some information could not be collected.
func CollectDoctorInfo(customPort int) (*DoctorReport, error) {
	report := &DoctorReport{
		SnagVersion: version,
		GoVersion:   runtime.Version(),
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		EnvVars:     make(map[string]string),
	}

	// Get working directory
	if wd, err := os.Getwd(); err == nil {
		report.WorkingDir = wd
	} else {
		report.WorkingDir = "(unknown)"
	}

	// Collect environment variables
	report.EnvVars["CHROME_PATH"] = os.Getenv("CHROME_PATH")
	report.EnvVars["CHROMIUM_PATH"] = os.Getenv("CHROMIUM_PATH")

	// Collect browser detection info
	bm := NewBrowserManager(BrowserOptions{Port: customPort})

	// Try to find browser
	path, err := bm.findBrowserPath()
	if err != nil {
		report.BrowserError = err
	} else {
		report.BrowserPath = path
		report.BrowserName = bm.browserName

		// Get browser version
		version, err := bm.GetBrowserVersion()
		if err == nil {
			report.BrowserVersion = version
		}

		// Get profile path
		profilePath, exists := bm.GetProfilePath()
		report.ProfilePath = profilePath
		report.ProfileExists = exists
	}

	// Collect connection status for default port 9222
	report.DefaultPortStatus = checkPortConnection(9222)

	// If custom port specified and different from default, check it too
	if customPort != 9222 {
		report.CustomPortStatus = checkPortConnection(customPort)
	}

	// Check latest version from GitHub (with 10s timeout)
	report.LatestVersion = checkLatestVersion()

	return report, nil
}

// checkPortConnection attempts to connect to a browser on the given port
// and returns connection status including tab count.
func checkPortConnection(port int) *PortStatus {
	status := &PortStatus{
		Port:    port,
		Running: false,
	}

	// Create a temporary browser manager for connection check
	bm := NewBrowserManager(BrowserOptions{Port: port})

	// Try to connect with a short timeout
	done := make(chan bool, 1)
	var tabCount int
	var connErr error

	go func() {
		browser, err := bm.connectToExisting()
		if err != nil {
			connErr = err
			done <- false
			return
		}

		// Get tab count
		pages, err := browser.Pages()
		if err == nil {
			tabCount = len(pages)
		}

		done <- true
	}()

	// Wait for connection attempt with timeout
	select {
	case success := <-done:
		if success {
			status.Running = true
			status.TabCount = tabCount
		} else {
			status.Error = connErr
		}
	case <-time.After(3 * time.Second):
		status.Error = fmt.Errorf("connection timeout")
	}

	return status
}

// checkLatestVersion queries the GitHub API for the latest release version.
// Returns the version string (without "v" prefix) or empty string on error.
// Uses a 10-second timeout to prevent hanging.
func checkLatestVersion() string {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get("https://api.github.com/repos/grantcarthew/snag/releases/latest")
	if err != nil {
		// Network error - fail silently
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// API error - fail silently
		return ""
	}

	var release struct {
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		// Parse error - fail silently
		return ""
	}

	// Strip "v" prefix if present
	version := strings.TrimPrefix(release.TagName, "v")
	return version
}

// String implements fmt.Stringer and generates the full diagnostic report.
// This enables testing and future features like --report-issue.
func (dr *DoctorReport) String() string {
	var buf strings.Builder

	buf.WriteString("snag Doctor Report\n")
	buf.WriteString("==================\n")
	buf.WriteString("Repository:    https://github.com/grantcarthew/snag\n")
	buf.WriteString("Report Issue:  https://github.com/grantcarthew/snag/issues/new\n")

	// Version Information
	buf.WriteString(dr.formatSection("Version Information"))
	buf.WriteString(dr.formatItem("snag version", dr.SnagVersion))
	if dr.LatestVersion != "" {
		if dr.LatestVersion != dr.SnagVersion {
			buf.WriteString(dr.formatItem("Latest version", fmt.Sprintf("%s (update available)", dr.LatestVersion)))
		} else {
			buf.WriteString(dr.formatItem("Latest version", dr.LatestVersion))
		}
	}
	buf.WriteString(dr.formatItem("Go version", dr.GoVersion))
	buf.WriteString(dr.formatItem("OS/Arch", fmt.Sprintf("%s/%s", dr.OS, dr.Arch)))

	// Working Directory
	buf.WriteString(dr.formatSection("Working Directory"))
	buf.WriteString(fmt.Sprintf("  %s\n", dr.WorkingDir))

	// Browser Detection
	buf.WriteString(dr.formatSection("Browser Detection"))
	if dr.BrowserError != nil {
		buf.WriteString(dr.formatCheck("Detected", "No Chromium-based browser found", false))
		buf.WriteString(dr.formatItem("Path", "(none)"))
		buf.WriteString(dr.formatItem("Version", "(none)"))
	} else {
		buf.WriteString(dr.formatItem("Detected", dr.BrowserName))
		buf.WriteString(dr.formatItem("Path", dr.BrowserPath))
		if dr.BrowserVersion != "" {
			buf.WriteString(dr.formatItem("Version", dr.BrowserVersion))
		} else {
			buf.WriteString(dr.formatItem("Version", "(unknown)"))
		}
	}

	// Profile Location
	if dr.ProfilePath != "" {
		buf.WriteString(dr.formatSection("Profile Location"))
		buf.WriteString(dr.formatCheck(dr.BrowserName, dr.ProfilePath, dr.ProfileExists))
	}

	// Connection Status
	buf.WriteString(dr.formatSection("Connection Status"))
	if dr.DefaultPortStatus != nil {
		buf.WriteString(dr.formatPortStatus(dr.DefaultPortStatus))
	}
	if dr.CustomPortStatus != nil {
		buf.WriteString(dr.formatPortStatus(dr.CustomPortStatus))
	}

	// Environment Variables
	buf.WriteString(dr.formatSection("Environment Variables"))
	for k, v := range dr.EnvVars {
		if v == "" {
			v = "(not set)"
		}
		buf.WriteString(dr.formatItem(k, v))
	}

	return buf.String()
}

// Print outputs the diagnostic report to stdout.
func (dr *DoctorReport) Print() {
	fmt.Print(dr.String())
}

// formatPortStatus returns a formatted port status line with checkmark.
func (dr *DoctorReport) formatPortStatus(status *PortStatus) string {
	label := fmt.Sprintf("Port %d", status.Port)
	if status.Running {
		value := fmt.Sprintf("Running (%d tabs open)", status.TabCount)
		return dr.formatCheck(label, value, true)
	}
	return dr.formatCheck(label, "Not running", false)
}

// formatSection returns a section header with underline.
func (dr *DoctorReport) formatSection(title string) string {
	return fmt.Sprintf("\n%s\n%s\n", title, strings.Repeat("─", len(title)))
}

// formatItem returns a labeled item with consistent formatting.
func (dr *DoctorReport) formatItem(label, value string) string {
	return fmt.Sprintf("  %-20s %s\n", label+":", value)
}

// formatCheck returns a labeled item with a checkmark or X.
func (dr *DoctorReport) formatCheck(label, value string, ok bool) string {
	mark := "✗"
	if ok {
		mark = "✓"
	}
	return fmt.Sprintf("  %-20s %s %s\n", label+":", mark, value)
}
