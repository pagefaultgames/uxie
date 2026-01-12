package utils

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// NB: These techincally have incorrect PC counters, but it hardly matters since we don't include said info in the logs anyhow.
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

// setLoggerDest configures [slog.Default] to output to the given file and stdout,
// overwriting any previously set outputs.
//
// The destination file will be created or truncated as necessary, alongside any required folders.
func setLoggerDest(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o0755); err != nil {
		return fmt.Errorf("error creating output directory %s: %w", dir, err)
	}

	outFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating output file %s: %w", path, err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(outFile, nil)))
	return nil
}

func init() {
	file, found := os.LookupEnv("LOG_FILE")
	if !found {
		file = "tmp/logs/oranguru.log"
	}
	if err := setLoggerDest(file); err != nil {
		panic("failed to set default log destination: " + err.Error())
	}
	if os.Getenv("VERBOSE") != "" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
}
