package config

type RabbitMQConf struct {
	URI          string `yaml:"uri"`
	QueueName    string `yaml:"queueName"`
	ExchangeName string `yaml:"exchangeName"`
}
