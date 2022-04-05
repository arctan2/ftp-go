package client

type env struct {
	curDir      string
	curDirFiles dirFiles
}

type localEnv interface {
	getCurDir() string
	setCurDir(string) error

	getCurDirFiles() *dirFiles
	setCurDirFiles(dirFiles)
}

func newLocalEnv() localEnv {
	var e localEnv = &env{}
	return e
}

func (e *env) getCurDir() string {
	return e.curDir
}

func (e *env) setCurDir(d string) error {
	e.curDir = d
	return nil
}

func (e *env) getCurDirFiles() *dirFiles {
	return &e.curDirFiles
}

func (e *env) setCurDirFiles(df dirFiles) {
	e.curDirFiles = df
}
