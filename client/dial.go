package client

import (
	"log"
	"net"
)

type dialer struct {
	addr string
	port string
	ipv4 string
}

func newDialer(addr string) dialer {
	h, p, err := net.SplitHostPort(addr)
	if err != nil {
		log.Fatal(err.Error())
	}
	return dialer{addr: addr, ipv4: h, port: p}
}

func (dlr *dialer) DialAndCmd(cmd string) (net.Conn, error) {
	conn, err := net.Dial("tcp", dlr.addr)
	if err != nil {
		return nil, err
	}
	conn.Write([]byte(cmd + "\n"))
	return conn, err
}
