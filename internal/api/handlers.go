package api

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"markdown-editor/internal/auth"
	"markdown-editor/internal/models"
	"markdown-editor/internal/services/store"
	"markdown-editor/internal/static"

	"github.com/gorilla/mux"
)

type Handler struct {
	store     store.Store
	userStore store.UserStore
}

func NewHandler(s store.Store, us store.UserStore) *Handler {
	return &Handler{
		store:     s,
		userStore: us,
	}
}

// Add CORS middleware
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

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
	// Public route to check auth status
	r.HandleFunc("/auth/status", h.authStatus).Methods("GET", "OPTIONS")

	// Only register auth routes if we have a user store (database backend)
	if h.userStore != nil {
		// Public routes (no authentication required)
		r.HandleFunc("/auth/signup", h.signup).Methods("POST", "OPTIONS")
		r.HandleFunc("/auth/login", h.login).Methods("POST", "OPTIONS")
	}

	// Legacy routes (for backward compatibility with markdown store)
	if h.store != nil {
		legacyRouter := r.PathPrefix("").Subrouter()
		legacyRouter.HandleFunc("/files", h.listFiles).Methods("GET", "OPTIONS")
		legacyRouter.HandleFunc("/file/{filename}", h.getFile).Methods("GET", "OPTIONS")
		legacyRouter.HandleFunc("/delete_links", h.deleteLinks).Methods("POST", "OPTIONS")
		legacyRouter.HandleFunc("/bulk_links", h.addBulkLinks).Methods("POST", "OPTIONS")
	}

	// Protected routes (require authentication and database backend)
	if h.userStore == nil {
		return
	}

	protected := r.PathPrefix("/v2").Subrouter()
	protected.Use(auth.GetAuthMiddleware())

	// Subject routes
	protected.HandleFunc("/subjects", h.listSubjects).Methods("GET", "OPTIONS")
	protected.HandleFunc("/subjects", h.createSubject).Methods("POST", "OPTIONS")
	protected.HandleFunc("/subjects/{id}", h.getSubject).Methods("GET", "OPTIONS")
	protected.HandleFunc("/subjects/{id}", h.deleteSubject).Methods("DELETE", "OPTIONS")

	// Topic routes (name-based - NEW)
	protected.HandleFunc("/subjects/{subjectName}/topics", h.listTopicsByName).Methods("GET", "OPTIONS")
	protected.HandleFunc("/subjects/{subjectName}/topics", h.createTopicByName).Methods("POST", "OPTIONS")

	// Link routes (name-based - NEW)
	// Use {topicName:.+} to match topic names with special characters including /
	protected.HandleFunc("/subjects/{subjectName}/topics/{topicName:.+}/links", h.listLinksByName).Methods("GET", "OPTIONS")
	protected.HandleFunc("/subjects/{subjectName}/topics/{topicName:.+}/links", h.createLinkByName).Methods("POST", "OPTIONS")
	protected.HandleFunc("/subjects/{subjectName}/topics/{topicName:.+}/links/bulk", h.bulkCreateLinksByName).Methods("POST", "OPTIONS")
	protected.HandleFunc("/links/{id}", h.deleteLink).Methods("DELETE", "OPTIONS")

	// Legacy ID-based routes (kept for backwards compatibility, can be removed later)
	protected.HandleFunc("/subjects/{subjectId:[0-9]+}/topics", h.listTopics).Methods("GET", "OPTIONS")
	protected.HandleFunc("/subjects/{subjectId:[0-9]+}/topics", h.createTopic).Methods("POST", "OPTIONS")
	protected.HandleFunc("/topics/{topicId:[0-9]+}/links", h.listLinks).Methods("GET", "OPTIONS")

	// Tag routes
	protected.HandleFunc("/tags", h.listTags).Methods("GET", "OPTIONS")
	protected.HandleFunc("/tags", h.createTag).Methods("POST", "OPTIONS")
	protected.HandleFunc("/links/{linkId}/tags/{tagId}", h.addTagToLink).Methods("POST", "OPTIONS")
	protected.HandleFunc("/links/{linkId}/tags/{tagId}", h.removeTagFromLink).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/links/{linkId}/tags", h.getLinkTags).Methods("GET", "OPTIONS")
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

	// Return the legacy subjects directly
	var response models.Response
	response.Data = storeSubjects
	json.NewEncoder(w).Encode(response)
}

// Convert DeleteLink to models.LegacyLink
func (h *Handler) toStoreLinks(links []models.DeleteLink) []models.LegacyLink {
	storeLinks := make([]models.LegacyLink, len(links))
	for i, link := range links {
		storeLinks[i] = models.LegacyLink{
			ID:    link.ID,
			Title: link.Title,
			URL:   link.URL,
		}
	}
	return storeLinks
}

// Convert BulkLink to models.LegacyLink
func (h *Handler) toBulkStoreLinks(links []models.BulkLink) []models.LegacyLink {
	storeLinks := make([]models.LegacyLink, len(links))
	for i, link := range links {
		storeLinks[i] = models.LegacyLink{
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

// Authentication handlers

func (h *Handler) authStatus(w http.ResponseWriter, r *http.Request) {
	disableAuth := os.Getenv("DISABLE_AUTH") == "true"
	response := map[string]interface{}{
		"authEnabled": !disableAuth,
		"defaultUser": map[string]string{
			"email": auth.DefaultEmail,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) signup(w http.ResponseWriter, r *http.Request) {
	var req models.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Create user
	user, err := h.userStore.CreateUser(req.Email, hashedPassword)
	if err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return response
	response := models.AuthResponse{
		Token: token,
		User:  *user,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get user by email
	user, err := h.userStore.GetUserByEmail(req.Email)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Skip password check if auth is disabled (local development only)
	disableAuth := os.Getenv("DISABLE_AUTH") == "true"
	if !disableAuth {
		// Check password
		if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return response
	response := models.AuthResponse{
		Token: token,
		User:  *user,
	}
	json.NewEncoder(w).Encode(response)
}

// Subject handlers

func (h *Handler) listSubjects(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	subjects, err := h.userStore.ListSubjects(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(subjects)
}

func (h *Handler) createSubject(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	subject, err := h.userStore.CreateSubject(userID, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(subject)
}

func (h *Handler) getSubject(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	vars := mux.Vars(r)
	id := vars["id"]

	var subjectID int
	if _, err := fmt.Sscanf(id, "%d", &subjectID); err != nil {
		http.Error(w, "Invalid subject ID", http.StatusBadRequest)
		return
	}

	subject, err := h.userStore.GetSubject(subjectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(subject)
}

func (h *Handler) deleteSubject(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	vars := mux.Vars(r)
	id := vars["id"]

	var subjectID int
	if _, err := fmt.Sscanf(id, "%d", &subjectID); err != nil {
		http.Error(w, "Invalid subject ID", http.StatusBadRequest)
		return
	}

	if err := h.userStore.DeleteSubject(subjectID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// Topic handlers

func (h *Handler) listTopics(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	vars := mux.Vars(r)
	subjectID := vars["subjectId"]

	var sid int
	if _, err := fmt.Sscanf(subjectID, "%d", &sid); err != nil {
		http.Error(w, "Invalid subject ID", http.StatusBadRequest)
		return
	}

	topics, err := h.userStore.ListTopics(sid, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(topics)
}

func (h *Handler) createTopic(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	vars := mux.Vars(r)
	subjectID := vars["subjectId"]

	var sid int
	if _, err := fmt.Sscanf(subjectID, "%d", &sid); err != nil {
		http.Error(w, "Invalid subject ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	topic, err := h.userStore.CreateTopic(sid, req.Name, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(topic)
}

func (h *Handler) getTopic(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	vars := mux.Vars(r)
	id := vars["id"]

	var topicID int
	if _, err := fmt.Sscanf(id, "%d", &topicID); err != nil {
		http.Error(w, "Invalid topic ID", http.StatusBadRequest)
		return
	}

	topic, err := h.userStore.GetTopic(topicID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(topic)
}

func (h *Handler) deleteTopic(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	vars := mux.Vars(r)
	id := vars["id"]

	var topicID int
	if _, err := fmt.Sscanf(id, "%d", &topicID); err != nil {
		http.Error(w, "Invalid topic ID", http.StatusBadRequest)
		return
	}

	if err := h.userStore.DeleteTopic(topicID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// Link handlers

func (h *Handler) listLinks(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	vars := mux.Vars(r)
	topicID := vars["topicId"]

	var tid int
	if _, err := fmt.Sscanf(topicID, "%d", &tid); err != nil {
		http.Error(w, "Invalid topic ID", http.StatusBadRequest)
		return
	}

	links, err := h.userStore.ListLinks(tid, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(links)
}

func (h *Handler) createLink(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	vars := mux.Vars(r)
	topicID := vars["topicId"]

	var tid int
	if _, err := fmt.Sscanf(topicID, "%d", &tid); err != nil {
		http.Error(w, "Invalid topic ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	link, err := h.userStore.CreateLink(tid, req.Title, req.URL, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(link)
}

func (h *Handler) bulkCreateLinks(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	vars := mux.Vars(r)
	topicID := vars["topicId"]

	var tid int
	if _, err := fmt.Sscanf(topicID, "%d", &tid); err != nil {
		http.Error(w, "Invalid topic ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Links []models.Link `json:"links"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.userStore.BulkCreateLinks(tid, req.Links, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) getLink(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	vars := mux.Vars(r)
	id := vars["id"]

	var linkID int
	if _, err := fmt.Sscanf(id, "%d", &linkID); err != nil {
		http.Error(w, "Invalid link ID", http.StatusBadRequest)
		return
	}

	link, err := h.userStore.GetLink(linkID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(link)
}

func (h *Handler) deleteLink(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	vars := mux.Vars(r)
	id := vars["id"]

	var linkID int
	if _, err := fmt.Sscanf(id, "%d", &linkID); err != nil {
		http.Error(w, "Invalid link ID", http.StatusBadRequest)
		return
	}

	if err := h.userStore.DeleteLink(linkID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// Tag handlers

func (h *Handler) listTags(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	tags, err := h.userStore.ListTags(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tags)
}

func (h *Handler) createTag(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tag, err := h.userStore.CreateTag(req.Name, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tag)
}

func (h *Handler) addTagToLink(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	vars := mux.Vars(r)

	var linkID, tagID int
	if _, err := fmt.Sscanf(vars["linkId"], "%d", &linkID); err != nil {
		http.Error(w, "Invalid link ID", http.StatusBadRequest)
		return
	}
	if _, err := fmt.Sscanf(vars["tagId"], "%d", &tagID); err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	if err := h.userStore.AddTagToLink(linkID, tagID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) removeTagFromLink(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	vars := mux.Vars(r)

	var linkID, tagID int
	if _, err := fmt.Sscanf(vars["linkId"], "%d", &linkID); err != nil {
		http.Error(w, "Invalid link ID", http.StatusBadRequest)
		return
	}
	if _, err := fmt.Sscanf(vars["tagId"], "%d", &tagID); err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	if err := h.userStore.RemoveTagFromLink(linkID, tagID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) getLinkTags(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	vars := mux.Vars(r)

	var linkID int
	if _, err := fmt.Sscanf(vars["linkId"], "%d", &linkID); err != nil {
		http.Error(w, "Invalid link ID", http.StatusBadRequest)
		return
	}

	tags, err := h.userStore.GetLinkTags(linkID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tags)
}

// Name-based handlers (NEW - no IDs needed!)

func (h *Handler) listTopicsByName(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	vars := mux.Vars(r)
	subjectName := vars["subjectName"]

	// Find subject by name
	subjects, err := h.userStore.ListSubjects(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var subjectID int
	found := false
	for _, s := range subjects {
		if s.Name == subjectName {
			subjectID = s.ID
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "Subject not found", http.StatusNotFound)
		return
	}

	// List topics for this subject
	topics, err := h.userStore.ListTopics(subjectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(topics)
}

func (h *Handler) createTopicByName(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	vars := mux.Vars(r)
	subjectName := vars["subjectName"]

	// Find subject by name
	subjects, err := h.userStore.ListSubjects(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var subjectID int
	found := false
	for _, s := range subjects {
		if s.Name == subjectName {
			subjectID = s.ID
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "Subject not found", http.StatusNotFound)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	topic, err := h.userStore.CreateTopic(subjectID, req.Name, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(topic)
}

func (h *Handler) listLinksByName(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	vars := mux.Vars(r)
	subjectName := vars["subjectName"]
	topicName := vars["topicName"]

	// URL decode the names (in case they contain special characters like /)
	// Note: Gorilla Mux already decodes, but let's be explicit
	// subjectName, _ = url.QueryUnescape(subjectName)
	// topicName, _ = url.QueryUnescape(topicName)

	// Find subject by name
	subjects, err := h.userStore.ListSubjects(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var subjectID int
	found := false
	for _, s := range subjects {
		if s.Name == subjectName {
			subjectID = s.ID
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "Subject not found", http.StatusNotFound)
		return
	}

	// Find topic by name
	topics, err := h.userStore.ListTopics(subjectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var topicID int
	found = false
	for _, t := range topics {
		if t.Name == topicName {
			topicID = t.ID
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "Topic not found", http.StatusNotFound)
		return
	}

	// List links for this topic
	links, err := h.userStore.ListLinks(topicID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(links)
}

func (h *Handler) createLinkByName(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	vars := mux.Vars(r)
	subjectName := vars["subjectName"]
	topicName := vars["topicName"]

	// Gorilla Mux automatically URL-decodes path parameters

	// Find subject and topic (same logic as listLinksByName)
	subjects, err := h.userStore.ListSubjects(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var subjectID int
	found := false
	for _, s := range subjects {
		if s.Name == subjectName {
			subjectID = s.ID
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "Subject not found", http.StatusNotFound)
		return
	}

	topics, err := h.userStore.ListTopics(subjectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var topicID int
	found = false
	for _, t := range topics {
		if t.Name == topicName {
			topicID = t.ID
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "Topic not found", http.StatusNotFound)
		return
	}

	var req struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	link, err := h.userStore.CreateLink(topicID, req.Title, req.URL, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(link)
}

func (h *Handler) bulkCreateLinksByName(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)
	vars := mux.Vars(r)
	subjectName := vars["subjectName"]
	topicName := vars["topicName"]

	// Gorilla Mux automatically URL-decodes path parameters

	// Find subject and topic
	subjects, err := h.userStore.ListSubjects(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var subjectID int
	found := false
	for _, s := range subjects {
		if s.Name == subjectName {
			subjectID = s.ID
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "Subject not found", http.StatusNotFound)
		return
	}

	topics, err := h.userStore.ListTopics(subjectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var topicID int
	found = false
	for _, t := range topics {
		if t.Name == topicName {
			topicID = t.ID
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "Topic not found", http.StatusNotFound)
		return
	}

	var req struct {
		Links []models.Link `json:"links"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.userStore.BulkCreateLinks(topicID, req.Links, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
