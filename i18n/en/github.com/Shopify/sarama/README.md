# Sarama(https://github.com/Shopify/sarama)

It traces the kafka produce and consume events processed through the sarama framework.
It configures the interceptor defined in the sarama config.

```
	brokers := []string{"192.168.200.65:9092"} //config kafka broker IP/Port

	interceptor := whatapsarama.Interceptor{Brokers: brokers}

	config.Producer.Interceptors = []sarama.ProducerInterceptor{&interceptor} //Applied only in Async.
	config.Consumer.Interceptors = []sarama.ConsumerInterceptor{&interceptor}

```

# Tracing the async produce

Asynchronous produce events are handled through the Interceptor.
At this time, if the trace context-related information is delivered to ProduceMessage, it connects to Multi Transaction.

```
import (
	"context"
	"net/http"
	"github.com/Shopify/sarama"
	"github.com/whatap/go-api/instrumentation/github.com/Shopify/sarama/whatapsarama"
	"github.com/whatap/go-api/trace"
)

func main() {
	config := sarama.NewConfig()

	brokers := []string{"192.168.200.65:9092"} //config kafka broker IP/Port

	interceptor := whatapsarama.Interceptor{Brokers: brokers}

	config.Producer.Interceptors = []sarama.ProducerInterceptor{&interceptor} //Applied only in Async.
	config.Consumer.Interceptors = []sarama.ConsumerInterceptor{&interceptor}


	whatapConfig := make(map[string]string)

	trace.Init(whatapConfig)
	defer trace.Shutdown()


	producer, err := sarama.NewAsyncProducer(brokers, config)
	consumerOffset := sarama.OffsetOldest
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			panic(err)
		}
	}()

	http.HandleFunc("/AsyncProduceInput", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer func() {
			trace.End(ctx, nil)
		}()
		msg := &sarama.ProducerMessage{
			Topic:    "tmp-topic",
			Key:      sarama.StringEncoder("Data Key"),
			Value:    sarama.StringEncoder("Data Value"),
			Metadata: trace.GetMTrace(ctx),
		}
		producer.Input() <- msg //error check

	})

	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

}

```

# Tracing consume

Consume is traced through the interceptor and it connects to Multi Transaction based on the message defined for Produce.

```

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"text/template"

	"github.com/Shopify/sarama"
	"github.com/whatap/go-api/instrumentation/github.com/Shopify/sarama/whatapsarama"
	"github.com/whatap/go-api/trace"
)



func main() {
	config := sarama.NewConfig()


	brokers := []string{"192.168.200.65:9092"} //config kafka broker IP/Port

	interceptor := whatapsarama.Interceptor{Brokers: brokers}

	config.Producer.Interceptors = []sarama.ProducerInterceptor{&interceptor}
	config.Consumer.Interceptors = []sarama.ConsumerInterceptor{&interceptor}



	// consume 1회당1tx
	consumer, err := sarama.NewConsumer(brokers, config)
	topic := "tmp-topic"

	partitions, _ := consumer.Partitions(topic)
	consume, _ := consumer.ConsumePartition(topic, partitions[0], consumerOffset)

	if consume == nil {
		fmt.Println("consume nil")
		return
	}

	go func() {
		for {
			select {
			case msg := <-consume.Messages():
				fmt.Println(msg)
			case consumerError := <-consume.Errors():
				fmt.Println("error", consumerError)
				return
			}
		}

	}()

}

```
