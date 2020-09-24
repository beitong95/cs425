package structs

//Membership  exported
type Membership struct {
	HeartBeat int64 `json:"heartbeat"`
	LocalTime int64 `json:"localtime"`
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

var B int
var Container []string
