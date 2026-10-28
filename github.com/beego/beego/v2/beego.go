// Beego v2 Framework Instrumentation Example
//
// This example demonstrates how to instrument a Beego v2 web application
// using the WhaTap Go API.
//
// WHY NO DEDICATED whatapbeego PACKAGE:
// Beego's middleware type is `MiddleWare func(http.Handler) http.Handler` —
// exactly the net/http middleware shape. So the existing core wrapper
//   github.com/whatap/go-api/instrumentation/net/http/whataphttp   (Go 1.18)
// plugs straight into Beego's RunWithMiddleWares as a middleware. No Beego
// import is pulled into go-api and no nested module is needed (unlike Echo v5 /
// Fiber v3, whose middleware signatures differ from net/http).
//
// INSTRUMENTATION OVERVIEW:
// 1. Call trace.Init() at application startup
// 2. Start the server with web.RunWithMiddleWares(addr, whataphttp.WrapHandler)
//    instead of web.Run(addr)
// 3. Use c.Ctx.Request.Context() to get the transaction context in a Controller
//
// AUTO-INSTRUMENTATION ALTERNATIVE:
// whatap-go-inst rewrites web.Run(...) to
//   web.RunWithMiddleWares(<addr or "">, whataphttp.WrapHandler)
// automatically at build time (§204 inline Transform rule) — no code changes:
//   whatap-go-inst go build ./...

package main

import (
	"github.com/beego/beego/v2/server/web"

	"github.com/whatap/go-api/instrumentation/net/http/whataphttp"
	"github.com/whatap/go-api/trace"
)

type MainController struct {
	web.Controller
}

func (c *MainController) Get() {
	// INSTRUMENTATION: the transaction started by whataphttp.WrapHandler is
	// carried on the request context — pass it to SQL / HTTP-client / method calls.
	ctx := c.Ctx.Request.Context()
	trace.Step(ctx, "lookup", "hello handler", 1, 1)

	c.Data["json"] = map[string]string{"message": "Hello, Beego v2!"}
	_ = c.ServeJSON()
}

type HealthController struct {
	web.Controller
}

func (c *HealthController) Get() {
	c.Data["json"] = map[string]string{"status": "ok"}
	_ = c.ServeJSON()
}

func main() {
	// ============================================================
	// INSTRUMENTATION STEP 1: Initialize the WhaTap agent
	// ============================================================
	trace.Init(nil)
	defer trace.Shutdown()

	web.Router("/", &MainController{})
	web.Router("/health", &HealthController{})

	// ============================================================
	// INSTRUMENTATION STEP 2: Start with the WhaTap middleware
	// ============================================================
	// BEFORE (Original):
	//   web.Run(":8080")
	// AFTER (Instrumented): wrap the handler chain with whataphttp.WrapHandler.
	// whataphttp.WrapHandler has signature func(http.Handler) http.Handler, which
	// is exactly Beego's MiddleWare type, so it is passed directly.
	web.RunWithMiddleWares(":8080", whataphttp.WrapHandler)
}
