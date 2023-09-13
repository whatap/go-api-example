HOME=/home/whatap/go-api-example
UDP_PORT=6600
DATA_SOURCE='doremimaker:doremimaker@tcp(go_api_example_db:33063)/doremimaker'
DATA_SOURCE_MYSQL='doremimaker:doremimaker@tcp(go_api_example_db:33063)/doremimaker'
DATA_SOURCE_MSSQL='sqlserver://NewUser:plokijuh!@21@go_api_example_db?database=bbs&encrypt:disable'
DATA_SOURCE_PGSQL='host=go_api_example_db port=5432 user=bbs password=bbs dbname=bbs sslmode=disable'
DATA_SOURCE_REDIS='go_api_example_db:6379'
DATA_SOURCE_KAFKA='go_api_example_db:9092'
IS_WHATAP=


CHECK_OS="`cat /etc/*-release`"
THIS_OS="Linux"

if [[ "$CHECK_OS" == *"CentOS"* ]]; then
    THIS_OS="CentOS"
elif [[ "$CHECK_OS" == *"Alpine"* ]]; then
    THIS_OS="ALPINE"
elif [[ "$CHECK_OS" == *"Ubuntu"* ]]; then
    THIS_OS="Ubuntu"
fi


echo "this os ${THIS_OS}"

start(){
    set -e
    
    start_http_server

	 start_grpc

    start_gin
    
    start_gorilla
    
    start_echo
    
    start_echov4
    
    start_sql
    start_mysql
    start_mssql
    start_pgsql
    
    start_chi
    start_gormv1
    start_gormv2
    
    start_redigo
    start_sarama
    start_chiv5
    start_fasthttp
    start_fiberv2
    
#    start_awsv2
    start_kuber
#    start_mongo
    
}
start_http_server() {
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_MYSQL}
   fi 
    APP_HOME=./app/http_server
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/http_server -p 8080 -up ${UDP_PORT} -ds ${DATA_SOURCE} > ${APP_HOME}/http_server.log & 
    echo $! >> run.pid

    APP_HOME=./app/http_server1
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/http_server -p 8081 -up ${UDP_PORT} -ds ${DATA_SOURCE} > ${APP_HOME}/http_server.log &
    echo $! >> run.pid
}

start_grpc() {
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_MYSQL}
   fi 
    APP_HOME=./app/grpc_server
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/grpc_server  -p 8085 -up ${UDP_PORT} > ${APP_HOME}/grpc_server.log &
    echo $! >> run.pid
    
    APP_HOME=./app/grpc_server1
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/grpc_server -gh localhost -gp 8085 -p 8084 -up ${UDP_PORT} -use_client > ${APP_HOME}/grpc_server.log &
    echo $! >> run.pid
    
    APP_HOME=./app/grpc_server2
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/grpc_server  -gh localhost -gp 8084 -p 8082 -up ${UDP_PORT} -use_client > ${APP_HOME}/grpc_server.log &
    echo $! >> run.pid

    APP_HOME=./app/grpc_client
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/grpc_client  -gh localhost -gp 8082 -up ${UDP_PORT} > ${APP_HOME}/grpc_client.log  &
    echo $! >> run.pid
}

start_gin(){
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_MYSQL}
   fi 
	 APP_HOME=./app/gin
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/gin  -p 8086 -up ${UDP_PORT} -ds ${DATA_SOURCE} > ${APP_HOME}/gin.log &
    echo $! >> run.pid
    
}
start_gorilla(){
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_MYSQL}
   fi 
	 APP_HOME=./app/gorilla
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/gorilla  -p 8087 -up ${UDP_PORT}  -ds ${DATA_SOURCE} > ${APP_HOME}/gorilla.log &
    echo $! >> run.pid
    
}
start_echo(){
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_MYSQL}
   fi 
    APP_HOME=./app/echo
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/echo ${IS_WHATAP} -p 8088 -up ${UDP_PORT} -ds ${DATA_SOURCE} > ${APP_HOME}/echo.log &
    echo $! >> run.pid
}
start_echov4(){
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_MYSQL}
   fi 
    APP_HOME=./app/echov4
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/echov4 ${IS_WHATAP} -p 8089 -up ${UDP_PORT} -ds ${DATA_SOURCE} > ${APP_HOME}/echov4.log &
    echo $! >> run.pid
}

start_sql(){
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_MYSQL}
   fi 
	
    APP_HOME=./app/sql
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/sql  -p 8090 -up ${UDP_PORT} -ds ${DATA_SOURCE} > ${APP_HOME}/sql.log &
    echo $! >> run.pid
}
start_mysql(){
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_MYSQL}
   fi 
    APP_HOME=./app/mysql
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/mysql  -p 8091 -up ${UDP_PORT} -ds ${DATA_SOURCE} > ${APP_HOME}/mysql.log &
    echo $! >> run.pid
}
start_mssql(){
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_MSSQL}
   fi 
    APP_HOME=./app/mssql
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/mssql  -p 8092 -up ${UDP_PORT} -ds ${DATA_SOURCE} > ${APP_HOME}/mssql.log &
    echo $! >> run.pid
}
start_pgsql(){
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_PGSQL}
   fi 
    APP_HOME=./app/pgsql
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/pgsql  -p 8093 -up ${UDP_PORT} -ds "${DATA_SOURCE}" > ${APP_HOME}/pgsql.log &
    echo $! >> run.pid	
}

start_chi(){
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_MYSQL}
   fi 
	APP_HOME=./app/chi
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/chi  -p 8094 -up ${UDP_PORT} -ds ${DATA_SOURCE} > ${APP_HOME}/chi.log &
    echo $! >> run.pid
}
start_gormv1(){
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_MYSQL}
   fi 
	 APP_HOME=./app/gormv1
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/gormv1  -p 8095 -up ${UDP_PORT} -ds ${DATA_SOURCE} > ${APP_HOME}/gormv1.log &
    echo $! >> run.pid
}
start_gormv2(){
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_MYSQL}
   fi 
    APP_HOME=./app/gormv2
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/gormv2 -p 8096 -up ${UDP_PORT} -ds ${DATA_SOURCE} > ${APP_HOME}/gormv2.log &
    echo $! >> run.pid
}

start_redigo(){
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_REDIS}
   fi 
	APP_HOME=./app/redigo
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/redigo  -p 8097 -up ${UDP_PORT} -ds ${DATA_SOURCE} > ${APP_HOME}/redigo.log &
    echo $! >> run.pid
}
start_sarama(){ 
if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_KAFKA}
   fi 
    APP_HOME=./app/sarama
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/sarama  -p 8098 -up ${UDP_PORT} -ds ${DATA_SOURCE} > ${APP_HOME}/sarama.log &
    echo $! >> run.pid
}

start_chiv5(){
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_MYSQL}
   fi 
	APP_HOME=./app/chiv5
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/chiv5  -p 8099 -up ${UDP_PORT} -ds ${DATA_SOURCE} > ${APP_HOME}/chiv5.log &
    echo $! >> run.pid
}
start_fasthttp(){
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_MYSQL}
   fi 
	 APP_HOME=./app/fasthttp
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/fasthttp  -p 8100 -up ${UDP_PORT} -ds ${DATA_SOURCE} > ${APP_HOME}/fasthttp.log &
    echo $! >> run.pid
}
start_fiberv2(){
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_MYSQL}
   fi 
	 APP_HOME=./app/fiberv2
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/fiberv2  -p 8101 -up ${UDP_PORT} -ds ${DATA_SOURCE} > ${APP_HOME}/fiberv2.log &
    echo $! >> run.pid
}
start_awsv2(){
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_MYSQL}
   fi 
	 APP_HOME=./app/awsv2
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/awsv2  -p 8102 -up ${UDP_PORT} -ds ${DATA_SOURCE} > ${APP_HOME}/awsv2.log &
    echo $! >> run.pid
}
start_kuber(){
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_MYSQL}
   fi 
    APP_HOME=./app/kuber
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/kuber  -p 8103 -up ${UDP_PORT} -ds ${DATA_SOURCE} > ${APP_HOME}/kuber.log &
    echo $! >> run.pid
}
start_mongo(){
	if [ "$DATA_SOURCE" != "" ]; then
     	DATA_SOURCE=${DATA_SOURCE_MYSQL}
   fi 
    APP_HOME=./app/mongo
    WHATAP_HOME=${APP_HOME} nohup ${APP_HOME}/mongo  -p 8104 -up ${UDP_PORT} -ds ${DATA_SOURCE_MONGGO} > ${APP_HOME}/mongo.log &
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
stop_stress(){
	pkill httpClient
}
usage(){
    echo ""
    echo "    start "
    echo "    stop "
    echo "    stress "
    echo "    stop_tress "
    echo "    start_http_server [ipc_port] [whatap_flag '-whatap'] "
    echo "    start_grpc [ipc_port] [whatap_flag '-whatap'] "
    echo "    start_gin [ipc_port] [whatap_flag '-whatap'] "
    echo "    start_gorilla [ipc_port] [whatap_flag '-whatap'] "
    echo "    start_echo [ipc_port] [whatap_flag '-whatap'] "
    echo "    start_echov4 [ipc_port] [whatap_flag '-whatap'] "
    echo "    start_sql [ipc_port] [whatap_flag '-whatap'] "
    echo "    start_mysql [ipc_port] [whatap_flag '-whatap'] "
    echo "    start_mssql [ipc_port] [whatap_flag '-whatap'] "
    echo "    start_pgsql [ipc_port] [whatap_flag '-whatap'] "
    echo "    start_chi [ipc_port] [whatap_flag '-whatap'] "
    echo "    start_gormv1 [ipc_port] [whatap_flag '-whatap'] "
    echo "    start_gormv2 [ipc_port] [whatap_flag '-whatap'] "
    echo "    start_redigo [ipc_port] [whatap_flag '-whatap'] "
    echo "    start_sarama [ipc_port] [whatap_flag '-whatap'] "
    echo "    start_chiv5 [ipc_port] [whatap_flag '-whatap'] "
    echo "    start_fasthttp [ipc_port] [whatap_flag '-whatap'] "
    echo "    start_fiberv5 [ipc_port] [whatap_flag '-whatap'] "
    #echo "    start_awsv2 [ipc_port] [whatap_flag '-whatap'] "
    echo "    start_kuber [ipc_port] [whatap_flag '-whatap'] "
    #echo "    start_mongo [ipc_port] [whatap_flag '-whatap'] "
}
set_param(){
	if [ "$1" != "" ]; then
     	UDP_PORT=$1
   fi 
	if [ "$2" != "" ]; then
     	IS_WHATAP=$2
   fi 
   if [ "$3" != "" ]; then
     	DATA_SOURCE=$3
   fi 
}

case "$1" in
    start)
        set_param $2 $3
        stop
        start
        ;;
    stop)
        stop
        ;;
    stress)
        start_stress
        ;;
    stop_stress)
        stop_stress
        ;;
    
    start_http_server)
    	 set_param $2 $3
        stop
        start_http_server
        ;;    
   	 start_grpc)
    	 set_param $2 $3
        stop
        start_grpc
        ;;    
    start_gin)
    	 set_param $2 $3
        stop
        start_gin
        ;;    
    start_gorilla)
    	 set_param $2 $3
        stop
        start_gorilla
        ;;    
    start_echo)
    	 set_param $2 $3
        stop
        start_echo
        ;;    
    start_echov4)
    	 set_param $2 $3
        stop
        start_echov4
        ;;    
    start_sql)
    	 set_param $2 $3
        stop
        start_sql
        ;;    
    start_mysql)
    	 set_param $2 $3
        stop
        start_mysql
        ;;    
    start_mssql)
    	 set_param $2 $3
        stop
        start_mssql
        ;;    
    start_pgsql)
    	 set_param $2 $3
        stop
        start_pgsql
        ;;    
    start_chi)
    	 set_param $2 $3
        stop
        start_chi
        ;;    
    start_gormv1)
    	 set_param $2 $3
        stop
        start_gormv1
        ;;    
    start_gormv2)
    	 set_param $2 $3
        stop
        start_gormv2
        ;;    
    start_redigo)
    	 set_param $2 $3
        stop
        start_redigo
        ;;    
    start_sarama)
    	 set_param $2 $3
        stop
        start_saram
        ;;    
    start_chiv5)
    	 set_param $2 $3
        stop
        start_chiv5
        ;;    
    start_fasthttp)
    	 set_param $2 $3
        stop
        start_fasthttp
        ;;    
    start_fiberv5)
    	 set_param $2 $3
        stop
        start_fiberv5
        ;;    
    start_awsv2)
    	 #set_param $2 $3
        #stop
        #start_awsv2
        ;;    
    start_kuber)
    	 set_param $2 $3
        stop
        start_kuber
        ;;    
    start_mongo)
    	 #set_param $2 $3
        #stop
        #start_mongo
        ;;    
    *)
        usage 
        exit 1
        ;;
esac

exit 0

