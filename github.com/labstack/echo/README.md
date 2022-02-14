# github.com/labstack/echo(https://github.com/labstack/echo)

echo 프레임워크의 웹 트랜잭션을 추적합니다.  
Use 함수를 통해 미들웨어를 설정합니다.

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