module github.com/go-api-example/database/sql

go 1.18

require (
	github.com/denisenkom/go-mssqldb v0.12.3
	github.com/go-sql-driver/mysql v1.6.0
	github.com/lib/pq v1.10.7
	github.com/whatap/go-api v0.2.3
)

require (
	github.com/golang-sql/civil v0.0.0-20190719163853-cb61b32ac6fe // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/whatap/golib v0.0.10 // indirect
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/text v0.3.7 // indirect
)

replace github.com/whatap/go-api => ../../../go-api

replace github.com/whatap/golib => ../../../golib
