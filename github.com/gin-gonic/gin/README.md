```
import (
    "github.com/go-gonic/gin"
    
    "github.com/whatap/go-api/trace"
    "github.com/whatap/go-api/instrumentation/github.com/go-gonic/gin/whatapgin"
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