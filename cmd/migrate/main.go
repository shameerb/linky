package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"markdown-editor/internal/auth"
	"markdown-editor/internal/models"
	"markdown-editor/internal/services/store/sqlite"
)

var (
	markdownDir string
	dbPath      string
	email       string
	password    string
)

func main() {
	flag.StringVar(&markdownDir, "markdown-dir", "./Links", "Path to markdown files directory")
	flag.StringVar(&dbPath, "db-path", "./linky.db", "Path to SQLite database")
	flag.StringVar(&email, "email", "", "Email for the default user")
	flag.StringVar(&password, "password", "password", "Password for the default user (default: 'password')")
	flag.Parse()

	if email == "" {
		log.Fatal("The -email flag is required")
	}

	// Use default password if not provided
	if password == "" {
		password = "password"
	}

	// Initialize SQLite store
	store, err := sqlite.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize SQLite store: %v", err)
	}
	defer store.Close()

	// Try to get existing user first
	user, err := store.GetUserByEmail(email)
	if err != nil {
		// User doesn't exist, create new one
		passwordHash, err := auth.HashPassword(password)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		user, err = store.CreateUser(email, passwordHash)
		if err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}

		log.Printf("Created new user: %s (ID: %d)", user.Email, user.ID)
	} else {
		log.Printf("Using existing user: %s (ID: %d)", user.Email, user.ID)
	}

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

		log.Printf("Processing file: %s -> Subject: %s", filename, subjectName)

		// Create subject
		subject, err := store.CreateSubject(user.ID, subjectName)
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

		log.Printf("  Found %d topics in %s", len(topics), filename)

		// Create topics and links
		for topicName, links := range topics {
			if len(links) == 0 {
				continue
			}

			// Create topic
			topic, err := store.CreateTopic(subject.ID, topicName, user.ID)
			if err != nil {
				log.Printf("Warning: Failed to create topic %s: %v", topicName, err)
				continue
			}

			log.Printf("    Created topic: %s with %d links", topicName, len(links))

			// Create links
			var modelLinks []models.Link
			for _, link := range links {
				modelLinks = append(modelLinks, models.Link{
					Title: link.Title,
					URL:   link.URL,
				})
			}

			if err := store.BulkCreateLinks(topic.ID, modelLinks, user.ID); err != nil {
				log.Printf("Warning: Failed to create links for topic %s: %v", topicName, err)
			}
		}
	}

	// Generate summary
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("Migration completed successfully!")
	log.Println(strings.Repeat("=", 60))
	log.Printf("Database: %s", dbPath)
	log.Printf("User: %s", email)
	log.Printf("Password: %s", password)

	// Get and display summary
	log.Println("\n" + strings.Repeat("-", 60))
	log.Println("SUMMARY")
	log.Println(strings.Repeat("-", 60))

	subjects, err := store.ListSubjects(user.ID)
	if err != nil {
		log.Printf("Warning: Failed to get subjects for summary: %v", err)
		return
	}

	totalLinks := 0
	totalTopics := 0

	for _, subject := range subjects {
		topics, err := store.ListTopics(subject.ID, user.ID)
		if err != nil {
			log.Printf("Warning: Failed to get topics for subject %s: %v", subject.Name, err)
			continue
		}

		subjectLinkCount := 0
		for _, topic := range topics {
			links, err := store.ListLinks(topic.ID, user.ID)
			if err != nil {
				continue
			}
			subjectLinkCount += len(links)
		}

		totalTopics += len(topics)
		totalLinks += subjectLinkCount
		log.Printf("  %-20s %3d topics  %4d links", subject.Name, len(topics), subjectLinkCount)
	}

	log.Println(strings.Repeat("-", 60))
	log.Printf("Total: %d subjects, %d topics, %d links", len(subjects), totalTopics, totalLinks)
	log.Println(strings.Repeat("=", 60))
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

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return topics, nil
}
