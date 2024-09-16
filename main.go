package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gorilla/mux"
)

const MARKDOWN_DIR = ""

type FileContent struct {
	Data []string `json:"data"`
}

type DeleteLineRequest struct {
	Filename  string `json:"filename"`
	LineIndex int    `json:"lineIndex"`
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/files", listFiles).Methods("GET")
	r.HandleFunc("/file/{filename}", getFile).Methods("GET")
	r.HandleFunc("/delete_line", deleteLine).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(".")))

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func listFiles(w http.ResponseWriter, r *http.Request) {
	files, err := filepath.Glob(filepath.Join(MARKDOWN_DIR, "*.md"))
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
	content, err := os.ReadFile(filepath.Join(MARKDOWN_DIR, filename))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lines := strings.Split(string(content), "\n")
	htmlContent := make([]string, 0, len(lines))
	for _, line := range lines {
		// if strings.TrimSpace(line) != "" {
		htmlContent = append(htmlContent, string(markdown.ToHTML([]byte(line), nil, nil)))
		// }
	}

	json.NewEncoder(w).Encode(FileContent{Data: htmlContent})
}

func deleteLine(w http.ResponseWriter, r *http.Request) {
	var req DeleteLineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filepath := filepath.Join(MARKDOWN_DIR, req.Filename)
	content, err := os.ReadFile(filepath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lines := strings.Split(string(content), "\n")
	if req.LineIndex < 0 || req.LineIndex >= len(lines) {
		http.Error(w, "Invalid line index", http.StatusBadRequest)
		return
	}

	lines = append(lines[:req.LineIndex], lines[req.LineIndex+1:]...)
	err = os.WriteFile(filepath, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
