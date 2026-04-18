package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	// Check default values
	if config.HTTP.Host != "localhost" {
		t.Errorf("expected HTTP.Host to be 'localhost', got %q", config.HTTP.Host)
	}
	if config.HTTP.Port != "8080" {
		t.Errorf("expected HTTP.Port to be '8080', got %q", config.HTTP.Port)
	}

	// Check that other fields have zero values
	if config.Logger.Level != "" {
		t.Errorf("expected Logger.Level to be empty, got %q", config.Logger.Level)
	}
	if config.Logger.Output != "" {
		t.Errorf("expected Logger.Output to be empty, got %q", config.Logger.Output)
	}
	if config.Logger.Format != "" {
		t.Errorf("expected Logger.Format to be empty, got %q", config.Logger.Format)
	}
	if config.Storage.Type != "" {
		t.Errorf("expected Storage.Type to be empty, got %q", config.Storage.Type)
	}
	if config.Storage.DSN != "" {
		t.Errorf("expected Storage.DSN to be empty, got %q", config.Storage.DSN)
	}
}

func TestLoadConfig_ValidFullConfig(t *testing.T) {
	configPath := filepath.Join("testdata", "full_config.yml")
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Check loaded values
	if config.Logger.Level != "DEBUG" {
		t.Errorf("expected Logger.Level to be 'DEBUG', got %q", config.Logger.Level)
	}
	if config.Logger.Output != "/var/log/calendar.log" {
		t.Errorf("expected Logger.Output to be '/var/log/calendar.log', got %q", config.Logger.Output)
	}
	if config.Logger.Format != "json" {
		t.Errorf("expected Logger.Format to be 'json', got %q", config.Logger.Format)
	}

	if config.Storage.Type != "sql" {
		t.Errorf("expected Storage.Type to be 'sql', got %q", config.Storage.Type)
	}
	expectedDSN := "host=localhost port=5432 user=postgres password=postgres dbname=calendar"
	if config.Storage.DSN != expectedDSN {
		t.Errorf("expected Storage.DSN to be %q, got %q", expectedDSN, config.Storage.DSN)
	}

	if config.HTTP.Host != "0.0.0.0" {
		t.Errorf("expected HTTP.Host to be '0.0.0.0', got %q", config.HTTP.Host)
	}
	if config.HTTP.Port != "9090" {
		t.Errorf("expected HTTP.Port to be '9090', got %q", config.HTTP.Port)
	}
}

func TestLoadConfig_PartialConfigWithDefaults(t *testing.T) {
	configPath := filepath.Join("testdata", "partial_config.yml")
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Check loaded values
	if config.Logger.Level != "INFO" {
		t.Errorf("expected Logger.Level to be 'INFO' (default), got %q", config.Logger.Level)
	}
	if config.Logger.Output != "" {
		t.Errorf("expected Logger.Output to be empty, got %q", config.Logger.Output)
	}
	if config.Logger.Format != "text" {
		t.Errorf("expected Logger.Format to be 'text' (default), got %q", config.Logger.Format)
	}

	if config.Storage.Type != "memory" {
		t.Errorf("expected Storage.Type to be 'memory', got %q", config.Storage.Type)
	}
	if config.Storage.DSN != "" {
		t.Errorf("expected Storage.DSN to be empty, got %q", config.Storage.DSN)
	}

	if config.HTTP.Host != "127.0.0.1" {
		t.Errorf("expected HTTP.Host to be '127.0.0.1', got %q", config.HTTP.Host)
	}
	if config.HTTP.Port != "8080" {
		t.Errorf("expected HTTP.Port to be '8080' (default), got %q", config.HTTP.Port)
	}
}

func TestLoadConfig_EmptyConfig(t *testing.T) {
	configPath := filepath.Join("testdata", "empty_config.yml")
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// All defaults should be applied
	if config.Logger.Level != "INFO" {
		t.Errorf("expected Logger.Level to be 'INFO' (default), got %q", config.Logger.Level)
	}
	if config.Logger.Output != "" {
		t.Errorf("expected Logger.Output to be empty, got %q", config.Logger.Output)
	}
	if config.Logger.Format != "text" {
		t.Errorf("expected Logger.Format to be 'text' (default), got %q", config.Logger.Format)
	}

	if config.Storage.Type != "memory" {
		t.Errorf("expected Storage.Type to be 'memory' (default), got %q", config.Storage.Type)
	}
	if config.Storage.DSN != "" {
		t.Errorf("expected Storage.DSN to be empty, got %q", config.Storage.DSN)
	}

	if config.HTTP.Host != "localhost" {
		t.Errorf("expected HTTP.Host to be 'localhost' (default), got %q", config.HTTP.Host)
	}
	if config.HTTP.Port != "8080" {
		t.Errorf("expected HTTP.Port to be '8080' (default), got %q", config.HTTP.Port)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	config, err := LoadConfig("nonexistent.yml")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read config file") {
		t.Errorf("error should contain 'failed to read config file', got: %v", err)
	}
	if !strings.Contains(err.Error(), "nonexistent.yml") {
		t.Errorf("error should contain 'nonexistent.yml', got: %v", err)
	}

	// Config should be zero value
	if config != (CalendarConfig{}) {
		t.Errorf("expected zero config, got %+v", config)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	configPath := filepath.Join("testdata", "invalid_config.yml")
	config, err := LoadConfig(configPath)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode YAML config file") {
		t.Errorf("error should contain 'failed to decode YAML config file', got: %v", err)
	}
	if !strings.Contains(err.Error(), "invalid_config.yml") {
		t.Errorf("error should contain 'invalid_config.yml', got: %v", err)
	}

	// Config should be zero value
	if config != (CalendarConfig{}) {
		t.Errorf("expected zero config, got %+v", config)
	}
}

func TestLoadConfigFromDefault_WithValidFile(t *testing.T) {
	configPath := filepath.Join("testdata", "full_config.yml")
	config := LoadConfigFromDefault(configPath)

	// Should load from file
	if config.Logger.Level != "DEBUG" {
		t.Errorf("expected Logger.Level to be 'DEBUG', got %q", config.Logger.Level)
	}
	if config.Logger.Output != "/var/log/calendar.log" {
		t.Errorf("expected Logger.Output to be '/var/log/calendar.log', got %q", config.Logger.Output)
	}
	if config.Logger.Format != "json" {
		t.Errorf("expected Logger.Format to be 'json', got %q", config.Logger.Format)
	}
	if config.Storage.Type != "sql" {
		t.Errorf("expected Storage.Type to be 'sql', got %q", config.Storage.Type)
	}
	if config.HTTP.Host != "0.0.0.0" {
		t.Errorf("expected HTTP.Host to be '0.0.0.0', got %q", config.HTTP.Host)
	}
	if config.HTTP.Port != "9090" {
		t.Errorf("expected HTTP.Port to be '9090', got %q", config.HTTP.Port)
	}
}

func TestLoadConfigFromDefault_WithEmptyString(t *testing.T) {
	config := LoadConfigFromDefault("")

	// Should return default config
	if config.HTTP.Host != "localhost" {
		t.Errorf("expected HTTP.Host to be 'localhost', got %q", config.HTTP.Host)
	}
	if config.HTTP.Port != "8080" {
		t.Errorf("expected HTTP.Port to be '8080', got %q", config.HTTP.Port)
	}
	if config.Logger.Level != "" {
		t.Errorf("expected Logger.Level to be empty (zero value), got %q", config.Logger.Level)
	}
}

func TestLoadConfigFromDefault_WithInvalidFile(t *testing.T) {
	// Create a temporary file that doesn't exist
	config := LoadConfigFromDefault("nonexistent.yml")

	// Should return default config (with warning printed to stderr)
	if config.HTTP.Host != "localhost" {
		t.Errorf("expected HTTP.Host to be 'localhost', got %q", config.HTTP.Host)
	}
	if config.HTTP.Port != "8080" {
		t.Errorf("expected HTTP.Port to be '8080', got %q", config.HTTP.Port)
	}
}

func TestLoadConfigFromDefault_WithInvalidYAML(t *testing.T) {
	configPath := filepath.Join("testdata", "invalid_config.yml")
	config := LoadConfigFromDefault(configPath)

	// Should return default config (with warning printed to stderr)
	if config.HTTP.Host != "localhost" {
		t.Errorf("expected HTTP.Host to be 'localhost', got %q", config.HTTP.Host)
	}
	if config.HTTP.Port != "8080" {
		t.Errorf("expected HTTP.Port to be '8080', got %q", config.HTTP.Port)
	}
}

func TestConfigDefaults(t *testing.T) {
	// Test that defaults are applied correctly when fields are empty
	testCases := []struct {
		name     string
		yaml     string
		expected CalendarConfig
	}{
		{
			name: "empty logger level defaults to INFO",
			yaml: `logger: {}`,
			expected: CalendarConfig{
				Logger:  LoggerConf{Level: "INFO", Format: "text"},
				Storage: StorageConf{Type: "memory"},
				HTTP:    HTTPConf{Host: "localhost", Port: "8080"},
			},
		},
		{
			name: "empty storage type defaults to memory",
			yaml: `storage: {}`,
			expected: CalendarConfig{
				Logger:  LoggerConf{Level: "INFO", Format: "text"},
				Storage: StorageConf{Type: "memory"},
				HTTP:    HTTPConf{Host: "localhost", Port: "8080"},
			},
		},
		{
			name: "empty http host defaults to localhost",
			yaml: `http: {}`,
			expected: CalendarConfig{
				Logger:  LoggerConf{Level: "INFO", Format: "text"},
				Storage: StorageConf{Type: "memory"},
				HTTP:    HTTPConf{Host: "localhost", Port: "8080"},
			},
		},
		{
			name: "explicit values override defaults",
			yaml: `
logger:
  level: "ERROR"
  format: "json"
storage:
  type: "sql"
  dsn: "test"
http:
  host: "example.com"
  port: "3000"
`,
			expected: CalendarConfig{
				Logger:  LoggerConf{Level: "ERROR", Format: "json"},
				Storage: StorageConf{Type: "sql", DSN: "test"},
				HTTP:    HTTPConf{Host: "example.com", Port: "3000"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary file
			tmpFile, err := os.CreateTemp("", "config_test_*.yml")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(tc.yaml)
			if err != nil {
				t.Fatalf("failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			config, err := LoadConfig(tmpFile.Name())
			if err != nil {
				t.Fatalf("LoadConfig failed: %v", err)
			}

			// Compare fields
			if config.Logger.Level != tc.expected.Logger.Level {
				t.Errorf("Logger.Level: expected %q, got %q", tc.expected.Logger.Level, config.Logger.Level)
			}
			if config.Logger.Format != tc.expected.Logger.Format {
				t.Errorf("Logger.Format: expected %q, got %q", tc.expected.Logger.Format, config.Logger.Format)
			}
			if config.Storage.Type != tc.expected.Storage.Type {
				t.Errorf("Storage.Type: expected %q, got %q", tc.expected.Storage.Type, config.Storage.Type)
			}
			if config.Storage.DSN != tc.expected.Storage.DSN {
				t.Errorf("Storage.DSN: expected %q, got %q", tc.expected.Storage.DSN, config.Storage.DSN)
			}
			if config.HTTP.Host != tc.expected.HTTP.Host {
				t.Errorf("HTTP.Host: expected %q, got %q", tc.expected.HTTP.Host, config.HTTP.Host)
			}
			if config.HTTP.Port != tc.expected.HTTP.Port {
				t.Errorf("HTTP.Port: expected %q, got %q", tc.expected.HTTP.Port, config.HTTP.Port)
			}
		})
	}
}

func TestConfigYAMLTags(t *testing.T) {
	// Test that YAML tags are correctly mapped
	yaml := `
logger:
  level: "WARN"
  output: "stderr"
  format: "json"
storage:
  type: "sql"
  dsn: "postgres://user:pass@localhost/db"
http:
  host: "192.168.1.1"
  port: "8081"
`

	tmpFile, err := os.CreateTemp("", "config_test_*.yml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yaml)
	if err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	config, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if config.Logger.Level != "WARN" {
		t.Errorf("expected Logger.Level to be 'WARN', got %q", config.Logger.Level)
	}
	if config.Logger.Output != "stderr" {
		t.Errorf("expected Logger.Output to be 'stderr', got %q", config.Logger.Output)
	}
	if config.Logger.Format != "json" {
		t.Errorf("expected Logger.Format to be 'json', got %q", config.Logger.Format)
	}
	if config.Storage.Type != "sql" {
		t.Errorf("expected Storage.Type to be 'sql', got %q", config.Storage.Type)
	}
	if config.Storage.DSN != "postgres://user:pass@localhost/db" {
		t.Errorf("expected Storage.DSN to be 'postgres://user:pass@localhost/db', got %q", config.Storage.DSN)
	}
	if config.HTTP.Host != "192.168.1.1" {
		t.Errorf("expected HTTP.Host to be '192.168.1.1', got %q", config.HTTP.Host)
	}
	if config.HTTP.Port != "8081" {
		t.Errorf("expected HTTP.Port to be '8081', got %q", config.HTTP.Port)
	}
}
