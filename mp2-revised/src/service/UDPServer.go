package service

import (
	"config"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"strings"
	. "structs"
	"sync"
	"time"
	"helper"
)

// bandwidth function
func countBandwidth() {
	for {
		time.Sleep(1 * time.Second)
		MT2.Lock()
		Bandwidth = 0
		MT2.Unlock()
	}
}

// cleanup function
func deleteIDAfterTcleanup(id string) {
	time.Sleep(time.Duration(Tclean) * time.Millisecond)
	MT.Lock()
	delete(MembershipList, id)
	MT.Unlock()
}

// fail detector
func selectFailedID(ticker *time.Ticker) {
	for {
		<-ticker.C
		MT.Lock()
		for id, member := range MembershipList {
			if id != MyID {
				diff := time.Now().UnixNano()/1000000 - member.HeartBeat
				_, ok := LeaveNodes[id]
				if diff > int64(Ttimeout) && MembershipList[id].HeartBeat != -1 && !ok {
					//fmt.Println(id + "might failed")
					helper.LogFail(Logger, MyVM, id, MembershipList[id].HeartBeat, diff)
					MembershipList[id] = Membership{-1, diff}
					FailedNodes[id] = 1
					go deleteIDAfterTcleanup(id)
					if currentFailTime1, ok := BroadcastAll[id]; ok {
						if currentFailTime1 < diff {
							BroadcastAll[id] = diff
						}
					} else {
						BroadcastAll[id] = diff
					}
					if currentFailTime2, ok := FirstDetect[id]; ok {
						if currentFailTime2 > diff {
							FirstDetect[id] = diff
						}
					} else {
						FirstDetect[id] = diff
					}
				}
			}
		}
		MT.Unlock()
	
	}
}

// gossip round robin function
func selectGossipID() []string {
	var num = len(Container)
	var res = make([]string, B)
	if num < 1 {
		rand.Seed(time.Now().Unix())
		MT.Lock()
		for key := range MembershipList {
			_, okLeave := LeaveNodes[key]
			_, okFail := FailedNodes[key]

			if key != MyID && !okLeave && !okFail {
				Container = append(Container, key)
			}
		}
		MT.Unlock()
		rand.Shuffle(len(Container), func(i, j int) { Container[i], Container[j] = Container[j], Container[i] })
	}
	num = len(Container)
	if num < B {
		res = Container[0:num]
		Container = Container[:0]
	} else {
		copy(res, Container[0:B])
		Container = Container[B:num]
	}
	return res
}

// merge membership list after receiving new membership list
func mergeMemberShipList(recievedMemberShipList map[string]Membership) {
	for key, _ := range recievedMemberShipList {
		if key == MyOldID {
			return
		}
	}
	for key, receivedMembership := range recievedMemberShipList {
		MT.Lock()
		if existedMembership, ok := MembershipList[key]; ok {
			if receivedMembership.HeartBeat == -2 {
				if _, ok := LeaveNodes[key]; !ok {
					helper.LogLeaver(Logger, MyVM, key)
					MembershipList[key] = receivedMembership
					LeaveNodes[key] = 1
					//build one row message
					tempMembershipList := map[string]Membership{key: MembershipList[key]}
					jsonString, err := json.Marshal(tempMembershipList)
					if err != nil {
						panic(err)
					}
					msg := string(jsonString)
					for id := range MembershipList {
						_, okLeave := LeaveNodes[id]
						_, okFail := FailedNodes[id]
						// dont send to myself, leave nodes and fail nodes
						if id != MyID && !okLeave && !okFail {
							sendMsgToID(id, msg)
						}
					}

					go deleteIDAfterTcleanup(key)
				} else {
					// we know it left, do nothing
				}
			} else if existedMembership.HeartBeat < receivedMembership.HeartBeat {
				_, okLeave := LeaveNodes[key]
				_, okFail := FailedNodes[key]
				if !okLeave && !okFail {
					MembershipList[key] = receivedMembership
				}
			}
		} else {
			if _, ok := FailedNodes[key]; ok {
				//refuse accept failed node
			} else if _, ok := LeaveNodes[key]; ok {
				// refuse accept leaved node
			} else {
				helper.LogJoiner(Logger, MyVM, key)
				MembershipList[key] = receivedMembership
			}
		}
		MT.Unlock()
	}
}

// handle UDP connection
func handleConnection(conn net.UDPConn) {
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	MT2.Lock()
	Bandwidth += n
	MT2.Unlock()
	if err != nil {
		Logger.Fatal(err)
	}
	msgString := string(buf)
	if msgString[:8] == "Command:" {
		command := strings.Split(msgString, ":")[1]
		C1 <- int(command[0]) + 8
	} else if msgString[:6] == "Leave:" {
		deleteID := strings.Split(msgString, ":")[1]
		deleteIDAfterTcleanup(deleteID)
	} else {
		//merge buf and membershiplist
		recievedMemberShipList := make(map[string]Membership)
		err = json.Unmarshal(buf[:n], &recievedMemberShipList)
		if err != nil {
			Logger.Fatal(err)
		}
		//introducer handle new joiner
		if len(recievedMemberShipList) == 1 && MyIP == IntroIP {
			for key, _ := range recievedMemberShipList {
				// if it is not in introducer membershiplist
				if _,ok := MembershipList[key]; !ok {
					MT.Lock()
					jsonString, err := json.Marshal(MembershipList)
					MT.Unlock()
					if err != nil {
						Logger.Fatal(err)
					}
					msg := string(jsonString)
					sendMsgToID(key, msg)
				}
			}
		}
		mergeMemberShipList(recievedMemberShipList)
	}
}

func listenUDP() {

	udpAddr, err := net.ResolveUDPAddr("udp4", ":"+MyPort)
	if err != nil {
		panic(err)
		return
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		panic(err)
	}
	for {
		if !IsJoin {
			continue
		}
		handleConnection(*conn)
	}
}

func sendMsgToID(id string, msg string) {
	//simulate for loss rate
	if rand.Intn(100) < LossRate {
		return
	}
	ip := strings.Split(id, "*")[0]
	conn, err := net.Dial("udp", ip)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(conn, msg+"\n")
}

// the name should be multicast UDP
func broadcastUDP() {
	if !IsJoin {
		return
	}
	if CurrentProtocol == "All2All" {
		selfNode := make(map[string]Membership)
		MT.Lock()
		selfNode[MyID] = MembershipList[MyID]
		MT.Unlock()
		jsonString, err := json.Marshal(selfNode)
		if err != nil {
			panic(err)
		}
		msg := string(jsonString)
		MT.Lock()
		for id, _ := range MembershipList {
			_, okLeave := LeaveNodes[id]
			_, okFail := FailedNodes[id]
			// dont send to myself, leave nodes and fail nodes
			if id != MyID && !okLeave && !okFail {
				//Logger.Info("All2All: " + helper.ConvertIDtoVM(id) + fmt.Sprintf(" %v", len(msg)))
				sendMsgToID(id, msg)
			}
		}
		MT.Unlock()
	} else if CurrentProtocol == "Gossip" {
		MT.Lock()
		jsonString, err := json.Marshal(MembershipList)
		MT.Unlock()
		if err != nil {
			panic(err)
		}
		msg := string(jsonString)
		idList := selectGossipID()
		// dont send to myself, leave nodes and fail nodes
		for _, id := range idList {
			//Logger.Info("Gossip: " + helper.ConvertIDtoVM(id) + fmt.Sprintf(" %v", len(msg)))
			sendMsgToID(id, msg)
		}
	}
}

func updateSelfHeartBeat() {
	for {
		t := time.Now().UnixNano() / 1000000
		MT.Lock()
		MembershipList[MyID] = Membership{t, t}
		MT.Unlock()
	}
}

func joinGroup() {
	MT.Lock()
	jsonString, err := json.Marshal(MembershipList)
	MT.Unlock()
	if err != nil {
		panic(err)
	}
	msg := string(jsonString)
	introIP, err := config.IntroducerIPAddresses()
	if err != nil {
		panic(err)
	}
	introPort, err := config.Port()
	if err != nil {
		panic(err)
	}
	id := string(introIP[0]) + ":" + introPort
	sendMsgToID(id, msg)
	helper.LogSelfJoin(Logger, MyVM, MyID)
	IsJoin = true
}

func leaveGroup() {
	MT.Lock()
	IsJoin = false
	helper.LogSelfLeave(Logger, MyVM, MyID)
	MembershipList[MyID] = Membership{-2, 0}
	LeaveNodes[MyID] = 1
	jsonString, err := json.Marshal(MembershipList)
	if err != nil {
		panic(err)
	}
	msg := string(jsonString)
	for id := range MembershipList {
		_, okFail := FailedNodes[id]
		_, okLeave := LeaveNodes[id]
		if id != MyID && !okFail && !okLeave{
			sendMsgToID(id, msg)
		}
	}
	millis := time.Now().UnixNano() / 1000000
	secs := millis / 1000
	heartBeat := millis
	MyOldID = MyID
	MyID = MyIP + ":" + MyPort + "*" + fmt.Sprint(secs)
	MembershipList = make(map[string]Membership)
	MembershipList[MyID] = Membership{HeartBeat: heartBeat, FailedTime: -1}
	FailedNodes = make(map[string]int)
	LeaveNodes = make(map[string]int)
	MT.Unlock()
}

func piggybackCommand(cmd int) {
	if cmd == CHANGE_TO_ALL2ALL {
		NextProtocol = "All2All"
		helper.LogReceiveProtocol(Logger, MyVM, MyID, NextProtocol)
	} else if cmd == CHANGE_TO_GOSSIP {
		NextProtocol = "Gossip"
		helper.LogReceiveProtocol(Logger, MyVM, MyID, NextProtocol)
	}
	msg := "Command:" + string(cmd)
	MT.Lock()
	for id := range MembershipList {
		_, okLeave := LeaveNodes[id]
		_, okFail := FailedNodes[id]
		// dont send to myself, leave nodes and fail nodes
		if id != MyID && !okLeave && !okFail {
			sendMsgToID(id, msg)
		}
	}
	MT.Unlock()

}
/**
func parseCmds(cmds []int) []int {
	//gossip or all2all
	//fmt.Println(cmds)
	if len(cmds) == 0 {
		return make([]int, 0)
	}
	gossipOrAll2All := -1
	joinGroupIndex := -1
	leaveGroupIndex := -1

	for i := len(cmds) - 1; i >= 0; i-- {
		if cmds[i] == CHANGE_TO_GOSSIP || cmds[i] == CHANGE_TO_ALL2ALL {
			if gossipOrAll2All == -1 {
				gossipOrAll2All = cmds[i]
			}
		}
		if cmds[i] == JOIN_GROUP {
			if joinGroupIndex == -1 {
				joinGroupIndex = i
			}
		}
		if cmds[i] == LEAVE_GROUP {
			if leaveGroupIndex == -1 {
				leaveGroupIndex = i
			}
		}
	}
	res := make([]int, 0)
	if gossipOrAll2All != -1 {
		res = append(res, gossipOrAll2All)
	}
	if joinGroupIndex > leaveGroupIndex && IsJoin == false {
		res = append(res, JOIN_GROUP)
		return res
	} else if joinGroupIndex > leaveGroupIndex && IsJoin == true {
		log.Println("cannot join group twice")
		if leaveGroupIndex != -1 {
			res = append(res, LEAVE_GROUP)
			return res
		}
		return res
	} else if joinGroupIndex < leaveGroupIndex && IsJoin == true {
		res = append(res, LEAVE_GROUP)
		return res
	} else if joinGroupIndex < leaveGroupIndex && IsJoin == false {
		log.Println("cannot leave group before join")
		if joinGroupIndex != -1 {
			res = append(res, JOIN_GROUP)
			return res
		}
		return res
	}
	return res
}
**/

//UDPServer is the udp server thread function
func UDPServer(wg *sync.WaitGroup, c chan int) {
	defer wg.Done()
	//ticker for gossip and all2all period; ClogN * gossip = all2all 
	ticker := time.NewTicker(time.Duration(Tgossip) * time.Millisecond)
	//ticker for fail detect period; gossip = all2all 
	tickerDetectFail := time.NewTicker(time.Duration(Tgossip) * time.Millisecond)
	//command from CLI
	cmd := 0
	gossipCounter := 0
	// bandwidth thread
	go countBandwidth()
	// udp handler thread
	go listenUDP()
	// fail detector thread
	go selectFailedID(tickerDetectFail)
	// main loop
	for {
		t1 := time.Now()
		<-ticker.C
		// here a new gossip period starts

		// step0: check if gossip period is long enough to run the code in each gossip period? 
		t2 := time.Now()
		diff := t2.Sub(t1)
		if float32(diff/time.Millisecond) < float32(float32(Tgossip)*0.05) {
			Logger.Fatal("gossip period time too short")
		}
		gossipCounter = gossipCounter + 1

		// step 1: change to other protocol if needed
		if CurrentProtocol != NextProtocol{
			helper.LogChangeProtocol(Logger, MyVM, MyID, CurrentProtocol, NextProtocol)
			if NextProtocol == "Gossip" {
				ticker.Reset(time.Duration(Tgossip) * time.Millisecond)
				CurrentProtocol = "Gossip"
				select {
					// for simple cli this is a channel with no receiver
				case ProtocolChangeACK <- "Gossip":
					Logger.Debug("Send Gossip to CLI")
				default:
				}
			} else {
				ticker.Reset(time.Duration(Tall2all) * time.Millisecond)
				CurrentProtocol = "All2All"
				select {
				case ProtocolChangeACK <- "All2All":
					Logger.Debug("Send All2All to CLI")
				default:
				}
			}
		}

		// step2: read commands
		cmds := make([]int, 0)
	forLoop:
		for {
			select {
			case cmd = <-c:
				cmds = append(cmds, cmd)
			default:
				if len(cmds) == 0 {
					Logger.Debug("No command from CLI. Do nothing")
				} else {
					Logger.Debug("No more commands.")
					Logger.Info("Commands received:", cmds)
				}
				break forLoop
			}
		}

		// step3: execute commands 
		if len(cmds) != 0 {
			for _, cmd := range cmds {
				switch cmd {
				case CHANGE_TO_ALL2ALL:
					// actually we just broadcast it
					piggybackCommand(CHANGE_TO_ALL2ALL)
				case CHANGE_TO_GOSSIP:
					// actually we just broadcast it
					piggybackCommand(CHANGE_TO_GOSSIP)
				case JOIN_GROUP:
					joinGroup()
				case LEAVE_GROUP:
					leaveGroup()
				case RECEIVE_CHANGE_TO_ALL2ALL:
					NextProtocol = "All2All"
					helper.LogReceiveProtocol(Logger, MyVM, MyID, NextProtocol)
				case RECEIVE_CHANGE_TO_GOSSIP:
					NextProtocol = "Gossip"
					helper.LogReceiveProtocol(Logger, MyVM, MyID, NextProtocol)
				}
			}
		}

		t := time.Now().UnixNano() / 1000000
		//if not leave
		MT.Lock()
		MembershipList[MyID] = Membership{t, -1}
		MT.Unlock()
		broadcastUDP()
	}
}
