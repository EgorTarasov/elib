package client

import "bytes"

type Credentials struct {
	StudentId int    `json:"studentId"`
	Name      string `json:"name"`
}

type DocuemntResponse struct {
	ID   string `json:"id"`
	Info []struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		Value           string `json:"value"`
		DocumentFieldID string `json:"document_field_id"`
		Title           string `json:"title"`
	} `json:"info"`
	Filetypes string `json:"filetypes"`
	StatShow  string `json:"stat_show"`
}

type DocumentPage struct {
	bookId int
	number int
	image  bytes.Buffer
}

type Folder struct {
	Id       int
	Name     string
	ParentId int
}

type FolderResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	FullPath string `json:"full_path"`
}

type ReadInFolderResponse struct {
	Error string `json:"error"`
	Data  struct {
		Count   int                `json:"count"`
		Docs    []DocuemntResponse `json:"docs"`
		Folders []struct {
			ID   string `json:"id"`
			Info []struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				FullPath string `json:"full_path"`
			} `json:"info"`
		} `json:"folders"`
		Price   string         `json:"price"`
		KolShow int            `json:"kol_show"`
		Start   int            `json:"start"`
		Folder  FolderResponse `json:"folder"`
		Time    float64        `json:"time"`
	} `json:"data"`
	Login bool `json:"login"`
}
