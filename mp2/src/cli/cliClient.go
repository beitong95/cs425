package cli

import (
	"os"
	"github.com/marcusolsson/tui-go"
	"time"
	"constant"
)

// Cli command line function
func CliClient() {
	done := make(chan string)
	//create shell
	
	createShell()
	createClientMasterStatusBox()
	// shell logic
	input.OnSubmit(func(e *tui.Entry) {
		// rejoin cmd
		_cmd := e.Text()[2:]
		if constant.IsKickout == true {
			rejoinCmd := _cmd
			if rejoinCmd == "Y" {
				Write2Shell("Y")
				constant.KickoutRejoinCmd <- "true"
			} else {
				Write2Shell("N")
				constant.KickoutRejoinCmd <- "false"
				ui.Quit()
				done <- "Done"
				os.Exit(1)
			}
		} else {
			cmd, _:= parseCmd(_cmd)
			if cmd == "" {
				// wrong command
				return
			}
			switch cmd {
			case "help":
				Write2Shell(getHelp())
			case "get":
				Write2Shell("TODO")
			case "set":
				Write2Shell("TODO")
			case "delete":
				Write2Shell("TODO")
			case "store":
				Write2Shell("TODO")
			case "exit":
				time.Sleep(time.Duration(500) * time.Millisecond)
				ui.Quit()
				done <- "Done"
				os.Exit(1)
			}
		} 
	})

	root := tui.NewVBox(clientMasterStatusBox, shell)
	var er error
	ui, er = tui.New(root)
	if er != nil {
	}

	ui.SetKeybinding("Esc", func() {
		ui.Quit()
		done <- "Done"
		os.Exit(1)
	})
	go ui.Run()
	go autoUpdateCLI()
	<-done
}
