function Dir(div) {
  this.div = div
  this.dirStr = ""

  this.set = newDir => {
    this.dirStr = newDir
    this.div.innerText = newDir
  }

  this.getDirFromServer = async (url, opts={ method: "GET", headers: { "Content-Type": "text/json" } }) => {
    const res = await fetch(url, opts)
    const data = await res.text()
    if(res.ok) {
      this.dirStr = data
      this.div.innerText = data
    }
  }
}

function File() {
  this.name = ""
  this.isDir = false
  this.size = 0
}

function ContentHandler(div) {
  this.div = div
  this.contents = []

  this.set = newContent => this.contents = newContent
}

function EnvHandler(div) {
  this.dir = new Dir(div)
  this.content = new ContentHandler()

  this.getContentFromServer = async (url, opts={ method: "GET", headers: { "Content-Type": "text/json" }, body: {} }) => {
    opts.body = JSON.stringify(opts.body)
    const res = await fetch(url, opts)
    const data = await res.text()
    if(res.ok)
      this.content.set(data)
  }

  this.init = async () => {
    await this.dir.getDirFromServer("/pwd")
    await this.getContentFromServer("/ls", { method: "POST", body: { path: this.dir.dirStr } })
  }
}

let env = new EnvHandler(document.getElementById("cur-working-dir"))

env.init()