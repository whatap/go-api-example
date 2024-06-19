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
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/whatap/go-api/httpc"
	wisql "github.com/whatap/go-api/instrumentation/database/sql/whatapsql"
	"github.com/whatap/go-api/instrumentation/net/http/whataphttp"
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
	var f WhatapHttpGet
	f = http.Get
	WrapResponse(callUrl, f)
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
		if resp != nil {
			fmt.Println(">>>>> httpGet error but resp not nil and close ")
			defer resp.Body.Close()
		}
		fmt.Println(err)
		return -1, "", err
	}
}

type WhatapHttpGet func(string) (resp *http.Response, err error)

func WrapClientDo(client *http.Client, req *http.Request) (*http.Response, error) {
	return client.Do(req)
}

func WrapResponse(url string, f WhatapHttpGet) (*http.Response, error) {
	return f(url)
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
		resp, err1 := client.Do(req)
		if err1 == nil {
			defer resp.Body.Close()
			if data, err2 := ioutil.ReadAll(resp.Body); err2 == nil {
				fmt.Println("status=", resp.StatusCode)
				return resp.StatusCode, string(data), err2
			} else {
				fmt.Println("Read response Error ", err2)
				return resp.StatusCode, "", err2
			}
		}

		if resp != nil {
			fmt.Println(">>>>> client.Do error but resp not nil and close ")
			defer resp.Body.Close()
		}
		fmt.Println("client.Do Error ", err1)

		return -2, "", err

	} else {
		fmt.Println("NewRequest Error ", err)
		return -1, "", err
	}

}

type AccessLogRoundTrip struct {
	transport http.RoundTripper
}

func (this *AccessLogRoundTrip) RoundTrip(req *http.Request) (res *http.Response, err error) {
	fmt.Println("AccessLogRoundTrip Start ")
	if res, err = this.transport.RoundTrip(req); err == nil {
		if res != nil {
			fmt.Println("AccessLogRoundTrip End res is nil ")
		} else {
			fmt.Println("AccessLogRoundTrip End res ", res)
		}
	} else {
		fmt.Println("AccessLogRoundTrip End error ", err)
	}
	return res, err
}

func NewAccessLogRoundTrip(t http.RoundTripper) http.RoundTripper {
	return &AccessLogRoundTrip{t}
}

type HTMLData struct {
	Title   string
	Content string
	//HTMLContent template.HTML
}

type HTMLLink struct {
	Href   string
	Name   string
	Alt    string
	Target string
}
type IndexHTMLData struct {
	HTMLData
	ALink []HTMLLink
}

func unescaped(str string) template.HTML { return template.HTML(str) }

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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)
		fmt.Println("Request -", r)

		tp, err := template.ParseFiles("templates/net/http/server/index.html")
		if err != nil {
			fmt.Println("Template not loaded, ", err)
			return
		}
		data := &HTMLData{}
		data.Title = "net/http server"
		data.Content = r.RequestURI
		tp.Execute(w, data)

		trace.Step(ctx, "Text Message", "Message", 3, 3)

		getUser(ctx)
		fmt.Println("Response -", r.Response)
	})

	http.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)
		// if wErr != nil {
		// 	fmt.Println("Trace start error ", wErr)
		// }
		// defer func() {
		// 	if err := trace.End(ctx, nil); err != nil {
		// 		fmt.Println("Error ", err)
		// 	}
		// }()
		fmt.Println("Request -", r)

		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer
		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		buffer.WriteString("</body></html>")

		_, _ = w.Write(buffer.Bytes())
		trace.Step(ctx, "Text Message", "Message", 3, 3)

		getUser(ctx)
		fmt.Println("Response -", r.Response)

	})

	http.HandleFunc("/main", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)
		w.Header().Add("Content-Type", "text/html")
		fmt.Println("Request -", r)
		var buffer bytes.Buffer
		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		buffer.WriteString("</body></html>")
		_, _ = w.Write(buffer.Bytes())
		trace.Step(ctx, "Text Message 2", "Message2", 6, 6)
		fmt.Println("Response -", r.Response)
	})

	http.HandleFunc("/httpc", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)
		w.Header().Add("Content-Type", "text/html")
		fmt.Println("Request -", r)
		var buffer bytes.Buffer
		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		callUrl := "http://localhost:8081/index"
		httpcCtx, _ := httpc.Start(ctx, callUrl)

		if statusCode, data, err := httpWithRequest("GET", callUrl, "", trace.GetMTrace(ctx)); err == nil {
			httpc.End(httpcCtx, statusCode, "", nil)
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", statusCode, ", data=", data))
		} else {
			httpc.End(httpcCtx, -1, "", err)
			buffer.WriteString(fmt.Sprintln("httpc Error callUrl=", callUrl, ", err=", err))
		}

		buffer.WriteString("</body></html>")
		_, _ = w.Write(buffer.Bytes())
		trace.Step(ctx, "Text Message 2", "Message2", 6, 6)
		fmt.Println("Response -", r.Response)
	})

	http.HandleFunc("/httpc/unknown", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)
		w.Header().Add("Content-Type", "text/html")
		fmt.Println("Request -", r)
		var buffer bytes.Buffer
		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		callUrl := "http://localhost:8081/unknown"
		httpcCtx, _ := httpc.Start(ctx, callUrl)

		if statusCode, data, err := httpWithRequest("GET", callUrl, "", trace.GetMTrace(ctx)); err == nil {
			httpc.End(httpcCtx, statusCode, "", nil)
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", statusCode, ", data=", data))
		} else {
			httpc.End(httpcCtx, -1, "", err)
			buffer.WriteString(fmt.Sprintln("httpc Error callUrl=", callUrl, ", err=", err))
		}

		buffer.WriteString("</body></html>")
		_, _ = w.Write(buffer.Bytes())
		trace.Step(ctx, "Text Message 2", "Message2", 6, 6)
		fmt.Println("Response -", r.Response)
	})

	http.HandleFunc("/roundTripper", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer
		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		callUrl := "http://localhost:8081/index"

		client := http.DefaultClient
		client.Transport = whataphttp.NewRoundTrip(ctx, http.DefaultTransport)
		if resp, err := client.Get(callUrl); err == nil {
			defer resp.Body.Close()
			if data, err := ioutil.ReadAll(resp.Body); err == nil {
				buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", resp.StatusCode, ", data=", string(data)))
			}
		} else {
			if resp != nil {
				fmt.Println(">>>>> client.Get error but resp not nil and close ")
				defer resp.Body.Close()
			}
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", error ", err))
			fmt.Printf("Error %s", err.Error())
		}

		buffer.WriteString("</body></html>")
		_, _ = w.Write(buffer.Bytes())
		trace.Step(ctx, "Text Message", "Message roundTripper", 6, 6)
		fmt.Println("Response -", r.Response)
	})

	http.HandleFunc("/roundTripper/inputCallUrl", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		r.ParseForm()
		callUrl := r.FormValue("url")

		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer
		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		//callUrl := "http://c7default.test.com/index"

		client := http.DefaultClient
		client.Transport = whataphttp.NewRoundTrip(ctx, http.DefaultTransport)
		if resp, err := client.Get(callUrl); err == nil {
			defer resp.Body.Close()
			if data, err := ioutil.ReadAll(resp.Body); err == nil {
				buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", resp.StatusCode, ", data=", string(data)))
			}
		} else {
			if resp != nil {
				fmt.Println(">>>>> client.Get error but resp not nil and close ")
				defer resp.Body.Close()
			}
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", error ", err))
			fmt.Printf("Error %s", err.Error())
		}

		buffer.WriteString("</body></html>")
		_, _ = w.Write(buffer.Bytes())
		trace.Step(ctx, "Text Message", "Message roundTripper", 6, 6)
		fmt.Println("Response -", r.Response)
	})

	http.HandleFunc("/roundTripper/unknown", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer
		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		callUrl := "http://localhost:8081/unknown"

		client := http.DefaultClient
		client.Transport = whataphttp.NewRoundTrip(ctx, http.DefaultTransport)
		if resp, err := client.Get(callUrl); err == nil {
			defer resp.Body.Close()
			if data, err := ioutil.ReadAll(resp.Body); err == nil {
				buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", resp.StatusCode, ", data=", string(data)))
			}
		} else {
			if resp != nil {
				fmt.Println(">>>>> client.Get error but resp not nil and close ")
				defer resp.Body.Close()
			}
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", error ", err))
			fmt.Printf("Error %s", err.Error())
		}

		buffer.WriteString("</body></html>")
		_, _ = w.Write(buffer.Bytes())
		trace.Step(ctx, "Text Message", "Message roundTripper", 6, 6)
		fmt.Println("Response -", r.Response)
	})

	http.HandleFunc("/roundTripper/nil", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer
		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		callUrl := "http://localhost:8081/index"

		client := http.DefaultClient
		client.Transport = whataphttp.NewRoundTrip(ctx, nil)
		if resp, err := client.Get(callUrl); err == nil {
			defer resp.Body.Close()
			if data, err := ioutil.ReadAll(resp.Body); err == nil {
				buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", resp.StatusCode, ", data=", string(data)))
			}
		} else {
			if resp != nil {
				fmt.Println(">>>>> client.Get error but resp not nil and close ")
				defer resp.Body.Close()
			}
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", error ", err))
			fmt.Printf("Error %s", err.Error())
		}

		buffer.WriteString("</body></html>")
		_, _ = w.Write(buffer.Bytes())
		trace.Step(ctx, "Text Message", "Message roundTripper", 6, 6)
		fmt.Println("Response -", r.Response)
	}))

	http.HandleFunc("/roundTripper/nil/unknown", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer
		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		callUrl := "http://localhost:8081/unknown"

		client := http.DefaultClient
		client.Transport = whataphttp.NewRoundTrip(ctx, nil)
		if resp, err := client.Get(callUrl); err == nil {
			defer resp.Body.Close()
			if data, err := ioutil.ReadAll(resp.Body); err == nil {
				buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", resp.StatusCode, ", data=", string(data)))
			}
		} else {
			if resp != nil {
				fmt.Println(">>>>> client.Get error but resp not nil and close ")
				defer resp.Body.Close()
			}
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", error ", err))
			fmt.Printf("Error %s", err.Error())
		}

		buffer.WriteString("</body></html>")
		_, _ = w.Write(buffer.Bytes())
		trace.Step(ctx, "Text Message", "Message roundTripper", 6, 6)
		fmt.Println("Response -", r.Response)
	}))

	http.HandleFunc("/roundTripper/multi", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer
		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		callUrl := "http://localhost:8081/index"

		client := http.DefaultClient
		client.Transport = NewAccessLogRoundTrip(whataphttp.NewRoundTrip(ctx, http.DefaultTransport))
		if resp, err := client.Get(callUrl); err == nil {
			defer resp.Body.Close()
			if data, err := ioutil.ReadAll(resp.Body); err == nil {
				buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", resp.StatusCode, ", data=", string(data)))
			}
		} else {
			if resp != nil {
				fmt.Println(">>>>> client.Get error but resp not nil and close ")
				defer resp.Body.Close()
			}
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", error ", err))
			fmt.Printf("Error %s", err.Error())
		}

		buffer.WriteString("</body></html>")
		_, _ = w.Write(buffer.Bytes())
		trace.Step(ctx, "Text Message", "Message roundTripper", 6, 6)
		fmt.Println("Response -", r.Response)
	})
	http.HandleFunc("/roundTripper/multi/unknown", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer
		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		callUrl := "http://localhost:8081/unknown"

		client := http.DefaultClient
		client.Transport = NewAccessLogRoundTrip(whataphttp.NewRoundTrip(ctx, http.DefaultTransport))
		if resp, err := client.Get(callUrl); err == nil {
			defer resp.Body.Close()
			if data, err := ioutil.ReadAll(resp.Body); err == nil {
				buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", resp.StatusCode, ", data=", string(data)))
			}
		} else {
			if resp != nil {
				fmt.Println(">>>>> client.Get error but resp not nil and close ")
				defer resp.Body.Close()
			}
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", error ", err))
			fmt.Printf("Error %s", err.Error())
		}

		buffer.WriteString("</body></html>")
		_, _ = w.Write(buffer.Bytes())
		trace.Step(ctx, "Text Message", "Message roundTripper", 6, 6)
		fmt.Println("Response -", r.Response)
	})

	http.HandleFunc("/roundTripper/multi/nil", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer
		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		callUrl := "http://localhost:8081/index"

		client := http.DefaultClient
		client.Transport = NewAccessLogRoundTrip(whataphttp.NewRoundTrip(ctx, nil))
		if resp, err := client.Get(callUrl); err == nil {
			defer resp.Body.Close()
			if data, err := ioutil.ReadAll(resp.Body); err == nil {
				buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", resp.StatusCode, ", data=", string(data)))
			}
		} else {
			if resp != nil {
				fmt.Println(">>>>> client.Get error but resp not nil and close ")
				defer resp.Body.Close()
			}
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", error ", err))
			fmt.Printf("Error %s", err.Error())
		}

		buffer.WriteString("</body></html>")
		_, _ = w.Write(buffer.Bytes())
		trace.Step(ctx, "Text Message", "Message roundTripper", 6, 6)
		fmt.Println("Response -", r.Response)
	}))
	http.HandleFunc("/roundTripper/multi/nil/unknown", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer
		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		callUrl := "http://localhost:8081/unknown"

		client := http.DefaultClient
		client.Transport = NewAccessLogRoundTrip(NewAccessLogRoundTrip(whataphttp.NewRoundTrip(ctx, nil)))
		if resp, err := client.Get(callUrl); err == nil {
			defer resp.Body.Close()
			if data, err := ioutil.ReadAll(resp.Body); err == nil {
				buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", statuscode=", resp.StatusCode, ", data=", string(data)))
			}
		} else {
			if resp != nil {
				fmt.Println(">>>>> client.Get error but resp not nil and close ")
				defer resp.Body.Close()
			}
			buffer.WriteString(fmt.Sprintln("httpc callUrl=", callUrl, ", error ", err))
			fmt.Printf("Error %s", err.Error())
		}

		buffer.WriteString("</body></html>")
		_, _ = w.Write(buffer.Bytes())
		trace.Step(ctx, "Text Message", "Message roundTripper", 6, 6)
		fmt.Println("Response -", r.Response)
	}))

	http.Handle("/fileTransport", whataphttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tr := &http.Transport{}
		tr.RegisterProtocol("file", http.NewFileTransport(http.Dir(".")))

		wrapTransport := whataphttp.NewRoundTrip(ctx, tr)

		c := &http.Client{Transport: wrapTransport}
		if r, err := c.Get("file:///file.txt"); err == nil {
			defer r.Body.Close()
			fmt.Println("file print")
			io.Copy(os.Stdout, r.Body)
		} else {
			if r != nil {
				fmt.Println(">>>>> client.Get error but resp not nil and close ")
				defer r.Body.Close()
			}
			fmt.Println("c.Get error ", err)
		}

		const badURL = "file:///no-exist.txt"
		res, err := c.Get(badURL)
		if err != nil {
			fmt.Println("badUrl get error", err)
		}
		if res != nil {
			defer res.Body.Close()
			io.Copy(os.Stdout, res.Body)
			if res.StatusCode != 404 {
				fmt.Printf("for %s, StatusCode = %d, want 404", badURL, res.StatusCode)
			}
		}
	}))

	http.HandleFunc("/wrapHandleFunc", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer
		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		buffer.WriteString("</body></html>")
		_, _ = w.Write(buffer.Bytes())
		trace.Step(r.Context(), "Text Message wrapHandleFunc", "wrapHandleFunc", 6, 6)
	}))

	http.Handle("/wrapHandleFunc1", whataphttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer
		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		buffer.WriteString("</body></html>")
		_, _ = w.Write(buffer.Bytes())
		trace.Step(r.Context(), "Text Message wrapHandleFunc1", "wrapHandleFunc1", 6, 6)
	}))

	http.Handle("/sql/select", whataphttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer
		buffer.WriteString("<html><head><title>net/http server</title></head><body>")
		buffer.WriteString(r.RequestURI + "<br/><hr/>")

		var query string

		// 복수 Row를 갖는 SQL 쿼리
		var id int
		var subject string
		query = "select id, subject from tbl_faq limit 10"
		rows, err := db.QueryContext(ctx, query)
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
		stmt, err := db.PrepareContext(ctx, query)
		if err != nil {
			return
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
				return
			}
			buffer.WriteString(fmt.Sprintln(id, subject))
		}

		rows2, _ := stmt.QueryContext(ctx, 8, 1) //Placeholder 파라미터 순서대로 전달
		defer rows2.Close()

		for rows1.Next() {
			err := rows1.Scan(&id, &subject)
			if err != nil {
				return
			}
			buffer.WriteString(fmt.Sprintln(id, subject))
		}

		buffer.WriteString("</body></html>")
		_, _ = w.Write(buffer.Bytes())

	}))

	http.Handle("/panic", whataphttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(fmt.Errorf("custom panic"))
	}))

	fmt.Println("Start :", port, ", Agent Udp Port:", udpPort)

	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
