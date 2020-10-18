package networking
import (
	"net"
)
func UDPsend(ip string, port string, message []byte ) error {
	addr, err := net.ResolveUDPAddr("udp", ip+":"+port)
		if err != nil {
			return err
		}
		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			return err
		}
		defer conn.Close()
		_, err = conn.Write(message)
		if err != nil {
			return err
		}
	return nil
}

func UDPlisten(port string, callback func(message []byte))error{
	port = ":" + port
	addr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	defer conn.Close()
	buffer := make([]byte, 4096)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return err
		}
		callback(buffer[0:n])
	}
}

func TCPsend(){

}

func TCPlisten(){
	
}