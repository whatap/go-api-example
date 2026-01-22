// IBM Sarama (Kafka) Instrumentation Example
//
// This example demonstrates how to instrument IBM Sarama (Kafka client)
// applications using WhaTap Go API.
//
// INSTRUMENTATION OVERVIEW:
// 1. Create whatapsarama.Interceptor with broker addresses
// 2. Add interceptor to sarama.Config for Producer and Consumer
// 3. Use trace.GetMTrace(ctx) for distributed tracing via message metadata
//
// FEATURES:
// - Producer message tracking (async and sync)
// - Consumer message tracking
// - Distributed tracing across Producer -> Consumer
// - Error tracking for failed messages
//
// DISTRIBUTED TRACING FLOW:
// Producer (Service A)                  Consumer (Service B)
// +------------------+                  +------------------+
// | Send message     |                  | Receive message  |
// | + inject headers | ----> Kafka ---> | + extract headers|
// | txid: 123        |                  | linked txid: 123 |
// +------------------+                  +------------------+

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"text/template"

	"github.com/IBM/sarama"
	"github.com/whatap/go-api/instrumentation/github.com/IBM/sarama/whatapsarama"
	"github.com/whatap/go-api/trace"
)

type HTMLData struct {
	Title   string
	Content string
}

func main() {

	udpPortPtr := flag.Int("up", 6600, "agent port(udp). default 6600")
	portPtr := flag.Int("p", 8080, "web port. default 8080")
	dataSourcePtr := flag.String("ds", "localhost:9092", "kafka broker address")
	setWhatapPtr := flag.Bool("whatap", false, "set whatap")

	flag.Parse()
	port := *portPtr
	udpPort := *udpPortPtr
	dataSource := *dataSourcePtr
	IsWhatap := *setWhatapPtr

	// ============================================================
	// INSTRUMENTATION STEP 1: Initialize WhaTap Agent
	// ============================================================
	if IsWhatap {
		config := make(map[string]string)
		config["net_udp_port"] = fmt.Sprintf("%d", udpPort)
		trace.Init(config)
	}
	defer trace.Shutdown()

	// ============================================================
	// INSTRUMENTATION STEP 2: Configure Sarama with Interceptors
	// ============================================================
	// Create Sarama configuration with WhaTap interceptors.
	// The interceptor handles:
	// - Injecting trace context into message headers (Producer)
	// - Extracting trace context from message headers (Consumer)
	config := sarama.NewConfig()
	config.Producer.Retry.Max = 5
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Consumer.Return.Errors = true

	brokers := []string{dataSource}

	// ============================================================
	// INSTRUMENTATION STEP 3: Create WhaTap Interceptor
	// ============================================================
	// The interceptor needs broker addresses for connection tracking.
	//
	// BEFORE (Original - no instrumentation):
	//   config := sarama.NewConfig()
	//   producer, _ := sarama.NewAsyncProducer(brokers, config)
	//
	// AFTER (Instrumented):
	//   interceptor := whatapsarama.Interceptor{Brokers: brokers}
	//   config.Producer.Interceptors = []sarama.ProducerInterceptor{&interceptor}
	//   config.Consumer.Interceptors = []sarama.ConsumerInterceptor{&interceptor}
	interceptor := whatapsarama.Interceptor{Brokers: brokers}

	// Add interceptors to config
	// - ProducerInterceptor: Called when message is sent
	// - ConsumerInterceptor: Called when message is received
	config.Producer.Interceptors = []sarama.ProducerInterceptor{&interceptor}
	config.Consumer.Interceptors = []sarama.ConsumerInterceptor{&interceptor}

	// ============================================================
	// Create Async Producer
	// ============================================================
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

	// ============================================================
	// Create Sync Producer
	// ============================================================
	syncProducer, err := sarama.NewSyncProducer(brokers, config)

	if err != nil {
		panic(err)
	}
	defer func() {
		if err := syncProducer.Close(); err != nil {
			panic(err)
		}
	}()

	templatePath := "templates/github.com/IBM/index.html"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tp, err := template.ParseFiles(templatePath)
		if err != nil {
			fmt.Println("Template not loaded, ", err)
			return
		}

		data := &HTMLData{}
		data.Title = "IBM Sarama Test Page"
		data.Content = r.RequestURI

		tp.Execute(w, data)
	})

	// ============================================================
	// Case 1: Async Producer
	// ============================================================
	// For async producers, use trace.GetMTrace(ctx) to pass trace context
	// via message Metadata. This enables distributed tracing when the
	// message is consumed by another service.
	//
	// IMPORTANT: The Metadata field carries trace context for async messages.
	// The interceptor's OnSend method extracts this and adds headers.
	http.HandleFunc("/AsyncProduceInput", func(w http.ResponseWriter, r *http.Request) {
		// Start HTTP transaction
		ctx, _ := trace.StartWithRequest(r)
		defer func() {
			trace.End(ctx, nil)
		}()

		// Create message with trace context in Metadata
		// trace.GetMTrace(ctx) returns http.Header with trace headers:
		// - x-whatap-mtid: Multi-transaction ID
		// - x-whatap-traceid: Trace ID
		// - x-whatap-poid: Parent OID
		msg := &sarama.ProducerMessage{
			Topic:    "tmp-topic",
			Key:      sarama.StringEncoder("Data Key"),
			Value:    sarama.StringEncoder("Data Value"),
			Metadata: trace.GetMTrace(ctx), // Pass trace context
		}

		// Send to input channel (async)
		producer.Input() <- msg

	})

	// ============================================================
	// Async Result Handling Goroutine
	// ============================================================
	// This goroutine handles async producer results (success/error).
	// For success messages, we create a new trace linked to the original.
	go func() {
		for {
			select {
			case msg, _ := <-producer.Successes():
				// Create trace for successful message
				name := fmt.Sprintf("produceSuccess/%s", msg.Topic)
				produceCtx, err := trace.Start(context.Background(), name)
				if err != nil {
					return
				}

				// Extract original trace context from Metadata
				header, ok := msg.Metadata.(http.Header)
				if ok != true {
					fmt.Println("Metadata Error")
				}

				// Link this trace to the original HTTP transaction
				if header != nil {
					trace.UpdateMtraceWithContext(produceCtx, header)
				}

				trace.Step(produceCtx, "Async Producer Successes Message", "Success", 2, 2)
				trace.End(produceCtx, nil)

			case err, ok := <-producer.Errors():
				// Handle producer errors
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

	// ============================================================
	// Case 2: Sync Producer
	// ============================================================
	// For sync producers, call interceptor.OnSend(msg) before SendMessage
	// to inject trace headers into the message.
	//
	// NOTE: Unlike async, we need to manually call OnSend for sync producers.
	http.HandleFunc("/SyncProduceInput", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer func() {
			trace.End(ctx, nil)
		}()

		msg := &sarama.ProducerMessage{
			Topic:    "tmp-topic",
			Key:      sarama.StringEncoder("Data Key"),
			Value:    sarama.StringEncoder("Data Value"),
			Metadata: trace.GetMTrace(ctx), // Pass trace context
		}

		// IMPORTANT: Call OnSend before SendMessage for sync producers
		// This injects trace headers into the message
		interceptor.OnSend(msg)

		// Send message synchronously
		_, _, err := syncProducer.SendMessage(msg)

		if err != nil {
			trace.Error(ctx, err)
		}

		trace.Step(ctx, "Sync Producer Success Message", "Success", 2, 2)
	})

	// ============================================================
	// Case 3: Consumer
	// ============================================================
	// The consumer interceptor automatically extracts trace context
	// from message headers, enabling distributed tracing.
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

	// Consumer message handling goroutine
	// The interceptor extracts trace context from message headers
	// and creates linked traces automatically.
	go func() {
		for {
			select {
			case msg := <-consume.Messages():
				// Message received - trace context is automatically extracted
				// by the ConsumerInterceptor
				fmt.Println(msg)
			case consumerError := <-consume.Errors():
				fmt.Println("error", consumerError)
				return
			}
		}

	}()

	fmt.Printf("Server starting on port %d...\n", port)
	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

}
