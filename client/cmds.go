package client

import (
	"fmt"
	"ftp/common"
	"path/filepath"

	"github.com/fatih/color"
)

var blue = color.New(color.FgBlue).PrintfFunc()

func getCurDirFiles(curDir string) (f []common.FileStruct, e error) {
	conn, err := DialAndCmd("ls")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	gh := common.NewGobHandler(conn, conn)
	if err := gh.Encode(curDir); err != nil {
		return nil, err
	}

	return common.Decode[[]common.FileStruct](gh)
}

func cd(cmdArgs []string, curDir string) string {
	if len(cmdArgs) == 1 {
		fmt.Println("err: missing operand after cd")
		return curDir
	}

	conn, err := DialAndCmd("cd")
	if err != nil {
		fmt.Println(err.Error())
		return curDir
	}
	defer conn.Close()

	gh := common.NewGobHandler(conn, conn)

	cdDirName := cmdArgs[1]
	if cdDirName != "" && cdDirName[0] == '"' && cdDirName[len(cdDirName)-1] == '"' {
		cdDirName = cdDirName[1 : len(cdDirName)-1]
	}

	if filepath.IsAbs(cdDirName) {
		gh.Encode(cdDirName)
	} else {
		gh.Encode(curDir + "/" + cdDirName)
	}

	exists, err := common.Decode[common.Res](gh)
	if err != nil {
		fmt.Println(err.Error())
	}

	if exists.Err {
		fmt.Println(exists.Error())
	}

	return exists.Data
}

func ls(curDir string) []common.FileStruct {
	curFiles, err := getCurDirFiles(curDir)

	if err != nil {
		fmt.Println(err.Error(), "\nunable to get files.")
		return nil
	}

	for _, f := range curFiles {
		if f.IsDir {
			blue("%s  ", f.Name)
		} else {
			fmt.Printf("%s  ", f.Name)
		}
	}

	fmt.Println()
	return curFiles
}
