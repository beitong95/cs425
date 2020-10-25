package datanode

import (
	"constant"
	"networking"
)

func ServerRun(serverPort string) {
	CreateFile()
	networking.HTTPlistenDownload(constant.Dir + "files_" + constant.DatanodeHTTPServerPort + "/") // handle put
	go networking.HTTPstart(constant.DatanodeHTTPServerUploadPort) // handle put
	go networking.HTTPfileServer(serverPort, constant.Dir + "files_" + constant.DatanodeHTTPServerPort) //handle get files

}
