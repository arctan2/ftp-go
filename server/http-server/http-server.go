package httpServer

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

func curWorkingDir(w http.ResponseWriter, r *http.Request) {
	fp, _ := filepath.Abs("./")
	w.Write([]byte(filepath.ToSlash(fp)))
}

func ls(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)

	var post struct {
		Path string `json:"path"`
	}

	err := json.Unmarshal(reqBody, &post)

	if err != nil {
		w.Write([]byte("something went wrong."))
	} else {

	}
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
