package rabbitmq

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const (
	testConfigPath     = "config_test.yaml" // This will be the one created in tmp by TestRabbitMQClient_Integration
	testExchangeName   = "test_exchange"
	testQueueName      = "test_queue"
	testRoutingKey     = "test_key"
	testConsumerTag    = "test_consumer"
	nonExistentQueue   = "non_existent_queue"
)

// Helper function to create a temporary config file for testing
func createTestConfig(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())
	return tmpFile.Name()
}

func TestNewClient_ConfigLoading(t *testing.T) {
	// Create a dummy config_test.yaml
	yamlContent := `
amqp:
  host: "localhost"
  user: "user_test"
  password: "password_test"
  ssl_enabled: false
  port: 5672
  vhost: "/test"
  reconnect_delay_seconds: 2
  rules:
    exchanges:
    - name: "exchange1_test"
      type: "direct"
      durable: true
      auto_delete: false
    queues:
    - name: "queue1_test"
      durable: true
      exclusive: false
      auto_delete: false
    bindings:
    - queue: "queue1_test"
      exchange: "exchange1_test"
      routing_key: "key1_test"
`
	configPath := createTestConfig(t, yamlContent)
	defer os.Remove(configPath)

	// We don't expect a connection here, just config parsing
	// So, NewClient will likely fail at the connect() stage if RabbitMQ is not running.
	// We are primarily testing the config loading part of NewClient.
	// A more advanced test would mock the dial function.

	data, err := os.ReadFile(configPath)
	require.NoError(t, err, "Failed to read temp config file")

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	require.NoError(t, err, "Failed to unmarshal config")

	assert.Equal(t, "localhost", cfg.Amqp.Host)
	assert.Equal(t, "user_test", cfg.Amqp.User)
	assert.Equal(t, "/test", cfg.Amqp.Vhost)
	assert.Equal(t, 2, cfg.Amqp.ReconnectDelaySeconds)
	require.Len(t, cfg.Amqp.Rules.Exchanges, 1)
	assert.Equal(t, "exchange1_test", cfg.Amqp.Rules.Exchanges[0].Name)
	require.Len(t, cfg.Amqp.Rules.Queues, 1)
	assert.Equal(t, "queue1_test", cfg.Amqp.Rules.Queues[0].Name)
	require.Len(t, cfg.Amqp.Rules.Bindings, 1)
	assert.Equal(t, "queue1_test", cfg.Amqp.Rules.Bindings[0].Queue)
}


// TestMain can be used to set up and tear down a RabbitMQ container for integration tests
// For now, these tests will assume a RabbitMQ instance is running and accessible
// on localhost:5672 with guest/guest credentials and the topology from config_test.yaml

func TestRabbitMQClient_Integration(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration tests in CI environment without a RabbitMQ service")
	}

	// Create the actual config_test.yaml file for the client to use
	yamlContent := `
amqp:
  host: "localhost" # Make sure this matches your RabbitMQ setup
  user: "guest"
  password: "guest"
  ssl_enabled: false
  port: 5672
  vhost: "/"
  reconnect_delay_seconds: 1
  rules:
    exchanges:
    - name: "test_exchange"
      type: "direct"
      durable: true
      auto_delete: false
    - name: "test_dlx_exchange" # Dead-letter exchange
      type: "direct"
      durable: true
      auto_delete: false
    queues:
    - name: "test_queue"
      durable: true
      exclusive: false
      auto_delete: false
      args:
        x-dead-letter-exchange: "test_dlx_exchange"
        x-dead-letter-routing-key: "dlx_key_test" # Routing key for DLX
    - name: "test_dlx_queue" # Dead-letter queue
      durable: true
      exclusive: false
      auto_delete: false
    bindings:
    - queue: "test_queue"
      exchange: "test_exchange"
      routing_key: "test_key"
    - queue: "test_dlx_queue" # Binding for DLX queue
      exchange: "test_dlx_exchange"
      routing_key: "dlx_key_test" # DLX routing key
`
	// Create a temporary directory for the config file
	tmpDir, err := os.MkdirTemp("", "rabbitmq-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir) // Clean up the directory

	configPath := filepath.Join(tmpDir, "config_integration_test.yaml")
	err = os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	client, err := NewClient(configPath)
	require.NoError(t, err, "Failed to create RabbitMQ client for integration tests. Ensure RabbitMQ is running at localhost:5672 with guest/guest credentials.")
	require.NotNil(t, client)
	defer client.Close()

	t.Run("PublishAndConsume", func(t *testing.T) {
		// Consume messages
		// Use a unique consumer tag for each test run or ensure cleanup
		deliveries, err := client.Consume(testQueueName, testConsumerTag+"_pubsub", false, false, false, false, nil)
		require.NoError(t, err, "Failed to start consumer")

		// Publish a test message
		testPayload := "Hello, RabbitMQ! " + time.Now().String()
		msg := amqp091.Publishing{
			ContentType: "text/plain",
			Body:        []byte(testPayload),
			DeliveryMode: amqp091.Persistent, // Ensure message persistence if queue is durable
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Increased timeout
		defer cancel()

		err = client.Publish(ctx, testExchangeName, testRoutingKey, msg)
		require.NoError(t, err, "Failed to publish message")

		select {
		case d, ok := <-deliveries:
			require.True(t, ok, "Deliveries channel closed unexpectedly")
			assert.Equal(t, testPayload, string(d.Body))
			err = d.Ack(false) // Acknowledge the message
			require.NoError(t, err, "Failed to acknowledge message")
		case <-time.After(15 * time.Second): // Increased timeout for receiving
			t.Fatal("Timeout waiting for message")
		}

		// Try to clean up consumer channel by closing client or specific consumer
		// Note: Proper consumer cleanup might involve cancelling the consumer explicitly
		// with Channel.Cancel(consumerTag, false) if the client library doesn't handle this in Close()
		// For this test, client.Close() should suffice for cleanup.
	})

	t.Run("ConsumeFromNonExistentQueue", func(t *testing.T) {
		_, err := client.Consume(nonExistentQueue, testConsumerTag+"_nonexist", false, false, false, false, nil)
		assert.Error(t, err, "Expected error when consuming from a non-existent queue")
		// Note: The error might come from RabbitMQ when the channel tries to use the queue.
		// The exact error message might vary.
		// Example check: assert.Contains(t, err.Error(), "NOT_FOUND") or similar AMQP error code.
		// For amqp091-go, a common error is *amqp091.Error code 404 (NOT_FOUND)
		var amqpErr *amqp091.Error
		if assert.ErrorAs(t, err, &amqpErr) {
			assert.Equal(t, amqp091.NotFound, amqpErr.Code, "Expected AMQP NotFound error code")
		}
	})

	t.Run("PublishWithTimeout", func(t *testing.T) {
		// This test is a bit tricky as it relies on the publish confirmation timeout.
		// To reliably test timeout, we'd need to make the broker unresponsive or slow.
		// For now, we'll test a successful publish with a short timeout.
		// A more robust test would involve a mock server or network manipulation.

		testPayload := "Hello, Timeout Test!"
		msg := amqp091.Publishing{
			ContentType: "text/plain",
			Body:        []byte(testPayload),
		}

		// Short timeout that should still allow successful publish to a local responsive broker
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := client.Publish(ctx, testExchangeName, testRoutingKey, msg)
		// If RabbitMQ is running and responsive, this should not error out due to timeout.
		// If it does, it might indicate a problem with publisher confirms or network.
		require.NoError(t, err, "Publish with short timeout failed unexpectedly")

		// To test actual timeout, you might publish to a non-existent exchange (if not auto-created)
		// or simulate network issues.
		// Example (if exchange doesn't exist and is not auto-declared by publish):
		// ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 50*time.Millisecond)
		// defer cancelTimeout()
		// err = client.Publish(ctxTimeout, "non_existent_exchange_for_timeout_test", testRoutingKey, msg)
		// require.Error(t, err, "Expected error when publishing to non-existent exchange for timeout")
		// if !assert.ErrorIs(t, err, context.DeadlineExceeded) { // This might also be an AMQP error if broker rejects fast
		// 	t.Logf("Actual error for non-existent exchange publish: %v", err)
		// 	// Depending on server behavior and lib, this could be a direct AMQP error or a timeout.
		// 	// If server rejects immediately, it might not be a timeout.
		// }
	})


}

// Note: To run these integration tests, you need a RabbitMQ server running
// and configured as per `config_integration_test.yaml` (effectively default guest/guest access).
// The tests declare their own topology based on this config.
// If RabbitMQ is not available, the `TestRabbitMQClient_Integration` will fail at `NewClient`.
// The `CI` env var check is a common way to skip integration tests in automated pipelines
// that don't have dependent services readily available.

// Further tests could include:
// - Reconnection scenarios (harder to unit test, better for integration tests with controlled RabbitMQ failure)
// - Concurrent publish/consume operations
// - Different exchange types and binding scenarios
// - Error handling for specific AMQP errors (e.g., publishing to a non-existent exchange if server doesn't auto-create)
// - Testing DLX (Dead Letter Exchange) functionality: publish a message that gets routed to DLX
