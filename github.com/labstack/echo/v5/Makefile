GO=go
BIN=../../../../bin
OUTPUT=$(BIN)/app
WHATAP_CONF=$(BIN)/whatap.conf

# Echo v5 requires Go 1.25+. GOTOOLCHAIN=auto lets the go command fetch it.
export GOTOOLCHAIN=auto

all: mod_tidy mod_download build

mod_tidy:
	$(GO) mod tidy

mod_download:
	$(GO) mod download -x

build:
	$(GO) build -o $(OUTPUT)/echov5/echov5 echo.go

# Build against the in-repo go-api / golib sources (local development).
local:
	cp go-local.mod go.mod
	$(GO) mod tidy
	$(GO) build -o $(OUTPUT)/echov5/echov5 echo.go

copy_env:
	cp -R $(WHATAP_CONF) $(OUTPUT)/echov5/

upgrade:
	$(GO) get github.com/whatap/go-api@latest
	$(GO) get github.com/whatap/golib@latest

upgrade_go_api:
	$(GO) get github.com/whatap/go-api@latest

upgrade_golib:
	$(GO) get github.com/whatap/golib@latest

clean:
	rm -rf $(OUTPUT)/echov5

clean_go:
	$(GO) clean -modcache
	$(GO) clean -testcache
	$(GO) clean -cache
	$(GO) clean
