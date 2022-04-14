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
    async getFilesFromServer(path) {
      if(this.loading) return
      this.loading = true
      if(this.errMsg !== "") {
        this.errMsg = ""
        this.files = []
      }

      const res = await fetch("/ls", {
        headers: { "Content-Type": "text/json" },
        method: "POST",
        body: JSON.stringify({ path: path + "/" })
      })
      const data = await res.json()
      
      if(data && data.err) {
        this.curDir = path
        this.errMsg = data.errMsg
      } else if(res.ok) {
        this.files = (data.files === null) ? [] : data.files
        this.curDir = path
      } else this.errMsg = "something went wrong."
      if(this.curDir === "" && this.dirType === "unix") this.curDir = "/"
      this.loading = false
    },
    async cd(dirName) {
      if(dirName === "..") {
        let lastIdx = this.curDir.lastIndexOf("/")
        if(lastIdx !== -1)
          await this.getFilesFromServer(this.curDir.slice(0, lastIdx))
      }
      else 
        await this.getFilesFromServer(this.curDir + "/" + dirName)
    },
    async getDirFromServer() {
      const res = await fetch("/pwd", { method: "GET", headers: { "Content-Type": "text/json" } })
      const data = await res.text()
      if(res.ok) {
        this.curDir = data
      }
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