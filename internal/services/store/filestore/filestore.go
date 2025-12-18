package filestore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"markdown-editor/internal/models"
)

// FileStore implements UserStore interface using JSON files
type FileStore struct {
	dataDir string
	mu      sync.RWMutex

	// In-memory cache
	subjects map[int]*models.Subject
	topics   map[int]*models.Topic
	links    map[int]*models.Link
	tags     map[int]*models.Tag
	linkTags map[int][]int // linkID -> []tagID

	// ID counters
	nextSubjectID int
	nextTopicID   int
	nextLinkID    int
	nextTagID     int
}

// Default user (password: "password")
// Password hash for "password" using bcrypt
const defaultPasswordHash = "$2a$10$TD.OX7dQDb06coImofJso.tEoVwfncKUOg9F/BYKIeSBdPXcNrbsq"

var defaultUser = &models.User{
	ID:           1,
	Email:        "shameer789@gmail.com",
	PasswordHash: defaultPasswordHash,
	CreatedAt:    time.Now(),
	UpdatedAt:    time.Now(),
}

// New creates a new file-based store
func New(dataDir string) (*FileStore, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	fs := &FileStore{
		dataDir:       dataDir,
		subjects:      make(map[int]*models.Subject),
		topics:        make(map[int]*models.Topic),
		links:         make(map[int]*models.Link),
		tags:          make(map[int]*models.Tag),
		linkTags:      make(map[int][]int),
		nextSubjectID: 1,
		nextTopicID:   1,
		nextLinkID:    1,
		nextTagID:     1,
	}

	// Load existing data
	if err := fs.load(); err != nil {
		return nil, fmt.Errorf("failed to load data: %w", err)
	}

	return fs, nil
}

// Data structure for JSON persistence
type persistedData struct {
	Subjects      map[int]*models.Subject `json:"subjects"`
	Topics        map[int]*models.Topic   `json:"topics"`
	Links         map[int]*models.Link    `json:"links"`
	Tags          map[int]*models.Tag     `json:"tags"`
	LinkTags      map[int][]int           `json:"link_tags"`
	NextSubjectID int                     `json:"next_subject_id"`
	NextTopicID   int                     `json:"next_topic_id"`
	NextLinkID    int                     `json:"next_link_id"`
	NextTagID     int                     `json:"next_tag_id"`
}

func (fs *FileStore) dataFile() string {
	return filepath.Join(fs.dataDir, "linky-data.json")
}

func (fs *FileStore) load() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := os.ReadFile(fs.dataFile())
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, that's okay
			return nil
		}
		return err
	}

	var pd persistedData
	if err := json.Unmarshal(data, &pd); err != nil {
		return err
	}

	fs.subjects = pd.Subjects
	fs.topics = pd.Topics
	fs.links = pd.Links
	fs.tags = pd.Tags
	fs.linkTags = pd.LinkTags
	fs.nextSubjectID = pd.NextSubjectID
	fs.nextTopicID = pd.NextTopicID
	fs.nextLinkID = pd.NextLinkID
	fs.nextTagID = pd.NextTagID

	if fs.subjects == nil {
		fs.subjects = make(map[int]*models.Subject)
	}
	if fs.topics == nil {
		fs.topics = make(map[int]*models.Topic)
	}
	if fs.links == nil {
		fs.links = make(map[int]*models.Link)
	}
	if fs.tags == nil {
		fs.tags = make(map[int]*models.Tag)
	}
	if fs.linkTags == nil {
		fs.linkTags = make(map[int][]int)
	}

	return nil
}

func (fs *FileStore) save() error {
	pd := persistedData{
		Subjects:      fs.subjects,
		Topics:        fs.topics,
		Links:         fs.links,
		Tags:          fs.tags,
		LinkTags:      fs.linkTags,
		NextSubjectID: fs.nextSubjectID,
		NextTopicID:   fs.nextTopicID,
		NextLinkID:    fs.nextLinkID,
		NextTagID:     fs.nextTagID,
	}

	data, err := json.MarshalIndent(pd, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(fs.dataFile(), data, 0644)
}

// User management (always returns default user)
func (fs *FileStore) CreateUser(email, passwordHash string) (*models.User, error) {
	return defaultUser, nil
}

func (fs *FileStore) GetUserByEmail(email string) (*models.User, error) {
	return defaultUser, nil
}

func (fs *FileStore) GetUserByID(id int) (*models.User, error) {
	return defaultUser, nil
}

// Subject management
func (fs *FileStore) ListSubjects(userID int) ([]models.Subject, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	subjects := make([]models.Subject, 0, len(fs.subjects))
	for _, s := range fs.subjects {
		subjects = append(subjects, *s)
	}
	return subjects, nil
}

func (fs *FileStore) CreateSubject(userID int, name string) (*models.Subject, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	subject := &models.Subject{
		ID:        fs.nextSubjectID,
		UserID:    userID,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	fs.subjects[subject.ID] = subject
	fs.nextSubjectID++

	if err := fs.save(); err != nil {
		return nil, err
	}

	return subject, nil
}

func (fs *FileStore) GetSubject(id int, userID int) (*models.Subject, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	subject, ok := fs.subjects[id]
	if !ok {
		return nil, fmt.Errorf("subject not found")
	}
	return subject, nil
}

func (fs *FileStore) DeleteSubject(id int, userID int) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Delete all topics and links in this subject
	for topicID, topic := range fs.topics {
		if topic.SubjectID == id {
			// Delete all links in this topic
			for linkID, link := range fs.links {
				if link.TopicID == topicID {
					delete(fs.links, linkID)
					delete(fs.linkTags, linkID)
				}
			}
			delete(fs.topics, topicID)
		}
	}

	delete(fs.subjects, id)
	return fs.save()
}

// Topic management
func (fs *FileStore) ListTopics(subjectID int, userID int) ([]models.Topic, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	topics := make([]models.Topic, 0)
	for _, t := range fs.topics {
		if t.SubjectID == subjectID {
			topics = append(topics, *t)
		}
	}
	return topics, nil
}

func (fs *FileStore) CreateTopic(subjectID int, name string, userID int) (*models.Topic, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	topic := &models.Topic{
		ID:        fs.nextTopicID,
		SubjectID: subjectID,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	fs.topics[topic.ID] = topic
	fs.nextTopicID++

	if err := fs.save(); err != nil {
		return nil, err
	}

	return topic, nil
}

func (fs *FileStore) GetTopic(id int, userID int) (*models.Topic, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	topic, ok := fs.topics[id]
	if !ok {
		return nil, fmt.Errorf("topic not found")
	}
	return topic, nil
}

func (fs *FileStore) DeleteTopic(id int, userID int) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Delete all links in this topic
	for linkID, link := range fs.links {
		if link.TopicID == id {
			delete(fs.links, linkID)
			delete(fs.linkTags, linkID)
		}
	}

	delete(fs.topics, id)
	return fs.save()
}

// Link management
func (fs *FileStore) ListLinks(topicID int, userID int) ([]models.Link, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	links := make([]models.Link, 0)
	for _, l := range fs.links {
		if l.TopicID == topicID {
			links = append(links, *l)
		}
	}
	return links, nil
}

func (fs *FileStore) CreateLink(topicID int, title, url string, userID int) (*models.Link, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	link := &models.Link{
		ID:        fs.nextLinkID,
		TopicID:   topicID,
		Title:     title,
		URL:       url,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	fs.links[link.ID] = link
	fs.nextLinkID++

	if err := fs.save(); err != nil {
		return nil, err
	}

	return link, nil
}

func (fs *FileStore) GetLink(id int, userID int) (*models.Link, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	link, ok := fs.links[id]
	if !ok {
		return nil, fmt.Errorf("link not found")
	}
	return link, nil
}

func (fs *FileStore) DeleteLink(id int, userID int) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	delete(fs.links, id)
	delete(fs.linkTags, id)
	return fs.save()
}

func (fs *FileStore) BulkCreateLinks(topicID int, links []models.Link, userID int) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for _, link := range links {
		newLink := &models.Link{
			ID:        fs.nextLinkID,
			TopicID:   topicID,
			Title:     link.Title,
			URL:       link.URL,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		fs.links[newLink.ID] = newLink
		fs.nextLinkID++
	}

	return fs.save()
}

// Tag management
func (fs *FileStore) ListTags(userID int) ([]models.Tag, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	tags := make([]models.Tag, 0, len(fs.tags))
	for _, t := range fs.tags {
		tags = append(tags, *t)
	}
	return tags, nil
}

func (fs *FileStore) CreateTag(name string, userID int) (*models.Tag, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	tag := &models.Tag{
		ID:   fs.nextTagID,
		Name: name,
	}

	fs.tags[tag.ID] = tag
	fs.nextTagID++

	if err := fs.save(); err != nil {
		return nil, err
	}

	return tag, nil
}

func (fs *FileStore) AddTagToLink(linkID, tagID int, userID int) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Check if already added
	for _, tid := range fs.linkTags[linkID] {
		if tid == tagID {
			return nil // Already exists
		}
	}

	fs.linkTags[linkID] = append(fs.linkTags[linkID], tagID)
	return fs.save()
}

func (fs *FileStore) RemoveTagFromLink(linkID, tagID int, userID int) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	tags := fs.linkTags[linkID]
	newTags := make([]int, 0, len(tags))
	for _, tid := range tags {
		if tid != tagID {
			newTags = append(newTags, tid)
		}
	}
	fs.linkTags[linkID] = newTags

	return fs.save()
}

func (fs *FileStore) GetLinkTags(linkID int, userID int) ([]models.Tag, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	tagIDs := fs.linkTags[linkID]
	tags := make([]models.Tag, 0, len(tagIDs))
	for _, tid := range tagIDs {
		if tag, ok := fs.tags[tid]; ok {
			tags = append(tags, *tag)
		}
	}
	return tags, nil
}

// Close is a no-op for file store
func (fs *FileStore) Close() error {
	return nil
}
