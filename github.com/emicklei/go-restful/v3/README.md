# go-restful Instrumentation Example

This example demonstrates how to instrument an [emicklei/go-restful](https://github.com/emicklei/go-restful) v3 application using the WhaTap Go API.

## Overview

go-restful's `Container` implements `http.Handler`, but it registers routes through its own `WebService`/`Route` API and bypasses net/http's `HandleFunc`/`ServeMux`. So WhaTap instruments it where the Container is handed to the server:

```go
http.ListenAndServe(":8080", container)                        // Before
http.ListenAndServe(":8080", whataphttp.WrapHandler(container)) // After
```

`whataphttp.WrapHandler` (part of the core `github.com/whatap/go-api` module, Go 1.18) starts a transaction for every request — **no dedicated go-restful module is needed**. The same applies to `restful.DefaultContainer` served via `http.ListenAndServe`.

- **Automatic Transaction Tracking**: each HTTP request creates a transaction
- **Request/Response Details**: URL, method, status code, response time
- **Distributed Tracing**: trace context propagation via headers

## Quick Start

### Option A — manual (wrap at ListenAndServe)

```go
import (
    restful "github.com/emicklei/go-restful/v3"

    "github.com/whatap/go-api/instrumentation/net/http/whataphttp"
    "github.com/whatap/go-api/trace"
)

func main() {
    trace.Init(nil)
    defer trace.Shutdown()

    ws := new(restful.WebService)
    ws.Route(ws.GET("/").To(handler))

    container := restful.NewContainer()
    container.Add(ws)

    http.ListenAndServe(":8080", whataphttp.WrapHandler(container))
}
```

### Option B — Auto-instrumentation (recommended)

Let `whatap-go-inst` wrap the handler automatically at build time (§95 `net/http.ListenAndServe` ArgWrap rule) — no code changes:

```bash
whatap-go-inst go build ./...
```

### Access the transaction context

```go
ws.Route(ws.GET("/users/{id}").To(func(req *restful.Request, resp *restful.Response) {
    ctx := req.Request.Context() // transaction started by WrapHandler
    // db.QueryContext(ctx, ...) linked to this transaction
}))
```

## Note on route naming

Transactions are named by the raw request path. go-restful's route templates (`/users/{id}`) are not currently used for naming; this does not affect transaction capture.

## Running the Example

```bash
make local
go run restful.go

curl http://localhost:8080/
curl http://localhost:8080/health
```

## See Also

- [WhaTap Go API Documentation](https://docs.whatap.io/golang/install-agent)
- [net/http Example](../../../../net/http/server/)
- [httprouter Example](../../../julienschmidt/httprouter/)
