package main

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	whatapsql "github.com/whatap/go-api/sql"
	"github.com/whatap/go-api/trace"
)

func getMysql(ctx context.Context) ([]string, error) {

	wCtx, _ := whatapsql.StartOpen(ctx, "doremimaker:doremimaker@tcp(192.168.56.101:3306)/doremimaker")
	db, err := sql.Open("mysql", "doremimaker:doremimaker@tcp(192.168.56.101:3306)/doremimaker")
	whatapsql.End(wCtx, err)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// 복수 Row를 갖는 SQL 쿼리
	var id int
	var subject string
	wCtx, _ = whatapsql.Start(ctx, "doremimaker:doremimaker@tcp(192.168.56.101:3306)/doremimaker", "select id, subject from tbl_faq limit 10")
	rows, err := db.QueryContext(ctx, "select id, subject from tbl_faq limit 10")
	whatapsql.End(wCtx, err)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //반드시 닫는다 (지연하여 닫기)

	result := make([]string, 0)

	for rows.Next() {
		err := rows.Scan(&id, &subject)
		if err != nil {
			return result, err
		}
		fmt.Println(id, subject)
		result = append(result, fmt.Sprintln(id, subject, "<br>"))
	}
	return result, nil
}

func main() {
	trace.Init(nil)
	defer trace.Shutdown()
	ctx := context.Background()
	wCtx, _ := trace.Start(ctx, "example-sql")
	if result, err := getMysql(wCtx); err != nil {
		fmt.Println("Error GetMysql() len=", len(result), ",err=", err)
	}
	trace.End(wCtx, err)
}
