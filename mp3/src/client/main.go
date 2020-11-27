package client

import (
	"encoding/json"
	"fmt"
	"networking"
	. "structs"
	"time"
	"helper"
	"strings"
)

/**
 TODO:
**/

/*
Function name: GetIPsFromMaster
Description: send get ips request to master and parse the returned body
Input: filename string
OutPut: ips []string
*/
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

/*
Function name: GetFile
Description: client get file from HDFS
Input: filename string, localfilename string
OutPut: nil
Related: HandleGet, DownloadFileFromDatanode
*/
func GetFile(filename string, localfilename string) {
	//t1 := time.Now()
	getFailFlag := true
	//step 1. get id and create url
	ID := fmt.Sprint(time.Now().UnixNano())
	newIP := IP2MasterHTTPServerIP(MasterIP)
	url := "http://" + newIP + "/get?id=" + ID + "&file=" + filename + "&memberID=" + MyID
	//Write2Shell("GetFile url: " + url)

	//step 2. send url and decode body
	body := networking.HTTPsend(url)
	var IPs []string
	IPs = []string{}
	err := json.Unmarshal(body, &IPs)
	if err != nil {
		Write2Shell("Unmarshal error in GetFile")
	}

	//step3. download files from the list
	for _, ip := range IPs {
		//ip: ip + udpPort  -> newIp: ip + datanodeHTTPServerPort
		newIp := IP2DatanodeHTTPServerIP(ip)
		status, _ := networking.DownloadFileFromDatanode(filename, localfilename, newIp)
		if status == "OK" {
			//t2 := time.Now()
			//diff := t2.Sub(t1)
			//Write2Shell("Get: " + filename + " time: " + diff.String())
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
		//Write2Shell("Get " + filename + " " + localfilename + " id: " + ID + " Fail")
	} else {
		//Write2Shell("Get " + filename + " " + localfilename + " id: " + ID + " Success")
	}
}

/*
Function name: PutFile
Description: client put file from HDFS
Input: filename string, remotefilename string
OutPut: nil
Related: HandlePut, UploadFileToDatanode
*/
func PutFile(filename string, remotefilename string) {
	//t1 := time.Now()
	putFailFlag := true
	//step 1. get id and create url
	ID := fmt.Sprint(time.Now().UnixNano())
	newIP := IP2MasterHTTPServerIP(MasterIP)
	url := "http://" + newIP + "/put?id=" + ID + "&file=" + remotefilename + "&memberID=" + MyID
	//Write2Shell("Putfile url: " + url)

	//step 2. send url and decode body
	body := networking.HTTPsend(url)
	var IPs []string
	IPs = []string{}
	err := json.Unmarshal(body, &IPs)
	if err != nil {
		Write2Shell("Unmarshal error in PutFile")
	}

	if len(IPs) == 0 {
		url = "http://" + newIP + "/clientBad?id=" + ID
		networking.HTTPsend(url)
		//Write2Shell("Put" + filename + " " + remotefilename + " id: " + ID + " Fail")
		//Write2Shell("Reason: IPs lenght == 0")
		return
	}

	//step 3. upload files to vms in the list
	successCounter := 0

	for _, ip := range IPs {
		//ip: ip + udpPort  -> newIp: ip + datanodeHTTPServerPort
		destinationIp := IP2DatanodeUploadIP(ip)
		status := networking.UploadFileToDatanode(filename, remotefilename, destinationIp)
		if status == "OK" {
			successCounter++
		} else {
			//rereplica will handle it
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
		//Write2Shell("Put" + filename + " " + remotefilename + " id: " + ID + " Fail")
	} else {
		//t2 := time.Now()
		//diff := t2.Sub(t1)
		//Write2Shell("Put: " + filename + " time: " + diff.String())
		url = "http://" + newIP + "/clientACK?id=" + ID
		networking.HTTPsend(url)
		//Write2Shell("Put" + filename + " " + remotefilename + " id: " + ID + " Success")
	}
}

/*
Function name: DeleteFile
Description: client delete file from HDFS(similiar to put file)
Input: remotefilename string
OutPut: nil
Related: HandleDelete
*/
func DeleteFile(remotefilename string) {
	putFailFlag := true
	//step 1. get id and create url
	ID := fmt.Sprint(time.Now().UnixNano())
	newIP := IP2MasterHTTPServerIP(MasterIP)
	url := "http://" + newIP + "/delete?id=" + ID + "&file=" + remotefilename + "&memberID=" + MyID
	//Write2Shell("Deletefile url: " + url)

	//step 2. send url and decode body
	body := networking.HTTPsend(url)
	//Write2Shell(string(body))
	var IPs []string
	IPs = []string{}
	err := json.Unmarshal(body, &IPs)
	if err != nil {
		Write2Shell("Unmarshal error in DeleteFile")
	}

	// no such file or not enough vms (4)
	if len(IPs) == 0 {
		url = "http://" + newIP + "/clientBad?id=" + ID
		networking.HTTPsend(url)
		Write2Shell("Delete" + remotefilename + " id: " + ID + " Fail")
		Write2Shell("Reason: IPs lenght == 0")
		return
	}
	// should always return 4 ips`

	//step 3. delete files to vms in the list
	successCounter := 0
	failedIPs := []string{}
	for _, ip := range IPs {
		//ip: ip + udpPort  -> newIp: ip + datanodeHTTPServerPort
		destinationIp := IP2DatanodeUploadIP(ip)
		status := networking.DeleteFileFromDatanode(remotefilename, destinationIp)
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
		//Write2Shell("Delete" + remotefilename + " id: " + ID + " Fail")
		//Write2Shell("Failed destination IPs:")
		for _, v := range failedIPs {
			//Write2Shell(v)
		}
	} else {
		url = "http://" + newIP + "/clientACK?id=" + ID
		networking.HTTPsend(url)
		//Write2Shell("Delete" + remotefilename + " id: " + ID + " Success")
	}

}

/*
Function name: Ls
Description: client query locations of a file in HDFS
Input: remotefilename string
OutPut: nil
Related: HandleLs
*/
func Ls(remotefilename string) {
	ID := fmt.Sprint(time.Now().UnixNano())
	newIP := IP2MasterHTTPServerIP(MasterIP)
	url := "http://" + newIP + "/ls?id=" + ID + "&file=" + remotefilename
	//Write2Shell("Ls file url: " + url)
	body := networking.HTTPsend(url)
	//Write2Shell(string(body))
	var IPs = []string{}
	err := json.Unmarshal(body, &IPs)
	if err != nil {
		Write2Shell("Unmarshal error in DeleteFile")
	}
	if len(IPs) == 0 {
		Write2Shell("no VMs hold such file")
	} else {
		Write2Shell(remotefilename + "existed in:")
		for _, ip := range IPs {
			Write2Shell(ip)
		}
	}
}

/*
Function name: Store
Description: client list all local files
Input:
OutPut: nil
*/
func Store() {
	list := helper.List()
	Write2Shell("This VM contain files:")
	for _, val := range list {
		Write2Shell(val)
	}
}

/*
Function name: Maple
Description: client execute the maple step
Input: 
OutPut: 
Related: 
*/
func Maple(maple_exe string, num_maples string, sdfs_intermediate_filename_prefix string, input_file string, _cmd string) {
	// we assume the maple_exe has already been stored in every datanode's main folder
	// we assume the input files have already been stored in master's main folder
	start := time.Now()
	newIP := IP2MasterHTTPServerIP(MasterIP)
	url := "http://" + newIP + "/maple?exe=" + maple_exe + "&num=" + num_maples + "&prefix=" + sdfs_intermediate_filename_prefix + "&file=" + input_file 
	Write2Shell("maple url: " + url)
	body := networking.HTTPsend(url)
	//Write2Shell("maple body: " + string(body))
	_cmd = strings.Replace(_cmd,"\n", "", 1)
	if string(body) == "OK" {
		Write2Shell(_cmd + " success")
	} else{
		Write2Shell(_cmd + " fail")
	}
	delta := time.Now().Sub(start).String()
	Write2Shell("Maple time: " +  delta)

}

/*
Function name: Juice
Description: client execute the juice step
Input: 
OutPut: 
Related: 
*/
func Juice(juice_exe string, num_juices string, sdfs_intermediate_filename_prefix string, destfile string, delete_input string, _cmd string) {
	// we assume the juice_exe has already been stored in every datanode's main folder
	// we assume the juice source file mapleResult_prefix_mapleworkerid_key has already been stored in hdfs
	start := time.Now()
	newIP := IP2MasterHTTPServerIP(MasterIP)
	url := "http://" + newIP + "/juice?exe=" + juice_exe + "&num=" + num_juices + "&prefix=" + sdfs_intermediate_filename_prefix + "&file=" + destfile + "&delete=" + delete_input
	Write2Shell("juice url: " + url)
	body := networking.HTTPsend(url)
	//Write2Shell("juice body: " + string(body))
	_cmd = strings.Replace(_cmd,"\n", "", 1)
	if string(body) == "OK" {
		Write2Shell(_cmd + " success")
	} else{
		Write2Shell(_cmd + " fail")
	}
	delta := time.Now().Sub(start).String()
	Write2Shell("Juice time: " +  delta)
}
