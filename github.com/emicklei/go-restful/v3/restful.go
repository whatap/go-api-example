// emicklei/go-restful/v3 Instrumentation Example
//
// This example demonstrates how to instrument a go-restful application using
// the WhaTap Go API.
//
// WHY ListenAndServe WRAPPING:
// go-restful's Container implements http.Handler (via ServeHTTP) but registers
// routes through its own WebService/Route API — it bypasses net/http's
// HandleFunc/ServeMux registration. So the Container is wrapped where it is
// handed to the server:
//
//   http.ListenAndServe(addr, container)                        // Before
//   http.ListenAndServe(addr, whataphttp.WrapHandler(container)) // After
//
// whataphttp.WrapHandler (net/http, Go 1.18) starts a transaction for every
// request the Container serves. No dedicated go-restful module is needed.
// (The same applies to restful.DefaultContainer served via http.ListenAndServe.)
//
// AUTO-INSTRUMENTATION ALTERNATIVE:
// whatap-go-inst rewrites http.ListenAndServe(addr, handler) automatically at
// build time (§95 ArgWrap rule) — no code changes:
//   whatap-go-inst go build ./...

package main

import (
	"net/http"

	restful "github.com/emicklei/go-restful/v3"

	"github.com/whatap/go-api/instrumentation/net/http/whataphttp"
	"github.com/whatap/go-api/trace"
)

func main() {
	// INSTRUMENTATION STEP 1: initialize the WhaTap agent.
	trace.Init(nil)
	defer trace.Shutdown()

	ws := new(restful.WebService)
	ws.Route(ws.GET("/").To(func(req *restful.Request, resp *restful.Response) {
		// The transaction started by WrapHandler is on the request context.
		trace.Step(req.Request.Context(), "lookup", "hello handler", 1, 1)
		_ = resp.WriteAsJson(map[string]string{"message": "Hello, go-restful!"})
	}))
	ws.Route(ws.GET("/health").To(func(req *restful.Request, resp *restful.Response) {
		_ = resp.WriteAsJson(map[string]string{"status": "ok"})
	}))

	container := restful.NewContainer()
	container.Add(ws)

	// INSTRUMENTATION STEP 2: wrap the Container with whataphttp.WrapHandler.
	http.ListenAndServe(":8080", whataphttp.WrapHandler(container))
}
