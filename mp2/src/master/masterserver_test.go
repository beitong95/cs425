package master_test
import (
	"testing"
	"master"
)
func TestMasterServer(t * testing.T){
	master.File2VmMap = make(map[string] []string)
	master.File2VmMap["1.txt"] = []string{"127.0.0.1:3001", "127.0.0.1:4001","127.0.0.1:5001"}
	master.ServerRun("4321")
}