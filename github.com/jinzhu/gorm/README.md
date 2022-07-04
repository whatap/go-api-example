# Gormv1(https://github.com/jinzhu/gorm)

grom 프레임워크를 통해 처리되는 DB Connection 및 SQL을 추적합니다.

# 일반적인 추적

gorm.Open 대신에 whatapgorm.OpenWithContext 함수를 사용합니다.
전달하는 context는 내부에 whatap TraceCtx를 포함해야 합니다.
trace.Start()를 통해 TraceCtx는 생성됩니다.


```

import (
	"net/http"

	"github.com/whatap/go-api/instrumentation/github.com/go-gorm/gorm/whatapgorm"
	"github.com/whatap/go-api/trace"
	_ "github.com/mattn/go-sqlite3"
	"github.com/jinzhu/gorm"
)

func main() {
	whatapConfig := make(map[string]string)
	trace.Init(whatapConfig)
	defer trace.Shutdown()

	http.HandleFunc("/InsertAndDelete", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		db, err := whatapgorm.OpenWithContext(ctx, "sqlite3", "test.db")
		defer db.Close()
		if err != nil {
			trace.Error(ctx, err)
			panic("Gorm Open Fail")
		}

		for i := 0; i < 100; i++ {
			db.Create(&Product{Code: i, Price: i * 100})
		}

		db.Unscoped().Delete(Product{}, "Code >= ? AND Code < ?", 0, 100)
	})


	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

```


# whatapsql driver를 통해 추적


whatapsql의 OpenContext 함수를 통해 만들어진 Connection을 통해 추적합니다.
gorm Open시 위에서 만들어진 만들어진 Connection을 전달하여 사용합니다.


```

import (
        "net/http"
        "github.com/whatap/go-api/instrumentation/database/sql/whatapsql"
        "github.com/whatap/go-api/trace"
	"github.com/jinzhu/gorm"
	_ "github.com/go-sql-driver/mysql"
)


func main() {
	whatapConfig := make(map[string]string)
	trace.Init(whatapConfig)
	defer trace.Shutdown()

	http.HandleFunc("/WhatapDriverTest", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		var conn gorm.SQLCommon
		var err error
		conn, err = whatapsql.OpenContext(ctx, "mysql", dataSource)
		if err != nil {
			trace.Error(ctx, err)
			panic("Whatapsql Open Fail")
		}
		db, err := gorm.Open("mysql", conn)
		if err != nil {
			trace.Error(ctx, err)
			panic("Gorm Open Fail")
		}
		for i := 0; i < 100; i++ {
			db.Create(&Product{Code: i, Price: i * 100})
		}
	})


	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
```
