package common

import (
	"encoding/gob"
	"fmt"
	"io"
	"net"
)

type GobHandler struct {
	enc *gob.Encoder
	dec *gob.Decoder
}

type Schema interface {
	[]FileStruct | FileStruct | DirName | string | []string | bool | Res | ZipProgress | int64
}

var (
	LOCAL_HOST = "127.0.0.1"
)

func (r Res) Error() string {
	return r.Msg
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

func (h *GobHandler) EncodeSuccess(data interface{}) error {
	if err := h.Encode(Res{Err: false, Msg: "success"}); err != nil {
		return err
	}
	return h.Encode(data)
}
func (h *GobHandler) EncodeErr(msg string) error {
	if err := h.Encode(Res{Err: true, Msg: msg}); err != nil {
		return err
	}
	return nil
}

func DecodeWithRes[T Schema](h *GobHandler) (T, error) {
	var data T
	var res Res
	if err := h.Decode(&res); err != nil {
		return data, err
	}
	if res.Err {
		return data, res
	}
	err := h.Decode(&data)
	return data, err
}

func Decode[T Schema](h *GobHandler) (T, error) {
	var data T
	err := h.Decode(&data)
	return data, err
}
