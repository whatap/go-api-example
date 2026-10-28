GO=go
BIN=../../../../bin
OUTPUT=$(BIN)/app

export GOTOOLCHAIN=auto

all: mod_tidy mod_download build

mod_tidy:
	$(GO) mod tidy

mod_download:
	$(GO) mod download -x

build:
	$(GO) build -o $(OUTPUT)/httprouter/httprouter router.go

local:
	cp go-local.mod go.mod
	$(GO) mod tidy
	$(GO) build -o $(OUTPUT)/httprouter/httprouter router.go

upgrade:
	$(GO) get github.com/whatap/go-api@latest
	$(GO) get github.com/whatap/golib@latest

clean:
	rm -rf $(OUTPUT)/httprouter

clean_go:
	$(GO) clean -modcache
	$(GO) clean -testcache
	$(GO) clean -cache
	$(GO) clean
