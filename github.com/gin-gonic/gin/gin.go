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

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"github.com/whatap/go-api/httpc"
	"github.com/whatap/go-api/instrumentation/github.com/gin-gonic/gin/whatapgin"
	"github.com/whatap/go-api/method"

	whatapsql "github.com/whatap/go-api/sql"
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
	portPtr := flag.Int("p", 8080, "web port. default 8080  ")
	udpPortPtr := flag.Int("up", 6600, "agent port(udp). defalt 6600 ")
	flag.Parse()
	port := *portPtr
	udpPort := *udpPortPtr

	config := make(map[string]string)
	config["net_udp_port"] = fmt.Sprintf("%d", udpPort)
	trace.Init(config)
	defer trace.Shutdown()

	r := gin.Default()
	r.Use(whatapgin.Middleware())
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
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

	r.GET("/sql/select", func(c *gin.Context) {
		ctx := c.Request.Context()
		var buffer bytes.Buffer
		var query string

		sqlCtx, _ := whatapsql.StartOpen(ctx, "doremimaker:doremimaker@tcp(192.168.56.101:3306)/doremimaker")
		db, err := sql.Open("mysql", "doremimaker:doremimaker@tcp(192.168.56.101:3306)/doremimaker")
		whatapsql.End(sqlCtx, err)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
		defer db.Close()

		// 복수 Row를 갖는 SQL 쿼리
		var id int
		var subject string
		query = "select id, subject from tbl_faq limit 10"
		sqlCtx, _ = whatapsql.Start(ctx, "doremimaker:doremimaker@tcp(192.168.56.101:3306)/doremimaker", query)
		rows, err := db.QueryContext(ctx, query)
		whatapsql.End(sqlCtx, err)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
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
		stmt, err := db.Prepare(query)
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

		sqlCtx, _ = whatapsql.StartWithParamArray(ctx, "doremimaker:doremimaker@tcp(192.168.56.101:3306)/doremimaker", query, params)
		rows1, err1 := stmt.QueryContext(ctx, params...) //Placeholder 파라미터 순서대로 전달
		whatapsql.End(sqlCtx, err1)
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

		sqlCtx, _ = whatapsql.StartWithParam(ctx, "doremimaker:doremimaker@tcp(192.168.56.101:3306)/doremimaker", query, params...)
		rows2, err2 := stmt.QueryContext(ctx, 8, 1) //Placeholder 파라미터 순서대로 전달
		whatapsql.End(sqlCtx, err2)
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

	fmt.Println("Start :", port, ", Agent Udp Port:", udpPort)
	r.Run(fmt.Sprintf(":%d", port))

}