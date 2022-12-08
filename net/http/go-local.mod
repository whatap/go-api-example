module github.com/whatap/go-api-example/net/http

go 1.18

require (
	github.com/go-sql-driver/mysql v1.6.0
	github.com/whatap/go-api v0.1.13
)

require (
	github.com/google/uuid v1.1.2 // indirect
	github.com/whatap/golib v0.0.1 // indirect
	golang.org/x/text v0.3.7 // indirect
)

replace github.com/whatap/go-api => ../../../go-api

replace github.com/whatap/golib => ../../../golib
