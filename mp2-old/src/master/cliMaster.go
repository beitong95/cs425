package master

import (
	"os"
	"github.com/marcusolsson/tui-go"
	"time"
	"fmt"
	"strings"
	"bufio"
	"cli"
	
)
var commandsMaster = []string{"ls", "store", "exit", "help"} 

func getMasterHelp() string {
	return `help                        -> help inFormation
			ls filename                 -> list where is this file stores in HDFS
			store                       -> list files stored on local disk
			exit                        -> exit from HDFS`
}
var (
	history *tui.Box
	input *tui.Entry
	shell *tui.Box
	masterMembershipBox *tui.Box
	masterClientMembershipLabel *tui.Label
	masterDatanodeMembershipLabel *tui.Label
	ui tui.UI
)

// Cli command line function
func cliMaster() {
	history, input, shell = cli.CreateShell()
	masterMembershipBox, masterClientMembershipLabel, masterDatanodeMembershipLabel = cli.CreateMasterMembershipBox()
	root := tui.NewVBox(masterMembershipBox, shell)
	ui,_ = tui.New(root)	
	done := make(chan string)
	// shell logic
	input.OnSubmit(func(e *tui.Entry) {
		cmd, filename1, _:= cli.ParseCmd(history, input,e.Text()[2:], commandsMaster)
		if cmd == "" {
			// wrong command
			return
		}
		switch cmd {
		case "help":
			cli.Write2Shell(history, getMasterHelp())
		case "ls":
			cli.Write2Shell(history, filename1)
		case "store":
			cli.Write2Shell(history, "TODO")
		case "exit":
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
	go cli.AutoUpdateCLI(ui)
	<-done
}

var reader *bufio.Reader 
// Cli command line function
func cliSimpleMaster() {
	reader = bufio.NewReader(os.Stdin)
	fmt.Println("Simple Shell")
	fmt.Println("---------------------")
	for {
		fmt.Print("-> ")
		cmd, _ := reader.ReadString('\n')
		cmd, filename1, _ := cli.ParseCmdSimple(cmd, commandsMaster)
		if cmd == "" {
			// wrong command
			continue
		}

		fmt.Printf("CLI send %s to UDP server\n", cmd)

		switch cmd {
		case "help":
			fmt.Println(strings.Replace(getMasterHelp(),"\t","",-1))
		case "ls":
			fmt.Println(filename1)
		case "store":
			fmt.Println("TODO")
		case "exit":
			os.Exit(1)
		}
	}
}