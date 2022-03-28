package client

import (
	"fmt"
	"ftp/common"
	"log"
	"net"
	"os"
	"strings"

	"github.com/fatih/color"
)

func DialAndSendCmd(cmd string) (net.Conn, error) {
	tcpAddr := common.GetTcpAddrStr("5000")
	conn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	conn.Write([]byte(cmd))
	return conn, err
}

func getWorkingDir() (common.DirName, error) {
	conn, err := DialAndSendCmd("pwd\n")
	defer conn.Close()

	if err != nil {
		return "", err
	}

	gh := common.NewGobHandler(conn)

	return common.Decode[common.DirName](gh)
}

func StartClient(PORT string) {
	blue := color.New(color.FgBlue).PrintfFunc()

	var (
		files      []common.FileStruct
		currentDir common.DirName
		conn       net.Conn
	)

	currentDir, err := getWorkingDir()
	if err != nil {
		log.Fatal("unable to get working directory from server. Closing...\n", err.Error())
	}

	for {
		cmd, err := common.Scan("> ")

		if err != nil {
			log.Fatal(err.Error())
		}

		switch strings.TrimSpace(cmd) {
		case "quit", "exit", "logout":
			os.Exit(0)
		case "clear":
			common.ClearScreen()
		case "pwd":
			fmt.Println(currentDir)
		case "ls":
			if files == nil {
				conn, err = DialAndSendCmd(cmd)
				if err != nil {
					break
				}
				gh := common.NewGobHandler(conn)
				files, _ = common.Decode[[]common.FileStruct](gh)
			}

			for _, f := range files {
				if f.IsDir {
					blue("%s  ", f.Name)
				} else {
					fmt.Printf("%s  ", f.Name)
				}
			}
			fmt.Println()
		}

		if conn != nil {
			conn.Close()
			conn = nil
		}
	}

	//if err != nil {
	//log.Fatal(err.Error())
	//}

	//defer conn.Close()

	//os.MkdirAll("./tmp-client", os.ModePerm)
	//file, err := os.Create("./tmp-client/test-recevied.zip")
	//if err != nil {
	//log.Fatal(err.Error())
	//}

	//io.Copy(file, conn)
	//file.Close()

	//common.UnzipSource("./tmp-client/test-recevied.zip", "./test")
	//os.RemoveAll("./tmp-client")
}
