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

type masterMembershipList struct {
	Heartbeat int64
}

var (
	_masterMembershipList masterMembershipList 
	isConnected bool
	muxMasterMembershipList sync.Mutex
	client2MasterMessageUDP constant.UDPMessageClient2Master
	isKickout bool
)

func readUDPMessageMaster2Client(message []byte) error {
	// only heartbeat from master
	remoteMessage, err := networking.DecodeUDPMessageMaster2Client(message)
	if err != nil {
		//log err
	}

	muxMasterMembershipList.Lock()
	newHeartbeat := remoteMessage.Heartbeat
	if newHeartbeat > _masterMembershipList.Heartbeat {
		_masterMembershipList.Heartbeat = newHeartbeat
		// update memebershiplist
	}
	muxMasterMembershipList.Unlock()

	if remoteMessage.MessageType == "ACK" {
		// this message is the ack to connect request
		isConnected = true 
		cli.Write2ClientMasterStatus(clientMasterStatusLabel, "CONN")
		// log success connect to master
		cli.Write2Shell(history, "Successfully connect to master")
		logger.LogSimpleInfo("Successfully connect to master")			
	} else if remoteMessage.MessageType == "KICKOUT" {
		cli.Write2ClientMasterStatus(clientMasterStatusLabel, "KICKED")
		cli.Write2Shell(history,"You are kicked out because of inactive")
		logger.LogSimpleInfo("You are kicked out because of inactive")	
		cli.Write2Shell(history, "Rejoin Y/N")
		constant.IsKickout = true
		cmd := <-constant.KickoutRejoinCmd
		if cmd == "true" {
			constant.IsKickout = false
			isConnected = false
		} else {
		}
	} 
	return nil
}

func detectMasterFail() {
	for {
		if isConnected == true && constant.IsKickout == false {
			muxMasterMembershipList.Lock()
			diff := time.Now().UnixNano()/1000000 - _masterMembershipList.Heartbeat
			muxMasterMembershipList.Unlock()
			
			if diff > constant.MasterTimeout {
				cli.Write2ClientMasterStatus(clientMasterStatusLabel, "FAIL")
				cli.Write2Shell(history, "detect master fail")
				logger.LogSimpleInfo("detect master fail")
				isConnected = false
				break
			}
		}
		time.Sleep(constant.ClientDetectMasterFailInterval * time.Millisecond)
	}
}

func connectMaster() {
	connectCount := 0
	for {
		if isConnected == false {
			if connectCount == 0 {
				cli.Write2Shell(history, "Send connect request to master")
				logger.LogSimpleInfo("Send connect request to master")
				// log send connect request to master
			} else {
				cli.Write2Shell(history, "Connect request Fail. Resend connect request to master")
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
	constant.KickoutRejoinCmd = make(chan string)
	_masterMembershipList = masterMembershipList{}
	_masterMembershipList.Heartbeat = 0
	clientIP, _ := networking.GetLocalIP()
	logger.LogSimpleInfo(clientIP)
	client2MasterMessageUDP = constant.UDPMessageClient2Master{clientIP,"CONNECT"}
	isConnected = false
	constant.IsKickout = false
	// try to connect to master, 
	go networking.UDPlisten(constant.UDPportMaster2Client, readUDPMessageMaster2Client)
	go connectMaster()
	go detectMasterFail()
	if cliLevel == "cli" {
		cliClient()
	} else {
		cliSimpleClient()
	}
} 