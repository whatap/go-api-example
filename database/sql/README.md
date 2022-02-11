database/sql 패키지의 sql.Open 함수 대신 whatapsql.OpenContext 함수를 사용합니다. 
PrepareContext, QueryContext, ExecContext 등 context를 전달하는 함수를 사용하기를 권장합니다. 

전달하는 context는 trace.Start()를 통해서 whatap TraceCtx 정보가 있어야합니다.

# 설치 안내 

```
import (
    _ "github.com/go-sql-driver/mysql"
	"github.com/whatap/go-api/instrumentation/database/sql/whatapsql"
)


func main() {
    config := make(map[string]string)
    trace.Init(config)
    defer trace.Shutdown()
    
    // whataphttp.Func 내부에서 whatap TraceCtx를 생성합니다. 
    http.HandleFunc("/query", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        
        // Use WhaTap Driver. whatap의 TraceCtx가 있는 context 를 전달합니다. 
        db, err := whatapsql.OpenContext(ctx, "mysql", dataSource)
    	if err != nil {
    		fmt.Println("Error whatapsql.Open ", err)
    		return
    	}
    	defer db.Close()

	    ... 
	    query := "select id, subject from tbl_faq limit 10"
	    
	    // whatap TraceCtx가 있는 context 를 전달합니다. 
	    if rows, err := db.QueryContext(ctx, query); err == nil {
		    ...
		}
	}
	
	...
}
```

|옵션명| 기본값| 테이터타입| 설명|
|----|----|----|----|
|go.sql_profile_enabled|true|bool|database/sql 정보 수집여부를 설정합니다. |

