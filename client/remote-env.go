package client

import (
	"errors"
	"fmt"
	"ftp/common"
	"io"
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
	initRemote(bool) error

	getRemoteName() string
	setRemoteName(string)

	get([]string, string)
}

func newRemoteEnv(dlr dialer, remoteName string, netListFunc func(string) []string) remoteEnv {
	es := &envStruct{curDirFiles: make(dirFiles, 0)}
	dirListFunc := es.curDirFiles.ListFunc()

	completer := readline.NewPrefixCompleter(
		readline.PcItem("cd", readline.PcItemDynamic(dirListFunc)),
		readline.PcItem("get", readline.PcItemDynamic(dirListFunc)),
		readline.PcItem("net switch", readline.PcItemDynamic(netListFunc)),
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

func (re *remoteEnvStruct) initRemote(logOutput bool) error {
	if logOutput {
		fmt.Println("getting current working dir...")
	}

	err := re.fetchCurDirFromServer()
	if err != nil {
		return errors.New(err.Error() + "\nunable to get working directory from server.\n")
	}

	if logOutput {
		fmt.Println("fetching file names...")
	}

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

	files, err := common.DecodeWithRes[[]common.FileStruct](gh)
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
	d, err := common.DecodeWithRes[string](gh)

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

	exists, err := common.DecodeWithRes[common.Res](gh)
	if err != nil {
		return err
	}

	if exists.Err {
		return exists
	}

	re.curDir = exists.Msg
	return nil
}

func (re *remoteEnvStruct) ls() error {
	err := re.fetchCurDirFilesFromServer()

	if err != nil {
		return errors.New(err.Error())
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

	paths := cmdArgs[1:]

	for i, p := range paths {
		if p[0] == '"' && p[len(paths[i])-1] == '"' {
			paths[i] = p[1 : len(paths[i])-1]
		}
		paths[i] = re.curDir + "/" + paths[i]
	}

	if err := gh.Encode(paths); err != nil {
		fmt.Println(err.Error())
		return
	}

	isZipping, err := common.DecodeWithRes[string](gh)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(isZipping)

	max, err := common.DecodeWithRes[int64](gh)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	zipPb := common.NewProgressBar(max, 30, "█", "zipping file: ")
	for {
		zp, err := common.Decode[int64](gh)
		if err != nil {
			break
		}
		if zipPb.Max() > 0 {
			zipPb.AddCurrent(zp)
			if zp == -1 {
				zipPb.UpdateCurrent(zipPb.Max())
				break
			}
		}
	}
	fmt.Print("\n\n")

	fileDetails, err := common.DecodeWithRes[common.FileStruct](gh)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	tmpDir, err := os.MkdirTemp("", "ftp-go-client")
	file, err := os.Create(tmpDir + "/" + fileDetails.Name)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer func() {
		file.Close()
		os.RemoveAll(tmpDir)
	}()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 256)
	pb := common.NewProgressBar(fileDetails.Size, 30, "█", "getting file: ")
	for {
		n, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				fmt.Println("read error:", err)
			}
			break
		}

		pb.AddCurrent(int64(n))

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
	common.UnzipSource(tmpDir+"/"+fileDetails.Name, ddir)
	fmt.Printf("got file '%s' successfully\n", cmdArgs[1])
}
