package datanode

import (
	_ "fmt"
	"cli"
	"logger"
	"time"
	"networking"
	"constant"
)

func sendHeartbeat2Master() {
	for {
		heartbeat := time.Now().UnixNano()/1000000
		message, _ := networking.EncodeUDPMessageDatanode2Master(&constant.UDPMessageDatanode2Master{constant.LocalIP ,heartbeat, "HEARTBEAT"})
		logger.LogSimpleInfo("send heartbeat to master" + constant.MasterIP)
		cli.Write2Shell(history, "send heartbeat to master"+ constant.MasterIP)
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