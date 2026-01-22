// fmt Package Instrumentation Example
//
// INSTRUMENTATION OVERVIEW:
// =========================
// This example demonstrates WhaTap instrumentation for fmt.Print functions.
// fmt.Print/Printf/Println are transformed to whatapfmt equivalents for
// transaction-linked stdout logging.
//
// BEFORE (Original):
//   fmt.Println("Hello")
//   fmt.Printf("Count: %d\n", 42)
//
// AFTER (Instrumented - Manual):
//   whatapfmt.Println("Hello")
//   whatapfmt.Printf("Count: %d\n", 42)
//
// AFTER (Instrumented - Auto with whatap-go-inst):
//   # Automatically transformed, no code changes needed
//   whatap-go-inst go build ./...
//
// FEATURES:
// - Transaction-linked stdout logging (txid, mtid, gid)
// - Same function signatures as fmt package
// - Only Print/Printf/Println are transformed
// - Sprint/Sprintf/Sprintln are NOT transformed (they return strings, not output)
//
// CONFIGURATION:
//   # whatap.conf
//   logsink_enabled=true
//   logsink_fmt_enabled=true

package main

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/whatap/go-api/instrumentation/fmt/whatapfmt"
	"github.com/whatap/go-api/instrumentation/net/http/whataphttp"
	"github.com/whatap/go-api/trace"
)

func main() {
	// ============================================================
	// INSTRUMENTATION STEP 1: Initialize WhaTap Agent
	// ============================================================
	// Call trace.Init() at application startup.
	// Configuration can be passed via map or loaded from whatap.conf file.
	trace.Init(nil)
	defer trace.Shutdown()

	// ============================================================
	// INSTRUMENTATION STEP 2: Use whatapfmt instead of fmt
	// ============================================================
	// Replace fmt.Print/Printf/Println with whatapfmt equivalents.
	// The output goes to stdout AND is collected by WhaTap logsink
	// with transaction context (txid, mtid, gid).

	whatapfmt.Println("Server starting...")
	whatapfmt.Printf("Listening on port %d\n", 8080)

	// ============================================================
	// HTTP Handler with whatapfmt logging
	// ============================================================
	// Inside HTTP handlers, whatapfmt automatically links logs
	// to the current transaction via goroutine ID.

	http.HandleFunc("/", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		// These logs are linked to the HTTP transaction
		whatapfmt.Printf("[%s] Request: %s %s\n", time.Now().Format(time.RFC3339), r.Method, r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Hello from whatapfmt example"}`))

		whatapfmt.Println("Response sent")
	}))

	http.HandleFunc("/print-test", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		// Test different print functions
		whatapfmt.Print("This is Print (no newline)")
		whatapfmt.Println(" - followed by Println")
		whatapfmt.Printf("This is Printf with args: method=%s, path=%s\n", r.Method, r.URL.Path)

		// fmt.Sprintf is NOT transformed - it returns a string, doesn't output
		msg := fmt.Sprintf("Formatted message: %s", r.URL.Path)
		whatapfmt.Println(msg)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "prints generated - check logs"}`))
	}))

	http.HandleFunc("/health", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	}))

	// ============================================================
	// Stack Trace Logging Test (§29 Issue Verification)
	// ============================================================
	// This endpoint tests how logsink handles multi-line stack traces.
	// Expected behavior: Stack trace lines should be grouped together.
	http.HandleFunc("/stack", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		// 1. Simulated stack trace (manual format)
		// First line has no tab, subsequent lines have tabs
		whatapfmt.Println("Error: simulated error occurred")
		whatapfmt.Println("\tat github.com/example/app.handler()")
		whatapfmt.Println("\tat github.com/example/app.middleware()")
		whatapfmt.Println("\tat main.main()")

		// 2. Real stack trace using runtime.Stack
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		whatapfmt.Printf("Real stack trace:\n%s\n", buf[:n])

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "stack traces generated - check logsink"}`))
	}))

	whatapfmt.Println("Server ready")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		whatapfmt.Printf("Server failed: %v\n", err)
	}
}
