package client

import (
	"fmt"
	"ftp/common"
	"io"
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

func handleCmd[T env](curEnv T, eh envHandler) bool {
	if eh.currentEnvType() == LOCAL {
		curEnv.curRln().SetPrompt("local> ")
	} else {
		curEnv.curRln().SetPrompt(eh.currentRemoteName() + "> ")
	}
	cmdExpr, err := curEnv.curRln().Readline()
	if err == readline.ErrInterrupt {
		return true
	} else if err == io.EOF {
		return true
	}

	cmdExpr = strings.TrimSpace(cmdExpr)
	cmdArgs := deleteEmptyStr(strings.Split(cmdExpr, " "))

	if len(cmdArgs) == 0 {
		return false
	}

	switch cmd := cmdArgs[0]; cmd {
	case "quit", "exit", "logout":
		return true
	case "clear":
		common.ClearScreen()
	case "pwd":
		curEnv.pwd()
	case "net":
		if err := eh.handleCmd(cmdArgs); err != nil {
			fmt.Println(err.Error())
		}
	case "cd":
		if err := curEnv.cd(cmdArgs); err != nil {
			fmt.Println(err.Error())
			break
		}
		if err := curEnv.refreshCurDirFiles(); err != nil {
			fmt.Println(err.Error())
		}
	case "ls":
		if err := curEnv.ls(); err != nil {
			fmt.Println(err.Error())
			break
		}
	default:
		if eh.currentEnvType() == REMOTE {
			rEnv := eh.currentRemote()
			switch cmd {
			case "get":
				rEnv.get(cmdArgs, eh.localEnv().getDownloadDir())
				return false
			}
		} else {
			lEnv := eh.localEnv()
			switch cmd {
			case "ddir":
				if err := lEnv.setDownloadDir(cmdArgs); err != nil {
					fmt.Printf("%s\ncouldn't set download directory\n", err.Error())
					break
				}
				fmt.Println(lEnv.getDownloadDir())
				return false
			}
		}
		fmt.Printf("unknown command '%s'\n", cmd)
	}
	return false
}

func StartClient() {
	eh := newEnvHandler()
	defer eh.closeAllRemotesRlns()
	for {
		if eh.currentEnvType() == LOCAL {
			if handleCmd(eh.localEnv(), eh) {
				break
			}
		} else {
			if handleCmd(eh.currentRemote(), eh) {
				break
			}
		}
	}
}
