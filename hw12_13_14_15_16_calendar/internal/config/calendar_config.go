package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type CalendarConfig struct {
	Logger  LoggerConf
	Storage StorageConf
	HTTP    HTTPConf
}

func NewConfig() CalendarConfig {
	// For backward compatibility, return empty config.
	// In production, LoadConfig should be used.
	return CalendarConfig{
		HTTP: HTTPConf{
			Host: "localhost",
			Port: "8080",
		},
	}
}

// LoadConfig reads configuration from the specified YAML file.
func LoadConfig(filename string) (CalendarConfig, error) {
	var config CalendarConfig

	data, err := os.ReadFile(filename)
	if err != nil {
		return config, fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to decode YAML config file %s: %w", filename, err)
	}

	// Set defaults
	if config.Logger.Level == "" {
		config.Logger.Level = "INFO"
	}
	if config.Logger.Format == "" {
		config.Logger.Format = "text"
	}
	if config.Storage.Type == "" {
		config.Storage.Type = "memory"
	}
	if config.HTTP.Host == "" {
		config.HTTP.Host = "localhost"
	}
	if config.HTTP.Port == "" {
		config.HTTP.Port = "8080"
	}

	// Apply environment variable overrides
	applyEnvOverrides(&config)

	return config, nil
}

// applyEnvOverrides updates config fields from environment variables.
func applyEnvOverrides(config *CalendarConfig) {
	// Logger
	if v := os.Getenv("CALENDAR_LOGGER_LEVEL"); v != "" {
		config.Logger.Level = v
	}
	if v := os.Getenv("CALENDAR_LOGGER_OUTPUT"); v != "" {
		config.Logger.Output = v
	}
	if v := os.Getenv("CALENDAR_LOGGER_FORMAT"); v != "" {
		config.Logger.Format = v
	}

	// Storage
	if v := os.Getenv("CALENDAR_STORAGE_TYPE"); v != "" {
		config.Storage.Type = v
	}
	if v := os.Getenv("CALENDAR_STORAGE_DSN"); v != "" {
		config.Storage.DSN = v
	}

	// HTTP
	if v := os.Getenv("CALENDAR_HTTP_HOST"); v != "" {
		config.HTTP.Host = v
	}
	if v := os.Getenv("CALENDAR_HTTP_PORT"); v != "" {
		config.HTTP.Port = v
	}
}

// LoadConfigFromDefault attempts to load config from the file specified by the -config flag.
// If the file does not exist or flag is not set, returns default config with environment overrides applied.
func LoadConfigFromDefault(configFile string) CalendarConfig {
	var config CalendarConfig
	if configFile == "" {
		config = NewConfig()
		applyEnvOverrides(&config)
		return config
	}
	loadedConfig, err := LoadConfig(configFile)
	if err != nil {
		// Log error but continue with default config (maybe we should exit)
		fmt.Fprintf(os.Stderr, "WARNING: failed to load config: %v\n", err)
		config = NewConfig()
		applyEnvOverrides(&config)
		return config
	}
	return loadedConfig
}
