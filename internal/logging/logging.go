package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Config drives how the logger writes structured events.
type Config struct {
	Dir string
}

// Logger wraps slog.Logger and tracks resources that require cleanup.
type Logger struct {
	closers []io.Closer
	logger  *slog.Logger
}

// New constructs a JSON slog logger that writes to stdout and a rotating file.
func New(cfg Config) (*Logger, error) {
	if cfg.Dir == "" {
		cfg.Dir = "logs"
	}
	rotating, err := newRotatingWriter(cfg.Dir)
	if err != nil {
		return nil, err
	}

	handler := slog.NewJSONHandler(io.MultiWriter(os.Stdout, rotating), &slog.HandlerOptions{Level: slog.LevelInfo})
	return &Logger{logger: slog.New(handler), closers: []io.Closer{rotating}}, nil
}

// Close releases any underlying writers.
func (l *Logger) Close() error {
	var firstErr error
	for _, c := range l.closers {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// With returns a child logger with additional attributes.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{logger: l.logger.With(args...), closers: l.closers}
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

// rotatingWriter rotates log files based on the current timestamp (minute resolution).
type rotatingWriter struct {
	dir         string
	currentName string
	file        *os.File
	mu          sync.Mutex
}

func newRotatingWriter(dir string) (*rotatingWriter, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}
	return &rotatingWriter{dir: dir}, nil
}

func (w *rotatingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	filename := time.Now().Format("2006-01-02-15:04") + ".log"
	if w.file == nil || w.currentName != filename {
		if err := w.rotate(filename); err != nil {
			return 0, err
		}
	}
	return w.file.Write(p)
}

func (w *rotatingWriter) rotate(filename string) error {
	if w.file != nil {
		_ = w.file.Close()
	}
	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return fmt.Errorf("ensure log dir: %w", err)
	}
	path := filepath.Join(w.dir, filename)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o640)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	w.file = file
	w.currentName = filename
	return nil
}

func (w *rotatingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}
