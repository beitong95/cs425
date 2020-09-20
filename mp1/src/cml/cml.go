package cml

import (
	"fmt"
	"sync"
	"time"
)

func Cml(wg *sync.WaitGroup, c chan int64) {
	defer wg.Done()

	for {
		currentTime := time.Now().Unix()
		fmt.Println("CML send ", currentTime)
		c <- currentTime
		time.Sleep(1 * time.Second)
	}
}
