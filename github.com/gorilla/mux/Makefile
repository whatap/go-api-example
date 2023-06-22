GO=/usr/local/go/bin/go
BIN=../../../bin
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
	$(GO) build -o $(OUTPUT)/gorilla/gorilla mux.go
	cp -R $(TEMPLATE) $(OUTPUT)/gorilla/
	cp -R $(WHATAP_CONF) $(OUTPUT)/gorilla
	
upgrade:
	$(GO) get github.com/whatap/go-api
	
clean :
	rm -rf $(OUTPUT)/gorilla

clean_go:
	$(GO) clean -modcache
	$(GO) clean -testcache
	$(GO) clean -cache
	$(GO) clean
