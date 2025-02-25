package store

import "markdown-editor/internal/models"

type Store interface {
	ListFiles() ([]string, error)
	GetLinks(filename string) ([]models.Subject, error)
	AddBulkLinks(filename string, subject string, links []models.Link) error
	DeleteLinks(filename string, links []models.Link) error
	CreateFile(filename string) error
	DeleteFile(filename string) error
}
