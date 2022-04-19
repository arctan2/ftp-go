package httpServer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"ftp/common"
	serverUtils "ftp/server/server-utils"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
)

type ResponseStruct struct {
	Err bool   `json:"err"`
	Msg string `json:"msg"`
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

	if err != nil {
		respondErrMsg(err.Error(), w)
	} else {
		var res struct {
			Files []common.FileStruct `json:"files"`
		}
		res.Files, err = serverUtils.GetFileList(path.Path)
		if err != nil {
			respondErrMsg(err.Error(), w)
			return
		}
		json.NewEncoder(w).Encode(res)
	}
}

func pathExists(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	reqBody, _ := ioutil.ReadAll(r.Body)

	var path struct {
		Path string `json:"path"`
	}

	err := json.Unmarshal(reqBody, &path)

	if err != nil {
		respondErrMsg(err.Error(), w)
	} else {
		var res struct {
			PathExists bool `json:"pathExists"`
		}
		res.PathExists = common.IsPathExists(path.Path)
		json.NewEncoder(w).Encode(res)
	}
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

func respondErrMsg(msg string, w http.ResponseWriter) {
	res := ResponseStruct{Msg: msg, Err: true}
	json.NewEncoder(w).Encode(res)
}

func respondSuccess(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(ResponseStruct{Err: false, Msg: "success"})
}

func upload(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if err := r.ParseMultipartForm(0); err != nil {
		fmt.Println(err.Error())
	}

	toSavePath := r.FormValue("path")

	if toSavePath == "" {
		respondErrMsg("no path provided", w)
	}

	for k := range r.MultipartForm.File {
		f, h, err := r.FormFile(k)
		if err != nil {
			continue
		}
		p := filepath.Join(toSavePath, h.Filename)
		if common.IsPathExists(p) {
			continue
		}

		downloadFile, err := os.Create(p)
		if err != nil {
			continue
		}
		io.Copy(downloadFile, f)
		downloadFile.Close()
		f.Close()
	}
	respondSuccess(w)
}

func LogGetFilesMiddleware(logger common.Logger, logDescr string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := ioutil.ReadAll(r.Body)
			r.Body.Close()
			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			next.ServeHTTP(w, r)

			var paths []string
			if json.Unmarshal(body, &paths) == nil {
				b, _ := json.MarshalIndent(paths, "", "\t")
				logger.Log(logDescr, time.Now(), string(b))
			}
		})
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
	r.HandleFunc("/upload", upload).Methods(http.MethodPost)
	fr := r.NewRoute().Subrouter()
	fr.Use(LogGetFilesMiddleware(logger, "get on "))
	fr.HandleFunc("/get-files", getMultipleFiles).Methods(http.MethodPost)
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./server/http-server/public/"))))
	printNetworks(":" + PORT)

	if err := http.ListenAndServe(":"+PORT, r); err != nil {
		log.Fatal("Error Starting the HTTP Server :", err)
		return
	}
}
