package main

import (
	"flag"

	"github.com/whatap/go-api/trace"

	"github.com/whatap/go-api-example/distributed_tracing/server"
)

func main() {
	portPtr := flag.Int("p", 8080, "web port. default 8080  ")
	apiHostPtr := flag.String("api", "http://localhost:8081/", "http://localhost:8081/")
	depthPtr := flag.Int("d", 0, ", depth. 0 ~")

	flag.Parse()
	port := *portPtr
	apiHost := *apiHostPtr
	depth := *depthPtr

	trace.Init(nil)
	defer trace.Shutdown()

	server := new(server.Server)
	server.Start(port, apiHost, depth)
}
