# Sarama(https://github.com/Shopify/sarama)

sarama 프레임워크를 통해서 처리되는 kafka produce, consume 이벤트를 추적합니다.
sarama config에 정의된 interceptor를 설정합니다.

```
	brokers := []string{"192.168.200.65:9092"} //config kafka broker IP/Port

	interceptor := whatapsarama.Interceptor{Brokers: brokers}

	config.Producer.Interceptors = []sarama.ProducerInterceptor{&interceptor} //Async에만 적용됨
	config.Consumer.Interceptors = []sarama.ConsumerInterceptor{&interceptor}

```

# async produce 추적

비동기로 처리되는 produce에 대해서는 Interceptor를 통해 처리됩니다.
이때 ProduceMessage에 Trace Context 관련 정보를 주면 Multi Transaction으로 연결됩니다.

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

	config.Producer.Interceptors = []sarama.ProducerInterceptor{&interceptor} //Async에만 적용됨
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

# consume 추적

consume도 Interceptor를 통해 추적되며 Produce 당시 정의된 Message 기준으로 Multi Transaction으로 연결됩니다.

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
