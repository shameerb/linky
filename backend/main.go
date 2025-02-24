package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var (
	markdownDir string
	isProd      bool
)

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
	File  string    `json:"filename"`
	Links []BulkLink `json:"links"`
	Content string  `json:"content"`
}

type ManifestEntry struct {
	File    string   `json:"file"`
	Src     string   `json:"src"`
	CSS     []string `json:"css,omitempty"`
	IsEntry bool     `json:"isEntry,omitempty"`
}

func loadManifest() (map[string]string, error) {
	assets := make(map[string]string)
	
	manifestPath := filepath.Join("dist", "manifest.json")
	data, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest map[string]ManifestEntry
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Find main entry and its CSS
	for _, entry := range manifest {
		if entry.IsEntry {
			assets["js"] = entry.File
			if len(entry.CSS) > 0 {
				assets["css"] = entry.CSS[0]
			}
			break
		}
	}

	return assets, nil
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if isProd {
		assets, err := loadManifest()
		if err != nil {
			log.Printf("Error loading manifest: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles("index.html")
		if err != nil {
			log.Printf("Error parsing template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, assets)
		if err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		// In development, just serve the basic index.html
		http.ServeFile(w, r, "index.html")
	}
}

// customFileServer wraps http.FileServer to set correct MIME types
func customFileServer(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ext := filepath.Ext(r.URL.Path)
		
		switch ext {
		case ".js":
			w.Header().Set("Content-Type", "application/javascript")
		case ".css":
			w.Header().Set("Content-Type", "text/css")
		case ".html":
			w.Header().Set("Content-Type", "text/html")
		default:
			if ct := mime.TypeByExtension(ext); ct != "" {
				w.Header().Set("Content-Type", ct)
			}
		}
		
		fs.ServeHTTP(w, r)
	})
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

	if isProd {
		// In production, serve static files from the dist directory
		r.PathPrefix("/dist/").Handler(http.StripPrefix("/dist/", customFileServer("dist")))
		
		// Serve index.html for all other routes (SPA support)
		r.PathPrefix("/").HandlerFunc(serveIndex)
		
		log.Printf("Server running in production mode on http://localhost:8080")
	} else {
		// In development, serve static assets with proper MIME types
		r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", customFileServer("assets")))
		
		// Serve index.html for all other routes (SPA support)
		r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			http.ServeFile(w, r, "index.html")
		})
		
		log.Printf("Server running in development mode on http://localhost:8080 (API only)")
	}

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
