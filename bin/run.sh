HOME=/home/whatap/go-api-example
UDP_PORT=6600

CHECK_OS="`cat /etc/*-release`"
THOS_OS="Linux"

if [[ "$CHECK_OS" == *"CentOS"* ]]; then
    THIS_OS="CentOS"
elif [[ "$CHECK_OS" == *"Alpine"* ]]; then
    THIS_OS="ALPINE"
elif [[ "$CHECK_OS" == *"Ubuntu"* ]]; then
    THIS_OS="Ubuntu"
fi


echo "this os ${THIS_OS}"
start(){
    nohup ./app/http_server -p 8080 -up ${UDP_PORT} > ./logs/http_server-8080.log & 
    echo $! >> run.pid
    nohup ./app/http_server -p 8081 -up ${UDP_PORT} > ./logs/http_server-8081.log &
    echo $! >> run.pid

    nohup ./app/grpc_client -gh localhost -gp 8082 -up ${UDP_PORT} > ./logs/grpc_client-8083.log  &
    echo $! >> run.pid

    nohup ./app/grpc_server -p 8085 -up ${UDP_PORT} > ./logs/grpc_server-8085.log &
    echo $! >> run.pid
    nohup ./app/grpc_server -gh localhost -gp 8085 -p 8084 -up ${UDP_PORT} > ./logs/grpc_server-8084.log &
    echo $! >> run.pid
    nohup ./app/grpc_server -gh localhost -gp 8084 -p 8082 -up ${UDP_PORT} > ./logs/grpc_server-8082.log &
    echo $! >> run.pid

    nohup ./app/grpc_client -gh localhost -gp 8082 -up ${UDP_PORT} > ./logs/grpc_server-8083.log  &
    echo $! >> run.pid

    nohup ./app/gin -p 8086 -up ${UDP_PORT} > ./logs/gin.log &
    echo $! >> run.pid
    nohup ./app/mux -p 8087 -up ${UDP_PORT} > ./logs/mux.log &
    echo $! >> run.pid
    nohup ./app/echo -p 8088 -up ${UDP_PORT} > ./logs/echo.log &
    echo $! >> run.pid
    nohup ./app/echo-v4 -p 8089 -up ${UDP_PORT} > ./logs/echo-v4.log &
    echo $! >> run.pid

    nohup ./app/sql -p 8090 -up ${UDP_PORT} > ./logs/sql.log &
    echo $! >> run.pid
    nohup ./app/mysql -p 8091 -up ${UDP_PORT} > ./logs/mysql.log &
    echo $! >> run.pid
    nohup ./app/mssql -p 8092 -up ${UDP_PORT} > ./logs/mssql.log &
    echo $! >> run.pid
    nohup ./app/pgsql -p 8093 -up ${UDP_PORT} > ./logs/pgsql.log &
    echo $! >> run.pid
    nohup ./app/chi -p 8094 -up ${UDP_PORT} > ./logs/chi.log &
    echo $! >> run.pid
    nohup ./app/gormv1 -p 8095 -up ${UDP_PORT} > ./logs/gormv1.log &
    echo $! >> run.pid
    nohup ./app/gormv2 -p 8096 -up ${UDP_PORT} > ./logs/gormv2.log &
    echo $! >> run.pid
    nohup ./app/redigo -p 8097 -up ${UDP_PORT} > ./logs/redigo.log &
    echo $! >> run.pid
    nohup ./app/sarama -p 8098 -up ${UDP_PORT} > ./logs/sarama.log &
    echo $! >> run.pid
    nohup ./app/chiv5 -p 8099 -up ${UDP_PORT} > ./logs/chiv5.log &
    echo $! >> run.pid
}

start_stress(){
    HTTPCLIENT_BIN="./httpClient"
    if [[ "$THIS_OS" == "ALPINE" ]]; then
        HTTPCLIENT_BIN="./httpClient_static"
    fi

    nohup ${HTTPCLIENT_BIN} -c 1 -mc 5 -f ./config_goapi_demo.json > ./logs/httpClient.log &
    echo $! >> run.pid
}

stop(){
    cat run.pid | while read line
    do
        echo "kill -9 ${line}" 
        kill -9 $line
    done
    cat /dev/null > run.pid
}
usage(){
    echo ""
    echo "    start "
    echo "    stop "
    echo "    stress "
}
case "$1" in
    start)
        stop
        start
        ;;
    stop)
        stop
        ;;
    stress)
        start_stress
        ;;
    *)
        usage 
        exit 1
        ;;
esac

exit 0
