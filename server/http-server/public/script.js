const abortCtlr = { controller: null, isActive: false };

const vueApp = new Vue({
  el: '#app',
  data: { 
    curDir: "",
    files: [],
    dirType: "",
    errMsg: "",
    loading: false
  },
  methods: {
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
        const res = await fetch("/ls", {
          headers: { "Content-Type": "text/json" },
          method: "POST",
          body: JSON.stringify({ path: this.curDir + "/" }),
          signal: abortCtlr.controller.signal
        })
        const data = await res.json()
        
        if(data && data.err) {
          this.errMsg = data.errMsg
        } else if(res.ok) {
          this.files = (data.files === null) ? [] : data.files
        } else this.errMsg = "something went wrong."
        abortCtlr.controller = null;
        abortCtlr.isActive = false;
        this.loading = false
      } catch(err) {
        console.log(err)
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
    async getDirFromServer() {
      const res = await fetch("/pwd", { method: "GET", headers: { "Content-Type": "text/json" } })
      const data = await res.text()
      if(res.ok) {
        this.curDir = data
      }
    },
    async dirChange() {
      const res = await fetch("/path-exists", {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({ path: this.curDir })
      })
      const data = await res.json()
      
      if(data && data.err)
        this.errMsg = data.errMsg
      else if(!data.pathExists)
        this.errMsg = "the specified path doesn't exist."
      else
        await this.getFilesFromServer()
    },
    async resetData() {
      await this.getDirFromServer()
      this.getFilesFromServer()
    }
  },
  async mounted() {
    await this.getDirFromServer()
    await this.getFilesFromServer(this.curDir)
    if(this.curDir !== "") {
      if(this.curDir[0] === "/") this.dirType = "unix"
      else this.dirType = "windows"
    }
  }
})