package master_test

import (
	"fmt"
	"master"
	"testing"
)

func TestMasterServer(t *testing.T) {
	master.File2VmMap = make(map[string][]string)
	master.File2VmMap["1.txt"] = []string{"127.0.0.1:3001", "127.0.0.1:4001", "127.0.0.1:5001"}
	//master.ServerRun("4321")
}

func TestMaster(t *testing.T) {
	// master.Vm2fileMap = map[string][]string{"1":[]string{},"2":[]string{},"3":[]string{},"4":[]string{},"5":[]string{},"6":[]string{},"7":[]string{},"8":[]string{},"9":[]string{},"10":[]string{}}
	fmt.Println("test Master")
	master.Vm2fileMap = map[string][]string{}
	master.File2VmMap = map[string][]string{}
	master.Recover("1", []string{"steam", "wechat", "qq"})
	master.Recover("2", []string{"steam", "wechat", "qq", "snapchat"})
	master.Recover("3", []string{"steam", "wechat", "qq", "linkedin"})
	master.Recover("4", []string{"steam", "wechat"})
	master.Recover("5", []string{"snapchat"})
	master.Recover("6", []string{"linkedin", "cs425"})
	master.Recover("7", []string{"cs425", "snapchat"})
	master.Recover("8", []string{"cs425", "linkedin", "snapchat"})
	master.Recover("9", []string{"cs425", "linkedin"})
	master.Recover("10", []string{"qq"})
	printmap()
	//fmt.Println("here")
	return
}
