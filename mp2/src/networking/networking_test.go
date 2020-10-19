package networking_test

import (
	"testing"
	"fmt"
	"networking"
	"net/http"
	"html"
)

func TestUDP(t *testing.T) {
	fmt.Println("test UDP")
	f := func(message []byte) error{
		fmt.Println(string(message))
		return nil
    }
	go networking.UDPlisten("2020", f)
	for {
	networking.UDPsend("127.0.0.1", "2020", []byte("hello test"))
	}
    
}

func TestTCP(t* testing.T){
	fmt.Println("test TCP")
}

func TestHTTP(t *testing.T){
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("some wrong with hello"))
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	}
	networking.HTTPlisten("/hello", handler)
	networking.HTTPlistenDownload("./")
	go networking.HTTPfileServer("5000")
	go networking.HTTPstart("3000")
	content := networking.HTTPuploadFile("http://127.0.0.1:3000/put", "networking_test.go", "test")
	fmt.Println(string(content))
	// body := networking.HTTPsend("http://127.0.0.1:3000/hello")
	// fmt.Println(string(body))
	
}