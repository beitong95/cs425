package main

import (
	"fmt"
)

func main() {
	myMembershipList := make([]membership, 0)
	myMembershipList = append(myMembershipList, membership{"1", 1, 1})
	fmt.Println(myMembershipList)
}
