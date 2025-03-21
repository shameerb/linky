package api

import (
	"encoding/json"
	"io/fs"
	"net/http"

	"markdown-editor/internal/models"
	"markdown-editor/internal/services/store"
	"markdown-editor/internal/static"

	"github.com/gorilla/mux"
)

type Handler struct {
	store store.Store
}

func NewHandler(s store.Store) *Handler {
	return &Handler{store: s}
}

// Add CORS middleware
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	// Don't apply CORS middleware here since it's already applied to the main router
	r.HandleFunc("/files", h.listFiles).Methods("GET", "OPTIONS")
	r.HandleFunc("/file/{filename}", h.getFile).Methods("GET", "OPTIONS")
	r.HandleFunc("/delete_links", h.deleteLinks).Methods("POST", "OPTIONS")
	r.HandleFunc("/bulk_links", h.addBulkLinks).Methods("POST", "OPTIONS")
}

func (h *Handler) StaticFiles() fs.FS {
	return static.Files()
}

func (h *Handler) listFiles(w http.ResponseWriter, r *http.Request) {
	files, err := h.store.ListFiles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(files)
}

func (h *Handler) getFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]
	storeSubjects, err := h.store.GetLinks(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert store.Subject to local Subject
	subjects := make([]models.Subject, len(storeSubjects))
	for i, s := range storeSubjects {
		subjects[i] = models.Subject{
			Subject: s.Subject,
			Links:   make([]models.Link, len(s.Links)),
		}
		for j, l := range s.Links {
			subjects[i].Links[j] = models.Link{
				ID:        l.ID,
				Title:     l.Title,
				URL:       l.URL,
				Timestamp: l.Timestamp,
			}
		}
	}

	var response models.Response
	response.Data = subjects
	json.NewEncoder(w).Encode(response)
}

// Convert DeleteLink to models.Link
func (h *Handler) toStoreLinks(links []models.DeleteLink) []models.Link {
	storeLinks := make([]models.Link, len(links))
	for i, link := range links {
		storeLinks[i] = models.Link{
			ID:    link.ID,
			Title: link.Title,
			URL:   link.URL,
		}
	}
	return storeLinks
}

// Convert BulkLink to models.Link
func (h *Handler) toBulkStoreLinks(links []models.BulkLink) []models.Link {
	storeLinks := make([]models.Link, len(links))
	for i, link := range links {
		storeLinks[i] = models.Link{
			Title: link.Title,
			URL:   link.URL,
		}
	}
	return storeLinks
}

func (h *Handler) deleteLinks(w http.ResponseWriter, r *http.Request) {
	var req models.DeleteLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.store.DeleteLinks(req.Filename, h.toStoreLinks(req.Links))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) addBulkLinks(w http.ResponseWriter, r *http.Request) {
	var req models.BulkLinksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.store.AddBulkLinks(req.Filename, req.Subject, h.toBulkStoreLinks(req.Links))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
