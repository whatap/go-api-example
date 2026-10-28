// julienschmidt/httprouter Instrumentation Example
//
// This example demonstrates how to instrument an httprouter application using
// the WhaTap Go API.
//
// WHY ListenAndServe WRAPPING:
// httprouter.Router implements http.Handler (via ServeHTTP) but registers its
// routes through its own router.GET/POST/Handle methods — it bypasses net/http's
// HandleFunc/ServeMux registration. So the per-handler wrappers don't apply.
// Instead, the router is wrapped where it is handed to the server:
//
//   http.ListenAndServe(addr, router)                       // Before
//   http.ListenAndServe(addr, whataphttp.WrapHandler(router)) // After
//
// whataphttp.WrapHandler (net/http, Go 1.18) starts a transaction for every
// request the router serves. No dedicated httprouter module is needed.
//
// AUTO-INSTRUMENTATION ALTERNATIVE:
// whatap-go-inst rewrites http.ListenAndServe(addr, handler) automatically at
// build time (§100 ArgWrap rule) — no code changes:
//   whatap-go-inst go build ./...

package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/whatap/go-api/instrumentation/net/http/whataphttp"
	"github.com/whatap/go-api/trace"
)

func main() {
	// INSTRUMENTATION STEP 1: initialize the WhaTap agent.
	trace.Init(nil)
	defer trace.Shutdown()

	router := httprouter.New()

	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// The transaction started by WrapHandler is on r.Context().
		trace.Step(r.Context(), "lookup", "hello handler", 1, 1)
		fmt.Fprint(w, `{"message":"Hello, httprouter!"}`)
	})
	router.GET("/health", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	// INSTRUMENTATION STEP 2: wrap the router with whataphttp.WrapHandler.
	http.ListenAndServe(":8080", whataphttp.WrapHandler(router))
}
