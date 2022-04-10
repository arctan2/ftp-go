package client

import (
	"errors"
	"fmt"
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

	closeAllRemotesRlns()
}

const (
	LOCAL int = iota
	REMOTE
)

func newEnvHandler() envHandler {
	pwd, _ := os.Getwd()
	var eh envHandler = &envHandlerStruct{
		lEnv:       newLocalEnv(filepath.ToSlash(pwd + "/downloads")),
		remotes:    make(map[string]remoteEnv),
		curEnvType: LOCAL,
	}
	return eh
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

func (eh *envHandlerStruct) isRemoteExist(addr, remoteName string) bool {
	for _, r := range eh.remotes {
		if r.getRemoteName() == remoteName || r.dialer().addr == addr {
			return false
		}
	}
	return false
}

func (eh *envHandlerStruct) handleCmd(cmdArgs []string) error {
	if len(cmdArgs) == 1 {
		return errors.New(`net usage:
add     add a new network
remove  remove a network
ls      list all networks
switch  switch network
`)
	}
	cmd := cmdArgs[1]
	cmdArgs = cmdArgs[2:]

	switch cmd {
	case "add":
		if len(cmdArgs) < 2 {
			return errors.New(`too few arguements for 'add'.
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
		if eh.isRemoteExist(cmdArgs[0], remoteName) {
			return errors.New("remote already exists.")
		}
		rEnv := newRemoteEnv(newDialer(cmdArgs[0]), remoteName)
		eh.addRemoteEnv(remoteName, rEnv)
		err = rEnv.initRemote()
		if err == nil {
			eh.handleCmd([]string{"net", "switch", remoteName})
		}
		return err
	case "ls":
		fmt.Print("Remotes ", len(eh.remotes), "\n\n")
		for _, r := range eh.remotes {
			fmt.Printf("%s => %s\n", r.getRemoteName(), r.dialer().addr)
		}
		return nil
	case "switch":
		if len(cmdArgs) < 1 {
			return errors.New("too few arguements for 'swtich'")
		}
		envName := cmdArgs[0]
		if envName == "local" {
			eh.setCurrentEnvType(LOCAL)
			return eh.setCurRemoteName("")
		}
		eh.setCurrentEnvType(REMOTE)
		return eh.setCurRemoteName(envName)
	case "remove":
		return nil
	}
	return errors.New("command '" + cmd + "' not found.")
}
