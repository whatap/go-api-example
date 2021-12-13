module github.com/whatap/go-api-example

go 1.14

require (
	github.com/denisenkom/go-mssqldb v0.11.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/godror/godror v0.29.0
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/labstack/echo v3.3.10+incompatible
	github.com/lib/pq v1.10.4
	github.com/whatap/go-api v0.1.6
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20200806141610-86f49bd18e98
	google.golang.org/grpc v1.42.0
	google.golang.org/grpc/examples v0.0.0-20211028161633-d47437c91e79 // indirect
	google.golang.org/protobuf v1.27.1
)

replace github.com/whatap/go-api => ../go-api
