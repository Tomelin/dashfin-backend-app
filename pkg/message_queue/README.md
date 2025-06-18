# Message Queue Library

Biblioteca Go para conexão e gerenciamento de RabbitMQ com suporte a múltiplas exchanges, queues, routing keys e dead letter queues via configuração YAML.

## Características

-  configuração via YAML
-  Múltiplas exchanges e queues
-  Dead Letter Queue suporte
-  Consumer com context cancellation
-  Publisher com configuraçães flexíveis
-  Reconnection handling
-  Error handling e logging
-  Interface clean e extensível

## Instalação

```bash
go get github.com/streadway/amqp
```

## configuração YAML

### Estrutura Básica

```yaml
# config.yaml
url: "amqp://guest:guest@localhost:5672/"

# configuraçães padrão para consumer e publisher
default_consumer:
  tag: "default-consumer"
  auto_ack: false
  exclusive: false
  no_local: false
  no_wait: false
  args: {}

default_publisher:
  mandatory: false
  immediate: false

# Lista de exchanges e suas queues
message_queues:
  - exchange: "orders_exchange"
    type: "direct"  # direct, topic, fanout, headers
    durable: true
    queues:
      - name: "order_processing_queue"
        durable: true
        auto_delete: false
        exclusive: false
        no_wait: false
        route_key: "order.process"
        args:
          x-message-ttl: 300000  # 5 minutos
        dead_letter:
          exchange: "orders_dlx"
          queue: "order_processing_dlq"
          route_key: "order.failed"
        consumer:
          tag: "order-processor"
          auto_ack: false
        publisher:
          mandatory: true
          immediate: false

      - name: "order_notification_queue"
        durable: true
        route_key: "order.notify"
        dead_letter:
          exchange: "orders_dlx"
          queue: "order_notification_dlq"
          route_key: "notify.failed"

  - exchange: "notifications_exchange"
    type: "topic"
    durable: true
    queues:
      - name: "email_queue"
        durable: true
        route_key: "notification.email.*"
        dead_letter:
          exchange: "notifications_dlx"
          queue: "email_dlq"
          route_key: "email.failed"
      
      - name: "sms_queue"
        durable: true
        route_key: "notification.sms.*"
      
      # Exemplo com múltiplas route keys
      - name: "financial_reporting_queue"
        durable: true
        route_keys:  # Múltiplas route keys
          - "finance.income.report"
          - "finance.expense.report"
          - "finance.budget.report"
        dead_letter:
          exchange: "notifications_dlx"
          queue: "financial_reports_dlq"
          route_key: "report.failed"
```

### configuração Mínima

```yaml
url: "amqp://localhost:5672"
message_queues:
  - exchange: "simple_exchange"
    queues:
      - name: "simple_queue"
        route_key: "simple.route"
```

## Uso da Biblioteca

### 1. Inicialização

```go
package main

import (
    "context"
    "log"
    "time"
    
    "your-project/pkg/message_queue"
    "gopkg.in/yaml.v2"
    "io/ioutil"
)

func main() {
    // Carregar configuração do YAML
    config, err := loadConfig("config.yaml")
    if err != nil {
        log.Fatal("Error loading config:", err)
    }

    // Criar instíncia do RabbitMQ
    mq, err := message_queue.NewRabbitMQ(config)
    if err != nil {
        log.Fatal("Error creating RabbitMQ:", err)
    }
    defer mq.Close()

    // Configurar exchanges e queues
    err = mq.Setup()
    if err != nil {
        log.Fatal("Error setting up RabbitMQ:", err)
    }

    log.Println("RabbitMQ initialized successfully")
}

func loadConfig(filename string) (message_queue.Config, error) {
    var config message_queue.Config
    
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return config, err
    }
    
    err = yaml.Unmarshal(data, &config)
    return config, err
}
```

### 2. Publisher - Enviando Mensagens

```go
func publishExample(mq message_queue.MessageQueue) {
    // Mensagem para processamento de pedido
    orderData := map[string]interface{}{
        "order_id": "12345",
        "customer_id": "67890",
        "amount": 99.99,
        "timestamp": time.Now(),
    }
    
    message, _ := json.Marshal(orderData)
    
    err := mq.Publisher("orders_exchange", "order_processing_queue", message)
    if err != nil {
        log.Printf("Error publishing message: %v", err)
        return
    }
    
    log.Println("Order message published successfully")
    
    // Mensagem de notificação por email
    emailData := map[string]interface{}{
        "to": "customer@example.com",
        "subject": "Order Confirmation",
        "body": "Your order has been confirmed",
    }
    
    emailMessage, _ := json.Marshal(emailData)
    
    err = mq.Publisher("notifications_exchange", "email_queue", emailMessage)
    if err != nil {
        log.Printf("Error publishing email: %v", err)
        return
    }
    
    log.Println("Email notification published successfully")
}
```

### 3. Consumer - Consumindo Mensagens

```go
func consumeOrders(mq message_queue.MessageQueue) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Handler para processar pedidos
    orderHandler := func(body []byte) error {
        var order map[string]interface{}
        
        err := json.Unmarshal(body, &order)
        if err != nil {
            log.Printf("Error unmarshaling order: %v", err)
            return err // Mensagem vai para dead letter queue
        }
        
        log.Printf("Processing order: %+v", order)
        
        // Simular processamento
        time.Sleep(100 * time.Millisecond)
        
        // Simular falha ocasional (para teste de dead letter)
        if order["order_id"] == "failed_order" {
            return fmt.Errorf("simulated processing failure")
        }
        
        log.Printf("Order %s processed successfully", order["order_id"])
        return nil
    }
    
    // Iniciar consumer
    err := mq.Consumer(ctx, "orders_exchange", "order_processing_queue", orderHandler)
    if err != nil {
        log.Printf("Consumer error: %v", err)
    }
}
```

### 4. Consumer com Goroutines

```go
func startConsumers(mq message_queue.MessageQueue) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Consumer para pedidos
    go func() {
        orderHandler := func(body []byte) error {
            // Processar pedido
            log.Printf("Processing order: %s", string(body))
            return nil
        }
        
        err := mq.Consumer(ctx, "orders_exchange", "order_processing_queue", orderHandler)
        if err != nil {
            log.Printf("Order consumer error: %v", err)
        }
    }()
    
    // Consumer para emails
    go func() {
        emailHandler := func(body []byte) error {
            // Enviar email
            log.Printf("Sending email: %s", string(body))
            return nil
        }
        
        err := mq.Consumer(ctx, "notifications_exchange", "email_queue", emailHandler)
        if err != nil {
            log.Printf("Email consumer error: %v", err)
        }
    }()
    
    // Consumer para SMS
    go func() {
        smsHandler := func(body []byte) error {
            // Enviar SMS
            log.Printf("Sending SMS: %s", string(body))
            return nil
        }
        
        err := mq.Consumer(ctx, "notifications_exchange", "sms_queue", smsHandler)
        if err != nil {
            log.Printf("SMS consumer error: %v", err)
        }
    }()
    
    // Aguardar sinal para parar
    select {}
}
```

### 5. Exemplo Completo com Graceful Shutdown

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"
    
    "your-project/pkg/message_queue"
)

func main() {
    // Carregar configuração
    config, err := loadConfig("config.yaml")
    if err != nil {
        log.Fatal("Error loading config:", err)
    }

    // Inicializar RabbitMQ
    mq, err := message_queue.NewRabbitMQ(config)
    if err != nil {
        log.Fatal("Error creating RabbitMQ:", err)
    }
    defer mq.Close()

    err = mq.Setup()
    if err != nil {
        log.Fatal("Error setting up RabbitMQ:", err)
    }

    // Context para cancelamento
    ctx, cancel := context.WithCancel(context.Background())
    var wg sync.WaitGroup

    // Iniciar consumers
    startAllConsumers(ctx, mq, &wg)
    
    // Simular publicação de mensagens
    go publishMessages(mq)

    // Aguardar sinal de interrupção
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    <-sigChan
    log.Println("Shutdown signal received")
    
    // Cancelar consumers
    cancel()
    
    // Aguardar consumers terminarem
    wg.Wait()
    log.Println("All consumers stopped")
}

func startAllConsumers(ctx context.Context, mq message_queue.MessageQueue, wg *sync.WaitGroup) {
    consumers := map[string]map[string]func([]byte) error{
        "orders_exchange": {
            "order_processing_queue": func(body []byte) error {
                log.Printf("Processing order: %s", string(body))
                time.Sleep(100 * time.Millisecond)
                return nil
            },
        },
        "notifications_exchange": {
            "email_queue": func(body []byte) error {
                log.Printf("Sending email: %s", string(body))
                time.Sleep(50 * time.Millisecond)
                return nil
            },
            "sms_queue": func(body []byte) error {
                log.Printf("Sending SMS: %s", string(body))
                time.Sleep(30 * time.Millisecond)
                return nil
            },
        },
    }
    
    for exchangeName, queues := range consumers {
        for queueName, handler := range queues {
            wg.Add(1)
            go func(ex, q string, h func([]byte) error) {
                defer wg.Done()
                err := mq.Consumer(ctx, ex, q, h)
                if err != nil && err != context.Canceled {
                    log.Printf("Consumer %s/%s error: %v", ex, q, err)
                }
            }(exchangeName, queueName, handler)
        }
    }
}

func publishMessages(mq message_queue.MessageQueue) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    counter := 0
    for range ticker.C {
        counter++
        
        // Publicar pedido
        order := map[string]interface{}{
            "order_id": fmt.Sprintf("order_%d", counter),
            "amount": 100.0 + float64(counter),
        }
        orderData, _ := json.Marshal(order)
        mq.Publisher("orders_exchange", "order_processing_queue", orderData)
        
        // Publicar notificação
        email := map[string]interface{}{
            "to": "user@example.com",
            "message": fmt.Sprintf("Order %d created", counter),
        }
        emailData, _ := json.Marshal(email)
        mq.Publisher("notifications_exchange", "email_queue", emailData)
        
        if counter >= 10 {
            break
        }
    }
}
```

## Interface da Biblioteca

```go
type MessageQueue interface {
    Consumer(ctx context.Context, exchangeName, queueName string, handler func([]byte) error) error
    Publisher(exchangeName, queueName string, message []byte) error
    PublisherWithRouteKey(exchangeName, routeKey string, message []byte) error
    Close() error
    Setup() error
}
```

### Métodos

- **`Setup()`**: Configura todas as exchanges, queues e dead letter queues definidas no YAML
- **`Publisher(exchangeName, queueName, message)`**: Publica mensagem na queue especificada (usa primeira route key)
- **`PublisherWithRouteKey(exchangeName, routeKey, message)`**: Publica mensagem com route key específica
- **`Consumer(ctx, exchangeName, queueName, handler)`**: Consome mensagens com handler customizado
- **`Close()`**: Fecha conexão com RabbitMQ

## Múltiplas Route Keys

Você pode configurar uma queue para receber mensagens de múltiplas route keys:

### Configuração YAML

```yaml
message_queues:
  - exchange: "finance_exchange"
    type: "topic"
    queues:
      # Queue com uma route key (método tradicional)
      - name: "income_queue"
        route_key: "finance.income.*"
      
      # Queue com múltiplas route keys
      - name: "financial_reporting_queue"
        route_keys:
          - "finance.income.report"
          - "finance.expense.report"
          - "finance.budget.analyze"
          - "finance.tax.calculate"
      
      # Queue que captura tudo relacionado a aprovações
      - name: "approval_queue"
        route_keys:
          - "finance.expense.approval"
          - "finance.budget.approval"
          - "finance.investment.approval"
```

### Exemplo de Uso

```go
// Configuração
config, _ := loadConfig("config.yaml")
mq, _ := message_queue.NewRabbitMQ(config)
mq.Setup()

// Publicar usando queue específica (usa primeira route key)
mq.Publisher("finance_exchange", "financial_reporting_queue", reportData)

// Publicar usando route key específica
mq.PublisherWithRouteKey("finance_exchange", "finance.income.report", incomeReportData)
mq.PublisherWithRouteKey("finance_exchange", "finance.expense.report", expenseReportData)
mq.PublisherWithRouteKey("finance_exchange", "finance.budget.analyze", budgetData)

// Todas essas mensagens serão entregues na "financial_reporting_queue"
```

### Compatibilidade

A biblioteca mantém compatibilidade com configurações antigas:

```yaml
# Método antigo (ainda funciona)
- name: "old_queue"
  route_key: "finance.income"

# Método novo
- name: "new_queue"
  route_keys:
    - "finance.income"
    - "finance.expense"
```

## Dead Letter Queue

As mensagens que falharam no processamento são automaticamente enviadas para a dead letter queue configurada:

```yaml
dead_letter:
  exchange: "orders_dlx"
  queue: "order_processing_dlq" 
  route_key: "order.failed"
```

## Error Handling

A biblioteca trata automaticamente:
- Reconnection em caso de queda de conexão
- Nack de mensagens com erro (envio para DLQ)
- Ack de mensagens processadas com sucesso
- Context cancellation para graceful shutdown

## Logs

A biblioteca faz log de:
- conexães fechadas
- Erros de processamento de mensagens
- Falhas de conexão

## Dependências

```go
import "github.com/streadway/amqp"
```

## Contribuição

Para contribuir com melhorias ou correções, abra uma issue ou pull request.