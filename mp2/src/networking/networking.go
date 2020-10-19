package networking
import (
	"net"
	"errors"
	"encoding/json"
	"constant"
	"net/http"
	"io/ioutil"
	"log"
	"os"
	"time"
	"io"
	"bytes"
	"mime/multipart"
)

var	c *http.Client = &http.Client{Timeout: time.Second * 3}

func GetLocalIP() (string, error) {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip == nil || ip.IsLoopback() {
			continue
		}
		ip = ip.To4()
		if ip == nil {
			continue // not an ipv4 address
		}
		return ip.String(), nil
	}
	return "", errors.New("Cannot find IP address, please check network connection")
}


func UDPsend(ip string, port string, message []byte ) error {
	addr, err := net.ResolveUDPAddr("udp", ip+":"+port)
		if err != nil {
			return err
		}
		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			return err
		}
		defer conn.Close()
		_, err = conn.Write(message)
		if err != nil {
			return err
		}
	return nil
}

func UDPlisten(port string, callback func(message []byte) error)error{
	port = ":" + port
	addr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	defer conn.Close()
	buffer := make([]byte, 4096)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return err
		}
		callback(buffer[0:n])
	}
}
// UDP messages
// client to master
func EncodeUDPMessageClient2Master(list *constant.UDPMessageClient2Master) ([]byte, error){
	message, err := json.Marshal(list)
	return message, err
}
func DecodeUDPMessageClient2Master(message []byte) (*constant.UDPMessageClient2Master, error) {
	list := &constant.UDPMessageClient2Master{}
	err := json.Unmarshal(message, list)

	return list, err
} 
// master to client
func EncodeUDPMessageMaster2Client(list *constant.UDPMessageMaster2Client) ([]byte, error){
	message, err := json.Marshal(list)
	return message, err
}
func DecodeUDPMessageMaster2Client(message []byte) (*constant.UDPMessageMaster2Client, error) {
	list := &constant.UDPMessageMaster2Client{}
	err := json.Unmarshal(message, list)

	return list, err
} 
// datanode to master 
func EncodeUDPMessageDatanode2Master(list *constant.UDPMessageDatanode2Master) ([]byte, error){
	message, err := json.Marshal(list)
	return message, err
}
func DecodeUDPMessageDatanode2Master(message []byte) (*constant.UDPMessageDatanode2Master, error) {
	list := &constant.UDPMessageDatanode2Master{}
	err := json.Unmarshal(message, list)

	return list, err
} 

// func TCPsend(ip string, port string, message []byte) {


// }

// func TCPlisten(port string, callback func(message []byte)) {
// 	port = ":" + port
// 	handleConnection := func(conn TCPconn) {
// 		for{
// 			buffer := make([]byte, 4096)
// 			n, err := conn.Read(buffer)
// 			if err != nil{
// 				panic(err)
// 			}
// 			callback(buffer[0:n])
// 		}
// 	}
// 	ln, err := net.Listen("tcp", port)
// 	if err != nil {
// 		panic(err)
// 	}
// 	for {
// 		conn, err := ln.Accept()
// 		if err != nil {
// 			panic(err)
// 		}
// 		go handleConnection(conn)
// 	}
// }
func HTTPsend(url string)[]byte{
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		panic(err)
	}
	return body
}

func HTTPlisten(endpoint string, handler func(w http.ResponseWriter, req *http.Request)){
	http.HandleFunc(endpoint, handler)
}
func HTTPfileServer(port string){
	fs := http.FileServer(http.Dir("/Users/chenxinhang/Downloads"))
	port = ":" + port
	log.Fatal(http.ListenAndServe(port, fs))
}
func HTTPuploadFile(url string, filename string, uploadFilename string) []byte {
    buf := new(bytes.Buffer)
    writer := multipart.NewWriter(buf)
    formFile, err := writer.CreateFormFile("uploadfile", uploadFilename)
    if err != nil {
        log.Fatalf("Create form file failed: %s\n", err)
    }
    srcFile, err := os.Open(filename)
    if err != nil {
        log.Fatalf("%Open source file failed: s\n", err)
    }
    defer srcFile.Close()
    _, err = io.Copy(formFile, srcFile)
    if err != nil {
        log.Fatalf("Write to form file falied: %s\n", err)
    }
    contentType := writer.FormDataContentType()
    writer.Close() 
    resp, err := http.Post(url, contentType, buf)
    if err != nil {
        log.Fatalf("Post failed: %s\n", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		panic(err)
	}
	return body
}
func HTTPlistenDownload(BaseUploadPath string){
	Download := func(w http.ResponseWriter, r *http.Request){
		formFile, header, err := r.FormFile("uploadfile")
    if err != nil {
        log.Printf("Get form file failed: %s\n", err)
        return
    }
	defer formFile.Close()
	
    destFile, err := os.Create("./" + header.Filename)
    if err != nil {
        log.Printf("Create failed: %s\n", err)
        return
    }
    defer destFile.Close()

    _, err = io.Copy(destFile, formFile)
		if err != nil {
			log.Printf("Write file failed: %s\n", err)
			return
		}
	}
	http.HandleFunc("/put", Download)
}
func HTTPstart(port string){
	port = ":" + port
	log.Fatal(http.ListenAndServe(port, nil))
}
