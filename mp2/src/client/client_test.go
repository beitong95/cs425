package client_test
import (
	"client"
	"testing"
	"fmt"
)
func TestGetIps(t *testing.T){
	client.GetIPsFromMaster("1.txt", "127.0.0.1")

}

func TestDownloadFile(t *testing.T){
	status, err := client.DownloadFileFromDatanode("1.txt","2.txt", "localhost")
	if err != nil{
		panic(err)
	}
	fmt.Println(status)
}