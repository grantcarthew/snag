// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// newTestLogger creates a logger for testing that writes to a buffer
func newTestLogger(level LogLevel, writer io.Writer) *Logger {
	return &Logger{
		level:  level,
		color:  false, // Disable color for testing
		writer: writer,
	}
}

func TestLogger_Success(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(LevelNormal, &buf)

	logger.Success("Operation completed")

	output := buf.String()
	if !strings.Contains(output, "Operation completed") {
		t.Errorf("expected success message in output, got: %s", output)
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(LevelNormal, &buf)

	logger.Error("Something went wrong")

	output := buf.String()
	if !strings.Contains(output, "Something went wrong") {
		t.Errorf("expected error message in output, got: %s", output)
	}
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(LevelNormal, &buf)

	logger.Info("Informational message")

	output := buf.String()
	if !strings.Contains(output, "Informational message") {
		t.Errorf("expected info message in output, got: %s", output)
	}
}

func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(LevelDebug, &buf)

	logger.Debug("Debug information")

	output := buf.String()
	if !strings.Contains(output, "Debug information") {
		t.Errorf("expected debug message in output, got: %s", output)
	}
}

func TestLogger_QuietMode(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(LevelQuiet, &buf)

	// These should be suppressed in quiet mode
	logger.Success("Success message")
	logger.Info("Info message")
	logger.Debug("Debug message")

	output := buf.String()
	if strings.Contains(output, "Success message") {
		t.Errorf("quiet mode should suppress success messages, got: %s", output)
	}
	if strings.Contains(output, "Info message") {
		t.Errorf("quiet mode should suppress info messages, got: %s", output)
	}

	// Errors should still appear in quiet mode
	logger.Error("Error message")
	output = buf.String()
	if !strings.Contains(output, "Error message") {
		t.Errorf("quiet mode should still show error messages, got: %s", output)
	}
}

func TestLogger_VerboseMode(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(LevelVerbose, &buf)

	logger.Success("Success message")
	logger.Info("Info message")

	output := buf.String()
	if !strings.Contains(output, "Success message") {
		t.Errorf("verbose mode should show success messages, got: %s", output)
	}
	if !strings.Contains(output, "Info message") {
		t.Errorf("verbose mode should show info messages, got: %s", output)
	}

	// Debug messages should NOT appear in verbose mode (only in debug mode)
	buf.Reset()
	logger.Debug("Debug message")
	output = buf.String()
	if strings.Contains(output, "Debug message") {
		t.Errorf("verbose mode should not show debug messages, got: %s", output)
	}
}

func TestLogger_StderrOnly(t *testing.T) {
	// This test verifies that logger writes to the provided writer (stderr in practice)
	// and not to stdout
	var stderr bytes.Buffer
	logger := newTestLogger(LevelNormal, &stderr)

	logger.Info("This should go to stderr")

	if stderr.Len() == 0 {
		t.Error("expected logger to write to stderr buffer")
	}

	output := stderr.String()
	if !strings.Contains(output, "This should go to stderr") {
		t.Errorf("expected message in stderr output, got: %s", output)
	}
}
