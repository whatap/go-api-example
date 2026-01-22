# IBM Sarama (Kafka) Instrumentation Example

This example demonstrates how to instrument IBM Sarama (Kafka client) applications using WhaTap Go API.

## Overview

The `whatapsarama` package provides automatic tracing for Kafka operations:

- **Producer Tracing**: Messages sent to Kafka are traced
- **Consumer Tracing**: Messages consumed from Kafka are traced
- **Distributed Tracing**: Trace context is propagated via message headers
- **Error Tracking**: Errors are captured and linked to transactions

## Prerequisites

- Go 1.18+
- Apache Kafka cluster
- WhaTap Go Agent installed

## Installation

```bash
go get github.com/whatap/go-api
go get github.com/whatap/go-api/instrumentation/github.com/IBM/sarama/whatapsarama
go get github.com/IBM/sarama
```

## Quick Start

### 1. Async Producer with Interceptor

**Before (Original Code):**
```go
import "github.com/IBM/sarama"

config := sarama.NewConfig()
producer, err := sarama.NewAsyncProducer(brokers, config)
```

**After (Instrumented Code):**
```go
import (
    "github.com/IBM/sarama"
    "github.com/whatap/go-api/instrumentation/github.com/IBM/sarama/whatapsarama"
)

config := sarama.NewConfig()
// Add WhaTap interceptor for tracing
config.Producer.Interceptors = []sarama.ProducerInterceptor{
    whatapsarama.NewProducerInterceptor(),
}
producer, err := sarama.NewAsyncProducer(brokers, config)
```

### 2. Sync Producer with Interceptor

```go
config := sarama.NewConfig()
config.Producer.Return.Successes = true
config.Producer.Interceptors = []sarama.ProducerInterceptor{
    whatapsarama.NewProducerInterceptor(),
}
producer, err := sarama.NewSyncProducer(brokers, config)
```

### 3. Consumer with Interceptor

```go
config := sarama.NewConfig()
config.Consumer.Interceptors = []sarama.ConsumerInterceptor{
    whatapsarama.NewConsumerInterceptor(),
}
consumer, err := sarama.NewConsumer(brokers, config)
```

## Instrumentation Methods

| Component | WhaTap Interceptor | Description |
|-----------|-------------------|-------------|
| AsyncProducer | `whatapsarama.NewProducerInterceptor()` | Trace async message sending |
| SyncProducer | `whatapsarama.NewProducerInterceptor()` | Trace sync message sending |
| Consumer | `whatapsarama.NewConsumerInterceptor()` | Trace message consumption |

## Examples in This File

| Function | Component | Description |
|----------|-----------|-------------|
| `asyncProduce()` | AsyncProducer | Async message production with goroutine |
| `syncProduce()` | SyncProducer | Sync message production |
| `consume()` | Consumer | Message consumption |

## Key Concepts

### Producer Interceptor

The producer interceptor:
1. Captures the message topic, partition, and key
2. Injects trace context into message headers (for distributed tracing)
3. Records send time and success/failure

```go
config.Producer.Interceptors = []sarama.ProducerInterceptor{
    whatapsarama.NewProducerInterceptor(),
}
```

### Consumer Interceptor

The consumer interceptor:
1. Extracts trace context from message headers
2. Creates a new trace linked to the producer trace
3. Records consumption time

```go
config.Consumer.Interceptors = []sarama.ConsumerInterceptor{
    whatapsarama.NewConsumerInterceptor(),
}
```

### Distributed Tracing Flow

```
Producer (Service A)                  Consumer (Service B)
+------------------+                  +------------------+
| Send message     |                  | Receive message  |
| + inject headers | ----> Kafka ---> | + extract headers|
| txid: 123        |                  | linked txid: 123 |
+------------------+                  +------------------+
```

## Complete Example

```go
func main() {
    trace.Init(nil)
    defer trace.Shutdown()

    brokers := []string{"localhost:9092"}

    // Producer setup
    producerConfig := sarama.NewConfig()
    producerConfig.Producer.Return.Successes = true
    producerConfig.Producer.Interceptors = []sarama.ProducerInterceptor{
        whatapsarama.NewProducerInterceptor(),
    }

    producer, _ := sarama.NewSyncProducer(brokers, producerConfig)
    defer producer.Close()

    // Consumer setup
    consumerConfig := sarama.NewConfig()
    consumerConfig.Consumer.Interceptors = []sarama.ConsumerInterceptor{
        whatapsarama.NewConsumerInterceptor(),
    }

    consumer, _ := sarama.NewConsumer(brokers, consumerConfig)
    defer consumer.Close()

    // HTTP handler that produces a message
    http.HandleFunc("/produce", func(w http.ResponseWriter, r *http.Request) {
        ctx, _ := trace.StartWithRequest(r)
        defer trace.End(ctx, nil)

        msg := &sarama.ProducerMessage{
            Topic: "my-topic",
            Value: sarama.StringEncoder("hello"),
        }

        // Message is traced via interceptor
        partition, offset, err := producer.SendMessage(msg)
        if err != nil {
            trace.Error(ctx, err)
        }
    })
}
```

## Running the Example

```bash
# Start Kafka (using Docker)
docker-compose up -d zookeeper kafka

# Run the example
go run sarama.go -whatap -brokers localhost:9092

# Test producing
curl http://localhost:8080/produce

# Test consuming (in another terminal)
curl http://localhost:8080/consume
```

## IBM vs Shopify Sarama

This example uses IBM Sarama (`github.com/IBM/sarama`). The API is compatible with Shopify Sarama (`github.com/Shopify/sarama`).

| Package | Import Path |
|---------|-------------|
| IBM Sarama | `github.com/IBM/sarama` |
| Shopify Sarama | `github.com/Shopify/sarama` |

WhaTap supports both:
- `whatapsarama` for IBM Sarama
- `whatapsarama` for Shopify Sarama (different import path)

## See Also

- [WhaTap Go API Documentation](https://docs.whatap.io/golang/install-agent)
- [IBM Sarama Documentation](https://github.com/IBM/sarama)
- [Shopify Sarama Example](../../Shopify/sarama/)
