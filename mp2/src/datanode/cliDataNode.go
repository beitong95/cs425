package datanode

import (
	"os"
	"github.com/marcusolsson/tui-go"
	"bufio"
	"fmt"
	"strings"
	"cli"
	"time"
)

var commandsDatanode = []string{"store", "exit", "help"}

func getDatanodeHelp() string {
	return `help                        -> help inFormation
			store                       -> list files stored on local disk
			exit                        -> exit from HDFS`
}


var (
	history *tui.Box
	input *tui.Entry
	shell *tui.Box
	ui tui.UI
)
// Cli command line function
func cliDatanode() {
	history, input, shell = cli.CreateShell()
	root := tui.NewVBox(shell)
	ui,_ = tui.New(root)
	done := make(chan string)
	// shell logic
	input.OnSubmit(func(e *tui.Entry) {
		cmd, _:= cli.ParseCmd(history, input,e.Text()[2:], commandsDatanode)
		if cmd == "" {
			// wrong command
			return
		}
		switch cmd {
		case "help":
			cli.Write2Shell(history,getDatanodeHelp())
		case "store":
			cli.Write2Shell(history,"TODO")
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
func cliSimpleDatanode() {
	reader = bufio.NewReader(os.Stdin)
	fmt.Println("Simple Shell")
	fmt.Println("---------------------")
	for {
		fmt.Print("-> ")
		cmd, _ := reader.ReadString('\n')
		cmd, _ = cli.ParseCmdSimple(cmd, commandsDatanode)
		if cmd == "" {
			// wrong command
			continue
		}

		fmt.Printf("CLI send %s to UDP server\n", cmd)

		switch cmd {
		case "help":
			fmt.Println(strings.Replace(getDatanodeHelp(),"\t","",-1))
		case "store":
			fmt.Println("TODO")
		case "exit":
			os.Exit(1)
		}
	}
}