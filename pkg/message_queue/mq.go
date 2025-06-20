package message_queue

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// QueueConfig configuração de uma queue específica
type QueueConfig struct {
	Name       string                 `yaml:"name" json:"name"`
	Durable    bool                   `yaml:"durable" json:"durable"`
	AutoDelete bool                   `yaml:"auto_delete" json:"auto_delete"`
	Exclusive  bool                   `yaml:"exclusive" json:"exclusive"`
	NoWait     bool                   `yaml:"no_wait" json:"no_wait"`
	Args       map[string]interface{} `yaml:"args" json:"args"`
	RouteKey   string                 `yaml:"route_key" json:"route_key"`   // Para compatibilidade
	RouteKeys  []string               `yaml:"route_keys" json:"route_keys"` // Múltiplas route keys
	DeadLetter DeadLetterConfig       `yaml:"dead_letter" json:"dead_letter"`
	Consumer   ConsumerConfig         `yaml:"consumer" json:"consumer"`
	Publisher  PublisherConfig        `yaml:"publisher" json:"publisher"`
}

// DeadLetterConfig configuração de dead letter
type DeadLetterConfig struct {
	Exchange string `yaml:"exchange" json:"exchange"`
	Queue    string `yaml:"queue" json:"queue"`
	RouteKey string `yaml:"route_key" json:"route_key"`
}

// ConsumerConfig configuração do consumer
type ConsumerConfig struct {
	Tag       string                 `yaml:"tag" json:"tag"`
	AutoAck   bool                   `yaml:"auto_ack" json:"auto_ack"`
	Exclusive bool                   `yaml:"exclusive" json:"exclusive"`
	NoLocal   bool                   `yaml:"no_local" json:"no_local"`
	NoWait    bool                   `yaml:"no_wait" json:"no_wait"`
	Args      map[string]interface{} `yaml:"args" json:"args"`
}

// PublisherConfig configuração do publisher
type PublisherConfig struct {
	Mandatory bool `yaml:"mandatory" json:"mandatory"`
	Immediate bool `yaml:"immediate" json:"immediate"`
}

// ExchangeConfig configuração de uma exchange
type ExchangeConfig struct {
	Name    string        `yaml:"exchange" json:"exchange"`
	Type    string        `yaml:"type" json:"type"`
	Durable bool          `yaml:"durable" json:"durable"`
	Queues  []QueueConfig `yaml:"queues" json:"queues"`
}

// Config estrutura principal de configuração para RabbitMQ
type Config struct {
	URL              string           `yaml:"url" json:"url"`
	MessageQueues    []ExchangeConfig `yaml:"message_queues" json:"message_queues"`
	DefaultConsumer  ConsumerConfig   `yaml:"default_consumer" json:"default_consumer"`
	DefaultPublisher PublisherConfig  `yaml:"default_publisher" json:"default_publisher"`
}

// MessageQueue interface principal da biblioteca
type MessageQueue interface {
	Consumer(ctx context.Context, exchangeName, queueName string, handler func([]byte, string) error) error
	Publisher(exchangeName, queueName string, message []byte, traceID string) error
	PublisherWithRouteKey(exchangeName, routeKey string, message []byte, traceID string) error
	Close() error
	Setup() error
}

// RabbitMQ implementação da interface MessageQueue
type RabbitMQ struct {
	config      Config
	conn        *amqp.Connection
	channel     *amqp.Channel
	closeChan   chan *amqp.Error
	isConnected bool
	exchanges   map[string]*ExchangeConfig
	queues      map[string]*QueueConfig
}

// NewRabbitMQ cria uma nova instância do RabbitMQ
func NewRabbitMQ(config Config) (MessageQueue, error) {
	log.Println("Initializing RabbitMQ")
	mq := &RabbitMQ{
		config:    config,
		closeChan: make(chan *amqp.Error),
		exchanges: make(map[string]*ExchangeConfig),
		queues:    make(map[string]*QueueConfig),
	}

	// Mapeia exchanges e queues para acesso rápido
	for i := range config.MessageQueues {
		exchange := &config.MessageQueues[i]
		mq.exchanges[exchange.Name] = exchange

		for j := range exchange.Queues {
			queue := &exchange.Queues[j]
			mq.queues[queue.Name] = queue
		}
	}

	err := mq.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return mq, nil
}

// connect estabelece conexão com RabbitMQ
func (mq *RabbitMQ) connect() error {
	conn, err := amqp.Dial(mq.config.URL)
	if err != nil {
		return err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}

	mq.conn = conn
	mq.channel = channel
	mq.isConnected = true

	// Monitora a conexão
	mq.conn.NotifyClose(mq.closeChan)

	return nil
}

// getRouteKeys retorna as route keys de uma queue (suporta tanto RouteKey quanto RouteKeys)
func (q *QueueConfig) getRouteKeys() []string {
	if len(q.RouteKeys) > 0 {
		return q.RouteKeys
	}
	if q.RouteKey != "" {
		return []string{q.RouteKey}
	}
	return []string{""} // Para fanout exchanges
}

// Setup configura todas as exchanges, queues e dead letters
func (mq *RabbitMQ) Setup() error {
	if !mq.isConnected {
		return fmt.Errorf("not connected to RabbitMQ")
	}

	// Declara todas as exchanges
	for _, exchange := range mq.config.MessageQueues {
		exchangeType := exchange.Type
		if exchangeType == "" {
			exchangeType = "direct"
		}

		err := mq.channel.ExchangeDeclare(
			exchange.Name,
			exchangeType,
			exchange.Durable,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", exchange.Name, err)
		}

		// Declara dead letter exchanges para cada queue
		for _, queue := range exchange.Queues {
			if queue.DeadLetter.Exchange != "" {
				err = mq.channel.ExchangeDeclare(
					queue.DeadLetter.Exchange,
					"direct",
					true,
					false,
					false,
					false,
					nil,
				)
				if err != nil {
					return fmt.Errorf("failed to declare dead letter exchange %s: %w", queue.DeadLetter.Exchange, err)
				}

				// Declara dead letter queue
				if queue.DeadLetter.Queue != "" {
					_, err = mq.channel.QueueDeclare(
						queue.DeadLetter.Queue,
						true,
						false,
						false,
						false,
						nil,
					)
					if err != nil {
						return fmt.Errorf("failed to declare dead letter queue %s: %w", queue.DeadLetter.Queue, err)
					}

					// Bind dead letter queue
					err = mq.channel.QueueBind(
						queue.DeadLetter.Queue,
						queue.DeadLetter.RouteKey,
						queue.DeadLetter.Exchange,
						false,
						nil,
					)
					if err != nil {
						return fmt.Errorf("failed to bind dead letter queue %s: %w", queue.DeadLetter.Queue, err)
					}
				}
			}
		}

		// Declara todas as queues da exchange
		for _, queue := range exchange.Queues {
			queueArgs := make(map[string]interface{})
			if queue.Args != nil {
				for k, v := range queue.Args {
					queueArgs[k] = v
				}
			}

			// Adiciona dead letter se configurado
			if queue.DeadLetter.Exchange != "" {
				queueArgs["x-dead-letter-exchange"] = queue.DeadLetter.Exchange
				if queue.DeadLetter.RouteKey != "" {
					queueArgs["x-dead-letter-routing-key"] = queue.DeadLetter.RouteKey
				}
			}

			// Declara queue
			_, err = mq.channel.QueueDeclare(
				queue.Name,
				queue.Durable,
				queue.AutoDelete,
				queue.Exclusive,
				queue.NoWait,
				queueArgs,
			)
			if err != nil {
				return fmt.Errorf("failed to declare queue %s: %w", queue.Name, err)
			}

			// Bind queue ao exchange com múltiplas route keys
			routeKeys := queue.getRouteKeys()
			for _, routeKey := range routeKeys {
				err = mq.channel.QueueBind(
					queue.Name,
					routeKey,
					exchange.Name,
					false,
					nil,
				)
				if err != nil {
					return fmt.Errorf("failed to bind queue %s to exchange %s with route key %s: %w",
						queue.Name, exchange.Name, routeKey, err)
				}
			}
		}
	}

	return nil
}

// Publisher publica uma mensagem na queue especificada
func (mq *RabbitMQ) Publisher(exchangeName, queueName string, message []byte, traceID string) error {
	if !mq.isConnected {
		return fmt.Errorf("not connected to RabbitMQ")
	}

	// Busca configuração da queue
	queueConfig, exists := mq.queues[queueName]
	if !exists {
		return fmt.Errorf("queue %s not found in configuration", queueName)
	}

	// Busca configuração da exchange
	exchangeConfig, exists := mq.exchanges[exchangeName]
	if !exists {
		return fmt.Errorf("exchange %s not found in configuration", exchangeName)
	}

	// Usa configuração específica da queue ou configuração padrão
	publisherConfig := queueConfig.Publisher
	if !publisherConfig.Mandatory && !publisherConfig.Immediate {
		publisherConfig = mq.config.DefaultPublisher
	}

	// Usa a primeira route key para publish (ou route key específica se fornecida)
	routeKeys := queueConfig.getRouteKeys()
	routeKey := routeKeys[0] // Usa a primeira route key por padrão

	// Cria headers incluindo X-TRACE-ID
	headers := amqp.Table{}
	if traceID != "" {
		headers["X-TRACE-ID"] = traceID
	}

	return mq.channel.Publish(
		exchangeConfig.Name,
		routeKey,
		publisherConfig.Mandatory,
		publisherConfig.Immediate,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         message,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Headers:      headers,
		},
	)
}

// PublisherWithRouteKey publica mensagem usando route key específica
func (mq *RabbitMQ) PublisherWithRouteKey(exchangeName, routeKey string, message []byte, traceID string) error {
	if !mq.isConnected {
		return fmt.Errorf("not connected to RabbitMQ")
	}

	// Busca configuração da exchange
	exchangeConfig, exists := mq.exchanges[exchangeName]
	if !exists {
		return fmt.Errorf("exchange %s not found in configuration", exchangeName)
	}

	// Usa configuração padrão do publisher
	publisherConfig := mq.config.DefaultPublisher

	// Cria headers incluindo X-TRACE-ID
	headers := amqp.Table{}
	if traceID != "" {
		headers["X-TRACE-ID"] = traceID
	}

	return mq.channel.Publish(
		exchangeConfig.Name,
		routeKey,
		publisherConfig.Mandatory,
		publisherConfig.Immediate,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         message,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Headers:      headers,
		},
	)
}

// Consumer consome mensagens da queue especificada
func (mq *RabbitMQ) Consumer(ctx context.Context, exchangeName, queueName string, handler func([]byte, string) error) error {
	if !mq.isConnected {
		return fmt.Errorf("not connected to RabbitMQ")
	}

	// Busca configuração da queue
	queueConfig, exists := mq.queues[queueName]
	if !exists {
		return fmt.Errorf("queue %s not found in configuration", queueName)
	}

	// Usa configuração específica da queue ou configuração padrão
	consumerConfig := queueConfig.Consumer
	if consumerConfig.Tag == "" && !consumerConfig.AutoAck && !consumerConfig.Exclusive &&
		!consumerConfig.NoLocal && !consumerConfig.NoWait && consumerConfig.Args == nil {
		consumerConfig = mq.config.DefaultConsumer
	}

	msgs, err := mq.channel.Consume(
		queueConfig.Name,
		consumerConfig.Tag,
		consumerConfig.AutoAck,
		consumerConfig.Exclusive,
		consumerConfig.NoLocal,
		consumerConfig.NoWait,
		consumerConfig.Args,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-mq.closeChan:
			if err != nil {
				return fmt.Errorf("connection closed: %w", err)
			}
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("message channel closed")
			}

			// Extrai X-TRACE-ID dos headers
			traceID := ""
			if msg.Headers != nil {
				if id, exists := msg.Headers["X-TRACE-ID"]; exists {
					if idStr, ok := id.(string); ok {
						traceID = idStr
					}
				}
			}

			err := handler(msg.Body, traceID)
			if err != nil {
				// Rejeita mensagem e envia para dead letter se configurado
				msg.Nack(false, false)
				log.Printf("Message processing failed: %v", err)
			} else {
				// Confirma processamento da mensagem
				if !consumerConfig.AutoAck {
					msg.Ack(false)
				}
			}
		}
	}
}

// Close fecha a conexão com RabbitMQ
func (mq *RabbitMQ) Close() error {
	if !mq.isConnected {
		return nil
	}

	mq.isConnected = false

	if mq.channel != nil {
		if err := mq.channel.Close(); err != nil {
			log.Printf("Error closing channel: %v", err)
		}
	}

	if mq.conn != nil {
		if err := mq.conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
			return err
		}
	}

	return nil
}
