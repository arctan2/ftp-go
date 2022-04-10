package client

import (
	"fmt"

	"github.com/chzyer/readline"
)

type envStruct struct {
	curDir      string
	curDirFiles dirFiles
	rln         *readline.Instance
}

type cmds interface {
	cd([]string) error
	ls() error
	pwd()
}

type env interface {
	cmds
	getCurDir() string
	setCurDir(string) error

	getCurDirFiles() *dirFiles
	setCurDirFiles(dirFiles)

	refreshCurDirFiles() error

	curRln() *readline.Instance
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
