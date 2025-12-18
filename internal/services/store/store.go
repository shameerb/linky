package store

import "markdown-editor/internal/models"

// Legacy store interface for backward compatibility
type Store interface {
	ListFiles() ([]string, error)
	GetLinks(filename string) ([]models.LegacySubject, error)
	AddBulkLinks(filename string, subject string, links []models.LegacyLink) error
	DeleteLinks(filename string, links []models.LegacyLink) error
	CreateFile(filename string) error
	DeleteFile(filename string) error
}

// New store interface for multi-user support
type UserStore interface {
	// User management
	CreateUser(email, passwordHash string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id int) (*models.User, error)

	// Subject management
	ListSubjects(userID int) ([]models.Subject, error)
	CreateSubject(userID int, name string) (*models.Subject, error)
	GetSubject(id int, userID int) (*models.Subject, error)
	DeleteSubject(id int, userID int) error

	// Topic management
	ListTopics(subjectID int, userID int) ([]models.Topic, error)
	CreateTopic(subjectID int, name string, userID int) (*models.Topic, error)
	GetTopic(id int, userID int) (*models.Topic, error)
	DeleteTopic(id int, userID int) error

	// Link management
	ListLinks(topicID int, userID int) ([]models.Link, error)
	CreateLink(topicID int, title, url string, userID int) (*models.Link, error)
	GetLink(id int, userID int) (*models.Link, error)
	DeleteLink(id int, userID int) error
	BulkCreateLinks(topicID int, links []models.Link, userID int) error

	// Tag management
	ListTags(userID int) ([]models.Tag, error)
	CreateTag(name string, userID int) (*models.Tag, error)
	AddTagToLink(linkID, tagID int, userID int) error
	RemoveTagFromLink(linkID, tagID int, userID int) error
	GetLinkTags(linkID int, userID int) ([]models.Tag, error)
}
