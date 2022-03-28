package main

import (
	"fmt"
	"os"

	"ftp/client"
	"ftp/server"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println(`
Please specify a flag:
--client | -c : run as client
--server | -s : run as server
--help   | -h : help
		`)
		os.Exit(1)
	}
	if len(os.Args) > 2 {
		fmt.Println("Please specify only one flag.")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "--server", "-s":
		server.StartServer("5000")
	case "--client", "-c":
		client.StartClient("5000")
	case "--help", "-h":
		fmt.Println(`
--client | -c : run as client
--server | -s : run as server
--help   | -h : help
		`)
		os.Exit(0)
	default:
		fmt.Printf(`Invalid flag "%s"
	use -h for help.
`, os.Args[1])
		os.Exit(1)
	}
}
