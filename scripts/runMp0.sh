#! /bin/bash
# kill all processes who is using port 1234
port=1234
pid=$(/usr/sbin/lsof -t -i:$port)
if [ ! -n "$pid" ]
then
	echo "no one is using $port"
else
	echo "the pid is $pid"
	kill -9 $pid
fi
cd /home/beitong2/cs425/mp0
./log_service_wrapper.sh
