package client

import (
	"os"
	"github.com/marcusolsson/tui-go"
	"bufio"
	"fmt"
	"strings"
	"cli"
	"constant"
	"time"
)

var commandsClient = []string{"get", "put", "delete", "store", "ls", "exit", "help"}

func getClientHelp() string{
	return `help                            -> help inFormation
			get sdfsfilename localfilename  -> read file from HDFS
			put localfilename sdfsfilename  -> write file to HDFS
			delete sdfsfilename             -> delete file in HDFS	
			ls sdfsfilename                 -> list all VMs where the file is stored
			store                           -> list files stored on local disk
			exit                            -> exit from HDFS`
}
var (
	history *tui.Box
	input *tui.Entry
	shell *tui.Box
	clientMasterStatusBox *tui.Box
	clientMasterStatusLabel *tui.Label
	ui tui.UI
)

// Cli command line function
func cliClient() {
	history, input, shell = cli.CreateShell()
	clientMasterStatusBox, clientMasterStatusLabel = cli.CreateClientMasterStatusBox()
	root := tui.NewVBox(clientMasterStatusBox, shell)
	ui,_ = tui.New(root)	
	done := make(chan string)
	// shell logic
	input.OnSubmit(func(e *tui.Entry) {
		// rejoin cmd
		_cmd := e.Text()[2:]
		if constant.IsKickout == true {
			rejoinCmd := _cmd
			if rejoinCmd == "Y" {
				cli.Write2Shell(history, "Y")
				constant.KickoutRejoinCmd <- "true"
			} else {
				cli.Write2Shell(history, "N")
				constant.KickoutRejoinCmd <- "false"
				ui.Quit()
				done <- "Done"
				os.Exit(1)
			}
		} else {
			cmd, filename1, filename2 := cli.ParseCmd(history,input, _cmd, commandsClient)
			if cmd == "" {
				// wrong command
				return
			}
			switch cmd {
			case "help":
				cli.Write2Shell(history,getClientHelp())
			case "get":
				cmdQueue.Enqueue([]string{cmd, filename1, filename2})
			case "put":
				cmdQueue.Enqueue([]string{cmd, filename1, filename2})
			case "delete":
				cmdQueue.Enqueue([]string{cmd, filename1, filename2})
			case "ls":
				cmdQueue.Enqueue([]string{cmd, filename1, filename2})
			case "store":
				cmdQueue.Enqueue([]string{cmd, filename1, filename2})
			case "exit":
				time.Sleep(time.Duration(500) * time.Millisecond)
				ui.Quit()
				done <- "Done"
				os.Exit(1)
			}
		} 
	})

	ui.SetKeybinding("Esc", func() {
		ui.Quit()
		done <- "Done"
		os.Exit(1)
	})
	go ui.Run()
	go cli.AutoUpdateCLI(ui)
	<-done
}


var reader *bufio.Reader
// Cli command line function
func cliSimpleClient() {
	reader = bufio.NewReader(os.Stdin)
	fmt.Println("Simple Shell")
	fmt.Println("---------------------")
	for {
		fmt.Print("-> ")
		cmd, _ := reader.ReadString('\n')

		if constant.IsKickout == true {
			rejoinCmd := cmd
			if rejoinCmd == "Y" {
				constant.KickoutRejoinCmd <- "true"
			} else {
				constant.KickoutRejoinCmd <- "false"
				os.Exit(1)
			}
		} else {
			cmd, filename1, filename2 := cli.ParseCmdSimple(cmd, commandsClient)
			if cmd == "" {
				// wrong command
				continue
			}

			fmt.Printf("CLI send %s to UDP server\n", cmd)

			switch cmd {
			case "help":
				fmt.Println(strings.Replace(getClientHelp(),"\t","",-1))
			case "get":
				cmdQueue.Enqueue([]string{cmd, filename1, filename2})
			case "put":
				cmdQueue.Enqueue([]string{cmd, filename1, filename2})
			case "delete":
				cmdQueue.Enqueue([]string{cmd, filename1, filename2})
			case "ls":
				cmdQueue.Enqueue([]string{cmd, filename1, filename2})
			case "store":
				cmdQueue.Enqueue([]string{cmd, filename1, filename2})
			case "exit":
				os.Exit(1)
			}

		}

	}
}