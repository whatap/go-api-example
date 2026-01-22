# net/http Server Instrumentation Example

This example demonstrates how to instrument standard net/http server applications using WhaTap Go API.

## Overview

The `whataphttp` package provides HTTP transaction tracing for Go's standard net/http package:

- **Handler Wrapping**: Wrap handlers for automatic transaction tracking
- **Request/Response Details**: URL, method, status code, response time
- **Error Tracking**: Errors are captured and linked to transactions
- **Distributed Tracing**: Trace context propagation via headers

## Prerequisites

- Go 1.18+
- WhaTap Go Agent installed

## Installation

```bash
go get github.com/whatap/go-api
go get github.com/whatap/go-api/instrumentation/net/http/whataphttp
```

## Quick Start

### Method 1: whataphttp.Func (Recommended)

**Before (Original Code):**
```go
http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello, World!"))
})
```

**After (Instrumented Code):**
```go
import "github.com/whatap/go-api/instrumentation/net/http/whataphttp"

// Wrap handler function
http.HandleFunc("/hello", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello, World!"))
}))
```

### Method 2: whataphttp.Handler (For http.Handler)

```go
type MyHandler struct{}

func (h *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello!"))
}

// Wrap http.Handler
http.Handle("/api", whataphttp.Handler(&MyHandler{}))
```

### Method 3: Manual Transaction Control

```go
http.HandleFunc("/manual", func(w http.ResponseWriter, r *http.Request) {
    // Start transaction manually
    ctx, _ := trace.StartWithRequest(r)
    defer trace.End(ctx, nil)
    
    // Your handler logic
    w.Write([]byte("Hello!"))
})
```

## Accessing Transaction Context

```go
http.HandleFunc("/users", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
    // Get transaction context from request
    ctx := r.Context()
    
    // Use ctx for downstream calls (database, HTTP client, etc.)
    // These calls will be linked to the HTTP transaction
    db.QueryContext(ctx, "SELECT * FROM users")
    
    w.Write([]byte("OK"))
}))
```

## Error Tracking

```go
http.HandleFunc("/error", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    result, err := someOperation()
    if err != nil {
        // Record error in transaction trace
        trace.Error(ctx, err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(result)
}))
```

## Instrumentation Methods

| Method | Use Case |
|--------|----------|
| `whataphttp.Func()` | Wrap `func(http.ResponseWriter, *http.Request)` |
| `whataphttp.Handler()` | Wrap `http.Handler` |
| `trace.StartWithRequest()` | Manual transaction control |

## Complete Example

```go
package main

import (
    "fmt"
    "net/http"
    
    "github.com/whatap/go-api/instrumentation/net/http/whataphttp"
    "github.com/whatap/go-api/trace"
)

func main() {
    // Initialize WhaTap agent
    trace.Init(nil)
    defer trace.Shutdown()
    
    // Handler with automatic tracing
    http.HandleFunc("/", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, World!")
    }))
    
    // Handler with database call
    http.HandleFunc("/users", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        
        // Database call will be linked to HTTP transaction
        users, err := db.QueryContext(ctx, "SELECT * FROM users")
        if err != nil {
            trace.Error(ctx, err)
            http.Error(w, err.Error(), 500)
            return
        }
        
        json.NewEncoder(w).Encode(users)
    }))
    
    http.ListenAndServe(":8080", nil)
}
```

## Examples

| Endpoint | Description |
|----------|-------------|
| `/` | Basic handler |
| `/hello` | Simple response |
| `/users` | Database query |
| `/error` | Error handling |

## Running the Example

```bash
# Run the example
go run server.go -whatap

# Test endpoints
curl http://localhost:8080/
curl http://localhost:8080/hello
curl http://localhost:8080/users
```

## See Also

- [WhaTap Go API Documentation](https://docs.whatap.io/golang/install-agent)
- [net/http Client Example](../client/)
- [Gin Example](../../github.com/gin-gonic/gin/)
- [Echo Example](../../github.com/labstack/echo/v4/)
