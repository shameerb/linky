package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"io/fs"

	"github.com/gorilla/mux"
)

var (
	markdownDir string
	isProd      bool
)

// Embed the entire dist directory
//
//go:embed dist/*
var embeddedFiles embed.FS

type Link struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Timestamp time.Time `json:"timestamp"`
}

type Subject struct {
	Subject string `json:"subject"`
	Links   []Link `json:"links"`
}

type Response struct {
	Data []Subject `json:"data"`
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
	URL         string `json:"url"`
	Description string `json:"description"`
}

type BulkLinksRequest struct {
	File    string     `json:"filename"`
	Links   []BulkLink `json:"links"`
	Content string     `json:"content"`
}

func main() {
	r := mux.NewRouter()

	// Check if we're in production mode
	isProd = os.Getenv("GO_ENV") == "production"

	markdownDir = os.Getenv("MARKDOWN_DIR")
	if markdownDir == "" {
		log.Fatal("MARKDOWN_DIR environment variable is required")
	}

	if _, err := os.Stat(markdownDir); os.IsNotExist(err) {
		log.Fatalf("Markdown directory does not exist: %s", markdownDir)
	}

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/files", listFiles).Methods("GET")
	api.HandleFunc("/file/{filename}", getFile).Methods("GET")
	api.HandleFunc("/delete_links", deleteLinks).Methods("POST")
	api.HandleFunc("/bulk_links", addBulkLinks).Methods("POST")

	// Serve static files from the embedded dist directory
	distFS, err := fs.Sub(embeddedFiles, "dist")
	if err != nil {
		log.Fatalf("Failed to create file system for dist: %v", err)
	}
	r.PathPrefix("/").Handler(http.FileServer(http.FS(distFS)))

	log.Printf("Using markdown directory: %s", markdownDir)
	log.Fatal(http.ListenAndServe(":8080", r))
}

func listFiles(w http.ResponseWriter, r *http.Request) {
	files, err := filepath.Glob(filepath.Join(markdownDir, "*.md"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i := range files {
		files[i] = filepath.Base(files[i])
	}

	json.NewEncoder(w).Encode(files)
}

func getFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]
	file, err := os.Open(filepath.Join(markdownDir, filename))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var response Response
	var currentSubject *Subject
	linkRegex := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			if currentSubject != nil {
				response.Data = append(response.Data, *currentSubject)
			}
			currentSubject = &Subject{
				Subject: strings.TrimSpace(strings.TrimLeft(line, "#")),
			}

		} else if strings.HasPrefix(line, "-") {
			if currentSubject == nil {
				currentSubject = &Subject{
					Subject: "Others",
				}
			}
			matches := linkRegex.FindStringSubmatch(line)
			if len(matches) == 3 {
				currentSubject.Links = append(currentSubject.Links, Link{
					Title: matches[1],
					URL:   matches[2],
				})
			}
		}
	}
	if currentSubject != nil {
		response.Data = append(response.Data, *currentSubject)
	}
	if err := scanner.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(response)
}

func deleteLinks(w http.ResponseWriter, r *http.Request) {
	var req DeleteLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filepath := filepath.Join(markdownDir, req.Filename)
	content, err := os.ReadFile(filepath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	linksToDelete := make(map[string]bool)
	for _, link := range req.Links {
		linksToDelete[fmt.Sprintf("[%s](%s)", link.Title, link.URL)] = true
	}

	for _, line := range lines {
		shouldDelete := false
		for linkToDelete := range linksToDelete {
			if strings.Contains(line, linkToDelete) {
				shouldDelete = true
				break
			}
		}
		if !shouldDelete {
			newLines = append(newLines, line)
		}
	}

	err = os.WriteFile(filepath, []byte(strings.Join(newLines, "\n")), 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func addBulkLinks(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received bulk links request")

	var req BulkLinksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Request data - Filename: %s, Content length: %d", req.File, len(req.Content))

	if req.File == "" {
		log.Printf("Error: filename is empty")
		http.Error(w, "filename is required", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(markdownDir, req.File)
	log.Printf("Target file path: %s", filePath)

	// Create file if it doesn't exist and open in read-write mode
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Get today's date in DD/MM/YYYY format
	now := time.Now()
	dateSection := fmt.Sprintf("# %s\n\n", now.Format("02/01/2006"))

	// Add newlines before new section if file isn't empty
	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("Error getting file stats: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare the content with the date section
	content := strings.TrimSpace(req.Content)
	if fileInfo.Size() > 0 {
		dateSection = "\n\n" + dateSection
	}

	finalContent := dateSection + content + "\n"
	log.Printf("Writing content:\n%s", finalContent)

	if _, err := file.WriteString(finalContent); err != nil {
		log.Printf("Error writing content: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully added bulk links")
	w.WriteHeader(http.StatusOK)
}
