package master

import (
	_ "errors"
	"sync"
	"time"
	"logger"
	"cli"
	"networking"
	"constant"
	"strings"
	"fmt"
)

/**
Finished:
1. detect inactive client
2. handle client connection
3. maintain client membershiplist
4. handle datanode connection
5. detect datanode fail
6. maintain datanode membershiplist

TODO:
1. master read in client commands and put them in queue
2. master process those commands, allow parallel read or single write (because we use a queue, there is no starving problem)
3. given a file name locate the 4 copies' ips
4. handle datanode fail, create an rereplica algorithm
5. maintain vm2file and file2vm

**/


var Vm2fileMap map[string] []string

var File2VmMap map[string] []string

var MessageQueue []string

var MQ sync.Mutex

var MW sync.Mutex

var MR sync.Mutex

var MF sync.Mutex

var MV sync.Mutex

var ReadCounter int = 0

var WriteCounter int = 0

type clientMembershipList map[string] int64
// map [client ip] last active time
type datanodeMembershipList map[string] int64
// map [datanode ip] heartbeat

var (
	_clientMembershipList clientMembershipList
	_datanodeMembershipList datanodeMembershipList
	muxClientMembershipList sync.Mutex
	muxDatanodeMembershipList sync.Mutex
)

func FindMaxLen(ips []string) (int,string) {
	var output = ""
	var idx = 0
    for i := 0; i < 4; i++ {
		if output == "" {
			output = ips[i]
		} else if len(Vm2fileMap[ips[i]]) > len(Vm2fileMap[output]) {
			output = ips[i]
			idx = i
		}
	}
	return idx,output
}

func Hash2Ips(filename string) {
	// assert filename is name of new file!
	var fourIps = []string{"","","",""}
	MV.Lock()
	for ip := range Vm2fileMap {
		if fourIps[0] == "" {
			fourIps[0] = ip
		} else if fourIps[1] == "" {
			fourIps[1] = ip
		} else if fourIps[2] == "" {
			fourIps[2] = ip
		} else if fourIps[3] == "" {
			fourIps[3] = ip
		} else {
			var idx,maxlen = FindMaxLen(fourIps)
			if len(Vm2fileMap[ip]) < len(Vm2fileMap[maxlen]) {
				fourIps[idx] = ip
			}
		}
	}
	for i := 0 ; i < 4; i++ {
		Vm2fileMap[fourIps[i]] = append(Vm2fileMap[fourIps[i]],filename)
	}
	MV.Unlock()
	MF.Lock()
	File2VmMap[filename] = fourIps
	MF.Unlock()
}

func find(filename string, ip string) bool {
	MF.Lock()
	var ips = File2VmMap[filename]
	for i := 0; i < len(ips); i++ {
		if ips[i] == ip {
			return true
		}
	}
	MF.Unlock()
	return false
}

func rereplica(filename string) {
	MV.Lock()
	var replica = ""
	for ip := range Vm2fileMap {
		var found = find(filename, ip)
		if replica == "" {
			replica = ip
		} else if len(Vm2fileMap[ip]) < len(Vm2fileMap[replica]) && !found {
			replica = ip
		}
	}
	Vm2fileMap[replica] = append(Vm2fileMap[replica],filename)
	MV.Unlock()
	MF.Lock()
	File2VmMap[filename] = append(File2VmMap[filename],replica)
	MF.Unlock()
}

func enqueue(cmd string){
	MQ.Lock()
	MessageQueue = append(MessageQueue,cmd)
	MQ.Unlock()
}

func enqueue_front(cmd string) {
	MQ.Lock()
	MessageQueue = append([]string{cmd},MessageQueue...)
	MQ.Unlock()
}

func dequeue() string{
	MQ.Lock()
	var num = len(MessageQueue)
	fmt.Println(num)
	MQ.Unlock()
	if  num == 0 {
		return ""
	}
	var output = MessageQueue[0]
	if num == 1 {
		MQ.Lock()
		MessageQueue = []string {}
		MQ.Unlock()
		return output
	}
	MQ.Lock()
	MessageQueue = MessageQueue[1:]
	MQ.Unlock()
	//fmt.Println(MessageQueue)
	return output
}
// when master receives (get put delete store ls) => (read write), handle read write concurrency problem.
func HandleMessage() {
	//MessageQueue = []string {"get1","get2","put1","get3","delete1","put2","ls1","store1"}
	for {
		//mutex
		fmt.Println(MessageQueue)
		var cmd = dequeue()
		if strings.Contains(cmd,"put") || strings.Contains(cmd,"delete") || strings.Contains(cmd,"replica") {
			//handle write, wait for reads (count acks)
			// I am not sure if this is hanging
			// we can use wait group(like semaphore)
			
			for {
				MW.Lock()
				MR.Lock()
				if ReadCounter == 0 && WriteCounter == 0 {
					MR.Unlock()
					MW.Unlock()
					break
				}
				MR.Unlock()
				MW.Unlock()
			}
			MW.Lock()
			WriteCounter++
			MW.Unlock()
			go handleCmd(cmd)
		} else {
			//handle read
			MR.Lock()
			ReadCounter++
			MR.Unlock()
			for {
				MW.Lock()
				if WriteCounter == 0 {
					MW.Unlock()
					break
				}
				MW.Unlock()
			}
			go handleCmd(cmd)
		}
	}
}

func handleCmd(cmd string) {
	//TODO: this version is just for test
	fmt.Println(cmd,time.Now())
	//simulate handle cmd
	time.Sleep(2*time.Second)
	if strings.Contains(cmd,"put") || strings.Contains(cmd,"delete") {
		MW.Lock()
		WriteCounter--
		MW.Unlock()
	} else {
		MR.Lock()
		ReadCounter--
		MR.Unlock()
	}
	fmt.Println(WriteCounter," ",ReadCounter)
}
// handle udp message from client
func readUDPMessageClient2Master(message []byte) error {
	remoteMessage, err := networking.DecodeUDPMessageClient2Master(message)
	if err != nil {
		// log err
	}
	if remoteMessage.MessageType == "CONNECT" {
		muxClientMembershipList.Lock()
		clientIP := remoteMessage.IP
		logger.LogSimpleInfo("receive connect from " + clientIP)
		cli.Write2Shell(history, "receive connect from " + clientIP)
		if _, ok := _clientMembershipList[clientIP]; ok {
			_clientMembershipList[clientIP] = time.Now().UnixNano()/1000000
		} else {
			_clientMembershipList[clientIP] = time.Now().UnixNano()/1000000
		}
		muxClientMembershipList.Unlock()
		cli.Write2MasterClientMembershipBox(masterClientMembershipLabel, cli.ConvertMasterClientMembershipList2String(_clientMembershipList, muxClientMembershipList))

		//update cli

		// send ack back
		heartbeat := time.Now().UnixNano()/1000000
		message, _ := networking.EncodeUDPMessageMaster2Client(&constant.UDPMessageMaster2Client{heartbeat, "ACK"})
		logger.LogSimpleInfo("send ack back to " + clientIP)
		cli.Write2Shell(history, "send ack back to " + clientIP)
		networking.UDPsend(clientIP, constant.UDPportMaster2Client, message)
	}
	return nil
}


func detectClientInactive() {
	for {
		muxClientMembershipList.Lock()
		currentTime := time.Now().UnixNano()/1000000
		for i,v := range _clientMembershipList {
			// kick out 1min inactive client (actual time 1min20s)
			if  currentTime - v > constant.KickoutTimeout  {
				// send kick out and delete
				logger.LogSimpleInfo("send kickout to " + i)
				cli.Write2Shell(history, "send kickout to "+ i)
				message, _ := networking.EncodeUDPMessageMaster2Client(&constant.UDPMessageMaster2Client{0, "KICKOUT"})
				networking.UDPsend(i, constant.UDPportMaster2Client, message)
				delete(_clientMembershipList, i)
				//update cli
			}
		}
		muxClientMembershipList.Unlock()
		cli.Write2MasterClientMembershipBox(masterClientMembershipLabel, cli.ConvertMasterClientMembershipList2String(_clientMembershipList, muxClientMembershipList))

		// every 20s check it
		time.Sleep(constant.CheckInactiveClientInterval* time.Millisecond)
	}
}

func sendHeartbeat2Clients() {
	for {
		heartbeat := time.Now().UnixNano()/1000000
		message, _ := networking.EncodeUDPMessageMaster2Client(&constant.UDPMessageMaster2Client{heartbeat, "HEARTBEAT"})
		for i,_ := range _clientMembershipList {
			//logger.LogSimpleInfo("send heartbeat to " + i)
			//cli.Write2Shell("send heartbeat to "+ i)
			networking.UDPsend(i, constant.UDPportMaster2Client, message)
		} 
		time.Sleep(constant.MasterSendHeartbeat2ClientInterval* time.Millisecond)
	}


}

// handle udp message from datanode
func readUDPMessageDatanode2Master(message []byte) error {
	remoteMessage, err := networking.DecodeUDPMessageDatanode2Master(message)
	id := remoteMessage.ID
	if err != nil {
		// log err
	}
	if remoteMessage.MessageType == "HEARTBEAT" {
		muxDatanodeMembershipList.Lock()
		newHeartbeat := remoteMessage.Heartbeat

		if _, ok := _datanodeMembershipList[id]; !ok {
			// new datanode
			logger.LogSimpleInfo(id + " join with heartbeat " + fmt.Sprintf("%v", newHeartbeat))
			cli.Write2Shell(history, id + " join with heartbeat " + fmt.Sprintf("%v", newHeartbeat))
			_datanodeMembershipList[id]= newHeartbeat
		} else {
			if newHeartbeat > _datanodeMembershipList[id]{
				logger.LogSimpleInfo(id + " update heartbeat from " + fmt.Sprintf("%v", _datanodeMembershipList[id]) + " to " + fmt.Sprintf("%v", newHeartbeat))
				_datanodeMembershipList[id]= newHeartbeat
			}
		}

		muxDatanodeMembershipList.Unlock()
		cli.Write2MasterDatanodeMembershipBox(masterDatanodeMembershipLabel,cli.ConvertMasterDatanodeMembershipList2String(_datanodeMembershipList, muxDatanodeMembershipList))

	}
	return nil
}

func detectDatanodeFail() {
	for {
		currentTime := time.Now().UnixNano()/1000000
		muxDatanodeMembershipList.Lock()
		for i, v := range _datanodeMembershipList {
			diff := currentTime - v
			if diff > constant.DatanodeTimeout {
				logger.LogSimpleInfo("detect data node fail " + i)
				cli.Write2Shell(history, "detect data node fail " + i)
				delete(_datanodeMembershipList, i)
			// TODO: handle rereplica
				logger.LogSimpleInfo("remove node " + i)
				cli.Write2Shell(history, "remove node " + i)
			}
		}
		muxDatanodeMembershipList.Unlock()
		cli.Write2MasterDatanodeMembershipBox(masterDatanodeMembershipLabel,cli.ConvertMasterDatanodeMembershipList2String(_datanodeMembershipList, muxDatanodeMembershipList))
		
		time.Sleep(constant.MasterDetectDatanodeFailInterval* time.Millisecond)
	}
}

func Run(cliLevel string) {
	_clientMembershipList = clientMembershipList{}
	_datanodeMembershipList = datanodeMembershipList{}
	go networking.UDPlisten(constant.UDPportClient2Master, readUDPMessageClient2Master)
	go networking.UDPlisten(constant.UDPportDatanode2Master, readUDPMessageDatanode2Master)
	go detectClientInactive()
	go detectDatanodeFail()
	go sendHeartbeat2Clients()
	if cliLevel == "cli" {
		cliMaster()
	} else {
		cliSimpleMaster()
	}
} 
