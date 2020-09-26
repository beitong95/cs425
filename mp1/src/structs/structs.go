package structs

import "sync"

//Membership  exported
type Membership struct {
	HeartBeat  int64 `json:"heartbeat"`
	FailedTime int64 `json:"failedtime"`
}

const (
	CHANGE_TO_ALL2ALL     = 1
	CHANGE_TO_GOSSIP      = 2
	LIST_MEMBERSHIPLIST   = "NA"
	PRINT_SELF_ID         = "NA"
	JOIN_GROUP            = 3
	LEAVE_GROUP           = 4
	FAIL                  = "NA"
	LEAVE_GROUP_HEARTBEAT = -2
	FAIL_HEARTBEAT        = -1
)

var MembershipList = make(map[string]Membership)
var MyIP string = ""
var MyPort string = ""
var MyID string = ""

// ms
var Tgossip int
var Tfail int
var Tclean int

// Ttimeout - Tgossip
var Ttimeout int

var B int = 3
var Container []string
var FailedNodes map[string]int = make(map[string]int)
var MT sync.Mutex //lock for global variable MembershipList
var UpdateGUI chan string = make(chan string)
var IsAll2All bool
var IsGossip bool
var BroadcastAll = make(map[string]int64)
var FirstDetect = make(map[string]int64)
