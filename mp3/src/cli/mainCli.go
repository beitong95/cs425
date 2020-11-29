package cli

import (
	"bufio"
	"client"
	"fmt"
	"helper"
	"log"
	"os"
	"strings"
	. "structs"
	"sync"
	"time"
	"github.com/marcusolsson/tui-go"
)

var commands = []string{"get", "put", "delete", "store", "ls", "exit", "help", "all2all", "gossip", "leave", "join", "id", "list", "para", "maple", "juice", "vote", "tree", "vote_large"}

func getHelp() string {
	return `help                            -> help inFormation
			mp3
			maple <maple_exe> <num_maples> <sdfs_intermediate_filename_prefix> <sdfs_src_directory>
			juice <juice_exe> <num_juices> <sdfs_intermediate_filename_prefix> <sdfs_dest_filename> <delete 0 1>
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
	input                 *tui.Entry
	shell                 *tui.Box
	bandwidthBox          *tui.Box
	bandwidthBoxLabel     *tui.Label
	protocolBox           *tui.Box
	protocolBoxLabel      *tui.Label
	currentStatusBox      *tui.Box
	currentStatusBoxLabel *tui.Label
	membershipBox         *tui.Box
	membershipBoxLabel    *tui.Label
	ui                    tui.UI
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
	ui, _ = tui.New(root)
	done := make(chan string)

	// shell logic
	input.OnSubmit(func(e *tui.Entry) {
		// rejoin cmd
		_cmd := e.Text()[2:]
		// mp2
		cmd, filename1, filename2 := ParseCmd(input, _cmd, commands)

		if cmd == "" {
			// wrong command
			return
		}
		switch cmd {
		case "vote":
			//filename 1 is maple count filename2 is juice count
			go client.MapleJuice("voteMaple", filename1, "vote", "votes.txt", "maplecommand", "countJuice", filename2, "voteOut.txt", "1", "juicecommand")
		case "vote_large":
			//filename 1 is maple count filename2 is juice count
			go client.MapleJuice("voteMaple", filename1, "vote", "votes_large.txt", "maplecommand", "countJuice", filename2, "vote_large_Out.txt", "1", "juicecommand")
		case "tree":
			//filename 1 is maple count filename2 is juice count
			go client.MapleJuice("treeMaple", filename1, "tree", "treetype.txt", "maplecommand", "countJuice", filename2, "treeOut.txt", "1", "juicecommand")
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
			Write2Shell("Tgossip: " + fmt.Sprintf("%v", Tgossip))
			Write2Shell("Tall2all: " + fmt.Sprintf("%v", Tall2all))
			Write2Shell("Tfail: " + fmt.Sprintf("%v", Tfail))
			Write2Shell("Tclean: " + fmt.Sprintf("%v", Tclean))
			Write2Shell("B: " + fmt.Sprintf("%v", B))
		case "help":
			Write2Shell(getHelp())
		case "get":
			go client.GetFile(filename1, filename2)
		case "put":
			go client.PutFile(filename1, filename2)
		case "delete":
			go client.DeleteFile(filename1)
		case "ls":
			go client.Ls(filename1)
		case "store":
			go client.Store()
		case "maple":
			cmds := strings.Fields(_cmd)
			if len(cmds) != 5 {
				Write2Shell("Wrong maple cmd")
			} else {
				maple_exe := cmds[1]
				num_maples := cmds[2]
				sdfs_intermediate_filename_prefix := cmds[3]
				sdfs_src_directory := cmds[4]
				go client.Maple(maple_exe, num_maples, sdfs_intermediate_filename_prefix, sdfs_src_directory, _cmd)
			}
		case "juice":
			cmds := strings.Fields(_cmd)
			if len(cmds) != 6 {
				Write2Shell("Wrong juice cmd")
			} else {
				juice_exe := cmds[1]
				num_juices := cmds[2]
				sdfs_intermediate_filename_prefix := cmds[3]
				sdfs_dest_filename := cmds[4]
				delete_input := cmds[5]
				go client.Juice(juice_exe, num_juices, sdfs_intermediate_filename_prefix, sdfs_dest_filename, delete_input, _cmd)
			}
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

		cmd, filename1, filename2 := ParseCmdSimple(_cmd, commands)
		if cmd == "" {
			// wrong command
			continue
		}

		fmt.Printf("CLI send %s to UDP server\n", cmd)

		switch cmd {
		case "vote":
			go client.MapleJuice("voteMaple", filename1, "vote", "votes.txt", "maplecommand", "countJuice", filename2, "out.txt", "1", "juicecommand")
		case "vote_large":
			//filename 1 is maple count filename2 is juice count
			go client.MapleJuice("voteMaple", filename1, "vote", "votes_large.txt", "maplecommand", "countJuice", filename2, "vote_large_Out.txt", "1", "juicecommand")
		case "tree":
			//filename 1 is maple count filename2 is juice count
			go client.MapleJuice("treeMaple", filename1, "tree", "treetype.txt", "maplecommand", "countJuice", filename2, "treeOut.txt", "1", "juicecommand")
		case "help":
			fmt.Println(strings.Replace(getHelp(), "\t", "", -1))
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
			go client.GetFile(filename1, filename2)
		case "put":
			fmt.Println(cmd)
			go client.PutFile(filename1, filename2)
		case "delete":
			go client.DeleteFile(filename1)
		case "ls":
			go client.Ls(filename1)
		case "store":
			go client.Store()
		case "maple":
			cmds := strings.Fields(_cmd)
			if len(cmds) != 5 {
				fmt.Println("Wrong maple cmd")
			} else {
				maple_exe := cmds[1]
				num_maples := cmds[2]
				sdfs_intermediate_filename_prefix := cmds[3]
				sdfs_src_directory := cmds[4]
				go client.Maple(maple_exe, num_maples, sdfs_intermediate_filename_prefix, sdfs_src_directory, _cmd)
			}
		case "juice":
			cmds := strings.Fields(_cmd)
			if len(cmds) != 6 {
				Write2Shell("Wrong juice cmd")
			} else {
				juice_exe := cmds[1]
				num_juices := cmds[2]
				sdfs_intermediate_filename_prefix := cmds[3]
				sdfs_dest_filename := cmds[4]
				delete_input := cmds[5]
				go client.Juice(juice_exe, num_juices, sdfs_intermediate_filename_prefix, sdfs_dest_filename, delete_input, _cmd)
			}
		case "exit":
			os.Exit(1)
		}
	}
}
