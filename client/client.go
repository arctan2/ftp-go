package client

import (
	"fmt"
	"ftp/common"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
)

type dirFiles []common.FileStruct

var tcpAddr string

func (df dirFiles) nameSlice() (fileNames []string) {
	for _, f := range df {
		fileNames = append(fileNames, f.Name)
	}
	return
}

func (df *dirFiles) ListFunc() func(string) []string {
	return func(line string) []string {
		return df.nameSlice()
	}
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

func filterInput(r rune) (rune, bool) {
	switch r {
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func StartClient(ipv4, port string) {
	pwd, _ := os.Getwd()
	rEnv := newRemoteEnv(filepath.ToSlash(pwd+"/downloads"), dialer{addr: ipv4 + ":" + port, port: port, ipv4: ipv4})

	fmt.Println("getting current working dir...")

	err := rEnv.fetchCurDirFromServer()
	if err != nil {
		log.Fatal(err.Error(), "\nunable to get working directory from server. Closing...\n")
	}

	fmt.Println("fetching file names...")

	err = rEnv.fetchCurDirFilesFromServer()
	if err != nil {
		log.Fatal(err.Error(), "\nunable to get directory files from server. Closing...\n")
	}

	fmt.Println()

	dirListFunc := rEnv.getCurDirFiles().ListFunc()

	completer := readline.NewPrefixCompleter(
		readline.PcItem("cd", readline.PcItemDynamic(dirListFunc)),
		readline.PcItem("get", readline.PcItemDynamic(dirListFunc)),
	)

	rln, err := readline.NewEx(&readline.Config{
		Prompt:              "> ",
		AutoComplete:        completer,
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer rln.Close()

	log.SetOutput(rln.Stderr())

	for {
		cmdExpr, err := rln.Readline()
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
			rln.Close()
			os.Exit(0)
		case "clear":
			common.ClearScreen()
		case "pwd":
			rEnv.pwd()
		case "ddir":
			if err := rEnv.setDownloadDir(cmdArgs); err != nil {
				fmt.Printf("%s\ncouldn't set download directory\n", err.Error())
				break
			}
			fmt.Println(rEnv.getDownloadDir())
		case "cd":
			if err := rEnv.cd(cmdArgs); err != nil {
				fmt.Println(err.Error())
				break
			}
			if err := rEnv.fetchCurDirFilesFromServer(); err != nil {
				fmt.Println(err.Error())
			}
		case "ls":
			if err := rEnv.ls(); err != nil {
				fmt.Println(err.Error())
				break
			}
		case "get":
			rEnv.get(cmdArgs)
		default:
			fmt.Printf("unknown command '%s'\n", cmd)
		}
	}
}
