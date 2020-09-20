package main

import (
	"cli"
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

func main() {
	// it seems flag.xxx() returns pointer
	isStartWithAll2All := flag.Bool("all2all", false, "start with all 2 all at the beginning")
	isIntroducer := flag.Bool("introducer", false, "start as an introducer")
	isMuteCli := flag.Bool("mute", false, "mute the command line interaction")
	configFilePtr := flag.String("config", "./config.json", "Location of Config File")
	flag.Parse()

	all2all := *isStartWithAll2All
	introducer := *isIntroducer
	mute := *isMuteCli
	//fmt.Println(all2all, introducer, mute)
	os.Setenv("CONFIG", *configFilePtr)

	//create the first memeber(myself)
	//ID: myIP + current time
	var err error
	MyIP, err = helper.GetLocalIP()
	if err != nil {
		log.Fatalln("get local IP error")
	}
	millis := time.Now().UnixNano() / 1000000
	secs := millis / 1000
	MyID = "*" + MyIP + "_" + fmt.Sprintf("%d", secs) + "*"
	heartBeat := millis
	currentTime := millis
	MembershipList = append(MembershipList, Membership{MyID, heartBeat, currentTime})
	helper.PrintMembershipListAsTable(MembershipList)

	// actually the server and cli will forever loop until receiving a kill command
	var wg sync.WaitGroup

	//start membership udp server
	//10 is enough for the channel buffer capacity
	c := make(chan int, 10)

	wg.Add(1)
	go service.UDPServer(all2all, introducer, &wg, c)

	if mute == false {
		wg.Add(1)
		go cli.Cli(&wg, c)
	}

	wg.Wait()
}
