// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/strikethrough"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/k3a/html2text"
)

const (
	DefaultFileMode = 0644   // Owner RW, Group R, Other R
	BytesPerKB      = 1024.0 // Bytes in a kilobyte
)

// markdownConverter is a reusable HTML-to-Markdown converter.
// Created once at package initialization with all necessary plugins.
// The converter is stateless and safe for concurrent use.
var markdownConverter = converter.NewConverter(
	converter.WithPlugins(
		base.NewBasePlugin(),
		commonmark.NewCommonmarkPlugin(),
		table.NewTablePlugin(),
		strikethrough.NewStrikethroughPlugin(),
	),
)

// ContentConverter handles content format conversion and output
type ContentConverter struct {
	format string
}

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
		content = html
		logger.Verbose("Output format: HTML (passthrough)")

	case FormatMarkdown:
		logger.Verbose("Converting HTML to Markdown...")
		content, err = cc.convertToMarkdown(html)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrConversionFailed, err)
		}
		logger.Debug("Converted to %d bytes of Markdown", len(content))

	case FormatText:
		logger.Verbose("Extracting plain text...")
		content = cc.extractPlainText(html)
		logger.Debug("Extracted %d bytes of plain text", len(content))

	default:
		return fmt.Errorf("unsupported format: %s", cc.format)
	}

	if outputFile != "" {
		return cc.writeToFile(content, outputFile)
	}

	return cc.writeToStdout(content)
}

func (cc *ContentConverter) convertToMarkdown(html string) (string, error) {
	// Reuse package-level converter (stateless, thread-safe)
	markdown, err := markdownConverter.ConvertString(html)
	if err != nil {
		return "", err
	}

	return markdown, nil
}

func (cc *ContentConverter) extractPlainText(htmlContent string) string {
	text := html2text.HTML2TextWithOptions(
		htmlContent,
		html2text.WithUnixLineBreaks(),
	)

	return text
}

func (cc *ContentConverter) writeToStdout(content string) error {
	logger.Verbose("Writing to stdout...")

	_, err := fmt.Print(content)
	if err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}

	// Log success to stderr (so it doesn't mix with content)
	logger.Debug("Wrote %d bytes to stdout", len(content))

	return nil
}

func (cc *ContentConverter) writeToFile(content string, filename string) error {
	logger.Verbose("Writing to file: %s", filename)

	if _, err := os.Stat(filename); err == nil {
		logger.Verbose("Overwriting existing file: %s", filename)
	}

	err := os.WriteFile(filename, []byte(content), DefaultFileMode)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}

	sizeKB := float64(len(content)) / BytesPerKB
	logger.Success("Saved to %s (%.1f KB)", filename, sizeKB)

	return nil
}

// ProcessPage processes content from a Rod page for binary formats (PDF, PNG)
func (cc *ContentConverter) ProcessPage(page *rod.Page, outputFile string) error {
	var data []byte
	var err error

	switch cc.format {
	case FormatPDF:
		logger.Verbose("Generating PDF...")
		data, err = cc.generatePDF(page)
		if err != nil {
			return fmt.Errorf("failed to generate PDF: %w", err)
		}
		logger.Debug("Generated %d bytes of PDF", len(data))

	case FormatPNG:
		logger.Verbose("Capturing PNG screenshot...")
		data, err = cc.captureScreenshot(page)
		if err != nil {
			return fmt.Errorf("failed to capture PNG screenshot: %w", err)
		}
		logger.Debug("Captured %d bytes of PNG", len(data))

	default:
		return fmt.Errorf("unsupported binary format: %s", cc.format)
	}

	if outputFile != "" {
		return cc.writeBinaryToFile(data, outputFile)
	}

	return cc.writeBinaryToStdout(data)
}

func (cc *ContentConverter) generatePDF(page *rod.Page) ([]byte, error) {
	// Use Chrome's print-to-PDF with default settings (locale-aware paper size)
	stream, err := page.PDF(&proto.PagePrintToPDF{
		PrintBackground: true, // Include background graphics
	})
	if err != nil {
		return nil, fmt.Errorf("PDF generation failed: %w", err)
	}

	pdfData, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF data: %w", err)
	}

	return pdfData, nil
}

func (cc *ContentConverter) captureScreenshot(page *rod.Page) ([]byte, error) {
	screenshotData, err := page.Screenshot(true, &proto.PageCaptureScreenshot{
		Format: proto.PageCaptureScreenshotFormatPng,
	})
	if err != nil {
		return nil, fmt.Errorf("screenshot capture failed: %w", err)
	}

	return screenshotData, nil
}

func (cc *ContentConverter) writeBinaryToStdout(data []byte) error {
	logger.Verbose("Writing binary data to stdout...")

	_, err := os.Stdout.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}

	// Log success to stderr (so it doesn't mix with content)
	logger.Debug("Wrote %d bytes to stdout", len(data))

	return nil
}

func (cc *ContentConverter) writeBinaryToFile(data []byte, filename string) error {
	logger.Verbose("Writing binary data to file: %s", filename)

	if _, err := os.Stat(filename); err == nil {
		logger.Verbose("Overwriting existing file: %s", filename)
	}

	err := os.WriteFile(filename, data, DefaultFileMode)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}

	sizeKB := float64(len(data)) / BytesPerKB
	logger.Success("Saved to %s (%.1f KB)", filename, sizeKB)

	return nil
}
