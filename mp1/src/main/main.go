package main

import (
	"cli"
	"config"
	"flag"
	"fmt"
	"helper"
	"io/ioutil"
	"math"
	"os"
	"service"
	. "structs"
	"sync"
	"time"
	log "github.com/sirupsen/logrus"
)

var logger = log.New()
/**
	logger.Fatal: Program cannot execute anymore
	logger.Error: Program can continue but the final result might be wrong
	logger.Warning: Node Fail, Node Leave, Node Join
	Warning Fields:
		Reason: Fail, Join, Leave, ChangeProtocol 
		Detail:
			Fail: xxID fail; Detect Time;
			Leave: xxID leave;
			Join: xxID join;
			ChangeProtocol: change to xx protocol;
	logger.Info: Basic info logger, like MyID, Introducer IP. We can also log when a go routine starts
	logger.Debug: Detailed info like the value of a counter or something
**/
//TODO: add log helper functions in the helper package
//TODO: add log in UDPServer.go
//TODO: Debug bandwidth

func init_logger(isAppendLog bool, logLevel string) {
	/** some possible settings
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)
	// Only log the warning severity or above.
	log.SetLevel(log.WarnLevel)
	**/
	switch logLevel {
	case "info":
		logger.SetLevel(log.InfoLevel)
	case "warning":
		fmt.Println("warning")
		logger.SetLevel(log.WarnLevel)
	//Dont print log; redirect the log to /dev/null
	case "mute":
		logger.Out = ioutil.Discard
		return
	default:
		logger.SetLevel(log.DebugLevel)
	} 
	homeDir := os.Getenv("HOME")
	vmNumber := os.Getenv("VMNUMBER")
	logFileDir := homeDir + "/cs425/mp1/log/"
	if _, err := os.Stat(logFileDir); os.IsNotExist(err) {
		os.Mkdir(logFileDir, 0755)
	}
	logFileName := logFileDir + vmNumber + "_MP1.log"
	if isAppendLog == true {
		file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
		if err != nil {
			logger.Info("Failed to log to file " + logFileName + ", using default stderr")
		} else {
			logger.Out = file
		}
	} else {
		file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			logger.Info("Failed to log to file " + logFileName + ", using default stderr")
		} else {
			file.Truncate(0)
			file.Seek(0,0)
			logger.Out = file
		}
	}
}

func main() {
	//Define Flags
	isStartWithAll2AllPtr := flag.Bool("all2all", false, "start with all 2 all at the beginning")
	//Currently we dont use isIntroducer Flag
	isIntroducerPtr := flag.Bool("introducer", false, "start as an introducer")
	isMuteCliPtr := flag.Bool("muteCli", false, "mute the command line interaction")
	isSimpleCliPtr := flag.Bool("simpleCli", false, "use simple cli")
	isAppendLogPtr := flag.Bool("append", false, "append log rather than start a new log")
	configFilePtr := flag.String("config", "../../config.json", "Location of Config File")
	myPortPtr := flag.String("port", "1234", "Port used for Debug on one machine")
	logLevelPtr := flag.String("logLevel", "debug", "log level: debug, info, warning, mute")
	flag.IntVar(&Tgossip, "gossip", 300, "Gossip Period")
	flag.IntVar(&Tfail, "fail", 3300, "Fail Time")
	flag.IntVar(&Tclean, "clean", 3000, "Cleanup Time; Remove the record from the membershiplist")
	flag.IntVar(&LossRate, "loss", 1, "message loss rate 1-100")

	//Parse and save flags
	flag.Parse()

	//step1 setup logger
	isAppendLog := *isAppendLogPtr
	logLevel := *logLevelPtr
	init_logger(isAppendLog, logLevel)

	//step2 setup all flags and parameters
	Ttimeout = Tfail - Tgossip
	//Ceil C*logN*Tgossip ;C = 1
	Tall2all = (int(math.Log(float64(VMMaxCount))) + 1) * Tgossip
	MyPort = *myPortPtr
	IsAll2All = *isStartWithAll2AllPtr
	IsGossip = !(IsAll2All)
	CurrentProtocol = IsAll2All
	isIntroducer := *isIntroducerPtr
	isMuteCli := *isMuteCliPtr
	isSimpleCli := *isSimpleCliPtr
	//Setup config file env variable
	os.Setenv("CONFIG", *configFilePtr)

	//Create the first memeber in the membership list
	//ID: myIP:myPort:currentTime(Unix s)
	var err error
	MyIP, err = helper.GetLocalIP()
	if err != nil {
		logger.WithFields(log.Fields{
			"package":	"helper",
			"function":	"helper.GetLocalIP",
			"error": err,
			"data": "",
		}).Fatal("Cannot get local IP address.")
	} else {
		logger.WithFields(log.Fields{
			"package":	"helper",
			"function":	"helper.GetLocalIP",
			"res": MyIP,
		}).Info("Local IP address.")
	}
	introIP, err := config.IntroducerIPAddresses()
	if err != nil {
		logger.WithFields(log.Fields{
			"package":	"config",
			"function":	"config.IntroducerIPAddresses",
			"error": err,
			"data": "",
		}).Fatal("Cannot get Introducer IP address.")
	} else {
		logger.WithFields(log.Fields{
			"package":	"config",
			"function":	"config.IntroducerIPAddresses",
			"res": introIP,
		}).Info("Introducer IP address.")
	}
	introPort, err := config.Port()
	if err != nil {
		logger.WithFields(log.Fields{
			"package":	"config",
			"function":	"config.Port",
			"error": err,
			"data": "",
		}).Fatal("Cannot get Introducer Port.")
	} else {
		logger.WithFields(log.Fields{
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
	logger.WithFields(log.Fields{
		"ID": MyID,
		"HeartBeat": heartBeat,
	}).Info("Create Local Membership Record.")

	var wg sync.WaitGroup

	//Start UDPServer thread
	//C1 is the channel for CLI command 
	//CLI <-> C1 <-> UDPServer
	wg.Add(1)
	go service.UDPServer(IsAll2All, isIntroducer, &wg, C1)
	logger.Info("Start UDPServer go routine")

	//Start CLI
	if isMuteCli == false {
		if isSimpleCli == false {
			wg.Add(1)
			go cli.Cli(&wg, C1)
			logger.Info("Start Cli go routine")
		} else {
			wg.Add(1)
			go cli.CliSimple(&wg, C1)
			logger.Info("Start CliSimple go routine")
		}
	}
	//Wait for UDPServer and CliSimple to return
	logger.Info("Main thread waits for other threads return")
	wg.Wait()
}
