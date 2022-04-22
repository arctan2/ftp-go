package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ftp/common"
	"ftp/config"
	serverUtils "ftp/server/server-utils"
)

func handleConn(conn net.Conn, logger common.Logger, c config.ConfigHandler) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	gh := common.NewGobHandler(reader, conn)

	switch cmd, _ := reader.ReadString('\n'); strings.TrimSpace(cmd) {
	case "pwd":
		if dirPath, err := serverUtils.GetAbsPath(c.GetInitDir()); err == nil {
			gh.EncodeSuccess(dirPath)
		} else {
			gh.EncodeErr(err.Error())
		}
	case "ls":
		dirName, err := common.Decode[string](gh)

		if err != nil {
			gh.EncodeErr(err.Error())
			return
		}

		if c.IsRestricted(dirName) {
			gh.EncodeErr("permission denied.")
			return
		}

		fileList, err := serverUtils.GetFileList(dirName)
		if err != nil {
			gh.EncodeErr(err.Error())
			return
		}
		gh.EncodeSuccess(fileList)
	case "cd":
		cdToDir, err := common.Decode[string](gh)
		if err != nil {
			gh.EncodeErr(err.Error())
			break
		}

		if fStat, err := os.Stat(cdToDir); err != nil {
			if os.IsNotExist(err) {
				gh.EncodeErr("The system cannot find the file specified.")
			} else {
				gh.EncodeErr(err.Error())
			}
			break
		} else {
			if !fStat.IsDir() {
				gh.EncodeErr(" is not a directory.")
				break
			}
		}
		absPath, err := filepath.Abs(cdToDir)
		gh.EncodeSuccess(common.Res{Err: false, Msg: filepath.ToSlash(absPath)})
	case "get":
		filePaths, err := common.Decode[[]string](gh)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		for _, f := range filePaths {
			if c.IsRestricted(f) {
				gh.EncodeErr("get file " + f + ": permission denied.")
				return
			}
			if _, err := os.Stat(f); err != nil {
				gh.EncodeErr("file " + f + " not found")
				return
			}
		}

		tmpDir, err := os.MkdirTemp("", "ftp-go-server")
		defer os.RemoveAll(tmpDir)
		fileName := "downloads"
		zipPath := tmpDir + "/" + fileName + ".zip"

		gh.EncodeSuccess("zipping...")
		common.ZipSource(filePaths, zipPath, gh)
		zfStat, err := os.Stat(zipPath)

		if err != nil {
			gh.EncodeErr(err.Error())
			break
		}

		gh.EncodeSuccess(common.FileStruct{Name: zfStat.Name(), IsDir: true, Size: zfStat.Size()})

		b, _ := json.MarshalIndent(filePaths, "", "\t")
		logger.Log("get on", time.Now(), string(b))

		serverUtils.SendFile(zipPath, conn)
	}
}

func StartTcpServer(tcpAddr string) {
	curDate := time.Now().Local().Format("01-02-2006")
	logger, err := common.NewLoggerWithDirAndFileName("./logs/tcp", curDate+".log")
	ln, err := net.Listen("tcp", tcpAddr)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer ln.Close()

	log.Println("running on", tcpAddr)

	c, err := config.ParseConfigFile("ftp-config.json")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err.Error())
		}
		go handleConn(conn, logger, c)
	}
}
