package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/whatap/go-api/instrumentation/github.com/gomodule/redigo/whatapredigo"
	"github.com/whatap/go-api/trace"
)

type HTMLData struct {
	Title   string
	Content string
	//HTMLContent template.HTML
}

func main() {
	udpPortPtr := flag.Int("up", 6600, "agent port(udp). defalt 6600 ")
	portPtr := flag.Int("p", 8080, "web port. default 8080  ")
	dataSourcePtr := flag.String("ds", "phpdemo3:6379", " dataSourceName ")
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

	templatePath := "templates/github.com/gomodule/index.html"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tp, err := template.ParseFiles(templatePath)
		if err != nil {
			fmt.Println("Template not loaded, ", err)
			return
		}

		data := &HTMLData{}
		data.Title = "Redigo Test Page"
		data.Content = r.RequestURI

		tp.Execute(w, data)
	})

	//Case 1. Pool Used
	servicePool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		DialContext: func(ctx context.Context) (redis.Conn, error) {
			return whatapredigo.DialContext(ctx, "tcp", dataSource)
		},
	}
	defer servicePool.Close()

	http.HandleFunc("/SetAndGetWithPool", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		conn, err := servicePool.GetContext(ctx)
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}
		defer conn.Close()

		_, err = conn.Do("SET", "DataKey", "DataValue")
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		data, err := redis.Bytes(conn.Do("GET", "DataKey"))
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		fmt.Println(string(data))
	})

	//Case 2. Dial Used
	http.HandleFunc("/SetAndGetWithDial", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		conn, err := whatapredigo.Dial("tcp", dataSource)
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}
		defer conn.Close()

		conn = conn.WithContext(ctx)

		_, err = conn.Do("SET", "DataKey", 1)
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		data, err := redis.Bytes(conn.Do("GET", "DataKey"))
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		fmt.Println(string(data))

	})

	//Case 3. DialContext Used
	http.HandleFunc("/SetAndGetWithDialContext", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		conn, err := whatapredigo.DialContext(ctx, "tcp", dataSource)
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}
		defer conn.Close()

		_, err = conn.Do("SET", "DataKey", "DataValue")
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		data, err := redis.Bytes(conn.Do("GET", "DataKey"))
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		fmt.Println(string(data))

	})

	//Case 4. DialURL Used
	http.HandleFunc("/SetAndGetWithDialURL", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		dUrl := fmt.Sprintf("redis://%s", dataSource)
		conn, err := whatapredigo.DialURL(dUrl)
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}
		defer conn.Close()

		conn = conn.WithContext(ctx)

		_, err = conn.Do("SET", "DataKey", "DataValue")
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		data, err := redis.Bytes(conn.Do("GET", "DataKey"))
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		fmt.Println(string(data))

	})

	//Case 5. DialURLContext Used
	http.HandleFunc("/SetAndGetWithDialURLContext", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		dUrl := fmt.Sprintf("redis://%s", dataSource)
		conn, err := whatapredigo.DialURLContext(ctx, dUrl)
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}
		defer conn.Close()

		_, err = conn.Do("SET", "DataKey", "DataValue")
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		data, err := redis.Bytes(conn.Do("GET", "DataKey"))
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		fmt.Println(string(data))

	})

	//Case 6. Dial Used With Timeout
	http.HandleFunc("/SetAndGetWithDialTimeout", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		//Dial Timeout Option
		conn, err := whatapredigo.Dial("tcp", dataSource, redis.DialConnectTimeout(time.Millisecond*1000), redis.DialReadTimeout(time.Millisecond*1000), redis.DialWriteTimeout(time.Millisecond*1000))
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}
		defer conn.Close()

		conn = conn.WithContext(ctx)

		_, err = conn.Do("SET", "DataKey", "DataValue")
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		data, err := redis.Bytes(conn.Do("GET", "DataKey"))
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		fmt.Println(string(data))

	})

	//Case 7. Send / Receive
	http.HandleFunc("/SetAndGetWithDialSendReceive", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		conn, err := whatapredigo.Dial("tcp", dataSource)
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}
		defer conn.Close()

		conn = conn.WithContext(ctx)

		err = conn.Send("SET", "DataKey", "DataValue")
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return

		}
		err = conn.Send("GET", "DataKey")
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		err = conn.Flush()
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}

		conn.Receive()              // SET
		data, err := conn.Receive() // GET
		if err != nil {
			fmt.Println(err)
			trace.Error(ctx, err)
			return
		}
		fmt.Println(data)
	})

	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	fmt.Println(port)
}
