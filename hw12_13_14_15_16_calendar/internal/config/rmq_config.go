package config

type RabbitMQConf struct {
	URI              string `yaml:"uri"`
	QueueName        string `yaml:"queueName"`
	ExchangeName     string `yaml:"exchangeName"`
	StatusQueueName  string `yaml:"statusQueueName,omitempty"`
	StatusRoutingKey string `yaml:"statusRoutingKey,omitempty"`
}
