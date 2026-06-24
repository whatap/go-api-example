// GORM v2 (gorm.io/gorm) Instrumentation Example
//
// This example demonstrates how to instrument GORM v2 applications using WhaTap Go API.
//
// INSTRUMENTATION OVERVIEW:
// There are TWO methods to instrument GORM:
//
// METHOD 1: Use existing whatapsql driver (for MySQL, PostgreSQL, MSSQL)
//   - Use whatapsql.OpenContext() to create a tracked DB connection
//   - Pass the connection to GORM via mysql.Config{Conn: dbConn}
//   - SQL queries are tracked through the whatapsql driver
//
// METHOD 2: Use whatapgorm hooks (recommended for all databases)
//   - Replace gorm.Open() with whatapgorm.Open()
//   - Or use whatapgorm.OpenWithContext() for context-aware tracking
//   - Or use whatapgorm.WithContext() to add context to existing DB
//   - GORM callbacks are hooked for query tracking
//
// FEATURES:
// - Automatic SQL query tracking (CREATE, UPDATE, DELETE, SELECT)
// - Connection tracking via StartOpen
// - Transaction tracking (Begin, Commit, Rollback)
// - Error tracking with trace.Error()
// - Distributed tracing support via context propagation
//
// CONTEXT PROPAGATION:
// - Use OpenWithContext() when opening connection in request handler
// - Use WithContext() to add HTTP context to existing global DB connection
// - Context links DB operations to HTTP transaction for distributed tracing

package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"text/template"

	"github.com/whatap/go-api/instrumentation/database/sql/whatapsql"
	"github.com/whatap/go-api/instrumentation/github.com/go-gorm/gorm/whatapgorm"
	"github.com/whatap/go-api/trace"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Code  int
	Price int
}

type HTMLData struct {
	Title   string
	Content string
	//HTMLContent template.HTML
}

func main() {
	udpPortPtr := flag.Int("up", 6600, "agent port(udp). defalt 6600 ")
	portPtr := flag.Int("p", 8080, "web port. default 8080  ")
	dataSourcePtr := flag.String("ds", "doremimaker:doremimaker@tcp(phpdemo2:3306)/doremimaker", " dataSourceName ")
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

	templatePath := "templates/github.com/go-gorm/index.html"

	http.HandleFunc("/", trace.Func(func(w http.ResponseWriter, r *http.Request) {
		tp, err := template.ParseFiles(templatePath)
		if err != nil {
			fmt.Println("Template not loaded, ", err)
			return
		}

		data := &HTMLData{}
		data.Title = "GormV2 Test Page"
		data.Content = r.RequestURI

		tp.Execute(w, data)
	}))

	// ============================================================
	// METHOD 1: Use whatapsql Driver with GORM
	// ============================================================
	// For MySQL, PostgreSQL, MSSQL - use existing whatapsql driver.
	// This method passes the tracked sql.DB connection to GORM.
	//
	// BEFORE (Original):
	//   db, err := gorm.Open(mysql.Open(dataSource), &gorm.Config{})
	//
	// AFTER (Instrumented with whatapsql):
	//   dbConn, err := whatapsql.OpenContext(ctx, "mysql", dataSource)
	//   db, err := gorm.Open(mysql.New(mysql.Config{Conn: dbConn}), &gorm.Config{})
	http.HandleFunc("/WhatapDriverTest", trace.Func(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Open connection with whatapsql for tracking
		dbConn, err := whatapsql.OpenContext(ctx, "mysql", dataSource)
		db, err := gorm.Open(mysql.New(mysql.Config{Conn: dbConn}), &gorm.Config{})
		if err != nil {
			panic("failed to connect to Db")
		}

		db.AutoMigrate(&Product{})

		for i := 0; i < 100; i++ {
			db.Create(&Product{Code: i, Price: i * 100})
		}

	}))

	// ============================================================
	// METHOD 2: Use whatapgorm Hooks (Global DB)
	// ============================================================
	// Use whatapgorm.Open() instead of gorm.Open() for automatic tracking.
	// This hooks into GORM's callback system for query tracking.
	//
	// BEFORE (Original):
	//   db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	//
	// AFTER (Instrumented):
	//   db, err := whatapgorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	//
	// NOTE: Global DB without context will track queries but won't link to HTTP transaction.
	serviceDB, err := whatapgorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to Db")
	}

	// ============================================================
	// Case 1: AutoMigrate without Context
	// ============================================================
	// AutoMigrate works without context but won't be linked to any transaction.
	serviceDB.AutoMigrate(&Product{})

	// ============================================================
	// Case 2: OpenWithContext - Insert and Delete
	// ============================================================
	// Use whatapgorm.OpenWithContext() when opening connection inside request handler.
	// This links all DB operations to the current HTTP transaction.
	//
	// BEFORE (Original):
	//   db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	//
	// AFTER (Instrumented with context):
	//   db, err := whatapgorm.OpenWithContext(sqlite.Open("test.db"), &gorm.Config{}, ctx)
	http.HandleFunc("/InsertAndDelete", trace.Func(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Open with context for transaction linking
		db, err := whatapgorm.OpenWithContext(sqlite.Open("test.db"), &gorm.Config{}, ctx)
		if err != nil {
			panic("failed to connect to Db")
		}

		// Create operations - tracked and linked to HTTP transaction
		for i := 0; i < 100; i++ {
			db.Create(&Product{Code: i, Price: i * 100})
		}

		// Delete operation - tracked
		db.Unscoped().Delete(&Product{}, "Code >= ? AND Code < ?", 0, 100)

	}))

	// ============================================================
	// Case 3: Transaction with Begin/Commit
	// ============================================================
	// GORM transactions (Begin, Commit, Rollback) are automatically tracked.
	http.HandleFunc("/InsertAndUpdate", trace.Func(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		db, err := whatapgorm.OpenWithContext(sqlite.Open("test.db"), &gorm.Config{}, ctx)
		if err != nil {
			panic("failed to connect to Db")
		}

		for i := 0; i < 100; i++ {
			db.Create(&Product{Code: i, Price: i * 100})
		}

		// Transaction operations - all tracked
		for i := 0; i < 100; i++ {
			var product Product
			tx := db.Begin() // Begin is tracked
			tx.First(&product, "Code = ?", i)
			tx.Model(&product).Update("price", product.Price*100)
			tx.Commit() // Commit is tracked
		}
	}))

	// ============================================================
	// Case 4: Global DB without Context (Not Recommended)
	// ============================================================
	// Using global DB without context will track queries
	// but won't link them to HTTP transaction for distributed tracing.
	// Use WithContext() instead for better tracing (see Case 5).
	http.HandleFunc("/Select", trace.Func(func(w http.ResponseWriter, r *http.Request) {
		var products []Product
		var buffer bytes.Buffer

		// Query without context - tracked but not linked to HTTP transaction
		serviceDB.Find(&products, "1 = 1")

		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		for i, product := range products {
			buffer.WriteString(fmt.Sprintf("Index : %d, Code : %d, Price : %d </br>", i, product.Code, product.Price))
		}

		buffer.WriteString("</body></html>")

		_, _ = w.Write(buffer.Bytes())
	}))

	// ============================================================
	// Case 5: WithContext - Link Global DB to HTTP Transaction
	// ============================================================
	// Use whatapgorm.WithContext() to add HTTP context to existing global DB.
	// This enables distributed tracing by linking DB operations to HTTP transaction.
	//
	// BEFORE (Original):
	//   db := serviceDB.WithContext(ctx)  // Standard GORM
	//
	// AFTER (Instrumented):
	//   db := whatapgorm.WithContext(ctx, serviceDB)  // WhaTap tracked
	http.HandleFunc("/SelectWithContext", trace.Func(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Add context to global DB for transaction linking
		db := whatapgorm.WithContext(ctx, serviceDB)

		var products []Product
		var buffer bytes.Buffer

		// Query with context - tracked AND linked to HTTP transaction
		db.Find(&products, "1 = 1")

		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		for i, product := range products {
			buffer.WriteString(fmt.Sprintf("Index : %d, Code : %d, Price : %d </br>", i, product.Code, product.Price))
		}

		buffer.WriteString("</body></html>")

		_, _ = w.Write(buffer.Bytes())
	}))

	// ============================================================
	// Case 6: Delete All
	// ============================================================
	http.HandleFunc("/DeleteAll", trace.Func(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		db, err := whatapgorm.OpenWithContext(sqlite.Open("test.db"), &gorm.Config{}, ctx)
		if err != nil {
			panic("failed to connect to Db")
		}

		db.Unscoped().Delete(&Product{}, "1 = 1")
	}))

	// ============================================================
	// Case 7: Transaction with Rollback/Commit and Custom Steps
	// ============================================================
	// Demonstrates:
	// - Transaction Begin/Rollback/Commit tracking
	// - Custom trace.Step() for business logic annotation
	// - Recover handling with Rollback
	http.HandleFunc("/DBTxTest", trace.Func(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		db, err := whatapgorm.OpenWithContext(sqlite.Open("test.db"), &gorm.Config{}, ctx)
		if err != nil {
			panic("failed to connect to Db")
		}

		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				// Add custom step for rollback logging
				trace.Step(ctx, "GormV2 Message", "Tranaction Rollback", 0, 0)
				tx.Rollback()
			}
		}()

		var beforeCount int64
		var afterCount int64

		for i := 0; i < 100; i++ {
			tx.Create(&Product{Code: i, Price: i * 100})
		}

		db.Model(&Product{}).Count(&beforeCount)
		tx.Rollback() // Rollback is tracked
		db.Model(&Product{}).Count(&afterCount)

		// Custom step annotation
		trace.Step(ctx, "TX TEST-Rollback", fmt.Sprintf("RollBack Test : input - %d,  rollback - %d", beforeCount, afterCount), 1, 1)

		tx = db.Begin()

		for i := 0; i < 100; i++ {
			tx.Create(&Product{Code: i, Price: i * 100})
		}

		db.Model(&Product{}).Count(&beforeCount)
		tx.Commit() // Commit is tracked
		db.Model(&Product{}).Count(&afterCount)

		trace.Step(ctx, "TX TEST-Commit", fmt.Sprintf("Commit Test : input - %d, commit - %d", beforeCount, afterCount), 1, 1)
	}))

	// ============================================================
	// Case 8: Multi-Goroutine DB Operations with Error Tracking
	// ============================================================
	// Demonstrates:
	// - Concurrent DB operations in goroutines
	// - Error tracking with trace.Error()
	// - Transaction handling per goroutine
	//
	// IMPORTANT: Each goroutine should have its own DB connection
	// or use context properly for tracing.
	http.HandleFunc("/DBTxTestMulti", trace.Func(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		for i := 0; i < 100; i++ {
			go func(i int) {
				db, err := whatapgorm.OpenWithContext(sqlite.Open("test.db"), &gorm.Config{}, ctx)
				if err != nil {
					panic("failed to connect to Db")
				}
				tx := db.Begin()

				defer func() {
					if r := recover(); r != nil {
						trace.Step(ctx, "GormV2 Message", "Tranaction Rollback", 0, 0)
						tx.Rollback()
					}
				}()

				size := 10

				// Write Lock case - may cause errors with SQLite
				for j := 0; j < size; j++ {
					code := i*size + j
					res := tx.Create(&Product{Code: code, Price: code * 100})
					if res.Error != nil {
						fmt.Println(res.Error)
						// Track error with WhaTap
						trace.Error(ctx, res.Error)
						trace.Step(ctx, "GormV2 Message", "Tranaction Rollback", 0, 0)
						tx.Rollback()
						return
					}
				}
				tx.Commit()
			}(i)
		}
	}))

	fmt.Println("Start :", port, ", Agent Udp Port:", udpPort)
	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

}
