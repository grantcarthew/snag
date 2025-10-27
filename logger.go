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

type LogLevel int

const (
	LevelQuiet LogLevel = iota
	LevelNormal
	LevelVerbose
	LevelDebug
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

type Logger struct {
	level  LogLevel
	color  bool
	writer io.Writer
}

func NewLogger(level LogLevel) *Logger {
	color := shouldUseColor()
	return &Logger{
		level:  level,
		color:  color,
		writer: os.Stderr,
	}
}

func shouldUseColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	if fileInfo, err := os.Stderr.Stat(); err == nil {
		return (fileInfo.Mode() & os.ModeCharDevice) != 0
	}

	return false
}

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

func (l *Logger) Info(format string, args ...interface{}) {
	if l.level >= LevelNormal {
		msg := fmt.Sprintf(format, args...)
		fmt.Fprintf(l.writer, "%s\n", msg)
	}
}

func (l *Logger) Verbose(format string, args ...interface{}) {
	if l.level >= LevelVerbose {
		msg := fmt.Sprintf(format, args...)
		if l.color {
			msg = colorCyan + msg + colorReset
		}
		fmt.Fprintf(l.writer, "%s\n", msg)
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level >= LevelDebug {
		msg := fmt.Sprintf(format, args...)
		fmt.Fprintf(l.writer, "[DEBUG] %s\n", msg)
	}
}

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

func (l *Logger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	prefix := "✗"
	if l.color {
		prefix = colorRed + "✗" + colorReset
	}
	fmt.Fprintf(l.writer, "%s %s\n", prefix, msg)
}

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

