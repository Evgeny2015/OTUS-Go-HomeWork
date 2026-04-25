package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type SchedulerConf struct {
	Interval     string `yaml:"interval"`     // e.g., "1m", "5m"
	LookAhead    string `yaml:"lookAhead"`    // how far ahead to check for notifications
	CleanupOlder string `yaml:"cleanupOlder"` // delete events older than this duration
}

// SchedulerConfig holds configuration specific to the scheduler service.
type SchedulerConfig struct {
	Logger    LoggerConf
	Storage   StorageConf
	RabbitMQ  RabbitMQConf  `yaml:"rabbitmq"`
	Scheduler SchedulerConf `yaml:"scheduler"`
}

func NewSchedulerConfig() SchedulerConfig {
	return SchedulerConfig{
		RabbitMQ: RabbitMQConf{
			URI:          "amqp://guest:guest@localhost:5672/",
			QueueName:    "calendar_notifications",
			ExchangeName: "calendar_exchange",
		},
		Scheduler: SchedulerConf{
			Interval:     "1m",
			LookAhead:    "24h",
			CleanupOlder: "8760h",
		},
	}
}

// LoadSchedulerConfig reads configuration from the specified YAML file.
func LoadSchedulerConfig(filename string) (SchedulerConfig, error) {
	var config SchedulerConfig

	data, err := os.ReadFile(filename)
	if err != nil {
		return config, fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to decode YAML config file %s: %w", filename, err)
	}

	if config.RabbitMQ.URI == "" {
		config.RabbitMQ.URI = "amqp://guest:guest@localhost:5672/"
	}
	if config.RabbitMQ.QueueName == "" {
		config.RabbitMQ.QueueName = "calendar_notifications"
	}
	if config.RabbitMQ.ExchangeName == "" {
		config.RabbitMQ.ExchangeName = "calendar_exchange"
	}
	if config.Scheduler.Interval == "" {
		config.Scheduler.Interval = "1m"
	}
	if config.Scheduler.LookAhead == "" {
		config.Scheduler.LookAhead = "24h"
	}
	if config.Scheduler.CleanupOlder == "" {
		config.Scheduler.CleanupOlder = "8760h" // 1 year in hours
	}
	return config, nil
}
