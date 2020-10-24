package datanode

import (
	"constant"
	"networking"
)

func ServerRun(serverPort string) {
	CreateFile()
	networking.HTTPlistenDownload(constant.Dir + "files_" + constant.DatanodeHTTPServerPort)
	go networking.HTTPstart(constant.DatanodeHTTPServerUploadPort)
	go networking.HTTPfileServer(serverPort, constant.Dir) //handle get files

}
