package datanode

import (
	_ "fmt"
	"cli"
	"logger"
	"time"
	"networking"
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

func sendHeartbeat2Master() {
	for {
		heartbeat := time.Now().UnixNano()/1000000
		message, _ := networking.EncodeUDPMessageDatanode2Master(&constant.UDPMessageDatanode2Master{constant.LocalIP ,heartbeat, "HEARTBEAT"})
		logger.LogSimpleInfo("send heartbeat to master " + constant.MasterIP)
		cli.Write2Shell(history, "send heartbeat to master "+ constant.MasterIP)
		networking.UDPsend(constant.MasterIP, constant.UDPportDatanode2Master, message)
		time.Sleep(constant.DatanodeSendHeartbeat2MasterInterval* time.Millisecond)
	}


}


func Run(cliLevel string) {
	go sendHeartbeat2Master()
	if cliLevel == "cli" {
		cliDatanode()
	} else {
		cliSimpleDatanode()
	}
} 