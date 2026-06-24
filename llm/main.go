// Manual LLM Instrumentation Example (llm.Start)
//
// This example demonstrates how to monitor an *arbitrary* LLM HTTP API with
// WhaTap Go API when no dedicated SDK adapter exists. You write the LLM call
// yourself (any HTTP client) and bracket it with the manual `llm.Start` API.
//
// WHEN TO USE:
//   - Your LLM provider has no WhaTap adapter (anthropic / openai-go /
//     sashabaranov / eino are covered by dedicated adapters — prefer those).
//   - You call a self-hosted or in-house model over plain HTTP.
//
// PATTERN:
//   1. ctx, step := llm.Start(ctx, llm.Config{Provider, Model, OperationType})
//   2. step.AddSystemMessage / step.AddInputMessage   (before the call)
//   3. http.Client{Transport: whataphttp.NewRoundTrip(ctx, base)} — so the
//      HTTPC step in the call chain attaches the pending LLM state and the
//      real endpoint URL is captured (§267).
//   4. step.SetTokens / step.AddOutputMessage          (after parsing)
//   5. defer step.End()                                (no-op when the
//      RoundTripper already closed the step — the common case)
//
// USAGE:
//   OPENAI_API_KEY=sk-... go run main.go
//   # or point at any OpenAI-compatible endpoint:
//   LLM_BASE_URL=https://my-host/v1 OPENAI_API_KEY=... go run main.go

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	whataphttp "github.com/whatap/go-api/instrumentation/net/http/whataphttp"
	"github.com/whatap/go-api/llm"
	"github.com/whatap/go-api/trace"
)

func baseURL() string {
	if v := os.Getenv("LLM_BASE_URL"); v != "" {
		return v
	}
	return "https://api.openai.com/v1"
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY env var required")
	}

	// 1. Initialise WhaTap (reads /usr/whatap/agent/whatap.conf)
	trace.Init(nil)
	defer trace.Shutdown()

	// 2. Open a transaction first — ctx is needed by whataphttp.NewRoundTrip
	txCtx, _ := trace.Start(context.Background(), "llm-manual-example")
	defer trace.End(txCtx, nil)

	if err := chat(txCtx, apiKey); err != nil {
		log.Fatalf("chat failed: %v", err)
	}
}

// chat performs one manually-instrumented LLM call.
func chat(ctx context.Context, apiKey string) error {
	// 3. Register a pending LLM step on ctx.
	ctx, step := llm.Start(ctx, llm.Config{
		Provider:      "openai",
		Model:         "gpt-4o",
		OperationType: "chat",
	})
	defer step.End()

	step.AddSystemMessage("You are a helpful assistant.")
	step.AddInputMessage("Say hello in one word.")

	body, err := json.Marshal(map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful assistant."},
			{"role": "user", "content": "Say hello in one word."},
		},
	})
	if err != nil {
		step.SetError(err, llm.ErrorTypeProgram)
		return err
	}

	// 4. Wrap the transport so the HTTPC step attaches the pending LLM state
	//    and captures the endpoint URL.
	client := &http.Client{Transport: whataphttp.NewRoundTrip(ctx, http.DefaultTransport)}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL()+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		step.SetError(err, llm.ErrorTypeProgram)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		step.SetError(err, llm.ErrorTypeAPI)
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		step.SetError(err, llm.ErrorTypeAPI)
		return err
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		step.SetError(err, llm.ErrorTypeAPI)
		return err
	}

	// 5. Record tokens and the model output.
	step.SetTokens(llm.Tokens{
		Input:  int64(parsed.Usage.PromptTokens),
		Output: int64(parsed.Usage.CompletionTokens),
	})

	content := ""
	if len(parsed.Choices) > 0 {
		content = parsed.Choices[0].Message.Content
	}
	step.AddOutputMessage(content)

	fmt.Println("Response:", content)
	fmt.Printf("Tokens: prompt=%d completion=%d\n",
		parsed.Usage.PromptTokens, parsed.Usage.CompletionTokens)
	return nil
}
