# Gormv2(https://github.com/go-gorm/gorm)

Grom 프레임워크를 통해 처리되는 DB Connection 및 SQL을 추적합니다.

# 일반적인 추적

gorm.Open 대신에 whatapgorm.OpenWithContext 함수를 사용합니다.
전달하는 context는 내부에 whatap TraceCtx를 포함해야 합니다.
trace.Start()를 통해 TraceCtx는 생성됩니다.


```

import (
	"net/http"

	"github.com/whatap/go-api/instrumentation/github.com/go-gorm/gorm/whatapgorm"
	"github.com/whatap/go-api/trace"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	whatapConfig := make(map[string]string)
	trace.Init(whatapConfig)
	defer trace.Shutdown()

	http.HandleFunc("/InsertAndDelete", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		db, err := whatapgorm.OpenWithContext(sqlite.Open("test.db"), &gorm.Config{}, ctx)
		if err != nil {
			panic("Db 연결에 실패하였습니다.")
		}

		for i := 0; i < 100; i++ {
			db.Create(&Product{Code: i, Price: i * 100})
		}

		db.Unscoped().Delete(&Product{}, "Code >= ? AND Code < ?", 0, 100)

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
        "gorm.io/driver/mysql"
        "gorm.io/gorm"
)


func main() {
	whatapConfig := make(map[string]string)
	trace.Init(whatapConfig)
	defer trace.Shutdown()

	http.HandleFunc("/WhatapDriverTest", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		dbConn, err := whatapsql.OpenContext(ctx, "mysql", dataSource)
		db, err := gorm.Open(mysql.New(mysql.Config{Conn: dbConn}), &gorm.Config{})
		if err != nil {
			panic("Db 연결에 실패하였습니다.")
		}
		for i := 0; i < 100; i++ {
			db.Create(&Product{Code: i, Price: i * 100})
		}

	})

	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
```
