package constant 
const ReconnectPeriod = 5000
const UDPportMaster2Client string = "1234"
const UDPportClient2Master string = "3456"
const UDPportDatanode2Master string = "3457"
const HTTPportClient2Master = "4321"
const HTTPClient2DataNodeDownload = "5000"
const MasterIP string = "172.22.156.12"
const Dir string = "/Users/chenxinhang/Files"

// inactive detector (master detect client)
const KickoutTimeout = 60000 
const CheckInactiveClientInterval = 20000 

// fail detector (client detect master)
const MasterTimeout = 4000
const ClientDetectMasterFailInterval = 2000
const MasterSendHeartbeat2ClientInterval = 1500

// fail detector (master detect datanode)
const DatanodeTimeout = 4000
const MasterDetectDatanodeFailInterval = 2000
const DatanodeSendHeartbeat2MasterInterval = 1500

var KickoutRejoinCmd chan string
var IsKickout bool

var LocalIP string