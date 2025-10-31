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
	slugNonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)
	slugMultipleHyphens = regexp.MustCompile(`-+`)
)

func SlugifyTitle(title string, maxLen int) string {
	slug := strings.ToLower(title)

	slug = slugNonAlphanumeric.ReplaceAllString(slug, "-")

	slug = slugMultipleHyphens.ReplaceAllString(slug, "-")

	slug = strings.Trim(slug, "-")

	if len(slug) > maxLen {
		slug = slug[:maxLen]
		slug = strings.TrimRight(slug, "-")
	}

	return slug
}

func GenerateURLSlug(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "page"
	}

	hostname := parsedURL.Host
	if hostname == "" {
		return "page"
	}

	return SlugifyTitle(hostname, MaxSlugLength)
}

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
		return ".md"
	}
}

func GenerateFilename(title string, format string, timestamp time.Time, urlStr string) string {
	timePrefix := timestamp.Format("2006-01-02-150405")

	titleSlug := SlugifyTitle(title, MaxSlugLength)
	logger.Debug("Title '%s' slugified to '%s'", title, titleSlug)

	if titleSlug == "" {
		titleSlug = GenerateURLSlug(urlStr)
		logger.Debug("Empty title slug, using URL slug: %s", titleSlug)
	}

	ext := GetFileExtension(format)

	filename := fmt.Sprintf("%s-%s%s", timePrefix, titleSlug, ext)
	logger.Debug("Generated filename: %s", filename)

	return filename
}

func ResolveConflict(dir, filename string) (string, error) {
	fullPath := filepath.Join(dir, filename)
	logger.Debug("Checking for conflicts: %s", fullPath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		logger.Debug("No conflict, using original filename")
		return filename, nil
	} else if err != nil {
		return "", fmt.Errorf("failed to check file existence: %w", err)
	}

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
			return "", fmt.Errorf("failed to check file existence: %w", err)
		}

		if counter > 10000 {
			return "", fmt.Errorf("too many conflicts for filename: %s", filename)
		}

		counter++
	}
}
