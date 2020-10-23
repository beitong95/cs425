package main

import (
	"helper"
	"log"
	"os"
	. "structs"
	"time"
	"github.com/marcusolsson/tui-go"
)

var testMap1 map[string]Membership = map[string]Membership{"11111": Membership{1111, 1111}, "11113": Membership{112321, 123123}}

type post struct {
	username string
	message  string
	time     string
}

func mergeMembership(ticker *time.Ticker) {
	for {
		<-ticker.C
		MT.Lock()
		testMap1["11111"] = Membership{time.Now().Unix(), time.Now().Unix()}
		MT.Unlock()
		UpdateGUI <- "Ping"
	}
}

func updateMembershipListInGUI(membershipBoxLabel *tui.Label, ui tui.UI) {
	for {
		<-UpdateGUI
		s, err := helper.PrintMembershipListAsTableInGUI(testMap1)
		if err != nil {
			log.Fatal("PrintMembershipListAsTableInGUI error")
		}
		ui.Update(func() {
			membershipBoxLabel.SetText(s)
		})
	}
}

func getHelp() string {
	return `help    -> help inFormation
			all2all -> change multicast to all2all
			gossip  -> change multicast to gossip
			leave   -> leave current group
			join    -> join current group
			id      -> print current id
			list    -> print current membershipList  
			kill    -> fail myself`
}

func main() {
	// shell history
	done := make(chan string)
	var ui tui.UI
	commands := []string{"help", "all2all", "gossip", "leave", "join", "id", "list", "kill"}
	f, err := os.OpenFile("text.log", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	logger := log.New(f, "test", log.LstdFlags)
	tui.SetLogger(logger)
	history := tui.NewVBox()
	// initialize history (change to shell help)
	history.Append(tui.NewHBox(
		tui.NewLabel(time.Now().Format("15:04")),
		tui.NewPadder(1, 0, tui.NewLabel("")),
		tui.NewLabel(getHelp()),
		tui.NewSpacer(),
	))

	historyScroll := tui.NewScrollArea(history)
	historyScroll.SetAutoscrollToBottom(true)

	historyBox := tui.NewVBox(historyScroll)
	historyBox.SetBorder(true)

	// shell input
	// NewEntry is a oneline input
	input := tui.NewEntry()
	input.SetFocused(true)
	input.SetText(">>")
	input.SetSizePolicy(tui.Expanding, tui.Maximum)

	inputBox := tui.NewHBox(input)
	inputBox.SetBorder(true)
	inputBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	// combine history and input to get shell
	shell := tui.NewVBox(historyBox, inputBox)
	shell.SetSizePolicy(tui.Expanding, tui.Expanding)

	// shell logic
	input.OnSubmit(func(e *tui.Entry) {

		cmd := e.Text()[2:]
		history.Append(tui.NewHBox(
			tui.NewLabel(time.Now().Format("15:04")),
			tui.NewPadder(1, 0, tui.NewLabel("")),
			tui.NewLabel(cmd),
			tui.NewSpacer(),
		))
		input.SetText(">>")
		wrongCommand := true
		for i := 0; i < len(commands); i++ {
			if commands[i] == cmd {
				wrongCommand = false
			}
		}

		if cmd == "help" || wrongCommand == true {
			history.Append(tui.NewHBox(
				tui.NewLabel(time.Now().Format("15:04")),
				tui.NewPadder(1, 0, tui.NewLabel("")),
				tui.NewLabel(getHelp()),
				tui.NewSpacer(),
			))
		} else {
			switch cmd {
			case "all2all":
				history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().Format("15:04")),
					tui.NewPadder(1, 0, tui.NewLabel("")),
					tui.NewLabel("change system to all to all mode"),
					tui.NewSpacer(),
				))
				//c <- CHANGE_TO_ALL2ALL
			case "gossip":
				history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().Format("15:04")),
					tui.NewPadder(1, 0, tui.NewLabel("")),
					tui.NewLabel("change system to gossip mode"),
					tui.NewSpacer(),
				))
				//c <- CHANGE_TO_GOSSIP
			case "leave":
				history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().Format("15:04")),
					tui.NewPadder(1, 0, tui.NewLabel("")),
					tui.NewLabel("leave group"),
					tui.NewSpacer(),
				))
				//c <- LEAVE_GROUP
			case "join":
				history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().Format("15:04")),
					tui.NewPadder(1, 0, tui.NewLabel("")),
					tui.NewLabel("join group"),
					tui.NewSpacer(),
				))
				//c <- JOIN_GROUP
			case "id":
				history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().Format("15:04")),
					tui.NewPadder(1, 0, tui.NewLabel("")),
					tui.NewLabel("ID: "+MyID),
					tui.NewSpacer(),
				))
			case "list":
				s, err := helper.PrintMembershipListAsTableInGUI(testMap1)
				if err != nil {
					log.Fatal("PrintMembershipListAsTableInGUI error")
				}
				logger.Printf(s)
				tmp := tui.NewLabel("")
				tmp.SetText(s)
				history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().Format("15:04")),
					tui.NewPadder(1, 0, tui.NewLabel("")),
					tmp,
					tui.NewSpacer(),
				))
			case "kill":
				history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().Format("15:04")),
					tui.NewPadder(1, 0, tui.NewLabel("")),
					tui.NewLabel("Got killed"),
					tui.NewSpacer(),
				))
				ui.Quit()
				done <- "done"
				os.Exit(1)
			}

		}
	})

	// membership list
	//to do update membership
	MyIP := "1:1:1:1"
	MyID := "1234"
	membershipBoxLabel := tui.NewLabel("")
	membershipBoxLabel.SetSizePolicy(tui.Expanding, tui.Expanding)

	membershipBox := tui.NewVBox(membershipBoxLabel)
	membershipBox.SetTitle("MembershipList on " + MyIP + ":" + MyID)
	membershipBox.SetBorder(true)
	s, err := helper.PrintMembershipListAsTableInGUI(testMap1)
	if err != nil {
		log.Fatal("PrintMembershipListAsTableInGUI error")
	}
	membershipBoxLabel.SetText(s)

	root := tui.NewVBox(membershipBox, shell)
	var er error
	ui, er = tui.New(root)
	if er != nil {
		log.Fatal(err)
	}

	ui.SetKeybinding("Esc", func() {
		ui.Quit()
		done <- "Done"
		os.Exit(1)
	})
	go ui.Run()
	ticker := time.NewTicker(time.Duration(1000) * time.Millisecond)
	go mergeMembership(ticker)
	go updateMembershipListInGUI(membershipBoxLabel, ui)

	<-done
}

/**
// Cli command line function
func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Simple Shell")
	fmt.Println("---------------------")
	commands := []string{"help", "all2all", "gossip", "leave", "join", "id", "list", "kill"}
	for {
		fmt.Print("-> ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.Replace(cmd, "\r\n", "", -1)
		cmd = strings.Replace(cmd, "\n", "", -1)

		wrongCommand := true
		for i := 0; i < len(commands); i++ {
			if commands[i] == cmd {
				wrongCommand = false
			}
		}

		fmt.Printf("CLI send %s to UDP server\n", cmd)

		if cmd == "help" || wrongCommand == true {
			fmt.Println("help    -> help inFormation")
			fmt.Println("all2all -> change multicast to all2all")
			fmt.Println("gossip  -> change multicast to gossip")
			fmt.Println("leave   -> leave current group")
			fmt.Println("join    -> join current group")
			fmt.Println("id      -> print current id")
			fmt.Println("list    -> print current membershipList")
			fmt.Println("kill    -> fail myself")
			continue
		}

		switch cmd {
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
			service.MT.Lock()
			helper.PrintMembershipListAsTable(MembershipList)
			service.MT.Unlock()
		case "kill":
			os.Exit(1)
		}
	}
}

**/
