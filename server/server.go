package server

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"ftp/common"
)

func sendFile(fileName string, conn net.Conn) {
	file, err := os.Open(strings.TrimSpace(fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	io.Copy(conn, file)
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	switch cmd, _ := bufio.NewReader(conn).ReadString('\n'); strings.TrimSpace(cmd) {
	case "pwd":
		fp, _ := filepath.Abs("./")
		gh := common.NewGobHandler(conn)
		gh.Encode(fp)
	case "ls":
		gh := common.NewGobHandler(conn)

		dirName, err := common.Decode[string](gh)

		if err != nil {
			fmt.Println(err.Error())
			break
		}

		files, err := ioutil.ReadDir(dirName)
		if err != nil {
			return
		}
		var fileList []common.FileStruct
		for _, f := range files {
			fileStruc := common.FileStruct{Name: f.Name(), IsDir: f.IsDir(), Size: f.Size()}
			fileList = append(fileList, fileStruc)
		}
		gh.Encode(fileList)
	case "get":
		os.Mkdir("./tmp", os.ModePerm)
		common.ZipSource("./files/test-dir", "./tmp/test-dir.zip")
		sendFile("./tmp/test-dir.zip", conn)
		os.RemoveAll("./tmp/")
	}
}

func StartServer(PORT string) {
	tcpAddr := common.GetTcpAddrStr(PORT)

	ln, err := net.Listen("tcp", tcpAddr)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("running on", tcpAddr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err.Error())
		}
		go handleConn(conn)
	}
}
