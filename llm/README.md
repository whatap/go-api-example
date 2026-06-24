# Manual LLM Instrumentation Example (`llm.Start`)

This example shows how to monitor an **arbitrary** LLM HTTP API with the WhaTap Go API when no dedicated SDK adapter exists.

## When to use

- Your LLM provider has **no** WhaTap adapter. (anthropic / openai-go / sashabaranov / eino are covered by dedicated adapters — prefer those, see the sibling examples.)
- You call a self-hosted or in-house model over plain HTTP.

The manual API lives in the **base** `github.com/whatap/go-api` module — no nested `instrumentation/llm` module is needed.

## Pattern

```go
ctx, step := llm.Start(ctx, llm.Config{
    Provider:      "openai",
    Model:         "gpt-4o",
    OperationType: "chat",
})
defer step.End()

step.AddSystemMessage("You are a helpful assistant.")
step.AddInputMessage("Say hello in one word.")

// Wrap the transport so the HTTPC step attaches the pending LLM state
// and the real endpoint URL is captured.
client := &http.Client{Transport: whataphttp.NewRoundTrip(ctx, http.DefaultTransport)}
resp, _ := client.Do(req)

// after parsing the response
step.SetTokens(llm.Tokens{Input: promptTokens, Output: completionTokens})
step.AddOutputMessage(content)
```

`step.End()` is a no-op when the wrapped `RoundTripper` already closed the HTTPC step (the common case); it is kept for symmetry with other Go resource patterns.

## Prerequisites

- Go 1.23+
- WhaTap Go Agent installed (`/usr/whatap/agent/whatap.conf` present)
- `OPENAI_API_KEY` environment variable

## Run

```bash
OPENAI_API_KEY=sk-... go run main.go

# point at any OpenAI-compatible endpoint:
LLM_BASE_URL=https://my-host/v1 OPENAI_API_KEY=... go run main.go
```

## Captured fields

Token counts (input/output), response content, the LLM step under the current
transaction, and the real endpoint URL (via `whataphttp.NewRoundTrip`). Errors
are recorded with `step.SetError(err, llm.ErrorType...)`.
