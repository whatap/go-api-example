# chi(https://github.com/go-chi/chi)

chi 프레임워크의 웹 트랜잭션을 추적합니다.

Use함수를 통해 미들웨어를 설정합니다.

```
r := chi.NewRouter()
r.Use(middleware.Logger)
r.Use(whatapchi.Middleware)
```


```
import (
    "github.com/go-chi/chi/v5"
    "github.com/whatap/go-api/trace"
    "github.com/whatap/go-api/instrumentation/github.com/go-chi/chi/whatapchi"
)		

func main() {
    config := make(map[string]string)
    trace.Init(config)
    defer trace.Shutdown()
    
    r := chi.NewRouter()
    r.Use(whatapchi.Middleware)
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("Response -", r.Response)
    })
}
```
