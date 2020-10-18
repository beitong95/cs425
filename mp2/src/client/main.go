package client

import (
	_ "fmt"
	_ "os"
	"sync"
	"time"
	"constant"
	"cli"
	"networking"
	"logger"
)

type MasterMembershipList struct {
	Heartbeat int64
}

var (
	masterMembershipList MasterMembershipList 
	isConnected bool
	muxMasterMembershipList sync.Mutex
	client2MasterMessageUDP constant.UDPMessageClient2Master
)

func readUDPMessageMaster2Client(message []byte) error {
	// only heartbeat from master
	remoteMessage, err := networking.DecodeUDPMessageMaster2Client(message)
	if err != nil {
		//log err
	}

	muxMasterMembershipList.Lock()
	newHeartbeat := remoteMessage.Heartbeat
	if newHeartbeat > masterMembershipList.Heartbeat {
		masterMembershipList.Heartbeat = newHeartbeat
	}
	muxMasterMembershipList.Unlock()

	if remoteMessage.MessageType == "ACK" {
		// this message is the ack to connect request
		isConnected = true 
		// log success connect to master
		cli.Write2Shell("Successfully connect to master")
		logger.LogSimpleInfo("Successfully connect to master")			
	} else if remoteMessage.MessageType == "KICKOUT" {
		isConnected = false
		cli.Write2Shell("You are kicked out because of inactive")
		logger.LogSimpleInfo("You are kicked out because of inactive")	
	} 
	return nil
}

func detectMasterFail() {
	for {
		if isConnected == true {
			muxMasterMembershipList.Lock()
			diff := time.Now().UnixNano()/1000000 - masterMembershipList.Heartbeat
			muxMasterMembershipList.Unlock()
			
			if diff > constant.MasterTimeout {
				cli.Write2Shell("detect master fail")
				logger.LogSimpleInfo("detect master fail")
				isConnected = false
				break
			}
		}
		time.Sleep(constant.ClientFailDetectPeriod * time.Millisecond)
	}
}

func connectMaster() {
	connectCount := 0
	for {
		if isConnected == false {
			if connectCount == 0 {
				cli.Write2Shell("Send connect request to master")
				logger.LogSimpleInfo("Send connect request to master")
				// log send connect request to master
			} else {
				cli.Write2Shell("Connect request Fail. Resend connect request to master")
				logger.LogSimpleInfo("Connect request Fail. Resend connect request to master")
				// log connect fail resend connect request
			}
			connectCount += 1
			message, _ := networking.EncodeUDPMessageClient2Master(&client2MasterMessageUDP)
			networking.UDPsend(constant.MasterIP, constant.UDPportClient2Master, message)
			// TODO: using TCP and detect error? 
		} else if connectCount != 0 {
			connectCount = 0
		}
		time.Sleep(constant.ReconnectPeriod * time.Millisecond)
	}
}

/** TODO:

func readFileFromDatanode(filename string, ip string) {
	networking.GetFTP(filename, ip)
}

func getFile(filename string, masterIp string) {
	IPs, err := getDestnationFromMaster(filename, masterIP)
	for i, v := range IPs {
		file, err := readFileFromDatanode(filename, ip)
		if err == nil {
			break 
		}
		if i == len(IPs) - 1 {
			//fatal error 
		}
	}
	//store the file into local location 
	//compare two files if the get file exist and output prompt
	 
}


func putFile(filename string, masterIP string, action string) {
	switch action {
	case "update": updateFile(filename, masterIP)
	case "delete": deleteFile(filename, masterIP)
	}

}

func updateFile(filename string, masterIP string) {
	IPs, err := getDestnationFromMaster(filename, masterIP)
	for _,v := range IPs {
		err := networking.FTPsend(filename, v) 
	}
	// wait for master's ACK
	
}

func deleteFile(filename string, masterIP string) {
	IPs, err := getDestnationFromMaster(filename, masterIP)
	for _,v := range IPs {
		err := networking.FTPsend(filename, v) 
	}
	// wait for master's ACK	
}

**/

func Run(cliLevel string) {
	// initialize
	masterMembershipList = MasterMembershipList{}
	masterMembershipList.Heartbeat = 0
	clientIP, _ := networking.GetLocalIP()
	logger.LogSimpleInfo(clientIP)
	client2MasterMessageUDP = constant.UDPMessageClient2Master{clientIP, "connect"}
	isConnected = false
	// try to connect to master, 
	go networking.UDPlisten(constant.UDPportMaster2Client, readUDPMessageMaster2Client)
	go connectMaster()
	go detectMasterFail()
	cli.Run(cliLevel, "client")
} 