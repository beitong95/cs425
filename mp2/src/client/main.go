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
	ds "datastructure"
)
/**
 Finished parts:
 1. connect to master
 2. detect master fail
 3. reconnect master if master fails
 4. kick out prompt
 5. kick out and rejoin 

 TODO: 
 1. get
 2. put
 3. abort current command when master fails
 4. resend current command if current command fails
 5. command queue or command mutual exclusion 
 ( command queue: allow user input multi commands in a short time. 
   command mutual exclusion: user cannot type new command until current command finishs)
**/
 type masterMembershipList struct {
	Heartbeat int64
}

var (
	_masterMembershipList masterMembershipList 
	isConnected bool
	muxMasterMembershipList sync.Mutex
	client2MasterMessageUDP constant.UDPMessageClient2Master
	isKickout bool
	cmdQueue ds.CommandQueue
	cmdStatusQueue ds.CommandQueue
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

func handleCommand(_cmd []string) {
	cmd := _cmd[0]
	filename1 := _cmd[1]
	filename2 := _cmd[2]
	cli.Write2Shell(history, cmd + filename1 + filename2)
	switch cmd {
		case "get":
			//go getFile()
			cli.Write2Shell(history, "TODO")
		case "put":
			//go putFile()
			cli.Write2Shell(history, "TODO")
		case "delete":
			//go putFile()
			cli.Write2Shell(history, "TODO")
		case "ls":
			//go lsFile()
			cli.Write2Shell(history, "TODO")
		case "store":
			//go storeFile()
			cli.Write2Shell(history, "TODO")
	}
}

func handleCommands() {
	cmd := []string{}
	for {
		for !cmdQueue.IsEmpty(){
			cmd = cmdQueue.Dequeue()
			handleCommand(cmd)
		}
		
	}
}
// all commands should be parallel? 
func downloadFileFromDatanode(filename string, ip string) (*file, error) {
	//XINHANG
	//http download file
	//store in local location
}

func GetIPsFromMaster(filename string, masterIP string) ([]string, error) {
	
	// XINHANG
	// send request to master server 
	// get IPs	
	
	return IPs, nil
}

func getFile(filename string, masterIp string) {
	// command start 
	IPs, err := networking.GetIPsFromMaster(filename, masterIP)
	for i, ip := range IPs {
		file, err := downloadFileFromDatanode(filename, ip)
		if err == nil {
			break 
		}
		if i == len(IPs) - 1 {
			// fatal error all 4 data nodes down 
		}
	}
	// command end
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

func lsFile() {

}

func storeFile() {

}

func Run(cliLevel string) {
	// initialize
	constant.KickoutRejoinCmd = make(chan string)
	cmdQueue = ds.CommandQueue{}
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
	go handleCommands()
	if cliLevel == "cli" {
		cliClient()
	} else {
		cliSimpleClient()
	}
} 