package master
import(
	"networking"
	"net/http"
	"fmt"
	"log"
	"encoding/json"
) 
func ServerRun(port string){
	networking.HTTPlisten("/getips", HandleGetIPs)
	networking.HTTPstart(port)

}
func HandleGetIPs(w http.ResponseWriter, req *http.Request){
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