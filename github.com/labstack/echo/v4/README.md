# Echo v4 Framework Instrumentation Example

This example demonstrates how to instrument Echo v4 web framework applications using WhaTap Go API.

## Overview

The `whatapecho` package provides automatic HTTP transaction tracing for Echo v4:

- **Automatic Transaction Tracking**: Each HTTP request creates a transaction
- **Request/Response Details**: URL, method, status code, response time
- **Error Tracking**: Errors are captured and linked to transactions
- **Distributed Tracing**: Trace context propagation via headers

## Prerequisites

- Go 1.18+
- WhaTap Go Agent installed

## Installation

```bash
go get github.com/whatap/go-api
go get github.com/whatap/go-api/instrumentation/github.com/labstack/echo/v4/whatapecho
go get github.com/labstack/echo/v4
```

## Quick Start

### 1. Add WhaTap Middleware

**Before (Original Code):**
```go
import "github.com/labstack/echo/v4"

e := echo.New()
e.GET("/hello", helloHandler)
e.Start(":8080")
```

**After (Instrumented Code):**
```go
import (
    "github.com/labstack/echo/v4"
    "github.com/whatap/go-api/instrumentation/github.com/labstack/echo/v4/whatapecho"
    "github.com/whatap/go-api/trace"
)

func main() {
    // Initialize WhaTap agent
    trace.Init(nil)
    defer trace.Shutdown()

    e := echo.New()
    
    // Add WhaTap middleware - must be first middleware
    e.Use(whatapecho.Middleware())
    
    e.GET("/hello", helloHandler)
    e.Start(":8080")
}
```

### 2. Access Transaction Context

```go
e.GET("/users/:id", func(c echo.Context) error {
    // Get transaction context from echo.Context
    ctx := c.Request().Context()
    
    // Use ctx for downstream calls (database, HTTP client, etc.)
    // These calls will be linked to the HTTP transaction
    db.QueryContext(ctx, "SELECT * FROM users WHERE id = ?", c.Param("id"))
    
    return c.JSON(200, map[string]string{"message": "ok"})
})
```

### 3. Custom Error Tracking

```go
e.GET("/error", func(c echo.Context) error {
    ctx := c.Request().Context()
    
    result, err := someOperation()
    if err != nil {
        // Record error in transaction trace
        trace.Error(ctx, err)
        return c.JSON(500, map[string]string{"error": err.Error()})
    }
    
    return c.JSON(200, result)
})
```

## Middleware Order

The WhaTap middleware should be added before other middlewares:

```go
e := echo.New()

// 1. WhaTap middleware (first)
e.Use(whatapecho.Middleware())

// 2. Other middlewares
e.Use(middleware.Logger())
e.Use(middleware.Recover())
e.Use(middleware.CORS())
```

## Echo v3 vs v4

| Version | Import Path | WhaTap Package |
|---------|-------------|----------------|
| Echo v4 | `github.com/labstack/echo/v4` | `whatapecho` (v4 path) |
| Echo v3 | `github.com/labstack/echo` | `whatapecho` (no version) |

For Echo v3, see the [Echo v3 example](../echo/).

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
go run echo.go -whatap

# Test endpoints
curl http://localhost:8080/
curl http://localhost:8080/hello/world
curl http://localhost:8080/users
```

## See Also

- [WhaTap Go API Documentation](https://docs.whatap.io/golang/install-agent)
- [Echo Documentation](https://echo.labstack.com/)
- [Echo v3 Example](../echo/)
- [Gin Example](../../gin-gonic/gin/)
