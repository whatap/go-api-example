# github.com/gin-gonic/gin(https://github.com/gin-gonic/gin)

gin 프레임워크의 웹 트랜잭션을 추적합니다.  
Use 함수를 통해 미들웨어를 설정합니다.

```
    r := gin.Default()
    
    // Set the whatap
    r.Use(whatapgin.Middleware())
```


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