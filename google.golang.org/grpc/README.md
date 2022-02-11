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

|옵션명| 기본값| 테이터타입| 설명|
|----|----|----|----|
|grpc_profile_enabled| true| bool| grpc 정보 수집여부를 설정합니다.
|grpc_profile_stream_client_enabled|true|bool|client stream method 정보 수집여부를 설정합니다.|
|grpc_profile_stream_server_enabled|true|bool|server stream method 정보 수집여부를 설정합니다.|
|grpc_profile_ignore_method|agent|string|지정된 method를 수집하지 않습니다.  여러개를 지정할 경우 콤마(,)로 구분합니다.|
}grpc_profile_stream_method|""|string|지정된 stream method를 별도의 트랜잭션으로 구성합니다.  여러개를 지정할 경우 콤마(,)로 구분합니다.  오랜시간 지속하는 stream 연결에 대해서 각각의 method 호출을 단독 트랜잭션으로 처리합니다.  히트맵, 트랜잭션 검색에서 full method 로 검색할 수 있습니다. |
|grpc_profile_stream_identify|false|boolean|Stream full method 를 트랜잭션 이름으로 수집하고,  동일한 full method 에 대해서 client, server를 구분할 수 있는 접두사를 추가합니다.  (/StreamClient/[full method]", /StreamServer/[full method])|