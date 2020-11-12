package master

import (
	"constant"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"networking"
	. "structs"
	"sync"
	"time"
	"strconv"
)

// track status
var ClientMap map[string]string = make(map[string]string)
var CM sync.Mutex

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
	networking.HTTPstart(port)

}

//TODELETE: can be deleted
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
		res, err = json.Marshal(val)
		if err != nil {
			panic(err)
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
			} else if elapsed := start.Sub(time.Now()); elapsed > constant.MasterPutTimeout*time.Second {
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
		list = Hash2Ips(filename)
		res, err = json.Marshal(list)
		if err != nil {
			panic(err)
		}
	} else {
		// the file exists
		if len(val) < 4 {
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
			} else if elapsed := start.Sub(time.Now()); elapsed > constant.MasterPutTimeout*time.Second {
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
			} else if elapsed := start.Sub(time.Now()); elapsed > constant.MasterGetTimeout*time.Second {
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

func HandleMaple(w http.ResponseWriter, req *http.Request) {
	// record current time for exit3

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
	Write2Shell("prefix: " + prefix)


	files, ok := req.URL.Query()["file"]
	if !ok {
		Logger.Error("Handle Maple Url Param 'file' is missing")
		return
	}
	file := files[0]
	Write2Shell("file: " + file)

	Write2Shell("Maple start")


	//step2. get, partition and put maple source files
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
	// get the file to local filesytem



	/**

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
		list = Hash2Ips(filename)
		res, err = json.Marshal(list)
		if err != nil {
			panic(err)
		}
	} else {
		// the file exists
		if len(val) < 4 {
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
			} else if elapsed := start.Sub(time.Now()); elapsed > constant.MasterPutTimeout*time.Second {
				Write2Shell("Timeout id: " + fmt.Sprintf("%v", id))
				CM.Unlock()
				break
			}

			// exit 1 compare time 5 mins
			// exit 2 Question add else if ClientMap[id] == "Node Fail" exit
			CM.Unlock()
		}
	}()
	**/

}

func HandleJuice(w http.ResponseWriter, req *http.Request) {
	Write2Shell("TODO handle maple")
	res := "Fail"
	w.Write([]byte(res))
}