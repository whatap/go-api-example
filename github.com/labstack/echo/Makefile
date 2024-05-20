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
	$(GO) build -o $(OUTPUT)/echo/echo echo.go
	$(GO) build -o $(OUTPUT)/echov4/echov4 v4/echo.go
	
copy_env:
	cp -R $(TEMPLATE) $(OUTPUT)/echo/
	cp -R $(WHATAP_CONF) $(OUTPUT)/echo/
	
	cp -R $(TEMPLATE) $(OUTPUT)/echov4/
	cp -R $(WHATAP_CONF) $(OUTPUT)/echov4/
	
upgrade:
	$(GO) get github.com/whatap/go-api@latest
	$(GO) get github.com/whatap/golib@latest
	
upgrade_go_api:
	$(GO) get github.com/whatap/go-api@latest

upgrade_golib:
	$(GO) get github.com/whatap/golib@latest

	
clean :
	rm -rf $(OUTPUT)/echo
	rm -rf $(OUTPUT)/echov4

clean_go:
	$(GO) clean -modcache
	$(GO) clean -testcache
	$(GO) clean -cache
	$(GO) clean
