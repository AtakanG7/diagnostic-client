package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"diagnostic-client/internal/db"
)

type Handler struct {
	db *db.DB
}

func NewHandler(db *db.DB) *Handler {
	return &Handler{db: db}
}

func normalizePath(path string) string {
	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Remove trailing slash unless it's the root path
	if path != "/" && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	return path
}

// internal/api/handler.go
func (h *Handler) GetFiles(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	} else {
		path = normalizePath(path)
	}

	// Get depth from query params, default to 1 if not specified
	depth := 1
	if depthStr := r.URL.Query().Get("depth"); depthStr != "" {
		if d, err := strconv.Atoi(depthStr); err == nil {
			depth = d
		}
	}

	// Limit maximum depth to prevent excessive recursion
	if depth > 10 {
		depth = 10
	}

	log.Printf("[API] Getting file tree for path: %s with depth: %d", path, depth)

	files, err := h.db.GetFileTree(r.Context(), path, depth)
	if err != nil {
		log.Printf("[API] Error getting file tree: %v", err)
		http.Error(w, fmt.Sprintf("Error getting file tree: %v", err), http.StatusInternalServerError)
		return
	}

	if len(files) == 0 {
		// Return empty array instead of null
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	log.Printf("[API] Found %d files at path: %s", len(files), path)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(files); err != nil {
		log.Printf("[API] Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetLogs(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("file")
	if filePath == "" {
		http.Error(w, "file parameter required", http.StatusBadRequest)
		return
	}

	beforeStr := r.URL.Query().Get("before")
	before := time.Now()
	if beforeStr != "" {
		var err error
		before, err = time.Parse(time.RFC3339, beforeStr)
		if err != nil {
			http.Error(w, "invalid before time", http.StatusBadRequest)
			return
		}
	}

	logs, err := h.db.GetLogs(r.Context(), filePath, before, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(logs)
}

func (h *Handler) SearchLogs(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query     string    `json:"query"`
		Files     []string  `json:"files"`
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logs, err := h.db.SearchLogs(r.Context(), req.Query, req.Files, req.StartTime, req.EndTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(logs)
}

func (h *Handler) GetNetworkMetrics(w http.ResponseWriter, r *http.Request) {
	var startTime, endTime time.Time
	var err error

	startStr := r.URL.Query().Get("start")
	if startStr != "" {
		startTime, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			http.Error(w, "invalid start time", http.StatusBadRequest)
			return
		}
	}

	endStr := r.URL.Query().Get("end")
	if endStr != "" {
		endTime, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			http.Error(w, "invalid end time", http.StatusBadRequest)
			return
		}
	}

	protocols := r.URL.Query()["protocol"]

	packets, err := h.db.GetNetworkPackets(r.Context(), startTime, endTime, protocols)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(packets)
}
