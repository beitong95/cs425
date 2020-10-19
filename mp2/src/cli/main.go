package cli
import (
	"bufio"
	"logger"
	"github.com/marcusolsson/tui-go"
	"os"
)
var _identity string = ""
var history *tui.Box
var input *tui.Entry
var shell *tui.Box
var reader *bufio.Reader
var ui tui.UI
var masterMembershipBox *tui.Box
var masterClientMembershipLabel *tui.Label
var masterDatanodeMembershipLabel *tui.Label
var clientMasterStatusBox *tui.Box
var clientMasterStatusLabel *tui.Label
func Run(cliLevel string, identity string){
	_identity = identity
	switch cliLevel {
	case "cli":
		switch _identity {
		case "client":
			CliClient()
		case "master":
			CliMaster()
		case "dataNode":
			CliDataNode()
		}
	case "cliSimple":
		reader = bufio.NewReader(os.Stdin)
		switch _identity {
		case "client":
			CliSimpleClient()
		case "master":
			CliSimpleMaster()
		case "dataNode":
			CliSimpleDataNode()
		}
	default:
		logger.LogSimpleFatal("no cli")
	}
}