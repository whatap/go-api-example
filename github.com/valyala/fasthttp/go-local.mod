module github.com/whatap/go-api-example/github.com/valyala/fasthttp

go 1.18

require (
	github.com/fasthttp/router v1.4.14
	github.com/go-sql-driver/mysql v1.6.0
	github.com/valyala/fasthttp v1.43.0
	github.com/whatap/go-api v0.1.14
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/savsgio/gotils v0.0.0-20220530130905-52f3993e8d6d // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/whatap/golib v0.0.10 // indirect
	golang.org/x/text v0.3.7 // indirect
)

replace github.com/whatap/go-api => ../../../../go-api

replace github.com/whatap/golib => ../../../../golib
