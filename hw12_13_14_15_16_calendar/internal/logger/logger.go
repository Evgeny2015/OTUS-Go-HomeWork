package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// LogLevel represents the severity of a log message.
type LogLevel int

const (
	LevelError LogLevel = iota
	LevelWarning
	LevelInfo
	LevelDebug
)

var levelNames = map[LogLevel]string{
	LevelError:   "ERROR",
	LevelWarning: "WARN",
	LevelInfo:    "INFO",
	LevelDebug:   "DEBUG",
}

var levelFromString = map[string]LogLevel{
	"error":   LevelError,
	"warn":    LevelWarning,
	"warning": LevelWarning,
	"info":    LevelInfo,
	"debug":   LevelDebug,
}

// Config holds configuration for the logger.
type Config struct {
	Level  string // "debug", "info", "warning", "error"
	Output string // file path, empty for stdout
	Format string // "text" or "json" (currently only text supported)
}

// Logger is a simple logger with level filtering.
type Logger struct {
	level  LogLevel
	output io.Writer
	format string
}

// New creates a new Logger instance with the specified level.
// If level is empty or invalid, defaults to LevelInfo.
// Output is set to stdout.
func New(level string) *Logger {
	lvl, ok := levelFromString[strings.ToLower(level)]
	if !ok {
		lvl = LevelInfo // default level
	}
	return &Logger{
		level:  lvl,
		output: os.Stdout,
		format: "text",
	}
}

// NewFromConfig creates a Logger from the provided configuration.
func NewFromConfig(cfg Config) *Logger {
	lvl, ok := levelFromString[strings.ToLower(cfg.Level)]
	if !ok {
		lvl = LevelInfo
	}
	var output io.Writer = os.Stdout
	if cfg.Output != "" {
		file, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0x666)
		if err != nil {
			// fallback to stdout, but we cannot log because logger not ready
			// just ignore and keep stdout
		} else {
			output = file
		}
	}
	format := cfg.Format
	if format == "" {
		format = "text"
	}
	return &Logger{
		level:  lvl,
		output: output,
		format: format,
	}
}

// logEntry represents a structured log entry for JSON output.
type logEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

// log writes a message with the given level if it's enabled.
func (l *Logger) log(level LogLevel, msg string) {
	if level > l.level {
		return // level is higher than configured, skip
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelName := levelNames[level]

	if l.format == "json" {
		entry := logEntry{
			Timestamp: timestamp,
			Level:     levelName,
			Message:   msg,
		}
		data, err := json.Marshal(&entry)
		if err != nil {
			// fallback to text format if JSON marshaling fails
			fmt.Fprintf(l.output, "%s [%s] %s\n", timestamp, levelName, msg)
			return
		}
		fmt.Fprintln(l.output, string(data))
	} else {
		fmt.Fprintf(l.output, "%s [%s] %s\n", timestamp, levelName, msg)
	}
}

// Info logs an informational message.
func (l *Logger) Info(msg string) {
	l.log(LevelInfo, msg)
}

// Error logs an error message.
func (l *Logger) Error(msg string) {
	l.log(LevelError, msg)
}

// Warning logs a warning message.
func (l *Logger) Warning(msg string) {
	l.log(LevelWarning, msg)
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string) {
	l.log(LevelDebug, msg)
}
