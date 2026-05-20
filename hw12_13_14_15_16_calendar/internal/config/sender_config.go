package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// SenderConf holds configuration specific to the sender service.
type SenderConf struct {
	PollInterval string `yaml:"pollInterval"` // e.g., "1s", "5s"
	BatchSize    int    `yaml:"batchSize"`    // number of messages to process at once
}

// SenderConfig holds configuration for the sender service.
type SenderConfig struct {
	Logger   LoggerConf   `yaml:"logger"`
	RabbitMQ RabbitMQConf `yaml:"rabbitmq"`
	Sender   SenderConf   `yaml:"sender"`
}

func NewSenderConfig() SenderConfig {
	return SenderConfig{
		RabbitMQ: RabbitMQConf{
			URI:          "amqp://guest:guest@localhost:5672/",
			QueueName:    "calendar_notifications",
			ExchangeName: "calendar_exchange",
		},
		Sender: SenderConf{
			PollInterval: "1s",
			BatchSize:    10,
		},
	}
}

// LoadSenderConfig reads configuration from the specified YAML file.
func LoadSenderConfig(filename string) (SenderConfig, error) {
	var config SenderConfig

	data, err := os.ReadFile(filename)
	if err != nil {
		return config, fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to decode YAML config file %s: %w", filename, err)
	}

	// Apply environment variable overrides
	applySenderEnvOverrides(&config)

	return config, nil
}

// applySenderEnvOverrides updates config fields from environment variables.
func applySenderEnvOverrides(config *SenderConfig) {
	// Logger
	if v := os.Getenv("SENDER_LOGGER_LEVEL"); v != "" {
		config.Logger.Level = v
	}
	if v := os.Getenv("SENDER_LOGGER_OUTPUT"); v != "" {
		config.Logger.Output = v
	}
	if v := os.Getenv("SENDER_LOGGER_FORMAT"); v != "" {
		config.Logger.Format = v
	}

	// RabbitMQ
	if v := os.Getenv("SENDER_RABBITMQ_URI"); v != "" {
		config.RabbitMQ.URI = v
	}
	if v := os.Getenv("SENDER_RABBITMQ_QUEUE"); v != "" {
		config.RabbitMQ.QueueName = v
	}
	if v := os.Getenv("SENDER_RABBITMQ_EXCHANGE"); v != "" {
		config.RabbitMQ.ExchangeName = v
	}

	// Sender
	if v := os.Getenv("SENDER_POLL_INTERVAL"); v != "" {
		config.Sender.PollInterval = v
	}
	if v := os.Getenv("SENDER_BATCH_SIZE"); v != "" {
		// Try to parse as int; if fails, keep default
		var batch int
		if _, err := fmt.Sscanf(v, "%d", &batch); err == nil {
			config.Sender.BatchSize = batch
		}
	}
}
