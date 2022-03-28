package common

import (
	"encoding/gob"
	"net"
	"os"
)

type GobHandler struct {
	enc *gob.Encoder
	dec *gob.Decoder
}

type DirName string

type Schema interface {
	[]FileStruct | DirName
}

func GetIPv4Str() string {
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			return ipv4.String()
		}
	}
	return ""
}

func GetTcpAddrStr(PORT string) string {
	return GetIPv4Str() + ":" + PORT
}

func NewGobHandler(conn net.Conn) *GobHandler {
	return &GobHandler{gob.NewEncoder(conn), gob.NewDecoder(conn)}
}

func (h *GobHandler) Encode(i interface{}) error {
	return h.enc.Encode(i)
}

func Decode[T Schema](h *GobHandler) (T, error) {
	var data T
	return data, h.dec.Decode(&data)
}
