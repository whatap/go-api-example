// Package main demonstrates WhaTap APM instrumentation for Aerospike Go Client.
//
// This example shows how to monitor Aerospike operations using the WhaTap Go API.
// Since Aerospike client doesn't provide Hook/Monitor interface like MongoDB or Redis,
// we use the sql.Wrap pattern for manual instrumentation.
//
// Key instrumentation points:
//   1. Use sql.WrapOpen() for client connection
//   2. Use sql.Wrap() for methods returning (T, error)
//   3. Use sql.WrapError() for methods returning only error
//   4. Pass context from HTTP transaction to Wrap functions
//
// WhaTap logs generated:
//   - [WA-SQL-04002] Open DB: Connection establishment tracking
//   - [WA-SQL-04003] Sql: Aerospike command execution tracking
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	// Aerospike client
	aerospike "github.com/aerospike/aerospike-client-go/v6"

	// WhaTap instrumentation packages
	// - sql: Provides Wrap functions for database operations
	// - trace: Core tracing API for transaction management
	"github.com/whatap/go-api/sql"
	"github.com/whatap/go-api/trace"
)

var client *aerospike.Client

func main() {
	// Command-line flags for configuration
	udpPortPtr := flag.Int("up", 6600, "WhaTap agent UDP port (default: 6600)")
	portPtr := flag.Int("p", 8104, "HTTP server port (default: 8104)")
	asHostPtr := flag.String("ashost", "localhost", "Aerospike host")
	asPortPtr := flag.Int("asport", 3000, "Aerospike port")
	setWhatapPtr := flag.Bool("whatap", false, "Enable WhaTap monitoring")

	flag.Parse()
	port := *portPtr
	udpPort := *udpPortPtr
	asHost := *asHostPtr
	asPort := *asPortPtr
	IsWhatap := *setWhatapPtr

	// ============================================================
	// INSTRUMENTATION STEP 1: Initialize WhaTap Agent
	// ============================================================
	if IsWhatap {
		config := make(map[string]string)
		config["net_udp_port"] = fmt.Sprintf("%d", udpPort)
		trace.Init(config)
	}
	defer trace.Shutdown()

	// ============================================================
	// INSTRUMENTATION STEP 2: Use sql.WrapOpen() for Connection
	// ============================================================
	// sql.WrapOpen() tracks connection establishment.
	// Parameters:
	//   - ctx: context for tracing
	//   - dbType: database type identifier ("aerospike")
	//   - connection: connection string for display
	//   - fn: function that returns (*T, error)
	//
	// Note: When called outside HTTP handler, txid will be 0.
	var err error
	connection := fmt.Sprintf("aerospike://%s:%d", asHost, asPort)
	client, err = sql.WrapOpen(context.Background(), connection,
		func() (*aerospike.Client, error) {
			return aerospike.NewClient(asHost, asPort)
		})
	if err != nil {
		fmt.Printf("Warning: Failed to connect to Aerospike: %v\n", err)
		// Continue anyway for testing without Aerospike server
	}

	// ============================================================
	// HTTP HANDLER EXAMPLES
	// ============================================================

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		json.NewEncoder(w).Encode(map[string]string{
			"message":   "Aerospike Example API",
			"endpoints": "/put, /get, /delete, /exists, /batch, /health",
		})
	})

	// ============================================================
	// CASE 1: Put - Store a Record
	// ============================================================
	// Demonstrates: sql.WrapError() for methods returning only error.
	// The Put method returns only error, so use WrapError.
	http.HandleFunc("/put", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		if client == nil {
			http.Error(w, "Not connected to Aerospike", http.StatusServiceUnavailable)
			return
		}

		name := r.URL.Query().Get("name")
		if name == "" {
			name = "default"
		}
		value := r.URL.Query().Get("value")

		key, keyErr := aerospike.NewKey("test", "demo", name)
		if keyErr != nil {
			trace.Error(ctx, keyErr)
			http.Error(w, keyErr.Error(), http.StatusInternalServerError)
			return
		}

		bins := aerospike.BinMap{
			"name":  name,
			"value": value,
		}

		// ============================================================
		// INSTRUMENTATION: sql.WrapError() for error-only returns
		// ============================================================
		// Parameters:
		//   - ctx: traced context from HTTP transaction
		//   - dbType: "aerospike"
		//   - methodName: "Put" (displayed in trace)
		//   - fn: function that returns error
		err = sql.WrapError(ctx, "aerospike", "Put", func() error {
			return client.Put(nil, key, bins)
		})
		if err != nil {
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"key":    name,
		})
	})

	// ============================================================
	// CASE 2: Get - Retrieve a Record
	// ============================================================
	// Demonstrates: sql.Wrap() for methods returning (T, error).
	// The Get method returns (*Record, error), so use Wrap.
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		if client == nil {
			http.Error(w, "Not connected to Aerospike", http.StatusServiceUnavailable)
			return
		}

		name := r.URL.Query().Get("name")
		if name == "" {
			name = "default"
		}

		key, keyErr := aerospike.NewKey("test", "demo", name)
		if keyErr != nil {
			trace.Error(ctx, keyErr)
			http.Error(w, keyErr.Error(), http.StatusInternalServerError)
			return
		}

		// ============================================================
		// INSTRUMENTATION: sql.Wrap() for (T, error) returns
		// ============================================================
		// Parameters:
		//   - ctx: traced context from HTTP transaction
		//   - dbType: "aerospike"
		//   - methodName: "Get" (displayed in trace)
		//   - fn: function that returns (T, error)
		//
		// Note: The return type *aerospike.Record is preserved.
		record, err := sql.Wrap(ctx, "aerospike", "Get", func() (*aerospike.Record, error) {
			return client.Get(nil, key)
		})
		if err != nil {
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if record == nil {
			http.Error(w, "Record not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"bins":   record.Bins,
		})
	})

	// ============================================================
	// CASE 3: Delete - Remove a Record
	// ============================================================
	// Demonstrates: sql.Wrap() for Delete returning (bool, error).
	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		if client == nil {
			http.Error(w, "Not connected to Aerospike", http.StatusServiceUnavailable)
			return
		}

		name := r.URL.Query().Get("name")
		if name == "" {
			name = "default"
		}

		key, keyErr := aerospike.NewKey("test", "demo", name)
		if keyErr != nil {
			trace.Error(ctx, keyErr)
			http.Error(w, keyErr.Error(), http.StatusInternalServerError)
			return
		}

		// Delete returns (bool, error) - existed flag
		existed, err := sql.Wrap(ctx, "aerospike", "Delete", func() (bool, error) {
			return client.Delete(nil, key)
		})
		if err != nil {
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"existed": existed,
		})
	})

	// ============================================================
	// CASE 4: Exists - Check Record Existence
	// ============================================================
	// Demonstrates: sql.Wrap() for Exists returning (bool, error).
	http.HandleFunc("/exists", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		if client == nil {
			http.Error(w, "Not connected to Aerospike", http.StatusServiceUnavailable)
			return
		}

		name := r.URL.Query().Get("name")
		if name == "" {
			name = "default"
		}

		key, keyErr := aerospike.NewKey("test", "demo", name)
		if keyErr != nil {
			trace.Error(ctx, keyErr)
			http.Error(w, keyErr.Error(), http.StatusInternalServerError)
			return
		}

		exists, err := sql.Wrap(ctx, "aerospike", "Exists", func() (bool, error) {
			return client.Exists(nil, key)
		})
		if err != nil {
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"exists": exists,
		})
	})

	// ============================================================
	// CASE 5: BatchGet - Batch Record Retrieval
	// ============================================================
	// Demonstrates: sql.Wrap() for batch operations.
	http.HandleFunc("/batch", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		if client == nil {
			http.Error(w, "Not connected to Aerospike", http.StatusServiceUnavailable)
			return
		}

		// Create multiple keys for batch retrieval
		keys := make([]*aerospike.Key, 3)
		for i := 0; i < 3; i++ {
			key, _ := aerospike.NewKey("test", "demo", fmt.Sprintf("item%d", i+1))
			keys[i] = key
		}

		// BatchGet returns ([]*Record, error)
		records, err := sql.Wrap(ctx, "aerospike", "BatchGet", func() ([]*aerospike.Record, error) {
			return client.BatchGet(nil, keys)
		})
		if err != nil {
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Convert records to response format
		results := make([]map[string]interface{}, len(records))
		for i, rec := range records {
			if rec != nil {
				results[i] = map[string]interface{}{
					"key":  keys[i].Value(),
					"bins": rec.Bins,
				}
			} else {
				results[i] = map[string]interface{}{
					"key":  keys[i].Value(),
					"bins": nil,
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"records": results,
		})
	})

	// ============================================================
	// CASE 6: Health Check
	// ============================================================
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		if client != nil && client.IsConnected() {
			json.NewEncoder(w).Encode(map[string]string{
				"status": "ok",
				"info":   "Connected to Aerospike",
			})
		} else {
			http.Error(w, "Not connected to Aerospike", http.StatusServiceUnavailable)
		}
	})

	fmt.Printf("Aerospike example server starting on port %d\n", port)
	fmt.Printf("Aerospike: %s:%d\n", asHost, asPort)
	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
