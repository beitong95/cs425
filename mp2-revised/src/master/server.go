package master

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"networking"
	"sync"
	. "structs"
	"time"
	"constant"
)

var ClientMap map[string]string = make(map[string]string)
var CM sync.Mutex

func ServerRun(port string) {
	networking.HTTPlisten("/getips", HandleGetIPs)
	networking.HTTPlisten("/put", HandlePut)
	networking.HTTPlisten("/get", HandleGet)             //client will send /put?id=1
	networking.HTTPlisten("/clientACK", HandleClientACK) // client will send /clientACK?id=1
	networking.HTTPlisten("/clientBad", HandleClientBad) //client will send /clientBad?id=1
	networking.HTTPstart(port)

}

// func HandleGetIPsPut(w http.ResponseWriter, req *http.Request) {
// 	file, ok := req.URL.Query()["file"]
// 	if !ok {
// 		log.Println("Get IPs Url Param 'key' is missing")
// 		return
// 	}
// 	//detect if can give write ips
// 	for {
// 		MW.Lock()
// 		MR.Lock()
// 		if ReadCounter == 0 && WriteCounter == 0 {
// 			MW.Unlock()
// 			MR.Unlock()
// 			break
// 		}
// 		MW.Unlock()
// 		MR.Unlock()
// 	}
// 	filename := file[0]
// 	var res []byte
// 	var err error
// 	if val, ok := File2VmMap[filename]; ok {
// 		res, err = json.Marshal(val)
// 		if err != nil {
// 			panic(err)
// 		}
// 	} else {
// 		val := Hash2Ips()
// 		res, err = json.Marshal(val)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// 	w.Write(res)
// 	fmt.Println(filename)

// }

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
		for _,v := range val {
			Write2Shell("Master sends IPS: " + v)
		}
	} else {
		res = []byte("[]")
		Write2Shell("File does not exist")
	}
	w.Write(res)
}

func HandlePut(w http.ResponseWriter, req *http.Request) {
	file, ok := req.URL.Query()["file"]
	if !ok {
		log.Println("Url Param 'key' is missing")
		return
	}
	filename := file[0]
	var res []byte
	var err error
	if val, ok := File2VmMap[filename]; ok {
		res, err = json.Marshal(val)
		if err != nil {
			panic(err)
		}
	} else {
		res = []byte("[]")
	}
	w.Write(res)
	fmt.Println(filename)
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
	Write2Shell("Master receive GET request id: " + fmt.Sprintf("%v",id))

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
	MR.Lock()
	ReadCounter++
	MR.Unlock()
	for {
		MW.Lock()
		if WriteCounter == 0 {
			MW.Unlock()
			break
		}
		MW.Unlock()
	}
	Write2Shell("Now Approve This Read id: " + fmt.Sprintf("%v",id))

	//step5. send ips back to client 
	var res []byte
	var err error
	if val, ok := File2VmMap[filename]; ok {
		res, err = json.Marshal(val)
		if err != nil {
			panic(err)
		}
		// print ips
		for _,v := range val {
			Write2Shell("Master sends IPS: " + v)
		}
	} else {
		res = []byte("[]")
		Write2Shell("File does not exist")
	} 
	w.Write(res)

	//step6. master wait ACK from client
	//exit 1: receive "Done" -> get success
	//exit 2: receive "Bad"  -> get fail
	//exit 3: timer timeout	 -> timeout
	Write2Shell("Now waiting ACK from id: " + fmt.Sprintf("%v",id))
	for {
		CM.Lock()
		if ClientMap[id] == "Done" {
			w.Write([]byte("OK"))
			//change readcounter to 0
			MR.Lock()
			ReadCounter--
			MR.Unlock()
			CM.Unlock()
			break
		} else if ClientMap[id] == "Bad" {
			w.Write([]byte("Bad"))
			//change readcounter to 0
			MR.Lock()
			ReadCounter--
			MR.Unlock()
			CM.Unlock()
			break
		} else if elapsed := start.Sub(time.Now()); elapsed > constant.MasterGetTimeout * time.Second {
			Write2Shell("Timeout id: " + fmt.Sprintf("%v",id))
			CM.Unlock()
			break
		} 

		// exit 1 compare time 15 mins 
		// exit 2 Question add else if ClientMap[id] == "Node Fail" exit
		CM.Unlock()
	}
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