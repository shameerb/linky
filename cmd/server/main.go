package main

import (
	"log"
	"net/http"
	"os"

	"markdown-editor/internal/api"
	"markdown-editor/internal/services/store/markdown"

	"github.com/gorilla/mux"
)

var (
	markdownDir string
	isProd      bool
)

func main() {
	r := mux.NewRouter()

	// Check if we're in production mode
	isProd = os.Getenv("GO_ENV") == "production"

	markdownDir = os.Getenv("MARKDOWN_DIR")
	if markdownDir == "" {
		log.Fatal("MARKDOWN_DIR environment variable is required")
	}

	// Initialize the store
	mdStore, err := markdown.New(markdownDir)
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}

	// Initialize handlers
	handler := api.NewHandler(mdStore)

	// Apply CORS middleware to the entire router
	r.Use(api.CORSMiddleware)

	// API routes
	apiRouter := r.PathPrefix("/api").Subrouter()
	handler.RegisterRoutes(apiRouter)

	// Serve static files
	r.PathPrefix("/").Handler(http.FileServer(http.FS(handler.StaticFiles())))

	log.Printf("Using markdown directory: %s", markdownDir)
	log.Fatal(http.ListenAndServe(":8080", r))
}
