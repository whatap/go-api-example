module github.com/whatap/go-api-example/goolg.golang.org/grpc

go 1.18

require (
	github.com/whatap/go-api v0.1.13
	github.com/whatap/go-api-example v0.0.0-20220831075702-c5e596fa1553
	google.golang.org/grpc v1.51.0
	google.golang.org/protobuf v1.28.1
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/whatap/golib v0.0.1 // indirect
	golang.org/x/net v0.0.0-20220722155237-a158d28d115b // indirect
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab // indirect
	golang.org/x/text v0.4.0 // indirect
	google.golang.org/genproto v0.0.0-20200825200019-8632dd797987 // indirect
)

replace github.com/whatap/go-api => ../../../go-api

replace github.com/whatap/golib => ../../../golib
