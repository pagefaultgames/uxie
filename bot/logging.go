package bot

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// SetLoggerDest configures [slog.Default] to output to the given file and stdout,
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

	slog.SetDefault(slog.New(slog.NewTextHandler(io.MultiWriter(outFile, os.Stdout), nil)))
	return nil
}

// debugAttrs is a wrapper around [slog.LogAttrs] for logging debug messages.
func debugAttrs(msg string, attrs ...slog.Attr) {
	slog.LogAttrs(nil, slog.LevelDebug, msg, attrs...)
}

// infoAttrs is a wrapper around [slog.LogAttrs] for logging informative messages.
func infoAttrs(msg string, attrs ...slog.Attr) {
	slog.LogAttrs(nil, slog.LevelInfo, msg, attrs...)
}
func warnAttrs(msg string, attrs ...slog.Attr) {
	slog.LogAttrs(nil, slog.LevelWarn, msg, attrs...)
}
func errorAttrs(msg string, attrs ...slog.Attr) {
	slog.LogAttrs(nil, slog.LevelError, msg, attrs...)
}

func init() {
	file, found := os.LookupEnv("LOG_FILE")
	if !found {
		file = "logs/oranguru.log"
	}
	if err := setLoggerDest(file); err != nil {
		panic("failed to set log destination: " + err.Error())
	}
}
