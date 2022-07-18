package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"text/template"

	"github.com/whatap/go-api/instrumentation/database/sql/whatapsql"
	"github.com/whatap/go-api/instrumentation/github.com/go-gorm/gorm/whatapgorm"
	"github.com/whatap/go-api/trace"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Code  int
	Price int
}

type HTMLData struct {
	Title   string
	Content string
	//HTMLContent template.HTML
}

func main() {
	udpPortPtr := flag.Int("up", 6600, "agent port(udp). defalt 6600 ")
	portPtr := flag.Int("p", 8080, "web port. default 8080  ")
	dataSourcePtr := flag.String("ds", "doremimaker:doremimaker@tcp(phpdemo2:3306)/doremimaker", " dataSourceName ")
	flag.Parse()

	udpPort := *udpPortPtr
	dataSource := *dataSourcePtr

	port := *portPtr

	whatapConfig := make(map[string]string)
	whatapConfig["net_udp_port"] = fmt.Sprintf("%d", udpPort)
	trace.Init(whatapConfig)
	defer trace.Shutdown()

	templatePath := "templates/github.com/go-gorm/index.html"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tp, err := template.ParseFiles(templatePath)
		if err != nil {
			fmt.Println("Template not loaded, ", err)
			return
		}

		data := &HTMLData{}
		data.Title = "GormV2 Test Page"
		data.Content = r.RequestURI

		tp.Execute(w, data)
	})

	// mysql, postgresql, mssql 등 기존 whatap driver 사용 케이스

	//Case 1. mysql
	http.HandleFunc("/WhatapDriverTest", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		dbConn, err := whatapsql.OpenContext(ctx, "mysql", dataSource)
		db, err := whatapgorm.Open(mysql.New(mysql.Config{Conn: dbConn}), &gorm.Config{})
		if err != nil {
			panic("Db 연결에 실패하였습니다.")
		}

		db.AutoMigrate(&Product{})

		for i := 0; i < 100; i++ {
			db.Create(&Product{Code: i, Price: i * 100})
		}

	})

	// gorm hooking 케이스
	serviceDB, err := whatapgorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("Db 연결에 실패하였습니다.")
	}

	//Case 1. Not Context + AutoMigrate
	serviceDB.AutoMigrate(&Product{})

	//Case 2. Insert Query + Delete Query
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

	//Case 3. Insert Query + Update Query
	http.HandleFunc("/InsertAndUpdate", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		db, err := whatapgorm.OpenWithContext(sqlite.Open("test.db"), &gorm.Config{}, ctx)
		if err != nil {
			panic("Db 연결에 실패하였습니다.")
		}

		for i := 0; i < 100; i++ {
			db.Create(&Product{Code: i, Price: i * 100})
		}

		for i := 0; i < 100; i++ {
			var product Product
			tx := db.Begin()
			tx.First(&product, "Code = ?", i)
			tx.Model(&product).Update("price", product.Price*100)
			tx.Commit()
		}
	})

	//Case 4. Non Context + Select
	http.HandleFunc("/Select", func(w http.ResponseWriter, r *http.Request) {
		var products []Product
		var buffer bytes.Buffer
		serviceDB.Find(&products, "1 = 1")

		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		for i, product := range products {
			buffer.WriteString(fmt.Sprintf("Index : %d, Code : %d, Price : %d </br>", i, product.Code, product.Price))
		}

		buffer.WriteString("</body></html>")

		_, _ = w.Write(buffer.Bytes())
	})

	//Case 5. WithContext + Select
	http.HandleFunc("/SelectWithContext", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		db := whatapgorm.WithContext(ctx, serviceDB)

		var products []Product
		var buffer bytes.Buffer
		db.Find(&products, "1 = 1")

		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		for i, product := range products {
			buffer.WriteString(fmt.Sprintf("Index : %d, Code : %d, Price : %d </br>", i, product.Code, product.Price))
		}

		buffer.WriteString("</body></html>")

		_, _ = w.Write(buffer.Bytes())
	})

	//Case 5. Delete ALl
	http.HandleFunc("/DeleteAll", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		db, err := whatapgorm.OpenWithContext(sqlite.Open("test.db"), &gorm.Config{}, ctx)
		if err != nil {
			panic("Db 연결에 실패하였습니다.")
		}

		db.Unscoped().Delete(&Product{}, "1 = 1")
	})

	//Case 6. DB Transaction
	http.HandleFunc("/DBTxTest", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		db, err := whatapgorm.OpenWithContext(sqlite.Open("test.db"), &gorm.Config{}, ctx)
		if err != nil {
			panic("Db 연결에 실패하였습니다.")
		}

		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				trace.Step(ctx, "GormV2 Message", "Tranaction Rollback", 0, 0)
				tx.Rollback()
			}
		}()

		var beforeCount int64
		var afterCount int64

		for i := 0; i < 100; i++ {
			tx.Create(&Product{Code: i, Price: i * 100})
		}

		db.Model(&Product{}).Count(&beforeCount)
		tx.Rollback()
		db.Model(&Product{}).Count(&afterCount)

		trace.Step(ctx, "TX TEST-Rollback", fmt.Sprintf("RollBack Test : input - %d,  rollback - %d", beforeCount, afterCount), 1, 1)

		tx = db.Begin()

		for i := 0; i < 100; i++ {
			tx.Create(&Product{Code: i, Price: i * 100})
		}

		db.Model(&Product{}).Count(&beforeCount)
		tx.Commit()
		db.Model(&Product{}).Count(&afterCount)

		trace.Step(ctx, "TX TEST-Commit", fmt.Sprintf("Commit Test : input - %d, commit - %d", beforeCount, afterCount), 1, 1)
	})

	//Case 7. Multi DB Connection + Error Trace
	http.HandleFunc("/DBTxTestMulti", func(w http.ResponseWriter, r *http.Request) {
		for i := 0; i < 100; i++ {
			go func(i int) {
				ctx, _ := trace.StartWithRequest(r)
				defer trace.End(ctx, nil)

				db, err := whatapgorm.OpenWithContext(sqlite.Open("test.db"), &gorm.Config{}, ctx)
				if err != nil {
					panic("Db 연결에 실패하였습니다.")
				}
				tx := db.Begin()

				defer func() {
					if r := recover(); r != nil {
						trace.Step(ctx, "GormV2 Message", "Tranaction Rollback", 0, 0)
						tx.Rollback()
					}
				}()

				size := 10

				// Write Lock 발생 Case
				for j := 0; j < size; j++ {
					code := i*size + j
					res := tx.Create(&Product{Code: code, Price: code * 100})
					if res.Error != nil {
						fmt.Println(res.Error)
						trace.Error(ctx, res.Error)
						trace.Step(ctx, "GormV2 Message", "Tranaction Rollback", 0, 0)
						tx.Rollback()
						return
					}
				}
				tx.Commit()
			}(i)
		}
	})

	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

}
