package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
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

	db, err := wisql.OpenContext(context.Background(), "mysql", dataSource)
	if err != nil {
		fmt.Println("Error service whatapsql.Open ", err)
		return
	}
	defer db.Close()

	e := echo.New()
	e.HTTPErrorHandler = whatapecho.WrapHTTPErrorHandler(e.DefaultHTTPErrorHandler)
	e.Pre(whatapecho.Middleware())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		var buffer bytes.Buffer
		buffer.WriteString(c.Request().RequestURI + "<br/><hr/>")

		buffer.WriteString("<a href='/index'>/index</a><br>")
		buffer.WriteString("<a href='/main'>/main</a><br>")
		buffer.WriteString("<a href='/httpc'>/httpc</a><br>")
		buffer.WriteString("<a href='/sql/select'>/sql/select</a><br>")
		buffer.WriteString("<a href='/panic'>/panic</a><br>")

		return c.HTMLBlob(http.StatusOK, buffer.Bytes())
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
		if statusCode, data, err := httpWithRequest("GET", callUrl, "", httpc.GetMTrace(httpcCtx)); err == nil {
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

	e.GET("/panic", func(c echo.Context) error {
		fmt.Println("Request -", c.Request())
		panic(fmt.Errorf("custom panic"))

		defer fmt.Println("Response -", c.Response())
		return c.String(http.StatusOK, "Hello, World!\n")
	})

	fmt.Println("Start :", port, ", Agent Udp Port:", udpPort)
	e.Start(fmt.Sprintf(":%d", port))
}
