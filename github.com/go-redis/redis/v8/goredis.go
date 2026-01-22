// go-redis v8 Instrumentation Example
//
// This example demonstrates how to instrument go-redis v8 applications using WhaTap Go API.
//
// INSTRUMENTATION OVERVIEW:
// 1. Replace redis.NewClient() with whatapgoredis.NewClient()
// 2. Or use whatapgoredis.WrapClient() for existing clients
// 3. Pass context from HTTP handlers to Redis commands for distributed tracing
//
// IMPORTANT DIFFERENCES FROM v9:
// - v8 does NOT support DialHook (no connection tracking via StartOpen)
// - v8 uses BeforeProcess/AfterProcess Hook interface
// - v8 import path: github.com/go-redis/redis/v8
// - v9 import path: github.com/redis/go-redis/v9
//
// RECOMMENDATION: Upgrade to v9 for full connection tracking support.

package main

import (
	"flag"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/whatap/go-api/instrumentation/github.com/go-redis/redis/v8/whatapgoredis"
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
	if IsWhatap {
		config := make(map[string]string)
		config["net_udp_port"] = fmt.Sprintf("%d", udpPort)
		trace.Init(config)
	}
	defer trace.Shutdown()

	templatePath := "templates/github.com/go-redis/index.html"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tp, err := template.ParseFiles(templatePath)
		if err != nil {
			fmt.Println("Template not loaded, ", err)
			return
		}

		data := &HTMLData{}
		data.Title = "go-redis v8 Test Page"
		data.Content = r.RequestURI

		tp.Execute(w, data)
	})

	// ============================================================
	// INSTRUMENTATION STEP 2: Create Instrumented Redis Client
	// ============================================================
	// Use whatapgoredis.NewClient() instead of redis.NewClient()
	//
	// BEFORE (Original):
	//   rdb := redis.NewClient(&redis.Options{...})
	//
	// AFTER (Instrumented):
	//   rdb := whatapgoredis.NewClient(&redis.Options{...})
	//
	// NOTE: v8 uses BeforeProcess/AfterProcess hooks (different from v9)
	// - BeforeProcess: Called before each command execution
	// - AfterProcess: Called after command completion (with error if any)
	// - NO DialHook: Connection establishment is NOT tracked in v8
	rdb := whatapgoredis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer rdb.Close()

	// ============================================================
	// Case 1: Basic SET and GET Operations
	// ============================================================
	// Pass context to enable distributed tracing.
	// Each command is tracked with: command name, key, duration, result
	http.HandleFunc("/SetAndGet", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// SET command - tracked via BeforeProcess/AfterProcess hooks
		err := rdb.Set(ctx, "mykey", "myvalue", time.Minute*10).Err()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// GET command
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
	http.HandleFunc("/Pipeline", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		pipe := rdb.Pipeline()

		pipe.Set(ctx, "key1", "value1", time.Minute*10)
		pipe.Set(ctx, "key2", "value2", time.Minute*10)
		pipe.Get(ctx, "key1")
		pipe.Get(ctx, "key2")

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
	http.HandleFunc("/Transaction", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		txPipe := rdb.TxPipeline()

		txPipe.Set(ctx, "tx_key1", "tx_value1", time.Minute*10)
		txPipe.Set(ctx, "tx_key2", "tx_value2", time.Minute*10)
		txPipe.Incr(ctx, "tx_counter")

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
	// Case 4: Hash Operations
	// ============================================================
	http.HandleFunc("/Hash", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// HSET
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

		// HGETALL
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
	// Case 5: List Operations
	// ============================================================
	http.HandleFunc("/List", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// LPUSH
		err := rdb.LPush(ctx, "mylist", "item1", "item2", "item3").Err()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// LRANGE
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
	// Case 6: Set Operations
	// ============================================================
	http.HandleFunc("/Set", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// SADD
		err := rdb.SAdd(ctx, "myset", "member1", "member2", "member3").Err()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// SMEMBERS
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
	// Case 7: Sorted Set Operations
	// ============================================================
	// NOTE: v8 uses &redis.Z{} (pointer), v9 uses redis.Z{} (value)
	http.HandleFunc("/SortedSet", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// ZADD - v8 uses pointer for redis.Z
		err := rdb.ZAdd(ctx, "leaderboard", &redis.Z{Score: 100, Member: "player1"}, &redis.Z{Score: 200, Member: "player2"}, &redis.Z{Score: 150, Member: "player3"}).Err()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// ZRANGE with scores
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

		// TTL
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
	// Use whatapgoredis.WrapClient() to add instrumentation to existing clients
	http.HandleFunc("/WrapClient", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// Create client using original method
		existingClient := redis.NewClient(&redis.Options{
			Addr: redisAddr,
		})
		defer existingClient.Close()

		// Add WhaTap instrumentation
		whatapgoredis.WrapClient(existingClient)

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
