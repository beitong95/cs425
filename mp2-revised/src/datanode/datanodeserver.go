package datanode

import (
	"constant"
	"networking"
	. "structs"
)

func ServerRun(serverPort string) {
	CreateFile()
	networking.HTTPlistenDownload(constant.Dir + "files_" + MyPort)
	go networking.HTTPstart(constant.DatanodeHTTPServerUploadPort)
	go networking.HTTPfileServer(serverPort, constant.Dir) //handle get files

}
