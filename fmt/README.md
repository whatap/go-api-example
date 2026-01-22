# fmt Package Instrumentation

WhaTap instrumentation guide for Go standard `fmt` package.

## Overview

`whatapfmt` transforms `fmt.Print/Printf/Println` calls to collect stdout output with transaction context (txid, mtid, gid).

## Quick Start

### Automatic Instrumentation (Recommended)

```bash
# No code changes needed - whatap-go-inst transforms fmt calls automatically
whatap-go-inst go build ./...
```

### Manual Instrumentation

```go
import "github.com/whatap/go-api/instrumentation/fmt/whatapfmt"

// Use whatapfmt instead of fmt for Print functions
whatapfmt.Println("Hello")
whatapfmt.Printf("Count: %d\n", 42)
```

## Before / After

| Original | Instrumented | Notes |
|----------|--------------|-------|
| `fmt.Print(...)` | `whatapfmt.Print(...)` | stdout + txid linked |
| `fmt.Printf(...)` | `whatapfmt.Printf(...)` | stdout + txid linked |
| `fmt.Println(...)` | `whatapfmt.Println(...)` | stdout + txid linked |
| `fmt.Sprintf(...)` | (not transformed) | Returns string, no output |
| `fmt.Sprint(...)` | (not transformed) | Returns string, no output |

## Configuration

```ini
# whatap.conf
logsink_enabled=true
logsink_fmt_enabled=true   # Enable whatapfmt collection

# Category is shared with ProxyStream stdout
# logsink_category_stdout=AppStdOut (default)
```

### Avoiding Duplicate Collection

| logsink_stdout_enabled | logsink_fmt_enabled | Result |
|------------------------|---------------------|--------|
| true | false | Pipe collects all stdout (no txid) |
| **false** | **true** | **whatapfmt only (with txid) - Recommended** |
| true | true | **Duplicate collection!** |
| false | false | No stdout collection |

## How It Works

1. `whatapfmt.Println("msg")` is called
2. Inside whatapfmt:
   - Get goroutine ID via `gid.GetGID()`
   - Find transaction context via `trace.GetGIDTraceCtx(gid)`
   - Send to logsink with txid/mtid/gid
   - Call original `fmt.Println("msg")`
3. Log appears in WhaTap with transaction linkage

## Features

- **Same signatures**: Drop-in replacement for fmt functions
- **Transaction linking**: Logs are linked to HTTP transactions via goroutine ID
- **Zero overhead when disabled**: If `logsink_fmt_enabled=false`, just calls original fmt
- **Automatic with whatap-go-inst**: No manual code changes needed

## Running the Example

```bash
cd go-api-example/fmt

# With whatap.conf in current directory
go run fmt.go

# Or with whatap-go-inst (if using original fmt code)
whatap-go-inst go run fmt.go
```

Test endpoints:
```bash
curl http://localhost:8080/
curl http://localhost:8080/print-test
curl http://localhost:8080/health
```

## Log Output

With proper configuration, logs will include transaction context:

```
{
  "category": "AppStdOut",
  "content": "Request: GET /print-test",
  "@txid": 1234567890,
  "@gid": 42
}
```
