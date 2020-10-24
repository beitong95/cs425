package datanode

import (
	"fmt"
	"strconv"
	. "structs"
)

/**
	Finished:
	1. send heartbeat to master

	TODO:
	1. handle rereplica request
	2. handle get from client
	3. handle put from client
	4. send local storage status to master(restore the system status when master fails)
	5. send ACK back to master to close the put\update\delete service loop
**/
var FileList []string

func Run() {
	i, err := strconv.Atoi(MyPort)
	if err != nil {
		panic(err)
	}
	serverPort := fmt.Sprint(int(i) + 1)
	ServerRun(serverPort)
}
