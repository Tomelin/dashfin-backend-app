package rabbitmq

type AmqpConfig struct {
	Host                  string `yaml:"host"`
	User                  string `yaml:"user"`
	Password              string `yaml:"password"`
	SslEnabled            bool   `yaml:"ssl_enabled"`
	Port                  int    `yaml:"port"`
	Vhost                 string `yaml:"vhost"`
	ReconnectDelaySeconds int    `yaml:"reconnect_delay_seconds"`
	Rules                 Rules  `yaml:"rules"`
}

type Rules struct {
	Exchanges []ExchangeConfig `yaml:"exchanges"`
	Queues    []QueueConfig    `yaml:"queues"`
	Bindings  []BindingConfig  `yaml:"bindings"`
}

type ExchangeConfig struct {
	Name       string `yaml:"name"`
	Type       string `yaml:"type"`
	Durable    bool   `yaml:"durable"`
	AutoDelete bool   `yaml:"auto_delete"`
}

type QueueConfig struct {
	Name       string                 `yaml:"name"`
	Durable    bool                   `yaml:"durable"`
	Exclusive  bool                   `yaml:"exclusive"`
	AutoDelete bool                   `yaml:"auto_delete"`
	Args       map[string]interface{} `yaml:"args"`
}

type BindingConfig struct {
	Queue      string `yaml:"queue"`
	Exchange   string `yaml:"exchange"`
	RoutingKey string `yaml:"routing_key"`
}

type Config struct {
	Amqp AmqpConfig `yaml:"amqp"`
}
