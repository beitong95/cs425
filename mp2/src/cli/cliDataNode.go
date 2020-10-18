package cli

import (
	"os"
	"github.com/marcusolsson/tui-go"
	"time"
)

// Cli command line function
func CliDataNode() {
	done := make(chan string)
	//create shell
	createShell()
	// shell logic
	input.OnSubmit(func(e *tui.Entry) {
		cmd, _:= parseCmd(e.Text()[2:])
		if cmd == "" {
			// wrong command
			return
		}
		switch cmd {
		case "help":
			Write2Shell(getHelp())
		case "store":
			Write2Shell("TODO")
		case "exit":
			time.Sleep(time.Duration(500) * time.Millisecond)
			ui.Quit()
			done <- "Done"
			os.Exit(1)
		}
	})

	root := tui.NewVBox(shell)
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
	go autoUpdateShell()
	<-done
}
