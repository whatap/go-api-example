// anthropics/anthropic-sdk-go LLM Instrumentation Example
//
// This example demonstrates how to instrument the official Anthropic Go SDK
// (github.com/anthropics/anthropic-sdk-go) using WhaTap Go API.
//
// SDK SHAPE:
//   The Anthropic SDK exposes the messages API as a service field on the
//   Client: `client.Messages.New(ctx, params)`. The interesting method lives
//   on *MessageService, not *Client, so the adapter wraps at the service
//   level (§253 §2.5).
//
// MANUAL USAGE (this example):
//   Replace `client.Messages.New(ctx, params)` with
//   `whatapanthropic.WrapAndNewMessage(ctx, client.Messages, params)`.
//   Same signature, no variable rebinding.
//
// AUTO-INJECT:
//   `whatap-go-inst go build` rewrites the call automatically — no manual
//   change needed.
//
// USAGE:
//   ANTHROPIC_API_KEY=sk-ant-... go run main.go

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	whataphttp "github.com/whatap/go-api/instrumentation/net/http/whataphttp"
	whatapanthropic "github.com/whatap/go-api/instrumentation/llm/github.com/anthropics/anthropic-sdk-go/whatapanthropic"
	"github.com/whatap/go-api/trace"
)

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY env var required")
	}

	// 1. Initialise WhaTap (reads /usr/whatap/agent/whatap.conf)
	trace.Init(nil)
	defer trace.Shutdown()

	// 2. Open a transaction first — ctx is needed by whataphttp.NewRoundTrip
	ctx, _ := trace.Start(context.Background(), "llm-example")
	defer trace.End(ctx, nil)

	// 3. Build SDK client with WhaTap HTTP transport (captures URL + status)
	httpClient := &http.Client{
		Transport: whataphttp.NewRoundTrip(ctx, http.DefaultTransport),
	}
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
		option.WithHTTPClient(httpClient),
	)

	// 4. Wrap the call — same params as client.Messages.New(ctx, params)
	resp, err := whatapanthropic.WrapAndNewMessage(ctx, client.Messages, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5HaikuLatest,
		MaxTokens: 64,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Say hello in one word.")),
		},
	})
	if err != nil {
		log.Fatalf("Anthropic call failed: %v", err)
	}
	if len(resp.Content) > 0 {
		fmt.Println("Response:", resp.Content[0].Text)
	}
	fmt.Printf("Tokens: input=%d output=%d\n",
		resp.Usage.InputTokens, resp.Usage.OutputTokens)
}
