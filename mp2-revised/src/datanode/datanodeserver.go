package datanode

import (
	"constant"
	"networking"
)

func ServerRun(serverPort string) {
	CreateFile()
	networking.HTTPlistenDownload(constant.Dir + "files_" + constant.DatanodeHTTPServerPort + "/") // handle put
	networking.HTTPlistenDelete(constant.Dir + "files_" + constant.DatanodeHTTPServerPort + "/")
	go networking.HTTPstart(constant.DatanodeHTTPServerUploadPort)                                  // handle put
	go networking.HTTPfileServer(serverPort, constant.Dir+"files_"+constant.DatanodeHTTPServerPort) //handle get files

}
