// go-redis v9 Instrumentation Example
//
// This example demonstrates how to instrument go-redis v9 applications using WhaTap Go API.
//
// INSTRUMENTATION OVERVIEW:
// 1. Replace redis.NewClient() with whatapgoredis.NewClient()
// 2. Or use whatapgoredis.WrapClient() for existing clients
// 3. Pass context from HTTP handlers to Redis commands for distributed tracing
//
// FEATURES:
// - Automatic command tracking (SET, GET, HSET, etc.)
// - Connection tracking via DialHook (StartOpen)
// - Distributed tracing support via context propagation
// - Error tracking for failed commands
//
// COMPARISON WITH v8:
// - v9 supports DialHook for connection tracking (v8 does not)
// - v9 has better Hook interface (ProcessHook, ProcessPipelineHook)

package main

import (
	"flag"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/whatap/go-api/instrumentation/github.com/redis/go-redis/v9/whatapgoredis"
	"github.com/whatap/go-api/trace"
)

type HTMLData struct {
	Title   string
	Content string
}

func main() {
	udpPortPtr := flag.Int("up", 6600, "agent port(udp). default 6600")
	portPtr := flag.Int("p", 8080, "web port. default 8080")
	redisAddrPtr := flag.String("redis", "localhost:6379", "redis address")
	setWhatapPtr := flag.Bool("whatap", false, "set whatap")

	flag.Parse()
	port := *portPtr
	udpPort := *udpPortPtr
	redisAddr := *redisAddrPtr
	IsWhatap := *setWhatapPtr

	// ============================================================
	// INSTRUMENTATION STEP 1: Initialize WhaTap Agent
	// ============================================================
	// Call trace.Init() at application startup to initialize the agent.
	// Configuration can be passed via map or loaded from whatap.conf file.
	if IsWhatap {
		config := make(map[string]string)
		config["net_udp_port"] = fmt.Sprintf("%d", udpPort)
		trace.Init(config)
	}
	defer trace.Shutdown()

	templatePath := "templates/github.com/redis/index.html"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tp, err := template.ParseFiles(templatePath)
		if err != nil {
			fmt.Println("Template not loaded, ", err)
			return
		}

		data := &HTMLData{}
		data.Title = "go-redis v9 Test Page"
		data.Content = r.RequestURI

		tp.Execute(w, data)
	})

	// ============================================================
	// INSTRUMENTATION STEP 2: Create Instrumented Redis Client
	// ============================================================
	// METHOD 1: Use whatapgoredis.NewClient() instead of redis.NewClient()
	//
	// BEFORE (Original):
	//   rdb := redis.NewClient(&redis.Options{...})
	//
	// AFTER (Instrumented):
	//   rdb := whatapgoredis.NewClient(&redis.Options{...})
	//
	// This wraps the client with WhaTap hooks for:
	// - DialHook: Tracks connection establishment (StartOpen)
	// - ProcessHook: Tracks individual commands
	// - ProcessPipelineHook: Tracks pipeline commands
	rdb := whatapgoredis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer rdb.Close()

	// ============================================================
	// Case 1: Basic SET and GET Operations
	// ============================================================
	// IMPORTANT: Always pass the context from HTTP request to Redis commands.
	// This links Redis operations to the HTTP transaction for distributed tracing.
	http.HandleFunc("/SetAndGet", func(w http.ResponseWriter, r *http.Request) {
		// Start HTTP transaction and get context
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// SET command - context is passed to link with HTTP transaction
		// WhaTap records: command=SET, key=mykey, duration, success/failure
		err := rdb.Set(ctx, "mykey", "myvalue", time.Minute*10).Err()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err) // Record error in transaction
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// GET command - same context for transaction linking
		val, err := rdb.Get(ctx, "mykey").Result()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "GET mykey = %s\n", val)
	})

	// ============================================================
	// Case 2: Pipeline Operations
	// ============================================================
	// Pipeline commands are batched and sent together.
	// WhaTap tracks the entire pipeline execution as one operation.
	http.HandleFunc("/Pipeline", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// Create pipeline - commands are queued but not executed yet
		pipe := rdb.Pipeline()

		// Queue multiple commands
		pipe.Set(ctx, "key1", "value1", time.Minute*10)
		pipe.Set(ctx, "key2", "value2", time.Minute*10)
		pipe.Get(ctx, "key1")
		pipe.Get(ctx, "key2")

		// Execute all commands in one round trip
		// WhaTap tracks this as a pipeline operation with all commands
		cmds, err := pipe.Exec(ctx)
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, cmd := range cmds {
			fmt.Fprintf(w, "%s\n", cmd.String())
		}
	})

	// ============================================================
	// Case 3: Transaction Pipeline (MULTI/EXEC)
	// ============================================================
	// TxPipeline wraps commands in MULTI/EXEC for atomicity.
	// WhaTap tracks the entire transaction as one operation.
	http.HandleFunc("/Transaction", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// Create transaction pipeline
		txPipe := rdb.TxPipeline()

		// Queue commands (will be executed atomically)
		txPipe.Set(ctx, "tx_key1", "tx_value1", time.Minute*10)
		txPipe.Set(ctx, "tx_key2", "tx_value2", time.Minute*10)
		txPipe.Incr(ctx, "tx_counter")

		// Execute transaction
		cmds, err := txPipe.Exec(ctx)
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, cmd := range cmds {
			fmt.Fprintf(w, "%s\n", cmd.String())
		}
	})

	// ============================================================
	// Case 4: Hash Operations (HSET, HGETALL)
	// ============================================================
	http.HandleFunc("/Hash", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// HSET - set multiple hash fields
		err := rdb.HSet(ctx, "user:1000", map[string]interface{}{
			"name":  "John",
			"email": "john@example.com",
			"age":   30,
		}).Err()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// HGETALL - get all hash fields
		result, err := rdb.HGetAll(ctx, "user:1000").Result()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for k, v := range result {
			fmt.Fprintf(w, "%s: %s\n", k, v)
		}
	})

	// ============================================================
	// Case 5: List Operations (LPUSH, LRANGE)
	// ============================================================
	http.HandleFunc("/List", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// LPUSH - add items to list head
		err := rdb.LPush(ctx, "mylist", "item1", "item2", "item3").Err()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// LRANGE - get list items
		items, err := rdb.LRange(ctx, "mylist", 0, -1).Result()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for i, item := range items {
			fmt.Fprintf(w, "[%d] %s\n", i, item)
		}
	})

	// ============================================================
	// Case 6: Set Operations (SADD, SMEMBERS)
	// ============================================================
	http.HandleFunc("/Set", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// SADD - add members to set
		err := rdb.SAdd(ctx, "myset", "member1", "member2", "member3").Err()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// SMEMBERS - get all set members
		members, err := rdb.SMembers(ctx, "myset").Result()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, member := range members {
			fmt.Fprintf(w, "%s\n", member)
		}
	})

	// ============================================================
	// Case 7: Sorted Set Operations (ZADD, ZRANGE)
	// ============================================================
	http.HandleFunc("/SortedSet", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// ZADD - add members with scores
		err := rdb.ZAdd(ctx, "leaderboard", redis.Z{Score: 100, Member: "player1"}, redis.Z{Score: 200, Member: "player2"}, redis.Z{Score: 150, Member: "player3"}).Err()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// ZRANGE with scores - get sorted members
		results, err := rdb.ZRangeWithScores(ctx, "leaderboard", 0, -1).Result()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, z := range results {
			fmt.Fprintf(w, "%v: %.0f\n", z.Member, z.Score)
		}
	})

	// ============================================================
	// Case 8: Pub/Sub - Publish
	// ============================================================
	http.HandleFunc("/Publish", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// PUBLISH - send message to channel
		err := rdb.Publish(ctx, "mychannel", "Hello, Redis!").Err()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Published message to mychannel\n")
	})

	// ============================================================
	// Case 9: Expire and TTL
	// ============================================================
	http.HandleFunc("/ExpireAndTTL", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// SET with expiration
		err := rdb.Set(ctx, "temp_key", "temp_value", time.Second*30).Err()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TTL - get remaining time to live
		ttl, err := rdb.TTL(ctx, "temp_key").Result()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "TTL for temp_key: %v\n", ttl)
	})

	// ============================================================
	// INSTRUMENTATION STEP 3: Wrap Existing Client
	// ============================================================
	// METHOD 2: Use whatapgoredis.WrapClient() for existing clients
	//
	// This is useful when:
	// - You already have a redis.Client created elsewhere
	// - You want to add instrumentation without changing the creation code
	//
	// BEFORE (Original):
	//   client := redis.NewClient(&redis.Options{...})
	//   // use client...
	//
	// AFTER (Instrumented):
	//   client := redis.NewClient(&redis.Options{...})
	//   whatapgoredis.WrapClient(client)  // Add WhaTap hooks
	//   // use client...
	http.HandleFunc("/WrapClient", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// Create client using original redis.NewClient
		existingClient := redis.NewClient(&redis.Options{
			Addr: redisAddr,
		})
		defer existingClient.Close()

		// Add WhaTap instrumentation to existing client
		whatapgoredis.WrapClient(existingClient)

		// Now all commands on existingClient are traced
		err := existingClient.Set(ctx, "wrapped_key", "wrapped_value", time.Minute*10).Err()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		val, err := existingClient.Get(ctx, "wrapped_key").Result()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "GET wrapped_key = %s\n", val)
	})

	fmt.Printf("Server starting on port %d...\n", port)
	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
