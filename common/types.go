package common

type FileStruct struct {
	Name  string
	IsDir bool
	Size  int64
}

type DirName string

type Res struct {
	Err  bool
	Data string
}

type ZipProgress struct {
	Max     int64
	Current int64
	IsDone  bool
}
