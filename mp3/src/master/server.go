package master

import (
	. "structs"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"networking"
	"sync"
	"time"
	"strconv"
	"helper"
	"strings"
	"os"
	"path/filepath"
	"github.com/cespare/xxhash"
	_ "errors"
	"io"
	"client"
)

// track status
var ClientMap map[string]string = make(map[string]string)
var CM sync.Mutex
var MapleMap map[string]int = make(map[string]int)
var MapleM sync.Mutex
var JuiceMap map[string]int = make(map[string]int)
var JuiceM sync.Mutex
// Start HTTP Server
func ServerRun(port string) {
	networking.HTTPlisten("/getips", HandleGetIPs)
	networking.HTTPlisten("/put", HandlePut)
	networking.HTTPlisten("/get", HandleGet) //client will send /put?id=1
	networking.HTTPlisten("/delete", HandleDelete)
	networking.HTTPlisten("/ls", HandleLs)
	networking.HTTPlisten("/clientACK", HandleClientACK) // client will send /clientACK?id=1
	networking.HTTPlisten("/clientBad", HandleClientBad) ///client will send /clientBad?id=1
	networking.HTTPlisten("/maple", HandleMaple)  // client 2 master maple
	networking.HTTPlisten("/juice", HandleJuice)  // client 2 master juice 
	networking.HTTPstart(port)

}

//TODELETE: can be deleted. Because we handle getips request in handle get/put/delete 
func HandleGetIPs(w http.ResponseWriter, req *http.Request) {
	file, ok := req.URL.Query()["file"]
	if !ok {
		log.Println("Get IPs Url Param 'key' is missing")
		return
	}
	//detect if can give read ips
	for {
		MW.Lock()
		if WriteCounter == 0 {
			Write2Shell("Now Approve This Read")
			//Question: reader ++ ?
			MW.Unlock()
			break
		}
		MW.Unlock()
	}
	filename := file[0]
	Write2Shell("Master receive GET request for file: " + filename)
	var res []byte
	var err error
	if val, ok := File2VmMap[filename]; ok {
		res, err = json.Marshal(val)
		if err != nil {
			panic(err)
		}
		/**
		for _, v := range val {
			Write2Shell("Master sends IPS: " + v)
		}
		**/
	} else {
		res = []byte("[]")
		Write2Shell("File does not exist")
	}
	w.Write(res)
}

/*
Function name: HandleGet
Description: master server handles client Get request
Input: writer, request
OutPut: nil
Related: GetFile
*/
func HandleGet(w http.ResponseWriter, req *http.Request) {
	// record current time for exit3
	start := time.Now()

	//handle get
	//step1. get "GET" request id.
	//step2. start tracking this request's state
	ids, ok := req.URL.Query()["id"]
	if !ok {
		Logger.Error("Handle Get Url Param 'key' is missing")
		return
	}
	id := ids[0]
	CM.Lock()
	ClientMap[id] = "Get"
	CM.Unlock()
	Write2Shell("Master receive GET request id: " + fmt.Sprintf("%v", id))

	//step3. get "GET" request file name
	file, ok := req.URL.Query()["file"]
	if !ok {
		Logger.Error("Get IPs Url Param 'key' is missing")
		return
	}
	filename := file[0]
	Write2Shell("Master receive GET request for file: " + filename)

	//step4. handle reader and writer logic
	//if we cannot read now, we stop here and wait for permission
	for {
		MR.Lock()
		MW.Lock()
		if WriteCounter == 0 {
			ReadCounter++
			MW.Unlock()
			MR.Unlock()
			break
		}
		MW.Unlock()
		MR.Unlock()
	}
	Write2Shell("Now Approve This Read id: " + fmt.Sprintf("%v", id))

	//step5. send ips back to client
	var res []byte
	var err error
	if val, ok := File2VmMap[filename]; ok {
		res, err = json.Marshal(val)
		if err != nil {
			panic(err)
		}
		// print ips
		Write2Shell("Master sends IPS: " + string(res))
	} else {
		res = []byte("[]")
		Write2Shell("File does not exist")
	}
	w.Write(res)

	//step6. master wait ACK from client
	//exit 1: receive "Done" -> get success
	//exit 2: receive "Bad"  -> get fail
	//exit 3: timer timeout	 -> timeout
	//Question local variable?
	go func() {
		Write2Shell("Now waiting ACK from id: " + fmt.Sprintf("%v", id))
		for {
			CM.Lock()
			if ClientMap[id] == "Done" {
				Write2Shell("Get success ACK from id: " + fmt.Sprintf("%v", id))
				w.Write([]byte("OK"))
				//change readcounter to 0
				MR.Lock()
				ReadCounter--
				MR.Unlock()
				CM.Unlock()
				break
			} else if ClientMap[id] == "Bad" {
				Write2Shell("Get fail ACK from id: " + fmt.Sprintf("%v", id))
				w.Write([]byte("Bad"))
				//change readcounter to 0
				MR.Lock()
				ReadCounter--
				MR.Unlock()
				CM.Unlock()
				break
			} else if elapsed := time.Now().Sub(start); elapsed > MasterGetTimeout*time.Second {
				Write2Shell("Timeout id: " + fmt.Sprintf("%v", id))
				CM.Unlock()
				break
			}

			// exit 1 compare time 15 mins
			// exit 2 Question add else if ClientMap[id] == "Node Fail" exit
			CM.Unlock()
		}
	}()
}

/*
Function name: HandlePut
Description: master server handles client Put request
Input: writer, request
OutPut: nil
Related: PutFile
*/
func HandlePut(w http.ResponseWriter, req *http.Request) {
	// record current time for exit3
	start := time.Now()

	//handle put
	//step1. get "PUT" request id.
	//step2. start tracking this request's state
	ids, ok := req.URL.Query()["id"]
	if !ok {
		Logger.Error("Handle PUT Url Param 'key' is missing")
		return
	}
	id := ids[0]
	CM.Lock()
	ClientMap[id] = "Put"
	CM.Unlock()
	Write2Shell("Master receive PUT request id: " + fmt.Sprintf("%v", id))

	//step3. get "PUT" request file name
	file, ok := req.URL.Query()["file"]
	if !ok {
		Logger.Error("PUT IPs Url Param 'key' is missing")
		return
	}
	filename := file[0]
	Write2Shell("Master receive PUT request for file: " + filename)

	// step4. handle reader and writer logic
	// if we cannot write now, we stop here for further permission
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

	// step5 send ips back to client
	var res []byte
	var err error
	val, ok := File2VmMap[filename]
	list := []string{}
	if !ok {
		//new file. hash2IP
		// TODO: cannot guarantee that we can get 4 ips
		list = Hash2Ips(filename)
		if len(list) < 4 {
			Write2Shell("Client try to put a new file, but we cannot find 4 VMs to store this file")
		}
		res, err = json.Marshal(list)
		if err != nil {
			panic(err)
		}
	} else {
		// the file exists
		if len(val) < 4 {
			Write2Shell("Client try to put a existed file, but we cannot find 4 VMs store this file.")
			res = []byte("[]")
		} else if len(val) >= 4 {
			res, err = json.Marshal(val)
			if err != nil {
				panic(err)
			}
		}
	}
	Write2Shell("Master sends IPS: " + string(res))
	w.Write(res)

	//step6. master wait ACK from client
	//exit 1: receive "Done" -> get success
	//exit 2: receive "Bad"  -> get fail
	//exit 3: timer timeout	 -> timeout
	//Question local variable?

	go func() {
		Write2Shell("Now waiting ACK from id: " + fmt.Sprintf("%v", id))
		for {
			CM.Lock()
			if ClientMap[id] == "Done" {

				MF.Lock()
				if !ok {
					File2VmMap[filename] = list
				} else {
					File2VmMap[filename] = val
				}
				MF.Unlock()
				Logger.Info(File2VmMap)

				MV.Lock()

				if !ok {
					for _, v := range list {
						Vm2fileMap[v] = append(Vm2fileMap[v], filename)
					}
				} else {
					for _, v := range val {
						Vm2fileMap[v] = append(Vm2fileMap[v], filename)
					}
				}
				MV.Unlock()
				Logger.Info(Vm2fileMap)

				Write2Shell("Put success ACK from id: " + fmt.Sprintf("%v", id))
				w.Write([]byte("OK"))
				//change readcounter to 0
				MW.Lock()
				WriteCounter--
				MW.Unlock()
				CM.Unlock()
				break
			} else if ClientMap[id] == "Bad" {
				Write2Shell("Put fail ACK from id: " + fmt.Sprintf("%v", id))
				w.Write([]byte("Bad"))
				//change readcounter to 0
				MW.Lock()
				WriteCounter--
				MW.Unlock()
				CM.Unlock()
				break
			} else if elapsed := time.Now().Sub(start); elapsed > MasterPutTimeout*time.Second {
				Write2Shell("Timeout id: " + fmt.Sprintf("%v", id))
				CM.Unlock()
				break
			}

			// exit 1 compare time 5 mins
			// exit 2 Question add else if ClientMap[id] == "Node Fail" exit
			CM.Unlock()
		}
	}()

}

/*
Function name: HandleDelete
Description: master server handles client Delete request
Input: writer, request
OutPut: nil
Related: DeleteFile
*/
func HandleDelete(w http.ResponseWriter, req *http.Request) {
	start := time.Now()

	//handle delete
	//step1. get "DELETE" request id.
	//step2. start tracking this request's state
	ids, ok := req.URL.Query()["id"]
	if !ok {
		Logger.Error("Handle DELETE Url Param 'key' is missing")
		return
	}
	id := ids[0]
	CM.Lock()
	ClientMap[id] = "Delete"
	CM.Unlock()
	Write2Shell("Master receive DELETE request id: " + fmt.Sprintf("%v", id))

	//step3. get "DELETE" request file name
	file, ok := req.URL.Query()["file"]
	if !ok {
		Logger.Error("DELETE IPs Url Param 'key' is missing")
		return
	}
	filename := file[0]
	Write2Shell("Master receive DELETE request for file: " + filename)

	// step4. handle reader and writer logic
	// if we cannot write now, we stop here for further permission
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

	// step5 send ips back to client
	var res []byte
	val, ok := File2VmMap[filename]
	Write2Shell("File VMs: " + fmt.Sprint(val))
	var err error
	if !ok {
		//no such file
		res = []byte("[]")
	} else {
		// the file exists
		
		//if len(val) < 4 {
		//	res = []byte("[]")
		//} else {

		res, err = json.Marshal(val)
		if err != nil {
			panic(err)
		}
		//}
	}
	Write2Shell("Master sends IPS: " + string(res))
	w.Write(res)

	//step6. master wait ACK from client

	go func() {
		Write2Shell("Now waiting ACK from id: " + fmt.Sprintf("%v", id))
		for {
			CM.Lock()
			if ClientMap[id] == "Done" {
				MF.Lock()
				if !ok {
					//no such file
				} else {
					delete(File2VmMap, filename)
				}
				MF.Unlock()
				Logger.Info(File2VmMap)

				if !ok {
					//no such file
				} else {
					MV.Lock()
					for _, v := range val { //each vm should delete this file
						for i := 0; i < len(Vm2fileMap[v]); i++ {
							if Vm2fileMap[v][i] == filename {
								Vm2fileMap[v] = append(Vm2fileMap[v][:i], Vm2fileMap[v][i+1:]...)
							}
						}
					}
					MV.Unlock()
				}
				Logger.Info(Vm2fileMap)

				Write2Shell("DELETE success ACK from id: " + fmt.Sprintf("%v", id))
				w.Write([]byte("OK"))
				//change readcounter to 0
				MR.Lock()
				WriteCounter--
				MR.Unlock()
				CM.Unlock()
				break
			} else if ClientMap[id] == "Bad" {
				Write2Shell("DELETE fail ACK from id: " + fmt.Sprintf("%v", id))
				w.Write([]byte("Bad"))
				//change readcounter to 0
				MR.Lock()
				WriteCounter--
				MR.Unlock()
				CM.Unlock()
				break
			} else if elapsed := time.Now().Sub(start); elapsed > MasterPutTimeout*time.Second {
				Write2Shell("Timeout id: " + fmt.Sprintf("%v", id))
				CM.Unlock()
				break
			}

			// exit 1 compare time 5 mins
			// exit 2 Question add else if ClientMap[id] == "Node Fail" exit
			CM.Unlock()
		}
	}()

}

/*
Function name: HandleLs
Description: master server handles client Ls request
Input: writer, request
OutPut: nil
Related: Ls
*/
func HandleLs(w http.ResponseWriter, req *http.Request) {
	file, ok := req.URL.Query()["file"]
	if !ok {
		Logger.Error("Get IPs Url Param 'key' is missing")
		return
	}
	filename := file[0]
	Write2Shell("Master receive LS request for file: " + filename)
	for {
		MR.Lock()
		MW.Lock()
		if WriteCounter == 0 {
			ReadCounter++
			MW.Unlock()
			MR.Unlock()
			break
		}
		MW.Unlock()
		MR.Unlock()
	}
	MF.Lock()
	val := File2VmMap[filename]
	MF.Unlock()
	res, err := json.Marshal(val)
	if err != nil {
		Logger.Error("LS json Marshal error")
	}
	MR.Lock()
	ReadCounter--
	MR.Unlock()
	w.Write(res)

}

/*
Function name: HandleClientACK
Description: Handle client(last node in the put/get/delete loop) ACK. Change status of a put/get/delete request
Input: writer, request
OutPut: nil
*/
func HandleClientACK(w http.ResponseWriter, req *http.Request) {
	ids, ok := req.URL.Query()["id"]
	if !ok {
		log.Println("Client Ack Url Param 'key' is missing")
		return
	}
	id := ids[0]
	CM.Lock()
	ClientMap[id] = "Done"
	CM.Unlock()
	w.Write([]byte("OK"))
}

/*
Function name: HandleClientBad
Description: Handle client(last node in the put/get/delete loop) Bad ACK. Change status of a put/get/delete request
Input: writer, request
OutPut: nil
*/
func HandleClientBad(w http.ResponseWriter, req *http.Request) {
	ids, ok := req.URL.Query()["id"]
	if !ok {
		log.Println("Client Bad Url Param 'key' is missing")
		return
	}
	id := ids[0]
	CM.Lock()
	ClientMap[id] = "Bad"
	CM.Unlock()
	w.Write([]byte("OK"))
}

// master send task to maple workers
func SendCmdToMapler(prefix string, filename string, exe string, destinationIp string, recoverFilename string) {
	status := networking.UploadFileToWorkers(filename, exe + "_" + filename, destinationIp)

	if status == "OK" {

		// remove the record
		MF.Lock()
		// gurantee there is only one copy 
		vm := File2VmMap[recoverFilename][0]
		delete(File2VmMap, recoverFilename)
		MF.Unlock()
		MV.Lock()
		// find index
		index := 0
		for i, f := range Vm2fileMap[vm]{
			if f == recoverFilename{
				index = i
				break
			}
		}
		Vm2fileMap[vm][index] = Vm2fileMap[vm][len(Vm2fileMap[vm])-1] 
		Vm2fileMap[vm][len(Vm2fileMap[vm])-1] = ""   
		Vm2fileMap[vm] = Vm2fileMap[vm][:len(Vm2fileMap[vm])-1]  
		MV.Unlock()
		//Write2Shell("Receive Worker OK")

		// subtract 1 
		MapleM.Lock()
		MapleMap[prefix]--
		//Write2Shell("remain worker: " + fmt.Sprint(MapleMap[prefix]))
		MapleM.Unlock()
	} else {
		//do nothing wait for fail detector
	}
}
/*
Function name: HandleMaple
Description: master server handles client Maple request
Input: writer, request
OutPut: nil
Related: 
*/
func HandleMaple(w http.ResponseWriter, req *http.Request) {
	//step1. get all parameters
	exes, ok := req.URL.Query()["exe"]
	if !ok {
		Logger.Error("Handle Maple Url Param 'exe' is missing")
		return
	}
	exe := exes[0]
	Write2Shell("maple exe: " + exe)

	nums, ok := req.URL.Query()["num"]
	if !ok {
		Logger.Error("Handle Maple Url Param 'num' is missing")
		return
	}
	num, _:= strconv.Atoi(nums[0])
	Write2Shell("num: " + fmt.Sprintf("%v",num))

	prefixs, ok := req.URL.Query()["prefix"]
	if !ok {
		Logger.Error("Handle Maple Url Param 'prefix' is missing")
		return
	}
	prefix := prefixs[0]
	Write2Shell("prefix: " + prefix)

	files, ok := req.URL.Query()["file"]
	if !ok {
		Logger.Error("Handle Maple Url Param 'file' is missing")
		return
	}
	file := files[0]
	Write2Shell("Source file: " + file)

	Write2Shell("Maple start")


	//step2. partition source files
	if num <= 0 {
		res := "maple num smaller than zero"
		w.Write([]byte(res))
		return
	}

	if file == "" {
		res := "source file is empty"
		w.Write([]byte(res))
		return
	}
	// partitions []string{prefix_maplerId}
	partitions, _ := helper.HashPartition(file,uint64(num), prefix) // partition the large file into num small files
	// record total maplers
	MapleM.Lock()
	MapleMap[prefix] = num
	MapleM.Unlock()
	Write2Shell("Finish Partition")

	// step3. send task to workers
	remainPartitions := len(partitions) // maplers count
	exitFlag := false
	for exitFlag != true {
		copyMembershipList := []string{}
		// get all available vm ip
		MT.Lock()
			for id, _ := range MembershipList {
				copyMembershipList = append(copyMembershipList, id)
			}
		MT.Unlock()
		for _, id := range copyMembershipList {
			ip := strings.Split(id,"*")[0]
			destinationIp := IP2DatanodeUploadIP(ip)
			filename := partitions[remainPartitions-1]
			recoverFilename := "maple:" + prefix + ":" + filename + ":" + exe // we can use the prefix and filename to recover the SendCmdToMapler 
			MF.Lock()
			File2VmMap[recoverFilename] = []string{ip}
			MF.Unlock()
			MV.Lock()
			Vm2fileMap[ip] = append(Vm2fileMap[ip], recoverFilename)
			MV.Unlock()
			// send exe with filename
			// filename: PartitionRes_prefix_maplerid
			go SendCmdToMapler(prefix, filename, exe, destinationIp, recoverFilename)
			remainPartitions--
			if remainPartitions == 0 {
				exitFlag = true
				break
			}
		}
	}

	// set barrier wait for those workers
	// Worker ACK
	//exit 1: receive "Done" -> get success
	//exit 2: receive "Bad"  -> get fail
	//exit 3: timer timeout	 -> timeout
	MapleFalseFlag := true	
	start := time.Now()
	Write2Shell("Send all tasks to Workers. Now waiting Workers ACKs")

	for {
		time.Sleep(1 * time.Second)
		MapleM.Lock()
		if MapleMap[prefix] == 0 {
			MapleFalseFlag = false
			Write2Shell("Maple Success")
			break
		}
		MapleM.Unlock()
		if elapsed := time.Now().Sub(start); elapsed > MapleTimeout*time.Second {
			Write2Shell("Timeout " + prefix)
			break
		}
	}
	MapleM.Unlock()

	// at that time all maple results have been uploaded to HDFS.
	// safe to delete partitions
	// delete all files
	Write2Shell("Removing files with prefix: " + prefix)
	files, err := filepath.Glob("PartitionRes_" + prefix + "*")
	if err != nil {
		Logger.Fatal(err)
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			Logger.Fatal(err)
		}
	}

	if MapleFalseFlag == true {
		res := "Bad"
		w.Write([]byte(res))
	} else {
		res := "OK"
		w.Write([]byte(res))
	}
	return
}

func SendCmdToJuicer(prefix string, commandString string, destinationIp string, id string, recoverFilename string) (string, error) {
	// in mapper filename contains id
	url := "http://" + destinationIp + "/juiceWorker?keys=" + commandString + "&prefix=" + prefix + "&id=" + id
	prefix = strings.Split(prefix, "_")[1]
	Write2Shell("Send command to juicer worker " + url)
	rsp, err:= http.Get(url)
	// if there is an error, the node fail
	if err != nil {
		return "Bad", nil	
	}
	// else
	// receive file
	localfilename := "juiceResult2Master_" + prefix + "_" + id 
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

	MF.Lock()
	// gurantee there is only one copy 
	// if fail we exit we keep those record
	vm := File2VmMap[recoverFilename][0]
	delete(File2VmMap, recoverFilename)
	MF.Unlock()
	MV.Lock()
	// find index
	index := 0
	for i, f := range Vm2fileMap[vm]{
		if f == recoverFilename{
			index = i
			break
		}
	}
	Vm2fileMap[vm][index] = Vm2fileMap[vm][len(Vm2fileMap[vm])-1] 
	Vm2fileMap[vm][len(Vm2fileMap[vm])-1] = ""   
	Vm2fileMap[vm] = Vm2fileMap[vm][:len(Vm2fileMap[vm])-1]  
	MV.Unlock()
	JuiceM.Lock()
	JuiceMap[prefix]--
	JuiceM.Unlock()
	return "OK", nil
}


/*
Function name: HandleJuice
Description: master server handles client Juice request
Input: writer, request
OutPut: nil
Related: 
*/
func HandleJuice(w http.ResponseWriter, req *http.Request) {
	// handle juice
	// step1. get all parameters
	exes, ok := req.URL.Query()["exe"]
	if !ok {
		Logger.Error("Handle Juice Url Param 'exe' is missing")
		return
	}
	exe := exes[0]
	Write2Shell("juice exe: " + exe)

	nums, ok := req.URL.Query()["num"]
	if !ok {
		Logger.Error("Handle Juice Url Param 'num' is missing")
		return
	}
	num, _:= strconv.Atoi(nums[0])
	Write2Shell("num: " + fmt.Sprintf("%v",num))

	prefixs, ok := req.URL.Query()["prefix"]
	if !ok {
		Logger.Error("Handle Juice Url Param 'prefix' is missing")
		return
	}
	prefix := prefixs[0]
	// datanode can decode exe 
	toSendPrefix := exe + "_" + prefix
	Write2Shell("prefix: " + prefix)

	files, ok := req.URL.Query()["file"]
	if !ok {
		Logger.Error("Handle Juice Url Param 'file' is missing")
		return
	}
	destFile := files[0]
	Write2Shell("dest file: " + destFile)

	deletes, ok := req.URL.Query()["delete"]
	if !ok {
		Logger.Error("Handle Juice Url Param 'file' is missing")
		return
	}
	delete := deletes[0]
	Write2Shell("is delete: " + delete)
	// convert it to int
	isDelete,_ := strconv.Atoi(delete)
	Write2Shell(fmt.Sprint(isDelete))
	Write2Shell("Juice start")

	//step2. find all keys and all available files (use maplerid as index)
	KeyList := make(map[string][]string) // [key][mapleworkerids]
	MV.Lock()
	for fname,_ := range File2VmMap {
		if isFileOfJuice:=strings.Contains(fname, prefix); isFileOfJuice == true {
			// extract the key
			key := strings.Split(fname, "_")[3]
			workerId := strings.Split(fname, "_")[2]
			if _, ok := KeyList[key]; ok == true {
				KeyList[key] = append(KeyList[key],workerId)
				continue
			} else {
				KeyList[key] = []string{workerId}
			}
		}
	}
	MV.Unlock()

	//step3. shuffle TODO: add shuffle option
	ShuffleRes := make(map[int][]string) // [juicer id][keys]
	for key, _ := range KeyList {
		res := HashShuffle(key, uint64(num))
		ShuffleRes[res] = append(ShuffleRes[res], key)
	}
	// we can use ShuffleRes and KeyList to send commands [juicer id][keys] [key][related file(maplerid)]
	
	// step4. send juice command to workers
	remainJuiceWorkers := len(ShuffleRes)
	JuiceM.Lock()
	JuiceMap[prefix] = num
	JuiceM.Unlock()
	realJuicerWorkers := []string{} // ignore those unlucky guys which have no key
	
	exitFlag := false
	for exitFlag != true {
		copyMembershipList := []string{}
		// get all available vm ip
		MT.Lock()
			for id, _ := range MembershipList {
				copyMembershipList = append(copyMembershipList, id)
			}
		MT.Unlock()
		for _, id := range copyMembershipList {
			ip := strings.Split(id,"*")[0]
			destinationIp := IP2DatanodeUploadIP(ip)
			//Write2Shell(destinationIp)
			// convert keys into a long string
			commandString := ""
			// if no keys, we skip this worker
			for len(ShuffleRes[remainJuiceWorkers-1]) == 0 {
				remainJuiceWorkers--
			}
			// at least we have one
			// add all real workers who get the command, later we can delete them
			realJuicerWorkers = append(realJuicerWorkers, fmt.Sprint(remainJuiceWorkers-1))
			for _, key := range ShuffleRes[remainJuiceWorkers-1] {
				// we also need to add subid (maple worker id)
				for _, workerId := range KeyList[key] {
					// commandString is a list of to be downloaded files index (maplerid + key)
					commandString = commandString + workerId + "_" + key + ","
				}
			}
			//Write2Shell(commandString)
			// record info so that we can recover if the node fails
			recoverFilename := "juice:" + toSendPrefix + ":" + commandString + ":" + fmt.Sprint(remainJuiceWorkers-1)
			MF.Lock()
			File2VmMap[recoverFilename] = []string{ip}
			MF.Unlock()
			MV.Lock()
			Vm2fileMap[ip] = append(Vm2fileMap[ip], recoverFilename)
			MV.Unlock()

			go SendCmdToJuicer(toSendPrefix, commandString, destinationIp, fmt.Sprint(remainJuiceWorkers-1), recoverFilename)
			remainJuiceWorkers--
			if remainJuiceWorkers == 0 {
				exitFlag = true
				break
			}
		}
	}

	// step 5. wait for ack and file name 
	// set barrier wait for those workers
	// Worker ACK
	//exit 1: receive "Done" -> get success
	//exit 2: receive "Bad"  -> get fail
	//exit 3: timer timeout	 -> timeout
	juiceFalseFlag := true	
	start := time.Now()
	Write2Shell("Now waiting Workers ACKs")

	for {
		time.Sleep(1 * time.Second)
		JuiceM.Lock()
		if JuiceMap[prefix] == 0 {
			Write2Shell("Juice Partial Success, now send the final res to HDFS/client")
			juiceFalseFlag = false
			break
		}
		JuiceM.Unlock()
		if elapsed := time.Now().Sub(start); elapsed > JuiceTimeout*time.Second {
			Write2Shell("Timeout " + prefix)
			break
		}

	}
	JuiceM.Unlock()

	//step 6. merge and upload the final res 
	// juice res file name: juiceResult2Master_prefix_juicerid 
	// final res, do we need to send it to client? just like datanode send it back to master
	finalRes, err := os.OpenFile(destFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		Logger.Fatal(err)
	}
	defer finalRes.Close()
	for _, realId := range realJuicerWorkers {
		filename := "juiceResult2Master_" + prefix + "_" + realId 
		inFile, err := os.Open(filename)
		if err != nil {
			Logger.Fatal(err)
		}
		_, err = io.Copy(finalRes, inFile)
		if err != nil {
			Logger.Fatal(err)
		}
		inFile.Close()
		if err := os.Remove(filename); err != nil {
			Logger.Fatal(err)
		}
	}

	//step 7. delete files
	if isDelete == 1 {
		Write2Shell("Request Delete") // delete what? map intermediate result?
		for key, maplerid := range KeyList {
			for _, id := range maplerid {
				todeleteFilename := "mapleResult_" + prefix + "_" + id + "_" + key
				client.DeleteFile(todeleteFilename)
			}
		}
	}

	//step 8. send ack to client 

	if juiceFalseFlag == true {
		res := "Bad"
		w.Write([]byte(res))
	} else {
		res := "OK"
		w.Write([]byte(res))
	}
	return
}

func HashShuffle(key string, num uint64) int {
	hash := xxhash.Sum64([]byte(key))
	return int(hash % num)
} 

//sort and range 
func RangeShuffle(KeyList map[string][]string) map[int][]string {
	Write2Shell("TODO")
	return nil
}