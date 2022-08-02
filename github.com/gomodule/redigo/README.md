# Redigo(https://github.com/gomodule/redigo)
redigo 프레임워크를 통해 redis에 전달되는 명령을 추적합니다.
redis.Dial 대신에 whatapredigo.DialContext를 함수를 사용합니다.
전달하는 context는 내부에 whatap TraceCtx를 포함해야 합니다.
trace.Start()를 통해 TraceCtx는 생성됩니다.

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
