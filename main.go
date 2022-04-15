package main

import (
	"flag"
	"fmt"
	"ftp/client"
	"ftp/common"
	"ftp/server"
	httpServer "ftp/server/http-server"
	"os"
	"strings"
)

func main() {
	var (
		isServerMode bool
		isClientMode bool
		isHttpMode   bool

		port string
		ipv4 string
		addr string
	)

	flag.BoolVar(&isHttpMode, "http", false, "run http server")

	flag.BoolVar(&isServerMode, "s", false, "run as server")
	flag.BoolVar(&isServerMode, "server", false, "run as server")

	flag.BoolVar(&isClientMode, "c", false, "run as client")
	flag.BoolVar(&isClientMode, "client", false, "run as client")

	flag.StringVar(&port, "port", "5000", "port to run on")
	flag.StringVar(&port, "p", "5000", "port to run on")

	i4 := common.GetIPv4Str()
	flag.StringVar(&ipv4, "ipv4", i4, "ipv4 address")
	flag.StringVar(&ipv4, "i4", i4, "ipv4 address")

	flag.StringVar(&addr, "addr", "", "server address")
	flag.StringVar(&addr, "a", "", "server address")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, `
-s, --server   run as server
-c, --client   run as client

--http  run http server 

-p, --port     port to run on
-i4, --ipv4    ipv4 address  
-a, --addr     server address
`)
	}

	flag.Parse()

	if addr != "" {
		addrParts := strings.Split(addr, ":")
		if len(addrParts) != 2 {
			fmt.Println("invalid address.")
			os.Exit(1)
		}
		ipv4, port = addrParts[0], addrParts[1]
	}

	if isHttpMode {
		httpServer.StartHttpServer(port)
		return
	}
	if isServerMode {
		server.StartTcpServer(ipv4 + ":" + port)
		return
	}
	if isClientMode {
		client.StartClient()
		return
	}
	fmt.Println("please specify -c or -s or --http flags.")
}
