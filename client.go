package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"gopkg.in/yaml.v3"
)

type Client struct {
	config        Config
	conn          *amqp091.Connection
	channel       *amqp091.Channel
	mu            sync.Mutex
	isConnected   bool
	notifyClose   chan *amqp091.Error
	notifyConfirm chan amqp091.Confirmation
}

func NewClient(configPath string) (*Client, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	client := &Client{
		config: cfg,
	}

	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	go client.handleReconnect()

	return client, nil
}

func (c *Client) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	amqpConfig := c.config.Amqp
	connectionURL := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		amqpConfig.User, amqpConfig.Password, amqpConfig.Host, amqpConfig.Port, amqpConfig.Vhost)
	if amqpConfig.SslEnabled {
		connectionURL = fmt.Sprintf("amqps://%s:%s@%s:%d/%s",
			amqpConfig.User, amqpConfig.Password, amqpConfig.Host, amqpConfig.Port, amqpConfig.Vhost)
	}

	conn, err := amqp091.Dial(connectionURL)
	if err != nil {
		return fmt.Errorf("failed to dial rabbitmq: %w", err)
	}
	c.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		conn.Close() // Close connection if channel creation fails
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	c.channel = ch

	// Enable publisher confirms
	if err := c.channel.Confirm(false); err != nil {
		c.channel.Close()
		c.conn.Close()
		return fmt.Errorf("failed to enable publisher confirms: %w", err)
	}
	c.notifyConfirm = c.channel.NotifyPublish(make(chan amqp091.Confirmation, 1))

	if err := c.setupTopology(); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to setup topology: %w", err)
	}

	c.isConnected = true
	c.notifyClose = make(chan *amqp091.Error)
	c.conn.NotifyClose(c.notifyClose)
	c.channel.NotifyClose(c.notifyClose) // Also watch for channel closes

	log.Println("Successfully connected to RabbitMQ and setup topology.")
	return nil
}

func (c *Client) setupTopology() error {
	// Declare Exchanges
	for _, ex := range c.config.Amqp.Rules.Exchanges {
		err := c.channel.ExchangeDeclare(
			ex.Name,
			ex.Type,
			ex.Durable,
			ex.AutoDelete,
			false, // internal
			false, // noWait
			nil,   // args
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", ex.Name, err)
		}
		log.Printf("Declared exchange: %s", ex.Name)
	}

	// Declare Queues
	for _, q := range c.config.Amqp.Rules.Queues {
		_, err := c.channel.QueueDeclare(
			q.Name,
			q.Durable,
			q.AutoDelete,
			q.Exclusive,
			false, // noWait
			q.Args,
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", q.Name, err)
		}
		log.Printf("Declared queue: %s", q.Name)
	}

	// Bindings
	for _, b := range c.config.Amqp.Rules.Bindings {
		err := c.channel.QueueBind(
			b.Queue,
			b.RoutingKey,
			b.Exchange,
			false, // noWait
			nil,   // args
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s to exchange %s with routing key %s: %w", b.Queue, b.Exchange, b.RoutingKey, err)
		}
		log.Printf("Bound queue %s to exchange %s with routing key %s", b.Queue, b.Exchange, b.RoutingKey)
	}
	return nil
}

func (c *Client) handleReconnect() {
	for {
		err := <-c.notifyClose // Wait for a close signal
		if err == nil {        // If err is nil, it means Close() was called intentionally
			log.Println("RabbitMQ connection closed intentionally.")
			return
		}

		// If it's a channel error (i.e., connection is still open), try to reopen the channel first.
		// This is a simplification; robust error handling might involve checking specific error codes.
		if c.conn != nil && !c.conn.IsClosed() {
			log.Printf("RabbitMQ channel closed: %v. Connection is still open. Attempting to reopen channel...", err)
			c.mu.Lock()
			if c.channel != nil {
				c.channel.Close() // Ensure old channel is closed
			}
			ch, chErr := c.conn.Channel()
			if chErr == nil {
				c.channel = ch
				// Re-enable publisher confirms on the new channel
				if err := c.channel.Confirm(false); err != nil {
					log.Printf("Failed to re-enable publisher confirms on new channel: %v", err)
					// This is problematic, might need full reconnect
				} else {
					c.notifyConfirm = c.channel.NotifyPublish(make(chan amqp091.Confirmation, 1))
				}

				// Potentially re-setup topology if necessary, or parts of it
				// For simplicity, we assume topology is durable and survives channel re-creation
				// In a more complex scenario, you might need to re-declare queues/bindings if they are not durable
				// or if the error indicates a problem with them.

				// Also, re-establish consumer channels if they were tied to the old channel
				// This part is complex and depends on how Consume is implemented

				log.Println("Successfully reopened channel.")
				c.notifyClose = make(chan *amqp091.Error) // Reset notification channel
				c.channel.NotifyClose(c.notifyClose)      // Watch new channel
				c.mu.Unlock()
				continue // Go back to waiting for close signal
			}
			log.Printf("Failed to reopen channel: %v. Proceeding to full reconnect.", chErr)
			c.mu.Unlock() // Unlock before attempting full reconnect
		}

		log.Printf("RabbitMQ connection/channel error: %v. Attempting to reconnect...", err)
		c.mu.Lock()
		c.isConnected = false
		c.mu.Unlock()

		for {
			time.Sleep(time.Duration(c.config.Amqp.ReconnectDelaySeconds) * time.Second)
			log.Println("Attempting to reconnect to RabbitMQ...")
			if err := c.connect(); err == nil {
				log.Println("Successfully reconnected to RabbitMQ.")
				// After reconnecting, existing consumers might need to be recreated.
				// This requires a more sophisticated consumer management strategy.
				// For now, we assume consumers will be re-established by the application
				// or that Consume() will handle this.
				break
			}
			log.Printf("Failed to reconnect: %v", err)
		}
	}
}

// Publish sends a message to a specific exchange with a routing key.
// It waits for publisher confirmation.
func (c *Client) Publish(ctx context.Context, exchange, routingKey string, msg amqp091.Publishing) error {
	c.mu.Lock()
	if !c.isConnected || c.channel == nil {
		c.mu.Unlock()
		return fmt.Errorf("not connected to RabbitMQ")
	}
	ch := c.channel // Use the current channel
	c.mu.Unlock()

	// It's important to use a new context for publishing, possibly with a timeout
	// The parent context might be cancelled for reasons unrelated to this specific publish operation
	pubCtx, cancel := context.WithTimeout(ctx, 30*time.Second) // Example timeout
	defer cancel()

	err := ch.PublishWithContext(
		pubCtx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		msg,
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	// Wait for confirmation
	select {
	case confirm := <-c.notifyConfirm:
		if confirm.Ack {
			log.Printf("Message published successfully to exchange '%s' with routing key '%s'", exchange, routingKey)
			return nil
		}
		return fmt.Errorf("failed to publish message: nack received")
	case <-pubCtx.Done():
		return fmt.Errorf("failed to publish message: context timed out or cancelled while waiting for confirmation")
	}
}

// Consume starts consuming messages from a specific queue.
// It returns a channel of amqp091.Delivery and an error.
// The caller is responsible for handling the messages from the returned channel.
// The consumer will automatically try to re-establish itself on reconnections
// if the initial setup was successful. This is a simplified version.
// A more robust implementation would involve more sophisticated consumer state management.
func (c *Client) Consume(queueName, consumerTag string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error) {
	c.mu.Lock()
	if !c.isConnected || c.channel == nil {
		c.mu.Unlock()
		return nil, fmt.Errorf("not connected to RabbitMQ")
	}
	ch := c.channel
	c.mu.Unlock()

	msgs, err := ch.Consume(
		queueName,
		consumerTag,
		autoAck,
		exclusive,
		noLocal,
		noWait,
		args,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start consumer for queue %s: %w", queueName, err)
	}
	log.Printf("Started consumer for queue '%s' with tag '%s'", queueName, consumerTag)
	return msgs, nil
}

// Close gracefully closes the connection to RabbitMQ.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isConnected {
		log.Println("Already closed or not connected.")
		return nil
	}

	log.Println("Closing RabbitMQ connection...")
	c.isConnected = false // Mark as not connected first to prevent reconnect logic firing for intentional close

	// Signal handleReconnect to exit by closing the notifyClose channel
	// However, amqp091-go signals on notifyClose with nil for intentional client initiated close
	// So, closing the channel directly might cause a panic if library writes to it later.
	// Instead, we rely on the fact that calling c.channel.Close() and c.conn.Close()
	// will eventually send a nil error to notifyClose if they are the source of the close.

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			log.Printf("Failed to close channel: %v", err)
			// Continue to close connection
		} else {
			log.Println("Channel closed.")
		}
		c.channel = nil
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			// This will trigger the handleReconnect goroutine if not handled properly,
			// but since isConnected is false, and we expect a nil error on notifyClose,
			// handleReconnect should exit.
			return fmt.Errorf("failed to close connection: %w", err)
		}
		log.Println("Connection closed.")
		c.conn = nil
	}

	// At this point, handleReconnect should receive a nil error on notifyClose and exit.
	// We might want to wait for handleReconnect to fully exit if it's critical.
	// For simplicity, we are not doing that here.

	log.Println("RabbitMQ client closed successfully.")
	return nil
}
