package cli

import (
	"os"
	"github.com/marcusolsson/tui-go"
	"bufio"
	"fmt"
	"strings"
	. "structs"
	"time"
	"sync"
	"helper"
	"log"
)

var commands = []string{"get", "put", "delete", "store", "ls", "exit", "help", "all2all", "gossip", "leave", "join", "id", "list", "para"}

func getHelp() string{
	return `help                            -> help inFormation
			mp2
			get sdfsfilename localfilename  -> read file from HDFS
			put localfilename sdfsfilename  -> write file to HDFS
			delete sdfsfilename             -> delete file in HDFS	
			ls sdfsfilename                 -> list all VMs where the file is stored
			store                           -> list files stored on local disk
			mp1
			all2all                         -> change multicast to all2all
			gossip                          -> change multicast to gossip
			leave                           -> leave current group
			join                            -> join current group
			id                              -> print current id
			list                            -> list membership list
			para                            -> print all gossip and all2all parameter
			exit                            -> exit from HDFS`
}

var (
	input *tui.Entry
	shell *tui.Box
	bandwidthBox *tui.Box
	bandwidthBoxLabel *tui.Label
	protocolBox *tui.Box
	protocolBoxLabel *tui.Label
	currentStatusBox *tui.Box
	currentStatusBoxLabel *tui.Label
	membershipBox *tui.Box 
	membershipBoxLabel *tui.Label
	ui tui.UI
)

// Cli command line function
func Cli(wg *sync.WaitGroup, c chan int) {
	defer wg.Done()
	History, input, shell = CreateShell()
	bandwidthBox, bandwidthBoxLabel = CreateBandwidthBox()
	protocolBox, protocolBoxLabel = CreateProtocolBox()
	currentStatusBox, currentStatusBoxLabel = CreateCurrentStatusBox()
	membershipBox, membershipBoxLabel = CreateMembershipBox()
	allStatusBox := tui.NewHBox(bandwidthBox, protocolBox, currentStatusBox)
	root := tui.NewVBox(membershipBox, allStatusBox, shell)
	ui,_ = tui.New(root)	
	done := make(chan string)

	// shell logic
	input.OnSubmit(func(e *tui.Entry) {
		// rejoin cmd
		_cmd := e.Text()[2:]
		cmd, _, _ := ParseCmd(input, _cmd, commands)
		
		if cmd == "" {
			// wrong command
			return
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
			Write2Shell(MyID)
		case "list":
			s, err := helper.PrintMembershipListAsTableInGUI(MembershipList)
			if err != nil {
				log.Fatal("PrintMembershipListAsTableInGUI error")
			}
			Write2Shell(s)
		case "para":
			Write2Shell("Tgossip: " + fmt.Sprintf("%v",Tgossip))
			Write2Shell("Tall2all: " + fmt.Sprintf("%v",Tall2all))
			Write2Shell("Tfail: " + fmt.Sprintf("%v",Tfail))
			Write2Shell("Tclean: " + fmt.Sprintf("%v",Tclean))
			Write2Shell("B: " + fmt.Sprintf("%v",B))
		case "help":
			Write2Shell(getHelp())
		case "get":
			Write2Shell(cmd)
		case "put":
			Write2Shell(cmd)
		case "delete":
			Write2Shell(cmd)
		case "ls":
			Write2Shell(cmd)
		case "store":
			Write2Shell(cmd)
		case "exit":
			Write2Shell(cmd)
			time.Sleep(time.Duration(500) * time.Millisecond)
			ui.Quit()
			done <- "Done"
			os.Exit(1)
			}
	})

	ui.SetKeybinding("Esc", func() {
		ui.Quit()
		done <- "Done"
		os.Exit(1)
	})
	go ui.Run()
	go AutoUpdateCLI(ui)
	<-done
}


var reader *bufio.Reader
// Cli command line function
func CliSimple(wg *sync.WaitGroup, c chan int) {
	defer wg.Done()
	reader = bufio.NewReader(os.Stdin)
	fmt.Println("Simple Shell")
	fmt.Println("---------------------")
	for {
		fmt.Print("-> ")
		_cmd, _ := reader.ReadString('\n')

		cmd, _, _ := ParseCmdSimple(_cmd, commands)
		if cmd == "" {
			// wrong command
			continue
		}

		fmt.Printf("CLI send %s to UDP server\n", cmd)

		switch cmd {
		case "help":
			fmt.Println(strings.Replace(getHelp(),"\t","",-1))
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
			helper.PrintMembershipListAsTable(MembershipList)
		case "get":
			fmt.Println(cmd)
		case "put":
			fmt.Println(cmd)
		case "delete":
			fmt.Println(cmd)
		case "ls":
			fmt.Println(cmd)
		case "store":
			fmt.Println(cmd)
		case "exit":
			os.Exit(1)
		}
	}
}