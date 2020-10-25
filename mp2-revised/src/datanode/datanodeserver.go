package datanode

import (
	"constant"
	"networking"
)

func ServerRun(serverPort string) {
	CreateFile()
	// register put server on port: DatanodeHTTPServerUploadPort
	networking.HTTPlistenDownload(constant.Dir + "files_" + constant.DatanodeHTTPServerPort + "/") 
	networking.HTTPlistenRereplica() // register rereplica server on port: DatanodeHTTPServerUploadPort
	go networking.HTTPstart(constant.DatanodeHTTPServerUploadPort) // start http server. main function: put, sub function: rereplica
	go networking.HTTPfileServer(serverPort, constant.Dir + "files_" + constant.DatanodeHTTPServerPort) //handle get files

}
