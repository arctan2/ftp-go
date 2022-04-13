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

	"github.com/chzyer/readline"
)

type remoteEnvStruct struct {
	*envStruct
	name string
	dlr  dialer
}

type remoteEnv interface {
	env

	fetchCurDirFromServer() error
	dialer() *dialer
	initRemote() error

	getRemoteName() string
	setRemoteName(string)

	get([]string, string)
}

func newRemoteEnv(dlr dialer, remoteName string) remoteEnv {
	es := &envStruct{curDirFiles: make(dirFiles, 0)}
	dirListFunc := es.curDirFiles.ListFunc()

	completer := readline.NewPrefixCompleter(
		readline.PcItem("cd", readline.PcItemDynamic(dirListFunc)),
		readline.PcItem("get", readline.PcItemDynamic(dirListFunc)),
	)

	rln, _ := readline.NewEx(&readline.Config{
		Prompt:              "> ",
		AutoComplete:        completer,
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		FuncFilterInputRune: filterInput,
	})
	es.rln = rln
	var rEnv = &remoteEnvStruct{dlr: dlr, envStruct: es, name: remoteName}
	return rEnv
}

func (re *remoteEnvStruct) getRemoteName() string {
	return re.name
}
func (re *remoteEnvStruct) setRemoteName(newName string) {
	re.name = newName
}

func (re *remoteEnvStruct) dialer() *dialer {
	return &re.dlr
}

func (re *remoteEnvStruct) initRemote() error {
	fmt.Println("getting current working dir...")

	err := re.fetchCurDirFromServer()
	if err != nil {
		return errors.New(err.Error() + "\nunable to get working directory from server.\n")
	}

	fmt.Println("fetching file names...")

	err = re.fetchCurDirFilesFromServer()
	if err != nil {
		return errors.New(err.Error() + "\nunable to get directory files from server.\n")
	}
	return nil
}

func (re *remoteEnvStruct) refreshCurDirFiles() error {
	return re.fetchCurDirFilesFromServer()
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
	d, err := common.Decode[string](gh)

	if err != nil {
		return err
	}

	re.curDir = strings.TrimSpace(d)
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

	for _, f := range re.curDirFiles {
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

func (re *remoteEnvStruct) get(cmdArgs []string, dest string) {
	if len(cmdArgs) == 1 {
		fmt.Println("err: missing operand after get")
		return
	}

	conn, err := re.dlr.DialAndCmd("get")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer conn.Close()

	gh := common.NewGobHandler(conn, conn)

	if err := gh.Encode(re.curDir + "/" + cmdArgs[1]); err != nil {
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

	ddir := dest
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
