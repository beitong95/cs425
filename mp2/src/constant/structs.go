package constant

type UDPMessageClient2Master struct {
	IP string
	MessageType string // first connect
}

type UDPMessageMaster2Client struct {
	Heartbeat int64
	MessageType string // ls ack heartbeat kickout
}

type UDPMessageDatanode2Master struct {
	ID string
	Heartbeat int64
	MessageType string // heartbeat 
}
/** TODO:
type Master2ClientMessageTCP struct {
	IPs []string
	ACK string
	MessageType string
}

type Master2DatanodeTCP struct {
	RereplicaFile string
	RereplicaIP string
	MessageType string
}
**/