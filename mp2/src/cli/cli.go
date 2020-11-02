package cli

import (
	"github.com/marcusolsson/tui-go"
	"strings"
	"fmt"
	"time"
	. "structs"
	"helper"
	"log"
)

const UPDATESHELL = 500

func CreateShell() (*tui.Box, *tui.Entry, *tui.Box){
	history := tui.NewVBox()
	historyScroll := tui.NewScrollArea(history)
	historyScroll.SetAutoscrollToBottom(true)

	historyBox := tui.NewVBox(historyScroll)
	historyBox.SetBorder(true)

	input := tui.NewEntry()
	input.SetFocused(true)
	input.SetText(">>")
	input.SetSizePolicy(tui.Expanding, tui.Maximum)

	inputBox := tui.NewHBox(input)
	inputBox.SetBorder(true)
	inputBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	// combine History and input to get shell
	shell := tui.NewVBox(historyBox, inputBox)
	shell.SetSizePolicy(tui.Expanding, tui.Expanding)
	return history, input, shell
}

/** template
func CreateClientMasterStatusBox() (*tui.Box, *tui.Label) {
	clientMasterStatusLabel := tui.NewLabel("UNCONN")
	clientMasterStatusLabel.SetSizePolicy(tui.Expanding, tui.Expanding)
	clientMasterStatusBox := tui.NewVBox(clientMasterStatusLabel)
	clientMasterStatusBox.SetTitle("MasterStatus")
	clientMasterStatusBox.SetBorder(true)
	return clientMasterStatusBox, clientMasterStatusLabel
}

func Write2ClientMasterStatus(clientMasterStatusLabel *tui.Label ,text string) {
	if clientMasterStatusLabel == nil {
		return
	}
	clientMasterStatusLabel.SetText(text)
}
**/

func CreateBandwidthBox() (*tui.Box, *tui.Label) {
	bandwidthBoxLabel := tui.NewLabel("")
	bandwidthBoxLabel.SetSizePolicy(tui.Expanding, tui.Expanding)
	bandwidthBox := tui.NewVBox(bandwidthBoxLabel)
	bandwidthBox.SetTitle("BandWidth")
	bandwidthBox.SetBorder(true)
	return bandwidthBox, bandwidthBoxLabel
}

func Write2BandwidthBox(bandwidthBox *tui.Box, bandwidthBoxLabel *tui.Label ,text string) {
	if bandwidthBoxLabel == nil {
		return
	}
	bandwidthBoxLabel.SetText(text)
}

func CreateProtocolBox() (*tui.Box, *tui.Label) {
	protocolBoxLabel := tui.NewLabel("")
	protocolBoxLabel.SetSizePolicy(tui.Expanding, tui.Expanding)
	protocolBox := tui.NewVBox(protocolBoxLabel)
	protocolBox.SetTitle("Protocol")
	protocolBox.SetBorder(true)
	return protocolBox, protocolBoxLabel
}

func Write2ProtocolBox(protocolBox *tui.Box, protocolBoxLabel *tui.Label ,text string) {
	if protocolBoxLabel == nil {
		return
	}
	protocolBoxLabel.SetText(text)
}
func CreateCurrentStatusBox() (*tui.Box, *tui.Label) {
	currentStatusBoxLabel:= tui.NewLabel("")
	currentStatusBoxLabel.SetSizePolicy(tui.Expanding, tui.Expanding)
	currentStatusBox:= tui.NewVBox(currentStatusBoxLabel)
	currentStatusBox.SetTitle("Status")
	currentStatusBox.SetBorder(true)
	return currentStatusBox, currentStatusBoxLabel 
}
func Write2CurrentStatusBox(currentStatusBox *tui.Box, currentStatusBoxLabel *tui.Label ,text string) {
	if currentStatusBoxLabel== nil {
		return
	}
	currentStatusBoxLabel.SetText(text)
}

func CreateMembershipBox() (*tui.Box, *tui.Label) {
	membershipBoxLabel := tui.NewLabel("")
	membershipBoxLabel.SetSizePolicy(tui.Expanding, tui.Expanding)

	membershipBox := tui.NewVBox(membershipBoxLabel)
	membershipBox.SetTitle("MembershipList on " + MyID)
	membershipBox.SetBorder(true)
	return membershipBox, membershipBoxLabel
}

func Write2MembershipBox(membershipBox *tui.Box, membershipBoxLabel *tui.Label ,text string) {
	if membershipBox == nil || membershipBoxLabel == nil {
		return
	}
	membershipBoxLabel.SetText(text)
	membershipBox.SetTitle("MembershipList on " + MyID)
}

func AutoUpdateCLI(ui tui.UI) {
	for {
		ui.Update(func(){
			s, err := helper.PrintMembershipListAsTableInGUI(MembershipList)
			if err != nil {
				log.Fatal("PrintMembershipListAsTableInGUI error")
			}
			Write2MembershipBox(membershipBox, membershipBoxLabel, s)
			Write2BandwidthBox(bandwidthBox, bandwidthBoxLabel, fmt.Sprintf("%v",Bandwidth))
			Write2ProtocolBox(protocolBox, protocolBoxLabel, CurrentProtocol)
			Write2CurrentStatusBox(currentStatusBox, currentStatusBoxLabel, CurrentStatus)
		})
		time.Sleep(UPDATESHELL * time.Millisecond)
	}
}

func ParseCmd(input *tui.Entry, cmd string, commands []string) (string, string, string) {
	Write2Shell(cmd)
	cmds := strings.Fields(cmd)
	mainCmd := ""
	filename1 := ""
	filename2 := ""
	if len(cmds) == 1 {
		mainCmd = cmds[0]
	} else if len(cmds) == 2{
		// delete or ls
		mainCmd = cmds[0]
		filename1 = cmds[1]
	} else if len(cmds) == 3{
		mainCmd = cmds[0]
		filename1 = cmds[1]
		filename2 = cmds[2]
	} else {
		Write2Shell("bad command format")
		return "","",""
	}
	input.SetText(">>")
	wrongCommand := true
	for i := 0; i < len(commands); i++ {
		if commands[i] == strings.Fields(cmd)[0] {
			wrongCommand = false
		}
	}
	if wrongCommand == true {
		Write2Shell("wrong command")
		return "","",""
	}
	return mainCmd, filename1, filename2
}

func ParseCmdSimple(cmd string, commands []string) (string,string,string) {
		cmd = strings.Replace(cmd, "\r\n", "", -1)
		cmd = strings.Replace(cmd, "\n", "", -1)
		cmds := strings.Fields(cmd)
		mainCmd := ""
		filename1 := ""
		filename2 := ""
		if len(cmds) == 1 {
			mainCmd = cmds[0]
		} else if len(cmds) == 2{
			// delete or ls
			mainCmd = cmds[0]
			filename1 = cmds[1]
		} else if len(cmds) == 3{
			mainCmd = cmds[0]
			filename1 = cmds[1]
			filename2 = cmds[2]
		} else {
			fmt.Println("bad command format")
			return "","",""
		}
		wrongCommand := true
		for i := 0; i < len(commands); i++ {
			if commands[i] == cmd {
				wrongCommand = false
			}
		}
		if wrongCommand == true {
			fmt.Println("wrong command")
			return "","",""
		}
		return mainCmd, filename1, filename2

}