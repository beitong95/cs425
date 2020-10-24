package datanode

import (
	"constant"
	"fmt"
	"networking"
	"strconv"
	. "structs"
)

func ServerRun(serverPort string) {
	CreateFile()
	i, err := strconv.Atoi(MyPort)
	if err != nil {
		panic(err)
	}
	serverUploadPort := fmt.Sprint(int(i) + 2)
	networking.HTTPlistenDownload(constant.Dir)
	go networking.HTTPstart(serverUploadPort)
	go networking.HTTPfileServer(serverPort, constant.Dir) //handle get files

}
