#  net/http

## Server

### 웹 트랜잭션 추적

```
// wrapping type of http.HanderFunc, example : http.Handle(pattern, http.HandlerFunc)
func HandlerFunc(handler func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conf := config.GetConfig()
		if !conf.TransactionEnabled {
			handler(w, r)
			return
		}
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)
		handler(w, r.WithContext(ctx))
	})
}

// wrapping handler function, example : http.HandleFunc(func(http.ResponseWriter, *http.Request))
func Func(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conf := config.GetConfig()
		if !conf.TransactionEnabled {
			handler(w, r)
			return
		}
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)
		handler(w, r.WithContext(ctx))
	}
}
```


```
import (
	"github.com/whatap/go-api/instrumentation/net/http/whataphttp"
	"github.com/whatap/go-api/trace"
)
	
func main(){	
	config := make(map[string]string)
	trace.Init(config)
	//It must be executed before closing the app.
	defer trace.Shutdown()
	
	http.HandleFunc("/wrapHandleFunc", whataphttp.Func(func(w http.ResponseWriter, r *http.Request) {
		...
	}))

	http.Handle("/wrapHandleFunc1", whataphttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		...
	}))
	
	
	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
```

## Client

httpc.Start(), httpc.End() 함수로 추적할 수 있습니다. 

```
func Start(ctx context.Context, url string) (*HttpcCtx, error)
func End(httpcCtx *HttpcCtx, status int, reason string, err error) error

```

import (
	"github.com/whatap/go-api/httpc"
)

func main(){
    config := make(map[string]string)
	trace.Init(config)
	//It must be executed before closing the app.
	defer trace.Shutdown()
	
	ctx, _ := trace.Start(context.Background(), "Http call")
	defer trace.End(ctx, nil)
    
    callUrl := "http://localhost:8081/index"
	httpcCtx, _ := httpc.Start(ctx, callUrl)
	if resp, err := http.Get(callUrl); err == nil {
		defer resp.Body.Close()
        httpc.End(httpcCtx, resp.StatusCode, "", err)
    } else {
        httpc.End(httpcCtx, resp.StatusCode, "", err)
    }
}
```


### RoundTripper 

RoundTripper 미들웨어를 설정하여 http call을 추적할 수 있습니다. 

전달하는 context는 내부에 whatap TraceCtx를 포함해야 합니다.  
trace.Start()를 통해 TraceCtx는 생성됩니다.

```
import (
	"github.com/whatap/go-api/instrumentation/net/http/whataphttp"
)


func main(){
	config := make(map[string]string)
	trace.Init(config)
	//It must be executed before closing the app.
	defer trace.Shutdown()
	
	ctx, _ := trace.Start(context.Background(), "Http call")
	defer trace.End(ctx, nil)
	
	callUrl = "http://localhost:8081/httpc"
	client := http.DefaultClient
	client.Transport = whataphttp.NewRoundTrip(ctx, http.DefaultTransport)
	resp, err := client.Get(callUrl)	
}
```
