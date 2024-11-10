package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
)

var markdownDir string

type Link struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type Subject struct {
	Subject string `json:"subject"`
	Links   []Link `json:"links"`
}

type Response struct {
	Data []Subject `json:"data"`
}

type DeleteLink struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type DeleteLinkRequest struct {
	Filename string       `json:"filename"`
	Links    []DeleteLink `json:"links"`
}

func main() {
	r := mux.NewRouter()
	markdownDir = os.Getenv("MARKDOWN_DIR")
	if markdownDir == "" {
		log.Fatal("set MARKDOWN_DIR env variable")
	}

	// staticFilesDir := os.Getenv("STATIC_FILES_DIR")

	// if staticFilesDir == "" {
	// 	log.Fatal("set STATIC_FILES_DIR env variable")
	// }
	staticFilesDir := "./"
	r.HandleFunc("/files", listFiles).Methods("GET")
	r.HandleFunc("/file/{filename}", getFile).Methods("GET")
	r.HandleFunc("/delete_links", deleteLinks).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(staticFilesDir))).Methods("GET")

	log.Println("Server running on http://localhost:8080")
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
