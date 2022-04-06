package client

import "ftp/common"

type dirFiles []common.FileStruct

func (df dirFiles) nameSlice() (fileNames []string) {
	for _, f := range df {
		fileNames = append(fileNames, f.Name)
	}
	return
}

func (df *dirFiles) ListFunc() func(string) []string {
	return func(line string) []string {
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
