package client

import (
	"fmt"
	"ftp/common"
	"log"
	"net"
	"os"
	"strings"

	"github.com/fatih/color"
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

func getCurDirFiles(curDir string) (f []common.FileStruct, e error) {
	conn, err := DialAndCmd("ls")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer conn.Close()

	gh := common.NewGobHandler(conn, conn)
	if err := gh.Encode(curDir); err != nil {
		return nil, err
	}

	return common.Decode[[]common.FileStruct](gh)
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
	blue := color.New(color.FgBlue).PrintfFunc()

	var (
		curFiles    []common.FileStruct
		curDir      string
		downloadDir = "./downloads"
		conn        net.Conn
	)

	DialAndCmd("hehe\n")
	fmt.Println("connecting...")
	fmt.Println("getting current working dir...")

	curDir, err := getWorkingDir()
	if err != nil {
		log.Fatal(err.Error(), "\nunable to get working directory from server. Closing...\n")
	}

	fmt.Println("fetching current directory file names...")

	if curFiles, err = getCurDirFiles(curDir); err != nil {
		log.Fatal(err.Error(), "\nunable to fetch current directory file names. Closing...\n")
	}

	fmt.Println()

	for {
		cmdExpr, err := common.Scan("> ")
		cmdExpr = strings.TrimSpace(cmdExpr)
		cmdArgs := deleteEmptyStr(strings.Split(cmdExpr, " "))

		if err != nil {
			log.Fatal(err.Error())
		}

	cmdSwh:
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
			if len(cmdArgs) == 1 {
				fmt.Println("missing operand for command: cd")
				break
			}
			dirName := cmdArgs[1]

			for _, f := range curFiles {
				if f.Name == dirName {
					if !f.IsDir {
						fmt.Printf("%s is not a directory.\n", dirName)
						break
					}
					curDir += "/" + dirName
					break cmdSwh
				}
			}
			fmt.Printf("There is no directory named '%s'\n", dirName)
		case "ls":
			curFiles, err := getCurDirFiles(curDir)
			if err != nil {
				fmt.Println(err.Error(), "\nunable to get files.")
				break
			}

			for _, f := range curFiles {
				if f.IsDir {
					blue("%s  ", f.Name)
				} else {
					fmt.Printf("%s  ", f.Name)
				}
			}
			fmt.Println()
		}

		if conn != nil {
			conn.Close()
			conn = nil
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
