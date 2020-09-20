package cli

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

// Cli command line function
func Cli(wg *sync.WaitGroup, c chan int) {
	defer wg.Done()

	time.Sleep(5 * time.Second)
	cmd := 1
	for {
		log.Printf("CLI send %d to UDP server\n", cmd)
		c <- cmd
		cmd = rand.Intn(4) + 1
		time.Sleep(500 * time.Millisecond)
	}
}
