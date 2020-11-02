package client_test

import (
	"client"
	"fmt"
	"structs"
	"testing"
)

func TestGetIps(t *testing.T) {
	client.GetIPsFromMaster("1.txt")

}

func TestDownloadFile(t *testing.T) {
	status, err := client.DownloadFileFromDatanode("1.txt", "2.txt", "localhost")
	if err != nil {
		panic(err)
	}
	fmt.Println(status)
	status, err = client.DownloadFileFromDatanode("2.txt", "joke", "localhost")
	if err != nil {
		panic(err)
	}
	fmt.Println(status)
	status, err = client.DownloadFileFromDatanode("dsad", "1", "localhost")
	if err != nil {
		panic(err)
	}
	fmt.Println(status)
}

func TestGetFile(t *testing.T) {
	structs.MasterIP = "10.180.129.247:1234"
	go client.GetFile("1.txt", "22")
	go client.GetFile("1.txt", "33")
	client.PutFile("joke", "joke.test")
}
