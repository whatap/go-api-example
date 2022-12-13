module github.com/whatap/go-api-example/go-gorm/gorm

go 1.18

require (
	github.com/whatap/go-api v0.1.14
	gorm.io/driver/mysql v1.4.4
	gorm.io/driver/sqlite v1.4.3
	gorm.io/gorm v1.24.2
)

require (
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.15 // indirect
	github.com/whatap/golib v0.0.10 // indirect
	golang.org/x/text v0.3.7 // indirect
)

replace github.com/whatap/go-api => ../../../../go-api

replace github.com/whatap/golib => ../../../../golib
