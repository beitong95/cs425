package logger

import (
	log "github.com/sirupsen/logrus"
	"fmt"
	"os"
	"io/ioutil"
)
var Logger = log.New()

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

func Init_Logger(isAppendLog bool, logLevel string, identity string) {
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

	//set log path
	homeDir := os.Getenv("HOME")
	vmNumber := os.Getenv("VMNUMBER")
	logFileDir := homeDir + "/cs425/mp2/log/"
	if _, err := os.Stat(logFileDir); os.IsNotExist(err) {
		os.Mkdir(logFileDir, 0755)
	}
	logFileName := logFileDir + vmNumber + "_" + identity + "_MP2.log"

	//open log file
	if isAppendLog == true {
		file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			Logger.Info("Failed to log to file " + logFileName + ", using default stderr")
		} else {
			Logger.Out = file
		}
	} else {
		file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			Logger.Info("Failed to log to file " + logFileName + ", using default stderr")
		} else{
			Logger.Out = file
		} 
	}
}
/**
func ConvertIDtoVM(ID string) string{
	ip := string.Split(ID, ":")[0]
	vmNumber := IPtoVM[ip]
	return "vm" + fmt.Sprintf("%02d",vmNumber)
}
**/
// reason is the first level key: fail; leave; join; protocol
// msg is the second level key: detailed info. exp: detect node leave

// how to grep
// 1. grep fail or Detect Node Fail; return all log related to failure detection
// 2. grep the fail detection info for a specific failed node: grep fail_vm=vm#
// 3. grep the fail detection infor from a specific node: grep current_vm=vm#

// self-defined wrapper
func LogSimpleFatal(reason string) {
	Logger.Fatal(reason)
}

func LogSimpleInfo(reason string) {
	Logger.Info(reason)
}

func LogSimpleError(reason string) {
	Logger.Error(reason)
}
/**

func LogFail(logger *log.Logger, VM string, ID string, heartbeat int64, detectTime int64) {
	logger.WithFields(log.Fields{
		"reason":	"fail",
		"current_vm": VM,
		"fail_id": ID,
		"fail_vm": ConvertIDtoVM(ID),
		"last_heartbeat": heartbeat,
		"detect_time": detectTime,
	}).Warn("Detect Node Fail")
}

func LogLeaver(logger *log.Logger, VM string, ID string) {
	logger.WithFields(log.Fields{
		"reason":	"leave",
		"current_vm": VM,
		"leave_id": ID,
		"leave_vm": ConvertIDtoVM(ID),
	}).Warn("Detect Node Leave")
}

func LogSelfLeave(logger *log.Logger, VM string, ID string) {
	logger.WithFields(log.Fields{
		"reason":	"leave",
		"current_vm": VM,
		"leave_id": ID,
	}).Warn("Node Leave Group")
}

func LogJoiner(logger *log.Logger, VM string, ID string) {
	logger.WithFields(log.Fields{
		"reason":	"join",
		"current_vm": VM,
		"join_id": ID,
		"join_vm": ConvertIDtoVM(ID),
	}).Warn("Detect Node Join")
}

func LogSelfJoin(logger *log.Logger, VM string, ID string) {
	logger.WithFields(log.Fields{
		"reason":	"join",
		"current_vm": VM,
		"join_id": ID,
	}).Warn("Node Join Group")
}

func LogReceiveProtocol(logger *log.Logger, VM string, ID string, receive_protocol string) {
	logger.WithFields(log.Fields{
		"reason":	"protocol",
		"current_vm": VM,
		"id": ID,
		"receive_protocol": receive_protocol,
	}).Warn("Receive Protocol Change Request")
}

func LogChangeProtocol(logger *log.Logger, VM string, ID string, previous_protocol string, current_protocol string) {
	logger.WithFields(log.Fields{
		"reason":	"protocol",
		"current_vm": VM,
		"id": ID,
		"previous_protocol": previous_protocol,
		"current_protocol": current_protocol,
	}).Warn("Change Protocol")
}
**/