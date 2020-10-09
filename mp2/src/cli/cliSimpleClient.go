package cli

import (
	"fmt"
	"os"
	"strings"
)

// Cli command line function
func CliSimpleClient() {
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
