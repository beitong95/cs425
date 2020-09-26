package cli

import (
	"bufio"
	"fmt"
	"helper"
	"os"
	"service"
	"strings"
	. "structs"
	"sync"
)

// Cli command line function
func Cli(wg *sync.WaitGroup, c chan int) {
	defer wg.Done()
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Simple Shell")
	fmt.Println("---------------------")
	commands := []string{"help", "all2all", "gossip", "leave", "join", "id", "list", "kill"}
	for {
		fmt.Print("-> ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.Replace(cmd, "\r\n", "", -1)
		cmd = strings.Replace(cmd, "\n", "", -1)

		wrongCommand := true
		for i := 0; i < len(commands); i++ {
			if commands[i] == cmd {
				wrongCommand = false
			}
		}

		fmt.Printf("CLI send %s to UDP server\n", cmd)

		if cmd == "help" || wrongCommand == true {
			fmt.Println("help    -> help information")
			fmt.Println("all2all -> change multicast to all2all")
			fmt.Println("gossip  -> change multicast to gossip")
			fmt.Println("leave   -> leave current group")
			fmt.Println("join    -> join current group")
			fmt.Println("id      -> print current id")
			fmt.Println("list    -> print current membershipList")
			fmt.Println("kill    -> fail myself")
			continue
		}

		switch cmd {
		case "all2all":
			c <- CHANGE_TO_ALL2ALL
		case "gossip":
			c <- CHANGE_TO_GOSSIP
		case "leave":
			c <- LEAVE_GROUP
		case "join":
			c <- JOIN_GROUP
		case "id":
			fmt.Println("ID:", MyID)
		case "list":
			service.MT.Lock()
			helper.PrintMembershipListAsTable(MembershipList)
			service.MT.Unlock()
		case "kill":
			os.Exit(1)
		}
	}
}
