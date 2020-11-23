package constant

var Dir string = "/home/beitong2/cs425/mp3/files/"
// we will change those three ports adaptively in main.go (based on my port)
// here we just wanna to how do we convert myport to other ports.
var MasterHTTPServerPort = "1238" // local myport + 3 
var DatanodeHTTPServerPort = "1239" // local myport + 1
var DatanodeHTTPServerUploadPort = "1240" //local myport + 2 
const MasterGetTimeout = 300 // exit3 timeout 5 mins 
const MasterPutTimeout = 300 // exit3 timeout 5 mins 