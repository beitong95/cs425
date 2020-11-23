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
)

// track status
var ClientMap map[string]string = make(map[string]string)
var CM sync.Mutex
var MapleMap map[string]int = make(map[string]int)
var MapleM sync.Mutex
// Start HTTP Server
func ServerRun(port string) {
	networking.HTTPlisten("/getips", HandleGetIPs)
	networking.HTTPlisten("/put", HandlePut)
	networking.HTTPlisten("/get", HandleGet) //client will send /put?id=1
	networking.HTTPlisten("/delete", HandleDelete)
	networking.HTTPlisten("/ls", HandleLs)
	networking.HTTPlisten("/clientACK", HandleClientACK) // client will send /clientACK?id=1
	networking.HTTPlisten("/clientBad", HandleClientBad) //client will send /clientBad?id=1
	networking.HTTPlisten("/maple", HandleMaple) //client will send /clientBad?id=1
	networking.HTTPlisten("/juice", HandleJuice) //client will send /clientBad?id=1
//	networking.HTTPlisten("/workerACK", HandleWorkerACK) // worker will send /workerACK?id=1
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
		if len(val) < 4 {
			res = []byte("[]")
		} else {
			res, err = json.Marshal(val)
			if err != nil {
				panic(err)
			}
		}
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


func UploadFileToWorkersWrapper(prefix string, filename string, destinationIp string) {
	status := networking.UploadFileToWorkers(filename, filename, destinationIp)
	// handle vm2file file2vm
	MF.Lock()
	// gurantee there is only one copy 
	vm := File2VmMap[filename][0]
	delete(File2VmMap, filename)
	MF.Unlock()
	MV.Lock()
	// find index
	index := 0
	for i, f := range Vm2fileMap[vm]{
		if f == filename {
			index = i
			break
		}
	}
	Vm2fileMap[vm][index] = Vm2fileMap[vm][len(Vm2fileMap[vm])-1] 
	Vm2fileMap[vm][len(Vm2fileMap[vm])-1] = ""   
	Vm2fileMap[vm] = Vm2fileMap[vm][:len(Vm2fileMap[vm])-1]  
	MV.Unlock()

	if status == "OK" {
		Write2Shell("Receive Worker OK")
		// subtract 1 
		MapleM.Lock()
		MapleMap[prefix]--
		Write2Shell("remain worker: " + fmt.Sprint(MapleMap[prefix]))
		MapleM.Unlock()
	} else {

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
	// record current time for exit
	//handle maple
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
	prefix = exe + "_" + prefix
	Write2Shell("prefix: " + prefix)

	files, ok := req.URL.Query()["file"]
	if !ok {
		Logger.Error("Handle Maple Url Param 'file' is missing")
		return
	}
	file := files[0]
	Write2Shell("file: " + file)

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
	partitions, _ := helper.HashPartition(file,uint64(num), prefix)
	MapleM.Lock()
	MapleMap[prefix] = num
	MapleM.Unlock()

	// step3. send task to workers
	remainPartitions := len(partitions)
	exitFlag := false
	for exitFlag != true {
		copyMembershipList := []string{}
		// get all available vm ip
		MT.Lock()
			for id, _ := range MembershipList {
				copyMembershipList = append(copyMembershipList, id)
			}
		MT.Unlock()
		fmt.Println(copyMembershipList)
		for _, id := range copyMembershipList {
			ip := strings.Split(id,"*")[0]
			destinationIp := IP2DatanodeUploadIP(ip)
			filename := partitions[remainPartitions-1]

			MF.Lock()
			File2VmMap[filename] = []string{ip}
			MF.Unlock()
			MV.Lock()
			Vm2fileMap[ip] = append(Vm2fileMap[ip], filename)
			MV.Unlock()
			// send exe with filename
			go UploadFileToWorkersWrapper(prefix, filename, destinationIp)
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
	Write2Shell("Now waiting Workers ACKs")

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
	files, err := filepath.Glob(prefix + "*")
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


/*
Function name: HandleJuice
Description: master server handles client Juice request
Input: writer, request
OutPut: nil
Related: 
*/
func HandleJuice(w http.ResponseWriter, req *http.Request) {
	Write2Shell("TODO handle maple")
	res := "Fail"
	w.Write([]byte(res))
}

/*
Function name: HandleClientACK
Description: Handle client(last node in the put/get/delete loop) ACK. Change status of a put/get/delete request
Input: writer, request
OutPut: nil
*/
// dont need that 
/**
func HandleWorkerACK(w http.ResponseWriter, req *http.Request) {
	filenames, ok := req.URL.Query()["filename"]
	if !ok {
		log.Println("Client Ack Url Param 'key' is missing")
		return
	}

	// filename
	filename := filenames[0]
	prefix := strings.Split(filename, "_")[0]
	Write2Shell("Receive Worker ACK")
	// subtract 1 
	MapleM.Lock()
	MapleMap[prefix]--
	Write2Shell("remain worker: " + fmt.Sprint(MapleMap[prefix]))
	MapleM.Unlock()
	// handle vm2file file2vm
	MF.Lock()
	// gurantee there is only one copy 
	vm := File2VmMap[filename][0]
	delete(File2VmMap, filename)
	MF.Unlock()
	MV.Lock()
	// find index
	index := 0
	for i, f := range Vm2fileMap[vm]{
		if f == filename {
			index = i
			break
		}
	}
	Vm2fileMap[vm][index] = Vm2fileMap[vm][len(Vm2fileMap[vm])-1] 
	Vm2fileMap[vm][len(Vm2fileMap[vm])-1] = ""   
	Vm2fileMap[vm] = Vm2fileMap[vm][:len(Vm2fileMap[vm])-1]  
	MV.Unlock()
	w.Write([]byte("OK"))
}
**/