package client

import (
	"ftp/common"
	"strings"
)

type dirFiles []common.FileStruct

func (df *dirFiles) nameSlice() (fileNames []string) {
	for _, f := range *df {
		if strings.Contains(f.Name, " ") {
			fileNames = append(fileNames, "\""+f.Name+"\"")
		} else {
			fileNames = append(fileNames, f.Name)
		}
	}
	return
}

func (df *dirFiles) ListFunc() func(string) []string {
	return func(s string) []string {
		return df.nameSlice()
	}
}

func deleteEmptyStr(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
