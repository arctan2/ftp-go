package common

import (
	"encoding/gob"
	"io"
	"net"
	"fmt"
)

type GobHandler struct {
	enc *gob.Encoder
	dec *gob.Decoder
}

type Schema interface {
	[]FileStruct | FileStruct | DirName | string | bool | Res | ZipProgress
}

var (
	LOCAL_HOST = "127.0.0.1"
)

func (r Res) Error() string {
	return r.Data
}

func GetIPv4Str() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err.Error())
		return LOCAL_HOST
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.IsGlobalUnicast() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return LOCAL_HOST
}

func GetTcpAddrStr(PORT string) string {
	return GetIPv4Str() + ":" + PORT
}

func NewGobHandler(r io.Reader, w io.Writer) *GobHandler {
	return &GobHandler{gob.NewEncoder(w), gob.NewDecoder(r)}
}

func (h *GobHandler) Encode(i interface{}) error {
	return h.enc.Encode(i)
}

func (h *GobHandler) Decode(data interface{}) error {
	return h.dec.Decode(data)
}

func Decode[T Schema](h *GobHandler) (T, error) {
	var data T
	err := h.Decode(&data)
	return data, err
}
