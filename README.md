# WhaTap Go API Instrumentation Examples

This repository contains example code demonstrating how to instrument Go applications using the [WhaTap Go API](https://github.com/whatap/go-api).

## Quick Start

All examples follow a common instrumentation pattern:

1. **Initialize** - Call `trace.Init(nil)` at application startup
2. **Add Middleware** - Add WhaTap middleware to your web framework
3. **Use Context** - Pass context to track SQL, HTTP calls, and custom methods
4. **Shutdown** - Call `defer trace.Shutdown()` for graceful cleanup

## Supported Libraries

### Web Frameworks

| Library | Example | Middleware Pattern |
|---------|---------|-------------------|
| [Gin](./github.com/gin-gonic/gin) | `github.com/gin-gonic/gin/` | `r.Use(whatapgin.Middleware())` |
| [Echo v4](./github.com/labstack/echo/v4) | `github.com/labstack/echo/v4/` | `e.Use(whatapecho.Middleware())` |
| [Echo v3](./github.com/labstack/echo) | `github.com/labstack/echo/` | `e.Use(whatapecho.Middleware())` |
| [Fiber v2](./github.com/gofiber/fiber) | `github.com/gofiber/fiber/` | `app.Use(whatapfiber.Middleware())` |
| [Chi](./github.com/go-chi/chi) | `github.com/go-chi/chi/` | `r.Use(whatapchi.Middleware)` (no parentheses) |
| [Gorilla Mux](./github.com/gorilla/mux) | `github.com/gorilla/mux/` | `r.Use(whatapmux.Middleware())` |
| [FastHTTP](./github.com/valyala/fasthttp) | `github.com/valyala/fasthttp/` | `whatapfasthttp.Func(handler)` wrapper |
| [net/http](./net/http/server) | `net/http/server/` | `whataphttp.Func()/Handler()` wrapper |

### Databases

| Library | Example | Instrumentation |
|---------|---------|-----------------|
| [database/sql](./database/sql) | `database/sql/` | `whatapsql.OpenContext()` or manual `Start/End` |
| [GORM v2](./github.com/go-gorm/gorm) | `github.com/go-gorm/gorm/` | `whatapgorm.OpenWithContext()` |
| [GORM v1 (jinzhu)](./github.com/jinzhu/gorm) | `github.com/jinzhu/gorm/` | `whatapgorm.Open()` |
| [sqlx](./github.com/jmoiron/sqlx) | `github.com/jmoiron/sqlx/` | `whatapsqlx.OpenContext()` |

### Cache & NoSQL

| Library | Example | Instrumentation |
|---------|---------|-----------------|
| [go-redis v9](./github.com/redis/go-redis/v9) | `github.com/redis/go-redis/v9/` | `whatapgoredis.NewClient()` |
| [go-redis v8](./github.com/go-redis/redis/v8) | `github.com/go-redis/redis/v8/` | `whatapgoredis.NewClient()` |
| [Redigo](./github.com/gomodule/redigo) | `github.com/gomodule/redigo/` | `whatapredigo.DialContext()` |
| [MongoDB](./go.mongodb.org/mongo-driver/mongo) | `go.mongodb.org/mongo-driver/mongo/` | `whatapmongo.Connect()` |

### Message Queues

| Library | Example | Instrumentation |
|---------|---------|-----------------|
| [IBM/sarama (Kafka)](./github.com/IBM/sarama) | `github.com/IBM/sarama/` | Producer/Consumer interceptors |
| [Shopify/sarama (Kafka)](./github.com/Shopify/sarama) | `github.com/Shopify/sarama/` | Producer/Consumer interceptors |

### HTTP Client

| Library | Example | Instrumentation |
|---------|---------|-----------------|
| [net/http client](./net/http/client) | `net/http/client/` | `httpc.Start/End` or `whataphttp.NewRoundTrip()` |

### Cloud & RPC

| Library | Example | Instrumentation |
|---------|---------|-----------------|
| [gRPC](./google.golang.org/grpc) | `google.golang.org/grpc/` | `whatapgrpc` interceptors |
| [Kubernetes client-go](./k8s.io/client-go/kubernetes) | `k8s.io/client-go/kubernetes/` | `whatapkubernetes.NewForConfig()` |

### Logging

| Library | Example | Instrumentation |
|---------|---------|-----------------|
| [fmt](./fmt) | `fmt/` | `whatapfmt.Print/Printf/Println()` |

## Key Features

- **Automatic HTTP transaction tracking** - Request/Response details
- **SQL query tracking** - Query text, parameters, execution time
- **HTTP client call tracking** - Outgoing HTTP calls with `httpc.Start/End`
- **Custom method tracing** - Track any function with `method.Start/End`
- **Distributed tracing** - Multi-transaction trace via `trace.GetMTrace(ctx)`
- **Error tracking** - Automatic error capture and reporting

## Running Examples

Each example can be run with:

```bash
cd <example-directory>
go run *.go -whatap
```

## Related Links

- [WhaTap Go API](https://github.com/whatap/go-api) - Main instrumentation library
- [WhaTap Documentation](https://docs.whatap.io/) - Official documentation
