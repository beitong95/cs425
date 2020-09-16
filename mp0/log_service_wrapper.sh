#!/bin/bash

log=logging_service.log
cd /home/beitong2/cs425/mp0/
exec &>$log
echo $(date +"%D %T")" Starting..."
export CONFIG=config.json
( exec ./logly --server &>>$log ) &
exit 0
