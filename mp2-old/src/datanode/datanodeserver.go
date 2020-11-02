package datanode
import (
	"networking"
	"constant"
)
func ServerRun(otherport string){
	CreateFile()
	networking.HTTPfileServer(constant.HTTPClient2DataNodeDownload, constant.Dir)//handle get files
	networking.HTTPlistenDownload(constant.Dir)//handle upload files
	networking.HTTPstart(otherport)
}