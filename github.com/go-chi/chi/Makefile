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
	$(GO) build -o $(OUTPUT)/chi/chi chi.go
	cp -R $(TEMPLATE) $(OUTPUT)/chi/
	cp -R $(WHATAP_CONF) $(OUTPUT)/chi/
	
	$(GO) build -o $(OUTPUT)/chiv5/chiv5 v5/chi.go
	cp -R $(TEMPLATE) $(OUTPUT)/chiv5/
	cp -R $(WHATAP_CONF) $(OUTPUT)/chiv5/
	
clean :
	rm -rf $(OUTPUT)/chi
	rm -rf $(OUTPUT)/chiv5

clean_go:
	$(GO) clean -modcache
	$(GO) clean -testcache
	$(GO) clean -cache
	$(GO) clean
