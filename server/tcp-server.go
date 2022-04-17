package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"ftp/common"
	serverUtils "ftp/server/server-utils"
)

func handleConn(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	gh := common.NewGobHandler(reader, conn)

	switch cmd, _ := reader.ReadString('\n'); strings.TrimSpace(cmd) {
	case "pwd":
		if dirPath, err := serverUtils.GetAbsPath("./"); err == nil {
			gh.Encode(dirPath)
		}
	case "ls":
		dirName, err := common.Decode[string](gh)

		if err != nil {
			gh.Encode(err.Error())
			return
		}

		fileList, err := serverUtils.GetFileList(dirName)
		if err != nil {
			gh.Encode(err.Error())
			return
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
				gh.Encode(common.Res{Err: true, Data: "The system cannot find the file specified."})
			} else {
				gh.Encode(common.Res{Err: true, Data: err.Error()})
			}
			break
		} else {
			if !fStat.IsDir() {
				gh.Encode(common.Res{Err: true, Data: cdToDir + " is not a directory."})
				break
			}
		}
		absPath, err := filepath.Abs(cdToDir)
		gh.Encode(common.Res{Err: false, Data: filepath.ToSlash(absPath)})
	case "get":
		filePath, err := common.Decode[string](gh)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if _, err := os.Stat(filePath); err != nil {
			break
		} else {
			tmpDir, err := os.MkdirTemp("", "ftp-go-server")
			defer os.RemoveAll(tmpDir)
			fileName := serverUtils.GetFileName(filePath)
			zipPath := tmpDir + "/" + fileName + ".zip"

			gh.Encode("zipping...")
			common.ZipSource([]string{filePath}, zipPath, gh)
			zfStat, err := os.Stat(zipPath)

			if err != nil {
				fmt.Println(err.Error())
				break
			}

			gh.Encode(common.FileStruct{Name: zfStat.Name(), IsDir: true, Size: zfStat.Size()})

			serverUtils.SendFile(zipPath, conn)
		}
	}
}

func StartTcpServer(tcpAddr string) {
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
