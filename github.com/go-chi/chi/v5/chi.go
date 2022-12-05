package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/go-sql-driver/mysql"
	"github.com/whatap/go-api/httpc"
	wisql "github.com/whatap/go-api/instrumentation/database/sql/whatapsql"
	"github.com/whatap/go-api/instrumentation/github.com/go-chi/chi/whatapchi"
	"github.com/whatap/go-api/method"
	"github.com/whatap/go-api/trace"
)

type HTMLData struct {
	Title   string
	Content string
	//HTMLContent template.HTML
}

func getUser(ctx context.Context) {
	methodCtx, _ := method.Start(ctx, "getUser")
	defer method.End(methodCtx, nil)
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

func main() {
	portPtr := flag.Int("p", 8080, "web port. default 8080  ")
	udpPortPtr := flag.Int("up", 6600, "agent port(udp). defalt 6600 ")
	dataSourcePtr := flag.String("ds", "doremimaker:doremimaker@tcp(phpdemo2:3306)/doremimaker", " dataSourceName ")
	flag.Parse()
	port := *portPtr
	udpPort := *udpPortPtr
	dataSource := *dataSourcePtr

	config := make(map[string]string)
	config["net_udp_port"] = fmt.Sprintf("%d", udpPort)
	trace.Init(config)
	defer trace.Shutdown()
	db, err := wisql.OpenContext(context.Background(), "mysql", dataSource)
	if err != nil {
		fmt.Println("Error service whatapsql.Open ", err)
		return
	}
	defer db.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(whatapchi.Middleware)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		tp, err := template.ParseFiles("templates/github.com/go-chi/v5/index.html")
		if err != nil {
			fmt.Println("Template not loaded, ", err)
			return
		}
		data := &HTMLData{}
		data.Title = "chi/v5 server"
		data.Content = r.RequestURI

		tp.Execute(w, data)

		ctx := r.Context()
		trace.Step(ctx, "Text Message", "/", 0, 0)

		getUser(ctx)

		fmt.Println("Response -", r.Response)

	})

	r.Get("/index", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request Index -", r)
		ctx := r.Context()
		trace.Step(ctx, "Text Message", "/index", 0, 0)

	})

	r.Get("/main", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		w.Header().Add("Content-Type", "text/html")

		fmt.Println("Request -", r)
		reply := r.RequestURI + "<br/><hr/>"
		_, _ = w.Write(([]byte)(reply))
		trace.Step(ctx, "Text Message 2", "Message2", 6, 6)
		fmt.Println("Response -", r.Response)
	})

	r.Get("/httpc", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		fmt.Println("Request httpc - ", r)
		callUrl := fmt.Sprintf("http://localhost:%d/index", port)
		httpcCtx, _ := httpc.Start(ctx, callUrl)
		var buffer bytes.Buffer
		if statusCode, data, err := httpWithRequest("GET", callUrl, "", httpc.GetMTrace(httpcCtx)); err == nil {
			httpc.End(httpcCtx, statusCode, "", nil)
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", statusCode, ", data=", data))
		} else {
			httpc.End(httpcCtx, -1, "", err)
			buffer.WriteString(fmt.Sprintln("httpc Error callUrl=", callUrl, ", err=", err))
		}

		trace.Step(ctx, "Text Message", "/httpc", 6, 6)
		w.Write(buffer.Bytes())

	})

	r.Get("/httpc/unknown", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		fmt.Println("Request httpc unknown - ", r)

		callUrl := fmt.Sprintf("http://localhost:%d/unknown", port)
		httpcCtx, _ := httpc.Start(ctx, callUrl)
		var buffer bytes.Buffer
		if statusCode, data, err := httpWithRequest("GET", callUrl, "", httpc.GetMTrace(httpcCtx)); err == nil {
			httpc.End(httpcCtx, statusCode, "", nil)
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", statusCode, ", data=", data))
		} else {
			httpc.End(httpcCtx, -1, "", err)
			buffer.WriteString(fmt.Sprintln("httpc Error callUrl=", callUrl, ", err=", err))
		}

		trace.Step(ctx, "Text Message", "/httpc-unknown", 6, 6)
		w.Write(buffer.Bytes())

	})

	r.Get("/sql/select", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var buffer bytes.Buffer
		var query string

		// 복수 Row를 갖는 SQL 쿼리
		var id int
		var subject string
		query = "select id, subject from tbl_faq limit 10"
		rows, err := db.QueryContext(ctx, query)
		defer rows.Close() //반드시 닫는다 (지연하여 닫기)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))

			fmt.Println(rows, err)

			return
		}

		for rows.Next() {
			err := rows.Scan(&id, &subject)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))

				return
			}
			buffer.WriteString(fmt.Sprintln(id, subject))
		}
		// Prepared Statement 생성
		query = "select id, subject from tbl_faq where id = ? limit ?"
		stmt, err := db.PrepareContext(ctx, query)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer stmt.Close()

		// Prepared Statement 실행
		params := make([]interface{}, 0)
		params = append(params, 8)
		params = append(params, 1)

		rows1, err1 := stmt.QueryContext(ctx, params...) //Placeholder 파라미터 순서대로 전달
		if err1 != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer rows1.Close()

		for rows1.Next() {
			err := rows1.Scan(&id, &subject)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			buffer.WriteString(fmt.Sprintln(id, subject))
		}

		rows2, err2 := stmt.QueryContext(ctx, 8, 1) //Placeholder 파라미터 순서대로 전달
		if err2 != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		defer rows2.Close()
		for rows1.Next() {
			err := rows2.Scan(&id, &subject)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			buffer.WriteString(fmt.Sprintln(id, subject))
		}

		w.WriteHeader(http.StatusOK)
		w.Write(buffer.Bytes())

	})

	r.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic(fmt.Errorf("custom panic"))
	})

	r.Get("/input", func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		buffer.WriteString("<html><body>")
		form := `<form action="/saveUrlencoded" method="post" >
    Name : <input type="text" name="name" value="">
    Value : <input type="text" name="value" value="">
    <input type="submit" value="Action" />
</form></body></html>`
		buffer.WriteString(r.RequestURI + "<br/><hr/>")
		buffer.WriteString(form)
		w.Write(buffer.Bytes())
	})
	r.Post("/saveUrlencoded", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request, ", r)
		if r != nil {
			name := r.FormValue("name")
			val := r.FormValue("value")
			fmt.Println("c.Request() FormValue ", name, ", ", val)

			params := r.Form
			for k, v := range params {
				fmt.Println("c.Request().Form k=", k, ", v=", v, ",")
			}

			name1 := r.PostFormValue("name")
			val1 := r.PostFormValue("value")
			fmt.Println("c.Request() PostFormValue ", name1, ", ", val1)

			params1 := r.PostForm
			for k, v := range params1 {
				fmt.Println("c.Request().PostForm k=", k, ", v=", v, ",")
			}
		}

		name := r.FormValue("name")
		val := r.FormValue("value")
		fmt.Println("params 11 saveUrlencoded ParseForm ", name, ", ", val)

		params := r.PostForm
		for k, v := range params {
			fmt.Println("c.FormParams() k=", k, ", v=", v, ",")
		}

		if err := r.ParseForm(); err != nil {
			fmt.Println("saveUrlencoded ParseForm error ", err)
		}

		var buffer bytes.Buffer
		buffer.WriteString("<html><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")
		buffer.WriteString("</body></html>")
		_, _ = w.Write(buffer.Bytes())

	})
	r.Get("/inputFile", func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer
		buffer.WriteString("<html><body>")
		form := `<form action="/upload" method="post" enctype="multipart/form-data">
    Name : <input type="text" name="name" value="">
    Value : <input type="text" name="email" value="">
    File : <input type="file" name="file" value="">
    <input type="submit" value="Action" />
</form></body></html>`
		buffer.WriteString(r.RequestURI + "<br/><hr/>")
		buffer.WriteString(form)
		_, _ = w.Write(buffer.Bytes())
	})
	r.Post("/upload", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request, ", r)
		var buffer bytes.Buffer

		// Read form fields
		name := r.FormValue("name")
		email := r.FormValue("email")
		fmt.Println("fields name=", name, ",email=", email)

		buffer.WriteString("name=" + name + "<br/>")
		buffer.WriteString("email=" + email + "<br/>")

		//-----------
		// Read file
		//-----------

		// Source
		file, handler, err := r.FormFile("file")
		if err != nil {
			fmt.Println("r.FromFile err=", err)
			return
		}

		f, err := os.OpenFile(handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println("OpenFile err", err)
			return
		}
		defer f.Close()
		io.Copy(f, file)

		fmt.Println("upload ok ", handler.Filename, ", size=", strconv.FormatInt(handler.Size, 10))
		buffer.WriteString("upload ok " + handler.Filename + ", size=" + strconv.FormatInt(handler.Size, 10))

		_, _ = w.Write(buffer.Bytes())

	})

	fmt.Println("Start :", port, ", Agent Udp Port:", udpPort)
	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
