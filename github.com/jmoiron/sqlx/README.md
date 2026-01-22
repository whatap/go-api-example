# sqlx Instrumentation Example

This example demonstrates how to instrument sqlx (github.com/jmoiron/sqlx) applications using WhaTap Go API.

## Overview

sqlx is an extension of database/sql that provides additional features like struct scanning. The `whatapsql` package provides automatic tracing for all database operations:

- **Connection Tracking**: OpenContext tracks connection establishment (StartOpen)
- **Query Monitoring**: All SQL queries are traced
- **Transaction Support**: Begin/Commit/Rollback are tracked
- **Error Tracking**: Errors are captured and linked to transactions

## Prerequisites

- Go 1.18+
- Database (PostgreSQL, MySQL, SQLite, etc.)
- WhaTap Go Agent installed

## Installation

```bash
go get github.com/whatap/go-api
go get github.com/whatap/go-api/instrumentation/database/sql/whatapsql
go get github.com/jmoiron/sqlx
go get github.com/lib/pq  # or your database driver
```

## Quick Start

### 1. Create Connection with whatapsql.OpenContext

**Before (Original Code):**
```go
import (
    "database/sql"
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
)

db, err := sql.Open("postgres", dsn)
sqlxDB := sqlx.NewDb(db, "postgres")
```

**After (Instrumented Code):**
```go
import (
    "github.com/jmoiron/sqlx"
    "github.com/whatap/go-api/instrumentation/database/sql/whatapsql"
    _ "github.com/lib/pq"
)

db, err := whatapsql.OpenContext(ctx, "postgres", dsn)
sqlxDB := sqlx.NewDb(db, "postgres")
```

### 2. Pass Context for Transaction Linkage

```go
http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
    ctx, _ := trace.StartWithRequest(r)
    defer trace.End(ctx, nil)

    // Create connection with transaction context
    db, err := whatapsql.OpenContext(ctx, "postgres", dsn)
    if err != nil {
        trace.Error(ctx, err)
        return
    }
    defer db.Close()

    sqlxDB := sqlx.NewDb(db, "postgres")

    // Use sqlx methods - all queries are traced
    var users []User
    err = sqlxDB.SelectContext(ctx, &users, "SELECT * FROM users")
})
```

### 3. Global Connection Pool

```go
// Create at startup (without HTTP context)
globalDB, _ := whatapsql.OpenContext(context.Background(), "postgres", dsn)
sqlxDB := sqlx.NewDb(globalDB, "postgres")

http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
    ctx, _ := trace.StartWithRequest(r)
    defer trace.End(ctx, nil)

    // Use Context methods to link to HTTP transaction
    var users []User
    err := sqlxDB.SelectContext(ctx, &users, "SELECT * FROM users")
})
```

## Instrumentation Methods

| sqlx Method | Context Method | Description |
|-------------|----------------|-------------|
| `db.Select()` | `db.SelectContext()` | Query multiple rows into struct |
| `db.Get()` | `db.GetContext()` | Query single row into struct |
| `db.Query()` | `db.QueryContext()` | Standard query |
| `db.QueryRow()` | `db.QueryRowContext()` | Query single row |
| `db.Exec()` | `db.ExecContext()` | Execute statement |
| `db.Prepare()` | `db.PrepareContext()` | Prepare statement |
| `db.BeginTx()` | - | Start transaction |

## Examples in This File

| Endpoint | Operation | Description |
|----------|-----------|-------------|
| `/query` | Select | Query with struct scanning |
| `/queryContext` | SelectContext | Context-aware query |
| `/queryRow` | QueryRow/QueryRowContext | Single row queries |
| `/prepare` | Prepare/PrepareContext | Prepared statements |
| `/named` | Named parameters | Named parameter queries |
| `/exec` | Exec/ExecContext | Execute statements |
| `/tx` | Transaction | Commit transaction |
| `/tx/rollback` | Transaction | Rollback transaction |
| `/tx/error` | Transaction | Error handling |
| `/service/index` | Global connection | Using global pool |
| `/service/selectContext` | SelectContext | Context with global pool |

## Key Concepts

### Per-Request vs Global Connection

| Approach | Use Case | Pros | Cons |
|----------|----------|------|------|
| Per-request | Simple apps | Automatic context | Connection overhead |
| Global pool | Production | Connection reuse | Manual context passing |

### Transaction Example

```go
ctx := r.Context()
db, _ := whatapsql.OpenContext(ctx, "postgres", dsn)

tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return
}
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

// Execute operations
_, err = tx.Exec("UPDATE users SET name = $1 WHERE id = $2", name, id)
if err != nil {
    tx.Rollback()
    return
}

rows, err := tx.Query("SELECT * FROM users WHERE id = $1", id)
// ... process rows

tx.Commit()
```

### Struct Scanning with sqlx

```go
type Person struct {
    Id   int       `db:"id"`
    Name string    `db:"name"`
    Adm  time.Time `db:"adm"`
}

// Select multiple rows
var people []Person
err := sqlxDB.SelectContext(ctx, &people, "SELECT * FROM people")

// Get single row
var person Person
err := sqlxDB.GetContext(ctx, &person, "SELECT * FROM people WHERE id = $1", 1)
```

## Running the Example

```bash
# Set up PostgreSQL
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=password --name postgres postgres

# Run the example
go run sqlx.go -whatap -ds "host=localhost port=5432 user=postgres password=password dbname=postgres sslmode=disable"

# Test endpoints
curl http://localhost:8080/query
curl http://localhost:8080/service/selectContext
curl http://localhost:8080/tx
```

## Supported Databases

| Database | Driver | DSN Example |
|----------|--------|-------------|
| PostgreSQL | `github.com/lib/pq` | `host=localhost port=5432 user=... dbname=...` |
| MySQL | `github.com/go-sql-driver/mysql` | `user:pass@tcp(localhost:3306)/dbname` |
| SQLite | `github.com/mattn/go-sqlite3` | `file:test.db` |
| MSSQL | `github.com/denisenkom/go-mssqldb` | `sqlserver://user:pass@localhost:1433` |

## See Also

- [WhaTap Go API Documentation](https://docs.whatap.io/golang/install-agent)
- [sqlx Documentation](https://github.com/jmoiron/sqlx)
- [GORM Example](../go-gorm/gorm/)
- [database/sql Example](../../database/sql/)
