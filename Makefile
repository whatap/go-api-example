

all: mod_tidy mod_download http database_sql grpc gin gorilla echo 

mod_download:
	go mod download -x

mod_tidy:
	go mod tidy

database_sql:
	#echo "database/sql"
	go build -o bin/app/sql database/sql/sql.go
	go build -o bin/app/mysql database/sql/mysql/mysql.go
	go build -o bin/app/mssql database/sql/mssql/mssql.go
	go build -o bin/app/pgsql database/sql/pgsql/pgsql.go

grpc:
	#echo "gooogle.golang.org/grpc"
	go build -o bin/app/grpc_client google.golang.org/grpc/client/client.go
	go build -o bin/app/grpc_server google.golang.org/grpc/server/server.go

gin:
	#echo "gin-gonic/gin"
	go build -o bin/app/gin github.com/gin-gonic/gin/gin.go

gorilla:
	#echo "gorilla/mux"
	go build -o bin/app/mux github.com/gorilla/mux/mux.go

echo:
	#echo "labstack/echo"
	go build -o bin/app/echo github.com/labstack/echo/echo.go
	go build -o bin/app/echo-v4 github.com/labstack/echo/v4/echo.go

chi:
	#echo "go-chi/chi"
	go build -o bin/app/chi github.com/go-chi/chi/chi.go

http:
	#echo "net/http"
	go build -o bin/app/http_client net/http/client/client.go
	go build -o bin/app/http_server net/http/server/server.go

clean:
	rm -f bin/app/*
 
