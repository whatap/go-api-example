# fiber (https://github.com/gofiber/fiber/)

Web transactions are traced in the fiber/v2 framework.
It sets the middleware through the Use function.
The request and response data is collected inside the middleware.

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
