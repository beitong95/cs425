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

var commandsClient = []string{"get", "set", "delete", "store", "exit", "help"}

func getClientHelp() string{
	return `help                        -> help inFormation
			get filename                -> read file from HDFS
			set filename (newfilename)  -> write file to HDFS
			delete filename             -> delete file in HDFS	
			store                       -> list files stored on local disk
			exit                        -> exit from HDFS`
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
			cmd, _:= cli.ParseCmd(history,input, _cmd, commandsClient)
			if cmd == "" {
				// wrong command
				return
			}
			switch cmd {
			case "help":
				cli.Write2Shell(history,getClientHelp())
			case "get":
				cli.Write2Shell(history, "TODO")
			case "set":
				cli.Write2Shell(history, "TODO")
			case "delete":
				cli.Write2Shell(history, "TODO")
			case "store":
				cli.Write2Shell(history, "TODO")
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
			cmd, _ = cli.ParseCmdSimple(cmd, commandsClient)
			if cmd == "" {
				// wrong command
				continue
			}

			fmt.Printf("CLI send %s to UDP server\n", cmd)

			switch cmd {
			case "help":
				fmt.Println(strings.Replace(getClientHelp(),"\t","",-1))
			case "get":
				fmt.Println("TODO")
			case "set":
				fmt.Println("TODO")
			case "delete":
				fmt.Println("TODO")
			case "store":
				fmt.Println("TODO")
			case "exit":
				os.Exit(1)
			}

		}

	}
}