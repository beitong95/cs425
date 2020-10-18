package networking_test

import (
	"testing"
	"fmt"
	"networking"
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