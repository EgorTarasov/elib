package main

import (
	"database/sql"
	"encoding/json"
	"os"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

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

func main() {
	db, err := sql.Open("sqlite3", "./elib.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS documents (
			id INTEGER PRIMARY KEY,
			name TEXT,
			page_cnt INTEGER,
			folder_id INTEGER
		);
		CREATE TABLE IF NOT EXISTS folders (
			id INTEGER PRIMARY KEY,
			name TEXT,
			parent_id INTEGER
		);
	`)

	if err != nil {
		panic(err)
	}

	// read folders directory and read every document into db
	files, err := os.ReadDir("./folders")
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		f, err := os.Open("folders/" + file.Name())
		if err != nil {
			panic(err)
		}
		folderId, err := strconv.Atoi(strings.Split(strings.Split(file.Name(), "_")[1], ".")[0])
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// read folder info from file
		// insert folder into db
		var data ReadInFolderResponse
		// read documents from folder

		err = json.NewDecoder(f).Decode(&data)
		if err != nil {
			panic(err)
		}
		for _, doc := range data.Data.Docs {
			var pageCnt, Id int
			var title string
			for _, value := range doc.Info {
				if value.ID == "32" {
					pageCnt, _ = strconv.Atoi(value.Value)
				}
				if value.ID == "28" {
					title = value.Title
				}
			}

			Id, err = strconv.Atoi(doc.ID)
			if err != nil {
				panic(err)
			}
			_, err := db.Exec("INSERT INTO documents (id, name, page_cnt, folder_id) VALUES (?, ?, ?, ?)", Id, title, pageCnt, folderId)
			if err != nil {
				panic(err)
			}
		}

	}

}
