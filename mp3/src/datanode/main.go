package datanode

import (
	"constant"
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

func Run() {
	ServerRun(constant.DatanodeHTTPServerPort)
}
