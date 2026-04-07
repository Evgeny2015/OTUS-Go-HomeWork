package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestLogger_LevelFiltering(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		messages []struct {
			method func(*Logger, string)
			text   string
			want   bool // whether it should appear in output
		}
	}{
		{
			name:  "debug level shows all messages",
			level: "debug",
			messages: []struct {
				method func(*Logger, string)
				text   string
				want   bool
			}{
				{(*Logger).Debug, "debug message", true},
				{(*Logger).Info, "info message", true},
				{(*Logger).Warning, "warning message", true},
				{(*Logger).Error, "error message", true},
			},
		},
		{
			name:  "info level filters debug",
			level: "info",
			messages: []struct {
				method func(*Logger, string)
				text   string
				want   bool
			}{
				{(*Logger).Debug, "debug message", false},
				{(*Logger).Info, "info message", true},
				{(*Logger).Warning, "warning message", true},
				{(*Logger).Error, "error message", true},
			},
		},
		{
			name:  "warning level filters debug and info",
			level: "warning",
			messages: []struct {
				method func(*Logger, string)
				text   string
				want   bool
			}{
				{(*Logger).Debug, "debug message", false},
				{(*Logger).Info, "info message", false},
				{(*Logger).Warning, "warning message", true},
				{(*Logger).Error, "error message", true},
			},
		},
		{
			name:  "error level filters all but error",
			level: "error",
			messages: []struct {
				method func(*Logger, string)
				text   string
				want   bool
			}{
				{(*Logger).Debug, "debug message", false},
				{(*Logger).Info, "info message", false},
				{(*Logger).Warning, "warning message", false},
				{(*Logger).Error, "error message", true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(tt.level)
			// Replace output for testing
			logger.output = &buf

			for _, msg := range tt.messages {
				msg.method(logger, msg.text)
			}

			output := buf.String()
			for _, msg := range tt.messages {
				hasMessage := strings.Contains(output, msg.text)
				if hasMessage != msg.want {
					t.Errorf("message %q: got output=%v, want %v", msg.text, hasMessage, msg.want)
				}
			}
		})
	}
}

func TestLogger_DefaultLevel(t *testing.T) {
	// Test that invalid level defaults to info
	logger := New("invalid")
	var buf bytes.Buffer
	logger.output = &buf

	logger.Debug("debug message")
	logger.Info("info message")

	output := buf.String()
	if strings.Contains(output, "debug message") {
		t.Error("debug message should be filtered with default info level")
	}
	if !strings.Contains(output, "info message") {
		t.Error("info message should appear with default info level")
	}
}

func TestLogger_OutputFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := New("info")
	logger.output = &buf

	logger.Info("test message")

	output := buf.String()
	// Check format: timestamp [LEVEL] message
	if !strings.Contains(output, "[INFO]") {
		t.Error("output should contain level prefix")
	}
	if !strings.Contains(output, "test message") {
		t.Error("output should contain the message")
	}
	// Check timestamp format (should have YYYY-MM-DD HH:MM:SS)
	// We'll just check it has numbers and dashes/colons
	if !strings.Contains(output, "-") || !strings.Contains(output, ":") {
		t.Error("output should contain timestamp with date and time")
	}
}

func TestLogger_EmptyLevel(t *testing.T) {
	// Empty string should default to info
	logger := New("")
	var buf bytes.Buffer
	logger.output = &buf

	logger.Debug("debug")
	logger.Info("info")

	output := buf.String()
	if strings.Contains(output, "debug") {
		t.Error("debug should be filtered with empty level (defaults to info)")
	}
	if !strings.Contains(output, "info") {
		t.Error("info should appear with empty level (defaults to info)")
	}
}

func TestLogger_JSONFormat(t *testing.T) {
	// Create logger with JSON format
	cfg := Config{
		Level:  "info",
		Format: "json",
	}
	logger := NewFromConfig(cfg)
	var buf bytes.Buffer
	logger.output = &buf

	// Log a message
	testMsg := "test json message"
	logger.Info(testMsg)

	output := buf.String()
	output = strings.TrimSpace(output)

	// Verify it's valid JSON
	var entry map[string]string
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("output is not valid JSON: %v, output: %q", err, output)
	}

	// Check required fields
	requiredFields := []string{"timestamp", "level", "message"}
	for _, field := range requiredFields {
		if _, ok := entry[field]; !ok {
			t.Errorf("JSON missing required field %q", field)
		}
	}

	// Check field values
	if entry["level"] != "INFO" {
		t.Errorf("expected level INFO, got %q", entry["level"])
	}
	if entry["message"] != testMsg {
		t.Errorf("expected message %q, got %q", testMsg, entry["message"])
	}

	// Check timestamp format (should match "2006-01-02 15:04:05")
	timestamp := entry["timestamp"]
	if len(timestamp) != 19 || timestamp[4] != '-' || timestamp[7] != '-' ||
		timestamp[10] != ' ' || timestamp[13] != ':' || timestamp[16] != ':' {
		t.Errorf("timestamp %q does not match expected format YYYY-MM-DD HH:MM:SS", timestamp)
	}

	// Also test that debug messages are filtered (level info)
	buf.Reset()
	logger.Debug("debug message")
	output2 := buf.String()
	if output2 != "" {
		t.Errorf("debug message should be filtered with info level, got output: %q", output2)
	}
}
