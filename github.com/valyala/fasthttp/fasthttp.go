// FastHTTP (github.com/valyala/fasthttp) Instrumentation Example
//
// This example demonstrates how to instrument FastHTTP web framework applications
// using WhaTap Go API.
//
// INSTRUMENTATION OVERVIEW:
// 1. Call trace.Init() at application startup
// 2. Wrap each handler with whatapfasthttp.Func()
// 3. Use ctx (*fasthttp.RequestCtx) as the context directly
//
// FEATURES:
// - Automatic HTTP transaction tracking
// - Request/Response details (URL, method, status code, response time)
// - SQL query tracking via whatapsql
// - HTTP client call tracking via httpc
// - Custom method tracing via method.Start/End
// - Error and panic tracking
// - Distributed tracing via header propagation
//
// IMPORTANT: FastHTTP does NOT use middleware pattern.
// Instead, wrap each handler function with whatapfasthttp.Func().
//
// CONTEXT ACCESS:
// FastHTTP's RequestCtx implements context.Context interface directly.
// Use ctx (*fasthttp.RequestCtx) as the context parameter for trace functions.

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

// ============================================================
// Example: Custom Method Tracing
// ============================================================
// Use method.Start/End to trace custom functions.
// FastHTTP's RequestCtx implements context.Context, so it can be used directly.
func getUser(ctx context.Context) {
	methodCtx, _ := method.Start(ctx, "getUser")
	defer method.End(methodCtx, nil)
	time.Sleep(time.Duration(1) * time.Second)
}

// Helper function for HTTP GET requests
func httpGet(callUrl string) (int, string, error) {
	fmt.Println("httpGet ", callUrl)
	// GET call
	if resp, err := http.Get(callUrl); err == nil {
		defer resp.Body.Close()
		fmt.Println("status=", resp.StatusCode)

		// print result
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

// Helper function for HTTP requests with custom headers
// Used with httpc.Start/End for tracking
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
	// Database Connection with Instrumentation
	// ============================================================
	// Use whatapsql.OpenContext() for automatic SQL query tracking.
	db, err := wisql.OpenContext(context.Background(), "mysql", dataSource)
	if err != nil {
		fmt.Println("Error service whatapsql.Open ", err)
		return
	}
	defer db.Close()

	r := router.New()

	// ============================================================
	// INSTRUMENTATION STEP 2: Wrap Handlers with whatapfasthttp.Func
	// ============================================================
	// FastHTTP does not use middleware pattern like other frameworks.
	// Instead, wrap each handler function with whatapfasthttp.Func().
	//
	// BEFORE (Original):
	//   r.GET("/", func(ctx *fasthttp.RequestCtx) {
	//       ctx.WriteString("Hello")
	//   })
	//
	// AFTER (Instrumented):
	//   r.GET("/", whatapfasthttp.Func(func(ctx *fasthttp.RequestCtx) {
	//       ctx.WriteString("Hello")
	//   }))
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

	// Example: URL path parameters
	r.GET("/hello/{name}", whatapfasthttp.Func(func(ctx *fasthttp.RequestCtx) {
		fmt.Fprintf(ctx, "Hello, %s!\n", ctx.UserValue("name"))
		ctx.SetContentType("text/html;charset=utf8")
	}))

	// ============================================================
	// INSTRUMENTATION STEP 3: Use ctx as Context Directly
	// ============================================================
	// FastHTTP's RequestCtx implements context.Context interface.
	// Use ctx directly as the context parameter for trace functions.
	// No need to call any method like c.Context() or r.Context().
	r.GET("/index", whatapfasthttp.Func(func(ctx *fasthttp.RequestCtx) {
		// fmt.Println("Request -", ctx.Request)

		// Use ctx directly as context - it implements context.Context
		trace.Step(ctx, "Text Message", "Message", 3, 3)

		// Pass ctx to custom method - works because ctx implements context.Context
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

	// ============================================================
	// HTTP Client Call with Distributed Tracing
	// ============================================================
	// Use httpc.Start/End for outgoing HTTP call tracking.
	// Use trace.GetMTrace(ctx) to propagate trace context via headers.
	r.GET("/httpc", whatapfasthttp.Func(func(ctx *fasthttp.RequestCtx) {
		// fmt.Println("Request -", c.Request)

		callUrl := "http://localhost:8081/index"

		// Start HTTP client call tracking - use ctx directly
		httpcCtx, _ := httpc.Start(ctx, callUrl)
		var buffer bytes.Buffer

		// Make HTTP call with trace headers for distributed tracing
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

	// ============================================================
	// SQL Query Tracking Example
	// ============================================================
	// SQL queries are automatically tracked when using whatapsql.
	// Pass ctx (RequestCtx) to QueryContext/ExecContext for transaction linking.
	r.GET("/sql/select", whatapfasthttp.Func(func(ctx *fasthttp.RequestCtx) {
		var buffer bytes.Buffer
		var query string

		// SQL query returning multiple rows
		var id int
		var subject string
		query = "select id, subject from tbl_faq limit 10"

		// Pass ctx directly to QueryContext - works because ctx implements context.Context
		rows, err := db.QueryContext(ctx, query)
		if err == nil {
			defer rows.Close() // must close (deferred)

			for rows.Next() {
				err := rows.Scan(&id, &subject)
				if err != nil {
					ctx.Error("message"+err.Error(), http.StatusInternalServerError)
					return
				}
				buffer.WriteString(fmt.Sprintln(id, subject))
			}
		}
		// create prepared statement
		query = "select id, subject from tbl_faq where id = ? limit ?"
		stmt, err := db.PrepareContext(ctx, query)
		if err != nil {
			ctx.Error("message"+err.Error(), http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		// execute prepared statement
		params := make([]interface{}, 0)
		params = append(params, 8)
		params = append(params, 1)

		rows1, err1 := stmt.QueryContext(ctx, params...) // pass placeholder parameters in order
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

		rows2, err2 := stmt.QueryContext(ctx, 8, 1) // pass placeholder parameters in order
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

	// ============================================================
	// Panic Tracking Example
	// ============================================================
	// Panics are captured by the whatapfasthttp.Func wrapper.
	r.GET("/panic", whatapfasthttp.Func(func(ctx *fasthttp.RequestCtx) {
		panic(fmt.Errorf("custom panic"))
		ctx.WriteString(string(ctx.RequestURI()) + "<br/><hr/>")
		ctx.SetContentType("text/html;charset=utf8")
	}))

	// Form handling examples (standard patterns)
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

		// save file

		// method 1.
		// save file using the built-in helper
		if err := fasthttp.SaveMultipartFile(file, file.Filename); err != nil {
			panic(err)
		}

		fmt.Println("upload ok ", file.Filename, ", size=", strconv.FormatInt(file.Size, 10))
		buffer.WriteString("upload ok " + file.Filename + ", size=" + strconv.FormatInt(file.Size, 10))

		ctx.WriteString(string(buffer.Bytes()))
		ctx.SetContentType("text/html;charset=utf8")

	}))

	// ============================================================
	// WrapHandler: Wrap an existing handler function
	// ============================================================
	// Use whatapfasthttp.WrapHandler() to wrap a handler that is already defined.
	// This is useful when assigning handlers to fasthttp.Server.Handler or
	// when you have a pre-existing handler function.
	//
	// BEFORE (Original):
	//   s := &fasthttp.Server{Handler: myHandler}
	//
	// AFTER (Instrumented):
	//   s := &fasthttp.Server{Handler: whatapfasthttp.WrapHandler(myHandler)}
	r.GET("/wrapHandler", whatapfasthttp.WrapHandler(func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("WrapHandler example")
		ctx.SetContentType("text/plain")
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
