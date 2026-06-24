// sashabaranov/go-openai LLM Instrumentation Example
//
// This example demonstrates how to instrument the sashabaranov/go-openai
// SDK using WhaTap Go API.
//
// INSTRUMENTATION OVERVIEW:
//  1. Call trace.Init() at startup (whatap.conf must be present)
//  2. Wrap *openai.Client with whatapopenai.WrapClient — the returned value
//     supports the same API but emits WhaTap LLM steps for the instrumented
//     methods (CreateChatCompletion / CreateChatCompletionStream /
//     CreateCompletion / CreateCompletionStream / CreateEmbeddings).
//  3. Wrap the HTTP transport with whataphttp.NewRoundTrip so the real
//     endpoint URL + status code are captured by the LLM step.
//
// AUTO-INJECT (whatap-go-inst):
//   The whatap-go-inst tool can perform steps 2 + 3 automatically — no code
//   change required. Run `whatap-go-inst go build` instead of `go build`.
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

	openai "github.com/sashabaranov/go-openai"

	whataphttp "github.com/whatap/go-api/instrumentation/net/http/whataphttp"
	whatapopenai "github.com/whatap/go-api/instrumentation/llm/github.com/sashabaranov/go-openai/whatapopenai"
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
	cfg := openai.DefaultConfig(apiKey)
	cfg.HTTPClient = &http.Client{
		Transport: whataphttp.NewRoundTrip(ctx, http.DefaultTransport),
	}
	inner := openai.NewClientWithConfig(cfg)

	// 4. Wrap client so LLM-shaped methods emit WhaTap LLM steps
	client := whatapopenai.WrapClient(inner)

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: "Say hello in one word."},
		},
	})
	if err != nil {
		log.Fatalf("OpenAI call failed: %v", err)
	}
	fmt.Println("Response:", resp.Choices[0].Message.Content)
	fmt.Printf("Tokens: prompt=%d completion=%d total=%d\n",
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
}
