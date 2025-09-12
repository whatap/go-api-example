# Redigo(https://github.com/gomodule/redigo)

It traces the commands delivered to redis through the redigo framework.
The whatapredigo.DialContext function is used instead of redis.Dial.
The context to deliver must include the whatap TraceCtx inside.
TreaceCtx is created through trace.Start().

```

import (
	"context"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/whatap/go-api/instrumentation/github.com/gomodule/redigo/whatapredigo"
	"github.com/whatap/go-api/trace"
)

func main() {
	http.HandleFunc("/SetAndGetWithDialContext", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		conn, err := whatapredigo.DialContext(ctx, "tcp", "192.168.200.65:6379")
		if err != nil {
			trace.Error(ctx, err)
			return
		}
		defer conn.Close()

		_, err = conn.Do("SET", "DataKey", "DataValue")
		if err != nil {
			trace.Error(ctx, err)
			return
		}

		data, err := redis.Bytes(conn.Do("GET", "DataKey"))
		if err != nil {
			trace.Error(ctx, err)
			return
		}


	})

}

```
