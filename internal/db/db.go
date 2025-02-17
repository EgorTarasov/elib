package db

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type SqliteDb struct {
	db *sql.DB
	mu *sync.Mutex
}

func NewDb() *SqliteDb {
	db, err := sql.Open("sqlite3", "./test.sqlite")
	if err != nil {
		panic(err)
	}
	return &SqliteDb{db: db, mu: &sync.Mutex{}}
}

func (d *SqliteDb) Close() {
	d.db.Close()
}

func (d *SqliteDb) CreateTables() {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS documents (
			id INTEGER PRIMARY KEY,
			name TEXT,
			page_cnt INTEGER,
			folder_id INTEGER,
			foreign key (folder_id) references folders(id)
		);
		CREATE TABLE IF NOT EXISTS folders (
			id INTEGER PRIMARY KEY,
			name TEXT,
			parent_id INTEGER,
			foreign key (parent_id) references folders(id)
		);
	`)
	if err != nil {
		panic(err)
	}
}

func (d *SqliteDb) InsertDocument(id int, name string, pageCnt int, folderId int) {
	d.mu.Lock()
	fmt.Println("Inserting document")
	defer d.mu.Unlock()
	_, err := d.db.Exec("INSERT INTO documents (id, name, page_cnt, folder_id) VALUES (?, ?, ?, ?)", id, name, pageCnt, folderId)
	if err != nil {
		panic(err)
	}
}

func (d *SqliteDb) InsertFolder(name string, parentId int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := d.db.Exec("INSERT INTO folders (name, parent_id) VALUES (?, ?)", name, parentId)
	if err != nil {
		panic(err)
	}
}

func (d *SqliteDb) GetFolderById(name string, parentId int) int {
	d.mu.Lock()
	defer d.mu.Unlock()
	var id int
	err := d.db.QueryRow("SELECT id FROM folders WHERE name = ? AND parent_id = ?", name, parentId).Scan(&id)
	if err != nil {
		panic(err)
	}
	return id
}
