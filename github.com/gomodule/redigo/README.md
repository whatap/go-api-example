# Redigo Instrumentation Example

This example demonstrates how to instrument Redigo (github.com/gomodule/redigo) applications using WhaTap Go API.

## Overview

The `whatapredigo` package provides automatic tracing for Redis operations using Redigo:

- **Connection Tracking**: Dial functions track connection establishment (StartOpen)
- **Command Monitoring**: All Redis commands are traced via Conn wrapper
- **Connection Pool Support**: Pool with traced DialContext function
- **Error Tracking**: Errors are captured and linked to transactions

## Prerequisites

- Go 1.18+
- Redis server
- WhaTap Go Agent installed

## Installation

```bash
go get github.com/whatap/go-api
go get github.com/whatap/go-api/instrumentation/github.com/gomodule/redigo/whatapredigo
go get github.com/gomodule/redigo/redis
```

## Quick Start

### 1. Replace Dial Functions

**Before (Original Code):**
```go
import "github.com/gomodule/redigo/redis"

conn, err := redis.Dial("tcp", "localhost:6379")
// or
conn, err := redis.DialContext(ctx, "tcp", "localhost:6379")
// or
conn, err := redis.DialURL("redis://localhost:6379")
```

**After (Instrumented Code):**
```go
import "github.com/whatap/go-api/instrumentation/github.com/gomodule/redigo/whatapredigo"

conn, err := whatapredigo.Dial("tcp", "localhost:6379")
// or
conn, err := whatapredigo.DialContext(ctx, "tcp", "localhost:6379")
// or
conn, err := whatapredigo.DialURL("redis://localhost:6379")
```

### 2. Use WithContext for Command Tracing

```go
http.HandleFunc("/redis", func(w http.ResponseWriter, r *http.Request) {
    ctx, _ := trace.StartWithRequest(r)
    defer trace.End(ctx, nil)

    conn, err := whatapredigo.Dial("tcp", "localhost:6379")
    if err != nil {
        trace.Error(ctx, err)
        return
    }
    defer conn.Close()

    // IMPORTANT: Use WithContext to link commands to transaction
    conn = conn.WithContext(ctx)

    _, err = conn.Do("SET", "key", "value")
    if err != nil {
        trace.Error(ctx, err)
    }
})
```

### 3. Connection Pool with Tracing

```go
pool := &redis.Pool{
    MaxIdle:     3,
    IdleTimeout: 240 * time.Second,
    DialContext: func(ctx context.Context) (redis.Conn, error) {
        // Use whatapredigo.DialContext for traced connections
        return whatapredigo.DialContext(ctx, "tcp", "localhost:6379")
    },
}

http.HandleFunc("/pool", func(w http.ResponseWriter, r *http.Request) {
    ctx, _ := trace.StartWithRequest(r)
    defer trace.End(ctx, nil)

    // GetContext passes the transaction context to the connection
    conn, err := pool.GetContext(ctx)
    if err != nil {
        trace.Error(ctx, err)
        return
    }
    defer conn.Close()

    // Commands are automatically linked to the transaction
    _, err = conn.Do("GET", "key")
})
```

## Instrumentation Methods

| Original Function | WhaTap Function | Notes |
|-------------------|-----------------|-------|
| `redis.Dial()` | `whatapredigo.Dial()` | Basic dial |
| `redis.DialContext()` | `whatapredigo.DialContext()` | Dial with context |
| `redis.DialURL()` | `whatapredigo.DialURL()` | Dial with URL |
| `redis.DialURLContext()` | `whatapredigo.DialURLContext()` | URL with context |

## Examples in This File

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/SetAndGetWithPool` | Pool + GetContext | Connection pool usage |
| `/SetAndGetWithDial` | Dial + WithContext | Basic dial usage |
| `/SetAndGetWithDialContext` | DialContext | Context-aware dial |
| `/SetAndGetWithDialURL` | DialURL + WithContext | URL-based dial |
| `/SetAndGetWithDialURLContext` | DialURLContext | URL with context |
| `/SetAndGetWithDialTimeout` | Dial + Timeout options | Connection timeout |
| `/SetAndGetWithDialSendReceive` | Send/Receive | Pipeline-style commands |

## Key Concepts

### WithContext

After calling `Dial()` or `DialURL()`, you must call `conn.WithContext(ctx)` to link commands to the HTTP transaction:

```go
conn, _ := whatapredigo.Dial("tcp", "localhost:6379")
conn = conn.WithContext(ctx)  // Links to HTTP transaction
conn.Do("SET", "key", "value")  // This command is now traced
```

### DialContext vs Dial + WithContext

| Method | Connection Tracking | Command Tracking |
|--------|---------------------|------------------|
| `DialContext(ctx, ...)` | Yes (via ctx) | Yes (automatic) |
| `Dial(...) + WithContext(ctx)` | No context at dial | Yes (via WithContext) |

**Recommendation**: Use `DialContext` when possible for full tracing support.

### Connection Pool Best Practice

```go
pool := &redis.Pool{
    DialContext: func(ctx context.Context) (redis.Conn, error) {
        return whatapredigo.DialContext(ctx, "tcp", addr)
    },
}

// In handler:
conn, _ := pool.GetContext(ctx)  // ctx is passed to DialContext
```

## Running the Example

```bash
# Start Redis
docker run -d -p 6379:6379 --name redis redis:latest

# Run the example
go run redigo.go -whatap -ds localhost:6379

# Test endpoints
curl http://localhost:8080/SetAndGetWithPool
curl http://localhost:8080/SetAndGetWithDialContext
```

## See Also

- [WhaTap Go API Documentation](https://docs.whatap.io/golang/install-agent)
- [Redigo Documentation](https://github.com/gomodule/redigo)
- [go-redis Example](../redis/go-redis/v9/) (Alternative Redis client)
