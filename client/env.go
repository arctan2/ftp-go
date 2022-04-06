package client

import "fmt"

type envStruct struct {
	curDir      string
	curDirFiles dirFiles
}

type env interface {
	getCurDir() string
	setCurDir(string) error

	getCurDirFiles() *dirFiles
	setCurDirFiles(dirFiles)
}

type cmds interface {
	cd([]string) error
	ls() error
	pwd()
}

type localEnvStruct struct {
	envStruct
}

type localEnv interface {
	cmds
	env
}

func newLocalEnv() localEnv {
	var e localEnv = &localEnvStruct{}
	return e
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

func (le *localEnvStruct) cd([]string) error {
	return nil
}

func (le *localEnvStruct) ls() error {
	return nil
}
