package structs

import (
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/marcusolsson/tui-go"
	"time"
	"fmt"
	"strings"
	"strconv"
)

//Membership struct
type Membership struct {
	HeartBeat  int64 `json:"heartbeat"`
	FailedTime int64 `json:"failedtime"`
}

var IPtoVM map[string]int = map[string]int{
	"172.22.156.12": 1,
	"172.22.158.12": 2,
	"172.22.94.12":  3,
	"172.22.156.13": 4,
	"172.22.158.13": 5,
	"172.22.94.13":  6,
	"172.22.156.14": 7,
	"172.22.158.14": 8,
	"172.22.94.14":  9,
	"172.22.156.15": 10,
}

const (
	CHANGE_TO_ALL2ALL         = 1
	CHANGE_TO_GOSSIP          = 2
	LIST_MEMBERSHIPLIST       = "NA"
	PRINT_SELF_ID             = "NA"
	JOIN_GROUP                = 3
	LEAVE_GROUP               = 4
	FAIL                      = "NA"
	LEAVE_GROUP_HEARTBEAT     = -2
	FAIL_HEARTBEAT            = -1
	RECEIVE_CHANGE_TO_ALL2ALL = 9
	RECEIVE_CHANGE_TO_GOSSIP  = 10
)

var Logger = log.New()

// variables for membershiplist
var MembershipList = make(map[string]Membership)
var MyIP string = ""
var MyPort string = ""
var MyID string = ""
var MyOldID string = ""
var MyVM string = ""
var IntroIP string = ""

// gossip parameter Unit:ms
var Tgossip int
var Tfail int
var Tclean int

// Ttimeout - Tgossip
var Ttimeout int
var VMMaxCount int = 10
var Tall2all int
var B int = 3

// slice and map
var Container []string
var FailedNodes map[string]int = make(map[string]int)
var LeaveNodes map[string]int = make(map[string]int)
var BroadcastAll = make(map[string]int64)
var FirstDetect = make(map[string]int64)

// 0 all2all 1 gossip
var CurrentProtocol string
var NextProtocol string
var IsJoin bool = false

// statistics
var Bandwidth int
var LossRate int //1 5 10 15 20 five points

// channels
var C1 chan int = make(chan int, 10) // command channel between cli and udpserver
// channel for ack between udpserver and cli;
// On receiving change protocol command, send ack to cli and print it.
var ProtocolChangeACK chan string = make(chan string)
var UpdateGUI chan string = make(chan string) // update membership list

// global mutex
var MT sync.Mutex  //Mutex for MembershipList
var MT2 sync.Mutex //mutex for Bandwidth

var Ack int = 0
var Master bool = false //if myself is master
var CandidateFail bool = false
var CandidateID string
var MasterIP string = ""

//mp2
/**
New nodes cannot join after master fails
Normal Nodes' CurrentStatus: Node unjoin -> Node join -> Node election -> Node join  
Default Master Node's CurrentStatus: Original Master -> Fail
New Master's CurrentStatus:  Node unjoin -> Node join -> Node election -> Master

**/
var CurrentStatus = "Node"

var History *tui.Box

func Write2Shell(text string) {
	
	if History == nil {
		// we are using simple cli
		fmt.Println(text)
	} else {
		History.Append(tui.NewHBox(
			tui.NewLabel(time.Now().Format("15:04")),
			tui.NewLabel(" "),
			tui.NewLabel(text),
			tui.NewSpacer(),
		))
	}
}
var IsMaster = false


func IP2MasterHTTPServerIP(oldIp string) string {
	ip := strings.Split(oldIp, ":")[0]
	oldPort := strings.Split(oldIp, ":")[1]
	oldPortInt, er := strconv.Atoi(oldPort)
	if er != nil {
		panic(er)
	}
	newPort := fmt.Sprintf("%v", oldPortInt + 3)
	newIP := ip + ":" +newPort
	return newIP

}

func IP2DatanodeHTTPServerIP(oldIp string) string{
	ip := strings.Split(oldIp, ":")[0]
	oldPort := strings.Split(oldIp, ":")[1]
	oldPortInt, er := strconv.Atoi(oldPort)
	if er != nil {
		panic(er)
	}
	newPort := fmt.Sprintf("%v", oldPortInt + 1)
	newIP := ip + ":" +newPort
	return newIP
}

func IP2DatanodeUploadIP(oldIp string) string{
	ip := strings.Split(oldIp, ":")[0]
	oldPort := strings.Split(oldIp, ":")[1]
	oldPortInt, er := strconv.Atoi(oldPort)
	if er != nil {
		panic(er)
	}
	newPort := fmt.Sprintf("%v", oldPortInt + 2)
	newIP := ip + ":" +newPort
	return newIP

}
var Vm2fileMap map[string][]string

var File2VmMap map[string][]string

var MW sync.Mutex

var MR sync.Mutex

var MF sync.Mutex

var MV sync.Mutex

var ReadCounter int = 0

var WriteCounter int = 0


var TimeBroadcastUDP time.Time
