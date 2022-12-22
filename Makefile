GO=/usr/local/go/bin/go

all: http database_sql grpc gin gorilla echo gormv2 gormv1 redigo sarama chi fasthttp fiber kubernetes
#all: mod_tidy mod_download http database_sql grpc gin gorilla echo redigo sarama chi chiv5 fasthttp fiberv2 mongo awsv2 kubernetes

mod_download:
	$(GO) mod download -x

mod_tidy:
	$(GO) mod tidy

database_sql:
	#echo "database/sql"
	make -C database/sql 

grpc:
	#echo "gooogle.golang.org/grpc"
	make -C google.golang.org/grpc

gin:
	#echo "gin-gonic/gin"
	make -C github.com/gin-gonic/gin

gorilla:
	#echo "gorilla/mux"
	make -C github.com/gorilla/mux

echo:
	#echo "labstack/echo"
	make -C github.com/labstack/echo

http:
	#echo "net/http"
	make -C net/http

gormv2:
	#echo "go-gorm/gorm"
	make -C github.com/go-gorm/gorm

gormv1:
	#echo "jinzhu/gorm"
	make -C github.com/jinzhu/gorm

redigo:
	#echo "gomodule/redigo"
	make -C github.com/gomodule/redigo

sarama:
	#echo "Shopify/sarama"
	make -C github.com/Shopify/sarama

chi:
	#echo "go-chi/chi"
	make -C github.com/go-chi/chi

fasthttp:
	#echo "valyala/fasthttp"
	make -C github.com/valyala/fasthttp

fiber:
	#echo "goviber/fiber"
	make -C github.com/gofiber/fiber

mongo:
	make -C github.com/mongodb/mongo-go-driver
	
aws:
	#echo "aws/aws-sdk-go-v2"
	make -C github.com/aws

kubernetes:
	#echo "k8s.io/client-go/kubernetes"
	make -C k8s.io/client-go/kubernetes
	
local:
	cp ./database/sql/go-local.mod ./database/sql/go.mod
	#cp ./github.com/aws/go-local.mod ./github.com/aws/go.mod
	cp ./github.com/gin-gonic/gin/go-local.mod ./github.com/gin-gonic/gin/go.mod
	cp ./github.com/go-chi/chi/go-local.mod ./github.com/go-chi/chi/go.mod
	cp ./github.com/go-gorm/gorm/go-local.mod ./github.com/go-gorm/gorm/go.mod
	cp ./github.com/gofiber/fiber/go-local.mod ./github.com/gofiber/fiber/go.mod
	cp ./github.com/gomodule/redigo/go-local.mod ./github.com/gomodule/redigo/go.mod
	cp ./github.com/gorilla/mux/go-local.mod ./github.com/gorilla/mux/go.mod
	cp ./github.com/jinzhu/gorm/go-local.mod ./github.com/jinzhu/gorm/go.mod
	cp ./github.com/labstack/echo/go-local.mod ./github.com/labstack/echo/go.mod
	#cp ./github.com/mongodb/mongo-go-driver/go-local.mod ./github.com/mongodb/mongo-go-driver/go.mod
	cp ./github.com/Shopify/sarama/go-local.mod ./github.com/Shopify/sarama/go.mod
	cp ./github.com/valyala/fasthttp/go-local.mod ./github.com/valyala/fasthttp/go.mod
	cp ./google.golang.org/grpc/go-local.mod ./google.golang.org/grpc/go.mod
	cp ./k8s.io/client-go/kubernetes/go-local.mod ./k8s.io/client-go/kubernetes/go.mod
	cp ./net/http/go-local.mod ./net/http/go.mod

	
clean:
	make -C database/sql clean 
	make -C google.golang.org/grpc clean
	make -C github.com/gin-gonic/gin clean
	make -C github.com/gorilla/mux clean
	make -C github.com/labstack/echo clean
	make -C net/http clean
	make -C github.com/go-gorm/gorm clean
	make -C github.com/jinzhu/gorm clean 
	make -C github.com/gomodule/redigo clean 
	make -C github.com/Shopify/sarama clean
	make -C github.com/go-chi/chi clean
	make -C github.com/valyala/fasthttp clean
	make -C github.com/gofiber/fiber clean
	#make -C github.com/mongodb/mongo-go-driver clean
	#make -C github.com/aws clean
	make -C k8s.io/client-go/kubernetes clean

go_clean:
	$(GO) clean -modcache
	$(GO) clean -testcache
	$(GO) clean -cache
	$(GO) clean 

	
