package datanode_test
import (
	"datanode"
	"testing"
)
func TestDataNode(t *testing.T){
	datanode.ServerRun("3000")
}