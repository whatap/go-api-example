# GORM v2 (gorm.io/gorm) Instrumentation Example

This example demonstrates how to instrument GORM v2 applications using WhaTap Go API.

## Overview

The `whatapgorm` package provides automatic tracing for GORM operations:

- **Connection Tracking**: Open/OpenWithContext track connection establishment (StartOpen)
- **Query Monitoring**: All GORM queries (Create, Find, Update, Delete) are traced
- **Transaction Support**: Begin/Commit/Rollback are tracked
- **Error Tracking**: Errors are captured and linked to transactions

## Prerequisites

- Go 1.18+
- Database (SQLite, MySQL, PostgreSQL, etc.)
- WhaTap Go Agent installed

## Installation

```bash
go get github.com/whatap/go-api
go get github.com/whatap/go-api/instrumentation/github.com/go-gorm/gorm/whatapgorm
go get gorm.io/gorm
go get gorm.io/driver/sqlite  # or mysql, postgres, etc.
```

## Quick Start

### Method 1: whatapgorm.Open (Recommended)

**Before (Original Code):**
```go
import "gorm.io/gorm"

db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
```

**After (Instrumented Code):**
```go
import "github.com/whatap/go-api/instrumentation/github.com/go-gorm/gorm/whatapgorm"

db, err := whatapgorm.Open(sqlite.Open("test.db"), &gorm.Config{})
```

### Method 2: OpenWithContext (Per-Request Connection)

```go
http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
    ctx, _ := trace.StartWithRequest(r)
    defer trace.End(ctx, nil)

    // Open connection with transaction context
    db, err := whatapgorm.OpenWithContext(sqlite.Open("test.db"), &gorm.Config{}, ctx)
    if err != nil {
        trace.Error(ctx, err)
        return
    }

    // All queries are automatically linked to HTTP transaction
    db.Create(&Product{Code: 1, Price: 100})
})
```

### Method 3: WithContext (Shared Connection)

```go
// Create shared connection at startup
serviceDB, _ := whatapgorm.Open(sqlite.Open("test.db"), &gorm.Config{})

http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
    ctx, _ := trace.StartWithRequest(r)
    defer trace.End(ctx, nil)

    // Add transaction context to shared connection
    db := whatapgorm.WithContext(ctx, serviceDB)

    // Now queries are linked to HTTP transaction
    var products []Product
    db.Find(&products)
})
```

### Method 4: Using whatapsql with GORM

For MySQL, PostgreSQL, etc., you can use whatapsql for connection:

```go
import (
    "github.com/whatap/go-api/instrumentation/database/sql/whatapsql"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

http.HandleFunc("/mysql", func(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Use whatapsql for connection with tracing
    sqlDB, err := whatapsql.OpenContext(ctx, "mysql", dsn)
    if err != nil {
        return
    }

    // Pass to GORM
    db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB}), &gorm.Config{})
    if err != nil {
        return
    }

    db.AutoMigrate(&Product{})
})
```

## Instrumentation Methods

| Original Function | WhaTap Function | Notes |
|-------------------|-----------------|-------|
| `gorm.Open()` | `whatapgorm.Open()` | Global connection |
| - | `whatapgorm.OpenWithContext()` | Per-request connection |
| `db.WithContext()` | `whatapgorm.WithContext()` | Add context to existing DB |

## Examples in This File

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/WhatapDriverTest` | whatapsql + GORM | Use whatapsql for MySQL |
| `/InsertAndDelete` | OpenWithContext | Insert and delete operations |
| `/InsertAndUpdate` | OpenWithContext + Transaction | Insert and update with Begin/Commit |
| `/Select` | Global DB | Select without context |
| `/SelectWithContext` | WithContext | Select with transaction context |
| `/DeleteAll` | OpenWithContext | Delete all records |
| `/DBTxTest` | Transaction | Rollback and Commit test |
| `/DBTxTestMulti` | Concurrent | Multi-goroutine transactions |

## Key Concepts

### Global vs Per-Request Connection

| Method | Use Case | Context Tracking |
|--------|----------|------------------|
| `Open()` + `WithContext()` | Connection pooling | Manual per-request |
| `OpenWithContext()` | Simple apps, short-lived | Automatic |

### Transaction Example

```go
ctx := r.Context()
db, _ := whatapgorm.OpenWithContext(sqlite.Open("test.db"), &gorm.Config{}, ctx)

tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

// Operations in transaction
tx.Create(&Product{Code: 1, Price: 100})
tx.Create(&Product{Code: 2, Price: 200})

// Commit or Rollback
if err != nil {
    tx.Rollback()
    return
}
tx.Commit()
```

### Error Tracking

```go
result := db.Create(&Product{Code: 1, Price: 100})
if result.Error != nil {
    trace.Error(ctx, result.Error)  // Record error in trace
    return
}
```

## Running the Example

```bash
# Run with SQLite (default)
go run gorm.go -whatap

# Run with MySQL
go run gorm.go -whatap -ds "user:pass@tcp(localhost:3306)/dbname"

# Test endpoints
curl http://localhost:8080/InsertAndDelete
curl http://localhost:8080/SelectWithContext
curl http://localhost:8080/DBTxTest
```

## jinzhu/gorm (v1) Migration

If you're using the old `github.com/jinzhu/gorm`:

| Old (v1) | New (v2) |
|----------|----------|
| `github.com/jinzhu/gorm` | `gorm.io/gorm` |
| `gorm.Open("sqlite3", ...)` | `gorm.Open(sqlite.Open(...), &gorm.Config{})` |
| `whatapgorm.Open("sqlite3", ...)` | `whatapgorm.Open(sqlite.Open(...), &gorm.Config{})` |

See [jinzhu/gorm example](../../jinzhu/gorm/) for v1 instrumentation.

## See Also

- [WhaTap Go API Documentation](https://docs.whatap.io/golang/install-agent)
- [GORM Documentation](https://gorm.io/docs/)
- [jinzhu/gorm (v1) Example](../../jinzhu/gorm/)
- [sqlx Example](../../jmoiron/sqlx/)
