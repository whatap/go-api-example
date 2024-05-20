# Gormv2(https://github.com/go-gorm/gorm)

It traces DB connections and SQLs that are processed through the Grom framework.

# General tracing

The whatapgorm.OpenWithContext function is used instead of gorm.Open.
The context to deliver must include the whatap TraceCtx inside.
TreaceCtx is created through trace.Start().

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
			panic("Db connection failed.")
		}

		for i := 0; i < 100; i++ {
			db.Create(&Product{Code: i, Price: i * 100})
		}

		db.Unscoped().Delete(&Product{}, "Code >= ? AND Code < ?", 0, 100)

	})


	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

```

# Tracing through the whatapsql driver

It traces through the connection created via the OpenContext function of whatapsql.
It is used by delivering the created connection upon the gorm Open.

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
			panic("DB connection failed.")
		}
		for i := 0; i < 100; i++ {
			db.Create(&Product{Code: i, Price: i * 100})
		}

	})

	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
```
