# Aerospike Client Instrumentation

WhaTap APM instrumentation example for Aerospike Go Client.

## Instrumentation Method: sql.Wrap Pattern

Since Aerospike client doesn't provide Hook/Monitor interface, we use the `sql.Wrap` pattern for manual instrumentation.

| Pattern | Use Case | Example |
|---------|----------|---------|
| `sql.WrapOpen()` | Client connection | `aerospike.NewClient()` |
| `sql.Wrap()` | Methods returning `(T, error)` | `Get`, `Delete`, `Exists`, `BatchGet` |
| `sql.WrapError()` | Methods returning only `error` | `Put` |

## Quick Start

```go
import (
    aerospike "github.com/aerospike/aerospike-client-go/v6"
    "github.com/whatap/go-api/sql"
    "github.com/whatap/go-api/trace"
)

func main() {
    trace.Init(nil)
    defer trace.Shutdown()

    // Connection with tracing
    client, err := sql.WrapOpen(ctx, "aerospike", "aerospike://host:3000",
        func() (*aerospike.Client, error) {
            return aerospike.NewClient("host", 3000)
        })()
}

func handler(w http.ResponseWriter, r *http.Request) {
    ctx, _ := trace.StartWithRequest(r)
    defer trace.End(ctx, nil)

    // Get with tracing
    record, err := sql.Wrap(ctx, "aerospike", "Get", func() (*aerospike.Record, error) {
        return client.Get(nil, key)
    })()

    // Put with tracing
    err = sql.WrapError(ctx, "aerospike", "Put", func() error {
        return client.Put(nil, key, bins)
    })()
}
```

## Before/After Comparison

### Connection
```go
// Before
client, err := aerospike.NewClient("localhost", 3000)

// After (with tracing)
client, err := sql.WrapOpen(ctx, "aerospike", "aerospike://localhost:3000",
    func() (*aerospike.Client, error) {
        return aerospike.NewClient("localhost", 3000)
    })()
```

### Get Operation
```go
// Before
record, err := client.Get(nil, key)

// After (with tracing)
record, err := sql.Wrap(ctx, "aerospike", "Get", func() (*aerospike.Record, error) {
    return client.Get(nil, key)
})()
```

### Put Operation
```go
// Before
err := client.Put(nil, key, bins)

// After (with tracing)
err = sql.WrapError(ctx, "aerospike", "Put", func() error {
    return client.Put(nil, key, bins)
})()
```

## Running the Example

```bash
# Start Aerospike (Docker)
docker run -d --name aerospike -p 3000:3000 aerospike/aerospike-server

# Run the example
go run aerospike.go -whatap -ashost localhost -asport 3000

# Test endpoints
curl http://localhost:8104/
curl "http://localhost:8104/put?name=test&value=hello"
curl "http://localhost:8104/get?name=test"
curl "http://localhost:8104/exists?name=test"
curl "http://localhost:8104/delete?name=test"
```

## WhaTap Logs

- `[WA-SQL-04002] Open DB`: Connection establishment
- `[WA-SQL-04003] Sql`: Aerospike command execution

## Supported Methods

| Method | Wrap Function | Return Type |
|--------|--------------|-------------|
| `NewClient` | `sql.WrapOpen()` | `(*Client, error)` |
| `Get` | `sql.Wrap()` | `(*Record, error)` |
| `Put` | `sql.WrapError()` | `error` |
| `Delete` | `sql.Wrap()` | `(bool, error)` |
| `Exists` | `sql.Wrap()` | `(bool, error)` |
| `BatchGet` | `sql.Wrap()` | `([]*Record, error)` |
| `Query` | `sql.Wrap()` | `(*Recordset, error)` |
| `Scan` | `sql.Wrap()` | `(*Recordset, error)` |
