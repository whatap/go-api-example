# MongoDB Go Driver Instrumentation Example

This example demonstrates how to instrument MongoDB Go Driver applications using WhaTap Go API.

## Overview

The `whatapmongo` package provides automatic tracing for MongoDB operations:

- **Connection Tracking**: `sql.StartOpen()` tracks connection establishment
- **Command Monitoring**: All MongoDB commands are traced via `CommandMonitor`
- **Error Tracking**: Errors are captured and linked to transactions
- **Distributed Tracing**: Context propagation enables multi-transaction tracking

## Prerequisites

- Go 1.18+
- MongoDB server
- WhaTap Go Agent installed

## Installation

```bash
go get github.com/whatap/go-api
go get github.com/whatap/go-api/instrumentation/go.mongodb.org/mongo-driver/mongo/whatapmongo
go get go.mongodb.org/mongo-driver/mongo
```

## Quick Start

### 1. Replace mongo.Connect with whatapmongo.Connect

**Before (Original Code):**
```go
import "go.mongodb.org/mongo-driver/mongo"

client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
```

**After (Instrumented Code):**
```go
import "github.com/whatap/go-api/instrumentation/go.mongodb.org/mongo-driver/mongo/whatapmongo"

client, err := whatapmongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
```

### 2. Initialize WhaTap Agent

```go
import "github.com/whatap/go-api/trace"

func main() {
    // Initialize at startup
    config := make(map[string]string)
    config["license"] = "your-license-key"
    config["whatap.server.host"] = "your-whatap-server"
    config["app_name"] = "my-mongodb-app"
    trace.Init(config)
    defer trace.Shutdown()

    // ... application code
}
```

### 3. Pass Context for Transaction Linkage

```go
http.HandleFunc("/insert", func(w http.ResponseWriter, r *http.Request) {
    // Start HTTP transaction
    ctx, _ := trace.StartWithRequest(r)
    defer trace.End(ctx, nil)

    // Pass ctx to MongoDB operations - this links them to the HTTP transaction
    result, err := collection.InsertOne(ctx, doc)
    if err != nil {
        trace.Error(ctx, err)
        return
    }
})
```

## Instrumentation Methods

| Original Function | WhaTap Function | Notes |
|-------------------|-----------------|-------|
| `mongo.Connect()` | `whatapmongo.Connect()` | Recommended for new connections |
| `mongo.NewClient()` | `whatapmongo.NewClient()` | Deprecated, use Connect |

## WhaTap Logs

When instrumented, the following logs are generated:

| Log Code | Description | Example |
|----------|-------------|---------|
| `WA-SQL-04002` | Connection establishment | `Open DB txid: 123, dbhost: mongodb://localhost:27017` |
| `WA-SQL-04003` | Command execution | `Sql txid: 123, cmd: insert, collection: items` |

## Examples in This File

| Endpoint | Operation | Description |
|----------|-----------|-------------|
| `/insert` | InsertOne | Single document insertion |
| `/find` | FindOne | Single document query |
| `/findall` | Find | Multiple documents with cursor |
| `/update` | UpdateOne | Document update |
| `/delete` | DeleteOne | Document deletion |
| `/connect` | Connect | Connection inside handler (shows txid linkage) |
| `/aggregate` | Aggregate | Aggregation pipeline |
| `/insertmany` | InsertMany | Bulk insert |
| `/findandupdate` | FindOneAndUpdate | Atomic find and update |
| `/findbyid` | FindOne | Query by ObjectID |

## Running the Example

```bash
# Start MongoDB
docker run -d -p 27017:27017 --name mongo mongo:latest

# Run the example
go run mongo.go -whatap -mongo mongodb://localhost:27017

# Test endpoints
curl http://localhost:8080/insert?name=test&value=hello
curl http://localhost:8080/find?name=test
curl http://localhost:8080/findall
curl http://localhost:8080/connect  # Check WA-SQL-04002 log for txid
```

## Important Notes

### Context Propagation

Always pass the traced context to MongoDB operations:

```go
// Good - MongoDB command linked to HTTP transaction
result, err := collection.InsertOne(ctx, doc)

// Bad - MongoDB command not linked (uses background context)
result, err := collection.InsertOne(context.Background(), doc)
```

### Connection at Startup vs Inside Handler

| Scenario | txid | Recommendation |
|----------|------|----------------|
| Connect in main() | 0 | OK for connection pooling |
| Connect in handler | non-zero | Better for transaction tracking |

### Error Handling

```go
result, err := collection.InsertOne(ctx, doc)
if err != nil {
    trace.Error(ctx, err)  // Record error in trace
    // handle error
}
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `license` | WhaTap license key | - |
| `whatap.server.host` | WhaTap server address | - |
| `app_name` | Application name | hostname |
| `net_udp_port` | Agent UDP port | 6600 |

## See Also

- [WhaTap Go API Documentation](https://docs.whatap.io/golang/install-agent)
- [MongoDB Go Driver Documentation](https://www.mongodb.com/docs/drivers/go/current/)
- [whatap-go-inst Auto-Instrumentation Tool](https://github.com/whatap/go-api-inst)
