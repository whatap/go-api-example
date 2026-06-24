// Fiber v2 (github.com/gofiber/fiber/v2) Instrumentation Example
//
// This example demonstrates how to instrument Fiber v2 web framework applications
// using WhaTap Go API.
//
// INSTRUMENTATION OVERVIEW:
// 1. Call trace.Init() at application startup
// 2. Add whatapfiber.Middleware() to Fiber app
// 3. Use c.Context() to get transaction context in handlers
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
// CONTEXT ACCESS:
// Fiber uses c.Context() to get the underlying fasthttp.RequestCtx,
// which implements context.Context interface.
// This context links all operations to the HTTP transaction.

package main

import (
	"bytes"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html"
	log "github.com/sirupsen/logrus"
	"github.com/whatap/go-api/httpc"
	"github.com/whatap/go-api/instrumentation/github.com/gofiber/fiber/v2/whatapfiber"
	"github.com/whatap/go-api/method"
	"github.com/whatap/go-api/trace"
)

var Database *sql.DB

type HTMLData struct {
	Title   string
	Content string
	//HTMLContent template.HTML
}

// ============================================================
// Example: Custom Method Tracing
// ============================================================
// Use method.Start/End to trace custom functions.
// This creates a "Method" step in the transaction trace.
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

	portPtr := flag.Int("p", 8080, "web port. default 8080")
	udpPortPtr := flag.Int("up", 6600, "agent port(udp). defalt 6600")
	dataSourcePtr := flag.String("ds", "whatap:whatap1234!@tcp(localhost:3306)/whatap_demo", "dataSourceName")
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

	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		log.Panic(err)
	}
	if db == nil {
		log.Panic("Db nil")
	}
	Database = db
	log.Infof("%+v\n", *Database)
	defer Database.Close()

	engine := html.New("./templates/github.com/fiber/v2", ".html")

	// ============================================================
	// INSTRUMENTATION STEP 2: Add WhaTap Middleware
	// ============================================================
	// BEFORE (Original):
	//   r := fiber.New(fiber.Config{...})
	//   r.Use(recover.New())
	//   r.Get("/hello", handler)
	//
	// AFTER (Instrumented):
	//   r := fiber.New(fiber.Config{...})
	//   r.Use(recover.New())
	//   r.Use(whatapfiber.Middleware())  // Add this line
	//   r.Get("/hello", handler)
	//
	// NOTE: Use Middleware() with parentheses (function call, not value).
	r := fiber.New(fiber.Config{
		StrictRouting: true,
		Views:         engine,
	})

	r.Use(recover.New())
	r.Use(whatapfiber.Middleware()) // WhaTap middleware

	// app.Get("/", index)
	// app.Get("/panic", panicFunc)
	// app.Get("/selectRows", selectRow)
	// app.Get("/sleepSecond", sleepSecond)

	// ============================================================
	// INSTRUMENTATION STEP 3: Access Transaction Context
	// ============================================================
	// Use c.Context() to get the transaction context.
	// This returns fasthttp.RequestCtx which implements context.Context.
	// Pass this context to trace.Step(), method.Start(), httpc.Start(), etc.
	r.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title": "fiber/v2",
		})
	})

	// Example: Using context with custom steps and method tracing
	r.Get("/index", func(c *fiber.Ctx) error {
		// fmt.Println("Request -", c.Request)

		// Get transaction context from Fiber context
		ctx := c.Context()

		// Add custom step to transaction trace
		trace.Step(ctx, "Text Message", "Message", 3, 3)

		// Call custom method - will be tracked as "Method" step
		getUser(ctx)
		return c.SendString("/index <br/>Test Body")
	})

	// Example: URL path parameters with context
	r.Get("/main/:idx", func(c *fiber.Ctx) error {
		// fmt.Println("Request -", c.Request)
		fmt.Println("Param idx=", c.Params("idx", "0"))
		ctx := c.Context()
		trace.Step(ctx, "Text Message 2", "Message2", 6, 6)
		return c.SendString("/main <br/>Test Body")
	})

	// ============================================================
	// HTTP Client Call with Distributed Tracing
	// ============================================================
	// Use httpc.Start/End for outgoing HTTP call tracking.
	// Use trace.GetMTrace(ctx) to propagate trace context via headers.
	r.Get("/httpc", func(c *fiber.Ctx) error {
		ctx := c.Context()
		//fmt.Println("Request -", c.Request())

		callUrl := "http://localhost:8081/index"

		// Start HTTP client call tracking
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
		return c.SendString(string(buffer.Bytes()))
	})

	r.Get("/httpc/unknown", func(c *fiber.Ctx) error {
		ctx := c.Context()
		//fmt.Println("Request -", c.Request())

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

		trace.Step(ctx, "Text Message 2", "Message2", 6, 6)
		return c.SendString(string(buffer.Bytes()))
	})

	// ============================================================
	// SQL Query Tracking Example
	// ============================================================
	// SQL queries are automatically tracked when using whatapsql.
	// Pass context to QueryContext/ExecContext for transaction linking.
	// Note: This example uses standard database/sql, not whatapsql.
	r.Get("/sql/select", func(c *fiber.Ctx) error {
		ctx := c.Context()
		var buffer bytes.Buffer
		var query string

		// SQL query returning multiple rows
		var id int
		var subject string
		query = "select id, subject from tbl_faq limit 10"

		// Pass context to QueryContext for tracking
		rows, err := db.QueryContext(ctx, query)
		defer rows.Close() // must close (deferred)

		for rows.Next() {
			err := rows.Scan(&id, &subject)
			if err != nil {
				c.SendStatus(http.StatusInternalServerError)
				return err
			}
			buffer.WriteString(fmt.Sprintln(id, subject))
		}

		// create prepared statement
		query = "select id, subject from tbl_faq where id = ? limit ?"
		stmt, err := db.PrepareContext(ctx, query)
		if err != nil {
			c.SendStatus(http.StatusInternalServerError)
			return err
		}
		defer stmt.Close()

		// execute prepared statement
		params := make([]interface{}, 0)
		params = append(params, 8)
		params = append(params, 1)

		rows1, err1 := stmt.QueryContext(ctx, params...) // pass placeholder parameters in order
		if err1 != nil {
			c.SendStatus(http.StatusInternalServerError)
			return err
		}
		defer rows1.Close()

		for rows1.Next() {
			err := rows1.Scan(&id, &subject)
			if err != nil {
				c.SendStatus(http.StatusInternalServerError)
				return err
			}
			buffer.WriteString(fmt.Sprintln(id, subject))
		}

		rows2, err2 := stmt.QueryContext(ctx, 8, 1) // pass placeholder parameters in order
		if err2 != nil {
			c.SendStatus(http.StatusInternalServerError)
			return err2
		}
		defer rows2.Close()

		for rows1.Next() {
			err := rows2.Scan(&id, &subject)
			if err != nil {
				c.SendStatus(http.StatusInternalServerError)
				return err
			}
			buffer.WriteString(fmt.Sprintln(id, subject))
		}

		return c.SendString(string(buffer.Bytes()))
	})

	// ============================================================
	// Panic Tracking Example
	// ============================================================
	// Panics are automatically captured by the middleware.
	r.Get("/panic", func(c *fiber.Ctx) error {
		panic(fmt.Errorf("custom panic"))
		return c.SendString("/panic <br/>Test Body")
	})

	// Form handling examples (standard patterns)
	r.Get("/input", func(c *fiber.Ctx) error {
		ctx := c.Context()
		var buffer bytes.Buffer
		buffer.WriteString("<html><body>")
		form := `<form action="/saveUrlencoded" method="post" >
	   Name : <input type="text" name="name" value="">
	   Value : <input type="text" name="value" value="">
	   <input type="submit" value="Action" />
	</form></body></html>`
		buffer.WriteString(string(ctx.RequestURI()) + "<br/><hr/>")
		buffer.WriteString(form)

		c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		return c.SendString(string(buffer.Bytes()))
	})

	r.Post("/saveUrlencoded", func(c *fiber.Ctx) error {
		ctx := c.Context()
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

		return c.SendString(string(buffer.Bytes()))
	})

	r.Get("/inputFile", func(c *fiber.Ctx) error {
		ctx := c.Context()
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

		c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		return c.SendString(string(buffer.Bytes()))
	})

	r.Post("/upload", func(c *fiber.Ctx) error {
		// fmt.Println("Request, ", c.Request)
		var buffer bytes.Buffer

		// Read form fields
		name := string(c.FormValue("name"))
		email := string(c.FormValue("email"))
		fmt.Println("fields name=", name, ",email=", email)

		buffer.WriteString("name=" + name + "<br/>")
		buffer.WriteString("email=" + email + "<br/>")

		//-----------
		// Read file
		//-----------
		file, _ := c.FormFile("file")
		fmt.Println(file.Filename + " uploaded")

		// save file

		// method 1.
		// save file using the built-in helper
		if err := c.SaveFile(file, file.Filename); err != nil {
			panic(err)
		}

		fmt.Println("upload ok ", file.Filename, ", size=", strconv.FormatInt(file.Size, 10))
		buffer.WriteString("upload ok " + file.Filename + ", size=" + strconv.FormatInt(file.Size, 10))

		return c.SendString(string(buffer.Bytes()))
	})

	log.Fatal(r.Listen(fmt.Sprintf(":%d", port)))
}
