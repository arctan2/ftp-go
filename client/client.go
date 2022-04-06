package client

import (
	"fmt"
	"ftp/common"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
)

var tcpAddr string

func filterInput(r rune) (rune, bool) {
	switch r {
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func StartClient(ipv4, port string) {
	pwd, _ := os.Getwd()
	eh := newEnvHandler(newLocalEnv(filepath.ToSlash(pwd + "/downloads")))
	eh.setCurrentEnvType(REMOTE)

	for i := 0; i < 3; i++ {
		eh.addRemoteEnv(newRemoteEnv(dialer{addr: ipv4 + ":" + port, port: port, ipv4: ipv4}))
		eh.setCurRemoteIdx(i)
		eh.currentRemote().initRemote()
	}
	eh.setCurRemoteIdx(0)
	curEnv := eh.currentRemote()

	for {
		cmdExpr, err := curEnv.curRln().Readline()
		if err == readline.ErrInterrupt {
			if len(cmdExpr) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		cmdExpr = strings.TrimSpace(cmdExpr)
		cmdArgs := deleteEmptyStr(strings.Split(cmdExpr, " "))

		if len(cmdArgs) == 0 {
			continue
		}

		switch cmd := cmdArgs[0]; cmd {
		case "quit", "exit", "logout":
			curEnv.curRln().Close()
			os.Exit(0)
		case "clear":
			common.ClearScreen()
		case "pwd":
			curEnv.pwd()
		case "ddir":
			//if err := curEnv.setDownloadDir(cmdArgs); err != nil {
			//fmt.Printf("%s\ncouldn't set download directory\n", err.Error())
			//break
			//}
			//fmt.Println(rEnv.getDownloadDir())
		case "se":
			i, _ := strconv.Atoi(cmdArgs[1])
			eh.setCurRemoteIdx(i)
			curEnv = eh.currentRemote()
		case "cd":
			if err := curEnv.cd(cmdArgs); err != nil {
				fmt.Println(err.Error())
				break
			}
			if err := curEnv.fetchCurDirFilesFromServer(); err != nil {
				fmt.Println(err.Error())
			}
		case "ls":
			if err := curEnv.ls(); err != nil {
				fmt.Println(err.Error())
				break
			}
		case "get":
			curEnv.get(cmdArgs, eh.localEnv().getDownloadDir())
		default:
			fmt.Printf("unknown command '%s'\n", eh)
		}
	}
}
