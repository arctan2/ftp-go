package common

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

type loggerStruct struct {
	logger *log.Logger
	file   string
}

type Logger interface {
	Log(...interface{})
}

func (l *loggerStruct) Log(i ...interface{}) {
	l.logger.Println(i...)
}

func NewLogger(fp string) (Logger, error) {
	parts := strings.Split(fp, "/")
	fileName := parts[len(parts)-1]
	fp = filepath.Join(parts[0 : len(parts)-1]...)
	return NewLoggerWithDirAndFileName(fp, fileName)
}

func NewLoggerWithDirAndFileName(dirName, fileName string) (Logger, error) {
	fp := filepath.Join(dirName, fileName)
	os.MkdirAll(dirName, os.ModePerm)
	var (
		f   *os.File
		err error
	)
	if PathExists(fp) {
		if f, err = os.OpenFile(fp, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
			return nil, err
		}
	} else {
		f, err = os.Create(fp)
	}
	var l Logger = &loggerStruct{logger: log.New(f, "", log.LstdFlags|log.Lshortfile), file: fp}
	return l, nil
}
