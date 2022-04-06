package client

import "errors"

type envHandlerStruct struct {
	lEnv         localEnv
	remotes      []remoteEnv
	curRemoteIdx int
	curEnvType   int
}

type envHandler interface {
	addRemoteEnv(...remoteEnv)

	currentEnvType() int
	setCurrentEnvType(int)

	currentRemote() remoteEnv
	currentRemoteIdx() int
	localEnv() localEnv
	setCurRemoteIdx(int) error
}

const (
	LOCAL int = iota
	REMOTE
)

func newEnvHandler(lEnv localEnv) envHandler {
	var eh envHandler = &envHandlerStruct{lEnv: lEnv, remotes: make([]remoteEnv, 0), curEnvType: LOCAL, curRemoteIdx: -1}
	return eh
}

func (eh *envHandlerStruct) addRemoteEnv(rEnv ...remoteEnv) {
	eh.remotes = append(eh.remotes, rEnv...)
}

func (eh *envHandlerStruct) currentEnvType() int {
	return eh.curEnvType
}

func (eh *envHandlerStruct) currentRemote() remoteEnv {
	if len(eh.remotes) == 0 || eh.curEnvType == LOCAL {
		return nil
	}
	return eh.remotes[eh.curRemoteIdx]
}

func (eh *envHandlerStruct) currentRemoteIdx() int {
	return eh.curRemoteIdx
}

func (eh *envHandlerStruct) localEnv() localEnv {
	return eh.lEnv
}

func (eh *envHandlerStruct) setCurrentEnvType(i int) {
	eh.curEnvType = i
}

func (eh *envHandlerStruct) setCurRemoteIdx(idx int) error {
	if idx >= len(eh.remotes) || idx < 0 {
		return errors.New("index out of range.")
	}
	eh.curRemoteIdx = idx
	return nil
}
