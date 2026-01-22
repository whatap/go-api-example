// Package main demonstrates WhaTap APM instrumentation for MongoDB Go Driver.
//
// This example shows how to monitor MongoDB operations using the WhaTap Go API.
// The whatapmongo package provides automatic tracing for:
//   - Connection establishment (StartOpen tracking)
//   - All MongoDB commands (insert, find, update, delete, aggregate, etc.)
//   - Error tracking
//
// Key instrumentation points:
//   1. Replace mongo.Connect() with whatapmongo.Connect()
//   2. Pass context from HTTP transaction to MongoDB operations
//   3. Use trace.StartWithRequest() for HTTP transaction tracking
//
// WhaTap logs generated:
//   - [WA-SQL-04002] Open DB: Connection establishment tracking
//   - [WA-SQL-04003] Sql: MongoDB command execution tracking
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"

	// WhaTap instrumentation packages
	// - whatapmongo: MongoDB driver instrumentation (replaces mongo.Connect)
	// - trace: Core tracing API for transaction management
	"github.com/whatap/go-api/instrumentation/go.mongodb.org/mongo-driver/mongo/whatapmongo"
	"github.com/whatap/go-api/trace"

	// Standard MongoDB driver packages
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Command-line flags for configuration
	udpPortPtr := flag.Int("up", 6600, "WhaTap agent UDP port (default: 6600)")
	portPtr := flag.Int("p", 8080, "HTTP server port (default: 8080)")
	mongoURIPtr := flag.String("mongo", "mongodb://localhost:27017", "MongoDB connection URI")
	setWhatapPtr := flag.Bool("whatap", false, "Enable WhaTap monitoring")

	flag.Parse()
	port := *portPtr
	udpPort := *udpPortPtr
	mongoURI := *mongoURIPtr
	IsWhatap := *setWhatapPtr

	// ============================================================
	// INSTRUMENTATION STEP 1: Initialize WhaTap Agent
	// ============================================================
	// trace.Init() must be called at application startup.
	// Configuration options can be passed via map or loaded from whatap.conf.
	//
	// Common configuration options:
	//   - net_udp_port: UDP port for agent communication
	//   - license: WhaTap license key
	//   - whatap.server.host: WhaTap server address
	//   - app_name: Application name displayed in WhaTap dashboard
	if IsWhatap {
		config := make(map[string]string)
		config["net_udp_port"] = fmt.Sprintf("%d", udpPort)
		trace.Init(config)
	}
	// trace.Shutdown() must be called before application exit
	// to ensure all pending traces are sent to the server.
	defer trace.Shutdown()

	// ============================================================
	// INSTRUMENTATION STEP 2: Use whatapmongo.Connect() instead of mongo.Connect()
	// ============================================================
	// whatapmongo.Connect() wraps the original mongo.Connect() and:
	//   1. Adds a CommandMonitor to track all MongoDB commands
	//   2. Tracks connection establishment via sql.StartOpen()
	//
	// Note: When called outside HTTP handler (like here in main()),
	// txid will be 0 because there's no active HTTP transaction context.
	// This is expected behavior for startup connections.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := whatapmongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Printf("Failed to connect to MongoDB: %v\n", err)
		return
	}
	defer client.Disconnect(context.Background())

	// Verify connection
	if err := client.Ping(ctx, nil); err != nil {
		fmt.Printf("Failed to ping MongoDB: %v\n", err)
		return
	}
	fmt.Println("Connected to MongoDB")

	collection := client.Database("testdb").Collection("items")

	// ============================================================
	// HTTP HANDLER EXAMPLES
	// ============================================================
	// Each handler demonstrates:
	//   1. trace.StartWithRequest(r) - Start HTTP transaction
	//   2. defer trace.End(ctx, nil) - End HTTP transaction
	//   3. Pass ctx to MongoDB operations for trace linkage
	//   4. trace.Error(ctx, err) - Record errors (optional)

	// Root handler - API information
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Start HTTP transaction tracking
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		json.NewEncoder(w).Encode(map[string]string{
			"message":   "MongoDB Example API",
			"endpoints": "/insert, /find, /findall, /update, /delete, /connect, /aggregate, /insertmany, /findandupdate, /findbyid",
		})
	})

	// ============================================================
	// CASE 1: InsertOne - Document Insertion
	// ============================================================
	// Demonstrates: Single document insertion with transaction tracing.
	// The MongoDB insert command is automatically tracked via CommandMonitor.
	http.HandleFunc("/insert", func(w http.ResponseWriter, r *http.Request) {
		// IMPORTANT: Start transaction and get traced context
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		name := r.URL.Query().Get("name")
		if name == "" {
			name = "default"
		}
		value := r.URL.Query().Get("value")

		doc := bson.M{
			"name":       name,
			"value":      value,
			"created_at": time.Now(),
		}

		// Pass the traced context to MongoDB operation
		// This links the MongoDB command to the HTTP transaction
		result, err := collection.InsertOne(ctx, doc)
		if err != nil {
			// Record error in transaction trace
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "ok",
			"inserted_id": result.InsertedID,
		})
	})

	// ============================================================
	// CASE 2: FindOne - Single Document Query
	// ============================================================
	// Demonstrates: Single document retrieval with error handling.
	http.HandleFunc("/find", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}

		var result bson.M
		// FindOne is tracked automatically via CommandMonitor
		err := collection.FindOne(ctx, bson.M{"name": name}).Decode(&result)
		if err != nil {
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// ============================================================
	// CASE 3: Find - Multiple Documents Query with Cursor
	// ============================================================
	// Demonstrates: Cursor-based document retrieval.
	// The Find command is tracked; cursor operations are lightweight.
	http.HandleFunc("/findall", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// Find returns a cursor for iterating documents
		cursor, err := collection.Find(ctx, bson.M{})
		if err != nil {
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(ctx)

		var results []bson.M
		// cursor.All() decodes all documents into the slice
		if err = cursor.All(ctx, &results); err != nil {
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	// ============================================================
	// CASE 4: UpdateOne - Document Update
	// ============================================================
	// Demonstrates: Document update with filter and update operators.
	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		name := r.URL.Query().Get("name")
		value := r.URL.Query().Get("value")
		if name == "" || value == "" {
			http.Error(w, "name and value required", http.StatusBadRequest)
			return
		}

		filter := bson.M{"name": name}
		update := bson.M{"$set": bson.M{"value": value, "updated_at": time.Now()}}

		// UpdateOne is tracked via CommandMonitor
		result, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":         "ok",
			"matched_count":  result.MatchedCount,
			"modified_count": result.ModifiedCount,
		})
	})

	// ============================================================
	// CASE 5: DeleteOne - Document Deletion
	// ============================================================
	// Demonstrates: Document deletion with filter.
	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}

		result, err := collection.DeleteOne(ctx, bson.M{"name": name})
		if err != nil {
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":        "ok",
			"deleted_count": result.DeletedCount,
		})
	})

	// ============================================================
	// CASE 6: Connect Inside Handler - Transaction-Linked Connection
	// ============================================================
	// Demonstrates: Creating a new MongoDB connection inside an HTTP handler.
	//
	// IMPORTANT: When whatapmongo.Connect() is called with a context derived
	// from trace.StartWithRequest(), the connection establishment is linked
	// to the HTTP transaction. This allows tracking of:
	//   - txid: Transaction ID (non-zero when inside HTTP handler)
	//   - uri: Request URI ("/connect" in this case)
	//
	// WhaTap log example:
	//   [WA-SQL-04002] Open DB txid: -4642369574789341408, dbhost: mongodb://localhost:27017
	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// Create connection context with timeout, derived from HTTP transaction context
		connCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		// Connection is now linked to HTTP transaction
		// Check WA-SQL-04002 log for txid and uri
		tempClient, err := whatapmongo.Connect(connCtx, options.Client().ApplyURI(mongoURI))
		if err != nil {
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tempClient.Disconnect(connCtx)

		// Verify connection
		if err := tempClient.Ping(connCtx, nil); err != nil {
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "Connected and disconnected successfully (check WA-SQL-04002 log for txid)",
		})
	})

	// ============================================================
	// CASE 7: Aggregate - Aggregation Pipeline
	// ============================================================
	// Demonstrates: MongoDB aggregation pipeline with tracing.
	// Complex aggregation queries are tracked as a single command.
	http.HandleFunc("/aggregate", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		// Aggregation pipeline: group by name and count
		pipeline := []bson.M{
			{"$group": bson.M{
				"_id":   "$name",
				"count": bson.M{"$sum": 1},
			}},
			{"$sort": bson.M{"count": -1}},
		}

		cursor, err := collection.Aggregate(ctx, pipeline)
		if err != nil {
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(ctx)

		var results []bson.M
		if err = cursor.All(ctx, &results); err != nil {
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	// ============================================================
	// CASE 8: InsertMany - Bulk Document Insertion
	// ============================================================
	// Demonstrates: Bulk insert with multiple documents.
	http.HandleFunc("/insertmany", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		docs := []interface{}{
			bson.M{"name": "item1", "value": "val1", "created_at": time.Now()},
			bson.M{"name": "item2", "value": "val2", "created_at": time.Now()},
			bson.M{"name": "item3", "value": "val3", "created_at": time.Now()},
		}

		result, err := collection.InsertMany(ctx, docs)
		if err != nil {
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":       "ok",
			"inserted_ids": result.InsertedIDs,
		})
	})

	// ============================================================
	// CASE 9: FindOneAndUpdate - Atomic Find and Update
	// ============================================================
	// Demonstrates: Atomic operation that finds and updates a document.
	http.HandleFunc("/findandupdate", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		name := r.URL.Query().Get("name")
		value := r.URL.Query().Get("value")
		if name == "" || value == "" {
			http.Error(w, "name and value required", http.StatusBadRequest)
			return
		}

		filter := bson.M{"name": name}
		update := bson.M{"$set": bson.M{"value": value, "updated_at": time.Now()}}
		// Return the updated document (After) instead of the original (Before)
		opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

		var result bson.M
		err := collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
		if err != nil {
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// ============================================================
	// CASE 10: FindById - Query by ObjectID
	// ============================================================
	// Demonstrates: Document retrieval using MongoDB ObjectID.
	http.HandleFunc("/findbyid", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}

		// Convert string to ObjectID
		objID, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			trace.Error(ctx, err)
			http.Error(w, "invalid ObjectID format", http.StatusBadRequest)
			return
		}

		var result bson.M
		err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&result)
		if err != nil {
			trace.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	fmt.Printf("MongoDB example server starting on port %d\n", port)
	fmt.Printf("MongoDB URI: %s\n", mongoURI)
	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
