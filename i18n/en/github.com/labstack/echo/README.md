# github.com/labstack/echo(https://github.com/labstack/echo)

Web transactions are traced in the echo framework.\
It sets the middleware through the Use function.

```
    e := echo.New()
    
    // Set the whatap
    e.Use(whatapecho.Middleware())
```

```
import (
    "github.com/labstack/echo"
    
    "github.com/whatap/go-api/trace"
    "github.com/whatap/go-api/instrumentation/github.com/labstack/echo/whatapecho"
)


func main() {

    config := make(map[string]string)
    trace.Init(config)
    defer trace.Shutdown()
    
    ...
    
    e := echo.New()
    
    // Set the whatap
    e.Use(whatapecho.Middleware())
    
    e.GET("/", func(c echo.Context) error {
        return c.String(http.StatusOK, "Hello, World!\n")
    })
}
```
