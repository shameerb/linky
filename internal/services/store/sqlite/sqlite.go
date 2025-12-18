package sqlite

import (
	"database/sql"
	_ "embed"
	"fmt"
	"markdown-editor/internal/models"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

type SQLiteStore struct {
	db *sql.DB
}

// New creates a new SQLite store and initializes the database
func New(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Initialize schema
	if _, err := db.Exec(schemaSQL); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// User management

func (s *SQLiteStore) CreateUser(email, passwordHash string) (*models.User, error) {
	now := time.Now()
	result, err := s.db.Exec(
		"INSERT INTO users (email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?)",
		email, passwordHash, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get user id: %w", err)
	}

	return &models.User{
		ID:           int(id),
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (s *SQLiteStore) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(
		"SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (s *SQLiteStore) GetUserByID(id int) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(
		"SELECT id, email, password_hash, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// Subject management

func (s *SQLiteStore) ListSubjects(userID int) ([]models.Subject, error) {
	rows, err := s.db.Query(
		"SELECT id, user_id, name, created_at, updated_at FROM subjects WHERE user_id = ? ORDER BY name",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list subjects: %w", err)
	}
	defer rows.Close()

	var subjects []models.Subject
	for rows.Next() {
		var subject models.Subject
		if err := rows.Scan(&subject.ID, &subject.UserID, &subject.Name, &subject.CreatedAt, &subject.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan subject: %w", err)
		}
		subjects = append(subjects, subject)
	}

	return subjects, rows.Err()
}

func (s *SQLiteStore) CreateSubject(userID int, name string) (*models.Subject, error) {
	now := time.Now()
	result, err := s.db.Exec(
		"INSERT INTO subjects (user_id, name, created_at, updated_at) VALUES (?, ?, ?, ?)",
		userID, name, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create subject: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get subject id: %w", err)
	}

	return &models.Subject{
		ID:        int(id),
		UserID:    userID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *SQLiteStore) GetSubject(id int, userID int) (*models.Subject, error) {
	var subject models.Subject
	err := s.db.QueryRow(
		"SELECT id, user_id, name, created_at, updated_at FROM subjects WHERE id = ? AND user_id = ?",
		id, userID,
	).Scan(&subject.ID, &subject.UserID, &subject.Name, &subject.CreatedAt, &subject.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("subject not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subject: %w", err)
	}

	return &subject, nil
}

func (s *SQLiteStore) DeleteSubject(id int, userID int) error {
	result, err := s.db.Exec("DELETE FROM subjects WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete subject: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("subject not found")
	}

	return nil
}

// Topic management

func (s *SQLiteStore) ListTopics(subjectID int, userID int) ([]models.Topic, error) {
	// Verify subject belongs to user
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM subjects WHERE id = ? AND user_id = ?", subjectID, userID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to verify subject ownership: %w", err)
	}
	if count == 0 {
		return nil, fmt.Errorf("subject not found or unauthorized")
	}

	rows, err := s.db.Query(
		"SELECT id, subject_id, name, created_at, updated_at FROM topics WHERE subject_id = ? ORDER BY name",
		subjectID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list topics: %w", err)
	}
	defer rows.Close()

	var topics []models.Topic
	for rows.Next() {
		var topic models.Topic
		if err := rows.Scan(&topic.ID, &topic.SubjectID, &topic.Name, &topic.CreatedAt, &topic.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan topic: %w", err)
		}
		topics = append(topics, topic)
	}

	return topics, rows.Err()
}

func (s *SQLiteStore) CreateTopic(subjectID int, name string, userID int) (*models.Topic, error) {
	// Verify subject belongs to user
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM subjects WHERE id = ? AND user_id = ?", subjectID, userID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to verify subject ownership: %w", err)
	}
	if count == 0 {
		return nil, fmt.Errorf("subject not found or unauthorized")
	}

	now := time.Now()
	result, err := s.db.Exec(
		"INSERT INTO topics (subject_id, name, created_at, updated_at) VALUES (?, ?, ?, ?)",
		subjectID, name, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create topic: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get topic id: %w", err)
	}

	return &models.Topic{
		ID:        int(id),
		SubjectID: subjectID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *SQLiteStore) GetTopic(id int, userID int) (*models.Topic, error) {
	var topic models.Topic
	err := s.db.QueryRow(`
		SELECT t.id, t.subject_id, t.name, t.created_at, t.updated_at
		FROM topics t
		JOIN subjects s ON t.subject_id = s.id
		WHERE t.id = ? AND s.user_id = ?`,
		id, userID,
	).Scan(&topic.ID, &topic.SubjectID, &topic.Name, &topic.CreatedAt, &topic.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("topic not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get topic: %w", err)
	}

	return &topic, nil
}

func (s *SQLiteStore) DeleteTopic(id int, userID int) error {
	result, err := s.db.Exec(`
		DELETE FROM topics
		WHERE id = ? AND subject_id IN (SELECT id FROM subjects WHERE user_id = ?)`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete topic: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("topic not found")
	}

	return nil
}

// Link management

func (s *SQLiteStore) ListLinks(topicID int, userID int) ([]models.Link, error) {
	// Verify topic belongs to user
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM topics t
		JOIN subjects s ON t.subject_id = s.id
		WHERE t.id = ? AND s.user_id = ?`,
		topicID, userID,
	).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to verify topic ownership: %w", err)
	}
	if count == 0 {
		return nil, fmt.Errorf("topic not found or unauthorized")
	}

	rows, err := s.db.Query(
		"SELECT id, topic_id, title, url, created_at, updated_at FROM links WHERE topic_id = ? ORDER BY created_at DESC",
		topicID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list links: %w", err)
	}
	defer rows.Close()

	var links []models.Link
	for rows.Next() {
		var link models.Link
		if err := rows.Scan(&link.ID, &link.TopicID, &link.Title, &link.URL, &link.CreatedAt, &link.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}
		links = append(links, link)
	}

	return links, rows.Err()
}

func (s *SQLiteStore) CreateLink(topicID int, title, url string, userID int) (*models.Link, error) {
	// Verify topic belongs to user
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM topics t
		JOIN subjects s ON t.subject_id = s.id
		WHERE t.id = ? AND s.user_id = ?`,
		topicID, userID,
	).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to verify topic ownership: %w", err)
	}
	if count == 0 {
		return nil, fmt.Errorf("topic not found or unauthorized")
	}

	now := time.Now()
	result, err := s.db.Exec(
		"INSERT INTO links (topic_id, title, url, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		topicID, title, url, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create link: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get link id: %w", err)
	}

	return &models.Link{
		ID:        int(id),
		TopicID:   topicID,
		Title:     title,
		URL:       url,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *SQLiteStore) GetLink(id int, userID int) (*models.Link, error) {
	var link models.Link
	err := s.db.QueryRow(`
		SELECT l.id, l.topic_id, l.title, l.url, l.created_at, l.updated_at
		FROM links l
		JOIN topics t ON l.topic_id = t.id
		JOIN subjects s ON t.subject_id = s.id
		WHERE l.id = ? AND s.user_id = ?`,
		id, userID,
	).Scan(&link.ID, &link.TopicID, &link.Title, &link.URL, &link.CreatedAt, &link.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("link not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get link: %w", err)
	}

	return &link, nil
}

func (s *SQLiteStore) DeleteLink(id int, userID int) error {
	result, err := s.db.Exec(`
		DELETE FROM links
		WHERE id = ? AND topic_id IN (
			SELECT t.id FROM topics t
			JOIN subjects s ON t.subject_id = s.id
			WHERE s.user_id = ?
		)`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("link not found")
	}

	return nil
}

func (s *SQLiteStore) BulkCreateLinks(topicID int, links []models.Link, userID int) error {
	// Verify topic belongs to user
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM topics t
		JOIN subjects s ON t.subject_id = s.id
		WHERE t.id = ? AND s.user_id = ?`,
		topicID, userID,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to verify topic ownership: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("topic not found or unauthorized")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO links (topic_id, title, url, created_at, updated_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, link := range links {
		_, err := stmt.Exec(topicID, link.Title, link.URL, now, now)
		if err != nil {
			return fmt.Errorf("failed to insert link: %w", err)
		}
	}

	return tx.Commit()
}

// Tag management

func (s *SQLiteStore) ListTags(userID int) ([]models.Tag, error) {
	rows, err := s.db.Query(
		"SELECT id, name FROM tags WHERE user_id = ? ORDER BY name",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

func (s *SQLiteStore) CreateTag(name string, userID int) (*models.Tag, error) {
	result, err := s.db.Exec(
		"INSERT INTO tags (user_id, name) VALUES (?, ?)",
		userID, name,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get tag id: %w", err)
	}

	return &models.Tag{
		ID:   int(id),
		Name: name,
	}, nil
}

func (s *SQLiteStore) AddTagToLink(linkID, tagID int, userID int) error {
	// Verify link belongs to user
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM links l
		JOIN topics t ON l.topic_id = t.id
		JOIN subjects s ON t.subject_id = s.id
		WHERE l.id = ? AND s.user_id = ?`,
		linkID, userID,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to verify link ownership: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("link not found or unauthorized")
	}

	// Verify tag belongs to user
	err = s.db.QueryRow("SELECT COUNT(*) FROM tags WHERE id = ? AND user_id = ?", tagID, userID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to verify tag ownership: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("tag not found or unauthorized")
	}

	_, err = s.db.Exec("INSERT OR IGNORE INTO link_tags (link_id, tag_id) VALUES (?, ?)", linkID, tagID)
	if err != nil {
		return fmt.Errorf("failed to add tag to link: %w", err)
	}

	return nil
}

func (s *SQLiteStore) RemoveTagFromLink(linkID, tagID int, userID int) error {
	result, err := s.db.Exec(`
		DELETE FROM link_tags
		WHERE link_id = ? AND tag_id = ? AND link_id IN (
			SELECT l.id FROM links l
			JOIN topics t ON l.topic_id = t.id
			JOIN subjects s ON t.subject_id = s.id
			WHERE s.user_id = ?
		)`,
		linkID, tagID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove tag from link: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("link tag not found")
	}

	return nil
}

func (s *SQLiteStore) GetLinkTags(linkID int, userID int) ([]models.Tag, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.name
		FROM tags t
		JOIN link_tags lt ON t.id = lt.tag_id
		JOIN links l ON lt.link_id = l.id
		JOIN topics tp ON l.topic_id = tp.id
		JOIN subjects s ON tp.subject_id = s.id
		WHERE l.id = ? AND s.user_id = ?
		ORDER BY t.name`,
		linkID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get link tags: %w", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}
