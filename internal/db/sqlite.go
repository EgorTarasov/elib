package db

type DocumentRepository interface {
	InsertDocument(doc Document) error
}
