package cli

import (
	"github.com/marcusolsson/tui-go"
	"strings"
	"fmt"
	"time"
)

func getHelp() string {
	switch _identity {
	case "client":
		return `help                        -> help inFormation
				get filename                -> read file from HDFS
				set filename (newfilename)  -> write file to HDFS
				delete filename             -> delete file in HDFS	
				store                       -> list files stored on local disk
				exit                        -> exit from HDFS`
	case "master":
		return `help                        -> help inFormation
				ls filename                 -> list where is this file stores in HDFS
				store                       -> list files stored on local disk
				exit                        -> exit from HDFS`
	case "dataNode":
		return `help                        -> help inFormation
				store                       -> list files stored on local disk
				exit                        -> exit from HDFS`
	default:
		return "unknown identity"
	}
}

var commandsClient = []string{"get", "set", "delete", "store", "exit", "help"}
var commandsMaster= []string{"ls", "store", "exit", "help"}
var commandsDataNode = []string{"store", "exit", "help"}

func createShell() {
	history = tui.NewVBox()
	historyScroll := tui.NewScrollArea(history)
	historyScroll.SetAutoscrollToBottom(true)

	historyBox := tui.NewVBox(historyScroll)
	historyBox.SetBorder(true)

	input = tui.NewEntry()
	input.SetFocused(true)
	input.SetText(">>")
	input.SetSizePolicy(tui.Expanding, tui.Maximum)

	inputBox := tui.NewHBox(input)
	inputBox.SetBorder(true)
	inputBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	// combine history and input to get shell
	shell = tui.NewVBox(historyBox, inputBox)
	shell.SetSizePolicy(tui.Expanding, tui.Expanding)
}

func write2Shell(text string) {
	history.Append(tui.NewHBox(
		tui.NewLabel(time.Now().Format("15:04")),
		tui.NewPadder(1, 0, tui.NewLabel("")),
		tui.NewLabel(text),
		tui.NewSpacer(),
	))
}

func parseCmd(cmd string) (string, string) {
	write2Shell(cmd)
	cmds := strings.Fields(cmd)
	mainCmd := ""
	subCmd := ""
	if len(cmds) == 1 {
		mainCmd = cmds[0]
		subCmd = ""
	} else if len(cmds) == 2{
		mainCmd = cmds[0]
		subCmd = cmds[1]
	} else {
		write2Shell("bad command format")
		write2Shell(getHelp())
		return "",""
	}
	input.SetText(">>")
	wrongCommand := true
	commands := make([]string, 0)
	switch _identity {
	case "client":
		commands = commandsClient
	case "master":
		commands = commandsMaster
	case "dataNode":
		commands = commandsDataNode
	}
	for i := 0; i < len(commands); i++ {
		if commands[i] == strings.Fields(cmd)[0] {
			wrongCommand = false
		}
	}
	if wrongCommand == true {
		return "",""
	}
	return mainCmd, subCmd
}

func parseCmdSimple(cmd string) (string,string) {
		cmd = strings.Replace(cmd, "\r\n", "", -1)
		cmd = strings.Replace(cmd, "\n", "", -1)
		cmds := strings.Fields(cmd)
		mainCmd := ""
		subCmd := ""
		if len(cmds) == 1 {
			mainCmd = cmds[0]
			subCmd = ""
		} else if len(cmds) == 2{
			mainCmd = cmds[0]
			subCmd = cmds[1]
		} else {
			fmt.Println("bad command format")
			fmt.Println(strings.Replace(getHelp(),"\t","",-1))
			return "",""
		}
		wrongCommand := true
		commands := make([]string, 0)
		switch _identity {
		case "client":
			commands = commandsClient
		case "master":
			commands = commandsMaster
		case "dataNode":
			commands = commandsDataNode
		}
		for i := 0; i < len(commands); i++ {
			if commands[i] == cmd {
				wrongCommand = false
			}
		}
		if wrongCommand == true {
			return "",""
		}
		return mainCmd, subCmd

}