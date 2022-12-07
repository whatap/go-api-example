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
	$(GO) build -o $(OUTPUT)/sql sql.go
	$(GO) build -o $(OUTPUT)/mssql mssql/mssql.go
	$(GO) build -o $(OUTPUT)/mysql mysql/mysql.go
	$(GO) build -o $(OUTPUT)/pgsql pgsql/pgsql.go

clean :
	rm -f $(OUTPUT)/sql
	rm -f $(OUTPUT)/mssql
	rm -f $(OUTPUT)/mysql
	rm -f $(OUTPUT)/pgsql

clean_go:
	$(GO) clean -modcache
	$(GO) clean -testcache
	$(GO) clean -cache
	$(GO) clean
