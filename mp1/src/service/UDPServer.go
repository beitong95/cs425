package service

import (
	"fmt"
	"sync"
	"time"
)

func UDPServer(isAll2All bool, isIntroducer bool, wg *sync.WaitGroup, c chan int64) {
	defer wg.Done()
	for {
		fmt.Println("UDPServer", <-c)
		time.Sleep(1 * time.Second)
	}
}
