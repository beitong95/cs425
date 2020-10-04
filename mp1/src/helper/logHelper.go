package helper

import (
	log "github.com/sirupsen/logrus"
	. "structs"
	"strings"
)

func convertIDtoVM(ID string) string{
	ip := strings.Split(ID, ":")[0]
	vmNumber := IPtoVM[ip]
	return "vm" + fmt.Sprintf("%02d",vmNumber)
}

// how to grep
// 1. grep fail or Detect Node Fail; return all log related to failure detection
// 2. grep the fail detection info for a specific failed node: grep fail vm=vm#
// 3. grep the fail detection infor from a specific node: grep current vm=vm#
func LogFail(logger *log.Logger, VM string, ID string, heartbeat int64, detectTime int64) {
	logger.WithFields(log.Fields{
		"reason":	"fail",
		"current vm": VM,
		"fail id": ID,
		"fail vm": convertIDtoVM(ID),
		"last heartbeat": heartbeat,
		"detect time": detectTime,
	}).Warn("Detect Node Fail")
}