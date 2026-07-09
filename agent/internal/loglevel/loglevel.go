// Package loglevel provides level-based logging that works alongside
// the standard log package. Existing log.Printf calls remain as-is
// (treated as Info level). Use Debugf/Infof/Warnf/Errorf for leveled output.
package loglevel

import (
	"log"
	"strings"
	"sync/atomic"
)

// Log levels ordered by verbosity.
const (
	LevelError = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

var currentLevel int32 = LevelInfo

// SetLevel sets the minimum log level. Messages below this level are suppressed.
func SetLevel(level string) {
	var l int
	switch strings.ToLower(level) {
	case "debug":
		l = LevelDebug
	case "info", "":
		l = LevelInfo
	case "warn", "warning":
		l = LevelWarn
	case "error":
		l = LevelError
	default:
		l = LevelInfo
	}
	atomic.StoreInt32(&currentLevel, int32(l))
}

// GetLevel returns the current log level as a string.
func GetLevel() string {
	switch atomic.LoadInt32(&currentLevel) {
	case LevelDebug:
		return "debug"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "info"
	}
}

// Debugf logs a message at debug level (only when level is "debug").
func Debugf(format string, args ...interface{}) {
	if atomic.LoadInt32(&currentLevel) >= LevelDebug {
		log.Printf("[DBG] "+format, args...)
	}
}

// Infof logs a message at info level.
func Infof(format string, args ...interface{}) {
	if atomic.LoadInt32(&currentLevel) >= LevelInfo {
		log.Printf(format, args...)
	}
}

// Warnf logs a message at warn level.
func Warnf(format string, args ...interface{}) {
	if atomic.LoadInt32(&currentLevel) >= LevelWarn {
		log.Printf("[WARN] "+format, args...)
	}
}

// Errorf logs a message at error level.
func Errorf(format string, args ...interface{}) {
	if atomic.LoadInt32(&currentLevel) >= LevelError {
		log.Printf("[ERR] "+format, args...)
	}
}
