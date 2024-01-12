HOME=/home/whatap/go-api-example
UDP_PORT=6600
DATA_SOURCE='doremimaker:doremimaker@tcp(u20default:33063)/doremimaker'
DATA_SOURCE_MYSQL='doremimaker:doremimaker@tcp(u20default:33063)/doremimaker'
DATA_SOURCE_MSSQL='sqlserver://NewUser:plokijuh!@21@u20default?database=bbs&encrypt:disable'
DATA_SOURCE_PGSQL='host=u20default port=5432 user=bbs password=bbs dbname=bbs sslmode=disable'
DATA_SOURCE_REDIS='u20default:6379'
DATA_SOURCE_KAFKA='u20default:9092'

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
	# stop at error
	set -e
	APP_HOME_ARR=("http_server" "http_server1" "grpc_server" "grpc_server1" "grpc_server2" \
					"grpc_client" "gin" "gorilla" "echo" "echov4" \
					"sql" "mysql" "mssql" "pgsql" "chi" \
					"gormv1" "gormv2" "redigo" "sarama" "chiv5" \
					"fasthttp" "fiberv2" "kuber")
	APP_BIN_ARR=("http_server" "http_server" "grpc_server" "grpc_server" "grpc_server" 
					"grpc_client" "gin" "gorilla" "echo" "echov4" \
					"sql" "mysql" "mssql" "pgsql" "chi" \
					"gormv1" "gormv2" "redigo" "sarama" "chiv5" \
					"fasthttp" "fiberv2" "kuber")
					
	# 8102 : aws , 8014: mongo
	APP_PORT_ARR=("8080" "8081" "8085" "8084" "8082" \
					"8083" "8086" "8087" "8088" "8089" \
					"8090" "8091" "8092" "8093" "8094" \
					"8095" "8096" "8097" "8098" "8099" \
					"8011" "8101" "8013")
					
	APP_PORT_G_ARR=("" "" "8085" "8084" "8082" 
					"" "" "" "" "" \
					"" "" "" "" "" \
					"" "" "" "" "" \
					"" "" "")
	APP_DS_ARR=("${DATA_SOURCE}" "${DATA_SOURCE}" "${DATA_SOURCE}" "${DATA_SOURCE}" "${DATA_SOURCE}" \
					"${DATA_SOURCE}" "${DATA_SOURCE}" "${DATA_SOURCE}" "${DATA_SOURCE}" "${DATA_SOURCE}" \
					"${DATA_SOURCE}" "${DATA_SOURCE}" "${DATA_SOURCE_MSSQL}" "${DATA_SOURCE_PGSQL}" "${DATA_SOURCE}" \
					"${DATA_SOURCE}" "${DATA_SOURCE}" "${DATA_SOURCE_REDIS}" "${DATA_SOURCE_KAFKA}" "${DATA_SOURCE}" \
					"${DATA_SOURCE}" "${DATA_SOURCE}" "${DATA_SOURCE}")

	for index in ${!APP_HOME_ARR[*]}; do
		if [ "${APP_PORT_G_ARR[$index]}" != "" ]; then
			echo "WHATAP_HOME=./app/${APP_HOME_ARR[$index]} \
			nohup ./app/${APP_HOME_ARR[$index]}/${APP_BIN_ARR[$index]} -p ${APP_PORT_ARR[$index]} \
			-gp ${APP_PORT_G_ARR[$index]} -up ${UDP_PORT} > ./app/${APP_HOME_ARR[$index]}/${APP_BIN_ARR[$index]}.log & "
			WHATAP_HOME=./app/${APP_HOME_ARR[$index]} nohup ./app/${APP_HOME_ARR[$index]}/${APP_BIN_ARR[$index]} -p ${APP_PORT_ARR[$index]} \
			-gp ${APP_PORT_G_ARR[$index]} -up ${UDP_PORT} > ./app/${APP_HOME_ARR[$index]}/${APP_BIN_ARR[$index]}.log & 
		else
			echo "WHATAP_HOME=./app/${APP_HOME_ARR[$index]} nohup ./app/${APP_HOME_ARR[$index]}/${APP_BIN_ARR[$index]} -p ${APP_PORT_ARR[$index]} \
			-up ${UDP_PORT} -ds ${APP_DS_ARR[$index]} > ./app/${APP_HOME_ARR[$index]}/${APP_BIN_ARR[$index]}.log & "
			WHATAP_HOME=./app/${APP_HOME_ARR[$index]} nohup ./app/${APP_HOME_ARR[$index]}/${APP_BIN_ARR[$index]} -p ${APP_PORT_ARR[$index]} \
			-up ${UDP_PORT} -ds ${APP_DS_ARR[$index]} > ./app/${APP_HOME_ARR[$index]}/${APP_BIN_ARR[$index]}.log & 
		fi
		echo $! >> run.pid
	done
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
    echo "    start ipc_port"
    echo "    stop "
    echo "    stress "
}
case "$1" in
    start)
        if [ "$2" != "" ]; then
        	UDP_PORT=$2
        fi  
        
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
