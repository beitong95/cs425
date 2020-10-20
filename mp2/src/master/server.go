package master
import(
	"networking"
	"net/http"
	"fmt"
	"log"
	"encoding/json"
	"sync"
) 
var ClientMap map[string] string = make(map[string] string)
var CM sync.Mutex

func ServerRun(port string){
	networking.HTTPlisten("/getips", HandleGetIPs)
	networking.HTTPlisten("/put", HandlePut)
	networking.HTTPlisten("/get", HandleGet)//client will send /put?id=1
	networking.HTTPlisten("/clientACK", HandleClientACK)// client will send /clientACK?id=1
	networking.HTTPlisten("/clientBad", HandleClientBad)//client will send /clientBad?id=1
	networking.HTTPstart(port)

}
func HandleGetIPs(w http.ResponseWriter, req *http.Request){
	fmt.Println(req)
	file, ok := req.URL.Query()["file"]
    if !ok {
        log.Println("Get IPs Url Param 'key' is missing")
        return
    }
	filename := file[0]
	var res []byte
	var err error
	if val, ok := File2VmMap[filename]; ok {
		res,err = json.Marshal(val)
		if err != nil{
			panic(err)
		}
	}else{
		res = []byte("[]")
	}
	w.Write(res)
	fmt.Println(filename)
}

func HandlePut(w http.ResponseWriter, req *http.Request){
	file, ok := req.URL.Query()["file"]
    if !ok {
        log.Println("Url Param 'key' is missing")
        return
    }
	filename := file[0]
	var res []byte
	var err error
	if val, ok := File2VmMap[filename]; ok {
		res,err = json.Marshal(val)
		if err != nil{
			panic(err)
		}
	}else{
		res = []byte("[]")
	}
	w.Write(res)
	fmt.Println(filename)
}

func HandleGet(w http.ResponseWriter, req *http.Request){
	ids, ok := req.URL.Query()["id"]
    if !ok {
        log.Println("Handle Get Url Param 'key' is missing")
        return
    }
	id := ids[0]
	CM.Lock()
	ClientMap[id] = "Get"
	CM.Unlock()
	for{
		CM.Lock()
		if ClientMap[id] == "Done"{
			w.Write([]byte("OK"))
			CM.Unlock()
			break;
		}else if ClientMap[id] == "Bad"{
			w.Write([]byte("Bad"))
			CM.Unlock()
			break;
		}
		CM.Unlock()
	}
}

func HandleClientACK(w http.ResponseWriter, req *http.Request){
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

func HandleClientBad(w http.ResponseWriter, req *http.Request){
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
