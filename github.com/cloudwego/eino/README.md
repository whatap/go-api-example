# cloudwego/eino LLM Instrumentation Example

This example shows how to instrument [eino](https://github.com/cloudwego/eino) ChatModel calls with WhaTap Go API LLM monitoring.

## Overview

eino defines the `model.ChatModel` interface (and the newer `model.ToolCallingChatModel`). [eino-ext](https://github.com/cloudwego/eino-ext) provides concrete implementations for OpenAI, Claude, Ollama, etc.

The WhaTap adapter is a transparent **decorator**:

```go
wrapped := whatapeino.WrapChatModel(inner)             // model.ChatModel
wrapped := whatapeino.WrapToolCallingChatModel(inner)  // newer interface
```

Every `Generate` / `Stream` call emits a WhaTap LLM step. Wrapping the same instance twice returns the existing wrapper (idempotent).

`BindTools` / `WithTools` are forwarded transparently — derived models stay wrapped.

## URL capture

The real HTTP endpoint URL is captured by the wrapped RoundTripper **inside the SDK**. For this to work, the SDK's HTTP transport must be wrapped with `whataphttp.NewRoundTrip(ctx, http.DefaultTransport)` — either:

- **Manual**: pass a custom `*http.Client` to the SDK config (this example).
- **Auto-inject**: `whatap-go-inst go build` does this automatically.

## Prerequisites

- Go 1.23+
- WhaTap Go Agent installed (`/usr/whatap/agent/whatap.conf` present)
- `OPENAI_API_KEY` environment variable (this example uses eino-ext OpenAI provider)

## Installation

```bash
go get github.com/cloudwego/eino
go get github.com/cloudwego/eino-ext/components/model/openai
go get github.com/whatap/go-api
go get github.com/whatap/go-api/instrumentation/llm
```

## Quick Start

### Manual (this example)

```go
import (
    "github.com/cloudwego/eino-ext/components/model/openai"
    "github.com/cloudwego/eino/components/model"
    whatapeino "github.com/whatap/go-api/instrumentation/llm/github.com/cloudwego/eino/whatapeino"
    whataphttp "github.com/whatap/go-api/instrumentation/net/http/whataphttp"
    "github.com/whatap/go-api/trace"
)

trace.Init(nil)
ctx, _ := trace.Start(context.Background(), "llm-example")

inner, _ := openai.NewChatModel(ctx, &openai.ChatModelConfig{
    APIKey:     apiKey,
    Model:      "gpt-3.5-turbo",
    HTTPClient: &http.Client{Transport: whataphttp.NewRoundTrip(ctx, http.DefaultTransport)},
})
var cm model.ChatModel = whatapeino.WrapChatModel(inner)

resp, _ := cm.Generate(ctx, []*schema.Message{...})
trace.End(ctx, nil)
```

### Auto-inject (whatap-go-inst)

```bash
whatap-go-inst go build .
```

The tool instruments eino-ext automatically:

- replaces `NewChatModel` to wrap the SDK's HTTP transport (URL capture),
- wraps compose call sites (`Chain.AppendChatModel`, `Graph.AddChatModelNode`, `Workflow.AddChatModelNode`, `ChainBranch.AddChatModel`, `Parallel.AddChatModel`) with `WrapBaseChatModel`,
- rewrites direct `cm.Generate` / `cm.Stream` calls to `WrapGenerate` / `WrapStream` for response-metadata extraction.

The manual `WrapChatModel` shown above stays available for explicit use.

## Claude provider

For Claude via eino-ext, replace the openai import with:

```go
import "github.com/cloudwego/eino-ext/components/model/claude"

inner, _ := claude.NewChatModel(ctx, &claude.Config{...})
cm := whatapeino.WrapChatModel(inner)
```

Same wrap pattern.

## Run

```bash
export OPENAI_API_KEY=sk-...
go run main.go
```

## References

- Adapter source: `go-api/instrumentation/llm/github.com/cloudwego/eino/whatapeino/`
