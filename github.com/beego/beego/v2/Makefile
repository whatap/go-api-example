GO=go
BIN=../../../../../bin
OUTPUT=$(BIN)/app
WHATAP_CONF=$(BIN)/whatap.conf

export GOTOOLCHAIN=auto

all: mod_tidy mod_download build

mod_tidy:
	$(GO) mod tidy

mod_download:
	$(GO) mod download -x

build:
	$(GO) build -o $(OUTPUT)/beegov2/beegov2 beego.go

# Build against the in-repo go-api / golib sources (local development).
local:
	cp go-local.mod go.mod
	$(GO) mod tidy
	$(GO) build -o $(OUTPUT)/beegov2/beegov2 beego.go

copy_env:
	cp -R $(WHATAP_CONF) $(OUTPUT)/beegov2/

upgrade:
	$(GO) get github.com/whatap/go-api@latest
	$(GO) get github.com/whatap/golib@latest

upgrade_go_api:
	$(GO) get github.com/whatap/go-api@latest

upgrade_golib:
	$(GO) get github.com/whatap/golib@latest

clean:
	rm -rf $(OUTPUT)/beegov2

clean_go:
	$(GO) clean -modcache
	$(GO) clean -testcache
	$(GO) clean -cache
	$(GO) clean
