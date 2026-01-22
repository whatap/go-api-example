# Gin Framework Instrumentation Example

This example demonstrates how to instrument Gin web framework applications using WhaTap Go API.

## Overview

The `whatapgin` package provides automatic HTTP transaction tracing for Gin:

- **Automatic Transaction Tracking**: Each HTTP request creates a transaction
- **Request/Response Details**: URL, method, status code, response time
- **Error Tracking**: Panics and errors are captured
- **Distributed Tracing**: Trace context propagation via headers

## Prerequisites

- Go 1.18+
- WhaTap Go Agent installed

## Installation

```bash
go get github.com/whatap/go-api
go get github.com/whatap/go-api/instrumentation/github.com/gin-gonic/gin/whatapgin
go get github.com/gin-gonic/gin
```

## Quick Start

### 1. Add WhaTap Middleware

**Before (Original Code):**
```go
import "github.com/gin-gonic/gin"

r := gin.Default()
r.GET("/hello", helloHandler)
r.Run(":8080")
```

**After (Instrumented Code):**
```go
import (
    "github.com/gin-gonic/gin"
    "github.com/whatap/go-api/instrumentation/github.com/gin-gonic/gin/whatapgin"
    "github.com/whatap/go-api/trace"
)

func main() {
    // Initialize WhaTap agent
    trace.Init(nil)
    defer trace.Shutdown()

    r := gin.Default()
    
    // Add WhaTap middleware - must be first middleware
    r.Use(whatapgin.Middleware())
    
    r.GET("/hello", helloHandler)
    r.Run(":8080")
}
```

### 2. Access Transaction Context

```go
r.GET("/users/:id", func(c *gin.Context) {
    // Get transaction context from gin.Context
    ctx := c.Request.Context()
    
    // Use ctx for downstream calls (database, HTTP client, etc.)
    // These calls will be linked to the HTTP transaction
    db.QueryContext(ctx, "SELECT * FROM users WHERE id = ?", c.Param("id"))
    
    c.JSON(200, gin.H{"message": "ok"})
})
```

### 3. Custom Error Tracking

```go
r.GET("/error", func(c *gin.Context) {
    ctx := c.Request.Context()
    
    result, err := someOperation()
    if err != nil {
        // Record error in transaction trace
        trace.Error(ctx, err)
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, result)
})
```

## Middleware Order

The WhaTap middleware should be added before other middlewares to ensure all requests are traced:

```go
r := gin.New()

// 1. WhaTap middleware (first)
r.Use(whatapgin.Middleware())

// 2. Other middlewares
r.Use(gin.Logger())
r.Use(gin.Recovery())
r.Use(cors.Default())
```

## Configuration

### Via whatap.conf

```properties
license=your-license-key
whatap.server.host=your-whatap-server
app_name=my-gin-app
```

### Via Code

```go
config := make(map[string]string)
config["license"] = "your-license-key"
config["whatap.server.host"] = "your-whatap-server"
config["app_name"] = "my-gin-app"
trace.Init(config)
```

## Examples

| Endpoint | Description |
|----------|-------------|
| `/` | Basic handler |
| `/hello/:name` | Path parameter |
| `/users` | JSON response |
| `/error` | Error handling |

## Running the Example

```bash
# Run the example
go run gin.go -whatap

# Test endpoints
curl http://localhost:8080/
curl http://localhost:8080/hello/world
curl http://localhost:8080/users
```

## See Also

- [WhaTap Go API Documentation](https://docs.whatap.io/golang/install-agent)
- [Gin Documentation](https://gin-gonic.com/docs/)
- [Echo Example](../../labstack/echo/v4/)
- [Fiber Example](../../gofiber/fiber/v2/)
