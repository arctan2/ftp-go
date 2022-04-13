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
    files: []
  },
  methods: {
    async getFilesFromServer() {
      const res = await fetch("/ls", {
        headers: { "Content-Type": "text/json" },
        method: "POST",
        body: JSON.stringify({ path: this.curDir + "/" })
      })
      const data = await res.json()
      
      if(data && data.err) return console.log(data.errMsg)

      if(res.ok)
        this.files = data.files
    },
    async cd(dirName) {
      if(dirName === "..") {
        let lastIdx = this.curDir.lastIndexOf("/")
        if(lastIdx !== -1)
          this.curDir = this.curDir.slice(0, lastIdx)
      }
      else 
        this.curDir = this.curDir + "/" + dirName
      this.getFilesFromServer()
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
    await this.getFilesFromServer()
  }
})