package httpServer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"ftp/common"
	serverUtils "ftp/server/server-utils"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
)

type ErrStruct struct {
	Err    bool   `json:"err"`
	ErrMsg string `json:"errMsg"`
}

func curWorkingDir(w http.ResponseWriter, r *http.Request) {
	if dirPath, err := serverUtils.GetAbsPath("./"); err == nil {
		w.Write([]byte(filepath.ToSlash(dirPath)))
	}
}

func ls(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	reqBody, _ := ioutil.ReadAll(r.Body)

	var path struct {
		Path string `json:"path"`
	}

	err := json.Unmarshal(reqBody, &path)

	var filesResponse struct {
		ErrStruct
		Files []common.FileStruct `json:"files"`
	}

	if err != nil {
		filesResponse.Err = true
		filesResponse.ErrMsg = err.Error()
	} else {
		filesResponse.Files, err = serverUtils.GetFileList(path.Path)
		if err != nil {
			filesResponse.Err = true
			filesResponse.ErrMsg = err.Error()
		}
	}
	json.NewEncoder(w).Encode(filesResponse)
}

func pathExists(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	reqBody, _ := ioutil.ReadAll(r.Body)

	var path struct {
		Path string `json:"path"`
	}

	err := json.Unmarshal(reqBody, &path)

	var filesResponse struct {
		ErrStruct
		PathExists bool `json:"pathExists"`
	}

	if err != nil {
		filesResponse.Err = true
		filesResponse.ErrMsg = err.Error()
	} else {
		filesResponse.PathExists = common.PathExists(path.Path)
	}
	json.NewEncoder(w).Encode(filesResponse)
}

func getMultipleFiles(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)

	var paths []string

	err := json.Unmarshal(body, &paths)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	tempDir, _ := os.MkdirTemp("", "ftp-go-http")
	defer os.RemoveAll(tempDir)
	fileName := "download"
	zipPath := filepath.Join(tempDir, fileName+".zip")

	if err := common.ZipSource(paths, zipPath, nil); err != nil {
		fmt.Println(err.Error())
		return
	}
	http.ServeFile(w, r, zipPath)
}

func printNetworks(port string) {
	fmt.Println("running on:")
	fmt.Println("    http://localhost" + port)
	if ipv4 := common.GetIPv4Str(); ipv4 != common.LOCAL_HOST {
		fmt.Println("    http://" + ipv4 + port)
	}
}

func LogGetFiles(logger common.Logger, logDescr string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			body, _ := ioutil.ReadAll(r.Body)
			r.Body.Close()
			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			next.ServeHTTP(w, r)

			var paths []string
			if json.Unmarshal(body, &paths) == nil {
				b, _ := json.MarshalIndent(paths, "", "\t")
				logger.Log(logDescr, time.Now(), string(b))
			}
		}
		return http.HandlerFunc(fn)
	}
}

func StartHttpServer(PORT string) {
	curDate := time.Now().Local().Format("01-02-2006")
	logger, err := common.NewLoggerWithDirAndFileName("./logs/http", curDate+".log")

	if err != nil {
		log.Fatal(err.Error())
	}
	r := mux.NewRouter()

	r.HandleFunc("/pwd", curWorkingDir)
	r.HandleFunc("/ls", ls).Methods(http.MethodPost)
	r.HandleFunc("/path-exists", pathExists).Methods(http.MethodPost)
	fr := r.NewRoute().Subrouter()
	fr.Use(LogGetFiles(logger, "get on "))
	fr.HandleFunc("/get-files", getMultipleFiles).Methods(http.MethodPost)
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./server/http-server/public/"))))
	printNetworks(":" + PORT)

	if err := http.ListenAndServe(":"+PORT, r); err != nil {
		log.Fatal("Error Starting the HTTP Server :", err)
		return
	}
}
