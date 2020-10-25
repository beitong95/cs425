package master

import (
	_ "errors"
	"constant"
	. "structs"
)

/**
Finished:
1. detect inactive client
2. handle client connection
3. maintain client membershiplist
4. handle datanode connection
5. detect datanode fail
6. maintain datanode membershiplist

TODO:
1. master read in client commands and put them in queue
2. master process those commands, allow parallel read or single write (because we use a queue, there is no starving problem)
3. given a file name locate the 4 copies' ips
4. handle datanode fail, create an rereplica algorithm
5. maintain vm2file and file2vm

**/


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

func find(filename string, ip string) bool {
	MF.Lock()
	var ips = File2VmMap[filename]
	for i := 0; i < len(ips); i++ {
		if ips[i] == ip {
			return true
		}
	}
	MF.Unlock()
	return false
}
// Notes: we need rereplica 
// 1) after new master elected, files have no more 4 replicas
// 2) after one datanode failed, all files stored in this datanode
func Rereplica(filename string) {
	MV.Lock()
	var replica = ""
	for ip := range Vm2fileMap {
		var found = find(filename, ip)
		if replica == "" && !found {
			replica = ip
		} else if len(Vm2fileMap[ip]) < len(Vm2fileMap[replica]) && !found {
			replica = ip
		}
	}
	MV.Unlock()
	// put file to replica
	MV.Lock()
	Vm2fileMap[replica] = append(Vm2fileMap[replica], filename)
	MV.Unlock()
	MF.Lock()
	File2VmMap[filename] = append(File2VmMap[filename], replica)
	MF.Unlock()
}

func Recover(ip string, list []string) {
	MV.Lock()
	Vm2fileMap[ip] = list
	MV.Unlock()
	for i := 0; i < len(list); i++ {
		MF.Lock()
		File2VmMap[list[i]] = append(File2VmMap[list[i]], ip)
		MF.Unlock()
	}
	return
}

func Run() {
	ServerRun(constant.MasterHTTPServerPort)
}
