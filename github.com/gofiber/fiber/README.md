# fiber (https://github.com/gofiber/fiber/)


fiber/v2 chi 프레임워크의 웹 트랜잭션을 추적합니다.
Use함수를 통해 미들웨어를 설정합니다.
미들웨어 내부에서 request, response 정보를 수집합니다. 

```
import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/whatap/go-api/instrumentation/github.com/gofiber/fiber/v2/whatapfiber"
)

func main() {
	...
	
	
	r.Use(recover.New())
	// set whatap middleware
	r.Use(whatapfiber.Middleware())

	// app.Get("/", index)
	// app.Get("/panic", panicFunc)
	// app.Get("/selectRows", selectRow)
	// app.Get("/sleepSecond", sleepSecond)

	r.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title": "fiber/v2",
		})
	})
	
	
	...
}

			
```