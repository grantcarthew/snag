// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"testing"
)

func TestDetectBrowserName(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		// Chrome
		{"chrome linux", "/usr/bin/google-chrome", "Chrome"},
		{"chrome stable", "/usr/bin/google-chrome-stable", "Chrome"},
		{"chrome macos", "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", "Chrome"},
		{"chrome windows", "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe", "Chrome"},
		{"chrome uppercase", "/usr/bin/CHROME", "Chrome"},
		{"chrome mixed case", "/usr/bin/Chrome", "Chrome"},

		// Chromium
		{"chromium linux", "/usr/bin/chromium", "Chromium"},
		{"chromium browser", "/usr/bin/chromium-browser", "Chromium"},
		{"chromium macos", "/Applications/Chromium.app/Contents/MacOS/Chromium", "Chromium"},
		{"chromium windows", "C:\\Program Files\\Chromium\\chromium.exe", "Chromium"},

		// Ungoogled Chromium (must be detected before regular Chromium)
		{"ungoogled chromium", "/usr/bin/ungoogled-chromium", "Ungoogled-Chromium"},
		{"ungoogled chromium app", "/Applications/Ungoogled Chromium.app", "Ungoogled-Chromium"},

		// Edge
		{"edge linux", "/usr/bin/microsoft-edge", "Edge"},
		{"edge macos", "/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge", "Edge"},
		{"edge windows", "C:\\Program Files\\Microsoft\\Edge\\Application\\msedge.exe", "Edge"},
		{"msedge", "/usr/bin/msedge", "Edge"},

		// Brave
		{"brave linux", "/usr/bin/brave", "Brave"},
		{"brave browser", "/usr/bin/brave-browser", "Brave"},
		{"brave macos", "/Applications/Brave Browser.app/Contents/MacOS/Brave Browser", "Brave"},
		{"brave windows", "C:\\Program Files\\BraveSoftware\\Brave-Browser\\brave.exe", "Brave"},

		// Opera
		{"opera linux", "/usr/bin/opera", "Opera"},
		{"opera macos", "/Applications/Opera.app/Contents/MacOS/Opera", "Opera"},

		// Vivaldi
		{"vivaldi linux", "/usr/bin/vivaldi", "Vivaldi"},
		{"vivaldi macos", "/Applications/Vivaldi.app/Contents/MacOS/Vivaldi", "Vivaldi"},

		// Arc
		{"arc macos", "/Applications/Arc.app/Contents/MacOS/Arc", "Arc"},

		// Yandex
		{"yandex linux", "/usr/bin/yandex-browser", "Yandex"},
		{"yandex macos", "/Applications/Yandex.app/Contents/MacOS/Yandex", "Yandex"},

		// Thorium
		{"thorium linux", "/usr/bin/thorium", "Thorium"},
		{"thorium browser", "/usr/bin/thorium-browser", "Thorium"},

		// Slimjet
		{"slimjet linux", "/usr/bin/slimjet", "Slimjet"},

		// Cent
		{"cent browser", "/usr/bin/cent-browser", "Cent"},

		// Extension handling
		{"exe extension", "C:\\chrome.exe", "Chrome"},
		{"app extension", "/Applications/Chrome.app", "Chrome"},

		// Path with directories
		{"deep path", "/home/user/.local/bin/chrome", "Chrome"},
		{"user directory", "/home/chrome-user/bin/chromium", "Chromium"},

		// Fallback cases
		{"unknown browser", "/usr/bin/firefox", "Firefox"},
		{"custom browser", "/usr/bin/mybrowser", "Mybrowser"},
		{"empty path", "", "."},      // filepath.Base("") returns "."
		{"just slash", "/", "/"},      // filepath.Base("/") returns "/"

		// Case sensitivity
		{"uppercase chromium", "/usr/bin/CHROMIUM", "Chromium"},
		{"mixed case edge", "/usr/bin/MsEdge", "Edge"},

		// Order of precedence (Chrome vs Chromium)
		{"chrome not chromium", "/usr/bin/chrome", "Chrome"},
		{"chromium not chrome", "/usr/bin/chromium", "Chromium"},

		// Complex paths
		{"windows complex", "C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe", "Chrome"},
		{"macos bundle", "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", "Chrome"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectBrowserName(tt.path)
			if result != tt.expected {
				t.Errorf("detectBrowserName(%q) = %q, expected %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestDetectBrowserName_OrderOfPrecedence(t *testing.T) {
	// Test that more specific matches take precedence over generic ones
	tests := []struct {
		name     string
		path     string
		expected string
		notExpected string
	}{
		{"ungoogled before chromium", "/usr/bin/ungoogled-chromium", "Ungoogled-Chromium", "Chromium"},
		{"chrome before chromium", "/usr/bin/google-chrome", "Chrome", "Chromium"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectBrowserName(tt.path)
			if result != tt.expected {
				t.Errorf("detectBrowserName(%q) = %q, expected %q", tt.path, result, tt.expected)
			}
			if result == tt.notExpected {
				t.Errorf("detectBrowserName(%q) = %q, should not match %q", tt.path, result, tt.notExpected)
			}
		})
	}
}

func TestDetectBrowserName_ExtensionStripping(t *testing.T) {
	// Test that .exe and .app extensions are properly stripped
	tests := []struct {
		name         string
		pathWithExt  string
		pathNoExt    string
		expectedName string
	}{
		{".exe stripping", "C:\\chrome.exe", "/usr/bin/chrome", "Chrome"},
		{".app stripping", "/Applications/Chrome.app", "/usr/bin/Chrome", "Chrome"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultWithExt := detectBrowserName(tt.pathWithExt)
			resultNoExt := detectBrowserName(tt.pathNoExt)

			if resultWithExt != tt.expectedName {
				t.Errorf("detectBrowserName(%q) = %q, expected %q", tt.pathWithExt, resultWithExt, tt.expectedName)
			}
			if resultNoExt != tt.expectedName {
				t.Errorf("detectBrowserName(%q) = %q, expected %q", tt.pathNoExt, resultNoExt, tt.expectedName)
			}
			if resultWithExt != resultNoExt {
				t.Errorf("Extension stripping failed: with ext = %q, without ext = %q", resultWithExt, resultNoExt)
			}
		})
	}
}

func TestDetectBrowserName_FallbackBehavior(t *testing.T) {
	// Test fallback behavior for unknown browsers
	tests := []struct {
		name     string
		path     string
		contains string // Should contain this substring
	}{
		{"capitalizes first letter", "/usr/bin/firefox", "F"},
		{"preserves rest", "/usr/bin/firefox", "irefox"},
		{"handles single char", "/usr/bin/x", "X"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectBrowserName(tt.path)
			if len(result) == 0 {
				t.Errorf("detectBrowserName(%q) returned empty string", tt.path)
			}
		})
	}
}
