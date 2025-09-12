# General tracing

The DB connection tracing is made using the whatapsql.StartOpen(), whatapsql.End() function.
The query execution tracing is made using the whatap.sql.Start(), whatapsql.End() function.

```
func Start(ctx context.Context, dbhost, sql string) (*SqlCtx, error) 
func StartOpen(ctx context.Context, dbhost string) (*SqlCtx, error)
func StartWithParam(ctx context.Context, dbhost, sql string, param ...interface{}) (*SqlCtx, error)
func StartWithParamArray(ctx context.Context, dbhost, sql string, param []interface{}) (*SqlCtx, error)
func End(sqlCtx *SqlCtx, err error) error 
```

```
import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	whatapsql "github.com/whatap/go-api/sql"
)

func main() {
	config := make(map[string]string)
	trace.Init(config)
	//It must be executed before closing the app.
	defer trace.Shutdown()
	   
	sqlCtx, _ := whatapsql.StartOpen(ctx, dataSource)
	db, err := sql.Open(mysql, dataSource)
	whatapsql.End(sqlCtx, err)
	
	if err != nil {
		fmt.Println("Error whatapsql.Open ", err)
		return
	}
	defer db.Close()
	
	// SQL query with multiple rows
	var id int
	var subject string
	var query = "select id, subject from tbl_faq limit 10"
	
	sqlCtx, _ = whatapsql.Start(ctx, dataSource, query)
	rows, err := db.Query(query)
	whatapsql.End(sqlCtx, err)
	
	if err != nil {
		fmt.Println("Error db.QueryContext ", err)
		return
	}
	defer rows.Close() //Be sure to close it (delayed closing).
	
	for rows.Next() {
		err := rows.Scan(&id, &subject)
		if err != nil {
			break
		}
		fmt.Println(id, subject)
		buffer.WriteString(fmt.Sprintln(id, subject, "<br>"))
	}
}        
```

## Tracing through a whatap driver

It uses the whatapsql.OpenContext function instead of the sql.Open function in the database/sql package.\
It is recommended to use the function that delivers contexts such as PrepareContext, QueryContext, and ExecContext.

The context to deliver must include the whatap TraceCtx inside.\
TreaceCtx is created through trace.Start().

```
import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/whatap/go-api/instrumentation/database/sql/whatapsql"
)


func main() {
	config := make(map[string]string)
	trace.Init(config)
	defer trace.Shutdown()
	
	// whatap TraceCtx is created inside whataphttp.Func. 
	http.HandleFunc("/query", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		// Use the WhaTap Driver. It delivers the context with the whatap's TraceCtx. 
		db, err := whatapsql.OpenContext(ctx, "mysql", dataSource)
		if err != nil {
		fmt.Println("Error whatapsql.Open ", err)
		return
		}
		defer db.Close()
		
		... 
		query := "select id, subject from tbl_faq limit 10"
		
		// It delivers the context with whatap TraceCtx. 
		if rows, err := db.QueryContext(ctx, query); err == nil {
		...
		}
	}
	
	...
}
```

| Option Name                                                                      | Default Value | Data Type | Description                                                                    |
| -------------------------------------------------------------------------------- | ------------- | --------- | ------------------------------------------------------------------------------ |
| go.sql_profile_enabled | true          | bool      | It determines whether or not to collect the database/sql data. |
