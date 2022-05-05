package client

import (
	"errors"
	"fmt"
	"ftp/common"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
)

type localEnvStruct struct {
	*envStruct
	downloadDir string
}

type localEnv interface {
	env
	getDownloadDir() string
	setDownloadDir([]string) error
}

func newLocalEnv(downloadDir, curDir string, netListFunc func(string) []string) localEnv {
	es := &envStruct{curDirFiles: make(dirFiles, 0), curDir: curDir}
	dirListFunc := es.curDirFiles.ListFunc()

	completer := readline.NewPrefixCompleter(
		readline.PcItem("cd", readline.PcItemDynamic(dirListFunc)),
		readline.PcItem("net switch", readline.PcItemDynamic(netListFunc)),
	)

	rln, _ := readline.NewEx(&readline.Config{
		Prompt:              "> ",
		AutoComplete:        completer,
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		FuncFilterInputRune: filterInput,
	})
	es.rln = rln
	var e localEnv = &localEnvStruct{downloadDir: filepath.ToSlash(downloadDir), envStruct: es}
	return e
}

func (le *localEnvStruct) getDownloadDir() string {
	return le.downloadDir
}

func (le *localEnvStruct) setDownloadDir(cmdArgs []string) error {
	if len(cmdArgs) > 1 && cmdArgs[1] != "" {
		arg1 := cmdArgs[1]
		switch arg1 {
		case "-d", "--default":
			abs, _ := filepath.Abs("./downloads")
			le.downloadDir = filepath.ToSlash(abs)
			return nil
		case "-s", "--set":
			le.downloadDir = le.curDir
			return nil
		}
		if len(arg1) > 2 {
			if arg1[0] == '"' {
				arg1 = arg1[1:]
			}
			if arg1[len(arg1)-1] == '"' {
				arg1 = arg1[0 : len(arg1)-1]
			}
		}

		if filepath.IsAbs(arg1) || arg1[0] == '/' {
			arg1, _ = filepath.Abs(arg1)
		}
		if _, err := os.Stat(arg1); err != nil {
			return err
		} else {
			ddir, err := filepath.Abs(arg1)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}
			le.downloadDir = filepath.ToSlash(ddir)
		}
	}
	return nil
}

func (le *localEnvStruct) refreshCurDirFiles() error {
	files, err := ioutil.ReadDir(le.curDir)
	if err != nil {
		return err
	}
	var fileList []common.FileStruct
	for _, f := range files {
		fileStruc := common.FileStruct{Name: f.Name(), IsDir: f.IsDir(), Size: f.Size()}
		fileList = append(fileList, fileStruc)
	}
	le.setCurDirFiles(fileList)
	return nil
}

func (le *localEnvStruct) cd(cmdArgs []string) error {
	if len(cmdArgs) == 1 {
		return errors.New("err: missing operand after cd")
	}

	cdDirNameArg := cmdArgs[1]
	cdToDir := ""

	if cdDirNameArg != "" && cdDirNameArg[0] == '"' && cdDirNameArg[len(cdDirNameArg)-1] == '"' {
		cdDirNameArg = cdDirNameArg[1 : len(cdDirNameArg)-1]
	}

	if !filepath.IsAbs(cdDirNameArg) && cdDirNameArg[0] != '/' {
		cdToDir = le.curDir + "/" + cdDirNameArg
	}

	if fStat, err := os.Stat(cdToDir); err != nil {
		if os.IsNotExist(err) {
			return errors.New("The system cannot find the file specified.")
		} else {
			return errors.New(err.Error())
		}
	} else {
		if !fStat.IsDir() {
			return errors.New(cdToDir + " is not a directory.")
		}
	}
	absPath, err := filepath.Abs(cdToDir)
	if err != nil {
		return err
	}
	le.curDir = filepath.ToSlash(absPath)
	le.refreshCurDirFiles()
	return nil
}

func (le *localEnvStruct) ls() error {
	for _, f := range le.curDirFiles {
		fName := f.Name
		if strings.ContainsRune(fName, ' ') {
			fName = "\"" + fName + "\""
		}
		if f.IsDir {
			blue("%s  ", fName)
		} else {
			fmt.Printf("%s  ", fName)
		}
	}
	return nil
}
