package structs

import (
	"sync"
	log "github.com/sirupsen/logrus"
)

//Membership struct 
type Membership struct {
	HeartBeat  int64 `json:"heartbeat"`
	FailedTime int64 `json:"failedtime"`
}

var IPtoVM map[string] int = map[string]int{
	    "172.22.156.12": 1,
        "172.22.158.12": 2,
        "172.22.94.12": 3,
        "172.22.156.13": 4,
        "172.22.158.13": 5,
        "172.22.94.13": 6,
        "172.22.156.14": 7,
        "172.22.158.14": 8,
        "172.22.94.14": 9,
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
// logger 
var Logger = log.New()

// variables for membershiplist
var MembershipList = make(map[string]Membership)
var MyIP string = ""
var MyPort string = ""
var MyID string = ""
var MyOldID string = ""
var MyVM string = ""

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
var CurrentProtocol bool
var IsAll2All bool
var IsGossip bool
var IsJoin bool = false

// statistics
var Bandwidth int
var LossRate int//1 5 10 15 20 five points

// channels
var C1 chan int = make(chan int, 10) // command channel between cli and udpserver
// channel for ack between udpserver and cli; 
// On receiving change protocol command, send ack to cli and print it.
var ProtocolChangeACK chan string = make(chan string) 
var UpdateGUI chan string = make(chan string) // update membership list 

// global mutex
var MT sync.Mutex //Mutex for MembershipList
var MT2 sync.Mutex //mutex for Bandwidth