package master

import (
	_ "errors"
	. "structs"
	"networking"
)

/**
TODO:
1. failure how to gracefully exit?
2. performance
**/

/*
Function name: FindMaxLen
Description: find the fullest VM (actually we should use the real size of all files rather than the count of all files)
Input: ips []string
OutPut: index, maxlength
*/
func FindMaxLen(ips []string) (int, string) {
	var output = ""
	var idx = 0
	for i := 0; i < 4; i++ {
		if output == "" {
			output = ips[i]
		} else if len(Vm2fileMap[ips[i]]) > len(Vm2fileMap[output]) {
			output = ips[i]
			idx = i
		}
	}
	return idx, output
}

/*
Function name: Hash2Ips
Description: find 4 VMs' IPs to store a new file
Input: filename
OutPut: 4 ips
*/
func Hash2Ips(filename string) []string {
	// assert filename is name of new file!
	var fourIps = []string{"", "", "", ""}
	MV.Lock()
	for ip := range Vm2fileMap {
		if fourIps[0] == "" {
			fourIps[0] = ip
		} else if fourIps[1] == "" {
			fourIps[1] = ip
		} else if fourIps[2] == "" {
			fourIps[2] = ip
		} else if fourIps[3] == "" {
			fourIps[3] = ip
		} else {
			var idx, maxlen = FindMaxLen(fourIps)
			if len(Vm2fileMap[ip]) < len(Vm2fileMap[maxlen]) {
				fourIps[idx] = ip
			}
		}
	}
	//just for test. should be implemented in put/delete.
	for i := 0; i < 4; i++ {
		if fourIps[i] == "" {
			fourIps = fourIps[:i]
			break
		}
		//Vm2fileMap[fourIps[i]] = append(Vm2fileMap[fourIps[i]], filename)
	}
	MV.Unlock()
	/**
	MF.Lock()
	File2VmMap[filename] = fourIps
	MF.Unlock()
	**/
	return fourIps
}

/*
Function name: find
Description: find if a specific vm has a specific file
Input: filename, ip of the vm
OutPut: true or false
*/
func find(filename string, ip string) bool {
	MF.Lock()
	var ips = File2VmMap[filename]
	for i := 0; i < len(ips); i++ {
		if ips[i] == ip {
			MF.Unlock()
			return true
		}
	}
	MF.Unlock()
	return false
}


/*
Function name: Rereplica
Description: rereplica a file in HDFS
Input: filename 
OutPut: nil
Related: UDPServer faildetector
*/
// 1) after new master elected, files have no more 4 replicas
// 2) after one datanode failed, all files stored in this datanode
func Rereplica(filename string) {
	//Write2Shell("Start rereplica " + filename)
	var replicas = []string{}
	var sources = []string{}
	MV.Lock()
	// find available VMS (1) sources: who have the file (2) replicas: where we can store the file 
	for ip := range Vm2fileMap{
		var found = find(filename, ip)
		if !found {
			replicas = append(replicas, ip)
		} else if found {
			sources = append(sources, ip)
		}
	}
	
	MV.Unlock()
	// put file to replica
	// send rereplica request
	// Ask one vm send a file to another vm
	rereplicaFailFlag := true
	// two for loops, find all combinations of sources and replica destinations
	for _,source := range sources {
		if rereplicaFailFlag == false {
			break
		}
		for _,replica := range replicas {
			if rereplicaFailFlag == false {
				break
			}
			for {
				MW.Lock()
				MR.Lock()
				if ReadCounter == 0 && WriteCounter == 0 {
					WriteCounter++
					MR.Unlock()
					MW.Unlock()
					break
				}
				MR.Unlock()
				MW.Unlock()
			}

			sourceIp := IP2DatanodeUploadIP(source)
			url := "http://" + sourceIp + "/rereplica?file=" + filename + "&destination=" + replica
			MV.Lock()
			Vm2fileMap[replica] = append(Vm2fileMap[replica], filename)
			MV.Unlock()
			MF.Lock()
			File2VmMap[filename] = append(File2VmMap[filename], replica)
			MF.Unlock()
			body := networking.HTTPsend(url)
			if string(body) == "OK" {
				rereplicaFailFlag = false
				//Write2Shell("Rereplica file " +  filename + " from " + source + " to " + replica + " Success!")
				// master update metadata
			} else if string(body) == "Bad" {
				//Write2Shell("Rereplica file " +  filename + " from " + source + " to " + replica + " Fail!")
			}
			MW.Lock()
			WriteCounter--
			MW.Unlock()
		}
	}
	if rereplicaFailFlag == true {
		//Write2Shell("Rereplica file " + filename + " Fail! We cannot reach the required rereplica factor = 3")
	}
	Logger.Info(Vm2fileMap)
}

/*
Function name: Recover
Description: recover the metadata file2vm and vm2file after election
Input: a datanode ip, local files stored on that datanode
OutPut: nil
Related: UDPServer faildetector
*/
func Recover(ip string, list []string) {
	MV.Lock()
	MF.Lock()
	Vm2fileMap[ip] = list
	for i := 0; i < len(list); i++ {
		File2VmMap[list[i]] = append(File2VmMap[list[i]], ip)
	}
	MF.Unlock()
	MV.Unlock()
	return
}

func Run() {
	ServerRun(MasterHTTPServerPort)
}
