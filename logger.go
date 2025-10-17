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
)

// LogLevel represents the logging verbosity level
type LogLevel int

const (
	// LevelQuiet suppresses all output except fatal errors
	LevelQuiet LogLevel = iota
	// LevelNormal shows key operations with emoji indicators
	LevelNormal
	// LevelVerbose shows detailed operation logs
	LevelVerbose
	// LevelDebug shows everything including CDP messages and timing
	LevelDebug
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

// Logger handles formatted output to stderr
type Logger struct {
	level  LogLevel
	color  bool
	writer io.Writer
}

// NewLogger creates a new logger instance
func NewLogger(level LogLevel) *Logger {
	color := shouldUseColor()
	return &Logger{
		level:  level,
		color:  color,
		writer: os.Stderr,
	}
}

// shouldUseColor determines if color output should be used
func shouldUseColor() bool {
	// Respect NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if stderr is a terminal
	if fileInfo, err := os.Stderr.Stat(); err == nil {
		return (fileInfo.Mode() & os.ModeCharDevice) != 0
	}

	return false
}

// Success logs a success message with green checkmark
func (l *Logger) Success(format string, args ...interface{}) {
	if l.level >= LevelNormal {
		msg := fmt.Sprintf(format, args...)
		prefix := "✓"
		if l.color {
			prefix = colorGreen + "✓" + colorReset
		}
		fmt.Fprintf(l.writer, "%s %s\n", prefix, msg)
	}
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level >= LevelNormal {
		msg := fmt.Sprintf(format, args...)
		if l.color {
			msg = colorCyan + msg + colorReset
		}
		fmt.Fprintf(l.writer, "%s\n", msg)
	}
}

// Verbose logs a detailed message (only in verbose/debug mode)
func (l *Logger) Verbose(format string, args ...interface{}) {
	if l.level >= LevelVerbose {
		msg := fmt.Sprintf(format, args...)
		if l.color {
			msg = colorCyan + msg + colorReset
		}
		fmt.Fprintf(l.writer, "%s\n", msg)
	}
}

// Debug logs a debug message (only in debug mode)
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level >= LevelDebug {
		msg := fmt.Sprintf(format, args...)
		fmt.Fprintf(l.writer, "[DEBUG] %s\n", msg)
	}
}

// Warning logs a warning message with yellow indicator
func (l *Logger) Warning(format string, args ...interface{}) {
	if l.level >= LevelNormal {
		msg := fmt.Sprintf(format, args...)
		prefix := "⚠"
		if l.color {
			prefix = colorYellow + "⚠" + colorReset
		}
		fmt.Fprintf(l.writer, "%s %s\n", prefix, msg)
	}
}

// Error logs an error message with red X (always shown, even in quiet mode)
func (l *Logger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	prefix := "✗"
	if l.color {
		prefix = colorRed + "✗" + colorReset
	}
	fmt.Fprintf(l.writer, "%s %s\n", prefix, msg)
}

// ErrorWithSuggestion logs an error with a helpful suggestion
func (l *Logger) ErrorWithSuggestion(errMsg string, suggestion string) {
	prefix := "✗"
	if l.color {
		prefix = colorRed + "✗" + colorReset
		suggestion = colorCyan + "  Try: " + suggestion + colorReset
	} else {
		suggestion = "  Try: " + suggestion
	}
	fmt.Fprintf(l.writer, "%s %s\n%s\n", prefix, errMsg, suggestion)
}

// Progress logs a progress message (operation in progress)
func (l *Logger) Progress(format string, args ...interface{}) {
	if l.level >= LevelNormal {
		msg := fmt.Sprintf(format, args...)
		fmt.Fprintf(l.writer, "%s\n", msg)
	}
}
