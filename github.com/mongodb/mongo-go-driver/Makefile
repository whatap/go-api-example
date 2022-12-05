GO=/usr/local/go/bin/go
OUTPUT=../../../bin/app

all: mod_tidy mod_download build test

mod_tidy:
	$(GO) mod tidy

mod_download:
	$(GO) mod download -x

test:   #compile warning 제거후 사용 가능
	$(GO) test ./...

build:
	$(GO) build -o $(OUTPUT)/mongo mongo.go

clean :
	rm -f $(OUTPUT)/mongo

clean_go:
	$(GO) clean -modcache
	$(GO) clean -testcache
	$(GO) clean -cache
	$(GO) clean
