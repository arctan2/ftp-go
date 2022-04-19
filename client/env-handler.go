package client

import (
	"encoding/gob"
	"errors"
	"fmt"
	"ftp/common"
	"net"
	"os"
	"path/filepath"
	"strconv"
)

type envHandlerStruct struct {
	lEnv          localEnv
	remotes       map[string]remoteEnv
	curRemoteName string
	curEnvType    int
}

type envHandler interface {
	addRemoteEnv(string, remoteEnv)

	currentEnvType() int
	setCurrentEnvType(int)

	currentRemote() remoteEnv
	currentRemoteName() string
	localEnv() localEnv
	setCurRemoteName(string) error
	handleCmd([]string) error
	loadRemotesFromGobFile()

	closeAllRemotesRlns()
}

const (
	LOCAL int = iota
	REMOTE
)

func newEnvHandler() envHandler {
	pwd, _ := os.Getwd()
	var eh envHandler = &envHandlerStruct{
		lEnv:       newLocalEnv(filepath.ToSlash(pwd+"/downloads"), pwd),
		remotes:    make(map[string]remoteEnv),
		curEnvType: LOCAL,
	}
	eh.loadRemotesFromGobFile()
	eh.localEnv().refreshCurDirFiles()
	return eh
}

func (eh *envHandlerStruct) loadRemotesFromGobFile() {
	rf, err := os.Open("remotes.gob")
	if err != nil {
		return
	}
	defer rf.Close()

	dec := gob.NewDecoder(rf)
	remotes := make(map[string]string)
	if err = dec.Decode(&remotes); err != nil {
		return
	}
	for rName, addr := range remotes {
		rEnv := newRemoteEnv(newDialer(addr), rName)
		rEnv.initRemote(false)
		eh.addRemoteEnv(rName, rEnv)
	}
}

func (eh *envHandlerStruct) saveRemotesToGobFile() {
	remoteMap := make(map[string]string)
	for rName, rEnv := range eh.remotes {
		remoteMap[rName] = rEnv.dialer().addr
	}

	var (
		rf  *os.File
		err error
	)

	if common.IsPathExists("remotes.gob") {
		rf, err = os.Open("remotes.gob")
	} else {
		rf, err = os.Create("remotes.gob")
	}
	if err != nil {
		return
	}
	defer rf.Close()

	enc := gob.NewEncoder(rf)
	enc.Encode(remoteMap)
}

func (eh *envHandlerStruct) closeAllRemotesRlns() {
	for _, r := range eh.remotes {
		r.curRln().Close()
	}
}

func (eh *envHandlerStruct) addRemoteEnv(remoteName string, rEnv remoteEnv) {
	eh.remotes[remoteName] = rEnv
}

func (eh *envHandlerStruct) currentEnvType() int {
	return eh.curEnvType
}

func (eh *envHandlerStruct) currentRemote() remoteEnv {
	return eh.remotes[eh.curRemoteName]
}

func (eh *envHandlerStruct) currentRemoteName() string {
	return eh.curRemoteName
}

func (eh *envHandlerStruct) localEnv() localEnv {
	return eh.lEnv
}

func (eh *envHandlerStruct) setCurrentEnvType(i int) {
	eh.curEnvType = i
}

func (eh *envHandlerStruct) setCurRemoteName(remoteName string) error {
	if remoteName != "" {
		if _, ok := eh.remotes[remoteName]; !ok {
			return errors.New("remote with that name doesn't exist.")
		}
	}
	eh.curRemoteName = remoteName
	return nil
}

func (eh *envHandlerStruct) isRemoteAddrExist(addr string) bool {
	for _, r := range eh.remotes {
		if r.dialer().addr == addr {
			return true
		}
	}
	return false
}

func (eh *envHandlerStruct) isRemoteNameExist(remoteName string) bool {
	_, exist := eh.remotes[remoteName]
	return exist
}

func errTooFewArgsFor(arg string, s ...string) error {
	var concat string
	if len(s) > 0 {
		concat = s[0]
	}
	return errors.New("too few arguements for '" + arg + "'." + concat)
}

func (eh *envHandlerStruct) handleCmd(cmdArgs []string) error {
	if len(cmdArgs) == 1 {
		return errors.New(`net usage:
add     add a new network
remove  remove network(s)
ls      list all networks
switch  switch network
`)
	}
	cmd := cmdArgs[1]
	cmdArgs = cmdArgs[2:]

	switch cmd {
	case "add":
		if len(cmdArgs) < 2 {
			return errTooFewArgsFor(cmd, `
usage: net add <address> <remote-name>
`)
		} else if len(cmdArgs) > 2 {
			return errors.New(`too many arguements for 'add'.
usage: net add <address> <remote-name>

(no spaces allowed in remote-name)
`)
		}
		h, p, err := net.SplitHostPort(cmdArgs[0])
		if err != nil {
			return err
		}
		if net.ParseIP(h) == nil {
			return errors.New("invalid ip host.")
		}
		if _, err := strconv.Atoi(p); err != nil {
			return errors.New("invalid port.")
		}
		remoteName := cmdArgs[1]
		if eh.isRemoteNameExist(remoteName) || eh.isRemoteAddrExist(cmdArgs[0]) {
			return errors.New("remote already exists.")
		}
		rEnv := newRemoteEnv(newDialer(cmdArgs[0]), remoteName)
		eh.addRemoteEnv(remoteName, rEnv)
		eh.saveRemotesToGobFile()

		if err = rEnv.initRemote(true); err != nil {
			return err
		}
		eh.handleCmd([]string{"net", "switch", remoteName})
		return nil
	case "ls":
		fmt.Print("Remotes ", len(eh.remotes), "\n\n")
		for _, r := range eh.remotes {
			fmt.Printf("%s => %s\n", r.getRemoteName(), r.dialer().addr)
		}
		return nil
	case "switch":
		if len(cmdArgs) < 1 {
			return errTooFewArgsFor(cmd)
		}
		envName := cmdArgs[0]
		if envName == "local" {
			eh.setCurrentEnvType(LOCAL)
			return eh.setCurRemoteName("")
		}
		if err := eh.setCurRemoteName(envName); err != nil {
			return err
		}
		eh.setCurrentEnvType(REMOTE)
		return nil
	case "remove":
		if len(cmdArgs) < 0 {
			return errTooFewArgsFor(cmd)
		}
		for _, rn := range cmdArgs {
			if eh.isRemoteNameExist(rn) {
				if eh.curRemoteName == rn {
					eh.curRemoteName = ""
					eh.curEnvType = LOCAL
				}
				delete(eh.remotes, rn)
			} else {
				return errors.New("unable to find remote '" + rn + "'.")
			}
		}
		return nil
	case "rename":
		if len(cmdArgs) != 2 {
			return errors.New(`need exactly two arguements for 'remove'
usage: net rename <remote-name> <new-name>
`)
		}
		if !eh.isRemoteNameExist(cmdArgs[0]) {
			return errors.New("there is no remote named '" + cmdArgs[0] + "'.")
		}
		if eh.isRemoteNameExist(cmdArgs[1]) {
			return errors.New("remote already exists.")
		}
		eh.remotes[cmdArgs[1]] = eh.remotes[cmdArgs[0]]
		eh.remotes[cmdArgs[1]].setRemoteName(cmdArgs[1])
		delete(eh.remotes, cmdArgs[0])
		if eh.curRemoteName == cmdArgs[0] {
			eh.setCurRemoteName(cmdArgs[1])
		}
		return nil
	}
	return errors.New("command '" + cmd + "' not found.")
}
