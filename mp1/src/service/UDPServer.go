package service

import (
	"config"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	. "structs"
	"sync"
	"time"
)

//var isJoin bool = false
//Gossip parameters
// var B int = 2
// var preservedB int = 1
func countBandwidth() {
	for {
		time.Sleep(1 * time.Second)
		MT2.Lock()
		//		fmt.Println("Current Bandwidth: ", Bandwidth)
		Bandwidth = 0
		MT2.Unlock()
	}
}
func deleteIDAfterTcleanup(id string) {
	time.Sleep(time.Duration(Tclean) * time.Millisecond)
	MT.Lock()
	delete(MembershipList, id)
	MT.Unlock()
}

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
					//fmt.Println("timeout: " + fmt.Sprint(diff))
				}
			}
		}
		MT.Unlock()
	
	}
}
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
				//			if key != MyID{
				Container = append(Container, key)
			}
		}
		MT.Unlock()
		//fmt.Println(Container)
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

func mergeMemberShipList(recievedMemberShipList map[string]Membership) {
	//MT.Lock()
	for key, _ := range recievedMemberShipList {
		if key == MyOldID {
			return
		}
	}
	//MT.Unlock()

	for key, receivedMembership := range recievedMemberShipList {
		MT.Lock()
		if existedMembership, ok := MembershipList[key]; ok {
			if receivedMembership.HeartBeat == -2 {
				if _, ok := LeaveNodes[key]; !ok {
					MembershipList[key] = receivedMembership
					LeaveNodes[key] = 1
					//build one row message
					// possible BUG
					tempMembershipList := map[string]Membership{key: MembershipList[key]}
					jsonString, err := json.Marshal(tempMembershipList)
					if err != nil {
						panic(err)
					}
					msg := string(jsonString)
					for id := range MembershipList {
						if id != MyID && LeaveNodes[id] != 1 && FailedNodes[id] != 1 {
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
				//fmt.Printf("key: %v, update time: %v\n", key, receivedMembership.HeartBeat-existedMembership.HeartBeat)
			}
		} else {
			if _, ok := FailedNodes[key]; ok {
				//refuse accept failed node
			} else if _, ok := LeaveNodes[key]; ok {
				// refuse accept leaved node
			} else {
				MembershipList[key] = receivedMembership
			}
		}
		MT.Unlock()
	}
	//UpdateGUI <- "Ping"

}
func handleConnection(conn net.UDPConn) {
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	MT2.Lock()
	Bandwidth += n
	//fmt.Println(Bandwidth)
	MT2.Unlock()
	if err != nil {
		fmt.Println(err)
	}
	msgString := string(buf)
	if msgString[:8] == "Command:" {
		command := strings.Split(msgString, ":")[1]
		//		fmt.Println(int(command[0]))
		C <- int(command[0]) + 8
		//fmt.Println(fmt.Sprint(int(command[0])))
	} else if msgString[:6] == "Leave:" {
		deleteID := strings.Split(msgString, ":")[1]
		deleteIDAfterTcleanup(deleteID)
	} else {
		// fmt.Println(string(buf) + " " + fmt.Sprint(n) + " bytes read")
		//merge buf and membershiplist
		recievedMemberShipList := make(map[string]Membership)

		//fix bug

		err = json.Unmarshal(buf[:n], &recievedMemberShipList)
		if err != nil {
			panic(err)
		}
		if len(recievedMemberShipList) == 1 {
			// this sender must be a new memeber
			// because he/she sends it to me
			// then i must be the introducer
			MT.Lock()
			jsonString, err := json.Marshal(MembershipList)
			MT.Unlock()
			if err != nil {
				panic(err)
			}
			msg := string(jsonString)
			for key, _ := range recievedMemberShipList {
				sendMsgToID(key, msg)
			}
		}
		mergeMemberShipList(recievedMemberShipList)
	}

}
func listenUDP() {

	udpAddr, err := net.ResolveUDPAddr("udp4", ":"+MyPort)
	//fmt.Println("listen on port:" + MyPort)
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
func broadcastUDP() {
	//fmt.Println("idList: ", idList, "len: ", len(idList))
	// for id := range MembershipList {
	// 	if MyID != id {
	// 		ip := strings.Split(id, "*")[0]
	// 		conn, err := net.Dial("udp", ip)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 		}
	// 		fmt.Fprintf(conn, msg+"\n")
	// 	}
	// }
	if !IsJoin {
		return
	}
	if IsAll2All {
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
			if id != MyID {
				sendMsgToID(id, msg)
			}
		}
		MT.Unlock()
	} else if IsGossip {
		MT.Lock()
		jsonString, err := json.Marshal(MembershipList)
		MT.Unlock()
		if err != nil {
			panic(err)
		}
		msg := string(jsonString)
		idList := selectGossipID()
		for _, id := range idList {
			sendMsgToID(id, msg)
			//fmt.Println(msg)
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
	//fmt.Println("joining group")
	MT.Lock()
	jsonString, err := json.Marshal(MembershipList)
	MT.Unlock()
	if err != nil {
		panic(err)
	}
	msg := string(jsonString)
	//fmt.Println(msg)
	introIP, err := config.IntroducerIPAddresses()
	if err != nil {
		panic(err)
	}
	introPort, err := config.Port()
	if err != nil {
		panic(err)
	}
	//fmt.Println(string(introIP[0]) + ":" + introPort)
	id := string(introIP[0]) + ":" + introPort
	sendMsgToID(id, msg)
	//fmt.Println("join group")
	IsJoin = true
}
func leaveGroup() {
	MT.Lock()
	IsJoin = false
	MembershipList[MyID] = Membership{-2, 0}
	LeaveNodes[MyID] = 1
	jsonString, err := json.Marshal(MembershipList)
	if err != nil {
		panic(err)
	}
	msg := string(jsonString)
	for id := range MembershipList {
		_, okFail := FailedNodes[id]
		if id != MyID && !okFail {
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
		IsAll2All = true
		IsGossip = false
	} else if cmd == CHANGE_TO_GOSSIP {
		IsGossip = true
		IsAll2All = false
	}
	msg := "Command:" + string(cmd)
	MT.Lock()
	for id := range MembershipList {
		sendMsgToID(id, msg)
	}
	MT.Unlock()

}

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

//UDPServer is the udp server thread function
func UDPServer(isAll2All bool, isIntroducer bool, wg *sync.WaitGroup, c chan int) {
	//log.SetOutput(ioutil.Discard)
	defer wg.Done()
	//timer for gossip period
	ticker := time.NewTicker(time.Duration(Tgossip) * time.Millisecond)
	tickerDetectFail := time.NewTicker(time.Duration(Tgossip) * time.Millisecond)
	//command from CLI
	cmd := 0
	gossipCounter := 0
	go countBandwidth()
	go listenUDP()
	// main loop
	go selectFailedID(tickerDetectFail)
	for {
		//helper.PrintMembershipListAsTable(MembershipList)
		// can go through here ever gossipPeriod
		log.Println("waiting for next gossip period")
		t1 := time.Now()
		//no wait means our gossip period is too short for gossip process
		<-ticker.C
		if CurrentProtocol != IsGossip {
			if IsGossip == true {
				ticker.Reset(time.Duration(Tgossip) * time.Millisecond)
				CurrentProtocol = true
				ProtocolChangeACK <- "Gossip"
			} else {
				ticker.Reset(time.Duration(Tall2all) * time.Millisecond)
				CurrentProtocol = false
				ProtocolChangeACK <- "All2All"
			}

		}
		t2 := time.Now()
		diff := t2.Sub(t1)
		log.Println("wait time:", diff)
		if float32(diff/time.Millisecond) < float32(float32(Tgossip)*0.05) {
			log.Fatalln("gossip period time too short")
		}
		gossipCounter = gossipCounter + 1
		log.Println("----------------------------------------------")
		log.Println("Start gossip period", gossipCounter)
		// in every gossipPeriod, the first thing is to read commands from CLI
		cmds := make([]int, 0)
		// read commands
	forLoop:
		for {
			select {
			case cmd = <-c:
				log.Printf("UDPServer receives cmd from CLI in %d gossip period: %d\n", gossipCounter, cmd)
				cmds = append(cmds, cmd)
			default:
				if len(cmds) == 0 {
					log.Println("No command from CLI. Do nothing")
				} else {
					log.Println("No more commands.")
					log.Println("Commands received:", cmds)
				}
				break forLoop
			}
		}
		log.Println("Doing Gossip work with commands", cmds)
		//we should process cmds in sequence, and there are some rules
		//for example, if join and leave are in the same cmd sequence, we should only execute leave
		//cmds = parseCmds(cmds)
		//fmt.Println(cmds)
		log.Println("Parsed commands ", cmds)
		//execute commands
		if len(cmds) != 0 {
			for _, cmd := range cmds {
				//fmt.Println(cmd)
				switch cmd {
				//if change gossip to all2all or all2all to gossip
				//change b, add command to membership list
				case CHANGE_TO_ALL2ALL:
					// B = len(MembershipList)
					piggybackCommand(CHANGE_TO_ALL2ALL)
				case CHANGE_TO_GOSSIP:
					// B = preservedB
					piggybackCommand(CHANGE_TO_GOSSIP)
				case JOIN_GROUP:
					joinGroup()
				case LEAVE_GROUP:
					leaveGroup()
				case RECEIVE_CHANGE_TO_ALL2ALL:
					IsAll2All = true
					IsGossip = false
				case RECEIVE_CHANGE_TO_GOSSIP:
					IsAll2All = false
					IsGossip = true
				}
			}
		}
		// TODO: Gossip logic
		//merge membershiplist
		/**
		jsonString, err := json.Marshal(MembershipList)
		if err != nil {
			panic(err)
		}
		//fmt.Println(string(jsonString))
		**/

		// helper.PrintMembershipListAsTable(MembershipList)
		t := time.Now().UnixNano() / 1000000
		//fmt.Println(t)
		//if not leave
		MT.Lock()
		MembershipList[MyID] = Membership{t, -1}
		//UpdateGUI <- "Ping"
		MT.Unlock()
		broadcastUDP()
		//update timer
		//control timers
		//failure detect
		//deseminate failure
		//execute global commands set B
		//time.Sleep(time.Duration(Tgossip) * time.Millisecond)
		log.Println("Finish Gossip work")
	}
}
