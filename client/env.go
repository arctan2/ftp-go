package client

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chzyer/readline"
)

type envStruct struct {
	curDir      string
	curDirFiles dirFiles
	rln         *readline.Instance
}

type env interface {
	getCurDir() string
	setCurDir(string) error

	getCurDirFiles() *dirFiles
	setCurDirFiles(dirFiles)

	curRln() *readline.Instance
}

type cmds interface {
	cd([]string) error
	ls() error
	pwd()
}

type localEnvStruct struct {
	envStruct
	downloadDir string
}

type localEnv interface {
	cmds
	env
	getDownloadDir() string
	setDownloadDir([]string) error
}

func newLocalEnv(downloadDir string) localEnv {
	var e localEnv = &localEnvStruct{downloadDir: downloadDir}
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

func (e *envStruct) getCurDir() string {
	return e.curDir
}

func (e *envStruct) setCurDir(d string) error {
	e.curDir = d
	return nil
}

func (e *envStruct) getCurDirFiles() *dirFiles {
	return &e.curDirFiles
}

func (e *envStruct) setCurDirFiles(df dirFiles) {
	e.curDirFiles = df
}

func (e *envStruct) pwd() {
	fmt.Println(e.curDir)
}

func (e *envStruct) curRln() *readline.Instance {
	return e.rln
}

func (le *localEnvStruct) cd([]string) error {
	return nil
}

func (le *localEnvStruct) ls() error {
	return nil
}
