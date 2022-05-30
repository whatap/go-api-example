package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"github.com/whatap/go-api/httpc"
	"github.com/whatap/go-api/instrumentation/github.com/gin-gonic/gin/whatapgin"
	"github.com/whatap/go-api/method"

	wisql "github.com/whatap/go-api/instrumentation/database/sql/whatapsql"
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

	r := gin.Default()
	r.Use(whatapgin.Middleware())

	r.LoadHTMLGlob("templates/github.com/gin-gonic/*")

	r.GET("/", func(c *gin.Context) {
		fmt.Println("Request -", c.Request)
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Title":   "gin server",
			"Content": c.Request.RequestURI,
		},
		)
	})
	r.GET("/index", func(c *gin.Context) {
		fmt.Println("Request -", c.Request)

		ctx := c.Request.Context()
		trace.Step(ctx, "Text Message", "Message", 3, 3)

		getUser(ctx)
		c.JSON(http.StatusOK, gin.H{
			"message": "/index <br/>Test Body",
		})
	})

	r.GET("/main", func(c *gin.Context) {
		fmt.Println("Request -", c.Request)
		ctx := c.Request.Context()
		trace.Step(ctx, "Text Message 2", "Message2", 6, 6)
		c.JSON(http.StatusOK, gin.H{
			"message": "/main <br/>Test Body",
		})
	})

	r.GET("/httpc", func(c *gin.Context) {
		ctx := c.Request.Context()
		fmt.Println("Request -", c.Request)

		callUrl := "http://localhost:8081/index"
		httpcCtx, _ := httpc.Start(ctx, callUrl)
		var buffer bytes.Buffer
		if statusCode, data, err := httpWithRequest("GET", callUrl, "", httpc.GetMTrace(httpcCtx)); err == nil {
			httpc.End(httpcCtx, statusCode, "", nil)
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", statusCode, ", data=", data))
		} else {
			httpc.End(httpcCtx, -1, "", err)
			buffer.WriteString(fmt.Sprintln("httpc Error callUrl=", callUrl, ", err=", err))
		}

		trace.Step(ctx, "Text Message 2", "Message2", 6, 6)
		c.JSON(http.StatusOK, gin.H{
			"message": string(buffer.Bytes()),
		})
	})

	r.GET("/httpc/unknown", func(c *gin.Context) {
		ctx := c.Request.Context()
		fmt.Println("Request -", c.Request)

		callUrl := "http://localhost:8081/unknown"
		httpcCtx, _ := httpc.Start(ctx, callUrl)
		var buffer bytes.Buffer
		if statusCode, data, err := httpWithRequest("GET", callUrl, "", httpc.GetMTrace(httpcCtx)); err == nil {
			httpc.End(httpcCtx, statusCode, "", nil)
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", statusCode, ", data=", data))
		} else {
			httpc.End(httpcCtx, -1, "", err)
			buffer.WriteString(fmt.Sprintln("httpc Error callUrl=", callUrl, ", err=", err))
		}

		trace.Step(ctx, "Text Message 2", "Message2", 6, 6)
		c.JSON(http.StatusOK, gin.H{
			"message": string(buffer.Bytes()),
		})
	})

	r.GET("/sql/select", func(c *gin.Context) {
		ctx := c.Request.Context()
		var buffer bytes.Buffer
		var query string

		// 복수 Row를 갖는 SQL 쿼리
		var id int
		var subject string
		query = "select id, subject from tbl_faq limit 10"
		rows, err := db.QueryContext(ctx, query)
		defer rows.Close() //반드시 닫는다 (지연하여 닫기)

		for rows.Next() {
			err := rows.Scan(&id, &subject)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": err.Error(),
				})
				return
			}
			buffer.WriteString(fmt.Sprintln(id, subject))
		}

		// Prepared Statement 생성
		query = "select id, subject from tbl_faq where id = ? limit ?"
		stmt, err := db.PrepareContext(ctx, query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
		defer stmt.Close()

		// Prepared Statement 실행
		params := make([]interface{}, 0)
		params = append(params, 8)
		params = append(params, 1)

		rows1, err1 := stmt.QueryContext(ctx, params...) //Placeholder 파라미터 순서대로 전달
		if err1 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err1.Error(),
			})
			return
		}
		defer rows1.Close()

		for rows1.Next() {
			err := rows1.Scan(&id, &subject)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": err.Error(),
				})
				return
			}
			buffer.WriteString(fmt.Sprintln(id, subject))
		}

		rows2, err2 := stmt.QueryContext(ctx, 8, 1) //Placeholder 파라미터 순서대로 전달
		if err2 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err2.Error(),
			})
			return
		}
		defer rows2.Close()

		for rows1.Next() {
			err := rows2.Scan(&id, &subject)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": err.Error(),
				})
				return
			}
			buffer.WriteString(fmt.Sprintln(id, subject))
		}

		c.JSON(http.StatusOK, gin.H{
			"message": string(buffer.Bytes()),
		})
	})

	r.GET("/panic", func(c *gin.Context) {
		panic(fmt.Errorf("custom panic"))
		c.JSON(http.StatusOK, gin.H{
			"message": "/main <br/>Test Body",
		})
	})

	r.GET("/input", func(c *gin.Context) {
		var buffer bytes.Buffer
		buffer.WriteString("<html><body>")
		form := `<form action="/saveUrlencoded" method="post" >
    Name : <input type="text" name="name" value="">
    Value : <input type="text" name="value" value="">
    <input type="submit" value="Action" />
</form></body></html>`
		buffer.WriteString(c.Request.RequestURI + "<br/><hr/>")
		buffer.WriteString(form)
		c.Data(http.StatusOK, "text/html;charset=UTF-8", buffer.Bytes())

	})

	r.POST("/saveUrlencoded", func(c *gin.Context) {
		fmt.Println("Request, ", c.Request)

		if c.Request != nil {
			name := c.Request.FormValue("name")
			val := c.Request.FormValue("value")
			fmt.Println("c.Request() FormValue ", name, ", ", val)

			params := c.Request.Form
			for k, v := range params {
				fmt.Println("c.Request().Form k=", k, ", v=", v, ",")
			}

			name1 := c.Request.PostFormValue("name")
			val1 := c.Request.PostFormValue("value")
			fmt.Println("c.Request() PostFormValue ", name1, ", ", val1)

			params1 := c.Request.PostForm
			for k, v := range params1 {
				fmt.Println("c.Request().PostForm k=", k, ", v=", v, ",")
			}
		}

		name := c.PostForm("name")
		val := c.PostForm("value")
		fmt.Println("params 11 saveUrlencoded ParseForm ", name, ", ", val)

		if err := c.Request.ParseForm(); err != nil {
			fmt.Println("saveUrlencoded ParseForm error ", err)
		}

		var buffer bytes.Buffer
		buffer.WriteString(c.Request.RequestURI + "<br/><hr/>")
		c.String(http.StatusOK, string(buffer.Bytes()))

	})
	r.GET("/inputFile", func(c *gin.Context) {
		var buffer bytes.Buffer
		buffer.WriteString("<html><body>")
		form := `<form action="/upload" method="post" enctype="multipart/form-data">
    Name : <input type="text" name="name" value="">
    Value : <input type="text" name="email" value="">
    File : <input type="file" name="file" value="">
    <input type="submit" value="Action" />
</form></body></html>`
		buffer.WriteString(c.Request.RequestURI + "<br/><hr/>")
		buffer.WriteString(form)
		c.Data(http.StatusOK, "text/html;charset=UTF-8", buffer.Bytes())
	})
	r.POST("/upload", func(c *gin.Context) {
		fmt.Println("Request, ", c.Request)
		var buffer bytes.Buffer

		// Read form fields
		name := c.PostForm("name")
		email := c.PostForm("email")
		fmt.Println("fields name=", name, ",email=", email)

		buffer.WriteString("name=" + name + "<br/>")
		buffer.WriteString("email=" + email + "<br/>")

		//-----------
		// Read file
		//-----------
		file, _ := c.FormFile("file")
		fmt.Println(file.Filename + " uploaded")

		// 파일 저장

		// 방법 1.
		// 기본 제공 함수로 파일 저장
		c.SaveUploadedFile(file, file.Filename)
		fmt.Println("upload ok ", file.Filename, ", size=", strconv.FormatInt(file.Size, 10))
		buffer.WriteString("upload ok " + file.Filename + ", size=" + strconv.FormatInt(file.Size, 10))

		c.String(http.StatusOK, string(buffer.Bytes()))

	})

	fmt.Println("Start :", port, ", Agent Udp Port:", udpPort)
	r.Run(fmt.Sprintf(":%d", port))

}
