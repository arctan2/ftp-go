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

	reader := bufio.NewReader(conn)
	gh := common.NewGobHandler(reader, conn)

	switch cmd, _ := reader.ReadString('\n'); strings.TrimSpace(cmd) {
	case "pwd":
		fp, _ := filepath.Abs("./")
		gh.Encode(common.DirName(filepath.ToSlash(fp)))
	case "ls":
		dirName, err := common.Decode[string](gh)

		if err != nil {
			fmt.Println("unable to get dirname: ", err.Error())
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
	case "cd":
		cdToDir, err := common.Decode[string](gh)
		if err != nil {
			fmt.Println(err.Error())
			break
		}

		if fStat, err := os.Stat(cdToDir); err != nil {
			if os.IsNotExist(err) {
				gh.Encode("The system cannot find the file specified.")
			} else {
				gh.Encode(err.Error())
			}
			break
		} else {
			if !fStat.IsDir() {
				gh.Encode(cdToDir + " is not a directory.")
				break
			}
		}
		absPath, err := filepath.Abs(cdToDir)
		gh.Encode(filepath.ToSlash(absPath))
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
