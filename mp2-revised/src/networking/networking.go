package networking

import (
	"bytes"
	"constant"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	. "structs"
	"time"
	"encoding/json"
	"fmt"
)

var c *http.Client = &http.Client{Timeout: time.Second * 3}

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
func HTTPsend(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return body
}

//HTTPlisten function
func HTTPlisten(endpoint string, handler func(w http.ResponseWriter, req *http.Request)) {
	http.HandleFunc(endpoint, handler)
}

//HTTPfileServer
func HTTPfileServer(port string, dir string) {
	fs := http.FileServer(http.Dir(dir))
	port = ":" + port
	log.Fatal(http.ListenAndServe(port, fs))
}

//HTTPuploadFile
func HTTPuploadFile(url string, filename string, uploadFilename string) []byte {
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
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
	contentType := writer.FormDataContentType()
	writer.Close()
	resp, err := http.Post(url, contentType, buf)
	if err != nil {
		Logger.Fatal("Post failed: %s\n", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return body
}
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

func UploadFileToDatanode(filename string, remotefilename string, ipPort string) string {
	url := "http://" + ipPort + "/putfile"
	//Write2Shell("Upload file to url:" + url)
	body := HTTPuploadFile(url, filename, remotefilename)
	//Write2Shell("Url: " + url + " Status: " + string(body))
	return string(body)
}

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
		filenameWithPath := constant.Dir + "files_" + constant.DatanodeHTTPServerPort + "/" + filename
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

func List() []string {
	var c, err = ioutil.ReadDir(constant.Dir + "files_" + constant.DatanodeHTTPServerPort) 
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var output []string
	for _, entry := range c {
		output = append(output, entry.Name())
	}
	return output
}

func HTTPlistenRecover() {
	Recover := func(w http.ResponseWriter, r *http.Request) {
		list := List()
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

func HTTPstart(port string) {
	port = ":" + port
	log.Fatal(http.ListenAndServe(port, nil))
}

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
