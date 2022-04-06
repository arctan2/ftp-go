package client

import (
	"errors"
	"fmt"
	"ftp/common"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type remoteEnvStruct struct {
	envStruct
	downloadDir string
	dlr         dialer
}

type remoteEnv interface {
	env
	cmds

	getDownloadDir() string
	setDownloadDir([]string) error

	fetchCurDirFromServer() error
	fetchCurDirFilesFromServer() error
	dialer() *dialer

	get([]string)
}

func newRemoteEnv(downloadDir string, dlr dialer) remoteEnv {
	var e remoteEnv = &remoteEnvStruct{dlr: dlr, downloadDir: downloadDir}
	return e
}

func (re *remoteEnvStruct) dialer() *dialer {
	return &re.dlr
}

func (re *remoteEnvStruct) getDownloadDir() string {
	return re.downloadDir
}

func (re *remoteEnvStruct) setDownloadDir(cmdArgs []string) error {
	if len(cmdArgs) > 1 && cmdArgs[1] != "" {
		arg1 := cmdArgs[1]
		switch arg1 {
		case "-d", "--default":
			abs, _ := filepath.Abs("./downloads")
			re.downloadDir = filepath.ToSlash(abs)
			return nil
		case "-s", "--set":
			re.downloadDir = re.curDir
			return nil
		}
		if len(arg1) > 2 {
			if arg1[0] == '"' {
				arg1 = arg1[1:]
			}
			if arg1[len(arg1)-1] == '"' {
				arg1 = arg1[0 : len(arg1)-1]
			}
		}

		if _, err := os.Stat(arg1); err != nil {
			return err
		} else {
			ddir, err := filepath.Abs(arg1)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}
			re.downloadDir = filepath.ToSlash(ddir)
		}
	}
	return nil
}

func (re *remoteEnvStruct) fetchCurDirFilesFromServer() error {
	conn, err := re.dlr.DialAndCmd("ls")
	if err != nil {
		return err
	}
	defer conn.Close()

	gh := common.NewGobHandler(conn, conn)
	if err := gh.Encode(re.curDir); err != nil {
		return err
	}

	files, err := common.Decode[[]common.FileStruct](gh)
	if err != nil {
		return err
	}
	re.curDirFiles = dirFiles(files)
	return nil
}

func (re *remoteEnvStruct) fetchCurDirFromServer() error {
	conn, err := re.dlr.DialAndCmd("pwd")

	if err != nil {
		return err
	}
	defer conn.Close()

	gh := common.NewGobHandler(conn, conn)
	d, err := common.Decode[common.DirName](gh)

	if err != nil {
		return err
	}

	re.curDir = strings.TrimSpace(string(d))
	return nil
}

func (re *remoteEnvStruct) cd(cmdArgs []string) error {
	if len(cmdArgs) == 1 {
		return errors.New("err: missing operand after cd")
	}

	conn, err := re.dialer().DialAndCmd("cd")
	if err != nil {
		return err
	}
	defer conn.Close()

	gh := common.NewGobHandler(conn, conn)

	cdDirName := cmdArgs[1]
	if cdDirName != "" && cdDirName[0] == '"' && cdDirName[len(cdDirName)-1] == '"' {
		cdDirName = cdDirName[1 : len(cdDirName)-1]
	}

	if filepath.IsAbs(cdDirName) || cdDirName[0] == '/' {
		if err := gh.Encode(cdDirName); err != nil {
			return err
		}
	} else {
		if err := gh.Encode(re.curDir + "/" + cdDirName); err != nil {
			return err
		}
	}

	exists, err := common.Decode[common.Res](gh)
	if err != nil {
		return err
	}

	if exists.Err {
		return errors.New(exists.Error())
	}

	re.curDir = exists.Data
	return nil
}

func (re *remoteEnvStruct) ls() error {
	err := re.fetchCurDirFilesFromServer()

	if err != nil {
		return errors.New(err.Error() + "\nunable to get files.")
	}

	for _, f := range *re.getCurDirFiles() {
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
	return nil
}

func (re *remoteEnvStruct) get(cmdArgs []string) {
	if len(cmdArgs) == 1 {
		fmt.Println("err: missing operand after get")
		return
	}

	conn, err := re.dialer().DialAndCmd("get")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer conn.Close()

	gh := common.NewGobHandler(conn, conn)

	if err := gh.Encode(re.getCurDir() + "/" + cmdArgs[1]); err != nil {
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

	ddir := re.getDownloadDir()
	if _, err := os.Stat(ddir); err != nil {
		err := os.Mkdir(ddir, os.ModePerm)
		if err != nil {
			fmt.Println("couldn't create downloads folder.")
			return
		}
	}

	fmt.Println("extracting...")
	common.UnzipSource("./.tmp-client/"+fileDetails.Name, ddir)
	fmt.Printf("got file '%s' successfully\n", cmdArgs[1])
}
