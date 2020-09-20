package structs

//Membership  exported
type Membership struct {
	ID        string
	HeartBeat int64
	LocalTime int64
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

var MembershipList []Membership
var MyIP string = ""
var MyID string = ""
