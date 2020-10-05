package helper

import (
	log "github.com/sirupsen/logrus"
	. "structs"
	"strings"
	"fmt"
)

func ConvertIDtoVM(ID string) string{
	ip := strings.Split(ID, ":")[0]
	vmNumber := IPtoVM[ip]
	return "vm" + fmt.Sprintf("%02d",vmNumber)
}

// reason is the first level key: fail; leave; join; protocol
// msg is the second level key: detailed info. exp: detect node leave

// how to grep
// 1. grep fail or Detect Node Fail; return all log related to failure detection
// 2. grep the fail detection info for a specific failed node: grep fail_vm=vm#
// 3. grep the fail detection infor from a specific node: grep current_vm=vm#
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