package serverUtils

import (
	"ftp/common"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
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

func SendFile(fileName string, conn net.Conn) (int64, error) {
	file, err := os.Open(strings.TrimSpace(fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	return io.Copy(conn, file)
}

func GetFileName(fp string) string {
	parts := strings.Split(fp, "/")
	return parts[len(parts)-1]
}
