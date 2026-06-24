// openai/openai-go (Official OpenAI Go SDK) LLM Instrumentation Example
//
// This example demonstrates how to instrument the official OpenAI Go SDK
// (github.com/openai/openai-go) using WhaTap Go API.
//
// SDK SHAPE:
//   The OpenAI SDK exposes the chat completions API as a nested service
//   field on the Client: `client.Chat.Completions.New(ctx, params)` — a
//   3-step selector chain. Because the interesting method lives on
//   *ChatCompletionService (not *Client), the adapter wraps at the service
//   level (§255).
//
// MANUAL USAGE (this example):
//   Replace `client.Chat.Completions.New(ctx, params)` with
//   `whatapopenaigo.WrapAndNewChatCompletion(ctx, client.Chat.Completions, params)`.
//   Same signature, no variable rebinding.
//
// AUTO-INJECT:
//   `whatap-go-inst go build` rewrites the call automatically.
//
// USAGE:
//   OPENAI_API_KEY=sk-... go run main.go

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"

	whataphttp "github.com/whatap/go-api/instrumentation/net/http/whataphttp"
	whatapopenaigo "github.com/whatap/go-api/instrumentation/llm/github.com/openai/openai-go/whatapopenaigo"
	"github.com/whatap/go-api/trace"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY env var required")
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
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithHTTPClient(httpClient),
	)

	// 4. Wrap the call — same params as client.Chat.Completions.New(ctx, params)
	resp, err := whatapopenaigo.WrapAndNewChatCompletion(ctx, client.Chat.Completions, openai.ChatCompletionNewParams{
		Model: shared.ChatModelGPT3_5Turbo,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Say hello in one word."),
		},
	})
	if err != nil {
		log.Fatalf("OpenAI call failed: %v", err)
	}
	if len(resp.Choices) > 0 {
		fmt.Println("Response:", resp.Choices[0].Message.Content)
	}
	fmt.Printf("Tokens: prompt=%d completion=%d total=%d\n",
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
}
