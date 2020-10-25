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
	Write2Shell("Downloading file from: " + url)
	rsp, err := http.Get(url)
	Write2Shell("get http.Get return")
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
			// store in main folder
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
	// store in main folder
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
	getFailFlag := true
	//step 1. get id and create url
	ID := fmt.Sprint(time.Now().UnixNano())
	newIP := IP2MasterHTTPServerIP(MasterIP)
	url := "http://" + newIP + "/get?id=" + ID + "&file=" + filename
	Write2Shell("Getfile url: " + url)

	//step 2. send url and decode body
	body := networking.HTTPsend(url)
	Write2Shell(string(body))
	var IPs []string
	IPs = []string{}
	err := json.Unmarshal(body, &IPs)
	if err != nil {
		Write2Shell("Unmarshal error in GetFile")
	}
	Write2Shell("Received IPs: ")
	for _, v := range IPs {
		Write2Shell(v)
	}

	//step3. download files from the list
	for _, ip := range IPs {
		//ip: ip + udpPort  -> newIp: ip + datanodeHTTPServerPort
		newIp := IP2DatanodeHTTPServerIP(ip)
		status, _ := DownloadFileFromDatanode(filename, localfilename, newIp)
		if status == "OK" {
			getFailFlag = false
			url = "http://" + newIP + "/clientACK?id=" + ID
			networking.HTTPsend(url)
			break
		}
	}

	//step4. check if get successes, print the reuslt to shell.
	// The user can resend the command manully.
	if getFailFlag == true {
		url = "http://" + newIP + "/clientBad?id=" + ID
		networking.HTTPsend(url)
		Write2Shell("Get " + filename + " " + localfilename + " id: " + ID + " Fail")
	} else {
		Write2Shell("Get " + filename + " " + localfilename + " id: " + ID + " Success")
	}
}

func UploadFileToDatanode(filename string, remotefilename string, ipPort string) string {
	url := "http://" + ipPort + "/putfile"
	Write2Shell("Upload file to url:" + url)
	body := networking.HTTPuploadFile(url, filename, remotefilename)
	Write2Shell("Url: " + url + " Status: " + string(body))
	return string(body)
}

func DeleteFileFromDatanode(remotefilename string, ipPort string) string {
	url := "http://" + ipPort + "/deletefile?file=" + remotefilename
	Write2Shell("Delete file from url:" + url)
	body := networking.HTTPsend(url)
	Write2Shell("Url: " + url + " Status: " + string(body))
	return string(body)
}

func PutFile(filename string, remotefilename string) {
	putFailFlag := true
	//step 1. get id and create url
	ID := fmt.Sprint(time.Now().UnixNano())
	newIP := IP2MasterHTTPServerIP(MasterIP)
	url := "http://" + newIP + "/put?id=" + ID + "&file=" + remotefilename
	Write2Shell("Getfile url: " + url)

	//step 2. send url and decode body
	body := networking.HTTPsend(url)
	Write2Shell(string(body))
	var IPs []string
	IPs = []string{}
	err := json.Unmarshal(body, &IPs)
	if err != nil {
		Write2Shell("Unmarshal error in PutFile")
	}

	if len(IPs) == 0 {
		url = "http://" + newIP + "/clientBad?id=" + ID
		networking.HTTPsend(url)
		Write2Shell("Put" + filename + " " + remotefilename + " id: " + ID + " Fail")
		Write2Shell("Reason: IPs lenght == 0")
		return
	}

	Write2Shell("Received IPs: ")
	// should always return 4 ips`
	for _, v := range IPs {
		Write2Shell(v)
	}

	//step 3. upload files to vms in the list
	successCounter := 0
	failedIPs := []string{}
	for _, ip := range IPs {
		//ip: ip + udpPort  -> newIp: ip + datanodeHTTPServerPort
		destinationIp := IP2DatanodeUploadIP(ip)
		status := UploadFileToDatanode(filename, remotefilename, destinationIp)
		if status == "OK" {
			successCounter++
		} else {
			failedIPs = append(failedIPs, destinationIp)
		}
	}
	if successCounter == len(IPs) {
		putFailFlag = false
	}

	//step4. check if put successes, print the reuslt to shell.
	// The user can resend the command manully.
	if putFailFlag == true {
		url = "http://" + newIP + "/clientBad?id=" + ID
		networking.HTTPsend(url)
		Write2Shell("Put" + filename + " " + remotefilename + " id: " + ID + " Fail")
		Write2Shell("Failed destination IPs:")
		for _, v := range failedIPs {
			Write2Shell(v)
		}
	} else {
		url = "http://" + newIP + "/clientACK?id=" + ID
		networking.HTTPsend(url)
		Write2Shell("Put" + filename + " " + remotefilename + " id: " + ID + " Success")
	}

}

func DeleteFile(remotefilename string) {
	putFailFlag := true
	//step 1. get id and create url
	ID := fmt.Sprint(time.Now().UnixNano())
	newIP := IP2MasterHTTPServerIP(MasterIP)
	url := "http://" + newIP + "/delete?id=" + ID + "&file=" + remotefilename
	Write2Shell("Deletefile url: " + url)

	//step 2. send url and decode body
	body := networking.HTTPsend(url)
	Write2Shell(string(body))
	var IPs []string
	IPs = []string{}
	err := json.Unmarshal(body, &IPs)
	if err != nil {
		Write2Shell("Unmarshal error in DeleteFile")
	}

	if len(IPs) == 0 {
		url = "http://" + newIP + "/clientBad?id=" + ID
		networking.HTTPsend(url)
		Write2Shell("Delete" + remotefilename + " id: " + ID + " Fail")
		Write2Shell("Reason: IPs lenght == 0")
		return
	}

	Write2Shell("Received IPs: ")
	// should always return 4 ips`
	for _, v := range IPs {
		Write2Shell(v)
	}

	//step 3. delete files to vms in the list
	successCounter := 0
	failedIPs := []string{}
	for _, ip := range IPs {
		//ip: ip + udpPort  -> newIp: ip + datanodeHTTPServerPort
		destinationIp := IP2DatanodeUploadIP(ip)
		status := DeleteFileFromDatanode(remotefilename, destinationIp)
		if status == "OK" {
			successCounter++
		} else {
			failedIPs = append(failedIPs, destinationIp)
		}
	}
	if successCounter == len(IPs) {
		putFailFlag = false
	}

	//step4. check if put successes, print the reuslt to shell.
	// The user can resend the command manully.
	if putFailFlag == true {
		url = "http://" + newIP + "/clientBad?id=" + ID
		networking.HTTPsend(url)
		Write2Shell("Delete" + remotefilename + " id: " + ID + " Fail")
		Write2Shell("Failed destination IPs:")
		for _, v := range failedIPs {
			Write2Shell(v)
		}
	} else {
		url = "http://" + newIP + "/clientACK?id=" + ID
		networking.HTTPsend(url)
		Write2Shell("Delete" + remotefilename + " id: " + ID + " Success")
	}

}

// func LsFile() {

// }

// func StoreFile() {

// }
