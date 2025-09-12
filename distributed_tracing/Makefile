all: mod_tidy mod_download build 

mod_tidy:
	go mod tidy

mod_download:
	go mod download -x

build:
	go build -o ./bin/mtrace ./main.go
	
