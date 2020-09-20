package service

import (
	"log"
	"sync"
	"time"
)

//UDPServer is the udp server thread function
func UDPServer(isAll2All bool, isIntroducer bool, wg *sync.WaitGroup, c chan int) {
	defer wg.Done()
	gossipPeriod := 2 * time.Second
	//timer for gossip period
	ticker := time.NewTicker(gossipPeriod)
	//command from CML
	cmd := 0
	gossipCounter := 0
	go func() {
		for {
			// can go through here ever gossipPeriod
			log.Println("waiting for next gossip period")
			<-ticker.C
			gossipCounter = gossipCounter + 1
			log.Println("----------------------------------------------")
			log.Println("Start gossip period", gossipCounter)
			// in every gossipPeriod, the first thing is to read commands from CML
			cmds := make([]int, 0)
		forLoop:
			for {
				select {
				case cmd = <-c:
					log.Printf("UDPServer receives cmd from CML in %d gossip period: %d\n", gossipCounter, cmd)
					cmds = append(cmds, cmd)
				default:
					if len(cmds) == 0 {
						log.Println("No command from CML. Do nothing")
					} else {
						log.Println("No more commands.")
						log.Println("Commands received:", cmds)
					}
					break forLoop
				}
			}
			log.Println("Doing Gossip work with commands", cmds)
			time.Sleep(1 * time.Second)
			log.Println("Finish Gossip work")
		}
	}()
}
