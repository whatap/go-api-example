module github.com/whatap/go-api-example/github.com/gofiber/fiber

go 1.18

require (
	github.com/whatap/go-api v0.1.12
	github.com/whatap/go-api/instrumentation/github.com/gofiber/fiber v0.0.0
	github.com/whatap/go-api/instrumentation/github.com/valyala/fasthttp v0.0.0 // indirect
	github.com/whatap/golib v0.0.3 // indirect
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/gofiber/fiber/v2 v2.36.0
	github.com/google/uuid v1.1.2 // indirect
	github.com/klauspost/compress v1.15.6 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.39.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/sys v0.0.0-20220227234510-4e6760a101f9 // indirect
	golang.org/x/text v0.3.7 // indirect
)

require (
	github.com/go-sql-driver/mysql v1.6.0
	github.com/sirupsen/logrus v1.8.1
)

replace (
	github.com/whatap/go-api/instrumentation/github.com/gofiber/fiber v0.0.0 => /home/ubuntu/whatap-go/go-api/instrumentation/github.com/gofiber/fiber
	github.com/whatap/go-api/instrumentation/github.com/valyala/fasthttp v0.0.0 => /home/ubuntu/whatap-go/go-api/instrumentation/github.com/valyala/fasthttp
)
