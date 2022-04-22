package common

type FileStruct struct {
	Name  string `json:"name"`
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size"`
}

type DirName string

type Res struct {
	Err bool
	Msg string
}

type ZipProgress struct {
	Max     int64
	Current int64
	IsDone  bool
}
