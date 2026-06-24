# openai/openai-go (Official OpenAI Go SDK) LLM Instrumentation Example

This example shows how to instrument the [official OpenAI Go SDK](https://github.com/openai/openai-go) with WhaTap Go API LLM monitoring.

> The unofficial community SDK [sashabaranov/go-openai](../sashabaranov/go-openai) has a separate adapter and example.

## Overview

The official SDK exposes the chat completions API as a 3-step selector chain:

```go
client.Chat.Completions.New(ctx, params)
```

Because the interesting method lives on `*ChatCompletionService`, the adapter wraps at the service level:

```go
whatapopenaigo.WrapAndNewChatCompletion(ctx, client.Chat.Completions, params)
```

Same signature, no variable rebinding.

Captured automatically:

- Token counts (prompt / completion / total)
- Finish reason
- Message content
- Streaming TTFT via the SDK's `ssestream` package (accumulates `ChatCompletionChunk` deltas)

## Prerequisites

- Go 1.18+
- WhaTap Go Agent installed (`/usr/whatap/agent/whatap.conf` present)
- `OPENAI_API_KEY` environment variable

## Installation

```bash
go get github.com/openai/openai-go
go get github.com/whatap/go-api
go get github.com/whatap/go-api/instrumentation/llm
```

## Quick Start

### Manual (this example)

```go
import (
    "github.com/openai/openai-go"
    "github.com/openai/openai-go/option"
    whatapopenaigo "github.com/whatap/go-api/instrumentation/llm/github.com/openai/openai-go/whatapopenaigo"
    whataphttp "github.com/whatap/go-api/instrumentation/net/http/whataphttp"
    "github.com/whatap/go-api/trace"
)

trace.Init(nil)
ctx, _ := trace.Start(context.Background(), "llm-example")

client := openai.NewClient(
    option.WithAPIKey(apiKey),
    option.WithHTTPClient(&http.Client{Transport: whataphttp.NewRoundTrip(ctx, http.DefaultTransport)}),
)

resp, _ := whatapopenaigo.WrapAndNewChatCompletion(ctx, client.Chat.Completions, openai.ChatCompletionNewParams{...})
trace.End(ctx, nil)
```

### Auto-inject (whatap-go-inst)

```bash
whatap-go-inst go build .
```

The tool rewrites the 3-step selector call automatically.

## Run

```bash
export OPENAI_API_KEY=sk-...
go run main.go
```

## References

- Adapter source: `go-api/instrumentation/llm/github.com/openai/openai-go/whatapopenaigo/`
