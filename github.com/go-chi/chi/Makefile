GO=go
BIN=../../../bin
OUTPUT=$(BIN)/app
TEMPLATE=$(BIN)/templates
WHATAP_CONF=$(BIN)/whatap.conf

all: mod_tidy mod_download build test copy_env

mod_tidy:
	$(GO) mod tidy

mod_download:
	$(GO) mod download -x

test:   #compile warning 제거후 사용 가능
	$(GO) test ./... -cover

build:
	$(GO) build -o $(OUTPUT)/chi/chi chi.go
	$(GO) build -o $(OUTPUT)/chiv5/chiv5 v5/chi.go
		
copy_env:
	cp -R $(TEMPLATE) $(OUTPUT)/chi/
	cp -R $(WHATAP_CONF) $(OUTPUT)/chi/
	
	cp -R $(TEMPLATE) $(OUTPUT)/chiv5/
	cp -R $(WHATAP_CONF) $(OUTPUT)/chiv5/

upgrade:
	$(GO) get github.com/whatap/go-api@latest
	$(GO) get github.com/whatap/golib@latest

upgrade_go_api:
	$(GO) get github.com/whatap/go-api@latest

upgrade_golib:
	$(GO) get github.com/whatap/golib@latest
		
clean :
	rm -rf $(OUTPUT)/chi
	rm -rf $(OUTPUT)/chiv5

clean_go:
	$(GO) clean -modcache
	$(GO) clean -testcache
	$(GO) clean -cache
	$(GO) clean
