package main

import (
	"cli"
	"config"
	"flag"
	"fmt"
	"helper"
	"io/ioutil"
	"log"
	"math"
	"os"
	"service"
	. "structs"
	"sync"
	"time"
)

func main() {
	//Define Flags
	isStartWithAll2AllPtr := flag.Bool("all2all", false, "start with all 2 all at the beginning")
	//Currently we dont use isIntroducer Flag
	isIntroducerPtr := flag.Bool("introducer", false, "start as an introducer")
	isMuteCliPtr := flag.Bool("mute", false, "mute the command line interaction")
	isSimpleCliPtr := flag.Bool("simpleCli", false, "use simple cli")
	isVerbosePtr := flag.Bool("v", false, "print log")
	configFilePtr := flag.String("config", "../../config.json", "Location of Config File")
	myPortPtr := flag.String("port", "1234", "Port used for Debug on one machine")
	flag.IntVar(&Tgossip, "gossip", 300, "Gossip Period")
	flag.IntVar(&Tfail, "fail", 3300, "Fail Time")
	flag.IntVar(&Tclean, "clean", 3000, "Cleanup Time")
	flag.IntVar(&LossRate, "loss", 1, "message loss rate 1-100")

	//Parse and save flags
	flag.Parse()
	Ttimeout = Tfail - Tgossip
	//Ceil C*logN*Tgossip C = 1
	Tall2all = (int(math.Log(float64(VMMaxCount))) + 1) * Tgossip
	MyPort = *myPortPtr
	IsAll2All = *isStartWithAll2AllPtr
	IsGossip = !(IsAll2All)
	CurrentProtocol = IsAll2All
	isIntroducer := *isIntroducerPtr
	isMuteCli := *isMuteCliPtr
	isSimpleCli := *isSimpleCliPtr
	isVerbose := *isVerbosePtr
	//Dont print log; redirect the log to /dev/null
	if !isVerbose {
		log.SetOutput(ioutil.Discard)
	}
	//Setup config file env variable
	os.Setenv("CONFIG", *configFilePtr)

	//Create the first memeber in the membership list
	//ID: myIP:myPort:currentTime(Unix s)
	var err error
	MyIP, err = helper.GetLocalIP()
	if err != nil {
		log.Fatalln("get local IP error")
	}
	introIP, err := config.IntroducerIPAddresses()
	if err != nil {
		panic(err)
	}
	introPort, err := config.Port()
	if err != nil {
		panic(err)
	}
	if MyIP == introIP[0] && MyPort == introPort {
		IsJoin = true
	}
	millis := time.Now().UnixNano() / 1000000
	secs := millis / 1000
	MyID = MyIP + ":" + MyPort + "*" + fmt.Sprint(secs)
	heartBeat := millis
	MembershipList[MyID] = Membership{HeartBeat: heartBeat, FailedTime: -1}

	var wg sync.WaitGroup

	//Start UDPServer thread
	//C1 is the channel for CLI command 
	//CLI <-> C1 <-> UDPServer
	wg.Add(1)
	go service.UDPServer(IsAll2All, isIntroducer, &wg, C1)

	//Start CLI
	if isMuteCli == false {
		if isSimpleCli == false {
			wg.Add(1)
			go cli.Cli(&wg, C1)
		} else {
			wg.Add(1)
			go cli.CliSimple(&wg, C1)
		}
	}
	//Wait for UDPServer and CliSimple to return
	wg.Wait()
}
