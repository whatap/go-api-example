package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"strconv"
	"strings"
	"time"

	"github.com/fasthttp/router"
	_ "github.com/go-sql-driver/mysql"
	"github.com/valyala/fasthttp"

	"github.com/whatap/go-api/httpc"
	"github.com/whatap/go-api/instrumentation/github.com/valyala/fasthttp/whatapfasthttp"
	"github.com/whatap/go-api/method"

	wisql "github.com/whatap/go-api/instrumentation/database/sql/whatapsql"
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
	setWhatapPtr := flag.Bool("whatap", false, "set whatap")

	flag.Parse()
	port := *portPtr
	udpPort := *udpPortPtr
	dataSource := *dataSourcePtr
	IsWhatap := *setWhatapPtr

	if IsWhatap {
		config := make(map[string]string)
		config["net_udp_port"] = fmt.Sprintf("%d", udpPort)
		trace.Init(config)
	}
	defer trace.Shutdown()

	db, err := wisql.OpenContext(context.Background(), "mysql", dataSource)
	if err != nil {
		fmt.Println("Error service whatapsql.Open ", err)
		return
	}
	defer db.Close()

	r := router.New()

	r.GET("/", whatapfasthttp.Func(func(ctx *fasthttp.RequestCtx) {
		tp, err := template.ParseFiles("templates/github.com/valyala/index.html")
		if err != nil {
			fmt.Println("Template not loaded, ", err)
			return
		}
		data := &HTMLData{}
		data.Title = "fasthttp server"
		data.Content = string(ctx.Request.RequestURI())

		tp.Execute(ctx, data)

		ctx.WriteString("Welcome!")
		ctx.SetContentType("text/html;charset=utf8")
	}))
	r.GET("/hello/{name}", whatapfasthttp.Func(func(ctx *fasthttp.RequestCtx) {
		fmt.Fprintf(ctx, "Hello, %s!\n", ctx.UserValue("name"))
		ctx.SetContentType("text/html;charset=utf8")
	}))

	r.GET("/index", whatapfasthttp.Func(func(ctx *fasthttp.RequestCtx) {
		// fmt.Println("Request -", ctx.Request)

		trace.Step(ctx, "Text Message", "Message", 3, 3)

		getUser(ctx)
		ctx.WriteString(fmt.Sprintln("message", "/index <br/>Test Body"))
		ctx.SetContentType("text/html;charset=utf8")
	}))

	r.GET("/main", whatapfasthttp.Func(func(ctx *fasthttp.RequestCtx) {
		// fmt.Println("Request -", ctx.Request)

		trace.Step(ctx, "Text Message 2", "Message2", 6, 6)

		ctx.WriteString(fmt.Sprintln("message", "/main <br/>Test Body"))
		ctx.SetContentType("text/html;charset=utf8")
	}))

	r.GET("/httpc", whatapfasthttp.Func(func(ctx *fasthttp.RequestCtx) {
		// fmt.Println("Request -", c.Request)

		callUrl := "http://localhost:8081/index"
		httpcCtx, _ := httpc.Start(ctx, callUrl)
		var buffer bytes.Buffer
		if statusCode, data, err := httpWithRequest("GET", callUrl, "", trace.GetMTrace(ctx)); err == nil {
			httpc.End(httpcCtx, statusCode, "", nil)
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", statusCode, ", data=", data))
		} else {
			httpc.End(httpcCtx, -1, "", err)
			buffer.WriteString(fmt.Sprintln("httpc Error callUrl=", callUrl, ", err=", err))
		}

		trace.Step(ctx, "Text Message 2", "Message2", 6, 6)

		ctx.WriteString(string(buffer.Bytes()))
		ctx.SetContentType("text/html;charset=utf8")
	}))

	r.GET("/httpc/unknown", whatapfasthttp.Func(func(ctx *fasthttp.RequestCtx) {
		// fmt.Println("Request -", c.Request)

		callUrl := "http://localhost:8081/unknown"
		httpcCtx, _ := httpc.Start(ctx, callUrl)
		var buffer bytes.Buffer
		if statusCode, data, err := httpWithRequest("GET", callUrl, "", trace.GetMTrace(ctx)); err == nil {
			httpc.End(httpcCtx, statusCode, "", nil)
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", statusCode, ", data=", data))
		} else {
			httpc.End(httpcCtx, -1, "", err)
			buffer.WriteString(fmt.Sprintln("httpc Error callUrl=", callUrl, ", err=", err))
		}

		ctx.WriteString(string(buffer.Bytes()))
		ctx.SetContentType("text/html;charset=utf8")
	}))

	r.GET("/sql/select", whatapfasthttp.Func(func(ctx *fasthttp.RequestCtx) {
		var buffer bytes.Buffer
		var query string

		// 복수 Row를 갖는 SQL 쿼리
		var id int
		var subject string
		query = "select id, subject from tbl_faq limit 10"
		rows, err := db.QueryContext(ctx, query)
		if err == nil {
			defer rows.Close() //반드시 닫는다 (지연하여 닫기)

			for rows.Next() {
				err := rows.Scan(&id, &subject)
				if err != nil {
					ctx.Error("message"+err.Error(), http.StatusInternalServerError)
					return
				}
				buffer.WriteString(fmt.Sprintln(id, subject))
			}
		}
		// Prepared Statement 생성
		query = "select id, subject from tbl_faq where id = ? limit ?"
		stmt, err := db.PrepareContext(ctx, query)
		if err != nil {
			ctx.Error("message"+err.Error(), http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		// Prepared Statement 실행
		params := make([]interface{}, 0)
		params = append(params, 8)
		params = append(params, 1)

		rows1, err1 := stmt.QueryContext(ctx, params...) //Placeholder 파라미터 순서대로 전달
		if err1 != nil {
			ctx.Error("message"+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows1.Close()

		for rows1.Next() {
			err := rows1.Scan(&id, &subject)
			if err != nil {
				ctx.Error("message"+err.Error(), http.StatusInternalServerError)
				return
			}
			buffer.WriteString(fmt.Sprintln(id, subject))
		}

		rows2, err2 := stmt.QueryContext(ctx, 8, 1) //Placeholder 파라미터 순서대로 전달
		if err2 != nil {
			ctx.Error("message"+err2.Error(), http.StatusInternalServerError)
			return
		}
		defer rows2.Close()

		for rows1.Next() {
			err := rows2.Scan(&id, &subject)
			if err != nil {
				ctx.Error("Error "+err.Error(), http.StatusInternalServerError)
				return
			}
			buffer.WriteString(fmt.Sprintln(id, subject))
		}

		ctx.WriteString(string(buffer.Bytes()))
		ctx.SetContentType("text/html;charset=utf8")
	}))

	r.GET("/panic", whatapfasthttp.Func(func(ctx *fasthttp.RequestCtx) {
		panic(fmt.Errorf("custom panic"))
		ctx.WriteString(string(ctx.RequestURI()) + "<br/><hr/>")
		ctx.SetContentType("text/html;charset=utf8")
	}))

	r.GET("/input", whatapfasthttp.Func(func(ctx *fasthttp.RequestCtx) {
		var buffer bytes.Buffer
		buffer.WriteString("<html><body>")
		form := `<form action="/saveUrlencoded" method="post" >
	   Name : <input type="text" name="name" value="">
	   Value : <input type="text" name="value" value="">
	   <input type="submit" value="Action" />
	</form></body></html>`
		buffer.WriteString(string(ctx.RequestURI()) + "<br/><hr/>")
		buffer.WriteString(form)

		ctx.WriteString(string(buffer.Bytes()))
		ctx.SetContentType("text/html;charset=utf8")
	}))

	r.POST("/saveUrlencoded", whatapfasthttp.Func(func(ctx *fasthttp.RequestCtx) {
		// fmt.Println("Request, ", c.Request)

		if ctx != nil {
			name := ctx.FormValue("name")
			val := ctx.FormValue("value")
			fmt.Println("c.Request() FormValue ", name, ", ", val)

			query_args := ctx.QueryArgs()
			form_args := ctx.PostArgs()
			visit_func_get := func(key, value []byte) {
				fmt.Println("Get key=", key, ",v=", string(value))
			}
			if query_args != nil {
				query_args.VisitAll(visit_func_get)
			}

			visit_func_post := func(key, value []byte) {
				fmt.Println("Post key=", key, ",v=", string(value))
			}

			if form_args != nil {
				form_args.VisitAll(visit_func_post)
			}

		}

		var buffer bytes.Buffer
		buffer.WriteString(string(ctx.RequestURI()) + "<br/><hr/>")
		ctx.WriteString(string(buffer.Bytes()))
		ctx.SetContentType("text/html;charset=utf8")

	}))

	r.GET("/inputFile", whatapfasthttp.Func(func(ctx *fasthttp.RequestCtx) {
		var buffer bytes.Buffer
		buffer.WriteString("<html><body>")
		form := `<form action="/upload" method="post" enctype="multipart/form-data">
	   Name : <input type="text" name="name" value="">
	   Value : <input type="text" name="email" value="">
	   File : <input type="file" name="file" value="">
	   <input type="submit" value="Action" />
	</form></body></html>`
		buffer.WriteString(string(ctx.RequestURI()) + "<br/><hr/>")
		buffer.WriteString(form)

		ctx.WriteString(string(buffer.Bytes()))
		ctx.SetContentType("text/html;charset=utf8")

	}))

	r.POST("/upload", whatapfasthttp.Func(func(ctx *fasthttp.RequestCtx) {
		// fmt.Println("Request, ", c.Request)
		var buffer bytes.Buffer

		// Read form fields
		name := string(ctx.FormValue("name"))
		email := string(ctx.FormValue("email"))
		fmt.Println("fields name=", name, ",email=", email)

		buffer.WriteString("name=" + name + "<br/>")
		buffer.WriteString("email=" + email + "<br/>")

		//-----------
		// Read file
		//-----------
		file, _ := ctx.FormFile("file")
		fmt.Println(file.Filename + " uploaded")

		// 파일 저장

		// 방법 1.
		// 기본 제공 함수로 파일 저장
		if err := fasthttp.SaveMultipartFile(file, file.Filename); err != nil {
			panic(err)
		}

		fmt.Println("upload ok ", file.Filename, ", size=", strconv.FormatInt(file.Size, 10))
		buffer.WriteString("upload ok " + file.Filename + ", size=" + strconv.FormatInt(file.Size, 10))

		ctx.WriteString(string(buffer.Bytes()))
		ctx.SetContentType("text/html;charset=utf8")

	}))

	s := &fasthttp.Server{
		Handler: r.Handler,

		// Every response will contain 'Server: My super server' header.
		Name: "My super server",

		// Other Server settings may be set here.
		ErrorHandler: ErrorHandler,
	}

	// Start the server listening for incoming requests on the given address.
	//
	// ListenAndServe returns only on error, so usually it blocks forever.
	fmt.Println("Start :", port, ", Agent Udp Port:", udpPort)
	if err := s.ListenAndServe(fmt.Sprintf(":%d", port)); err != nil {
		// fmt.Fatalf("error in ListenAndServe: %v", err)
	}
	// fasthttp.ListenAndServe(fmt.Sprintf(":%d", port), r.Handler)
}

func ErrorHandler(ctx *fasthttp.RequestCtx, err error) {
	fmt.Println("ErrorHandler", ctx.RequestURI(), ", error=", err)
}
