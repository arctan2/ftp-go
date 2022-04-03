package client

import (
	"fmt"
	"ftp/common"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

type dirFiles []common.FileStruct

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

func filterInput(r rune) (rune, bool) {
	switch r {
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func StartClient(PORT string) {
	var (
		curDir      string
		curDirFiles dirFiles
		downloadDir = "./downloads"
	)

	fmt.Println("getting current working dir...")

	curDir, err := getWorkingDir()
	if err != nil {
		log.Fatal(err.Error(), "\nunable to get working directory from server. Closing...\n")
	}

	fmt.Println("fetching file names...")

	curDirFiles, err = getCurDirFiles(curDir)
	if err != nil {
		log.Fatal(err.Error(), "\nunable to get directory files from server. Closing...\n")
	}

	fmt.Println()

	completer := readline.NewPrefixCompleter(
		readline.PcItem("cd", readline.PcItemDynamic(curDirFiles.ListFunc())),
		readline.PcItem("get", readline.PcItemDynamic(curDirFiles.ListFunc())),
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

		switch cmd := cmdArgs[0]; cmd {
		case "quit", "exit", "logout":
			rln.Close()
			os.Exit(0)
		case "clear":
			common.ClearScreen()
		case "pwd":
			fmt.Println(curDir)
		case "ddir":
			fmt.Println(downloadDir)
		case "cd":
			curDir = cd(cmdArgs, curDir)
			cf, err := getCurDirFiles(curDir)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				curDirFiles = cf
			}
		case "ls":
			cf := ls(curDir)
			if cf != nil {
				curDirFiles = cf
			}
		case "get":
			get(curDir, cmdArgs)
		default:
			fmt.Printf("unknown command '%s'\n", cmd)
			func(i interface{}) {}(curDirFiles)
		}
	}
}
