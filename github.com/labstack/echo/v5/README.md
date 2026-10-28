# Echo v5 Framework Instrumentation Example

This example demonstrates how to instrument Echo v5 web framework applications using the WhaTap Go API.

## Overview

Echo v5 is traced with the `whatapecho` package (`whatapecho.Middleware()`, `whatapecho.WrapEcho()`), the same one-liner as Echo v4:

- **Automatic Transaction Tracking**: each HTTP request creates a transaction
- **Request/Response Details**: URL, method, status code, response time
- **Error Tracking**: HTTP status is captured even for 4xx/5xx responses
- **Distributed Tracing**: trace context propagation via headers

## The `whatapecho/v5` nested module

Echo v3 and v4 ship their `whatapecho` package inside the core `github.com/whatap/go-api` module (Go 1.18).

Echo v5 is different:

- It **requires Go 1.25+**.
- `echo.Context` changed from an **interface to a struct**, so handlers now receive `*echo.Context`.

Because bundling an `echo/v5` package inside the core module would force the whole SDK to Go 1.25, the Echo v5 helper ships as a **separate nested module**:

```
github.com/whatap/go-api/instrumentation/github.com/labstack/echo/v5/whatapecho   (go 1.25)
```

The core `go-api` module stays on Go 1.18; only applications that use Echo v5 (already on Go 1.25) pull this module.

## Prerequisites

- Go 1.25+
- WhaTap Go Agent installed

## Installation

```bash
go get github.com/labstack/echo/v5
go get github.com/whatap/go-api/instrumentation/github.com/labstack/echo/v5/whatapecho
```

## Quick Start

### Option A — whatapecho middleware (manual)

**Before (Original Code):**
```go
import echo "github.com/labstack/echo/v5"

e := echo.New()
e.GET("/hello", helloHandler)
e.Start(":8080")
```

**After (Instrumented Code):**
```go
import (
    echo "github.com/labstack/echo/v5"

    "github.com/whatap/go-api/instrumentation/github.com/labstack/echo/v5/whatapecho"
    "github.com/whatap/go-api/trace"
)

func main() {
    trace.Init(nil)
    defer trace.Shutdown()

    e := echo.New()
    e.Use(whatapecho.Middleware()) // add before your other middlewares
    // or: whatapecho.WrapEcho(e)

    e.GET("/hello", helloHandler)
    e.Start(":8080")
}
```

### Option B — Auto-instrumentation (recommended)

Let `whatap-go-inst` inject equivalent middleware automatically at build time — no code changes and no dependency needed (it uses an inline Transform rule that relies only on `trace`):

```bash
whatap-go-inst go build ./...
```

### Access the transaction context

```go
e.GET("/users/:id", func(c *echo.Context) error {
    // Get the transaction context created by the middleware.
    ctx := c.Request().Context()

    // Downstream calls linked to this HTTP transaction.
    db.QueryContext(ctx, "SELECT * FROM users WHERE id = ?", c.Param("id"))

    return c.JSON(200, map[string]string{"message": "ok"})
})
```

## Status code capture

`whatapecho.Middleware()` records the HTTP status on the transaction:

- Committed responses → `echo.Response.Status`
- Returned `*echo.HTTPError` → its `.Code` (e.g. 404, 500)
- Any other error → 500

> Compatibility note: the middleware uses only symbols available in **all v5.0.0+** releases (`Response.Status`, `Response.Committed`, `HTTPError.Code`). It intentionally avoids `ResolveResponseStatus` / `StatusCode`, which were added in v5.0.4 and would break builds on v5.0.0.

## Echo version matrix

| Version | Import Path | Instrumentation |
|---------|-------------|-----------------|
| Echo v5 | `github.com/labstack/echo/v5` | `whatapecho.Middleware()` — nested module (Go 1.25) |
| Echo v4 | `github.com/labstack/echo/v4` | `whatapecho.Middleware()` |
| Echo v3 | `github.com/labstack/echo` | `whatapecho.Middleware()` |

## Examples

| Endpoint | Description |
|----------|-------------|
| `/` | Basic handler |
| `/health` | Health check |
| `/products` | JSON response |
| `/products/:id` | Path parameter + custom step + method tracing |
| `/error` | Returns 500 (status captured on the transaction) |

## Running the Example

```bash
# Build (Go 1.25+; GOTOOLCHAIN=auto will fetch the toolchain if needed)
make

# Run with instrumentation enabled
go run echo.go -whatap

# Test endpoints
curl http://localhost:8080/
curl http://localhost:8080/products/7
curl http://localhost:8080/error
```

## See Also

- [WhaTap Go API Documentation](https://docs.whatap.io/golang/install-agent)
- [Echo Documentation](https://echo.labstack.com/)
- [Echo v4 Example](../v4/)
- [Gin Example](../../../gin-gonic/gin/)
