package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger  LoggerConf
	Storage StorageConf
	HTTP    HTTPConf
}

type LoggerConf struct {
	Level  string `yaml:"level"`
	Output string `yaml:"output"` // file path, empty for stdout
	Format string `yaml:"format"` // "text" or "json"
}

type StorageConf struct {
	Type string `yaml:"type"` // "memory" or "sql"
	DSN  string `yaml:"dsn"`  // Data Source Name for SQL storage, optional for memory
}

type HTTPConf struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func NewConfig() Config {
	// For backward compatibility, return empty config.
	// In production, LoadConfig should be used.
	return Config{
		HTTP: HTTPConf{
			Host: "localhost",
			Port: "8080",
		},
	}
}

// LoadConfig reads configuration from the specified YAML file.
func LoadConfig(filename string) (Config, error) {
	var config Config

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
	return config, nil
}

// LoadConfigFromDefault attempts to load config from the file specified by the -config flag.
// If the file does not exist or flag is not set, returns default config.
func LoadConfigFromDefault(configFile string) Config {
	if configFile == "" {
		return NewConfig()
	}
	config, err := LoadConfig(configFile)
	if err != nil {
		// Log error but continue with default config (maybe we should exit)
		fmt.Fprintf(os.Stderr, "WARNING: failed to load config: %v\n", err)
		return NewConfig()
	}
	return config
}
