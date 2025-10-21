// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"io"
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

func TestConvertToMarkdown_Headings(t *testing.T) {
	html := `<html><body>
		<h1>Heading 1</h1>
		<h2>Heading 2</h2>
		<h3>Heading 3</h3>
	</body></html>`

	converter := NewContentConverter(FormatMarkdown)
	md, err := converter.convertToMarkdown(html)
	if err != nil {
		t.Fatalf("convertToMarkdown failed: %v", err)
	}

	// Check for standard markdown heading format with space after #
	if !strings.Contains(md, "# Heading 1") {
		t.Errorf("expected h1 to convert to standard markdown heading '# Heading 1', got:\n%s", md)
	}
	if !strings.Contains(md, "## Heading 2") {
		t.Errorf("expected h2 to convert to standard markdown heading '## Heading 2', got:\n%s", md)
	}
	if !strings.Contains(md, "### Heading 3") {
		t.Errorf("expected h3 to convert to standard markdown heading '### Heading 3', got:\n%s", md)
	}
}

func TestConvertToMarkdown_Links(t *testing.T) {
	html := `<html><body>
		<a href="https://example.com">Example Link</a>
	</body></html>`

	converter := NewContentConverter(FormatMarkdown)
	md, err := converter.convertToMarkdown(html)
	if err != nil {
		t.Fatalf("convertToMarkdown failed: %v", err)
	}

	// Check for markdown link format [text](url) as a combined string
	if !strings.Contains(md, "[Example Link](https://example.com)") {
		t.Errorf("expected link to convert to markdown format [Example Link](https://example.com), got:\n%s", md)
	}
}

func TestConvertToMarkdown_Tables(t *testing.T) {
	html := `<html><body>
		<table>
			<thead>
				<tr><th>Name</th><th>Value</th></tr>
			</thead>
			<tbody>
				<tr><td>Item 1</td><td>100</td></tr>
			</tbody>
		</table>
	</body></html>`

	converter := NewContentConverter(FormatMarkdown)
	md, err := converter.convertToMarkdown(html)
	if err != nil {
		t.Fatalf("convertToMarkdown failed: %v", err)
	}

	// Verify proper markdown table syntax (not just content preservation)
	if !strings.Contains(md, "|") {
		t.Errorf("expected markdown table with pipe characters, got:\n%s", md)
	}

	// Verify header row
	if !strings.Contains(md, "Name") || !strings.Contains(md, "Value") {
		t.Errorf("expected table headers 'Name' and 'Value', got:\n%s", md)
	}

	// Verify data row
	if !strings.Contains(md, "Item 1") || !strings.Contains(md, "100") {
		t.Errorf("expected table data 'Item 1' and '100', got:\n%s", md)
	}

	// Verify separator row (some variation of dashes)
	if !strings.Contains(md, "---") {
		t.Errorf("expected table separator with dashes, got:\n%s", md)
	}
}

func TestConvertToMarkdown_Strikethrough(t *testing.T) {
	html := `<html><body>
		<p>This is <del>deleted</del> text.</p>
		<p>This is <s>strikethrough</s> text.</p>
		<p>This is <strike>old strikethrough</strike> text.</p>
	</body></html>`

	converter := NewContentConverter(FormatMarkdown)
	md, err := converter.convertToMarkdown(html)
	if err != nil {
		t.Fatalf("convertToMarkdown failed: %v", err)
	}

	// Verify strikethrough syntax
	if !strings.Contains(md, "~~") {
		t.Errorf("expected strikethrough with ~~ syntax, got:\n%s", md)
	}

	// Verify content is preserved for all three strikethrough tags
	if !strings.Contains(md, "deleted") {
		t.Errorf("expected 'deleted' content from <del> tag, got:\n%s", md)
	}
	if !strings.Contains(md, "strikethrough") {
		t.Errorf("expected 'strikethrough' content from <s> tag, got:\n%s", md)
	}
	if !strings.Contains(md, "old") {
		t.Errorf("expected 'old' content from <strike> tag, got:\n%s", md)
	}
}

func TestConvertToMarkdown_Lists(t *testing.T) {
	html := `<html><body>
		<ul>
			<li>Unordered 1</li>
			<li>Unordered 2</li>
		</ul>
		<ol>
			<li>Ordered 1</li>
			<li>Ordered 2</li>
		</ol>
	</body></html>`

	converter := NewContentConverter(FormatMarkdown)
	md, err := converter.convertToMarkdown(html)
	if err != nil {
		t.Fatalf("convertToMarkdown failed: %v", err)
	}

	// Check for unordered list syntax (*, -, or +) and content
	hasUnorderedMarker := strings.Contains(md, "* ") || strings.Contains(md, "- ") || strings.Contains(md, "+ ")
	if !hasUnorderedMarker {
		t.Errorf("expected unordered list markers (*, -, or +) in markdown, got:\n%s", md)
	}
	if !strings.Contains(md, "Unordered 1") || !strings.Contains(md, "Unordered 2") {
		t.Errorf("expected unordered list items in markdown, got:\n%s", md)
	}

	// Check for ordered list syntax (1., 2., etc.) and content
	hasOrderedMarker := strings.Contains(md, "1. ") || strings.Contains(md, "1) ")
	if !hasOrderedMarker {
		t.Errorf("expected ordered list markers (1., 2., etc.) in markdown, got:\n%s", md)
	}
	if !strings.Contains(md, "Ordered 1") || !strings.Contains(md, "Ordered 2") {
		t.Errorf("expected ordered list items in markdown, got:\n%s", md)
	}
}

func TestConvertToMarkdown_CodeBlocks(t *testing.T) {
	html := `<html><body>
		<pre><code>function hello() {
	console.log("Hello");
}</code></pre>
	</body></html>`

	converter := NewContentConverter(FormatMarkdown)
	md, err := converter.convertToMarkdown(html)
	if err != nil {
		t.Fatalf("convertToMarkdown failed: %v", err)
	}

	// Check for fenced code block markers
	if !strings.Contains(md, "```") {
		t.Errorf("expected code to be wrapped in fenced code block (```), got:\n%s", md)
	}

	// Check for code content
	if !strings.Contains(md, "hello") || !strings.Contains(md, "console.log") {
		t.Errorf("expected code block content in markdown, got:\n%s", md)
	}
}

func TestConvertToMarkdown_Minimal(t *testing.T) {
	html := `<html><body>Hello</body></html>`

	converter := NewContentConverter(FormatMarkdown)
	md, err := converter.convertToMarkdown(html)
	if err != nil {
		t.Fatalf("convertToMarkdown failed: %v", err)
	}

	// Check for basic content
	if !strings.Contains(md, "Hello") {
		t.Errorf("expected 'Hello' in markdown output, got:\n%s", md)
	}
}
