```
import (
    "github.com/gorilla/mux"
    
    "github.com/whatap/go-api/trace"
    "github.com/whatap/go-api/instrumentation/github.com/gorilla/mux/whatapmux"
)


func main() {

    config := make(map[string]string)
    trace.Init(config)
    defer trace.Shutdown()
    
    r := mux.NewRouter()
    
    // Set the whatap
    r.Use(whatapmux.Middleware())
    
    r.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Add("Content-Type", "text/html")
        reply := "/index <br/>Test Body"
        _, _ = w.Write(([]byte)(reply))
        fmt.Println("Response -", r.Response)
    }
}
```
