package models

import "time"

// User represents a user in the system
type User struct {
	ID           int       `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Subject represents a top-level category (e.g., "now", "learn", "work")
// Previously these were file names
type Subject struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Topic represents a sub-category within a Subject (e.g., "Meta", "Learn", "Python")
// Previously these were the ### headings within files
type Topic struct {
	ID        int       `json:"id" db:"id"`
	SubjectID int       `json:"subject_id" db:"subject_id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Link represents a bookmarked link
type Link struct {
	ID        int       `json:"id" db:"id"`
	TopicID   int       `json:"topic_id" db:"topic_id"`
	Title     string    `json:"title" db:"title"`
	URL       string    `json:"url" db:"url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Tag represents a tag that can be applied to links
type Tag struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

// LinkTag represents the many-to-many relationship between links and tags
type LinkTag struct {
	LinkID int `json:"link_id" db:"link_id"`
	TagID  int `json:"tag_id" db:"tag_id"`
}

// Legacy structures for backward compatibility with current API
type LegacySubject struct {
	Subject string       `json:"subject"`
	Links   []LegacyLink `json:"links"`
}

type LegacyLink struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Timestamp time.Time `json:"timestamp"`
}

type Response struct {
	Data []LegacySubject `json:"data"`
}

type DeleteLink struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type DeleteLinkRequest struct {
	Filename string       `json:"filename"`
	Links    []DeleteLink `json:"links"`
}

type BulkLink struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type BulkLinksRequest struct {
	Filename string     `json:"filename"`
	Subject  string     `json:"subject"`
	Links    []BulkLink `json:"links"`
}

// Auth request/response types
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
