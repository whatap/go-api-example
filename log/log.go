// log Package Instrumentation Example
//
// INSTRUMENTATION OVERVIEW:
// =========================
// This example demonstrates WhaTap instrumentation for Go standard log package.
// log.SetOutput is used to redirect log output through WhaTap's TraceLogWriter
// for transaction-linked logging.
//
// BEFORE (Original):
//   log.Println("Hello")
//   log.Printf("Count: %d", 42)
//
// AFTER (Instrumented - Auto with whatap-go-inst):
//   trace.Init(nil)
//   log.SetOutput(logsink.GetTraceLogWriter(&logsink.LogConfig{...}))
//   log.Println("Hello")  // Now linked to transaction via goroutine ID
//
// FEATURES:
// - Transaction-linked logging (txid, mtid, gid)
// - Works with existing log.Print/Printf/Println calls
// - No changes needed to logging code itself
//
// CONFIGURATION:
//   # whatap.conf
//   logsink_enabled=true

package main

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/whatap/go-api/instrumentation/net/http/whataphttp"
	"github.com/whatap/go-api/logsink"
	"github.com/whatap/go-api/trace"
)

func main() {
	// ============================================================
	// INSTRUMENTATION STEP 1: Initialize WhaTap Agent
	// ============================================================
	trace.Init(nil)
	defer trace.Shutdown()

	// ============================================================
	// INSTRUMENTATION STEP 2: Redirect log output to TraceLogWriter
	// ============================================================
	// This enables transaction-linked logging via goroutine ID.
	// Logs are collected by WhaTap logsink with txid, mtid, gid context.
	log.SetOutput(logsink.GetTraceLogWriter(os.Stdout))

	log.Println("Server starting...")
	log.Printf("Listening on port %d", 8080)

	// ============================================================
	// HTTP Handlers with log output
	// ============================================================
	http.HandleFunc("/", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] Request: %s %s", time.Now().Format(time.RFC3339), r.Method, r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Hello from log example"}`))

		log.Println("Response sent")
	}))

	http.HandleFunc("/log-test", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		// Test different log functions
		log.Print("This is Print (with timestamp)")
		log.Println("This is Println")
		log.Printf("This is Printf with args: method=%s, path=%s", r.Method, r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "logs generated - check WhaTap logsink"}`))
	}))

	http.HandleFunc("/health", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	}))

	// ============================================================
	// Stack Trace Logging Test (Issue §29 Verification)
	// ============================================================
	// This endpoint tests how logsink handles multi-line stack traces.
	// Expected behavior: Stack trace lines should be grouped together.
	http.HandleFunc("/stack", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		// 1. Simulated stack trace (manual format)
		// First line has no tab, subsequent lines have tabs
		log.Println("Error: simulated error occurred")
		log.Println("\tat github.com/example/app.handler()")
		log.Println("\tat github.com/example/app.middleware()")
		log.Println("\tat main.main()")

		// 2. Real stack trace using runtime.Stack
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		log.Printf("Real stack trace:\n%s", buf[:n])

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "stack traces generated - check logsink"}`))
	}))

	log.Println("Server ready")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Printf("Server failed: %v", err)
	}
}
