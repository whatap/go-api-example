package main

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

type HTMLData struct {
	Title   string
	Content string
	//HTMLContent template.HTML
}

func main() {

	udpPortPtr := flag.Int("up", 6600, "agent port(udp). defalt 6600 ")
	portPtr := flag.Int("p", 8080, "web port. default 8080  ")
	dataSourcePtr := flag.String("ds", "phpdemo3:9092", " dataSourceName ")
	setWhatapPtr := flag.Bool("whatap", false, "set whatap")

	flag.Parse()
	port := *portPtr
	udpPort := *udpPortPtr
	dataSource := *dataSourcePtr
	IsWhatap := *setWhatapPtr

	if IsWhatap {
		config := make(map[string]string)
		config["net_udp_port"] = fmt.Sprintf("%d", udpPort)
		trace.Init(config)
	}
	defer trace.Shutdown()

	config := sarama.NewConfig()
	config.Producer.Retry.Max = 5
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Consumer.Return.Errors = true

	brokers := []string{dataSource} //config kafka broker IP/Port

	interceptor := whatapsarama.Interceptor{Brokers: brokers}

	config.Producer.Interceptors = []sarama.ProducerInterceptor{&interceptor} //Async에만 적용됨
	config.Consumer.Interceptors = []sarama.ConsumerInterceptor{&interceptor}

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

	syncProducer, err := sarama.NewSyncProducer(brokers, config)

	if err != nil {
		panic(err)
	}
	defer func() {
		if err := syncProducer.Close(); err != nil {
			panic(err)
		}
	}()

	templatePath := "templates/github.com/Shopify/index.html"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tp, err := template.ParseFiles(templatePath)
		if err != nil {
			fmt.Println("Template not loaded, ", err)
			return
		}

		data := &HTMLData{}
		data.Title = "Sarama Test Page"
		data.Content = r.RequestURI

		tp.Execute(w, data)
	})

	//Case 1. ASyncProduce
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

	// Async Result 처리 루틴
	go func() {
		for {
			select {
			case msg, _ := <-producer.Successes():
				name := fmt.Sprintf("produceSuccess/%s", msg.Topic)
				produceCtx, err := trace.Start(context.Background(), name)
				if err != nil {
					return
				}

				header, ok := msg.Metadata.(http.Header)
				if ok != true {
					fmt.Println("Metadata Error")
				}
				if header != nil {
					trace.UpdateMtraceWithContext(produceCtx, header)
				}
				trace.Step(produceCtx, "Async Producer Successes Message", "Success", 2, 2)
				trace.End(produceCtx, nil)

			case err, ok := <-producer.Errors():
				if ok {
					ctx, ok := err.Msg.Metadata.(context.Context)
					if ok != true {
						fmt.Println("Metadata Error")
					}
					if ctx != nil {
						errMsg := fmt.Sprintf("Error : %s", err)
						trace.Step(ctx, "Async Producer Error Message", errMsg, 2, 2)
						trace.End(ctx, err)
					}
				}
			}
		}
	}()

	//Case 2. SyncProduce
	http.HandleFunc("/SyncProduceInput", func(w http.ResponseWriter, r *http.Request) {
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

		interceptor.OnSend(msg)
		_, _, err := syncProducer.SendMessage(msg)

		if err != nil {
			trace.Error(ctx, err)
		}

		trace.Step(ctx, "Sync Producer Success Message", "Success", 2, 2)
	})

	// consume 1회당1tx
	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		fmt.Println("error new consumer ", err)
	}
	topic := "tmp-topic"

	partitions, err := consumer.Partitions(topic)
	if err != nil {
		fmt.Println("error consumer partitions ", err)
	}
	consume, err := consumer.ConsumePartition(topic, partitions[0], consumerOffset)
	if err != nil {
		fmt.Println("error consumer ConsumePartition ", err)
	}

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

	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

}
