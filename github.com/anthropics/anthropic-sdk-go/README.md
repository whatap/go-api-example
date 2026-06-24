# anthropics/anthropic-sdk-go LLM Instrumentation Example

This example shows how to instrument the [official Anthropic Go SDK](https://github.com/anthropics/anthropic-sdk-go) with WhaTap Go API LLM monitoring.

## Overview

The SDK exposes the messages API as a service field on the Client:

```go
client.Messages.New(ctx, params)
```

Because the interesting method lives on `*MessageService`, the adapter wraps at the service level:

```go
whatapanthropic.WrapAndNewMessage(ctx, client.Messages, params)
```

Same signature, no variable rebinding.

Captured automatically:

- Token counts including **cache creation / cache read** tokens
- Finish reason
- Content blocks (text + tool_use)
- Streaming TTFT via the SDK's discriminated union event types (MessageStartEvent / ContentBlockDeltaEvent / MessageDeltaEvent / ...)

## Prerequisites

- Go 1.18+
- WhaTap Go Agent installed (`/usr/whatap/agent/whatap.conf` present)
- `ANTHROPIC_API_KEY` environment variable

## Installation

```bash
go get github.com/anthropics/anthropic-sdk-go
go get github.com/whatap/go-api
go get github.com/whatap/go-api/instrumentation/llm
```

## Quick Start

### Manual (this example)

```go
import (
    "github.com/anthropics/anthropic-sdk-go"
    "github.com/anthropics/anthropic-sdk-go/option"
    whatapanthropic "github.com/whatap/go-api/instrumentation/llm/github.com/anthropics/anthropic-sdk-go/whatapanthropic"
    whataphttp "github.com/whatap/go-api/instrumentation/net/http/whataphttp"
    "github.com/whatap/go-api/trace"
)

trace.Init(nil)
ctx, _ := trace.Start(context.Background(), "llm-example")

client := anthropic.NewClient(
    option.WithAPIKey(apiKey),
    option.WithHTTPClient(&http.Client{Transport: whataphttp.NewRoundTrip(ctx, http.DefaultTransport)}),
)

resp, _ := whatapanthropic.WrapAndNewMessage(ctx, client.Messages, anthropic.MessageNewParams{...})
trace.End(ctx, nil)
```

### Auto-inject (whatap-go-inst)

```bash
whatap-go-inst go build .
```

The tool rewrites `client.Messages.New(ctx, params)` → `whatapanthropic.WrapAndNewMessage(ctx, client.Messages, params)` automatically.

## Run

```bash
export ANTHROPIC_API_KEY=sk-ant-...
go run main.go
```

## Streaming

For `client.Messages.NewStreaming(ctx, params)`, see the adapter's `stream.go`. The streaming wrapper accumulates text + tool_use deltas across SDK event types automatically.

## References

- Adapter source: `go-api/instrumentation/llm/github.com/anthropics/anthropic-sdk-go/whatapanthropic/`
