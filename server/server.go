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

func getFileName(fp string) string {
	parts := strings.Split(fp, "/")
	return parts[len(parts)-1]
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
			return
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

		if fStat, err := os.Stat(filePath); err != nil {
			break
		} else {
			fileName := getFileName(filePath)
			zipPath := "./.tmp/" + fileName + ".zip"

			if fStat.IsDir() {
				os.Mkdir("./.tmp", os.ModePerm)
				common.ZipSource(filePath, zipPath)
				if err := gh.Encode(fileName + ".zip"); err != nil {
					break
				}
				sendFile(zipPath, conn)
				os.RemoveAll("./.tmp/")
			}
		}
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
