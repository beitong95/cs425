package main

import (
	"flag"
	"logger"
	"client"
	"datanode"
	"master"
)


func main() {
	//Define Flags
	cliLevelPtr := flag.String("cli", "cliSimple", "cliSimple, cli")

	logLevelPtr := flag.String("logLevel", "debug", "log level: debug, info, warning, mute")
	isAppendLogPtr := flag.Bool("append", false, "append log rather than start a new log")

	//configFilePtr := flag.String("config", "../../config.json", "Location of Config File")
	//myPortPtr := flag.String("port", "1234", "Port used for Debug on one machine")

	identityPtr := flag.String("identity", "client", "identity: client, master, dataNode")

	/**
	flag.IntVar(&Tgossip, "gossip", 300, "Gossip Period")
	flag.IntVar(&Tfail, "fail", 3300, "Fail Time")
	flag.IntVar(&Tclean, "clean", 3000, "Cleanup Time; Remove the record from the membershiplist")
	flag.IntVar(&LossRate, "loss", 0, "message loss rate 1-100")
	**/
	//Parse and save flags
	flag.Parse()

	//step0 setup Logger
	logger.Init_Logger(*isAppendLogPtr, *logLevelPtr, *identityPtr)

	//step1 check identity
	logger.LogSimpleInfo("Using " + *identityPtr)

	//step2 run according to identity
	switch *identityPtr {
	case "client":
		client.Run(*cliLevelPtr)
	case "master":
		master.Run(*cliLevelPtr)
	case "dataNode":
		datanode.Run(*cliLevelPtr)
	}

	/**
	//step2 setup all flags and parameters
	Ttimeout = Tfail - Tgossip
	//Ceil C*logN*Tgossip ;C = 1
	Tall2all = (int(math.Log(float64(VMMaxCount))) + 1) * Tgossip
	MyPort = *myPortPtr
	isAll2All := *isStartWithAll2AllPtr
	if isAll2All == true {
		CurrentProtocol = "All2All"
		NextProtocol = "All2All"
	} else {
		CurrentProtocol = "Gossip"
		NextProtocol = "Gossip"
	}
	isMuteCli := *isMuteCliPtr
	isSimpleCli := *isSimpleCliPtr
	//Setup config file env variable
	os.Setenv("CONFIG", *configFilePtr)

	//Create the first memeber in the membership list
	//ID: myIP:myPort:currentTime(Unix s)
	var err error
	MyIP, err = helper.GetLocalIP()
	if err != nil {
		Logger.WithFields(log.Fields{
			"package":	"helper",
			"function":	"helper.GetLocalIP",
			"error": err,
			"data": "",
		}).Fatal("Cannot get local IP address.")
	} else {
		Logger.WithFields(log.Fields{
			"package":	"helper",
			"function":	"helper.GetLocalIP",
			"res": MyIP,
		}).Info("Local IP address.")
	}
	introIP, err := config.IntroducerIPAddresses()
	IntroIP = introIP[0]
	if err != nil {
		Logger.WithFields(log.Fields{
			"package":	"config",
			"function":	"config.IntroducerIPAddresses",
			"error": err,
			"data": "",
		}).Fatal("Cannot get Introducer IP address.")
	} else {
		Logger.WithFields(log.Fields{
			"package":	"config",
			"function":	"config.IntroducerIPAddresses",
			"res": introIP,
		}).Info("Introducer IP address.")
	}
	introPort, err := config.Port()
	if err != nil {
		Logger.WithFields(log.Fields{
			"package":	"config",
			"function":	"config.Port",
			"error": err,
			"data": "",
		}).Fatal("Cannot get Introducer Port.")
	} else {
		Logger.WithFields(log.Fields{
			"package":	"config",
			"function":	"config.Port",
			"res": introPort,
		}).Info("Introducer Port.")
	}
	if MyIP == introIP[0] && MyPort == introPort {
		IsJoin = true
	}
	millis := time.Now().UnixNano() / 1000000
	secs := millis / 1000
	MyID = MyIP + ":" + MyPort + "*" + fmt.Sprint(secs)
	heartBeat := millis
	MembershipList[MyID] = Membership{HeartBeat: heartBeat, FailedTime: -1}
	Logger.WithFields(log.Fields{
		"ID": MyID,
		"HeartBeat": heartBeat,
	}).Info("Create Local Membership Record.")

	var wg sync.WaitGroup

	//Start UDPServer thread
	//C1 is the channel for CLI command 
	//CLI <-> C1 <-> UDPServer
	wg.Add(1)
	go service.UDPServer(&wg, C1)
	Logger.Info("Start UDPServer go routine")

	//Start CLI
	if isMuteCli == false {
		if isSimpleCli == false {
			wg.Add(1)
			go cli.Cli(&wg, C1)
			Logger.Info("Start Cli go routine")
		} else {
			wg.Add(1)
			go cli.CliSimple(&wg, C1)
			Logger.Info("Start CliSimple go routine")
		}
	}
	//Wait for UDPServer and CliSimple to return
	Logger.Info("Main thread waits for other threads return")
	wg.Wait()
	*/
}
