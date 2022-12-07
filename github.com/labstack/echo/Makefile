GO=/usr/local/go/bin/go
OUTPUT=../../../bin/app

all: mod_tidy mod_download build test

mod_tidy:
	$(GO) mod tidy

mod_download:
	$(GO) mod download -x

test:   #compile warning 제거후 사용 가능
	$(GO) test ./... -cover

build:
	$(GO) build -o $(OUTPUT)/echo echo.go
	$(GO) build -o $(OUTPUT)/echov4 v4/echo.go

clean :
	rm -f $(OUTPUT)/echo
	rm -f $(OUTPUT)/echov4

clean_go:
	$(GO) clean -modcache
	$(GO) clean -testcache
	$(GO) clean -cache
	$(GO) clean
