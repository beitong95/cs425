package networking

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"time"
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
	http.HandleFunc("/putfile", Download)
}
func HTTPstart(port string) {
	port = ":" + port
	log.Fatal(http.ListenAndServe(port, nil))
}
