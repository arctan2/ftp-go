package client

import (
	"net"
)

type dialer struct {
	addr string
	port string
	ipv4 string
}

func (dlr *dialer) DialAndCmd(cmd string) (net.Conn, error) {
	conn, err := net.Dial("tcp", dlr.addr)
	if err != nil {
		return nil, err
	}
	conn.Write([]byte(cmd + "\n"))
	return conn, err
}
