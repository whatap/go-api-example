GO=/usr/local/go/bin/go
OUTPUT=../../bin/app

all: mod_tidy mod_download build test

mod_tidy:
	$(GO) mod tidy

mod_download:
	$(GO) mod download -x

test:   #compile warning 제거후 사용 가능
	$(GO) test ./... -cover

build:
	$(GO) build -o $(OUTPUT)/http_client client/client.go
	$(GO) build -o $(OUTPUT)/http_server server/server.go
	

clean :
	rm -f $(OUTPUT)/http_client
	rm -f $(OUTPUT)/http_server

clean_go:
	$(GO) clean -modcache
	$(GO) clean -testcache
	$(GO) clean -cache
	$(GO) clean 

