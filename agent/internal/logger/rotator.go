// Package logger provides a size-based rotating file writer for agent logs.
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

const (
	defaultMaxSize  = 10 * 1024 * 1024 // 10 MB
	defaultBackups  = 3
)

// RotatingFileWriter is an io.Writer that writes to a file and rotates
// when the file exceeds MaxSize. Old log files are named <base>.log.1,
// <base>.log.2, etc., with at most MaxBackups retained.
type RotatingFileWriter struct {
	mu        sync.Mutex
	path      string
	maxSize   int64
	maxBackups int
	file      *os.File
	size      int64
}

// NewRotatingFileWriter creates a RotatingFileWriter.
// If maxSize <= 0, 10MB is used. If maxBackups <= 0, 3 backups are kept.
func NewRotatingFileWriter(path string, maxSize int64, maxBackups int) (*RotatingFileWriter, error) {
	if maxSize <= 0 {
		maxSize = defaultMaxSize
	}
	if maxBackups <= 0 {
		maxBackups = defaultBackups
	}

	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create log directory %s: %w", dir, err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file %s: %w", path, err)
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("stat log file %s: %w", path, err)
	}

	w := &RotatingFileWriter{
		path:        path,
		maxSize:     maxSize,
		maxBackups:  maxBackups,
		file:        f,
		size:        info.Size(),
	}

	return w, nil
}

// Write implements io.Writer.
func (w *RotatingFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	n, err := w.file.Write(p)
	if err != nil {
		return n, err
	}

	w.size += int64(n)
	if w.size >= w.maxSize {
		if err := w.rotate(); err != nil {
			fmt.Fprintf(os.Stderr, "log rotation failed: %v\n", err)
		}
	}

	return n, nil
}

// Close closes the underlying log file.
func (w *RotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// rotate performs the actual rotation: renames the current file to .1,
// shifts existing backups up by one, then opens a fresh file.
func (w *RotatingFileWriter) rotate() error {
	if err := w.file.Close(); err != nil {
		return fmt.Errorf("close current log: %w", err)
	}

	base := w.path

	// Remove the oldest backup if it exists
	oldest := fmt.Sprintf("%s.%d", base, w.maxBackups)
	os.Remove(oldest)

	// Shift backups: .N -> .N+1
	for i := w.maxBackups - 1; i >= 1; i-- {
		old := fmt.Sprintf("%s.%d", base, i)
		new := fmt.Sprintf("%s.%d", base, i+1)
		os.Rename(old, new)
	}

	// Rename current to .1
	if err := os.Rename(base, fmt.Sprintf("%s.%d", base, 1)); err != nil {
		return fmt.Errorf("rename log for rotation: %w", err)
	}

	// Open fresh file
	f, err := os.OpenFile(base, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("create new log after rotation: %w", err)
	}

	w.file = f
	w.size = 0

	// Clean up any excess backups (in case maxBackups was reduced)
	w.cleanup()

	return nil
}

// cleanup removes excess backup files beyond maxBackups.
func (w *RotatingFileWriter) cleanup() {
	base := w.path
	pattern := fmt.Sprintf("%s.*", base)

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	// Sort ascending (natural order: .1, .2, .3, ...)
	sort.Strings(matches)

	// Keep the most recent maxBackups, remove the rest
	if len(matches) > w.maxBackups {
		for _, m := range matches[:len(matches)-w.maxBackups] {
			os.Remove(m)
		}
	}
}