package main

import (
	"cli"
	"config"
	"datanode"
	"flag"
	"fmt"
	"helper"
	"io/ioutil"
	"master"
	"math"
	"os"
	"service"
	. "structs"
	"sync"
	"time"
	"constant"
	"strconv"

	log "github.com/sirupsen/logrus"
)

/**
	Logger.Fatal: Program cannot execute anymore
	Logger.Error: Program can continue but the final result might be wrong
	Logger.Warning: Node Fail, Node Leave, Node Join
	Warning Fields:
		Reason: Fail, Join, Leave, ChangeProtocol
		Detail:
			Fail: check helper/logHelper.go
			Leave: check helper/logHelper.go
			Join: check helper/logHelper.go
			ChangeProtocol: check helper/logHelper.go
	Logger.Info: Basic info Logger, like MyID, Introducer IP. We can also log when a go routine starts
	Logger.Debug: Detailed info like the value of a counter or something
**/

func init_Logger(isAppendLog bool, logLevel string) {
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
		Logger.SetLevel(log.InfoLevel)
	case "warning":
		fmt.Println("warning")
		Logger.SetLevel(log.WarnLevel)
	//Dont print log; redirect the log to /dev/null
	case "mute":
		Logger.Out = ioutil.Discard
		return
	default:
		Logger.SetLevel(log.DebugLevel)
	}
	homeDir := os.Getenv("HOME")
	vmNumber := os.Getenv("VMNUMBER")
	MyVM = vmNumber
	logFileDir := homeDir + "/cs425/mp2-revised/log/"
	if _, err := os.Stat(logFileDir); os.IsNotExist(err) {
		os.Mkdir(logFileDir, 0755)
	}
	logFileName := logFileDir + vmNumber + "_MP2.log"
	if isAppendLog == true {
		file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
		if err != nil {
			Logger.Info("Failed to log to file " + logFileName + ", using default stderr")
		} else {
			Logger.Out = file
		}
	} else {
		file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			Logger.Info("Failed to log to file " + logFileName + ", using default stderr")
		} else {
			file.Truncate(0)
			file.Seek(0, 0)
			Logger.Out = file
		}
	}
}
func checkIfMasterThenRunMasterLogic() {
	for {
		if IsMaster {
			go master.Run()
			break
		}
	}
}

func main() {
	//Define Flags
	isStartWithAll2AllPtr := flag.Bool("all2all", false,  "start with all 2 all at the beginning")
	isMuteCliPtr := flag.Bool("muteCli", false, "mute the command line interaction")
	isSimpleCliPtr := flag.Bool("simpleCli", false, "use simple cli")
	isAppendLogPtr := flag.Bool("append", false, "append log rather than start a new log")
	configFilePtr := flag.String("config", "../../config.json", "Location of Config File")
	myPortPtr := flag.String("port", "1234", "Port used for Debug on one machine")
	logLevelPtr := flag.String("logLevel", "debug", "log level: debug, info, warning, mute")
	dirPtr := flag.String("dir", "", "dir for saving files")
	flag.IntVar(&Tgossip, "gossip", 300, "Gossip Period")
	flag.IntVar(&Tfail, "fail", 5000, "Fail Time")
	flag.IntVar(&Tclean, "clean", 3000, "Cleanup Time; Remove the record from the membershiplist")
	flag.IntVar(&LossRate, "loss", 0, "message loss rate 1-100")
	//Parse and save flags
	flag.Parse()

	//step1 setup Logger
	isAppendLog := *isAppendLogPtr
	logLevel := *logLevelPtr
	init_Logger(isAppendLog, logLevel)
	//if master master.ServerRun(myPort)

	//step2 setup all flags and parameters
	if *dirPtr != "" {
		constant.Dir = *dirPtr
	}
	Ttimeout = Tfail - Tgossip
	//Ceil C*logN*Tgossip ;C = 1
	Tall2all = (int(math.Log(float64(VMMaxCount))) + 1) * Tgossip
	MyPort = *myPortPtr

	//setup local test ports
	//Delete for released version
	MyPortInt, er := strconv.Atoi(MyPort)
	if er != nil {
		panic(er)
	}
	constant.MasterHTTPServerPort = fmt.Sprint(int(MyPortInt) + 3)
	constant.DatanodeHTTPServerPort = fmt.Sprint(int(MyPortInt) + 1)
	constant.DatanodeHTTPServerUploadPort = fmt.Sprint(int(MyPortInt) + 2)
	//Delete for released version


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

	//step3 Create the first memeber in the membership list
	//ID: myIP:myPort:currentTime(Unix s)
	var err error
	MyIP, err = helper.GetLocalIP()
	if err != nil {
		panic(err)
	}
	introIP, err := config.IntroducerIPAddresses()
	if err != nil {
		panic(err)
	}
	IntroIP = introIP[0] // 172.22.156.12
	introPort, err := config.Port() // 1234
	if err != nil {
		panic(err)
	}
	if MyIP == introIP[0] && MyPort == introPort {
		IsMaster = true
		IsJoin = true
		CurrentStatus = "Original Master" // In this state, the master has join the group and acts as the introducer
	}
	millis := time.Now().UnixNano() / 1000000
	secs := millis / 1000
	MyID = MyIP + ":" + MyPort + "*" + fmt.Sprint(secs)
	//test locally
	MasterIP = IntroIP + ":1234"
	heartBeat := millis
	MembershipList[MyID] = Membership{HeartBeat: heartBeat, FailedTime: -1}
	Vm2fileMap = make(map[string][]string)
	File2VmMap = make(map[string][]string)
	if IsMaster == true {
		MV.Lock()
		Vm2fileMap[MasterIP] = []string{}
		Logger.Info(Vm2fileMap)
		MV.Unlock()
	}


	//step4 Start UDPServer thread
	//C1 is the channel for CLI command
	//CLI <-> C1 <-> UDPServer
	var wg sync.WaitGroup
	wg.Add(1)
	go checkIfMasterThenRunMasterLogic()
	go datanode.Run()
	go service.UDPServer(&wg, C1)
	Logger.Info("Start UDPServer go routine")

	//step5 Start CLI
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
}
