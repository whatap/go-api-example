# Beego v1 Framework Instrumentation Example

This example demonstrates how to instrument a Beego v1 web application (`github.com/beego/beego`, v1.12.x — the pre-rebrand root package) using the WhaTap Go API.

## Overview

Beego is an MVC framework whose server is started by `beego.Run()` rather than by wrapping a router/engine object. WhaTap instruments it at the server-start call:

```go
beego.Run(":8080")                                            // Before
beego.RunWithMiddleWares(":8080", whataphttp.WrapHandler)     // After
```

- **Automatic Transaction Tracking**: each HTTP request creates a transaction
- **Request/Response Details**: URL, method, status code, response time
- **Distributed Tracing**: trace context propagation via headers

## Why there is no dedicated `whatapbeego` package

Beego's middleware type — identical in v1 and v2 — is:

```go
type MiddleWare func(http.Handler) http.Handler
```

That is exactly the **net/http middleware** signature. So the core wrapper `whataphttp.WrapHandler` (shipped inside `github.com/whatap/go-api`, Go 1.18) plugs straight into `RunWithMiddleWares` — no Beego import is added to `go-api` and **no nested module is required**.

Real-world Beego v1 apps such as [casdoor](https://github.com/casdoor/casdoor) use exactly this `beego.Run(...)` entry point.

## Prerequisites

- Go 1.13+ (Beego v1.12.x)
- WhaTap Go Agent installed

## Installation

```bash
go get github.com/beego/beego@v1.12.12
go get github.com/whatap/go-api
```

## Quick Start

### Option A — manual (one line)

```go
import (
    "github.com/beego/beego"

    "github.com/whatap/go-api/instrumentation/net/http/whataphttp"
    "github.com/whatap/go-api/trace"
)

func main() {
    trace.Init(nil)
    defer trace.Shutdown()

    beego.Router("/", &MainController{})

    // Start with the WhaTap middleware.
    beego.RunWithMiddleWares(":8080", whataphttp.WrapHandler)
}
```

### Option B — Auto-instrumentation (recommended)

Let `whatap-go-inst` rewrite `beego.Run(...)` automatically at build time (§50/§204 inline Transform rule) — no code changes:

```bash
whatap-go-inst go build ./...
```

### Access the transaction context

```go
func (c *MainController) Get() {
    ctx := c.Ctx.Request.Context()
    trace.Step(ctx, "lookup", "hello handler", 1, 1)
    c.Data["json"] = map[string]string{"message": "ok"}
    c.ServeJSON()   // v1 ServeJSON() returns no value (v2 returns error)
}
```

## Beego version matrix

| Version | Import Path | Instrumentation |
|---------|-------------|-----------------|
| Beego v1 | `github.com/beego/beego` | `beego.RunWithMiddleWares(addr, whataphttp.WrapHandler)` |
| Beego v2 | `github.com/beego/beego/v2/server/web` | `web.RunWithMiddleWares(addr, whataphttp.WrapHandler)` — see [v2/](./v2/) |

## Running the Example

```bash
make local         # build against in-repo go-api sources
go run beego.go     # run (configure whatap.conf / trace.Init as needed)

curl http://localhost:8080/
curl http://localhost:8080/health
```

## See Also

- [WhaTap Go API Documentation](https://docs.whatap.io/golang/install-agent)
- [Beego v2 Example](./v2/)
- [Gin Example](../../gin-gonic/gin/)
