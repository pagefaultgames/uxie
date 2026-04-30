// SPDX-FileCopyrightText: 2026 Pagefault Games
// SPDX-FileContributor: Bertie690
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
)

// NB: These technically have incorrect PC counters, but it hardly matters since we don't include said info in the logs anyhow.
// (it takes up too much space in practice)

// DebugAttrs is a wrapper around [slog.LogAttrs] for logging debug messages.
func DebugAttrs(msg string, attrs ...slog.Attr) {
	slog.LogAttrs(context.Background(), slog.LevelDebug, msg, attrs...)
}

// InfoAttrs is a wrapper around [slog.LogAttrs] for logging informative messages.
func InfoAttrs(msg string, attrs ...slog.Attr) {
	slog.LogAttrs(context.Background(), slog.LevelInfo, msg, attrs...)
}

// WarnAttrs is a wrapper around [slog.LogAttrs] for logging warning messages.
func WarnAttrs(msg string, attrs ...slog.Attr) {
	slog.LogAttrs(context.Background(), slog.LevelWarn, msg, attrs...)
}

// ErrorAttrs is a wrapper around [slog.LogAttrs] for logging error messages.
func ErrorAttrs(msg string, attrs ...slog.Attr) {
	slog.LogAttrs(context.Background(), slog.LevelError, msg, attrs...)
}

func init() {
	if os.Getenv("VERBOSE") != "" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	path, found := os.LookupEnv("LOG_FILE")
	if !found {
		path = "./tmp/logs/uxie.log"
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o0755); err != nil {
		panic(fmt.Sprintf("error creating logfile output directory %q: %v", dir, err))
	}

	outFile, err := os.Create(path)
	if err != nil {
		panic(fmt.Sprintf("error creating logfile output file %q: %v", path, err))
	}

	// NB: slog by default forwards to the standard library's log package
	log.SetPrefix("[UXIE] ")
	log.SetOutput(io.MultiWriter(outFile, os.Stdout))
}
