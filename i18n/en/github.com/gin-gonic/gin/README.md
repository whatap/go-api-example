# github.com/gin-gonic/gin(https://github.com/gin-gonic/gin)

Web transactions are traced in the gin framework.\
It sets the middleware through the Use function.

```
    r := gin.Default()
    
    // Set the whatap
    r.Use(whatapgin.Middleware())
```

```
import (
    "github.com/gin-gonic/gin"
    
    "github.com/whatap/go-api/trace"
    "github.com/whatap/go-api/instrumentation/github.com/gin-gonic/gin/whatapgin"
)

func main() {
    config := make(map[string]string)
    trace.Init(config)
    defer trace.Shutdown()
    
    r := gin.Default()
    
    // Set the whatap
    r.Use(whatapgin.Middleware())
    
    r.GET("/", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "message": "ok",
        })
    })
}
```
