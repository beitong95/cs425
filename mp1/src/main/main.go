package main

import (
	"cml"
	"flag"
	"fmt"
	"helper"
	"log"
	"os"
	"service"
	. "structs"
	"sync"
	"time"
)

// my membership list
var membershipList []Membership
var myIP string
var myID string

func main() {
	// it seems flag.xxx() returns pointer
	isStartWithAll2All := flag.Bool("all2all", false, "start with all 2 all at the beginning")
	isIntroducer := flag.Bool("introducer", false, "start as an introducer")
	isMuteCML := flag.Bool("mute", false, "mute the command line interaction")
	configFilePtr := flag.String("config", "./config.json", "Location of Config File")
	flag.Parse()

	all2all := *isStartWithAll2All
	introducer := *isIntroducer
	mute := *isMuteCML
	//fmt.Println(all2all, introducer, mute)
	os.Setenv("CONFIG", *configFilePtr)

	//create the first memeber(myself)
	//ID: myIP + current time
	myIP, err := helper.GetLocalIP()
	if err != nil {
		log.Fatalln("get local IP error")
	}
	millis := time.Now().UnixNano() / 1000000
	secs := millis / 1000
	myID = "*" + myIP + "_" + fmt.Sprintf("%d", secs) + "*"
	heartBeat := millis
	currentTime := millis
	membershipList = append(membershipList, Membership{myID, heartBeat, currentTime})
	helper.PrintMembershipListAsTable(membershipList)

	// actually the server and cml will forever loop until receiving a kill command
	var wg sync.WaitGroup

	//start membership udp server
	c := make(chan int64)

	wg.Add(1)
	go service.UDPServer(all2all, introducer, &wg, c)

	if mute == false {
		wg.Add(1)
		go cml.Cml(&wg, c)
	}

	wg.Wait()
}
