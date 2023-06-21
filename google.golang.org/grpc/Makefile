GO=/usr/local/go/bin/go
BIN=../../bin
OUTPUT=$(BIN)/app
TEMPLATE=$(BIN)/templates
WHATAP_CONF=$(BIN)/whatap.conf

all: mod_tidy mod_download build test

mod_tidy:
	$(GO) mod tidy

mod_download:
	$(GO) mod download -x

test:   #compile warning 제거후 사용 가능
	$(GO) test ./... -cover

build:
	$(GO) build -o $(OUTPUT)/grpc_client/grpc_client client/client.go
	cp -R $(TEMPLATE) $(OUTPUT)/grpc_client/
	cp -R $(WHATAP_CONF) $(OUTPUT)/grpc_client/
	
	$(GO) build -o $(OUTPUT)/grpc_server/grpc_server server/server.go
	cp -R $(TEMPLATE) $(OUTPUT)/grpc_server/
	cp -R $(WHATAP_CONF) $(OUTPUT)/grpc_server/
	
	cp -R $(OUTPUT)/grpc_server $(OUTPUT)/grpc_server1
	cp -R $(OUTPUT)/grpc_server $(OUTPUT)/grpc_server2
	

clean :
	rm -rf $(OUTPUT)/grpc_client
	rm -rf $(OUTPUT)/grpc_server
	rm -rf $(OUTPUT)/grpc_server1
	rm -rf $(OUTPUT)/grpc_server2

clean_go:
	$(GO) clean -modcache
	$(GO) clean -testcache
	$(GO) clean -cache
	$(GO) clean
