const abortCtlr = { controller: null, isActive: false };

function downloadFiles(filePathList) {
  fetch("/get-files", { body: JSON.stringify(filePathList), method: "POST" })
    .then(res => res.blob())
    .then(blob => {
      const fileURL = window.URL.createObjectURL(blob);
      const fileLink = document.createElement('a');
      fileLink.style.display = "none";
      fileLink.href = fileURL;
      fileLink.download = "download.zip";
      document.body.appendChild(fileLink)
      fileLink.click();
      document.body.removeChild(fileLink)
    })
}

const vueApp = new Vue({
  el: '#app',
  data: { 
    curDir: "",
    files: [],
    dirType: "",
    errMsg: "",
    loading: false,
    isSelectMode: false,
    selected: new Set(),
    displayUploadProgressBar: false
  },
  methods: {
    async upload(files) {
      const formData = new FormData()
      for(const f of files)
        formData.append(f.name, f)
      this.displayUploadProgressBar = true;
      const res = await axios.post("/upload", formData, {
        headers: { "path": this.curDir + "/" },
        onUploadProgress: progressEvent => {
          let percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          this.$refs.uploadProgressBar.style.left = percentCompleted + "%";
        }
      })
      const data = res.data
      if(!data?.err)
        this.getFilesFromServer()
      this.displayUploadProgressBar = false;
    },
    filesChange(e) {
      this.upload([...e.target.files])
      e.target.type = ""
      e.target.type = "file"
    },
    async getFilesFromServer() {
      if(abortCtlr.isActive) abortCtlr.controller?.abort();
      this.loading = true
      if(this.errMsg !== "") {
        this.errMsg = ""
        this.files = []
      }

      abortCtlr.controller = new AbortController();
      abortCtlr.isActive = true;

      try {
        const res = await axios.post("/ls", { path: this.curDir === "" ? "" : this.curDir + "/" }, {
          headers: { "Content-Type": "text/json" },
          signal: abortCtlr.controller.signal
        })
        const data = res.data
        if(data && data.err) {
          this.errMsg = data.msg
        } else if(res.status === 200) {
          this.files = (data.files === null) ? [] : data.files
        } else this.errMsg = "something went wrong."
        abortCtlr.controller = null;
        abortCtlr.isActive = false;
        this.loading = false
      } catch(err) {
        this.errMsg = "something went wrong";
      }
    },
    async cd(dirName) {
      if(dirName === "..") {
        let lastIdx = this.curDir.lastIndexOf("/")
        if(lastIdx !== -1) {
          this.curDir = this.curDir.slice(0, lastIdx)
          if(this.curDir === "" && this.dirType === "unix") this.curDir = "/"
          await this.getFilesFromServer()
        }
      }
      else {
        this.curDir = this.curDir + "/" + dirName
        if(this.curDir === "" && this.dirType === "unix") this.curDir = "/"
        await this.getFilesFromServer()
      }
    },
    async getInitDirFromServer() {
      try {
        const res = await axios.get("/init-dir", {})
        if(res.data.err && res.data.msg) {
          this.errMsg = res.data.msg
          return
        }
        this.curDir = res.data.initDir
      } catch(err) {
        console.log(err)
      }
    },
    async dirChange() {
      const res = await axios.post("/path-exists", { path: this.curDir }, {
        headers: {
          "Content-Type": "application/json"
        },
      })
      const data = await res.data;
      
      if(data && data.err)
        this.errMsg = data.msg
      else if(!data.pathExists)
        this.errMsg = "the specified path doesn't exist."
      else
        await this.getFilesFromServer()
    },
    async resetData() {
      await this.getInitDirFromServer()
      this.getFilesFromServer()
    },
    handleFileClick(fName, isDir) {
      if(this.isSelectMode) {
        if(this.selected.has(fName)) this.selected.delete(fName)
        else this.selected.add(fName)
        this.$forceUpdate()
      }
      else isDir && this.cd(fName)
    },
    async done() {
      if(this.selected.size === 0) return this.isSelectMode = false;
      downloadFiles([...this.selected].map(fileName => this.curDir + "/" + fileName));
      this.selected.clear();
      this.isSelectMode = false;
    }
  },
  async mounted() {
    await this.getInitDirFromServer()
    await this.getFilesFromServer()
    if(this.curDir !== "") {
      if(this.curDir[0] === "/") this.dirType = "unix"
      else this.dirType = "windows"
    }
  }
})