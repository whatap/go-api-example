// Redigo (github.com/gomodule/redigo) Instrumentation Example
//
// This example demonstrates how to instrument Redigo (Redis client) applications
// using WhaTap Go API.
//
// INSTRUMENTATION OVERVIEW:
// Replace redis.Dial* functions with whatapredigo.Dial* equivalents:
//   - redis.Dial() -> whatapredigo.Dial()
//   - redis.DialContext() -> whatapredigo.DialContext()
//   - redis.DialURL() -> whatapredigo.DialURL()
//   - redis.DialURLContext() -> whatapredigo.DialURLContext()
//
// FEATURES:
// - Connection tracking via StartOpen
// - Command tracking (SET, GET, etc.) via Do() method
// - Pipeline tracking via Send/Flush/Receive
// - Error tracking with trace.Error()
// - Distributed tracing via context propagation
//
// CONTEXT PROPAGATION:
// There are TWO ways to pass context for distributed tracing:
// 1. Use DialContext/DialURLContext - context passed at connection time
// 2. Use Dial + conn.WithContext(ctx) - context added after connection
//
// For connection pool (redis.Pool), use DialContext in the pool configuration.

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/whatap/go-api/instrumentation/github.com/gomodule/redigo/whatapredigo"
	"github.com/whatap/go-api/trace"
)

type HTMLData struct {
	Title   string
	Content string
	//HTMLContent template.HTML
}

func main() {
	udpPortPtr := flag.Int("up", 6600, "agent port(udp). defalt 6600 ")
	portPtr := flag.Int("p", 8080, "web port. default 8080  ")
	dataSourcePtr := flag.String("ds", "phpdemo3:6379", " dataSourceName ")
	setWhatapPtr := flag.Bool("whatap", false, "set whatap")

	flag.Parse()
	port := *portPtr
	udpPort := *udpPortPtr
	dataSource := *dataSourcePtr
	IsWhatap := *setWhatapPtr

	// ============================================================
	// INSTRUMENTATION STEP 1: Initialize WhaTap Agent
	// ============================================================
	// Call trace.Init() at application startup.
	// Configuration can be passed via map or loaded from whatap.conf file.
	if IsWhatap {
		config := make(map[string]string)
		config["net_udp_port"] = fmt.Sprintf("%d", udpPort)
		trace.Init(config)
	}
	defer trace.Shutdown()

	templatePath := "templates/github.com/gomodule/index.html"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tp, err := template.ParseFiles(templatePath)
		if err != nil {
			fmt.Println("Template not loaded, ", err)
			return
		}

		data := &HTMLData{}
		data.Title = "Redigo Test Page"
		data.Content = r.RequestURI

		tp.Execute(w, data)
	})

	// ============================================================
	// Case 1: Connection Pool with DialContext
	// ============================================================
	// Use whatapredigo.DialContext in pool's DialContext function.
	// This enables connection tracking for each pool.Get() call.
	//
	// BEFORE (Original):
	//   servicePool := &redis.Pool{
	//       DialContext: func(ctx context.Context) (redis.Conn, error) {
	//           return redis.DialContext(ctx, "tcp", address)
	//       },
	//   }
	//
	// AFTER (Instrumented):
	//   servicePool := &redis.Pool{
	//       DialContext: func(ctx context.Context) (redis.Conn, error) {
	//           return whatapredigo.DialContext(ctx, "tcp", address)
	//       },
	//   }
	servicePool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		DialContext: func(ctx context.Context) (redis.Conn, error) {
			return whatapredigo.DialContext(ctx, "tcp", dataSource)
		},
	}
	defer servicePool.Close()

	// Pool usage example
	// - Use pool.GetContext(ctx) to get connection with context
	// - All Do() commands on this connection are tracked
	http.HandleFunc("/SetAndGetWithPool", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// Get connection from pool with context
		conn, err := servicePool.GetContext(ctx)
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}
		defer conn.Close()

		// SET command - tracked
		_, err = conn.Do("SET", "DataKey", "DataValue")
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		// GET command - tracked
		data, err := redis.Bytes(conn.Do("GET", "DataKey"))
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		fmt.Println(string(data))
	})

	// ============================================================
	// Case 2: Dial + WithContext (Manual Context Assignment)
	// ============================================================
	// Use whatapredigo.Dial() then add context with conn.WithContext(ctx).
	// Useful when connection is created before request context is available.
	//
	// BEFORE (Original):
	//   conn, err := redis.Dial("tcp", address)
	//   conn = conn.WithContext(ctx)
	//
	// AFTER (Instrumented):
	//   conn, err := whatapredigo.Dial("tcp", address)
	//   conn = conn.WithContext(ctx)  // Add context for tracking
	http.HandleFunc("/SetAndGetWithDial", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// Dial without context first
		conn, err := whatapredigo.Dial("tcp", dataSource)
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}
		defer conn.Close()

		// Add context for command tracking
		conn = conn.WithContext(ctx)

		// Commands are now tracked and linked to HTTP transaction
		_, err = conn.Do("SET", "DataKey", 1)
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		data, err := redis.Bytes(conn.Do("GET", "DataKey"))
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		fmt.Println(string(data))

	})

	// ============================================================
	// Case 3: DialContext (Recommended)
	// ============================================================
	// Use whatapredigo.DialContext() - context is passed at connection time.
	// This is the recommended approach for request handlers.
	//
	// BEFORE (Original):
	//   conn, err := redis.DialContext(ctx, "tcp", address)
	//
	// AFTER (Instrumented):
	//   conn, err := whatapredigo.DialContext(ctx, "tcp", address)
	http.HandleFunc("/SetAndGetWithDialContext", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// Connection and all commands linked to HTTP transaction
		conn, err := whatapredigo.DialContext(ctx, "tcp", dataSource)
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}
		defer conn.Close()

		_, err = conn.Do("SET", "DataKey", "DataValue")
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		data, err := redis.Bytes(conn.Do("GET", "DataKey"))
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		fmt.Println(string(data))

	})

	// ============================================================
	// Case 4: DialURL + WithContext
	// ============================================================
	// Use whatapredigo.DialURL() with Redis URL format.
	// Context is added separately with WithContext().
	//
	// BEFORE (Original):
	//   conn, err := redis.DialURL("redis://host:port")
	//
	// AFTER (Instrumented):
	//   conn, err := whatapredigo.DialURL("redis://host:port")
	//   conn = conn.WithContext(ctx)
	http.HandleFunc("/SetAndGetWithDialURL", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		dUrl := fmt.Sprintf("redis://%s", dataSource)
		conn, err := whatapredigo.DialURL(dUrl)
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}
		defer conn.Close()

		// Add context for tracking
		conn = conn.WithContext(ctx)

		_, err = conn.Do("SET", "DataKey", "DataValue")
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		data, err := redis.Bytes(conn.Do("GET", "DataKey"))
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		fmt.Println(string(data))

	})

	// ============================================================
	// Case 5: DialURLContext (URL + Context Combined)
	// ============================================================
	// Use whatapredigo.DialURLContext() for URL-based connection with context.
	//
	// BEFORE (Original):
	//   conn, err := redis.DialURLContext(ctx, "redis://host:port")
	//
	// AFTER (Instrumented):
	//   conn, err := whatapredigo.DialURLContext(ctx, "redis://host:port")
	http.HandleFunc("/SetAndGetWithDialURLContext", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		dUrl := fmt.Sprintf("redis://%s", dataSource)
		conn, err := whatapredigo.DialURLContext(ctx, dUrl)
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}
		defer conn.Close()

		_, err = conn.Do("SET", "DataKey", "DataValue")
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		data, err := redis.Bytes(conn.Do("GET", "DataKey"))
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		fmt.Println(string(data))

	})

	// ============================================================
	// Case 6: Dial with Timeout Options
	// ============================================================
	// Timeout options work the same way with whatapredigo.
	// Pass redis.DialOption as variadic arguments.
	http.HandleFunc("/SetAndGetWithDialTimeout", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// Dial with timeout options - same as original redigo
		conn, err := whatapredigo.Dial("tcp", dataSource, redis.DialConnectTimeout(time.Millisecond*1000), redis.DialReadTimeout(time.Millisecond*1000), redis.DialWriteTimeout(time.Millisecond*1000))
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}
		defer conn.Close()

		conn = conn.WithContext(ctx)

		_, err = conn.Do("SET", "DataKey", "DataValue")
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		data, err := redis.Bytes(conn.Do("GET", "DataKey"))
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		fmt.Println(string(data))

	})

	// ============================================================
	// Case 7: Pipeline with Send/Flush/Receive
	// ============================================================
	// Pipeline commands are tracked via Send/Flush/Receive pattern.
	// - Send(): Queues command (tracked)
	// - Flush(): Sends all queued commands (tracked)
	// - Receive(): Receives response
	http.HandleFunc("/SetAndGetWithDialSendReceive", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		conn, err := whatapredigo.Dial("tcp", dataSource)
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}
		defer conn.Close()

		conn = conn.WithContext(ctx)

		// Queue SET command
		err = conn.Send("SET", "DataKey", "DataValue")
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return

		}
		// Queue GET command
		err = conn.Send("GET", "DataKey")
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		// Flush all queued commands to server
		err = conn.Flush()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		// Receive responses in order
		conn.Receive()              // SET response
		data, err := conn.Receive() // GET response
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}
		fmt.Println(data)
	})

	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	fmt.Println(port)
}
