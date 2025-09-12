# Multi-Transaction (Distributed Tracing)

## Key Points

### When using `trace.Start` or `trace.StartWithContext`:
* You need logic to parse the incoming request headers.
    * `trace.UpdateMtrace()`
* When making an external API call, you need to add the WhaTap-specific header.
    * `trace.GetMtrace()`
    * `header.Set()`

### Regarding `trace.StartWithContext`:
* `trace.StartWithContext(ctx context.Context, name)`
* Note: This function is somewhat special and could potentially be confusing for users.
* The `ctx` passed to this function **must** be a context that has already been created or processed by a WhaTap API. This ensures it internally contains the necessary WhaTap trace context.

    ```go
    ctx, _ = trace.NewTraceContext(ctx)
    ctx, _ = trace.StartWithContext(ctx, fmt.Sprintf("%s_%d", "/trace3", depth))
    ```

---

## `distributed_tracing/server/server.go`

### Makefile
* Generates the `bin/mtrace` executable.

### bin/run.sh
* **`whatap.conf`**
    * Modify your license key and server host information in this file.
* **Execution**
    * This script starts four separate servers on ports 8080, 8081, 8082, and 8083.
* **Test URLs**
    * `http://localhost:8080/trace1_0`
        * Multi-transaction trace flow:
        * `trace1_0` → `trace1_1` → `trace1_2` → `trace1_3`
    * `http://localhost:8080/trace2_0`
        * Multi-transaction trace flow:
        * `trace2_0` → `trace2_1` → `trace2_2` → `trace2_3`
    * `http://localhost:8080/trace3_0`
        * Multi-transaction trace flow:
        * `trace3_0` → `trace3_1` → `trace3_2` → `trace3_3`

---

### trace1
* Example using the `trace.StartWithRequest` function.

### trace2
* Example using the `trace.Start` function.

### trace3
* Example using the `trace.StartWithContext` function.
