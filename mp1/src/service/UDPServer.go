package service

import (
	"config"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	. "structs"
	"sync"
	"time"
)

var isJoin bool = false

//Gossip parameters
var B int = 1
var preservedB int = 1

func mergeMemberShipList(recievedMemberShipList map[string]Membership) {
	for key, receivedMembership := range recievedMemberShipList {
		if existedMembership, ok := MembershipList[key]; ok {
			if existedMembership.HeartBeat < receivedMembership.HeartBeat {
				MembershipList[key] = receivedMembership
			}
		} else {
			MembershipList[key] = receivedMembership
		}
	}
}
func handleConnection(conn net.UDPConn) {
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println(err)
	}
	n, err = conn.Read(buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(buf) + " " + fmt.Sprint(n) + " bytes read")
	//merge buf and membershiplist
	recievedMemberShipList := make(map[string]Membership)
	err = json.Unmarshal(buf[:n], &recievedMemberShipList)
	if err != nil {
		panic(err)
	}
	mergeMemberShipList(recievedMemberShipList)

}
func listenUDP() {
	udpAddr, err := net.ResolveUDPAddr("udp4", ":"+MyPort)
	fmt.Println("listen on port:" + MyPort)
	if err != nil {
		panic(err)
		return
	}
	conn, err := net.ListenUDP("udp", udpAddr)

	if err != nil {
		panic(err)
	}
	for {
		handleConnection(*conn)
	}
}
func boardcastUDP() {
	jsonString, err := json.Marshal(MembershipList)
	if err != nil {
		panic(err)
	}
	msg := string(jsonString)
	for id := range MembershipList {
		if MyID != id {
			ip := strings.Split(id, "*")[0]
			conn, err := net.Dial("udp", ip)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Fprintf(conn, msg+"\n")
		}
	}

}
func joinGroup() {
	fmt.Println("joining group")
	jsonString, err := json.Marshal(MembershipList)
	if err != nil {
		panic(err)
	}
	msg := string(jsonString)
	fmt.Println(msg)
	introIP, err := config.IntroducerIPAddresses()
	if err != nil {
		panic(err)
	}
	introPort, err := config.Port()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(introIP[0]) + ":" + introPort)
	conn, err := net.Dial("udp", string(introIP[0])+":"+introPort)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(conn, msg)
	fmt.Println("join group")
}

func leaveGroup() {
	log.Println("leave group")
}

func piggybackCommand(cmd int) {
	log.Println("Sending cmd:", cmd)
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
	if joinGroupIndex > leaveGroupIndex && isJoin == false {
		res = append(res, JOIN_GROUP)
		return res
	} else if joinGroupIndex > leaveGroupIndex && isJoin == true {
		log.Println("cannot join group twice")
		if leaveGroupIndex != -1 {
			res = append(res, LEAVE_GROUP)
			return res
		}
		return res
	} else if joinGroupIndex < leaveGroupIndex && isJoin == true {
		res = append(res, LEAVE_GROUP)
		return res
	} else if joinGroupIndex < leaveGroupIndex && isJoin == false {
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
	gossipPeriodMillisecond := 2000
	//timer for gossip period
	ticker := time.NewTicker(time.Duration(gossipPeriodMillisecond) * time.Millisecond)
	//command from CLI
	cmd := 0
	gossipCounter := 0
	go listenUDP()
	// main loop
	for {
		// can go through here ever gossipPeriod
		log.Println("waiting for next gossip period")
		t1 := time.Now()
		//no wait means our gossip period is too short for gossip process
		<-ticker.C
		t2 := time.Now()
		diff := t2.Sub(t1)
		log.Println("wait time:", diff)
		if float32(diff/time.Millisecond) < float32(float32(gossipPeriodMillisecond)*0.05) {
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
		cmds = parseCmds(cmds)
		log.Println("Parsed commands ", cmds)
		//execute commands
		if len(cmds) != 0 {
			for _, cmd := range cmds {
				fmt.Println(cmd)
				switch cmd {
				//if change gossip to all2all or all2all to gossip
				//change b, add command to membership list
				case CHANGE_TO_ALL2ALL:
					B = len(MembershipList)
					piggybackCommand(CHANGE_TO_ALL2ALL)
				case CHANGE_TO_GOSSIP:
					B = preservedB
					piggybackCommand(CHANGE_TO_GOSSIP)
				case JOIN_GROUP:
					joinGroup()
				case LEAVE_GROUP:
					leaveGroup()
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
		boardcastUDP()
		//update timer
		t := time.Now().UnixNano() / 1000000
		MembershipList[MyID] = Membership{t, t}
		//control timers
		//failure detect
		//deseminate failure
		//execute global commands set B
		time.Sleep(1500 * time.Millisecond)
		log.Println("Finish Gossip work")
	}
}
