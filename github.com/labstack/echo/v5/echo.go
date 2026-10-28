// Echo v5 Framework Instrumentation Example
//
// This example demonstrates how to instrument Echo v5 web framework
// applications using the WhaTap Go API.
//
// ECHO v5 AND THE nested whatapecho/v5 MODULE:
// Echo v5 requires Go 1.25+ and changed `echo.Context` from an interface to a
// struct (handlers receive `*echo.Context`). Because of that Go version
// requirement the echo/v5 helper ships as a SEPARATE nested module
//   github.com/whatap/go-api/instrumentation/github.com/labstack/echo/v5/whatapecho
// (go 1.25) so the core go-api module stays on Go 1.18. Manual usage is the same
// one-liner as Echo v4:
//
//   e.Use(whatapecho.Middleware())   // or whatapecho.WrapEcho(e)
//
// INSTRUMENTATION OVERVIEW:
// 1. Call trace.Init() at application startup
// 2. Add whatapecho.Middleware() with e.Use()
// 3. Use c.Request().Context() to get the transaction context in handlers
//
// FEATURES:
// - Automatic HTTP transaction tracking
// - Request/Response details (URL, method, status code, response time)
// - Error tracking (HTTP status captured even on 4xx/5xx)
// - Distributed tracing via header propagation
//
// AUTO-INSTRUMENTATION ALTERNATIVE:
// whatap-go-inst injects equivalent middleware automatically at build time
// (via an inline Transform rule, no dependency needed) — no code changes:
//   whatap-go-inst go build ./...
// See the go-api-inst docs.

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"time"

	echo "github.com/labstack/echo/v5"

	"github.com/whatap/go-api/instrumentation/github.com/labstack/echo/v5/whatapecho"
	"github.com/whatap/go-api/method"
	"github.com/whatap/go-api/trace"
)

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// getUser demonstrates custom method tracing with the method package.
func getUser(ctx context.Context) {
	methodCtx, _ := method.Start(ctx, "getUser")
	defer method.End(methodCtx, nil)
	time.Sleep(100 * time.Millisecond)
}

func main() {
	port := flag.Int("p", 8080, "web port (default 8080)")
	udpPort := flag.Int("up", 6600, "agent UDP port (default 6600)")
	useWhatap := flag.Bool("whatap", false, "enable WhaTap instrumentation")
	flag.Parse()

	// ============================================================
	// INSTRUMENTATION STEP 1: Initialize the WhaTap agent
	// ============================================================
	if *useWhatap {
		config := map[string]string{
			"net_udp_port": fmt.Sprintf("%d", *udpPort),
		}
		trace.Init(config)
	}
	defer trace.Shutdown()

	e := echo.New()

	// ============================================================
	// INSTRUMENTATION STEP 2: Register the WhaTap middleware
	// ============================================================
	// whatapecho.Middleware() traces every request (transaction start/end,
	// status code, error). Add it before your other middlewares.
	if *useWhatap {
		e.Use(whatapecho.Middleware())
	}

	// Basic handler.
	e.GET("/", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Hello, Echo v5!"})
	})

	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.GET("/products", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, []Product{
			{ID: 1, Name: "Laptop", Price: 999.99},
			{ID: 2, Name: "Phone", Price: 599.99},
		})
	})

	// ============================================================
	// INSTRUMENTATION STEP 3: Access the transaction context
	// ============================================================
	// c.Request().Context() carries the transaction started by the
	// middleware. Pass it to SQL / HTTP-client / custom-method calls so
	// they are linked to this HTTP transaction.
	e.GET("/products/:id", func(c *echo.Context) error {
		ctx := c.Request().Context()

		// Add a custom step to the transaction trace.
		trace.Step(ctx, "lookup", "product lookup", 1, 1)

		// Call a custom method - traced via the method package.
		getUser(ctx)

		id := c.Param("id")
		return c.JSON(http.StatusOK, map[string]string{"id": id, "name": "Product " + id})
	})

	// Error handling: returning an echo.HTTPError makes the middleware
	// record the status code (500 here) on the transaction.
	e.GET("/error", func(c *echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong")
	})

	fmt.Printf("Start :%d, Agent UDP Port: %d\n", *port, *udpPort)
	if err := e.Start(fmt.Sprintf(":%d", *port)); err != nil {
		panic(err)
	}
}
