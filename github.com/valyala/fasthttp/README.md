# fasthttp(https://github.com/valyala/fasthttp)

fasthttp 핸들러 함수를 whatapfasthttp.Func()로 wrapping 합니다. 
내부에서 fasthttp request, response 관련된 정보를 수집합니다. 

```
import(
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	"github.com/whatap/go-api/httpc"
	"github.com/whatap/go-api/instrumentation/github.com/valyala/fasthttp/whatapfasthttp"
)

func main(){
	
	r := router.New()
	// set whatap 
	r.GET("/", whatapfasthttp.Func(func(ctx *fasthttp.RequestCtx) {
		...
	}
}

```
