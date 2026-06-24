// cloudwego/eino LLM Instrumentation Example
//
// This example demonstrates how to instrument eino ChatModel calls using
// WhaTap Go API.
//
// SDK SHAPE:
//   eino (cloudwego/eino) defines the ChatModel interface and eino-ext
//   provides concrete implementations (OpenAI, Claude, etc.). The adapter
//   is a transparent decorator that wraps the ChatModel:
//
//     wrapped := whatapeino.WrapChatModel(inner)            // model.ChatModel
//     wrapped := whatapeino.WrapToolCallingChatModel(inner) // newer interface
//
//   Every Generate / Stream call emits a WhaTap LLM step.
//
// URL CAPTURE:
//   The real HTTP endpoint URL is captured by the wrapped RoundTripper
//   inside the SDK (§267). For this to work, the SDK's HTTP transport must
//   be wrapped with whataphttp.NewRoundTrip — typically via §254 auto-inject,
//   or by supplying a custom http.Client in the SDK config.
//
// AUTO-INJECT:
//   `whatap-go-inst go build` instruments eino-ext automatically (§254/§282):
//   it replaces the constructor (NewChatModel) to wrap the SDK's HTTP transport
//   for URL capture, wraps compose call sites — Chain.AppendChatModel etc. —
//   with WrapBaseChatModel, and rewrites direct cm.Generate / cm.Stream calls
//   to WrapGenerate / WrapStream for response-metadata extraction. The manual
//   WrapChatModel below stays available for explicit use.
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

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	whataphttp "github.com/whatap/go-api/instrumentation/net/http/whataphttp"
	whatapeino "github.com/whatap/go-api/instrumentation/llm/github.com/cloudwego/eino/whatapeino"
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
	txCtx, _ := trace.Start(context.Background(), "llm-example")
	defer trace.End(txCtx, nil)

	// 3. Build the underlying ChatModel (eino-ext OpenAI provider) with
	//    WhaTap HTTP transport so the URL is captured by the LLM step.
	httpClient := &http.Client{
		Transport: whataphttp.NewRoundTrip(txCtx, http.DefaultTransport),
	}
	inner, err := openai.NewChatModel(txCtx, &openai.ChatModelConfig{
		APIKey:     apiKey,
		Model:      "gpt-3.5-turbo",
		HTTPClient: httpClient,
	})
	if err != nil {
		log.Fatalf("create ChatModel: %v", err)
	}

	// 4. Wrap so Generate / Stream emit WhaTap LLM steps
	var cm model.ChatModel = whatapeino.WrapChatModel(inner)

	resp, err := cm.Generate(txCtx, []*schema.Message{
		{Role: schema.User, Content: "Say hello in one word."},
	})
	if err != nil {
		log.Fatalf("Generate failed: %v", err)
	}
	fmt.Println("Response:", resp.Content)
	if resp.ResponseMeta != nil && resp.ResponseMeta.Usage != nil {
		fmt.Printf("Tokens: prompt=%d completion=%d total=%d\n",
			resp.ResponseMeta.Usage.PromptTokens,
			resp.ResponseMeta.Usage.CompletionTokens,
			resp.ResponseMeta.Usage.TotalTokens)
	}
}
