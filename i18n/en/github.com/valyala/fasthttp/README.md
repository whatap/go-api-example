# fasthttp(https://github.com/valyala/fasthttp)

It wraps the fasthttp handler function with whatapfasthttp.Func().
It collects the data related to the fasthttp request and response internally.

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
