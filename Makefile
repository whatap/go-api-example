

all: http database_sql grpc gin gorilla echo 

database_sql:
	#echo "database/sql"
	go build -o database/sql/mysql/ database/sql/mysql/mysql.go
	go build -o database/sql/mssql/ database/sql/mssql/mssql.go
	go build -o database/sql/pgsql/ database/sql/pgsql/pgsql.go

grpc:
	#echo "gooogle.golang.org/grpc"
	go build -o google.golang.org/grpc/client/ google.golang.org/grpc/client/client.go
	go build -o google.golang.org/grpc/server/ google.golang.org/grpc/server/server.go

gin:
	#echo "gin-gonic/gin"
	go build -o github.com/gin-gonic/gin/ github.com/gin-gonic/gin/gin.go

gorilla:
	#echo "gorilla/mux"
	go build -o github.com/gorilla/mux/ github.com/gorilla/mux/mux.go

echo:
	#echo "labstack/echo"
	go build -o github.com/labstack/echo/ github.com/labstack/echo/echo.go

http:
	#echo "net/http"
	go build -o net/http/client/ net/http/client/client.go
	go build -o net/http/server/ net/http/server/server.go

clean:
	rm -rf database/sql/mysql/mysql
	rm -rf database/sql/mssql/mssql
	rm -rf database/sql/pgsql/pgsql
	rm -rf google.golang.org/grpc/client/client
	rm -rf google.golang.org/grpc/server/server
	rm -rf github.com/gin-gonic/gin/gin
	rm -rf github.com/gorilla/mux/mux
	rm -rf github.com/labstack/echo/echo
	rm -rf net/http/client/client
	rm -rf net/http/server/server

 
