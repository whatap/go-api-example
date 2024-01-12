GO=/usr/local/go/bin/go

all: database_sql http grpc gin gorilla echo chi fasthttp fiber gormv2 gormv1 redigo sarama  kubernetes 
#all: database_sql http grpc gin gorilla echo chi fasthttp fiber gormv2 gormv1 redigo sarama  kubernetes aws mongo
#all: mod_tidy mod_download http database_sql grpc gin gorilla echo gormv2 gormv1 redigo sarama chi fasthttp fiber kubernetes aws mongo

mod_download:
	$(GO) mod download -x

mod_tidy:
	$(GO) mod tidy

database_sql:
	#echo "database/sql"
	make -C database/sql 

http:
	#echo "net/http"
	make -C net/http

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

chi:
	#echo "go-chi/chi"
	make -C github.com/go-chi/chi

fasthttp:
	#echo "valyala/fasthttp"
	make -C github.com/valyala/fasthttp

fiber:
	#echo "goviber/fiber"
	make -C github.com/gofiber/fiber


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
	cp ./net/http/go-local.mod ./net/http/go.mod
	cp ./google.golang.org/grpc/go-local.mod ./google.golang.org/grpc/go.mod

	cp ./github.com/gin-gonic/gin/go-local.mod ./github.com/gin-gonic/gin/go.mod
	cp ./github.com/gorilla/mux/go-local.mod ./github.com/gorilla/mux/go.mod
	cp ./github.com/labstack/echo/go-local.mod ./github.com/labstack/echo/go.mod
	cp ./github.com/go-chi/chi/go-local.mod ./github.com/go-chi/chi/go.mod
	cp ./github.com/valyala/fasthttp/go-local.mod ./github.com/valyala/fasthttp/go.mod
	cp ./github.com/gofiber/fiber/go-local.mod ./github.com/gofiber/fiber/go.mod

	cp ./github.com/go-gorm/gorm/go-local.mod ./github.com/go-gorm/gorm/go.mod
	cp ./github.com/gomodule/redigo/go-local.mod ./github.com/gomodule/redigo/go.mod
	cp ./github.com/jinzhu/gorm/go-local.mod ./github.com/jinzhu/gorm/go.mod
	cp ./github.com/Shopify/sarama/go-local.mod ./github.com/Shopify/sarama/go.mod

	cp ./k8s.io/client-go/kubernetes/go-local.mod ./k8s.io/client-go/kubernetes/go.mod
	#cp ./github.com/aws/go-local.mod ./github.com/aws/go.mod
	#cp ./github.com/mongodb/mongo-go-driver/go-local.mod ./github.com/mongodb/mongo-go-driver/go.mod
	
upgrade:
	make -C database/sql upgrade 
	make -C net/http upgrade
	make -C google.golang.org/grpc upgrade
	make -C github.com/gin-gonic/gin upgrade
	make -C github.com/gorilla/mux upgrade
	make -C github.com/labstack/echo upgrade
	make -C github.com/go-chi/chi upgrade
	make -C github.com/valyala/fasthttp upgrade
	make -C github.com/gofiber/fiber upgrade
	make -C github.com/go-gorm/gorm upgrade
	make -C github.com/gomodule/redigo upgrade 
	make -C github.com/jinzhu/gorm upgrade 
	make -C github.com/Shopify/sarama upgrade
	make -C k8s.io/client-go/kubernetes upgrade
	#make -C github.com/aws upgrade
	#make -C github.com/mongodb/mongo-go-driver upgrade

upgrade_go_api:
	make -C database/sql upgrade_go_api
	make -C net/http upgrade_go_api
	make -C google.golang.org/grpc upgrade_go_api
	make -C github.com/gin-gonic/gin upgrade_go_api
	make -C github.com/gorilla/mux upgrade_go_api
	make -C github.com/labstack/echo upgrade_go_api
	make -C github.com/go-chi/chi upgrade_go_api
	make -C github.com/valyala/fasthttp upgrade_go_api
	make -C github.com/gofiber/fiber upgrade_go_api
	make -C github.com/go-gorm/gorm upgrade_go_api
	make -C github.com/gomodule/redigo upgrade_go_api 
	make -C github.com/jinzhu/gorm upgrade_go_api 
	make -C github.com/Shopify/sarama upgrade_go_api
	make -C k8s.io/client-go/kubernetes upgrade_go_api
	#make -C github.com/aws upgrade_go_api
	#make -C github.com/mongodb/mongo-go-driver upgrade_go_api
	
	
upgrade_golib:
	make -C database/sql upgrade_golib 
	make -C net/http upgrade_golib
	make -C google.golang.org/grpc upgrade_golib
	make -C github.com/gin-gonic/gin upgrade_golib
	make -C github.com/gorilla/mux upgrade_golib
	make -C github.com/labstack/echo upgrade_golib
	make -C github.com/go-chi/chi upgrade_golib
	make -C github.com/valyala/fasthttp upgrade_golib
	make -C github.com/gofiber/fiber upgrade_golib
	make -C github.com/go-gorm/gorm upgrade_golib
	make -C github.com/gomodule/redigo upgrade_golib 
	make -C github.com/jinzhu/gorm upgrade_golib 
	make -C github.com/Shopify/sarama upgrade_golib
	make -C k8s.io/client-go/kubernetes upgrade_golib
	#make -C github.com/aws upgrade_golib
	#make -C github.com/mongodb/mongo-go-driver upgrade_golib
	

clean:
	make -C database/sql clean 
	make -C net/http clean
	make -C google.golang.org/grpc clean
	make -C github.com/gin-gonic/gin clean
	make -C github.com/gorilla/mux clean
	make -C github.com/labstack/echo clean
	make -C github.com/go-chi/chi clean
	make -C github.com/valyala/fasthttp clean
	make -C github.com/gofiber/fiber clean
	make -C github.com/go-gorm/gorm clean
	make -C github.com/gomodule/redigo clean 
	make -C github.com/jinzhu/gorm clean 
	make -C github.com/Shopify/sarama clean
	make -C k8s.io/client-go/kubernetes clean
	#make -C github.com/aws clean
	#make -C github.com/mongodb/mongo-go-driver clean
	

clean_go:
	$(GO) clean -modcache
	$(GO) clean -testcache
	$(GO) clean -cache
	$(GO) clean 


	make -C database/sql clean_go
	make -C net/http clean_go
	make -C google.golang.org/grpc clean_go
	make -C github.com/gin-gonic/gin clean_go
	make -C github.com/gorilla/mux clean_go
	make -C github.com/labstack/echo clean_go
	make -C github.com/go-chi/chi clean_go
	make -C github.com/valyala/fasthttp clean_go
	make -C github.com/gofiber/fiber clean_go
	make -C github.com/go-gorm/gorm clean_go
	make -C github.com/gomodule/redigo clean_go 
	make -C github.com/jinzhu/gorm clean_go 
	make -C github.com/Shopify/sarama clean_go
	make -C k8s.io/client-go/kubernetes clean_go
	#make -C github.com/aws clean_go
	#make -C github.com/mongodb/mongo-go-driver clean_go
	
