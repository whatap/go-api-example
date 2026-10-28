# httprouter Instrumentation Example

This example demonstrates how to instrument a [julienschmidt/httprouter](https://github.com/julienschmidt/httprouter) application using the WhaTap Go API.

## Overview

`httprouter.Router` implements `http.Handler`, but it registers routes through its own `router.GET/POST/Handle` methods and bypasses net/http's `HandleFunc`/`ServeMux`. So WhaTap instruments it where the router is handed to the server:

```go
http.ListenAndServe(":8080", router)                        // Before
http.ListenAndServe(":8080", whataphttp.WrapHandler(router)) // After
```

`whataphttp.WrapHandler` (part of the core `github.com/whatap/go-api` module, Go 1.18) starts a transaction for every request — **no dedicated httprouter module is needed**.

- **Automatic Transaction Tracking**: each HTTP request creates a transaction
- **Request/Response Details**: URL, method, status code, response time
- **Distributed Tracing**: trace context propagation via headers

## Quick Start

### Option A — manual (wrap at ListenAndServe)

```go
import (
    "github.com/julienschmidt/httprouter"

    "github.com/whatap/go-api/instrumentation/net/http/whataphttp"
    "github.com/whatap/go-api/trace"
)

func main() {
    trace.Init(nil)
    defer trace.Shutdown()

    router := httprouter.New()
    router.GET("/", handler)

    http.ListenAndServe(":8080", whataphttp.WrapHandler(router))
}
```

Same idea when using `http.Server`:

```go
srv := &http.Server{Addr: ":8080", Handler: whataphttp.WrapHandler(router)}
srv.ListenAndServe()
```

### Option B — Auto-instrumentation (recommended)

Let `whatap-go-inst` wrap the handler automatically at build time (§100 `net/http.ListenAndServe` ArgWrap rule) — no code changes:

```bash
whatap-go-inst go build ./...
```

### Access the transaction context

```go
router.GET("/users/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    ctx := r.Context() // transaction started by WrapHandler
    // db.QueryContext(ctx, ...) linked to this transaction
})
```

## Note on route naming

Transactions are named by the raw request path. httprouter's route pattern (`/users/:id`) is not currently used for naming; this does not affect transaction capture.

## Running the Example

```bash
make local
go run router.go

curl http://localhost:8080/
curl http://localhost:8080/health
```

## See Also

- [WhaTap Go API Documentation](https://docs.whatap.io/golang/install-agent)
- [net/http Example](../../../net/http/server/)
- [go-restful Example](../../emicklei/go-restful/v3/)
