package cml

import (
	"log"
	"sync"
	"time"
)

// Cml command line function
func Cml(wg *sync.WaitGroup, c chan int) {
	defer wg.Done()

	time.Sleep(5 * time.Second)
	cmd := 1
	for {
		log.Printf("CML send %d to UDP server\n", cmd)
		c <- cmd
		cmd = cmd + 1
		time.Sleep(500 * time.Millisecond)
	}
}
