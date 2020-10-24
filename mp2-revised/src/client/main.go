package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"networking"
	"os"
	"strings"
	. "structs"
	"time"
)

/**
 Finished parts:
 1. connect to master
 2. detect master fail
 3. reconnect master if master fails
 4. kick out prompt
 5. kick out and rejoin

 TODO:
 1. get
 2. put
 3. abort current command when master fails
 4. resend current command if current command fails
 5. command queue or command mutual exclusion
 ( command queue: allow user input multi commands in a short time.
   command mutual exclusion: user cannot type new command until current command finishs)
**/
func DownloadFileFromDatanode(filename string, localfilename string, ipPort string) (string, error) {
	url := "http://" + ipPort + "/" + filename
	fmt.Println(url)
	rsp, err := http.Get(url)
	if err != nil {
		return "Connection error", err
	}
	if rsp.Header["Content-Length"][0] == "19" {
		fmt.Println("Possible empty")
		buffer := make([]byte, 19)
		rsp.Body.Read(buffer)
		if string(buffer) == "404 page not found\n" {
			return "File not found", errors.New("networking: file not found")
		} else {
			file := strings.NewReader(string(buffer))
			destFile, err := os.Create("./" + localfilename)
			if err != nil {
				log.Printf("Create file failed: %s\n", err)
				return "Create Failed", err
			}
			_, err = io.Copy(destFile, file)
			if err != nil {
				log.Printf("Write file failed: %s\n", err)
				return "Write error", err
			}
			return "OK", nil
		}
	}
	destFile, err := os.Create("./" + localfilename)
	if err != nil {
		log.Printf("Create file failed: %s\n", err)
		return "Create Failed", err
	}
	_, err = io.Copy(destFile, rsp.Body)
	if err != nil {
		log.Printf("Write file failed: %s\n", err)
		return "Write error", err
	}
	return "OK", nil
}

func GetIPsFromMaster(filename string) ([]string, error) {
	url := "http://" + MasterIP + "/getips?file=" + filename
	body := networking.HTTPsend(url)
	var ipList []string
	err := json.Unmarshal([]byte(body), &ipList)
	if err != nil {
		return []string{}, err
	}
	fmt.Println(ipList)
	return ipList, nil
}

func GetIPsPutFromMaster(filename string) ([]string, error) {
	url := "http://" + MasterIP + "/getipsput?file=" + filename
	body := networking.HTTPsend(url)
	var ipList []string
	err := json.Unmarshal([]byte(body), &ipList)
	if err != nil {
		return []string{}, err
	}
	fmt.Println(ipList)
	return ipList, nil
}

func GetFile(filename string, localfilename string) {
	ID := fmt.Sprint(time.Now().UnixNano())
	// my ip + my port + current time
	url := "http://" + MasterIP + "/get?id=" + ID + "&file=" + filename 
	go networking.HTTPsend(url)
	IPs, err := GetIPsFromMaster(filename)
	if len(IPs) == 0 {
		url = "http://" + MasterIP + "/clientBad?id=" + ID
		networking.HTTPsend(url)
	}
	fmt.Println(IPs)
	if err != nil {
		panic(err)
	}
	for _, ip := range IPs {
		status, _ := DownloadFileFromDatanode(filename, localfilename, ip)
		if status == "OK" {
			url = "http://" + MasterIP + "/clientACK?id=" + ID
			networking.HTTPsend(url)
			return
		}
	}
	// command end
}

func PutFile(filename string, remotefilename string) {
	ID := fmt.Sprint(time.Now().UnixNano())
	url := "http://" + MasterIP + "/put?id=" + ID
	go networking.HTTPsend(url)
	IPs, err := GetIPsPutFromMaster(filename)
	if err != nil {
		panic(err)
	}
	// if len(IPs) == 0 {
	// 	url = "http://" + MasterIP + ":" + constant.HTTPportClient2Master + "/clientBad?id=" + ID
	// 	networking.HTTPsend(url)
	// }
	fmt.Println(IPs)
	// if err != nil {
	// 	panic(err)
	// }
	// for _, ip := range IPs {
	// 	status, _ := DownloadFileFromDatanode(filename, localfilename, ip)
	// 	if status == "OK" {
	// 		url = "http://" + MasterIP + ":" + constant.HTTPportClient2Master + "/clientACK?id=" + ID
	// 		networking.HTTPsend(url)
	// 		return
	// 	}
	// }
}

// func UpdateFile(filename string, MasterIP string) {
// 	IPs, err := getDestnationFromMaster(filename, MasterIP)
// 	for _,v := range IPs {
// 		err := networking.FTPsend(filename, v)
// 	}
// 	// wait for master's ACK

// }

// func DeleteFile(filename string, MasterIP string) {
// 	IPs, err := getDestnationFromMaster(filename, MasterIP)
// 	for _,v := range IPs {
// 		err := networking.FTPsend(filename, v)
// 	}
// 	// wait for master's ACK
// }

// func LsFile() {

// }

// func StoreFile() {

// }
