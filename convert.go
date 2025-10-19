// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"os"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/strikethrough"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
)

const (
	DefaultFileMode = 0644   // Owner RW, Group R, Other R
	BytesPerKB      = 1024.0 // Bytes in a kilobyte
)

// ContentConverter handles HTML to Markdown conversion and output
type ContentConverter struct {
	format string
}

// NewContentConverter creates a new content converter
func NewContentConverter(format string) *ContentConverter {
	return &ContentConverter{
		format: format,
	}
}

// Process processes the HTML content based on format and outputs it
func (cc *ContentConverter) Process(html string, outputFile string) error {
	var content string
	var err error

	switch cc.format {
	case FormatHTML:
		// Pass through HTML as-is
		content = html
		logger.Verbose("Output format: HTML (passthrough)")

	case FormatMarkdown:
		// Convert HTML to Markdown
		logger.Verbose("Converting HTML to Markdown...")
		content, err = cc.convertToMarkdown(html)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrConversionFailed, err)
		}
		logger.Debug("Converted to %d bytes of Markdown", len(content))

	default:
		return fmt.Errorf("unsupported format: %s", cc.format)
	}

	// Output content
	if outputFile != "" {
		return cc.writeToFile(content, outputFile)
	}

	return cc.writeToStdout(content)
}

// convertToMarkdown converts HTML to Markdown
func (cc *ContentConverter) convertToMarkdown(html string) (string, error) {
	// Create converter with table and strikethrough plugin support
	conv := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			table.NewTablePlugin(),
			strikethrough.NewStrikethroughPlugin(),
		),
	)

	markdown, err := conv.ConvertString(html)
	if err != nil {
		return "", err
	}

	logger.Success("Converted to Markdown")
	return markdown, nil
}

// writeToStdout writes content to stdout
func (cc *ContentConverter) writeToStdout(content string) error {
	logger.Verbose("Writing to stdout...")

	// Write to stdout
	_, err := fmt.Print(content)
	if err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}

	// Log success to stderr (so it doesn't mix with content)
	logger.Debug("Wrote %d bytes to stdout", len(content))

	return nil
}

// writeToFile writes content to a file
func (cc *ContentConverter) writeToFile(content string, filename string) error {
	logger.Verbose("Writing to file: %s", filename)

	// Check if file exists and warn in verbose mode
	if _, err := os.Stat(filename); err == nil {
		logger.Verbose("Overwriting existing file: %s", filename)
	}

	// Write to file
	err := os.WriteFile(filename, []byte(content), DefaultFileMode)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}

	// Calculate size in KB
	sizeKB := float64(len(content)) / BytesPerKB
	logger.Success("Saved to %s (%.1f KB)", filename, sizeKB)

	return nil
}
