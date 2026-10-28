# Beego v2 Framework Instrumentation Example

This example demonstrates how to instrument a Beego v2 web application using the WhaTap Go API.

## Overview

Beego is an MVC framework whose server is started by `web.Run()` rather than by wrapping a router/engine object. WhaTap instruments it at the server-start call:

```go
web.Run(":8080")                                            // Before
web.RunWithMiddleWares(":8080", whataphttp.WrapHandler)     // After
```

- **Automatic Transaction Tracking**: each HTTP request creates a transaction
- **Request/Response Details**: URL, method, status code, response time
- **Distributed Tracing**: trace context propagation via headers

## Why there is no dedicated `whatapbeego` package

Beego's middleware type is:

```go
type MiddleWare func(http.Handler) http.Handler
```

That is exactly the **net/http middleware** signature. So the core wrapper `whataphttp.WrapHandler` (shipped inside `github.com/whatap/go-api`, Go 1.18) plugs straight into `RunWithMiddleWares` — no Beego import is added to `go-api` and **no nested module is required** (unlike Echo v5 / Fiber v3, whose middleware signatures differ from net/http).

A nice side effect: manual users need nothing beyond the core module — just the one-line change above.

## Prerequisites

- Go 1.20+ (Beego v2 requirement)
- WhaTap Go Agent installed

## Installation

```bash
go get github.com/beego/beego/v2
go get github.com/whatap/go-api
```

## Quick Start

### Option A — manual (one line)

```go
import (
    "github.com/beego/beego/v2/server/web"

    "github.com/whatap/go-api/instrumentation/net/http/whataphttp"
    "github.com/whatap/go-api/trace"
)

func main() {
    trace.Init(nil)
    defer trace.Shutdown()

    web.Router("/", &MainController{})

    // Start with the WhaTap middleware.
    web.RunWithMiddleWares(":8080", whataphttp.WrapHandler)
}
```

### Option B — Auto-instrumentation (recommended)

Let `whatap-go-inst` rewrite `web.Run(...)` automatically at build time (§204 inline Transform rule) — no code changes:

```bash
whatap-go-inst go build ./...
```

### Access the transaction context

Inside a Controller, the transaction lives on the request context:

```go
func (c *MainController) Get() {
    ctx := c.Ctx.Request.Context()
    trace.Step(ctx, "lookup", "hello handler", 1, 1)
    // db.QueryContext(ctx, ...) / httpc calls linked to this transaction
    c.Data["json"] = map[string]string{"message": "ok"}
    _ = c.ServeJSON()
}
```

## Beego version matrix

| Version | Import Path | Instrumentation |
|---------|-------------|-----------------|
| Beego v2 | `github.com/beego/beego/v2/server/web` | `web.RunWithMiddleWares(addr, whataphttp.WrapHandler)` |
| Beego v1 | `github.com/beego/beego` | `beego.RunWithMiddleWares(addr, whataphttp.WrapHandler)` — see [../](../) |

Both versions reuse the core `whataphttp.WrapHandler`; the only difference is the package the `Run`/`RunWithMiddleWares` functions live in.

## Running the Example

```bash
make local        # build against in-repo go-api sources
go run beego.go    # run (configure whatap.conf / trace.Init as needed)

curl http://localhost:8080/
curl http://localhost:8080/health
```

## See Also

- [WhaTap Go API Documentation](https://docs.whatap.io/golang/install-agent)
- [Beego v1 Example](../)
- [Gin Example](../../../gin-gonic/gin/)
