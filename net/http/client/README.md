# net/http Client Instrumentation Example

This example demonstrates how to instrument HTTP client calls using WhaTap Go API.

## Overview

The `whataphttp` package provides HTTP client call tracing:

- **Outgoing Request Tracking**: HTTP calls to external services are traced
- **Request/Response Details**: URL, method, status code, response time
- **Distributed Tracing**: Trace context is propagated via HTTP headers
- **Error Tracking**: Connection errors and timeouts are captured

## Prerequisites

- Go 1.18+
- WhaTap Go Agent installed

## Installation

```bash
go get github.com/whatap/go-api
go get github.com/whatap/go-api/instrumentation/net/http/whataphttp
```

## Quick Start

### Method 1: Replace http.Get/Post Functions

**Before (Original Code):**
```go
resp, err := http.Get("https://api.example.com/users")
resp, err := http.Post("https://api.example.com/users", "application/json", body)
```

**After (Instrumented Code):**
```go
import "github.com/whatap/go-api/instrumentation/net/http/whataphttp"

// Pass context from HTTP handler
ctx := r.Context()

resp, err := whataphttp.Get(ctx, "https://api.example.com/users")
resp, err := whataphttp.Post(ctx, "https://api.example.com/users", "application/json", body)
```

### Method 2: Wrap http.Client with RoundTripper

**Before (Original Code):**
```go
client := &http.Client{}
resp, err := client.Do(req)
```

**After (Instrumented Code):**
```go
import "github.com/whatap/go-api/instrumentation/net/http/whataphttp"

client := &http.Client{
    Transport: whataphttp.NewRoundTrip(nil),  // nil uses http.DefaultTransport
}
resp, err := client.Do(req)
```

### Method 3: Wrap Existing Transport

```go
client := &http.Client{
    Transport: whataphttp.NewRoundTrip(&http.Transport{
        MaxIdleConns:    100,
        IdleConnTimeout: 90 * time.Second,
    }),
}
```

## Instrumentation Methods

| Original Function | WhaTap Function | Description |
|-------------------|-----------------|-------------|
| `http.Get()` | `whataphttp.Get()` | Simple GET request |
| `http.Post()` | `whataphttp.Post()` | POST request |
| `http.PostForm()` | `whataphttp.PostForm()` | Form POST request |
| `http.Client{}` | `Transport: whataphttp.NewRoundTrip()` | Custom client |

## Complete Example

```go
package main

import (
    "io"
    "net/http"
    
    "github.com/whatap/go-api/instrumentation/net/http/whataphttp"
    "github.com/whatap/go-api/trace"
)

func main() {
    trace.Init(nil)
    defer trace.Shutdown()
    
    // Server handler that makes HTTP client calls
    http.HandleFunc("/proxy", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        
        // Method 1: Direct function replacement
        resp, err := whataphttp.Get(ctx, "https://httpbin.org/get")
        if err != nil {
            trace.Error(ctx, err)
            http.Error(w, err.Error(), 500)
            return
        }
        defer resp.Body.Close()
        
        body, _ := io.ReadAll(resp.Body)
        w.Write(body)
    }))
    
    // Using custom client with RoundTripper
    http.HandleFunc("/custom", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        
        // Method 2: Custom client with RoundTripper
        client := &http.Client{
            Transport: whataphttp.NewRoundTrip(nil),
        }
        
        req, _ := http.NewRequestWithContext(ctx, "GET", "https://httpbin.org/get", nil)
        resp, err := client.Do(req)
        if err != nil {
            trace.Error(ctx, err)
            http.Error(w, err.Error(), 500)
            return
        }
        defer resp.Body.Close()
        
        body, _ := io.ReadAll(resp.Body)
        w.Write(body)
    }))
    
    http.ListenAndServe(":8080", nil)
}
```

## Distributed Tracing

HTTP client instrumentation automatically propagates trace context via headers:

```
                    x-whatap-mtid: abc123
Service A ────────────────────────────────> Service B
(client)                                    (server)
  │                                            │
  ├─ txid: 1001                               ├─ txid: 2001
  │                                           │  mtid: abc123
  └─ Records HTTP call                        └─ Linked to txid: 1001
```

Headers propagated:
- `x-whatap-mtid`: Multi-transaction ID
- `x-whatap-traceid`: Trace ID
- `x-whatap-poid`: Parent OID

## Examples

| Endpoint | Description |
|----------|-------------|
| `/test-get` | GET request using whataphttp.Get |
| `/test-post` | POST request using whataphttp.Post |
| `/test-client` | Custom client with RoundTripper |
| `/test-roundtrip` | RoundTripper direct usage |

## Running the Example

```bash
# Run the example
go run client.go -whatap

# Test endpoints
curl http://localhost:8080/test-get
curl http://localhost:8080/test-post
curl http://localhost:8080/test-client
```

## See Also

- [WhaTap Go API Documentation](https://docs.whatap.io/golang/install-agent)
- [net/http Server Example](../server/)
- [Distributed Tracing Guide](../../distributed_tracing/)
