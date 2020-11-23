package datanode

import (
	"constant"
	"networking"
	"helper"
)

func ServerRun(serverPort string) {
	// create a folder
	helper.CreateFile()

	// client puts; datanode downloads the file to local disk
	// endpoint /putfile
	networking.HTTPlistenDownload(constant.Dir + "files_" + constant.DatanodeHTTPServerPort + "/") 

	// master sends rereplica request to source; source sends file to datanode; datanode downloads file to local disk
	// endpoint /rereplica
	networking.HTTPlistenRereplica() 

	// new master sends recover request to datanode; datanode sends its local file list to master.
	// endpoint /recover
	networking.HTTPlistenRecover() 

	// client sends delete request to datanode; datanode deletes the file.
	// endpoint /deletefile
	networking.HTTPlistenDelete(constant.Dir + "files_" + constant.DatanodeHTTPServerPort + "/")

	// start above http services
	go networking.HTTPstart(constant.DatanodeHTTPServerUploadPort) // start http server. main function: put, sub function: rereplica

	// client gets; datanode send the file to the client.
	go networking.HTTPfileServer(serverPort, constant.Dir + "files_" + constant.DatanodeHTTPServerPort) //handle get files
}
