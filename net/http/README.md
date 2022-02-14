#  net/http


## 웹 트랜잭션 추적

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

## RoundTripper 

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
