package logging

import (
	"log/slog"
	"os"
)

// Logger is a thin wrapper around slog.Logger to provide structured logging helpers.
type Logger struct {
	logger *slog.Logger
}

// New constructs a JSON slog logger writing to stdout.
func New() *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	return &Logger{logger: slog.New(handler)}
}

// With returns a child logger with additional attributes.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{logger: l.logger.With(args...)}
}

// Info logs an informational message.
func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// Debug logs a debug message when the level permits.
func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}
