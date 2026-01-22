# go-redis v8 Instrumentation Example

This example demonstrates how to instrument go-redis v8 applications using WhaTap Go API.

## Overview

The `whatapgoredis` package (v8 version) provides automatic tracing for Redis operations:

- **Command Monitoring**: All Redis commands are traced via BeforeProcess/AfterProcess hooks
- **Pipeline Support**: Pipeline and transaction commands are tracked
- **Error Tracking**: Errors are captured and linked to transactions

**Note**: go-redis v8 does not support DialHook, so connection establishment (StartOpen) tracking is not available. Consider upgrading to v9 for full instrumentation support.

## Prerequisites

- Go 1.18+
- Redis server
- WhaTap Go Agent installed

## Installation

```bash
go get github.com/whatap/go-api
go get github.com/whatap/go-api/instrumentation/github.com/go-redis/redis/v8/whatapgoredis
go get github.com/go-redis/redis/v8
```

## Quick Start

### 1. Replace redis.NewClient with whatapgoredis.NewClient

**Before (Original Code):**
```go
import "github.com/go-redis/redis/v8"

rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})
```

**After (Instrumented Code):**
```go
import "github.com/whatap/go-api/instrumentation/github.com/go-redis/redis/v8/whatapgoredis"

rdb := whatapgoredis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})
```

### 2. Pass Context for Transaction Linkage

```go
http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
    ctx, _ := trace.StartWithRequest(r)
    defer trace.End(ctx, nil)

    // Pass ctx to Redis operations
    err := rdb.Set(ctx, "key", "value", time.Minute).Err()
    if err != nil {
        trace.Error(ctx, err)
        return
    }
})
```

## v8 vs v9 API Differences

| Feature | v8 | v9 |
|---------|----|----|
| Import path | `github.com/go-redis/redis/v8` | `github.com/redis/go-redis/v9` |
| Hook interface | BeforeProcess/AfterProcess | ProcessHook |
| DialHook | **Not available** | Available |
| Connection tracking | **Not supported** | Supported |
| ZAdd syntax | `&redis.Z{Score: 100, Member: "a"}` | `redis.Z{Score: 100, Member: "a"}` |

## Instrumentation Methods

| Original Function | WhaTap Function | Notes |
|-------------------|-----------------|-------|
| `redis.NewClient()` | `whatapgoredis.NewClient()` | Standard client |
| `redis.NewClusterClient()` | `whatapgoredis.NewClusterClient()` | Cluster mode |
| `redis.NewFailoverClient()` | `whatapgoredis.NewFailoverClient()` | Sentinel mode |
| - | `whatapgoredis.WrapClient()` | Wrap existing client |

## Examples in This File

| Endpoint | Operation | Description |
|----------|-----------|-------------|
| `/SetAndGet` | SET, GET | Basic key-value operations |
| `/Pipeline` | Pipeline | Batch command execution |
| `/Transaction` | MULTI/EXEC | Atomic transactions |
| `/Hash` | HSET, HGETALL | Hash data structure |
| `/List` | LPUSH, LRANGE | List data structure |
| `/Set` | SADD, SMEMBERS | Set data structure |
| `/SortedSet` | ZADD, ZRANGE | Sorted set (leaderboard) |
| `/Publish` | PUBLISH | Pub/Sub publishing |
| `/ExpireAndTTL` | SET, TTL | Key expiration |
| `/WrapClient` | WrapClient | Instrument existing client |

## Running the Example

```bash
# Start Redis
docker run -d -p 6379:6379 --name redis redis:latest

# Run the example
go run goredis.go -whatap -redis localhost:6379

# Test endpoints
curl http://localhost:8080/SetAndGet
curl http://localhost:8080/Pipeline
```

## Migration to v9

If you're using go-redis v8, consider upgrading to v9 for:
- Full connection tracking support (DialHook)
- Better Hook interface (ProcessHook)
- Active maintenance and new features

Migration steps:
1. Change import path: `github.com/go-redis/redis/v8` → `github.com/redis/go-redis/v9`
2. Change WhaTap import path accordingly
3. Update ZAdd calls: `&redis.Z{}` → `redis.Z{}`

## See Also

- [WhaTap Go API Documentation](https://docs.whatap.io/golang/install-agent)
- [go-redis v9 Example](../../../redis/go-redis/v9/) (Recommended)
- [go-redis Documentation](https://redis.uptrace.dev/)
