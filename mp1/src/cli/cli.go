package cli

import (
	"helper"
	"log"
	"os"
	. "structs"
	"sync"
	"time"
	"github.com/marcusolsson/tui-go"
	"fmt"
)

func updateMembershipListInGUI(membershipBoxLabel *tui.Label, membershipBox *tui.Box, ui tui.UI, ticker2 *time.Ticker) {
	for {
		<-ticker2.C
		s, err := helper.PrintMembershipListAsTableInGUI(MembershipList)
		if err != nil {
			log.Fatal("PrintMembershipListAsTableInGUI error")
		}
		ui.Update(func() {
			membershipBoxLabel.SetText(s)
			membershipBox.SetTitle("MembershipList on " + MyID)
		})
	}
}

func updateBandwidthAndJoin(bandwidthBoxLabel *tui.Label, joinBoxLabel *tui.Label, ui tui.UI, ticker *time.Ticker) {
	for {
		<-ticker.C
		ui.Update(func() {
			bandwidthBoxLabel.SetText(fmt.Sprintf("%v",Bandwidth))
			joinBoxLabel.SetText(fmt.Sprintf("%v",IsJoin))
		})
	}
}

func updateProtocolChangeACK(history *tui.Box, protocolBoxLabel *tui.Label, ui tui.UI) {
	for {
		msg := <-ProtocolChangeACK
		ui.Update(func() {
			protocolBoxLabel.SetText(msg)
            history.Append(tui.NewHBox(
                tui.NewLabel(time.Now().Format("15:04")),
                tui.NewPadder(1, 0, tui.NewLabel("")),
                tui.NewLabel("Change to " + msg),
                tui.NewSpacer(),
            ))
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
			para    -> print all gossip and all2all parameter
			kill    -> fail myself`
}

// Cli command line function
func Cli(wg *sync.WaitGroup, c chan int) {
	defer wg.Done()
	done := make(chan string)
	var ui tui.UI
	commands := []string{"help", "all2all", "gossip", "leave", "join", "id", "list", "para","kill"}

	//set up gui
	// set shell history
	history := tui.NewVBox()
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
				c <- CHANGE_TO_ALL2ALL
			case "gossip":
				history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().Format("15:04")),
					tui.NewPadder(1, 0, tui.NewLabel("")),
					tui.NewLabel("change system to gossip mode"),
					tui.NewSpacer(),
				))
				c <- CHANGE_TO_GOSSIP
			case "leave":
				history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().Format("15:04")),
					tui.NewPadder(1, 0, tui.NewLabel("")),
					tui.NewLabel("leave group"),
					tui.NewSpacer(),
				))
				c <- LEAVE_GROUP
			case "join":
				history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().Format("15:04")),
					tui.NewPadder(1, 0, tui.NewLabel("")),
					tui.NewLabel("join group"),
					tui.NewSpacer(),
				))
				c <- JOIN_GROUP
			case "id":
				history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().Format("15:04")),
					tui.NewPadder(1, 0, tui.NewLabel("")),
					tui.NewLabel("ID: "+MyID),
					tui.NewSpacer(),
				))
			case "list":
				s, err := helper.PrintMembershipListAsTableInGUI(MembershipList)
				if err != nil {
					log.Fatal("PrintMembershipListAsTableInGUI error")
				}
				history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().Format("15:04")),
					tui.NewPadder(1, 0, tui.NewLabel("")),
					tui.NewLabel(s),
					tui.NewSpacer(),
				))
			case "para":
				history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().Format("15:04")),
					tui.NewPadder(1, 0, tui.NewLabel("")),
					tui.NewLabel("Tgossip: "+ fmt.Sprintf("%v",Tgossip)),
					tui.NewSpacer(),
				))
				history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().Format("15:04")),
					tui.NewPadder(1, 0, tui.NewLabel("")),
					tui.NewLabel("Tall2all: "+ fmt.Sprintf("%v",Tall2all)),
					tui.NewSpacer(),
				))
				history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().Format("15:04")),
					tui.NewPadder(1, 0, tui.NewLabel("")),
					tui.NewLabel("Tfail: "+ fmt.Sprintf("%v",Tfail)),
					tui.NewSpacer(),
				))
				history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().Format("15:04")),
					tui.NewPadder(1, 0, tui.NewLabel("")),
					tui.NewLabel("Tclean: "+ fmt.Sprintf("%v",Tclean)),
					tui.NewSpacer(),
				))
				history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().Format("15:04")),
					tui.NewPadder(1, 0, tui.NewLabel("")),
					tui.NewLabel("B: "+ fmt.Sprintf("%v",B)),
					tui.NewSpacer(),
				))
			case "kill":
				history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().Format("15:04")),
					tui.NewPadder(1, 0, tui.NewLabel("")),
					tui.NewLabel("Got killed"),
					tui.NewSpacer(),
				))
				time.Sleep(time.Duration(500) * time.Millisecond)
				ui.Quit()
				done <- "Done"
				os.Exit(1)
			}

		}
	})

	// membership list
	membershipBoxLabel := tui.NewLabel("")
	membershipBoxLabel.SetSizePolicy(tui.Expanding, tui.Expanding)

	membershipBox := tui.NewVBox(membershipBoxLabel)
	membershipBox.SetTitle("MembershipList on " + MyID)
	membershipBox.SetBorder(true)
	s, err := helper.PrintMembershipListAsTableInGUI(MembershipList)
	if err != nil {
		log.Fatal("PrintMembershipListAsTableInGUI error")
	}
	membershipBoxLabel.SetText(s)

	//  bandwidth and protocol
	bandwidthBoxLabel := tui.NewLabel("")
	bandwidthBoxLabel.SetSizePolicy(tui.Expanding, tui.Expanding)

	bandwidthBox := tui.NewVBox(bandwidthBoxLabel)
	bandwidthBox.SetTitle("BandWidth")
	bandwidthBox.SetBorder(true)
	bandwidthBoxLabel.SetText(fmt.Sprintf("%v",Bandwidth))

	protocolBoxLabel := tui.NewLabel("")
	protocolBoxLabel.SetSizePolicy(tui.Expanding, tui.Expanding)

	protocolBox := tui.NewVBox(protocolBoxLabel)
	protocolBox.SetTitle("Protocol")
	protocolBox.SetBorder(true)
	protocolBoxLabel.SetText(CurrentProtocol)

	joinBoxLabel := tui.NewLabel("")
	joinBoxLabel.SetSizePolicy(tui.Expanding, tui.Expanding)

	joinBox:= tui.NewVBox(joinBoxLabel)
	joinBox.SetTitle("Join")
	joinBox.SetBorder(true)
	joinBoxLabel.SetText(fmt.Sprintf("%v",IsJoin))

	bandpandjBox := tui.NewHBox(bandwidthBox, protocolBox, joinBox)

	root := tui.NewVBox(membershipBox, bandpandjBox, shell)

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
	tickerMembershipList := time.NewTicker(time.Duration(Tgossip) * time.Millisecond)
	go updateMembershipListInGUI(membershipBoxLabel, membershipBox, ui, tickerMembershipList)
    go updateProtocolChangeACK(history, protocolBoxLabel, ui)
	ticker := time.NewTicker(time.Duration(1000) * time.Millisecond)
	go updateBandwidthAndJoin(bandwidthBoxLabel, joinBoxLabel, ui, ticker)
	<-done
}
