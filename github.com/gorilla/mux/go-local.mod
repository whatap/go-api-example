module github.com/whatap/go-api-example/github.com/gorilla/mux

go 1.18

require (
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gorilla/mux v1.8.0
	github.com/whatap/go-api v0.2.4
)

require (
	github.com/google/uuid v1.1.2 // indirect
	github.com/whatap/golib v0.0.16 // indirect
	golang.org/x/text v0.3.7 // indirect
)

replace github.com/whatap/go-api => ../../../../go-api

replace github.com/whatap/golib => ../../../../golib
