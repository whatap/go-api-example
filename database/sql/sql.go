// database/sql Manual Instrumentation Example
//
// This example demonstrates how to manually instrument database/sql applications
// using WhaTap Go API's low-level SQL tracking functions.
//
// INSTRUMENTATION OVERVIEW:
// This example shows TWO approaches to SQL tracking:
//
// METHOD 1: Manual whatapsql.Start/End (explicit control)
//   - Call whatapsql.Start(ctx, dataSource, query) before SQL execution
//   - Call whatapsql.End(sqlCtx, err) after SQL execution
//   - Use StartWithParam() or StartWithParamArray() for parameter tracking
//
// METHOD 2: Use whatapsql.OpenContext() (recommended, automatic)
//   - Replace sql.Open() with whatapsql.OpenContext()
//   - All queries are automatically tracked
//   - See other examples (gorilla/mux, gin, echo) for this approach
//
// FEATURES:
// - Connection tracking via whatapsql.StartOpen/End
// - Query tracking via whatapsql.Start/End
// - Parameter tracking via whatapsql.StartWithParam
// - Named parameter support via whatapsql.StartWithParamArray
// - Transaction tracking (Begin, Commit, Rollback)
// - Error tracking
//
// WHEN TO USE MANUAL TRACKING:
// - When you need fine-grained control over SQL tracking
// - When using connection pools with shared *sql.DB
// - When you want to track specific queries only
// - Legacy code where automatic instrumentation is difficult

package main

import (
	"bytes"
	//"context"
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/whatap/go-api/instrumentation/net/http/whataphttp"
	whatapsql "github.com/whatap/go-api/sql"
	"github.com/whatap/go-api/trace"
)

const (
	MYSQL_DRIVER_NAME  = "mysql"
	MSSQL_DRIVER_NAME  = "mssql"
	ORACLE_DRIVER_NAME = "godror"
	PGSQL_DRIVER_NAME  = "postgres"
)

type HTMLData struct {
	Title   string
	Content string
	//HTMLContent template.HTML
}

func main() {
	portPtr := flag.Int("p", 8080, "web port. default 8080  ")
	udpPortPtr := flag.Int("up", 6600, "agent port(udp). defalt 6600 ")
	dataSourcePtr := flag.String("ds", "doremimaker:doremimaker@tcp(phpdemo:3306)/doremimaker", " dataSourceName ")
	setWhatapPtr := flag.Bool("whatap", false, "set whatap")

	flag.Parse()
	port := *portPtr
	udpPort := *udpPortPtr
	dataSource := *dataSourcePtr
	IsWhatap := *setWhatapPtr

	// ============================================================
	// INSTRUMENTATION STEP 1: Initialize WhaTap Agent
	// ============================================================
	// Call trace.Init() at application startup.
	// Configuration can be passed via map or loaded from whatap.conf file.
	if IsWhatap {
		config := make(map[string]string)
		config["net_udp_port"] = fmt.Sprintf("%d", udpPort)
		trace.Init(config)
	}
	defer trace.Shutdown()

	// ============================================================
	// Global Database Connection (No Tracking)
	// ============================================================
	// This connection is opened at startup without HTTP context.
	// Use manual tracking (whatapsql.Start/End) for queries.
	serviceDB, err := sql.Open(MYSQL_DRIVER_NAME, dataSource)
	if err != nil {
		fmt.Println("Error service sql Open ", err)
		return
	}
	defer serviceDB.Close()

	http.HandleFunc("/", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request -", r)

		tp, err := template.ParseFiles("templates/database/sql/index.html")
		if err != nil {
			fmt.Println("Template not loaded, ", err)
			return
		}
		data := &HTMLData{}
		data.Title = "database/sql server"
		data.Content = r.RequestURI
		tp.Execute(w, data)

		fmt.Println("Response -", r.Response)
	}))

	// ============================================================
	// CASE 1: Query with Manual Tracking
	// ============================================================
	// Demonstrates manual SQL tracking using whatapsql.Start/End.
	// Also shows connection tracking with whatapsql.StartOpen/End.
	http.HandleFunc("/query", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		w.Header().Add("Content-Type", "text/html")

		ctx := r.Context()
		fmt.Println("Request -", r)
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		// Track database connection opening
		sqlCtx, _ := whatapsql.StartOpen(ctx, dataSource)
		db, err := sql.Open(MYSQL_DRIVER_NAME, dataSource)
		whatapsql.End(sqlCtx, err)

		if err != nil {
			fmt.Println("Error whatapsql.Open ", err)
			return
		}
		defer db.Close()

		// Track SQL query execution
		var id int
		var subject string
		var query = "select id, subject from tbl_faq limit 10"

		// Start SQL tracking
		sqlCtx, _ = whatapsql.Start(ctx, dataSource, query)
		rows, err := db.Query(query)
		// End SQL tracking with error status
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

		// Track SQL query with context
		query = "select id, subject from tbl_faq limit 10"
		sqlCtx, _ = whatapsql.Start(ctx, dataSource, query)
		rows, err = db.QueryContext(ctx, query)
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

		_, _ = w.Write(buffer.Bytes())

		fmt.Println("Response -", r.Response)

	}))

	// ============================================================
	// CASE 2: QueryRow with Manual Tracking
	// ============================================================
	// Demonstrates tracking single row queries.
	http.HandleFunc("/queryRow", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		w.Header().Add("Content-Type", "text/html")

		ctx := r.Context()
		fmt.Println("Request -", r)
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		sqlCtx, _ := whatapsql.StartOpen(ctx, dataSource)
		db, err := sql.Open(MYSQL_DRIVER_NAME, dataSource)
		whatapsql.End(sqlCtx, err)
		if err != nil {
			fmt.Println("Error whatapsql.Open ", err)
			return
		}
		defer db.Close()
		var id int
		var subject string
		var query string
		query = "select id, subject from tbl_faq limit 1"

		// Track QueryRow
		sqlCtx, _ = whatapsql.Start(ctx, dataSource, query)
		row := db.QueryRow(query)
		whatapsql.End(sqlCtx, nil)

		// Scan and close
		if err := row.Scan(&id, &subject); err != nil {
			fmt.Println("Error Row.Scan ", err)
		} else {
			fmt.Println(id, subject)
			buffer.WriteString(fmt.Sprintln(id, subject, "<br>"))
		}

		// Track QueryRowContext
		sqlCtx, _ = whatapsql.Start(ctx, dataSource, query)
		row = db.QueryRowContext(ctx, query)
		whatapsql.End(sqlCtx, nil)

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

	// ============================================================
	// CASE 3: Prepared Statement with Parameter Tracking
	// ============================================================
	// Demonstrates tracking prepared statements with parameters.
	// Use whatapsql.StartWithParam() to include parameter values in trace.
	http.HandleFunc("/prepare", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		w.Header().Add("Content-Type", "text/html")

		ctx := r.Context()
		fmt.Println("Request -", r)
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		sqlCtx, _ := whatapsql.StartOpen(ctx, dataSource)
		db, err := sql.Open(MYSQL_DRIVER_NAME, dataSource)
		whatapsql.End(sqlCtx, err)
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
			// Track with parameters using StartWithParam
			sqlCtx, _ = whatapsql.StartWithParam(ctx, dataSource, query, params...)
			if rows, err1 := stmt.Query(params...); err1 == nil {
				whatapsql.End(sqlCtx, err1)
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
				whatapsql.End(sqlCtx, err1)
				buffer.WriteString(fmt.Sprintln("Error db.Prepare stmt.Query ", err1, "<br>"))
				fmt.Println("Error stmt.Query ", err1)
			}

			sqlCtx, _ = whatapsql.StartWithParam(ctx, dataSource, query, params...)
			if rows, err1 := stmt.QueryContext(ctx, params...); err == nil {
				whatapsql.End(sqlCtx, err1)
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
				whatapsql.End(sqlCtx, err1)
				buffer.WriteString(fmt.Sprintln("Error db.Prepare stmt.QueryContext ", err1, "<br>"))
				fmt.Println("Error stmt.QueryContext ", err1)
			}
		} else {
			fmt.Println("Error db.Prepare ", err)
			buffer.WriteString(fmt.Sprintln("Error db.Prepare ", err, "<br>"))
		}

		if stmt, err := db.PrepareContext(ctx, query); err == nil {
			sqlCtx, _ = whatapsql.StartWithParam(ctx, dataSource, query, params...)
			if rows, err1 := stmt.Query(params...); err1 == nil {
				whatapsql.End(sqlCtx, err1)
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
				whatapsql.End(sqlCtx, err1)
				buffer.WriteString(fmt.Sprintln("Error db.PrepareContext stmt.QueryContex ", err1, "<br>"))
				fmt.Println("Error stmt.QueryContext ", err1)
			}

			sqlCtx, _ = whatapsql.StartWithParam(ctx, dataSource, query, params...)
			if rows, err1 := stmt.QueryContext(ctx, params...); err1 == nil {
				whatapsql.End(sqlCtx, err1)
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
				whatapsql.End(sqlCtx, err1)
				buffer.WriteString(fmt.Sprintln("Error db.PrepareContext stmt.QueryContex ", err1, "<br>"))
				fmt.Println("Error stmt.QueryContext ", err1)
			}

		} else {
			fmt.Println("Error db.PrepareContext ", err)
			buffer.WriteString(fmt.Sprintln("Error db.PrepareContext ", err, "<br>"))
		}

		query = "select id, subject from tbl_faq where id in (?,?) limit 1"

		if stmt, err := db.Prepare(query); err == nil {
			sqlCtx, _ = whatapsql.StartWithParam(ctx, dataSource, query, params...)
			row := stmt.QueryRow(params...)
			whatapsql.End(sqlCtx, nil)
			if err1 := row.Scan(&id, &subject); err1 == nil {
				fmt.Println(id, subject)
				buffer.WriteString(fmt.Sprintln(id, subject, "<br>"))
			} else {
				fmt.Println("Error row.Scan ", err1)
				buffer.WriteString(fmt.Sprintln("Error  stmt.QueryRow row.Scan ", err1, "<br>"))
			}

			sqlCtx, _ = whatapsql.StartWithParam(ctx, dataSource, query, params...)
			row = stmt.QueryRowContext(ctx, params...)
			whatapsql.End(sqlCtx, nil)
			if err1 := row.Scan(&id, &subject); err1 == nil {
				fmt.Println(id, subject)
				buffer.WriteString(fmt.Sprintln(id, subject, "<br>"))
			} else {
				fmt.Println("Error row.Scan ", err1)
				buffer.WriteString(fmt.Sprintln("Error  stmt.QueryRowContext row.Scan", err1, "<br>"))
			}
		} else {
			fmt.Println("Error db.Prepare ", err)
			buffer.WriteString(fmt.Sprintln("Error db.Prepare row.Scan ", err, "<br>"))
		}

		query = "update tbl_faq set subject='aaa' where id in (?,?) limit 1"
		if stmt, err := db.Prepare(query); err == nil {
			sqlCtx, _ = whatapsql.StartWithParam(ctx, dataSource, query, params...)
			if res, err1 := stmt.Exec(params...); err1 == nil {
				whatapsql.End(sqlCtx, err1)
				fmt.Println("Result ", res)
			} else {
				whatapsql.End(sqlCtx, err1)
				buffer.WriteString(fmt.Sprintln("Error stmt.Exec ", err1, "<br>"))
				fmt.Println("Error stmt.Exec ", err1)
			}

			sqlCtx, _ = whatapsql.StartWithParam(ctx, dataSource, query, params...)
			if res, err1 := stmt.ExecContext(ctx, params...); err1 == nil {
				whatapsql.End(sqlCtx, err1)
				fmt.Println("Result ", res)
			} else {
				whatapsql.End(sqlCtx, err1)
				buffer.WriteString(fmt.Sprintln("Error stmt.ExecContext ", err1, "<br>"))
				fmt.Println("Error stmt.ExecContext ", err1)
			}
		} else {
			fmt.Println("Error db.Prepare ", err)
			buffer.WriteString(fmt.Sprintln("Error db.Prepare Exec ", err, "<br>"))
		}
		_, _ = w.Write(buffer.Bytes())

		fmt.Println("Response -", r.Response)
	}))

	// ============================================================
	// CASE 4: Named Parameters
	// ============================================================
	// Demonstrates tracking with named parameters using sql.Named().
	// Use whatapsql.StartWithParamArray() for named parameter arrays.
	http.HandleFunc("/named", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		w.Header().Add("Content-Type", "text/html")

		ctx := r.Context()
		fmt.Println("Request -", r)
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		sqlCtx, _ := whatapsql.StartOpen(ctx, dataSource)
		db, err := sql.Open(MYSQL_DRIVER_NAME, dataSource)
		whatapsql.End(sqlCtx, err)
		if err != nil {
			fmt.Println("Error whatapsql.Open")
			http.Error(w, fmt.Sprintln("Error whatapsql.Open", err), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		query := "select id, subject from tbl_faq where id in (?, ?) limit 10"
		var id int
		var subject string

		// Named parameters using sql.Named()
		params := make([]interface{}, 0)
		params = append(params, sql.Named("idx1", 8))
		params = append(params, sql.Named("idx2", 1))

		if stmt, err := db.Prepare(query); err == nil {
			// Use StartWithParamArray for named parameters
			sqlCtx, _ = whatapsql.StartWithParamArray(ctx, dataSource, query, params)
			if rows, err1 := stmt.QueryContext(ctx, params...); err1 == nil {
				whatapsql.End(sqlCtx, err1)
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
				whatapsql.End(sqlCtx, err1)
				fmt.Println("Error db.QueryContext", err1)
				http.Error(w, fmt.Sprintln("Error db.QueryContext", err1), http.StatusInternalServerError)
			}

		} else {
			fmt.Println("Error db.Prepard ", err)
			http.Error(w, fmt.Sprintln("Error db.Prepared", err), http.StatusInternalServerError)
		}
		// 복수 Row를 갖는 SQL 쿼리
		_, _ = w.Write(buffer.Bytes())

		fmt.Println("Response -", r.Response)
	}))

	// ============================================================
	// CASE 5: Exec (INSERT, UPDATE, DELETE)
	// ============================================================
	// Demonstrates tracking DML statements.
	http.HandleFunc("/exec", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		w.Header().Add("Content-Type", "text/html")

		ctx := r.Context()
		if _, traceCtx := trace.GetTraceContext(ctx); traceCtx != nil {
			fmt.Println("Txid=", traceCtx.Txid)
		}
		fmt.Println("Request -", r)
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		sqlCtx, _ := whatapsql.StartOpen(ctx, dataSource)
		db, err := sql.Open(MYSQL_DRIVER_NAME, dataSource)
		whatapsql.End(sqlCtx, err)
		if err != nil {
			fmt.Println("Error whatapsql.Open")
			return
		}
		defer db.Close()

		params := make([]interface{}, 0)
		params = append(params, 8)
		params = append(params, 1)

		// Track UPDATE statement
		query := "update tbl_faq set subject = 'aaa' where id in (?,?)"
		sqlCtx, _ = whatapsql.StartWithParamArray(ctx, dataSource, query, params)
		if res, err := db.Exec(query, params...); err == nil {
			whatapsql.End(sqlCtx, err)
			fmt.Println("Result ", res)
			buffer.WriteString(fmt.Sprintln("Result ", res, "<br>"))
		} else {
			whatapsql.End(sqlCtx, err)
			fmt.Println("Error db.Exec ", err)
		}

		sqlCtx, _ = whatapsql.StartWithParamArray(ctx, dataSource, query, params)
		if res, err := db.ExecContext(ctx, query, params...); err == nil {
			whatapsql.End(sqlCtx, err)
			fmt.Println("Result ", res)
			buffer.WriteString(fmt.Sprintln("Result ", res, "<br>"))
		} else {
			whatapsql.End(sqlCtx, err)
			fmt.Println("Error db.ExecContext ", err)
		}

		_, _ = w.Write(buffer.Bytes())

		fmt.Println("Response -", r.Response)
	}))

	// ============================================================
	// CASE 6: Transaction with Begin/Commit/Rollback
	// ============================================================
	// Demonstrates tracking database transactions.
	http.HandleFunc("/tx", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		w.Header().Add("Content-Type", "text/html")
		ctx := r.Context()
		fmt.Println("Request -", r)
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		sqlCtx, _ := whatapsql.StartOpen(ctx, dataSource)
		db, err := sql.Open(MYSQL_DRIVER_NAME, dataSource)
		whatapsql.End(sqlCtx, err)
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

		// Track transaction begin
		sqlCtx, _ = whatapsql.Start(ctx, dataSource, "Begin Tx")
		if tx, err := db.BeginTx(ctx, nil); err == nil {
			whatapsql.End(sqlCtx, err)

			// Track queries within transaction
			query = "update tbl_faq set subject = 'bbb' where id in (?,?)"
			sqlCtx1, _ := whatapsql.StartWithParam(ctx, dataSource, query, params...)
			if res, err := tx.Exec(query, params...); err != nil {
				whatapsql.End(sqlCtx1, err)
				fmt.Println("Error tx.Exec ", err)
			} else {
				whatapsql.End(sqlCtx1, err)
				fmt.Println("tx.Exec  Result ", res)
			}

			query = "select id, subject from tbl_faq where id in (?,?)"
			sqlCtx1, _ = whatapsql.StartWithParamArray(ctx, dataSource, query, params)
			rows, err := tx.Query(query, params...)
			whatapsql.End(sqlCtx1, err)
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
			sqlCtx1, _ = whatapsql.StartWithParamArray(ctx, dataSource, query, params)
			if res, err := tx.ExecContext(ctx, query, params...); err != nil {
				whatapsql.End(sqlCtx1, err)
				fmt.Println("Error tx.ExecContext ", err)
			} else {
				whatapsql.End(sqlCtx1, err)
				fmt.Println("tx.ExecContext Result", res)
			}

			query = "select id, subject from tbl_faq where id in (?,?)"

			sqlCtx1, _ = whatapsql.StartWithParamArray(ctx, dataSource, query, params)
			rows, err = tx.QueryContext(ctx, query, params...)
			whatapsql.End(sqlCtx1, err)

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
			whatapsql.End(sqlCtx, err)
			fmt.Println("Error tx.BeginTx ", err)
		}

		_, _ = w.Write(buffer.Bytes())

		fmt.Println("Response -", r.Response)
	}))

	// ============================================================
	// CASE 7: Using Shared Service DB with Manual Tracking
	// ============================================================
	// Demonstrates tracking queries on a shared/global database connection.
	http.HandleFunc("/service/index", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		w.Header().Add("Content-Type", "text/html")

		ctx := r.Context()
		fmt.Println("Request -", r)
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		var id int
		var subject string
		query := "select id, subject from tbl_faq limit 10"

		// Track query on shared service DB
		sqlCtx, _ := whatapsql.Start(ctx, dataSource, query)
		if rows, err := serviceDB.QueryContext(ctx, query); err == nil {
			whatapsql.End(sqlCtx, err)
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
			whatapsql.End(sqlCtx, err)
			fmt.Println("Error db.QueryContext ", err)
		}

		if rows, err := serviceDB.Query(query); err == nil {
			whatapsql.End(sqlCtx, err)
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
			whatapsql.End(sqlCtx, err)
			fmt.Println("Error db.Query", err)
		}

		buffer.WriteString("DB Statas <hr/>")
		buffer.WriteString(fmt.Sprintln(serviceDB.Stats()))
		_, _ = w.Write(buffer.Bytes())

		fmt.Println("Response -", r.Response)
	}))

	// ============================================================
	// CASE 8: Using Goroutine ID for Context-less Tracking
	// ============================================================
	// Demonstrates using trace.GetGID() when context is not available.
	// Enable go.use_goroutine_id_enabled=true in whatap.conf.
	http.HandleFunc("/service/gid", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		w.Header().Add("Content-Type", "text/html")

		// ctx := r.Context()
		fmt.Println("Request -", r)
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		// Get goroutine ID for debugging
		fmt.Println("GID=", trace.GetGID())
		buffer.WriteString(fmt.Sprintf("GID=%d<br/><hr/>", trace.GetGID()))

		var id int
		var subject string
		query := "select id, subject from tbl_faq limit 10"
		if rows, err := serviceDB.Query(query); err == nil {
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
		}

		buffer.WriteString("DB Statas <hr/>")
		buffer.WriteString(fmt.Sprintln(serviceDB.Stats()))
		_, _ = w.Write(buffer.Bytes())

		fmt.Println("Response -", r.Response)
	}))

	// ============================================================
	// CASE 9: No HTTP Transaction Context
	// ============================================================
	// Handler without whataphttp.Func wrapper - no HTTP transaction.
	// SQL tracking still works but won't be linked to HTTP transaction.
	http.HandleFunc("/notx/select", func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		w.Header().Add("Content-Type", "text/html")

		ctx := r.Context()
		fmt.Println("Request -", r)
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		var id int
		var subject string
		query := "select id, subject from tbl_faq limit 10"
		sqlCtx, _ := whatapsql.Start(ctx, dataSource, query)
		if rows, err := serviceDB.QueryContext(ctx, query); err == nil {
			whatapsql.End(sqlCtx, err)
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
			whatapsql.End(sqlCtx, err)
			fmt.Println("Error db.QueryContext ", err)
			return
		}

		buffer.WriteString("DB Statas <hr/>")
		buffer.WriteString(fmt.Sprintln(serviceDB.Stats()))
		_, _ = w.Write(buffer.Bytes())

		fmt.Println("Response -", r.Response)
	})

	// ============================================================
	// CASE 10: Custom Error Tracking
	// ============================================================
	// Demonstrates passing custom errors to whatapsql.End().
	http.HandleFunc("/notx/error", func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		w.Header().Add("Content-Type", "text/html")

		ctx := r.Context()
		fmt.Println("Request -", r)
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		var id int
		var subject string
		query := "select id, subject from tbl_faq limit 10"
		sqlCtx, _ := whatapsql.Start(ctx, dataSource, query)
		if rows, err := serviceDB.QueryContext(ctx, query); err == nil {
			// Pass custom error to End() for tracking
			//whatapsql.End(sqlCtx, err)
			whatapsql.End(sqlCtx, fmt.Errorf("custom error"))
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
			whatapsql.End(sqlCtx, err)
			fmt.Println("Error db.QueryContext ", err)
			return
		}

		buffer.WriteString("DB Statas <hr/>")
		buffer.WriteString(fmt.Sprintln(serviceDB.Stats()))
		_, _ = w.Write(buffer.Bytes())

		fmt.Println("Response -", r.Response)
	})

	fmt.Println("Start :", port, ", Agent Udp Port:", udpPort)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		fmt.Println("Error ListenAndServe ", err)
	}
}
