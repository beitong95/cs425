package helper

import (
	. "structs"
)

//merge recievedMemberShipList and the node memberShipList
//modify MemberShipList(global var in structs.go)
//e.g.{"127.0.0.1:5000":{"h":1000,"t":5000}} {"127.0.0.1:6000":{"h":1234,"t":5678}} => {"127.0.0.1:5000":{"h":1000,"t":5000}, "127.0.0.1:6000":{"h":1234,"t":5678}}
//If members have the same heartbeat, then update the time of these members to the latest time, coresponding to the time in membershiplist struct.
func memberShipList(recievedMemberShipList map[string]MemberShip) {
	for key, receivedMembership := range recievedMemberShipList {
		if existedMembership, ok := MembershipList[key]; ok {
			if existedMembership.LocalTime < receivedMembership.LocalTime {
				MembershipList[key] = receivedMembership
			}
		} else {
			MembershipList[key] = receivedMembership
		}
	}
}
