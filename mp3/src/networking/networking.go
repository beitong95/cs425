package networking

import (
	. "structs"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"time"
	"helper"
	"strings"
)

var c *http.Client = &http.Client{Timeout: time.Second * 3}

/*
Function name: GetLocalIP
Description: get local machine ip
Input: 
OutPut: local ip address
*/
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

// UDP services

/*
Function name: UDPsend
Description: send udp message to a vm 
Input: ip, port(udp port), message
OutPut: err
*/
func UDPsend(ip string, port string, message []byte) error {
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

/*
Function name: UDPlisten
Description: start to receive udp message on a port, and use the callback function to process the message
Input: port(udp port), call back function
OutPut: err
*/
func UDPlisten(port string, callback func(message []byte) error) error {
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


// HTTP services

/*
Function name: HTTPsend
Description: send http message to a vm and get the body 
Input: url(contain ip, port and endpoint, and key value pairs)
OutPut: body
*/
func HTTPsend(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	return body
}

// HTTP service used by master
// here we need to unify the http services in master and datanode

// listen to a endpoint and execute the call back function 
func HTTPlisten(endpoint string, handler func(w http.ResponseWriter, req *http.Request)) {
	http.HandleFunc(endpoint, handler)
}
// all callback functions are defined in master/server.go

// HTTP services used by master and client 

// upload a single file to a datanode, use pipe
func HTTPuploadFile(url string, filename string, uploadFilename string) []byte {
	r, w := io.Pipe()
	writer := multipart.NewWriter(w)

	go func() {
		defer w.Close()
		defer writer.Close()

		formFile, err := writer.CreateFormFile("uploadfile", uploadFilename)
		if err != nil {
			Logger.Fatal("Create form file failed: %s\n", err)
		}

		srcFile, err := os.Open(filename)

		if err != nil {
			Logger.Fatal("%Open source file failed: s\n", err)
		}

		defer srcFile.Close()
		_, err = io.Copy(formFile, srcFile)
		if err != nil {
			Logger.Fatal("Write to form file falied: %s\n", err)
		}
	}()

	contentType := writer.FormDataContentType()

	resp, err := http.Post(url, contentType, r)

	if err != nil {
		Logger.Fatal("Post failed: %s\n", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	return body
}

// a wrapper to upload file
func UploadFileToDatanode(filename string, remotefilename string, ipPort string) string {
	url := "http://" + ipPort + "/putfile"
	Write2Shell("Upload file to url:" + url)
	body := HTTPuploadFile(url, filename, remotefilename)
	Write2Shell("Url: " + url + " Status: " + string(body))
	return string(body)
}

// upload file to datanode
func UploadFileToWorkers(filename string, remotefilename string, ipPort string) string {
	url := "http://" + ipPort + "/mapleWorker"
	Write2Shell("Upload file to url:" + url)
	body := HTTPuploadFile(url, filename, remotefilename)
	Write2Shell("Url: " + url + " Status: " + string(body))
	return string(body)
}

// download a single file to local disk
func DownloadFileFromDatanode(filename string, localfilename string, ipPort string) (string, error) {
	url := "http://" + ipPort + "/" + filename
	//Write2Shell("Downloading file from: " + url)
	rsp, err := http.Get(url)
	//Write2Shell("get http.Get return")
	if err != nil {
		return "Connection error", err
	}
	if rsp.Header["Content-Length"][0] == "19" {
		fmt.Println("Possible empty")
		buffer := make([]byte, 19)
		rsp.Body.Read(buffer)
		if string(buffer) == "404 page not found\n" {
			return "File not found", errors.New("networking: file not found")
		} else {
			file := strings.NewReader(string(buffer))
			// store in main folder
			destFile, err := os.Create("./" + localfilename)
			if err != nil {
				log.Printf("Create file failed: %s\n", err)
				return "Create Failed", err
			}
			_, err = io.Copy(destFile, file)
			if err != nil {
				log.Printf("Write file failed: %s\n", err)
				return "Write error", err
			}
			return "OK", nil
		}
	}
	// store in main folder
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
	return "OK", nil
}

// delete a file just send a request
func DeleteFileFromDatanode(remotefilename string, ipPort string) string {
	url := "http://" + ipPort + "/deletefile?file=" + remotefilename
	Write2Shell("Delete file from url:" + url)
	body := HTTPsend(url)
	Write2Shell("Url: " + url + " Status: " + string(body))
	return string(body)
}

// HTTP services used by datanode

// handle get
func HTTPfileServer(port string, dir string) {
	fs := http.FileServer(http.Dir(dir))
	port = ":" + port
	log.Fatal(http.ListenAndServe(port, fs))
}

// handle put
// use this as template for maple
func HTTPlistenDownload(BaseUploadPath string) {
	Download := func(w http.ResponseWriter, r *http.Request) {
		formFile, header, err := r.FormFile("uploadfile")
		if err != nil {
			log.Printf("Get form file failed: %s\n", err)
			//TODO: w.write add return status
			w.Write([]byte("error"))
			return
		}
		defer formFile.Close()

		destFile, err := os.Create(BaseUploadPath + header.Filename)
		if err != nil {
			log.Printf("Create failed: %s\n", err)
			w.Write([]byte("error"))
			return
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, formFile)
		if err != nil {
			log.Printf("Write file failed: %s\n", err)
			w.Write([]byte("error"))
			return
		}
		w.Write([]byte("OK"))
	}
	http.HandleFunc("/putfile", Download)
}

// handle rereplica
func HTTPlistenRereplica() {
	Rereplica := func(w http.ResponseWriter, r *http.Request) {
		// get filename
		filenames, ok := r.URL.Query()["file"]
		if !ok {
			Logger.Error("Handle rereplica but the key is missing")
			return
		}
		filename := filenames[0]

		// get desitnation
		destinations, ok := r.URL.Query()["destination"]
		if !ok {
			Logger.Error("Handle rereplica but the key is missing")
			return
		}
		destination := destinations[0]

		//send
		filenameWithPath := Dir + "files_" + DatanodeHTTPServerPort + "/" + filename
		ipPort := IP2DatanodeUploadIP(destination)
		status := UploadFileToDatanode(filenameWithPath, filename, ipPort)
		if status != "OK" {
			w.Write([]byte("Bad"))
		} else {
			w.Write([]byte("OK"))
		}
	}
	http.HandleFunc("/rereplica", Rereplica)
}

// handle recover
func HTTPlistenRecover() {
	Recover := func(w http.ResponseWriter, r *http.Request) {
		list := helper.List()
		var res []byte
		var err error
		if len(list) == 0 {
			res = []byte("[]")
		} else {
			res, err = json.Marshal(list)
			if err != nil {
				Write2Shell("Unmarshal error in HTTPlistenRecover")
			}
		}
		w.Write(res)
	}
	http.HandleFunc("/recover", Recover)
}

//handle delete
func HTTPlistenDelete(BaseDeletePath string) {
	Delete := func(w http.ResponseWriter, r *http.Request) {
		file, ok := r.URL.Query()["file"]
		if !ok {
			log.Println("Get IPs Url Param 'key' is missing")
			return
		}
		filename := file[0]
		err := os.Remove(BaseDeletePath + filename)
		if err != nil {
			w.Write([]byte("Write Failed"))

		} else {
			w.Write([]byte("OK"))
		}
	}
	http.HandleFunc("/deletefile", Delete)
}


// start all http.HandleFunc()
func HTTPstart(port string) {
	port = ":" + port
	log.Fatal(http.ListenAndServe(port, nil))
}


// TODO:
// 1. master networking.HTTPlisten -> http.handlefunc() -> callback -> networking.HTTPstart -> http.ListenAndServe
// 2. datanode: networking.HTTPlistenDownload() -> http.handlefunc() -> callback -> networking.HTTPstart -> http.ListenAndServe