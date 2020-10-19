package cli

import (
	"github.com/marcusolsson/tui-go"
	"strings"
	"fmt"
	"time"
	"sort"
	"sync"
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
const UPDATESHELL = 500
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

func Write2Shell(text string) {
	history.Append(tui.NewHBox(
		tui.NewLabel(time.Now().Format("15:04")),
		tui.NewLabel(" "),
		tui.NewLabel(text),
		tui.NewSpacer(),
	))
}

func createClientMasterStatusBox() {
	clientMasterStatusLabel = tui.NewLabel("UNCONN")
	clientMasterStatusLabel.SetSizePolicy(tui.Expanding, tui.Expanding)
	clientMasterStatusBox = tui.NewVBox(clientMasterStatusLabel)
	clientMasterStatusBox.SetTitle("MasterStatus")
	clientMasterStatusBox.SetBorder(true)
}

func Write2ClientMasterStatus(text string) {
	clientMasterStatusLabel.SetText(text)
}

func createMasterMembershipBox() {
	masterClientMembershipLabel = tui.NewLabel("")
	masterDatanodeMembershipLabel = tui.NewLabel("")
	masterClientMembershipLabel.SetSizePolicy(tui.Expanding, tui.Expanding)
	masterDatanodeMembershipLabel.SetSizePolicy(tui.Expanding, tui.Expanding)

	masterClientMembershipBox := tui.NewVBox(masterClientMembershipLabel)
	masterDatanodeMembershipBox := tui.NewVBox(masterDatanodeMembershipLabel)
	masterClientMembershipBox.SetTitle("client membershiplist")
	masterClientMembershipBox.SetBorder(true)
	masterDatanodeMembershipBox.SetTitle("datanode membershiplist")
	masterDatanodeMembershipBox.SetBorder(true)

	masterMembershipBox = tui.NewHBox(masterClientMembershipBox, masterDatanodeMembershipBox)
}

func Write2MasterClientMembershipBox(text string) {
	masterClientMembershipLabel.SetText(text)
}

func Write2MasterDatanodeMembershipBox(text string) {
	masterDatanodeMembershipLabel.SetText(text)
}
func ConvertMasterClientMembershipList2String(membershipList map[string] int64, muxClientMembershipList sync.Mutex) string {
	var res []string
	if membershipList == nil {
		return ""
	}

	membershipAttributeCount := 1
	tableWidth := membershipAttributeCount + 1
	muxClientMembershipList.Lock()
	tableHeight := len(membershipList)
	maxL := make([]int, tableWidth)

	//get table header info
	s1 := make([]interface{}, tableWidth)
	keyName := "ID"
	s1[0] = keyName
	maxL[0] = len(keyName)
	attrName := "last active time" 
	s1[1] =  attrName
	maxL[1] = len(attrName)

	// get table body info
	s3 := make([][]interface{}, tableHeight)
	i := 0
	for k, v := range membershipList {
		s3[i] = make([]interface{}, tableWidth)
		s3[i][0] = k
		if l := len(k); l > maxL[0] {
			maxL[0] = l
		}
		s3[i][1] = v
		str := fmt.Sprintln(s3[i][1])
		if l := len(str); l > maxL[1] {
			maxL[1] = l
		}
		i++
	}
	muxClientMembershipList.Unlock()
	// sort 2d interface{} type slice in column 0
	sort.SliceStable(s3, func(i, j int) bool {
		return s3[i][0].(string) < s3[j][0].(string)
	})

	// get table border info
	tableWidthByCharacter := 0
	pad := 3
	for i := 0; i < tableWidth; i++ {
		tableWidthByCharacter += maxL[i]
	}
	tableWidthByCharacter += pad*tableWidth + (1*tableWidth + 1)

	// create print format command
	printCommand := ""
	for i := 0; i < tableWidth; i++ {
		printCommand = printCommand + "|%-" + fmt.Sprintf("%v", maxL[i]+3) + "v"
	}
	printCommand = printCommand + "|"
	//fmt.Printf("%#v\n", printCommand)

	// print border
	border := strings.Repeat("-", tableWidthByCharacter)
	res = append(res, border)

	// print header
	s := fmt.Sprintf(printCommand, s1...)
	res = append(res, s)
	s2 := make([]interface{}, tableWidth)
	for i := 0; i < tableWidth; i++ {
		s2[i] = strings.Repeat("-", maxL[i]+3)
	}
	s = fmt.Sprintf(printCommand, s2...)
	res = append(res, s)

	// print body
	for i := 0; i < tableHeight; i++ {
		s := fmt.Sprintf(printCommand, s3[i]...)
		res = append(res, s)
	}

	// print border
	res = append(res, border)
	return strings.Join(res, "\n")
	
	
}

func ConvertMasterDatanodeMembershipList2String(membershipList map[string] int64, muxDatanodeMembershipList sync.Mutex) string {
	var res []string
	if membershipList == nil {
		return ""
	}

	membershipAttributeCount := 1
	tableWidth := membershipAttributeCount + 1
	muxDatanodeMembershipList.Lock()
	tableHeight := len(membershipList)
	maxL := make([]int, tableWidth)

	//get table header info
	s1 := make([]interface{}, tableWidth)
	keyName := "ID"
	s1[0] = keyName
	maxL[0] = len(keyName)
	attrName := "Heartbeat" 
	s1[1] =  attrName
	maxL[1] = len(attrName)

	// get table body info
	s3 := make([][]interface{}, tableHeight)
	i := 0
	for k, v := range membershipList {
		s3[i] = make([]interface{}, tableWidth)
		s3[i][0] = k
		if l := len(k); l > maxL[0] {
			maxL[0] = l
		}
		s3[i][1] = v
		str := fmt.Sprintln(s3[i][1])
		if l := len(str); l > maxL[1] {
			maxL[1] = l
		}
		i++
	}
	muxDatanodeMembershipList.Unlock()
	// sort 2d interface{} type slice in column 0
	sort.SliceStable(s3, func(i, j int) bool {
		return s3[i][0].(string) < s3[j][0].(string)
	})

	// get table border info
	tableWidthByCharacter := 0
	pad := 3
	for i := 0; i < tableWidth; i++ {
		tableWidthByCharacter += maxL[i]
	}
	tableWidthByCharacter += pad*tableWidth + (1*tableWidth + 1)

	// create print format command
	printCommand := ""
	for i := 0; i < tableWidth; i++ {
		printCommand = printCommand + "|%-" + fmt.Sprintf("%v", maxL[i]+3) + "v"
	}
	printCommand = printCommand + "|"
	//fmt.Printf("%#v\n", printCommand)

	// print border
	border := strings.Repeat("-", tableWidthByCharacter)
	res = append(res, border)

	// print header
	s := fmt.Sprintf(printCommand, s1...)
	res = append(res, s)
	s2 := make([]interface{}, tableWidth)
	for i := 0; i < tableWidth; i++ {
		s2[i] = strings.Repeat("-", maxL[i]+3)
	}
	s = fmt.Sprintf(printCommand, s2...)
	res = append(res, s)

	// print body
	for i := 0; i < tableHeight; i++ {
		s := fmt.Sprintf(printCommand, s3[i]...)
		res = append(res, s)
	}

	// print border
	res = append(res, border)
	return strings.Join(res, "\n")
	
	
}


func autoUpdateCLI() {
	for {
		ui.Update(func(){

		})
		time.Sleep(UPDATESHELL * time.Millisecond)
	}
}

func parseCmd(cmd string) (string, string) {
	Write2Shell(cmd)
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
		Write2Shell("bad command format")
		Write2Shell(getHelp())
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