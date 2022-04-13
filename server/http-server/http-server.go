package httpServer

import (
	"encoding/json"
	"ftp/common"
	serverUtils "ftp/server/server-utils"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

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

func StartHttpServer(PORT string) {
	r := mux.NewRouter()

	r.HandleFunc("/pwd", curWorkingDir)
	r.HandleFunc("/ls", ls).Methods("POST")
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./server/http-server/public/"))))
	log.Println("running http server: http://localhost:" + PORT)

	if err := http.ListenAndServe(":"+PORT, r); err != nil {
		log.Fatal("Error Starting the HTTP Server :", err)
		return
	}
}
