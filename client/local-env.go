package client

import (
	"fmt"
	"os"
	"path/filepath"

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

func newLocalEnv(downloadDir string) localEnv {
	es := &envStruct{curDirFiles: make(dirFiles, 0)}
	dirListFunc := es.curDirFiles.ListFunc()

	completer := readline.NewPrefixCompleter(
		readline.PcItem("cd", readline.PcItemDynamic(dirListFunc)),
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
	return nil
}

func (le *localEnvStruct) cd([]string) error {
	return nil
}

func (le *localEnvStruct) ls() error {
	return nil
}
