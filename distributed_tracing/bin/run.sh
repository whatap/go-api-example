nohup ./mtrace -p 8080 -api http://localhost:8081 -d 0 > mtrace_8080.log &

nohup ./mtrace -p 8081 -api http://localhost:8082 -d 1 > mtrace_8081.log &

nohup ./mtrace -p 8082 -api http://localhost:8083 -d 2 > mtrace_8082.log &

nohup ./mtrace -p 8083 -api http://localhost:8084 -d 3 > mtrace_8083.log &
