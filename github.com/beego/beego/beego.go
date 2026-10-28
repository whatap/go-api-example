// Beego v1 Framework Instrumentation Example
//
// This example demonstrates how to instrument a Beego v1 web application
// (github.com/beego/beego, v1.12.x — the pre-rebrand root package) using the
// WhaTap Go API.
//
// WHY NO DEDICATED whatapbeego PACKAGE:
// Beego's middleware type is `MiddleWare func(http.Handler) http.Handler` —
// exactly the net/http middleware shape (identical in v1 and v2). So the
// existing core wrapper
//   github.com/whatap/go-api/instrumentation/net/http/whataphttp   (Go 1.18)
// plugs straight into Beego's RunWithMiddleWares as a middleware. No Beego
// import is pulled into go-api and no nested module is needed.
//
// INSTRUMENTATION OVERVIEW:
// 1. Call trace.Init() at application startup
// 2. Start the server with beego.RunWithMiddleWares(addr, whataphttp.WrapHandler)
//    instead of beego.Run(addr)
// 3. Use c.Ctx.Request.Context() to get the transaction context in a Controller
//
// AUTO-INSTRUMENTATION ALTERNATIVE:
// whatap-go-inst rewrites beego.Run(...) to
//   beego.RunWithMiddleWares(<addr or "">, whataphttp.WrapHandler)
// automatically at build time (§50/§204 inline Transform rule) — no code changes:
//   whatap-go-inst go build ./...

package main

import (
	"github.com/beego/beego"

	"github.com/whatap/go-api/instrumentation/net/http/whataphttp"
	"github.com/whatap/go-api/trace"
)

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	// INSTRUMENTATION: the transaction started by whataphttp.WrapHandler is
	// carried on the request context — pass it to SQL / HTTP-client / method calls.
	ctx := c.Ctx.Request.Context()
	trace.Step(ctx, "lookup", "hello handler", 1, 1)

	c.Data["json"] = map[string]string{"message": "Hello, Beego v1!"}
	// Beego v1's ServeJSON() returns no value (v2 returns error).
	c.ServeJSON()
}

type HealthController struct {
	beego.Controller
}

func (c *HealthController) Get() {
	c.Data["json"] = map[string]string{"status": "ok"}
	c.ServeJSON()
}

func main() {
	// ============================================================
	// INSTRUMENTATION STEP 1: Initialize the WhaTap agent
	// ============================================================
	trace.Init(nil)
	defer trace.Shutdown()

	beego.Router("/", &MainController{})
	beego.Router("/health", &HealthController{})

	// ============================================================
	// INSTRUMENTATION STEP 2: Start with the WhaTap middleware
	// ============================================================
	// BEFORE (Original):
	//   beego.Run(":8080")
	// AFTER (Instrumented): wrap the handler chain with whataphttp.WrapHandler.
	// whataphttp.WrapHandler has signature func(http.Handler) http.Handler, which
	// is exactly Beego's MiddleWare type, so it is passed directly.
	beego.RunWithMiddleWares(":8080", whataphttp.WrapHandler)
}
