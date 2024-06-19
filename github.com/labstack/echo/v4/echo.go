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

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/whatap/go-api/httpc"
	wisql "github.com/whatap/go-api/instrumentation/database/sql/whatapsql"
	"github.com/whatap/go-api/instrumentation/github.com/labstack/echo/v4/whatapecho"
	"github.com/whatap/go-api/method"
	"github.com/whatap/go-api/trace"
)

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

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

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

	t := &Template{
		templates: template.Must(template.ParseGlob("templates/github.com/labstack/v4/*.html")),
	}

	e := echo.New()
	e.Renderer = t

	e.HTTPErrorHandler = whatapecho.WrapHTTPErrorHandler(e.DefaultHTTPErrorHandler)
	if IsWhatap {
		e.Pre(whatapecho.Middleware())
	}
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		data := &HTMLData{}
		data.Title = "echo/v4 server"
		data.Content = c.Request().RequestURI
		return c.Render(http.StatusOK, "index.html", data)
	})
	e.GET("/index", func(c echo.Context) error {
		fmt.Println("Request -", c.Request())

		ctx := c.Request().Context()
		trace.Step(ctx, "Text Message", "Message", 3, 3)

		getUser(ctx)
		fmt.Println("Response -", c.Response())
		return c.String(http.StatusOK, c.Request().RequestURI+"<br/><hr/>")
	})

	e.GET("/main", func(c echo.Context) error {
		fmt.Println("Request -", c.Request())
		ctx := c.Request().Context()
		trace.Step(ctx, "Text Message 2", "Message2", 6, 6)
		fmt.Println("Response -", c.Response())
		return c.String(http.StatusOK, c.Request().RequestURI+"<br/><hr/>")
	})

	e.GET("/httpc", func(c echo.Context) error {
		ctx := c.Request().Context()
		fmt.Println("Request -", c.Request())
		var buffer bytes.Buffer
		buffer.WriteString(c.Request().RequestURI + "<br/><hr/>")

		callUrl := "http://localhost:8081/index"
		httpcCtx, _ := httpc.Start(ctx, callUrl)
		if statusCode, data, err := httpWithRequest("GET", callUrl, "", trace.GetMTrace(ctx)); err == nil {
			httpc.End(httpcCtx, statusCode, "", nil)
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", statusCode, ", data=", data))
		} else {
			httpc.End(httpcCtx, -1, "", err)
			buffer.WriteString(fmt.Sprintln("httpc Error callUrl=", callUrl, ", err=", err))
		}

		trace.Step(ctx, "Text Message 2", "Message2", 6, 6)
		fmt.Println("Response -", c.Response())
		return c.HTMLBlob(http.StatusOK, buffer.Bytes())
	})

	e.GET("/httpc/query", func(c echo.Context) error {
		var cnt int
		if i, err := strconv.Atoi(c.QueryParam("cnt")); err == nil {
			cnt = i
		} else {
			cnt = 1
		}
		var sleep int
		if i, err := strconv.Atoi(c.QueryParam("sleep")); err == nil {
			sleep = i
		} else {
			sleep = 1
		}
		var loop int
		if i, err := strconv.Atoi(c.QueryParam("loop")); err == nil {
			loop = i
		} else {
			loop = 1
		}

		ctx := c.Request().Context()
		fmt.Println("Request -", c.Request())
		var buffer bytes.Buffer
		buffer.WriteString(c.Request().RequestURI + "<br/><hr/>")

		callUrl := "http://localhost:8081/index"
		for i := 0; i < loop; i++ {
			for j := 0; j < cnt; j++ {

				httpcCtx, _ := httpc.Start(ctx, callUrl)
				if statusCode, data, err := httpWithRequest("GET", callUrl, "", trace.GetMTrace(ctx)); err == nil {
					httpc.End(httpcCtx, statusCode, "", nil)
					buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", statusCode, ", data=", data))
				} else {
					httpc.End(httpcCtx, -1, "", err)
					buffer.WriteString(fmt.Sprintln("httpc Error callUrl=", callUrl, ", err=", err))
				}
			}
		}
		if sleep > 0 {
			time.Sleep(time.Duration(sleep) * time.Millisecond)
		}
		trace.Step(ctx, "Text Message 2", "Message2", 6, 6)
		fmt.Println("Response -", c.Response())
		return c.HTMLBlob(http.StatusOK, buffer.Bytes())
	})

	e.GET("/httpc/unknown", func(c echo.Context) error {
		ctx := c.Request().Context()
		fmt.Println("Request -", c.Request())
		var buffer bytes.Buffer
		buffer.WriteString(c.Request().RequestURI + "<br/><hr/>")

		callUrl := "http://localhost:8081/unknown"
		httpcCtx, _ := httpc.Start(ctx, callUrl)
		if statusCode, data, err := httpWithRequest("GET", callUrl, "", trace.GetMTrace(ctx)); err == nil {
			httpc.End(httpcCtx, statusCode, "", nil)
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", statusCode, ", data=", data))
		} else {
			httpc.End(httpcCtx, -1, "", err)
			buffer.WriteString(fmt.Sprintln("httpc Error callUrl=", callUrl, ", err=", err))
		}

		trace.Step(ctx, "Text Message 2", "Message2", 6, 6)
		fmt.Println("Response -", c.Response())
		return c.HTMLBlob(http.StatusOK, buffer.Bytes())
	})

	e.GET("/sql/select", func(c echo.Context) error {

		if err := c.Request().ParseForm(); err != nil {
			fmt.Println("ParseForm error ", err)
		}

		ctx := c.Request().Context()
		var buffer bytes.Buffer
		buffer.WriteString(c.Request().RequestURI + "<br/><hr/>")

		var query string

		// 복수 Row를 갖는 SQL 쿼리
		var id int
		var subject string
		query = "select id, subject from tbl_faq limit 10"
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return fmt.Errorf("db.QueryContext error:%s", err.Error())
		}
		defer rows.Close() //반드시 닫는다 (지연하여 닫기)

		for rows.Next() {
			err := rows.Scan(&id, &subject)
			if err != nil {
				return fmt.Errorf("rows.Scan error:%s", err.Error())
			}
			buffer.WriteString(fmt.Sprintln(id, subject))
		}

		// Prepared Statement 생성
		query = "select id, subject from tbl_faq where id = ? limit ?"
		stmt, err := db.PrepareContext(ctx, query)
		if err != nil {
			return fmt.Errorf("db.Prepare error:%s", err.Error())
		}
		defer stmt.Close()

		// Prepared Statement 실행
		params := make([]interface{}, 0)
		params = append(params, 8)
		params = append(params, 1)

		rows1, _ := stmt.QueryContext(ctx, params...) //Placeholder 파라미터 순서대로 전달
		defer rows1.Close()

		for rows1.Next() {
			err := rows1.Scan(&id, &subject)
			if err != nil {
				return fmt.Errorf("rows1.Scan error:%s", err.Error())
			}
			buffer.WriteString(fmt.Sprintln(id, subject))
		}

		rows2, _ := stmt.QueryContext(ctx, 8, 1) //Placeholder 파라미터 순서대로 전달
		defer rows2.Close()

		for rows1.Next() {
			err := rows2.Scan(&id, &subject)
			if err != nil {
				return fmt.Errorf("rows2.Scan error:%s", err.Error())
			}
			buffer.WriteString(fmt.Sprintln(id, subject))
		}

		return c.HTMLBlob(http.StatusOK, buffer.Bytes())

	})

	e.GET("/sql/query", func(c echo.Context) error {

		var cnt int
		if i, err := strconv.Atoi(c.QueryParam("cnt")); err == nil {
			cnt = i
		} else {
			cnt = 1
		}
		var sleep int
		if i, err := strconv.Atoi(c.QueryParam("sleep")); err == nil {
			sleep = i
		} else {
			sleep = 1
		}
		var loop int
		if i, err := strconv.Atoi(c.QueryParam("loop")); err == nil {
			loop = i
		} else {
			loop = 1
		}

		ctx := c.Request().Context()
		var buffer bytes.Buffer
		buffer.WriteString(c.Request().RequestURI + "<br/><hr/>")

		var query string

		// 복수 Row를 갖는 SQL 쿼리
		var id int
		var subject string
		for i := 0; i < loop; i++ {
			for j := 0; j < cnt; j++ {
				query = "select id, subject from tbl_faq limit 10"
				rows, err := db.QueryContext(ctx, query)
				if err != nil {
					return fmt.Errorf("db.QueryContext error:%s", err.Error())
				}
				defer rows.Close() //반드시 닫는다 (지연하여 닫기)

				for rows.Next() {
					err := rows.Scan(&id, &subject)
					if err != nil {
						return fmt.Errorf("rows.Scan error:%s", err.Error())
					}
					buffer.WriteString(fmt.Sprintln(id, subject))
				}

				// Prepared Statement 생성
				query = "select id, subject from tbl_faq where id = ? limit ?"
				stmt, err := db.PrepareContext(ctx, query)
				if err != nil {
					return fmt.Errorf("db.Prepare error:%s", err.Error())
				}
				defer stmt.Close()

				// Prepared Statement 실행
				params := make([]interface{}, 0)
				params = append(params, 8)
				params = append(params, 1)

				rows1, _ := stmt.QueryContext(ctx, params...) //Placeholder 파라미터 순서대로 전달
				defer rows1.Close()

				for rows1.Next() {
					err := rows1.Scan(&id, &subject)
					if err != nil {
						return fmt.Errorf("rows1.Scan error:%s", err.Error())
					}
					buffer.WriteString(fmt.Sprintln(id, subject))
				}

				rows2, _ := stmt.QueryContext(ctx, 8, 1) //Placeholder 파라미터 순서대로 전달
				defer rows2.Close()

				for rows1.Next() {
					err := rows2.Scan(&id, &subject)
					if err != nil {
						return fmt.Errorf("rows2.Scan error:%s", err.Error())
					}
					buffer.WriteString(fmt.Sprintln(id, subject))
				}
			}
		}

		if sleep > 0 {
			time.Sleep(time.Duration(sleep) * time.Millisecond)
		}
		return c.HTMLBlob(http.StatusOK, buffer.Bytes())
	})

	e.GET("/panic", func(c echo.Context) error {
		fmt.Println("Request -", c.Request())
		panic(fmt.Errorf("custom panic"))

		defer fmt.Println("Response -", c.Response())
		return c.String(http.StatusOK, "Hello, World!\n")
	})

	e.GET("/input", func(c echo.Context) error {
		var buffer bytes.Buffer
		buffer.WriteString("<html><body>")
		form := `<form action="/saveUrlencoded" method="post" >
    Name : <input type="text" name="name" value="">
    Value : <input type="text" name="value" value="">
    <input type="submit" value="Action" />
</form></body></html>`
		buffer.WriteString(c.Request().RequestURI + "<br/><hr/>")
		buffer.WriteString(form)
		return c.HTMLBlob(http.StatusOK, buffer.Bytes())

	})

	e.POST("/saveUrlencoded", func(c echo.Context) error {
		fmt.Println("Request, ", c.Request())

		if c.Request() != nil {
			name := c.Request().FormValue("name")
			val := c.Request().FormValue("value")
			fmt.Println("c.Request() FormValue ", name, ", ", val)

			params := c.Request().Form
			for k, v := range params {
				fmt.Println("c.Request().Form k=", k, ", v=", v, ",")
			}

			name1 := c.Request().PostFormValue("name")
			val1 := c.Request().PostFormValue("value")
			fmt.Println("c.Request() PostFormValue ", name1, ", ", val1)

			params1 := c.Request().PostForm
			for k, v := range params1 {
				fmt.Println("c.Request().PostForm k=", k, ", v=", v, ",")
			}
		}

		name := c.FormValue("name")
		val := c.FormValue("value")
		fmt.Println("params 11 saveUrlencoded ParseForm ", name, ", ", val)

		if params, err := c.FormParams(); err == nil {
			for k, v := range params {
				fmt.Println("c.FormParams() k=", k, ", v=", v, ",")
			}
		}

		if err := c.Request().ParseForm(); err != nil {
			fmt.Println("saveUrlencoded ParseForm error ", err)
		}

		var buffer bytes.Buffer
		buffer.WriteString(c.Request().RequestURI + "<br/><hr/>")
		return c.HTMLBlob(http.StatusOK, buffer.Bytes())

	})
	e.GET("/inputFile", func(c echo.Context) error {
		var buffer bytes.Buffer
		buffer.WriteString("<html><body>")
		form := `<form action="/upload" method="post" enctype="multipart/form-data">
    Name : <input type="text" name="name" value="">
    Value : <input type="text" name="email" value="">
    File : <input type="file" name="file" value="">
    <input type="submit" value="Action" />
</form></body></html>`
		buffer.WriteString(c.Request().RequestURI + "<br/><hr/>")
		buffer.WriteString(form)
		return c.HTMLBlob(http.StatusOK, buffer.Bytes())

	})
	e.POST("/upload", func(c echo.Context) error {
		fmt.Println("Request, ", c.Request())
		var buffer bytes.Buffer

		// Read form fields
		name := c.FormValue("name")
		email := c.FormValue("email")
		fmt.Println("fields name=", name, ",email=", email)

		buffer.WriteString("name=" + name + "<br/>")
		buffer.WriteString("email=" + email + "<br/>")

		//-----------
		// Read file
		//-----------

		// Source
		file, err := c.FormFile("file")
		if err != nil {
			fmt.Println("c.FromFile err=", err)
			return err
		}
		src, err := file.Open()
		if err != nil {
			fmt.Println("file.Open() err=", err)
			return err
		}
		defer src.Close()

		// Destination
		dst, err := os.Create("./" + file.Filename)
		if err != nil {
			fmt.Println("os.Create ", file.Filename, ", err=", err)
			return err
		}
		defer dst.Close()

		// Copy
		if _, err = io.Copy(dst, src); err != nil {
			fmt.Println("io.Copy src to dest err", err)
			return err
		}
		fmt.Println("upload ok ", file.Filename, ", size=", strconv.FormatInt(file.Size, 10))
		buffer.WriteString("upload ok " + file.Filename + ", size=" + strconv.FormatInt(file.Size, 10))

		return c.HTMLBlob(http.StatusOK, buffer.Bytes())

	})

	fmt.Println("Start :", port, ", Agent Udp Port:", udpPort)
	e.Start(fmt.Sprintf(":%d", port))
}
