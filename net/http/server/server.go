package main

import (
	"bytes"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/whatap/go-api/httpc"
	"github.com/whatap/go-api/method"
	whatapsql "github.com/whatap/go-api/sql"
	"github.com/whatap/go-api/trace"
)

func getUser(ctx context.Context) {
	wCtx, _ := method.Start(ctx, "getUser")
	defer method.End(wCtx, nil)
	time.Sleep(time.Duration(1) * time.Second)
}

func httpGet(callUrl string) (int, string, error) {
	fmt.Println("httpGet ", callUrl)
	// GET 호출
	if resp, err := http.Get(callUrl); err == nil {
		defer resp.Body.Close()
		fmt.Println("status=", resp.StatusCode)

		// 결과 출력
		if data, err := ioutil.ReadAll(resp.Body); err == nil {
			return resp.StatusCode, string(data), err
		} else {
			return resp.StatusCode, "", err
		}

	} else {
		fmt.Println(err)
		return -1, "", err
	}
}

func httpWithRequest(method string, callUrl string, body string, headers http.Header) (int, string, error) {
	fmt.Println("httpGetWithRequest ", method, ", ", callUrl, ", ", body, ", ", headers)
	timeout := time.Duration(10 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	if req, err := http.NewRequest(strings.ToUpper(method), callUrl, bytes.NewBufferString(body)); err == nil {
		if headers != nil {
			for key, _ := range headers {
				req.Header.Add(key, headers.Get(key))
			}
		}
		if resp, err := client.Do(req); err == nil {
			defer resp.Body.Close()
			if data, err := ioutil.ReadAll(resp.Body); err == nil {
				fmt.Println("status=", resp.StatusCode)
				return resp.StatusCode, string(data), err
			} else {
				fmt.Println("Read response Error ", err)
				return resp.StatusCode, "", err
			}
		} else {
			fmt.Println("client.Do Error ", err)
			return -2, "", err
		}

	} else {
		fmt.Println("NewRequest Error ", err)
		return -1, "", err
	}
}

func getMysql(ctx context.Context) ([]string, error) {

	db, err := sql.Open("mysql", "doremimaker:doremimaker@tcp(192.168.56.101:3306)/doremimaker")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// 복수 Row를 갖는 SQL 쿼리
	var id int
	var subject string
	rows, err := db.QueryContext(ctx, "select id, subject from tbl_faq limit 10")
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
	portPtr := flag.Int("p", 8080, "part ")
	flag.Parse()
	port := *portPtr

	trace.Init(make(map[string]string))

	fmt.Println("Start")

	http.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		wCtx, _ := trace.StartWithRequest(r)
		defer trace.End(wCtx, nil)
		// if wErr != nil {
		// 	fmt.Println("Trace start error ", wErr)
		// }
		// defer func() {
		// 	if err := trace.End(wCtx, nil); err != nil {
		// 		fmt.Println("Error ", err)
		// 	}
		// }()
		fmt.Println("Request -", r)

		w.Header().Add("Content-Type", "text/html")

		reply := "/index <br/>Test Body"

		_, _ = w.Write(([]byte)(reply))
		trace.Step(wCtx, "Text Message", "Message", 3, 3)

		getUser(wCtx)
		fmt.Println("Response -", r.Response)

	})

	http.HandleFunc("/main", func(w http.ResponseWriter, r *http.Request) {
		wCtx, _ := trace.StartWithRequest(r)
		defer trace.End(wCtx, nil)
		w.Header().Add("Content-Type", "text/html")

		fmt.Println("Request -", r)
		reply := "/main <br/>Test Body"
		_, _ = w.Write(([]byte)(reply))
		trace.Step(wCtx, "Text Message 2", "Message2", 6, 6)
		fmt.Println("Response -", r.Response)
	})

	http.HandleFunc("/httpc", func(w http.ResponseWriter, r *http.Request) {
		wCtx, _ := trace.StartWithRequest(r)
		defer trace.End(wCtx, nil)
		w.Header().Add("Content-Type", "text/html")
		fmt.Println("Request -", r)
		callUrl := "http://localhost:8081/index"
		mCtx, _ := httpc.Start(wCtx, callUrl)
		var buffer bytes.Buffer
		if statusCode, data, err := httpWithRequest("GET", callUrl, "", httpc.MTrace(mCtx)); err == nil {
			httpc.End(mCtx, statusCode, "", nil)
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", statusCode, ", data=", data))
		} else {
			httpc.End(mCtx, -1, "", err)
			buffer.WriteString(fmt.Sprintln("httpc Error callUrl=", callUrl, ", err=", err))
		}

		_, _ = w.Write(buffer.Bytes())
		trace.Step(wCtx, "Text Message 2", "Message2", 6, 6)
		fmt.Println("Response -", r.Response)
	})

	http.HandleFunc("/wrapHandleFunc", trace.Func(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer
		buffer.WriteString("wrapHandleFunc")
		_, _ = w.Write(buffer.Bytes())
		trace.Step(r.Context(), "Text Message wrapHandleFunc", "wrapHandleFunc", 6, 6)
	}))

	http.Handle("/wrapHandleFunc1", trace.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer
		buffer.WriteString("wrapHandleFunc1")
		_, _ = w.Write(buffer.Bytes())
		trace.Step(r.Context(), "Text Message wrapHandleFunc1", "wrapHandleFunc1", 6, 6)
	}))

	http.Handle("/sql/select", trace.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer
		var query string

		wCtx, _ := whatapsql.StartOpen(ctx, "doremimaker:doremimaker@tcp(192.168.56.101:3306)/doremimaker")
		db, err := sql.Open("mysql", "doremimaker:doremimaker@tcp(192.168.56.101:3306)/doremimaker")
		whatapsql.End(wCtx, err)
		if err != nil {
			return
		}
		defer db.Close()

		// 복수 Row를 갖는 SQL 쿼리
		var id int
		var subject string
		query = "select id, subject from tbl_faq limit 10"
		wCtx, _ = whatapsql.Start(ctx, "doremimaker:doremimaker@tcp(192.168.56.101:3306)/doremimaker", query)
		rows, err := db.QueryContext(ctx, query)
		whatapsql.End(wCtx, err)
		if err != nil {
			return
		}
		defer rows.Close() //반드시 닫는다 (지연하여 닫기)

		for rows.Next() {
			err := rows.Scan(&id, &subject)
			if err != nil {
				return
			}
			buffer.WriteString(fmt.Sprintln(id, subject))
		}

		// Prepared Statement 생성
		query = "select id, subject from tbl_faq where id = ? limit ?"
		stmt, err := db.Prepare(query)
		if err != nil {
			return
		}
		defer stmt.Close()

		// Prepared Statement 실행
		params := make([]interface{}, 0)
		params = append(params, 8)
		params = append(params, 1)

		wCtx, _ = whatapsql.StartWithParamArray(ctx, "doremimaker:doremimaker@tcp(192.168.56.101:3306)/doremimaker", query, params)
		rows1, err1 := stmt.QueryContext(ctx, params...) //Placeholder 파라미터 순서대로 전달
		whatapsql.End(wCtx, err1)
		defer rows1.Close()

		for rows1.Next() {
			err := rows1.Scan(&id, &subject)
			if err != nil {
				return
			}
			buffer.WriteString(fmt.Sprintln(id, subject))
		}

		wCtx, _ = whatapsql.StartWithParam(ctx, "doremimaker:doremimaker@tcp(192.168.56.101:3306)/doremimaker", query, fmt.Sprintln(8, ",", 1))
		rows2, err2 := stmt.QueryContext(ctx, 8, 1) //Placeholder 파라미터 순서대로 전달
		whatapsql.End(wCtx, err2)
		defer rows2.Close()

		for rows1.Next() {
			err := rows1.Scan(&id, &subject)
			if err != nil {
				return
			}
			buffer.WriteString(fmt.Sprintln(id, subject))
		}

		_, _ = w.Write(buffer.Bytes())

	}))

	fmt.Println("Start :", port)

	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
