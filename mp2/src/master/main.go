package master

import (
	_ "errors"
	"sync"
	"time"
	"logger"
	"cli"
	"networking"
	"constant"
	_ "fmt"
)


type Vm2fileMap map[string] []string

type File2VmMap map[string] []string


type ClientMembershipList map[string] int64
// map [client ip] last active time

var (
	clientMembershipList ClientMembershipList
	muxClientMembershipList sync.Mutex
)

func readUDPMessageClient2Master(message []byte) error {
	remoteMessage, err := networking.DecodeUDPMessageClient2Master(message)
	if err != nil {
		// log err
	}
	if remoteMessage.MessageType == "connect" {
		muxClientMembershipList.Lock()
		clientIP := remoteMessage.IP
		logger.LogSimpleInfo("receive connect from " + clientIP)
		cli.Write2Shell("receive connect from " + clientIP)
		if _, ok := clientMembershipList[clientIP]; ok {
			clientMembershipList[clientIP] = time.Now().UnixNano()/1000000
		} else {
			clientMembershipList[clientIP] = time.Now().UnixNano()/1000000
		}
		muxClientMembershipList.Unlock()
		// send ack back
		heartbeat := time.Now().UnixNano()/1000000
		message, _ := networking.EncodeUDPMessageMaster2Client(&constant.UDPMessageMaster2Client{heartbeat, "ACK"})
		logger.LogSimpleInfo("send ack back to " + clientIP)
		cli.Write2Shell("send ack back to " + clientIP)
		networking.UDPsend(clientIP, constant.UDPportMaster2Client, message)
	}
	return nil
}


func detectClientInactive() {
	for {
		muxClientMembershipList.Lock()
		currentTime := time.Now().UnixNano()/1000000
		for i,v := range clientMembershipList {
			// kick out 1min inactive client (actual time 1min20s)
			if  currentTime - v > constant.KickoutInactiveClientPeriod  {
				// send kick out and delete
				logger.LogSimpleInfo("send kickout to " + i)
				cli.Write2Shell("send kickout to "+ i)
				message, _ := networking.EncodeUDPMessageMaster2Client(&constant.UDPMessageMaster2Client{0, "KICKOUT"})
				networking.UDPsend(i, constant.UDPportMaster2Client, message)
				delete(clientMembershipList, i)
			}
		}
		muxClientMembershipList.Unlock()
		// every 20s check it
		time.Sleep(constant.CheckInactiveClientPeriod* time.Millisecond)
	}
}

func sendHeartbeat2Clients() {
	for {
		heartbeat := time.Now().UnixNano()/1000000
		message, _ := networking.EncodeUDPMessageMaster2Client(&constant.UDPMessageMaster2Client{heartbeat, "HEARTBEAT"})
		for i,_ := range clientMembershipList {
			logger.LogSimpleInfo("send heartbeat to " + i)
			cli.Write2Shell("send heartbeat to "+ i)
			networking.UDPsend(i, constant.UDPportMaster2Client, message)
		} 
		time.Sleep(constant.MasterSendHeartbeat2ClientPeriod* time.Millisecond)
	}


}

func Run(cliLevel string) {
	clientMembershipList = ClientMembershipList{}
	go networking.UDPlisten(constant.UDPportClient2Master, readUDPMessageClient2Master)
	go detectClientInactive()
	go sendHeartbeat2Clients()
	cli.Run(cliLevel, "master")
} 
