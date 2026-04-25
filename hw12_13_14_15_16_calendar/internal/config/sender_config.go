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

	return config, nil
}
