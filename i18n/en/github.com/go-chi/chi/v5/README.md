# chi(https://github.com/go-chi/chi)

Web transactions are traced in the chi framework.

It sets the middleware through the Use function.

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
