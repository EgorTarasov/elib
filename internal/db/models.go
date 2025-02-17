package db

type Document struct {
	Id       int
	Name     string
	PageCnt  int
	FolderId int
}

type Folder struct {
	Id   int
	Name string
}

