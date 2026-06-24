# sashabaranov/go-openai LLM Instrumentation Example

This example shows how to instrument the [sashabaranov/go-openai](https://github.com/sashabaranov/go-openai) SDK with WhaTap Go API LLM monitoring.

## Overview

`whatapopenai.WrapClient` returns a client that wraps the SDK and emits a WhaTap LLM step for each instrumented call:

- `CreateChatCompletion` / `CreateChatCompletionStream`
- `CreateCompletion` / `CreateCompletionStream`
- `CreateEmbeddings`

Every other method (image / audio / moderation / file / fine-tune / ...) forwards transparently through the embedded `*openai.Client`.

Token counts, finish reason, response content and streaming TTFT are extracted automatically.

## Prerequisites

- Go 1.18+
- WhaTap Go Agent installed (`/usr/whatap/agent/whatap.conf` present)
- `OPENAI_API_KEY` environment variable

## Installation

```bash
go get github.com/sashabaranov/go-openai
go get github.com/whatap/go-api
go get github.com/whatap/go-api/instrumentation/llm
```

## Quick Start

### Manual (this example)

```go
import (
    openai "github.com/sashabaranov/go-openai"
    whatapopenai "github.com/whatap/go-api/instrumentation/llm/github.com/sashabaranov/go-openai/whatapopenai"
    whataphttp "github.com/whatap/go-api/instrumentation/net/http/whataphttp"
    "github.com/whatap/go-api/trace"
)

trace.Init(nil)
ctx, _ := trace.Start(context.Background(), "llm-example")

cfg := openai.DefaultConfig(apiKey)
cfg.HTTPClient = &http.Client{Transport: whataphttp.NewRoundTrip(ctx, http.DefaultTransport)}
client := whatapopenai.WrapClient(openai.NewClientWithConfig(cfg))

resp, _ := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{...})
trace.End(ctx, nil)
```

### Auto-inject (whatap-go-inst)

```bash
whatap-go-inst go build .
```

The instrumentation tool rewrites the SDK calls automatically — no manual `WrapClient` / transport setup needed.

## Run

```bash
export OPENAI_API_KEY=sk-...
go run main.go
```

## What gets captured

Per call the agent emits an LLM step containing:

- Provider (`openai`) and model (`gpt-3.5-turbo`, `gpt-4`, ...)
- Endpoint URL + HTTP status
- Token usage (prompt / completion / total)
- Finish reason
- TTFT for streaming calls

## References

- Adapter source: `go-api/instrumentation/llm/github.com/sashabaranov/go-openai/whatapopenai/`
