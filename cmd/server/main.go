package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"markdown-editor/internal/api"
	"markdown-editor/internal/auth"
	"markdown-editor/internal/services/store"

	"github.com/gorilla/mux"
)

var (
	markdownDir string
	dbPath      string
	jwtSecret   string
	isProd      bool
	port        string
)

// spaHandler implements the http.Handler interface for SPA routing
type spaHandler struct {
	staticFS   fs.FS
	indexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// If a file is found, it will be served. If not, the file located at the index
// path on the SPA will be served.
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the absolute path to prevent directory traversal
	path := r.URL.Path

	// Remove leading slash for fs.FS compatibility
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		path = "index.html"
	}

	// Try to open the file
	f, err := h.staticFS.Open(path)
	if err == nil {
		// File exists, serve it
		defer f.Close()

		// Get file info to check if it's a directory
		stat, err := f.Stat()
		if err == nil && !stat.IsDir() {
			// Set content type based on file extension
			ext := filepath.Ext(path)
			switch ext {
			case ".html":
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
			case ".css":
				w.Header().Set("Content-Type", "text/css; charset=utf-8")
			case ".js":
				w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			case ".json":
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
			case ".png":
				w.Header().Set("Content-Type", "image/png")
			case ".jpg", ".jpeg":
				w.Header().Set("Content-Type", "image/jpeg")
			case ".svg":
				w.Header().Set("Content-Type", "image/svg+xml")
			}

			// Copy file content to response
			w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
			io.Copy(w, f)
			return
		}
	}

	// File doesn't exist or is a directory, serve index.html for SPA routing
	indexFile, err := h.staticFS.Open(h.indexPath)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	defer indexFile.Close()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.Copy(w, indexFile)
}

func main() {
	r := mux.NewRouter()

	// Check if we're in production mode
	isProd = os.Getenv("GO_ENV") == "production"

	// Get configuration from environment variables
	markdownDir = os.Getenv("MARKDOWN_DIR")
	if markdownDir == "" {
		markdownDir = "./Links" // Default to ./Links if not set
	}

	dbPath = os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./linky.db" // Default database path
	}

	jwtSecret = os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-this-in-production"
		log.Println("WARNING: Using default JWT secret. Set JWT_SECRET environment variable in production!")
	}

	port = os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	// Set JWT secret
	auth.SetJWTSecret(jwtSecret)

	// Check if auth is disabled
	disableAuth := os.Getenv("DISABLE_AUTH") == "true"
	if disableAuth {
		log.Println("⚠️  WARNING: Authentication is DISABLED. Using default user: shameer789@gmail.com")
		log.Println("⚠️  This should ONLY be used for local development!")
	}

	// Initialize store based on configuration
	// Use STORAGE_TYPE environment variable: "file", "sqlite", or "postgres"
	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = "file"
	}

	legacyStore, userStore, err := store.NewStoreFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}

	// Log storage configuration
	log.Printf("Storage backend: %s", storageType)
	if storageType == "file" {
		log.Printf("Using markdown directory: %s", markdownDir)
	} else {
		if storageType == "sqlite" {
			log.Printf("Using SQLite database: %s", dbPath)
		}
		// For postgres, would log DATABASE_URL here
	}

	// Close database connection if using database store
	if closer, ok := userStore.(interface{ Close() error }); ok {
		defer closer.Close()
	}

	// Initialize handlers
	handler := api.NewHandler(legacyStore, userStore)

	// Apply CORS middleware to the entire router
	r.Use(api.CORSMiddleware)

	// API routes
	apiRouter := r.PathPrefix("/api").Subrouter()
	handler.RegisterRoutes(apiRouter)

	// Serve static files with SPA fallback
	spa := spaHandler{
		staticFS:  handler.StaticFiles(),
		indexPath: "index.html",
	}
	r.PathPrefix("/").Handler(spa)

	log.Printf("Server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
