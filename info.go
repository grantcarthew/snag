// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/spf13/cobra"
)

// PageInfo represents metadata about a web page for JSON output.
type PageInfo struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Domain    string `json:"domain"`
	Slug      string `json:"slug"`
	Timestamp string `json:"timestamp"`
}

// ExtractPageInfo extracts metadata from a rod.Page and returns a PageInfo struct.
func ExtractPageInfo(page *rod.Page) (*PageInfo, error) {
	if page == nil {
		return nil, fmt.Errorf("cannot extract info: page is nil")
	}

	pageInfo, err := page.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get page info: %w", err)
	}

	domain := extractDomain(pageInfo.URL)
	slug := SlugifyTitle(pageInfo.Title, MaxSlugLength)

	return &PageInfo{
		Title:     pageInfo.Title,
		URL:       pageInfo.URL,
		Domain:    domain,
		Slug:      slug,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// extractDomain extracts the domain from a URL string.
func extractDomain(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	host := parsedURL.Host

	// Remove port if present
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// Remove www. prefix if present
	host = strings.TrimPrefix(host, "www.")

	return host
}

// OutputPageInfo writes the PageInfo as JSON to the specified output (stdout or file).
func OutputPageInfo(info *PageInfo, outputFile string) error {
	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal page info to JSON: %w", err)
	}

	if outputFile == "" {
		fmt.Println(string(jsonData))
		return nil
	}

	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write info to file: %w", err)
	}

	logger.Success("Saved info to %s", outputFile)
	return nil
}

// handleInfoFromURL fetches page info from a URL and outputs as JSON.
func handleInfoFromURL(cmd *cobra.Command, urlStr string) error {
	validatedURL, err := validateURL(urlStr)
	if err != nil {
		return err
	}

	if err := validateTimeout(timeout); err != nil {
		return err
	}

	if err := validatePort(port); err != nil {
		return err
	}

	outputFile := strings.TrimSpace(output)
	if cmd.Flags().Changed("output") && outputFile != "" {
		if err := validateOutputPath(outputFile); err != nil {
			return err
		}
	}

	validatedUserDataDir := ""
	if cmd.Flags().Changed("user-data-dir") {
		validatedDir, err := validateUserDataDir(userDataDir)
		if err != nil {
			return err
		}
		validatedUserDataDir = validatedDir
	}

	validatedWaitFor := validateWaitFor(waitFor, cmd.Flags().Changed("wait-for"))

	bm := NewBrowserManager(BrowserOptions{
		Port:          port,
		ForceHeadless: forceHead,
		UserDataDir:   validatedUserDataDir,
	})

	browserMutex.Lock()
	browserManager = bm
	browserMutex.Unlock()

	defer func() {
		bm.Close()
		browserMutex.Lock()
		browserManager = nil
		browserMutex.Unlock()
	}()

	_, err = bm.Connect()
	if err != nil {
		return err
	}

	page, err := bm.NewPage()
	if err != nil {
		return err
	}

	if closeTab || bm.launchedHeadless {
		defer bm.ClosePage(page)
	}

	fetcher := NewPageFetcher(page, timeout)
	_, err = fetcher.Fetch(FetchOptions{
		URL:     validatedURL,
		Timeout: timeout,
		WaitFor: validatedWaitFor,
	})
	if err != nil {
		return err
	}

	pageInfo, err := ExtractPageInfo(page)
	if err != nil {
		return err
	}

	return OutputPageInfo(pageInfo, outputFile)
}

// handleInfoFromTab fetches page info from an existing tab and outputs as JSON.
func handleInfoFromTab(cmd *cobra.Command) error {
	tabValue := strings.TrimSpace(tab)
	if tabValue == "" {
		logger.Error("Tab pattern cannot be empty")
		return fmt.Errorf("tab pattern cannot be empty")
	}

	if cmd.Flags().Changed("user-agent") {
		logger.Warning("--user-agent is ignored with --tab (cannot change existing tab's user agent)")
	}
	if cmd.Flags().Changed("user-data-dir") {
		logger.Warning("--user-data-dir ignored when connecting to existing browser")
	}

	outputFile := strings.TrimSpace(output)
	if cmd.Flags().Changed("output") && outputFile != "" {
		if err := validateOutputPath(outputFile); err != nil {
			return err
		}
	}

	bm, err := connectToExistingBrowser(port)
	if err != nil {
		return err
	}
	defer func() {
		browserMutex.Lock()
		browserManager = nil
		browserMutex.Unlock()
	}()

	var page *rod.Page

	// Try parsing as tab index
	if tabIndex, err := strconv.Atoi(tabValue); err == nil {
		logger.Verbose("Getting info from tab index: %d", tabIndex)
		page, err = bm.GetTabByIndex(tabIndex)
		if err != nil {
			if errors.Is(err, ErrTabIndexInvalid) {
				logger.Error("Tab index out of range")
				logger.Info("Run 'snag --list-tabs' to see available tabs")
			}
			return err
		}
	} else {
		// Pattern matching - but --info only supports single tab
		logger.Verbose("Getting info from tab matching pattern: %s", tabValue)
		matchedPages, err := bm.GetTabsByPattern(tabValue)
		if err != nil {
			if errors.Is(err, ErrNoTabMatch) {
				logger.Error("No tab matches pattern '%s'", tabValue)
				logger.Info("Run 'snag --list-tabs' to see available tabs")
			}
			return err
		}

		if len(matchedPages) > 1 {
			logger.Error("Pattern '%s' matched %d tabs, --info requires exactly one", tabValue, len(matchedPages))
			logger.Info("Use a more specific pattern or tab index")
			return fmt.Errorf("pattern matched multiple tabs")
		}

		page = matchedPages[0]
	}

	// Wait for selector if specified
	if cmd.Flags().Changed("wait-for") {
		validatedWaitFor := validateWaitFor(waitFor, true)
		if validatedWaitFor != "" {
			err := waitForSelector(page, validatedWaitFor, time.Duration(timeout)*time.Second)
			if err != nil {
				return err
			}
		}
	}

	pageInfo, err := ExtractPageInfo(page)
	if err != nil {
		return err
	}

	return OutputPageInfo(pageInfo, outputFile)
}
