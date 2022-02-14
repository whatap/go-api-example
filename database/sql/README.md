
# 일반적인 추적

DB Connection 추적은 whatapsql.StartOpen(), whatapsql.End() 함수로 추적합니다. 
쿼리 실행에 대한 추적은 whatap.sql.Start(), whatapsql.End() 함수로 추적합니다. 

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
	
	// 복수 Row를 갖는 SQL 쿼리
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
	defer rows.Close() //반드시 닫는다 (지연하여 닫기)
	
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

## whatapsql의 Driver를 통해서 추적

database/sql 패키지의 sql.Open 함수 대신 whatapsql.OpenContext 함수를 사용합니다.  
PrepareContext, QueryContext, ExecContext 등 context를 전달하는 함수를 사용하기를 권장합니다. 

전달하는 context는 내부에 whatap TraceCtx를 포함해야 합니다.  
trace.Start()를 통해 TraceCtx는 생성됩니다.

# 설치 안내 

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

