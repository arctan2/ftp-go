package common

import (
	"encoding/gob"
	"io"
	"net"
	"os"
)

type GobHandler struct {
	enc *gob.Encoder
	dec *gob.Decoder
}

type Schema interface {
	[]FileStruct | FileStruct | DirName | string | bool | Res
}

func (r Res) Error() string {
	return r.Data
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
