package applog

import (
	"fmt"
	"sync"
	"time"
)

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

type Logger struct {
	mu      sync.Mutex
	entries []LogEntry
	maxSize int
}

func New(maxSize int) *Logger {
	return &Logger{entries: make([]LogEntry, 0, maxSize), maxSize: maxSize}
}

func (l *Logger) add(level, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.entries) >= l.maxSize {
		l.entries = l.entries[1:]
	}
	l.entries = append(l.entries, LogEntry{Timestamp: time.Now(), Level: level, Message: msg})
}

func (l *Logger) Info(format string, args ...interface{})  { l.add("info", fmt.Sprintf(format, args...)) }
func (l *Logger) Warn(format string, args ...interface{})  { l.add("warn", fmt.Sprintf(format, args...)) }
func (l *Logger) Error(format string, args ...interface{}) { l.add("error", fmt.Sprintf(format, args...)) }

func (l *Logger) GetEntries() []LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	result := make([]LogEntry, len(l.entries))
	copy(result, l.entries)
	return result
}

func (l *Logger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = l.entries[:0]
}
