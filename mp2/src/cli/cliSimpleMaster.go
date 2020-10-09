package cli

import (
	"fmt"
	"os"
	"strings"
)

// Cli command line function
func CliSimpleMaster() {
	fmt.Println("Simple Shell")
	fmt.Println("---------------------")
	for {
		fmt.Print("-> ")
		cmd, _ := reader.ReadString('\n')
		cmd, _ = parseCmdSimple(cmd)
		if cmd == "" {
			// wrong command
			continue
		}

		fmt.Printf("CLI send %s to UDP server\n", cmd)

		switch cmd {
		case "help":
			fmt.Println(strings.Replace(getHelp(),"\t","",-1))
		case "ls":
			fmt.Println("TODO")
		case "store":
			fmt.Println("TODO")
		case "exit":
			os.Exit(1)
		}
	}
}
