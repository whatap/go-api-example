package main

import (
	"bytes"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/whatap/go-api/instrumentation/database/sql/whatapsql"
	"github.com/whatap/go-api/instrumentation/net/http/whataphttp"
	"github.com/whatap/go-api/trace"
)

const (
	MYSQL_DRIVER_NAME  = "mysql"
	MSSQL_DRIVER_NAME  = "mssql"
	ORACLE_DRIVER_NAME = "godror"
	PGSQL_DRIVER_NAME  = "postgres"
)

func main() {
	portPtr := flag.Int("p", 8080, "web port. default 8080  ")
	udpPortPtr := flag.Int("up", 6600, "agent port(udp). defalt 6600 ")
	dataSourcePtr := flag.String("ds", "doremimaker:doremimaker@tcp(phpdemo:3306)/doremimaker", " dataSourceName ")
	flag.Parse()
	port := *portPtr
	udpPort := *udpPortPtr
	dataSource := *dataSourcePtr

	config := make(map[string]string)
	config["net_udp_port"] = fmt.Sprintf("%d", udpPort)
	trace.Init(config)
	defer trace.Shutdown()

	serviceDB, err := whatapsql.OpenContext(context.Background(), MYSQL_DRIVER_NAME, dataSource)
	if err != nil {
		fmt.Println("Error service whatapsql.Open ", err)
		return
	}
	defer serviceDB.Close()

	http.HandleFunc("/index", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		w.Header().Add("Content-Type", "text/html")
		buffer.WriteString("/index <br/>Test Body<hr/>")

		ctx := r.Context()
		fmt.Println("Request -", r)
		db, err := whatapsql.OpenContext(ctx, MYSQL_DRIVER_NAME, dataSource)
		if err != nil {
			fmt.Println("Error whatapsql.Open ", err)
			return
		}
		defer db.Close()

		var id int
		var subject string
		query := "select id, subject from tbl_faq limit 10"
		if rows, err := db.QueryContext(ctx, query); err == nil {
			defer rows.Close()
			for rows.Next() {
				err := rows.Scan(&id, &subject)
				if err != nil {
					break
				}
				fmt.Println(id, subject)
				buffer.WriteString(fmt.Sprintln(id, subject, "<br>"))
			}
		} else {
			fmt.Println("Error db.QueryContext ", err)
			return
		}

		_, _ = w.Write(buffer.Bytes())

		fmt.Println("Response -", r.Response)

	}))

	http.HandleFunc("/query", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		w.Header().Add("Content-Type", "text/html")
		buffer.WriteString("/index <br/>Test Body<hr/>")

		ctx := r.Context()
		fmt.Println("Request -", r)
		db, err := whatapsql.OpenContext(ctx, MYSQL_DRIVER_NAME, dataSource)
		if err != nil {
			fmt.Println("Error whatapsql.Open ", err)
			return
		}
		defer db.Close()

		// 복수 Row를 갖는 SQL 쿼리
		var id int
		var subject string
		rows, err := db.Query("select id, subject from tbl_faq limit 10")
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

		// 복수 Row를 갖는 SQL 쿼리
		rows, err = db.QueryContext(ctx, "select id, subject from tbl_faq limit 10")
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

		_, _ = w.Write(buffer.Bytes())

		fmt.Println("Response -", r.Response)

	}))

	http.HandleFunc("/queryRow", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		w.Header().Add("Content-Type", "text/html")
		buffer.WriteString("/index <br/>Test Body<hr/>")

		ctx := r.Context()
		fmt.Println("Request -", r)
		db, err := whatapsql.OpenContext(ctx, MYSQL_DRIVER_NAME, dataSource)
		if err != nil {
			fmt.Println("Error whatapsql.Open ", err)
			return
		}
		defer db.Close()
		var id int
		var subject string

		row := db.QueryRow("select id, subject from tbl_faq limit 1")
		// Scan and close
		if err := row.Scan(&id, &subject); err != nil {
			fmt.Println("Error Row.Scan ", err)
		} else {
			fmt.Println(id, subject)
			buffer.WriteString(fmt.Sprintln(id, subject, "<br>"))
		}

		row = db.QueryRowContext(ctx, "select id, subject from tbl_faq limit 1")
		// Scan and close
		if err := row.Scan(&id, &subject); err != nil {
			fmt.Println("Error db.QueryRowContext")
		} else {
			fmt.Println(id, subject)
			buffer.WriteString(fmt.Sprintln(id, subject, "<br>"))
		}

		_, _ = w.Write(buffer.Bytes())

		fmt.Println("Response -", r.Response)

	}))

	http.HandleFunc("/prepare", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		w.Header().Add("Content-Type", "text/html")
		buffer.WriteString("/index <br/>Test Body<hr/>")

		ctx := r.Context()
		fmt.Println("Request -", r)
		db, err := whatapsql.OpenContext(ctx, MYSQL_DRIVER_NAME, dataSource)
		if err != nil {
			fmt.Println("Error whatapsql.Open")
			return
		}
		defer db.Close()

		var id int
		var subject string
		params := make([]interface{}, 0)
		params = append(params, 8)
		params = append(params, 1)

		query := "select id, subject from tbl_faq where id in (?,?) limit 10"

		if stmt, err := db.Prepare(query); err == nil {
			if rows, err1 := stmt.Query(params...); err1 == nil {
				defer rows.Close()
				for rows.Next() {
					err2 := rows.Scan(&id, &subject)
					if err2 != nil {
						break
					}
					fmt.Println(id, subject)
					buffer.WriteString(fmt.Sprintln(id, subject, "<br>"))
				}
			} else {
				fmt.Println("Error stmt.Query ", err1)
			}

			if rows, err1 := stmt.QueryContext(ctx, params...); err == nil {
				defer rows.Close() //반드시 닫는다 (지연하여 닫기)
				for rows.Next() {
					err2 := rows.Scan(&id, &subject)
					if err2 != nil {
						break
					}
					fmt.Println(id, subject)
					buffer.WriteString(fmt.Sprintln(id, subject, "<br>"))
				}
			} else {
				fmt.Println("Error stmt.QueryContext ", err1)
			}
		} else {
			fmt.Println("Error db.Prepare ", err)
		}

		if stmt, err := db.PrepareContext(ctx, query); err == nil {
			if rows, err1 := stmt.Query(params...); err1 == nil {
				defer rows.Close() //반드시 닫는다 (지연하여 닫기)

				for rows.Next() {
					err2 := rows.Scan(&id, &subject)
					if err2 != nil {
						break
					}
					fmt.Println(id, subject)
					buffer.WriteString(fmt.Sprintln(id, subject, "<br>"))
				}
			} else {
				fmt.Println("Error stmt.QueryContext ", err1)
			}

			if rows, err1 := stmt.QueryContext(ctx, params...); err1 == nil {
				defer rows.Close() //반드시 닫는다 (지연하여 닫기)

				for rows.Next() {
					err2 := rows.Scan(&id, &subject)
					if err2 != nil {
						break
					}
					fmt.Println(id, subject)
					buffer.WriteString(fmt.Sprintln(id, subject, "<br>"))
				}
			} else {
				fmt.Println("Error stmt.QueryContext ", err1)
			}

		} else {
			fmt.Println("Error db.PrepareContext ", err)
		}

		query = "select id, subject from tbl_faq where id in (?,?) limit 1"

		if stmt, err := db.Prepare(query); err == nil {
			row := stmt.QueryRow(params...)
			if err1 := row.Scan(&id, &subject); err1 == nil {
				fmt.Println(id, subject)
				buffer.WriteString(fmt.Sprintln(id, subject, "<br>"))
			} else {
				fmt.Println("Error row.Scan ", err1)
			}

			row = stmt.QueryRowContext(ctx, params...)
			if err1 := row.Scan(&id, &subject); err1 == nil {
				fmt.Println(id, subject)
				buffer.WriteString(fmt.Sprintln(id, subject, "<br>"))
			} else {
				fmt.Println("Error row.Scan ", err1)
			}
		} else {
			fmt.Println("Error db.Prepare ", err)
		}

		query = "update tbl_faq set subject='aaa' where id in (?,?) limit 1"
		if stmt, err := db.Prepare(query); err == nil {
			if res, err1 := stmt.Exec(params...); err1 == nil {
				fmt.Println("Result ", res)
			} else {
				fmt.Println("Error stmt.Exec ", err1)
			}

			if res, err1 := stmt.ExecContext(ctx, params...); err1 == nil {
				fmt.Println("Result ", res)
			} else {
				fmt.Println("Error stmt.Exec ", err1)
			}
		} else {
			fmt.Println("Error db.Prepare ", err)
		}
		_, _ = w.Write(buffer.Bytes())

		fmt.Println("Response -", r.Response)
	}))

	http.HandleFunc("/named", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		w.Header().Add("Content-Type", "text/html")
		buffer.WriteString("/index <br/>Test Body<hr/>")

		ctx := r.Context()
		fmt.Println("Request -", r)
		db, err := whatapsql.OpenContext(ctx, MYSQL_DRIVER_NAME, dataSource)
		if err != nil {
			fmt.Println("Error whatapsql.Open")
			http.Error(w, fmt.Sprintln("Error whatapsql.Open", err), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		query := "select id, subject from tbl_faq where id in (?, ?) limit 10"
		var id int
		var subject string
		params := make([]interface{}, 0)
		params = append(params, sql.Named("idx1", 8))
		params = append(params, sql.Named("idx2", 1))
		if stmt, err := db.Prepare(query); err == nil {
			if rows, err1 := stmt.QueryContext(ctx, params...); err1 == nil {
				defer rows.Close() //반드시 닫는다 (지연하여 닫기)

				for rows.Next() {
					err := rows.Scan(&id, &subject)
					if err != nil {
						break
					}
					fmt.Println(id, subject)
					buffer.WriteString(fmt.Sprintln(id, subject, "<br>"))
				}

			} else {
				fmt.Println("Error db.QueryContext", err)
				http.Error(w, fmt.Sprintln("Error db.QueryContext", err), http.StatusInternalServerError)
			}

		} else {
			fmt.Println("Error db.Prepard ", err)
			http.Error(w, fmt.Sprintln("Error db.Prepared", err), http.StatusInternalServerError)
		}
		// 복수 Row를 갖는 SQL 쿼리
		_, _ = w.Write(buffer.Bytes())

		fmt.Println("Response -", r.Response)
	}))

	http.HandleFunc("/exec", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		w.Header().Add("Content-Type", "text/html")
		buffer.WriteString("/index <br/>Test Body<hr/>")

		ctx := r.Context()
		if _, traceCtx := trace.GetTraceContext(ctx); traceCtx != nil {
			fmt.Println("Txid=", traceCtx.Txid)
		}
		fmt.Println("Request -", r)
		db, err := whatapsql.OpenContext(ctx, MYSQL_DRIVER_NAME, dataSource)
		if err != nil {
			fmt.Println("Error whatapsql.Open")
			return
		}
		defer db.Close()

		params := make([]interface{}, 0)
		params = append(params, 8)
		params = append(params, 1)

		query := "update tbl_faq set subject = 'aaa' where id in (?,?)"
		if res, err := db.Exec(query, params...); err == nil {
			fmt.Println("Result ", res)
			buffer.WriteString(fmt.Sprintln("Result ", res, "<br>"))
		} else {
			fmt.Println("Error db.Exec ", err)
		}

		if res, err := db.ExecContext(ctx, query, params...); err == nil {
			fmt.Println("Result ", res)
			buffer.WriteString(fmt.Sprintln("Result ", res, "<br>"))
		} else {
			fmt.Println("Error db.ExecContext ", err)
		}

		_, _ = w.Write(buffer.Bytes())

		fmt.Println("Response -", r.Response)
	}))

	http.HandleFunc("/tx", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		w.Header().Add("Content-Type", "text/html")
		buffer.WriteString("/index <br/>Test Body<hr/>")
		ctx := r.Context()
		fmt.Println("Request -", r)

		db, err := whatapsql.OpenContext(ctx, MYSQL_DRIVER_NAME, dataSource)
		if err != nil {
			fmt.Println("Error whatapsql.Open")
			return
		}
		defer db.Close()
		var (
			query   = ""
			id      = 0
			subject = ""
		)
		params := make([]interface{}, 0)
		params = append(params, 8)
		params = append(params, 1)

		if tx, err := db.BeginTx(ctx, nil); err == nil {

			query = "update tbl_faq set subject = 'bbb' where id in (?,?)"
			if res, err := tx.Exec(query, params...); err != nil {
				fmt.Println("Error tx.Exec ", err)
			} else {
				fmt.Println("tx.Exec  Result ", res)
			}

			query = "select id, subject from tbl_faq where id in (?,?)"

			rows, err := tx.Query(query, params...)
			if err != nil {
				fmt.Println("Error tx.Query ", err)
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

			query = "update tbl_faq set subject = 'ccc' where id in (?,?)"
			if res, err := tx.ExecContext(ctx, query, params...); err != nil {
				fmt.Println("Error tx.ExecContext ", err)
			} else {
				fmt.Println("tx.ExecContext Result", res)
			}

			query = "select id, subject from tbl_faq where id in (?,?)"

			rows, err = tx.QueryContext(ctx, query, params...)
			if err != nil {
				fmt.Println("Error tx.QueryContext ", err)
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

			tx.Commit()

		} else {
			fmt.Println("Error tx.BeginTx ", err)
		}

		_, _ = w.Write(buffer.Bytes())

		fmt.Println("Response -", r.Response)
	}))

	http.HandleFunc("/service/index", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		w.Header().Add("Content-Type", "text/html")
		buffer.WriteString("/index <br/>Test Body<hr/>")

		ctx := r.Context()
		fmt.Println("Request -", r)

		var id int
		var subject string
		query := "select id, subject from tbl_faq limit 10"
		if rows, err := serviceDB.QueryContext(ctx, query); err == nil {
			defer rows.Close()
			for rows.Next() {
				err := rows.Scan(&id, &subject)
				if err != nil {
					break
				}
				fmt.Println(id, subject)
				buffer.WriteString(fmt.Sprintln(id, subject, "<br>"))
			}
		} else {
			fmt.Println("Error db.QueryContext ", err)
			return
		}

		buffer.WriteString("DB Statas <hr/>")
		buffer.WriteString(fmt.Sprintln(serviceDB.Stats()))
		_, _ = w.Write(buffer.Bytes())

		fmt.Println("Response -", r.Response)
	}))

	fmt.Println("Start :", port, ", Agent Udp Port:", udpPort)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		fmt.Println("Error ListenAndServe ", err)
	}
}
