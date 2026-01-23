package aura

import (
	"log/slog"
	"os"
)

// testLogger creates a logger for testing that outputs to stderr
func testLogger() *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelWarn}
	handler := slog.NewTextHandler(os.Stderr, opts)
	return slog.New(handler)
}
