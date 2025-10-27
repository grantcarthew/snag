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
	"regexp"
	"strings"
	"time"
)

var (
	// Compile regex patterns once at package level for performance
	slugNonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)
	slugMultipleHyphens = regexp.MustCompile(`-+`)
)

// SlugifyTitle converts a page title into a URL-safe slug.
// Rules:
//   - Convert to lowercase
//   - Replace all non-alphanumeric characters with hyphen
//   - Collapse multiple consecutive hyphens to single hyphen
//   - Trim leading and trailing hyphens
//   - Truncate to maxLen characters
func SlugifyTitle(title string, maxLen int) string {
	slug := strings.ToLower(title)

	// Keep only a-z, 0-9, and replace everything else with -
	slug = slugNonAlphanumeric.ReplaceAllString(slug, "-")

	slug = slugMultipleHyphens.ReplaceAllString(slug, "-")

	slug = strings.Trim(slug, "-")

	if len(slug) > maxLen {
		slug = slug[:maxLen]
		// Trim trailing hyphen if truncation created one
		slug = strings.TrimRight(slug, "-")
	}

	return slug
}

// GenerateURLSlug creates a slug from URL hostname as fallback when title is empty.
// Applies same slugification rules as SlugifyTitle.
func GenerateURLSlug(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		// Fallback to "page" if URL can't be parsed
		return "page"
	}

	hostname := parsedURL.Host
	if hostname == "" {
		// For file:// URLs or other cases without host
		return "page"
	}

	// Apply slugification to hostname
	return SlugifyTitle(hostname, 80)
}

// GetFileExtension returns the file extension for a given format.
// Supported formats: md, html, text, pdf, png
func GetFileExtension(format string) string {
	switch format {
	case FormatMarkdown:
		return ".md"
	case FormatHTML:
		return ".html"
	case FormatText:
		return ".txt"
	case FormatPDF:
		return ".pdf"
	case FormatPNG:
		return ".png"
	default:
		return ".md" // Default to markdown
	}
}

// GenerateFilename creates an auto-generated filename using timestamp and page title.
// Format: yyyy-mm-dd-hhmmss-{title-slug}{ext}
// If title is empty after slugification, falls back to URL hostname.
func GenerateFilename(title string, format string, timestamp time.Time, urlStr string) string {
	// Generate timestamp prefix (yyyy-mm-dd-hhmmss)
	timePrefix := timestamp.Format("2006-01-02-150405")

	// Slugify title
	titleSlug := SlugifyTitle(title, 80)

	// If title slug is empty, use URL hostname as fallback
	if titleSlug == "" {
		titleSlug = GenerateURLSlug(urlStr)
	}

	// Get file extension
	ext := GetFileExtension(format)

	// Combine: timestamp-titleslug.ext
	filename := fmt.Sprintf("%s-%s%s", timePrefix, titleSlug, ext)

	return filename
}

// ResolveConflict checks if filename exists in directory and appends counter if needed.
// Returns the final unique filename (not full path).
// Examples:
//   - file.md exists → file-1.md
//   - file-1.md exists → file-2.md
func ResolveConflict(dir, filename string) (string, error) {
	fullPath := filepath.Join(dir, filename)

	// If file doesn't exist, return original filename
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return filename, nil
	} else if err != nil {
		// Handle other errors (permission denied, etc.)
		return "", fmt.Errorf("failed to check file existence: %w", err)
	}

	// File exists, need to add counter
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	counter := 1
	for {
		newFilename := fmt.Sprintf("%s-%d%s", nameWithoutExt, counter, ext)
		newFullPath := filepath.Join(dir, newFilename)

		_, err := os.Stat(newFullPath)
		if os.IsNotExist(err) {
			return newFilename, nil
		} else if err != nil {
			// Handle other errors (permission denied, etc.)
			return "", fmt.Errorf("failed to check file existence: %w", err)
		}

		// File exists, try next counter
		// Add safety check to prevent infinite loop if filesystem is behaving oddly
		if counter > 10000 {
			return "", fmt.Errorf("too many conflicts for filename: %s", filename)
		}

		counter++
	}
}
