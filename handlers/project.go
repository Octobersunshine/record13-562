package handlers

import (
	"decoration/models"
	"decoration/store"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type ProjectHandler struct {
	store store.PhaseStore
}

func NewProjectHandler(s store.PhaseStore) *ProjectHandler {
	return &ProjectHandler{store: s}
}

func (h *ProjectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/projects")
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	projectID := parts[0]
	action := parts[1]

	switch r.Method {
	case http.MethodPut, http.MethodPatch:
		switch action {
		case "foreman":
			h.setForeman(w, r, projectID)
			return
		}
	case http.MethodGet:
		switch action {
		case "foreman":
			h.getForeman(w, r, projectID)
			return
		case "alerts":
			h.getAlerts(w, r, projectID)
			return
		}
	}
	http.Error(w, "not found", http.StatusNotFound)
}

func (h *ProjectHandler) setForeman(w http.ResponseWriter, r *http.Request, projectID string) {
	var req models.SetForemanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.ForemanName == "" {
		http.Error(w, "foreman_name is required", http.StatusBadRequest)
		return
	}

	info, err := h.store.SetForeman(projectID, req.ForemanName, req.ForemanPhone)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, info)
}

func (h *ProjectHandler) getForeman(w http.ResponseWriter, r *http.Request, projectID string) {
	info, err := h.store.GetForeman(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if info == nil {
		http.Error(w, "foreman not configured", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, info)
}

func (h *ProjectHandler) getAlerts(w http.ResponseWriter, r *http.Request, projectID string) {
	alerts := h.store.GetAlerts(projectID)
	if alerts == nil {
		alerts = []models.PhaseAlert{}
	}
	writeJSON(w, http.StatusOK, alerts)
}

func HandlePhasePlannedDate(store store.PhaseStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut && r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/api/phases")
		path = strings.Trim(path, "/")
		parts := strings.Split(path, "/")
		if len(parts) < 3 || parts[2] != "planned_date" {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		projectID := parts[0]
		phaseType := parts[1]

		var req models.SetPlannedDateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.PlannedEndDate == "" {
			http.Error(w, "planned_end_date is required", http.StatusBadRequest)
			return
		}
		parsed, err := time.Parse(time.RFC3339, req.PlannedEndDate)
		if err != nil {
			parsed, err = time.Parse("2006-01-02", req.PlannedEndDate)
			if err != nil {
				http.Error(w, "invalid planned_end_date format, use RFC3339 or YYYY-MM-DD", http.StatusBadRequest)
				return
			}
		}

		phase, err := store.SetPlannedEndDate(projectID, phaseType, parsed)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, phase)
	}
}
