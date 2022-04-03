package client

import (
	"fmt"
	"ftp/common"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

var blue = color.New(color.FgBlue).PrintfFunc()

func getCurDirFiles(curDir string) (f []common.FileStruct, e error) {
	conn, err := DialAndCmd("ls")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	gh := common.NewGobHandler(conn, conn)
	if err := gh.Encode(curDir); err != nil {
		return nil, err
	}

	return common.Decode[[]common.FileStruct](gh)
}

func cd(cmdArgs []string, curDir string) string {
	if len(cmdArgs) == 1 {
		fmt.Println("err: missing operand after cd")
		return curDir
	}

	conn, err := DialAndCmd("cd")
	if err != nil {
		fmt.Println(err.Error())
		return curDir
	}
	defer conn.Close()

	gh := common.NewGobHandler(conn, conn)

	cdDirName := cmdArgs[1]
	if cdDirName != "" && cdDirName[0] == '"' && cdDirName[len(cdDirName)-1] == '"' {
		cdDirName = cdDirName[1 : len(cdDirName)-1]
	}

	if filepath.IsAbs(cdDirName) {
		gh.Encode(cdDirName)
	} else {
		gh.Encode(curDir + "/" + cdDirName)
	}

	exists, err := common.Decode[common.Res](gh)
	if err != nil {
		fmt.Println(err.Error())
	}

	if exists.Err {
		fmt.Println(exists.Error())
	}

	return exists.Data
}

func ls(curDir string) []common.FileStruct {
	curFiles, err := getCurDirFiles(curDir)

	if err != nil {
		fmt.Println(err.Error(), "\nunable to get files.")
		return nil
	}

	for _, f := range curFiles {
		fName := f.Name
		if strings.ContainsRune(fName, ' ') {
			fName = "\"" + fName + "\""
		}
		if f.IsDir {
			blue("%s  ", fName)
		} else {
			fmt.Printf("%s  ", fName)
		}
	}

	fmt.Println()
	return curFiles
}

func get(curDir string, cmdArgs []string) {
	if len(cmdArgs) == 1 {
		fmt.Println("err: missing operand after get")
		return
	}

	conn, err := DialAndCmd("get")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer conn.Close()

	gh := common.NewGobHandler(conn, conn)

	if err := gh.Encode(curDir + "/" + cmdArgs[1]); err != nil {
		fmt.Println(err.Error())
		return
	}

	isZipping, err := common.Decode[string](gh)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(isZipping)

	zipPb := newProgressBar(0, 30, "█", "zipping file: ")
	for {
		zp, err := common.Decode[common.ZipProgress](gh)
		if err != nil {
			break
		}
		if zp.Max > 0 {
			zipPb.curPercent = int(math.Round(float64(100 * zp.Current / zp.Max)))
			zipPb.filledLength = zipPb.length * zp.Current / zp.Max
			zipPb.print()
			if zp.IsDone {
				zipPb.curPercent = 100
				zipPb.filledLength = zipPb.length
				zipPb.print()
				break
			}
		}
	}
	fmt.Print("\n\n")

	fileDetails, err := common.Decode[common.FileStruct](gh)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	os.MkdirAll("./.tmp-client", os.ModePerm)
	file, err := os.Create("./.tmp-client/" + fileDetails.Name)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer func() {
		file.Close()
		os.RemoveAll("./.tmp-client")
	}()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 256)
	pb := newProgressBar(fileDetails.Size, 30, "█", "getting file: ")
	for {
		n, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				fmt.Println("read error:", err)
			}
			break
		}

		pb.update(int64(n))

		buf = append(buf, tmp[:n]...)
	}
	fmt.Println()

	file.Write(buf)

	if _, err := os.Stat("./downloads"); err != nil {
		err := os.Mkdir("./downloads", os.ModePerm)
		if err != nil {
			fmt.Println("couldn't create downloads folder.")
			return
		}
	}

	fmt.Println("extracting...")
	common.UnzipSource("./.tmp-client/"+fileDetails.Name, "./downloads")
	fmt.Printf("got file '%s' successfully\n", cmdArgs[1])
}
