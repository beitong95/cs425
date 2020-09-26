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
)

// my membership list

func main() {
	// it seems flag.xxx() returns pointer
	isStartWithAll2All := flag.Bool("all2all", false, "start with all 2 all at the beginning")
	isIntroducer := flag.Bool("introducer", false, "start as an introducer")
	isMuteCli := flag.Bool("mute", false, "mute the command line interaction")
	isVerbose := flag.Bool("v", false, "print log")
	configFilePtr := flag.String("config", "../../config.json", "Location of Config File")
	MyPortPtr := flag.String("port", "1234", "Port used for Debug")
	// parameters
	flag.IntVar(&Tgossip, "gossip", 300, "Gossip Period")
	flag.IntVar(&Tfail, "fail", 3000, "Fail Time")
	flag.IntVar(&Tclean, "clean", 3000, "Cleanup Time")

	flag.Parse()
	//parameter
	Ttimeout = Tfail - Tgossip
	/**
	fmt.Println(Tgossip)
	fmt.Println(Tfail)
	fmt.Println(Tclean)
	fmt.Println(Ttimeout)
	**/
	MyPort = *MyPortPtr
	fmt.Printf("Using Port: %s\n", MyPort)
	all2all := *isStartWithAll2All
	introducer := *isIntroducer
	mute := *isMuteCli
	verbose := *isVerbose
	if !verbose {
		log.SetOutput(ioutil.Discard)
	}
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
	MyID = MyIP + ":" + MyPort + "*" + fmt.Sprint(secs)
	heartBeat := millis
	//currentTime := millis
	//MembershipList = append(MembershipList, Membership{MyID, heartBeat, currentTime})
	// change to map 09222020
	MembershipList[MyID] = Membership{HeartBeat: heartBeat, LocalTime: -1}
	// MembershipList["test"] = Membership{111, 111} //test table

	// test
	// fmt.Println("map based membershiplist", MembershipList)

	//fmt.Println("cannot print table now, TODO PrintMembershipListAsTableFromMap")
	service.MT.Lock()
	helper.PrintMembershipListAsTable(MembershipList)
	service.MT.Unlock()

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
