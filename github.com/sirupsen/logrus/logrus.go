// logrus Package Instrumentation Example
//
// INSTRUMENTATION OVERVIEW:
// =========================
// This example demonstrates WhaTap instrumentation for logrus logging library.
// WhaTap uses logrus Hook pattern to capture logs without interfering with
// the application's logrus configuration.
//
// BEFORE (Original):
//   logrus.Info("Hello")
//   logrus.WithFields(logrus.Fields{"user": "john"}).Info("Login")
//
// AFTER (Instrumented - Auto with whatap-go-inst):
//   import _ "github.com/whatap/go-api/instrumentation/github.com/sirupsen/logrus/whataplogrus"
//   logrus.Info("Hello")  // Now captured by WhaTap via Hook
//
// KEY BENEFITS:
// - Hook pattern: Does not override app's logrus configuration
// - Auto-registration: Blank import triggers init() which registers Hook
// - Transaction linking: Logs are linked to active transactions via goroutine ID
// - No code changes: Works with existing logrus.Info/Warn/Error/etc calls
//
// CONFIGURATION:
//   # whatap.conf
//   logsink_enabled=true
//   logsink_trace_enabled=true     # Enable txid/mtid linking
//   logsink_trace_txid_enabled=true
//   logsink_trace_mtid_enabled=true

package main

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/whatap/go-api/instrumentation/net/http/whataphttp"
	"github.com/whatap/go-api/trace"

	// ============================================================
	// INSTRUMENTATION: Blank import auto-registers WhaTap Hook
	// ============================================================
	// METHOD 1: Blank import triggers init() in whataplogrus package which:
	// 1. Uses sync.Once to ensure single registration
	// 2. Calls logrus.AddHook() to register WhaTap Hook
	// 3. Hook captures all log levels and sends to logsink
	//
	// METHOD 2: For custom logrus.Logger instances (logrus.New()),
	// use whataplogrus.WrapLogger() to register Hook on each instance.
	"github.com/whatap/go-api/instrumentation/github.com/sirupsen/logrus/whataplogrus"
)

func main() {
	// ============================================================
	// INSTRUMENTATION: Initialize WhaTap Agent
	// ============================================================
	trace.Init(nil)
	defer trace.Shutdown()

	// ============================================================
	// METHOD 2: WrapLogger() for custom logrus.Logger instances
	// ============================================================
	// When using logrus.New() to create separate logger instances,
	// the blank import (init() Hook) only applies to the global logger.
	// Use whataplogrus.WrapLogger() to instrument custom instances.
	//
	// BEFORE (Original):
	//   logger := logrus.New()
	//
	// AFTER (Instrumented):
	//   logger := whataplogrus.WrapLogger(logrus.New())
	customLogger := whataplogrus.WrapLogger(logrus.New())
	customLogger.SetFormatter(&logrus.JSONFormatter{})
	customLogger.SetLevel(logrus.DebugLevel)
	customLogger.Info("Custom logger initialized with WrapLogger")

	// ============================================================
	// Application's logrus configuration (unaffected by WhaTap)
	// ============================================================
	// WhaTap Hook pattern respects the app's logrus settings.
	// These settings won't affect WhaTap's log collection.
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.DebugLevel)

	logrus.Info("Server starting...")
	logrus.WithFields(logrus.Fields{
		"port":    8080,
		"version": "1.0.0",
	}).Info("Configuration loaded")

	// ============================================================
	// HTTP Handlers with logrus logging
	// ============================================================
	http.HandleFunc("/", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		logrus.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
		}).Info("Request received")

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Hello from logrus example"}`))

		logrus.Debug("Response sent")
	}))

	http.HandleFunc("/log-test", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		// Test different log levels - all captured by WhaTap Hook
		logrus.Trace("This is Trace level")
		logrus.Debug("This is Debug level")
		logrus.Info("This is Info level")
		logrus.Warn("This is Warn level")
		logrus.Error("This is Error level")

		// Test structured logging with fields
		logrus.WithFields(logrus.Fields{
			"user_id":    123,
			"session_id": "abc-xyz",
			"action":     "log-test",
		}).Info("Structured log with fields")

		// Test error logging with error field
		logrus.WithError(nil).Warn("Warning with error field")

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "logs generated - check WhaTap logsink"}`))
	}))

	// ============================================================
	// Custom Logger Instance with WrapLogger
	// ============================================================
	// Logs from customLogger are also captured by WhaTap,
	// even though it's a separate logrus.Logger instance.
	http.HandleFunc("/custom-logger", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		customLogger.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		}).Info("Request handled by custom logger")

		customLogger.Debug("Custom logger debug message")

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Logged via custom logger with WrapLogger"}`))
	}))

	http.HandleFunc("/health", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	}))

	// ============================================================
	// Simulated background task with logging
	// ============================================================
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			logrus.WithFields(logrus.Fields{
				"timestamp": time.Now().Format(time.RFC3339),
			}).Debug("Background task heartbeat")
		}
	}()

	logrus.Info("Server ready")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logrus.WithError(err).Fatal("Server failed to start")
	}
}
