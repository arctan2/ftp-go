package client

import (
	"fmt"
	"ftp/common"
	"log"
	"net"
	"os"
	"strings"
)

func DialAndCmd(cmd string) (net.Conn, error) {
	tcpAddr := common.GetTcpAddrStr("5000")
	conn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}
	conn.Write([]byte(cmd + "\n"))
	return conn, err
}

func getWorkingDir() (string, error) {
	conn, err := DialAndCmd("pwd")

	if err != nil {
		return "", err
	}
	defer conn.Close()

	gh := common.NewGobHandler(conn, conn)
	d, err := common.Decode[common.DirName](gh)

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(d)), err
}

func deleteEmptyStr(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func StartClient(PORT string) {
	var (
		curDir      string
		downloadDir = "./downloads"
	)

	DialAndCmd("hehe\n")
	fmt.Println("connecting...")
	fmt.Println("getting current working dir...")

	curDir, err := getWorkingDir()
	if err != nil {
		log.Fatal(err.Error(), "\nunable to get working directory from server. Closing...\n")
	}

	fmt.Println()

	for {
		cmdExpr, err := common.Scan("> ")
		cmdExpr = strings.TrimSpace(cmdExpr)
		cmdArgs := deleteEmptyStr(strings.Split(cmdExpr, " "))

		if err != nil {
			log.Fatal(err.Error())
		}

		switch cmd := cmdArgs[0]; cmd {
		case "quit", "exit", "logout":
			os.Exit(0)
		case "clear":
			common.ClearScreen()
		case "pwd":
			fmt.Println(curDir)
		case "ddir":
			fmt.Println(downloadDir)
		case "cd":
			curDir = cd(cmdArgs, curDir)
		case "ls":
			ls(curDir)
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
