package httpServer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"ftp/common"
	"ftp/config"
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

func initDir(c config.ConfigHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p struct {
			InitDir string `json:"initDir"`
		}
		p.InitDir = c.GetInitDir()
		json.NewEncoder(w).Encode(p)
	})
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
	} else if !common.IsPathExists(path.Path) {
		respondErrMsg("no path provided", w)
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

func uploadWithLog(c config.ConfigHandler, logger common.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		if err := r.ParseMultipartForm(0); err != nil {
			respondErrMsg(err.Error(), w)
			return
		}

		toSavePath := r.Header.Get("path")

		if toSavePath == "" {
			respondErrMsg("no path provided", w)
		}

		var uploadedPaths []string

		for k := range r.MultipartForm.File {
			f, h, err := r.FormFile(k)
			if err != nil {
				continue
			}
			p := filepath.Join(toSavePath, h.Filename)
			if common.IsPathExists(p) {
				continue
			}
			p = filepath.ToSlash(p)
			log.Printf("upload: %s\n", p)
			uploadedPaths = append(uploadedPaths, p)

			downloadFile, err := os.Create(p)
			if err != nil {
				continue
			}
			io.Copy(downloadFile, f)
			downloadFile.Close()
			f.Close()
		}
		js, _ := json.MarshalIndent(uploadedPaths, "", "\t")
		logger.Log("upload on", time.Now(), string(js))
		respondSuccess(w)
	}
}

func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func getMultipleFilesWithLog(c config.ConfigHandler, logger common.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		var paths []string
		if json.Unmarshal(body, &paths) == nil {
			for i, p := range paths {
				p, _ = common.ToAbsToSlash(p)
				if c.IsRestricted(p) {
					paths = remove(paths, i)
				}
			}
			b, _ := json.MarshalIndent(paths, "", "\t")
			logger.Log("get on", time.Now(), string(b))

			tempDir, _ := os.MkdirTemp("", "ftp-go-http")
			defer os.RemoveAll(tempDir)
			fileName := "download"
			zipPath := filepath.Join(tempDir, fileName+".zip")

			log.Println("zipping: ", paths)

			if err := common.ZipSource(paths, zipPath, nil); err != nil {
				fmt.Println(err.Error())
				return
			}
			http.ServeFile(w, r, zipPath)
		}
		respondErrMsg("something went wrong", w)
	}
}

func verifyPath(c config.ConfigHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				respondErrMsg(err.Error(), w)
				return
			}
			r.Body.Close()
			var p struct {
				Path string `json:"path"`
			}

			if path := r.Header.Get("path"); path != "" {
				p.Path = path
			} else {
				err = json.Unmarshal(body, &p)
				if err != nil {
					respondErrMsg(err.Error(), w)
					return
				}
			}
			if c.IsRestricted(p.Path) {
				respondErrMsg("err: Permission denied.", w)
			} else {
				r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
				next.ServeHTTP(w, r)
			}
		})
	}
}

func StartHttpServer(PORT string) {
	c, err := config.ParseConfigFile("ftp-config.json")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	curDate := time.Now().Local().Format("01-02-2006")
	logger, err := common.NewLoggerWithDirAndFileName("./logs/http", curDate+".log")

	if err != nil {
		log.Fatal(err.Error())
	}
	r := mux.NewRouter()
	r.Handle("/init-dir", initDir(c))
	r.HandleFunc("/get-files", getMultipleFilesWithLog(c, logger)).Methods(http.MethodPost)

	m := r.NewRoute().Subrouter()
	m.Use(verifyPath(c))
	m.HandleFunc("/ls", ls).Methods(http.MethodPost)
	m.HandleFunc("/path-exists", pathExists).Methods(http.MethodPost)
	m.HandleFunc("/upload", uploadWithLog(c, logger)).Methods(http.MethodPost)

	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./server/http-server/public/"))))
	printNetworks(":" + PORT)

	if err := http.ListenAndServe(":"+PORT, r); err != nil {
		log.Fatal("Error Starting the HTTP Server :", err)
		return
	}
}
