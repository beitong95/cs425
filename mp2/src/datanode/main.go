package datanode

import (
	_ "fmt"
	"cli"
	"logger"
	"time"
	"networking"
	"constant"
)
var (
	localIP string
)

func sendHeartbeat2Master() {
	for {
		heartbeat := time.Now().UnixNano()/1000000
		message, _ := networking.EncodeUDPMessageDatanode2Master(&constant.UDPMessageDatanode2Master{localIP ,heartbeat, "HEARTBEAT"})
		logger.LogSimpleInfo("send heartbeat to master" + constant.MasterIP)
		cli.Write2Shell("send heartbeat to master"+ constant.MasterIP)
		networking.UDPsend(constant.MasterIP, constant.UDPportDatanode2Master, message)
		time.Sleep(constant.DatanodeSendHeartbeat2MasterInterval* time.Millisecond)
	}


}


func Run(cliLevel string) {
	localIP,_ = networking.GetLocalIP()
	go sendHeartbeat2Master()
	cli.Run(cliLevel, "dataNode")
} 