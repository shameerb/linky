package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"markdown-editor/internal/models"
	"markdown-editor/internal/services/store/filestore"
)

var (
	markdownDir string
	dataDir     string
)

func main() {
	flag.StringVar(&markdownDir, "markdown-dir", "./Links", "Path to markdown files directory")
	flag.StringVar(&dataDir, "data-dir", "./data", "Path to file store data directory")
	flag.Parse()

	// Initialize file store
	store, err := filestore.New(dataDir)
	if err != nil {
		log.Fatalf("Failed to initialize file store: %v", err)
	}
	defer store.Close()

	// User ID is always 1 for file store
	userID := 1

	log.Printf("Importing from: %s", markdownDir)
	log.Printf("Saving to: %s", dataDir)

	// Get all markdown files
	files, err := filepath.Glob(filepath.Join(markdownDir, "*.md"))
	if err != nil {
		log.Fatalf("Failed to list markdown files: %v", err)
	}

	log.Printf("Found %d markdown files", len(files))

	// Process each file
	for _, filePath := range files {
		filename := filepath.Base(filePath)
		subjectName := strings.TrimSuffix(filename, ".md")

		log.Printf("Processing: %s -> Subject: %s", filename, subjectName)

		// Create subject
		subject, err := store.CreateSubject(userID, subjectName)
		if err != nil {
			log.Printf("Warning: Failed to create subject %s: %v", subjectName, err)
			continue
		}

		// Parse markdown file
		topics, err := parseMarkdownFile(filePath)
		if err != nil {
			log.Printf("Warning: Failed to parse file %s: %v", filename, err)
			continue
		}

		log.Printf("  Found %d topics", len(topics))

		// Create topics and links
		for topicName, links := range topics {
			if len(links) == 0 {
				continue
			}

			// Create topic
			topic, err := store.CreateTopic(subject.ID, topicName, userID)
			if err != nil {
				log.Printf("Warning: Failed to create topic %s: %v", topicName, err)
				continue
			}

			log.Printf("    Topic: %s (%d links)", topicName, len(links))

			// Create links
			var modelLinks []models.Link
			for _, link := range links {
				modelLinks = append(modelLinks, models.Link{
					Title: link.Title,
					URL:   link.URL,
				})
			}

			if err := store.BulkCreateLinks(topic.ID, modelLinks, userID); err != nil {
				log.Printf("Warning: Failed to create links: %v", err)
			}
		}
	}

	log.Println("✅ Migration completed!")
	log.Printf("Data saved to: %s/linky-data.json", dataDir)
	log.Printf("Login with: shameer789@gmail.com / password")
}

type LinkData struct {
	Title string
	URL   string
}

func parseMarkdownFile(filePath string) (map[string][]LinkData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	topics := make(map[string][]LinkData)
	currentTopic := "Uncategorized"
	linkRegex := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Check for topic header (### Topic Name)
		if strings.HasPrefix(line, "### ") {
			currentTopic = strings.TrimSpace(strings.TrimPrefix(line, "### "))
			if topics[currentTopic] == nil {
				topics[currentTopic] = []LinkData{}
			}
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			// Extract link from markdown list item
			matches := linkRegex.FindStringSubmatch(line)
			if len(matches) == 3 {
				if topics[currentTopic] == nil {
					topics[currentTopic] = []LinkData{}
				}
				topics[currentTopic] = append(topics[currentTopic], LinkData{
					Title: matches[1],
					URL:   matches[2],
				})
			}
		}
	}

	return topics, scanner.Err()
}
