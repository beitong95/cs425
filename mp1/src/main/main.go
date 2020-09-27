package main

import (
	"cli"
	"flag"
	"fmt"
	"helper"
	"io/ioutil"
	"log"
	"os"
	"service"
	. "structs"
	"sync"
	"time"
	"math"
)

// my membership list

func main() {
	// flags
	isStartWithAll2AllPtr := flag.Bool("all2all", false, "start with all 2 all at the beginning")
	// actually we dont need that, we config introducer in config.json
	isIntroducerPtr := flag.Bool("introducer", false, "start as an introducer")
	isMuteCliPtr := flag.Bool("mute", false, "mute the command line interaction")
	isVerbosePtr := flag.Bool("v", false, "print log")
	configFilePtr := flag.String("config", "../../config.json", "Location of Config File")
	myPortPtr := flag.String("port", "1234", "Port used for Debug on one machine")
	// parameters
	flag.IntVar(&Tgossip, "gossip", 300, "Gossip Period")
	flag.IntVar(&Tfail, "fail", 4800, "Fail Time")
	flag.IntVar(&Tclean, "clean", 3000, "Cleanup Time")

	// parse and save flags
	flag.Parse()
	Ttimeout = Tfail - Tgossip
	Tall2all = int(math.Log(float64(VMMaxCount))* float64(Tgossip))
	MyPort = *myPortPtr
	//fmt.Printf("Using Port: %s\n", MyPort)
	IsAll2All = *isStartWithAll2AllPtr
	IsGossip = !(IsAll2All)
	isIntroducer := *isIntroducerPtr
	isMuteCli := *isMuteCliPtr
	isVerbose := *isVerbosePtr
	if !isVerbose {
		log.SetOutput(ioutil.Discard)
	}
	os.Setenv("CONFIG", *configFilePtr)

	//create the first memeber
	//ID: myIP:myPort:currentTime(Unix s)
	var err error
	MyIP, err = helper.GetLocalIP()
	if err != nil {
		log.Fatalln("get local IP error")
	}
	millis := time.Now().UnixNano() / 1000000
	secs := millis / 1000
	MyID = MyIP + ":" + MyPort + "*" + fmt.Sprint(secs)
	heartBeat := millis
	MembershipList[MyID] = Membership{HeartBeat: heartBeat, FailedTime: -1}
	// MembershipList["test"] = Membership{111, 111} //test table

	// test
	// fmt.Println("map based membershiplist", MembershipList)
	/**
	i add the lock inside the print function
	service.MT.Lock()
	helper.PrintMembershipListAsTable(MembershipList)
	service.MT.Unlock()
	**/
	// actually the server and cli will forever loop until receiving a kill command
	var wg sync.WaitGroup

	//start membership udp server
	//10 is enough for the channel buffer capacity

	wg.Add(1)
	go service.UDPServer(IsAll2All, isIntroducer, &wg, C)

	if isMuteCli == false {
		wg.Add(1)
		go cli.Cli(&wg, C)
	} else {
		wg.Add(1)
		go cli.CliSimple(&wg, C)
	}

	wg.Wait()
}
