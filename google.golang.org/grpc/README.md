```
import (
    "github.com/whatap/go-api/instrumentation/google.golang.org/grpc/whatapgrpc"
)


func main() {
    ...
    // client
    // Set the whatap interceptor to grpc
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", *grpcHost, *grpcPort),
	    grpc.WithInsecure(),
	    grpc.WithBlock(),
	    grpc.WithUnaryInterceptor(whatapgrpc.UnaryClientInterceptor()),
	    grpc.WithStreamInterceptor(whatapgrpc.StreamClientInterceptor()))
	    
	//-------------------------------------------
	
	// server
	// Set the whatap interceptor to grpc
	grpcServer := grpc.NewServer(
	    grpc.UnaryInterceptor(whatapgrpc.UnaryServerInterceptor()),
	    grpc.StreamInterceptor(whatapgrpc.StreamServerInterceptor()))
	    
	...
}
```

| Option Name                                                                                                            | Default Value | Data Type | Description                                                                                                                                                                                                                                                                                                                                                          |
| ---------------------------------------------------------------------------------------------------------------------- | ------------- | --------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| grpc_profile_enabled                                                         | true          | bool      | It determines whether or not to collect the grpc data.                                                                                                                                                                                                                                                                                               |
| grpc_profile_stream_client_enabled | true          | bool      | It determines whether or not to collect the client stream method data.                                                                                                                                                                                                                                                                               |
| grpc_profile_stream_server_enabled | true          | bool      | It determines whether or not to collect the server stream method data.                                                                                                                                                                                                                                                                               |
| grpc_profile_ignore_method                              | agent         | string    | The specified method is not collected.  Use comma (,) to set multiple items.                                                                                                                                                                                                                                      |
| grpc_profile_stream_method                              | ""            | string    | The specified stream method is configured as a separate transaction.  Use comma (,) to set multiple items.  For long-lasting stream connections, each method call is processed as a standalone transaction.  You can search with full method for hitmap and transaction searches. |
| grpc_profile_stream_identify                            | false         | boolean   | It collects the stream full method as the transaction name and adds prefixes to distinguish between the client and server for the same full method.  (/StreamClient/[full method]", /StreamServer/[full method])          |
