function Dir() {
  this.dirStr = ""

  this.set = newDir => {
    this.dirStr = newDir
  }

}

const vueApp = new Vue({
  el: '#app',
  data: { 
    curDir: "",
    files: [],
    dirType: "",
    errMsg: ""
  },
  methods: {
    async getFilesFromServer(path) {
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
        return this.errMsg = data.errMsg
      }

      if(res.ok) {
        this.files = data.files
        this.curDir = path
        if(this.curDir === "" && this.dirType === "unix") this.curDir = "/"
      }
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