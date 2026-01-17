package markdownstore

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"markdown-editor/internal/models"
)

// MarkdownStore implements UserStore interface using markdown files
// Files are subjects, ### headings are topics, links are stored as markdown links
type MarkdownStore struct {
	dataDir string
	mu      sync.RWMutex
}

// Default user (password: "password")
const defaultPasswordHash = "$2a$10$TD.OX7dQDb06coImofJso.tEoVwfncKUOg9F/BYKIeSBdPXcNrbsq"

var defaultUser = &models.User{
	ID:           1,
	Email:        "shameer789@gmail.com",
	PasswordHash: defaultPasswordHash,
	CreatedAt:    time.Now(),
	UpdatedAt:    time.Now(),
}

// New creates a new markdown-based store
func New(dataDir string) (*MarkdownStore, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &MarkdownStore{
		dataDir: dataDir,
	}, nil
}

// User management (always returns default user)
func (ms *MarkdownStore) CreateUser(email, passwordHash string) (*models.User, error) {
	return defaultUser, nil
}

func (ms *MarkdownStore) GetUserByEmail(email string) (*models.User, error) {
	return defaultUser, nil
}

func (ms *MarkdownStore) GetUserByID(id int) (*models.User, error) {
	return defaultUser, nil
}

// Subject management (subjects = markdown files)
func (ms *MarkdownStore) ListSubjects(userID int) ([]models.Subject, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	files, err := filepath.Glob(filepath.Join(ms.dataDir, "*.md"))
	if err != nil {
		return nil, err
	}

	subjects := make([]models.Subject, 0, len(files))
	for i, filePath := range files {
		filename := filepath.Base(filePath)
		name := strings.TrimSuffix(filename, ".md")

		subjects = append(subjects, models.Subject{
			ID:        i + 1,
			UserID:    userID,
			Name:      name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	return subjects, nil
}

func (ms *MarkdownStore) CreateSubject(userID int, name string) (*models.Subject, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	filename := filepath.Join(ms.dataDir, name+".md")

	// Create empty file if it doesn't exist
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if err := os.WriteFile(filename, []byte(""), 0644); err != nil {
			return nil, err
		}
	}

	// Get ID by counting files
	files, _ := filepath.Glob(filepath.Join(ms.dataDir, "*.md"))

	return &models.Subject{
		ID:        len(files),
		UserID:    userID,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (ms *MarkdownStore) GetSubject(id int, userID int) (*models.Subject, error) {
	subjects, err := ms.ListSubjects(userID)
	if err != nil {
		return nil, err
	}

	for _, s := range subjects {
		if s.ID == id {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("subject not found")
}

func (ms *MarkdownStore) DeleteSubject(id int, userID int) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Find subject by ID without calling methods that acquire locks
	files, err := filepath.Glob(filepath.Join(ms.dataDir, "*.md"))
	if err != nil {
		return err
	}

	for i, filePath := range files {
		if i+1 == id {
			return os.Remove(filePath)
		}
	}

	return fmt.Errorf("subject not found")
}

// Topic management (topics = ### headings in markdown files)
func (ms *MarkdownStore) ListTopics(subjectID int, userID int) ([]models.Topic, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	subject, err := ms.GetSubject(subjectID, userID)
	if err != nil {
		return nil, err
	}

	filename := filepath.Join(ms.dataDir, subject.Name+".md")
	// Use ordered parsing to preserve topic order from file
	orderedTopics, err := ms.parseTopicsOrdered(filename)
	if err != nil {
		return nil, err
	}

	result := make([]models.Topic, 0, len(orderedTopics))
	for topicIndex, topicData := range orderedTopics {
		// Make topic ID globally unique by combining subject ID with topic index
		// Format: subjectID * 1000 + (topicIndex + 1)
		// This ensures topics from different subjects have different IDs
		topicID := subjectID*1000 + (topicIndex + 1)
		result = append(result, models.Topic{
			ID:        topicID,
			SubjectID: subjectID,
			Name:      topicData.Name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	return result, nil
}

func (ms *MarkdownStore) CreateTopic(subjectID int, name string, userID int) (*models.Topic, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	subject, err := ms.GetSubject(subjectID, userID)
	if err != nil {
		return nil, err
	}

	filename := filepath.Join(ms.dataDir, subject.Name+".md")

	// Read existing content
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Append new topic heading
	newContent := string(content)
	if len(newContent) > 0 && !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	newContent += fmt.Sprintf("\n### %s\n", name)

	// Write back
	if err := os.WriteFile(filename, []byte(newContent), 0644); err != nil {
		return nil, err
	}

	// Get topic ID
	topics, _ := ms.parseTopics(filename)

	return &models.Topic{
		ID:        len(topics),
		SubjectID: subjectID,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (ms *MarkdownStore) GetTopic(id int, userID int) (*models.Topic, error) {
	// Need to search through all subjects to find the topic
	subjects, err := ms.ListSubjects(userID)
	if err != nil {
		return nil, err
	}

	for _, subject := range subjects {
		topics, err := ms.ListTopics(subject.ID, userID)
		if err != nil {
			continue
		}

		for _, topic := range topics {
			if topic.ID == id {
				return &topic, nil
			}
		}
	}

	return nil, fmt.Errorf("topic not found")
}

func (ms *MarkdownStore) DeleteTopic(id int, userID int) error {
	// This is complex - would need to remove the ### heading and all links under it
	// For now, return error
	return fmt.Errorf("deleting topics not yet implemented for markdown store")
}

// Link management
func (ms *MarkdownStore) ListLinks(topicID int, userID int) ([]models.Link, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// Find the topic
	subjects, err := ms.ListSubjects(userID)
	if err != nil {
		return nil, err
	}

	for _, subject := range subjects {
		filename := filepath.Join(ms.dataDir, subject.Name+".md")
		topicsMap, err := ms.parseTopics(filename)
		if err != nil {
			continue
		}

		// Match topic by ID (counted from all topics in this file)
		topics, _ := ms.ListTopics(subject.ID, userID)
		for _, topic := range topics {
			if topic.ID == topicID {
				// Found the topic, return its links
				links := topicsMap[topic.Name]
				result := make([]models.Link, len(links))
				for i, link := range links {
					// Make link ID globally unique by combining topic ID with link index
					// Format: topicID * 10000 + (i + 1)
					linkID := topicID*10000 + (i + 1)
					result[i] = models.Link{
						ID:        linkID,
						TopicID:   topicID,
						Title:     link.Title,
						URL:       link.URL,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}
				}
				return result, nil
			}
		}
	}

	return []models.Link{}, nil
}

func (ms *MarkdownStore) CreateLink(topicID int, title, url string, userID int) (*models.Link, error) {
	return nil, ms.BulkCreateLinks(topicID, []models.Link{{Title: title, URL: url}}, userID)
}

func (ms *MarkdownStore) BulkCreateLinks(topicID int, links []models.Link, userID int) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Find the subject and topic
	subjects, err := ms.ListSubjects(userID)
	if err != nil {
		return err
	}

	for _, subject := range subjects {
		topics, err := ms.ListTopics(subject.ID, userID)
		if err != nil {
			continue
		}

		for _, topic := range topics {
			if topic.ID == topicID {
				// Found the topic, append links
				filename := filepath.Join(ms.dataDir, subject.Name+".md")
				return ms.appendLinksToTopic(filename, topic.Name, links)
			}
		}
	}

	return fmt.Errorf("topic not found")
}

func (ms *MarkdownStore) GetLink(id int, userID int) (*models.Link, error) {
	return nil, fmt.Errorf("not implemented")
}

func (ms *MarkdownStore) DeleteLink(id int, userID int) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Decode link ID: topicID * 10000 + (linkIndex + 1)
	topicID := id / 10000
	linkIndex := (id % 10000) - 1

	if linkIndex < 0 {
		return fmt.Errorf("invalid link ID")
	}

	// Find subject and topic without calling methods that acquire locks
	// (to avoid deadlock since we already hold the write lock)
	files, err := filepath.Glob(filepath.Join(ms.dataDir, "*.md"))
	if err != nil {
		return err
	}

	for subjectIndex, filePath := range files {
		subjectID := subjectIndex + 1
		filename := filepath.Base(filePath)
		subjectName := strings.TrimSuffix(filename, ".md")

		// Parse topics from this file
		orderedTopics, err := ms.parseTopicsOrdered(filePath)
		if err != nil {
			continue
		}

		for topicIndex, topic := range orderedTopics {
			// Calculate topic ID the same way ListTopics does
			calculatedTopicID := subjectID*1000 + (topicIndex + 1)
			if calculatedTopicID == topicID {
				// Found the topic, now remove the link from the file
				return ms.removeLinkFromTopic(filepath.Join(ms.dataDir, subjectName+".md"), topic.Name, linkIndex)
			}
		}
	}

	return fmt.Errorf("link not found")
}

// Tag management (not supported in markdown files)
func (ms *MarkdownStore) ListTags(userID int) ([]models.Tag, error) {
	return []models.Tag{}, nil
}

func (ms *MarkdownStore) CreateTag(name string, userID int) (*models.Tag, error) {
	return nil, fmt.Errorf("tags not supported in markdown store")
}

func (ms *MarkdownStore) AddTagToLink(linkID, tagID int, userID int) error {
	return fmt.Errorf("tags not supported in markdown store")
}

func (ms *MarkdownStore) RemoveTagFromLink(linkID, tagID int, userID int) error {
	return fmt.Errorf("tags not supported in markdown store")
}

func (ms *MarkdownStore) GetLinkTags(linkID int, userID int) ([]models.Tag, error) {
	return []models.Tag{}, nil
}

// Helper functions

type linkData struct {
	Title string
	URL   string
}

type topicData struct {
	Name  string
	Links []linkData
}

// parseTopicsOrdered returns topics in the order they appear in the file
// Merges duplicate topic names into a single topic
func (ms *MarkdownStore) parseTopicsOrdered(filename string) ([]topicData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var topics []topicData
	topicIndex := make(map[string]int) // Track topic name -> index in topics slice
	var currentTopic *topicData
	linkRegex := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "### ") {
			topicName := strings.TrimSpace(strings.TrimPrefix(line, "### "))

			// Check if topic already exists (duplicate name)
			if idx, exists := topicIndex[topicName]; exists {
				// Reuse existing topic (merge duplicates)
				currentTopic = &topics[idx]
			} else {
				// Add new topic
				topics = append(topics, topicData{
					Name:  topicName,
					Links: []linkData{},
				})
				topicIndex[topicName] = len(topics) - 1
				currentTopic = &topics[len(topics)-1]
			}
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			matches := linkRegex.FindStringSubmatch(line)
			if len(matches) == 3 && currentTopic != nil {
				currentTopic.Links = append(currentTopic.Links, linkData{
					Title: matches[1],
					URL:   matches[2],
				})
			}
		}
	}

	return topics, scanner.Err()
}

// parseTopics returns topics as a map (legacy, order not guaranteed)
func (ms *MarkdownStore) parseTopics(filename string) (map[string][]linkData, error) {
	ordered, err := ms.parseTopicsOrdered(filename)
	if err != nil {
		return nil, err
	}

	// Convert to map for backward compatibility
	topics := make(map[string][]linkData)
	for _, topic := range ordered {
		topics[topic.Name] = topic.Links
	}

	return topics, nil
}

func (ms *MarkdownStore) removeLinkFromTopic(filename, topicName string, linkIndex int) error {
	// Read entire file
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	newLines := make([]string, 0, len(lines))
	linkRegex := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)

	inTopic := false
	currentLinkIndex := 0
	removed := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check if entering our target topic
		if trimmedLine == fmt.Sprintf("### %s", topicName) {
			inTopic = true
			currentLinkIndex = 0
			newLines = append(newLines, line)
			continue
		}

		// Check if leaving topic (hit another heading)
		if strings.HasPrefix(trimmedLine, "### ") {
			inTopic = false
		}

		// If in topic and this is a link line
		if inTopic && (strings.HasPrefix(trimmedLine, "- ") || strings.HasPrefix(trimmedLine, "* ")) {
			if linkRegex.MatchString(trimmedLine) {
				if currentLinkIndex == linkIndex {
					// Skip this line (delete the link)
					removed = true
					currentLinkIndex++
					continue
				}
				currentLinkIndex++
			}
		}

		newLines = append(newLines, line)
	}

	if !removed {
		return fmt.Errorf("link not found in topic")
	}

	// Write back
	newContent := strings.Join(newLines, "\n")
	return os.WriteFile(filename, []byte(newContent), 0644)
}

func (ms *MarkdownStore) appendLinksToTopic(filename, topicName string, links []models.Link) error {
	// Read entire file
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	newLines := make([]string, 0, len(lines)+len(links))

	inTopic := false
	topicFound := false
	inserted := false

	for i, line := range lines {
		newLines = append(newLines, line)

		// Check if this is our topic
		if strings.TrimSpace(line) == fmt.Sprintf("### %s", topicName) {
			inTopic = true
			topicFound = true
			continue
		}

		// If we're in the topic and hit another topic or end, insert links before it
		if inTopic && !inserted {
			if strings.HasPrefix(strings.TrimSpace(line), "###") || i == len(lines)-1 {
				// Insert links before this line
				newLines = newLines[:len(newLines)-1] // Remove the line we just added

				for _, link := range links {
					newLines = append(newLines, fmt.Sprintf("- [%s](%s)", link.Title, link.URL))
				}

				newLines = append(newLines, line) // Add back the line
				inserted = true
				inTopic = false
			}
		}
	}

	// If topic found but didn't insert (topic is at end of file)
	if topicFound && !inserted {
		for _, link := range links {
			newLines = append(newLines, fmt.Sprintf("- [%s](%s)", link.Title, link.URL))
		}
	}

	// Write back
	newContent := strings.Join(newLines, "\n")
	return os.WriteFile(filename, []byte(newContent), 0644)
}

// Close is a no-op for markdown store
func (ms *MarkdownStore) Close() error {
	return nil
}
