package service

import (
	"log"
	"sync"
	"time"
)

//UDPServer is the udp server thread function
func UDPServer(isAll2All bool, isIntroducer bool, wg *sync.WaitGroup, c chan int) {
	defer wg.Done()
	gossipPeriodMillisecond := 2000
	//timer for gossip period
	ticker := time.NewTicker(time.Duration(gossipPeriodMillisecond) * time.Millisecond)
	//command from CML
	cmd := 0
	gossipCounter := 0
	for {
		// can go through here ever gossipPeriod
		log.Println("waiting for next gossip period")
		t1 := time.Now()
		//no wait means our gossip period is too short for gossip process
		<-ticker.C
		t2 := time.Now()
		diff := t2.Sub(t1)
		log.Println("wait time:", diff)
		if float32(diff/time.Millisecond) < float32(float32(gossipPeriodMillisecond)*0.05) {
			log.Fatalln("gossip period time too short")
		}
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
		// TODO: Gossip logic
		time.Sleep(1500 * time.Millisecond)
		log.Println("Finish Gossip work")
	}
}
