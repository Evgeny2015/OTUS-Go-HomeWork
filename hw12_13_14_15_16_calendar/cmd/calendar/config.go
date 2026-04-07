package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger  LoggerConf
	Storage StorageConf
}

type LoggerConf struct {
	Level  string `toml:"level"`
	Output string `toml:"output"` // file path, empty for stdout
	Format string `toml:"format"` // "text" or "json"
}

type StorageConf struct {
	Type string `toml:"type"` // "memory" or "sql"
	DSN  string `toml:"dsn"`  // Data Source Name for SQL storage, optional for memory
}

func NewConfig() Config {
	// For backward compatibility, return empty config.
	// In production, LoadConfig should be used.
	return Config{}
}

// LoadConfig reads configuration from the specified TOML file.
func LoadConfig(filename string) (Config, error) {
	var config Config
	if _, err := toml.DecodeFile(filename, &config); err != nil {
		return config, fmt.Errorf("failed to decode config file %s: %w", filename, err)
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
	return config, nil
}

// LoadConfigFromDefault attempts to load config from the file specified by the -config flag.
// If the file does not exist or flag is not set, returns default config.
func LoadConfigFromDefault() Config {
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
