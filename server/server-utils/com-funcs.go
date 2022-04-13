package serverUtils

import (
	"ftp/common"
	"io/ioutil"
	"path/filepath"
)

func GetFileList(dirName string) ([]common.FileStruct, error) {
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		return nil, err
	}
	var fileList []common.FileStruct
	for _, f := range files {
		fileStruc := common.FileStruct{Name: f.Name(), IsDir: f.IsDir(), Size: f.Size()}
		fileList = append(fileList, fileStruc)
	}
	return fileList, nil
}

func GetAbsPath(relPath string) (string, error) {
	fp, err := filepath.Abs("./")
	return filepath.ToSlash(fp), err
}
