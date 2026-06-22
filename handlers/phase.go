package handlers

import (
	"decoration/models"
	"decoration/store"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type PhaseHandler struct {
	store store.PhaseStore
}

func NewPhaseHandler(s store.PhaseStore) *PhaseHandler {
	return &PhaseHandler{store: s}
}

func (h *PhaseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/phases")
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	switch r.Method {
	case http.MethodPost:
		if len(parts) >= 1 && parts[0] != "" {
			h.createProject(w, r, parts[0])
			return
		}
		http.Error(w, "project_id required", http.StatusBadRequest)
	case http.MethodGet:
		if len(parts) >= 2 {
			h.getPhase(w, r, parts[0], parts[1])
			return
		}
		if len(parts) >= 1 && parts[0] != "" {
			h.getPhases(w, r, parts[0])
			return
		}
		http.Error(w, "project_id required", http.StatusBadRequest)
	case http.MethodPut, http.MethodPatch:
		if len(parts) >= 3 && parts[2] == "planned_date" {
			h.setPlannedDate(w, r, parts[0], parts[1])
			return
		}
		if len(parts) >= 2 {
			h.updatePhase(w, r, parts[0], parts[1])
			return
		}
		http.Error(w, "project_id and phase_type required", http.StatusBadRequest)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *PhaseHandler) createProject(w http.ResponseWriter, r *http.Request, projectID string) {
	phases, err := h.store.CreateProject(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	writeJSON(w, http.StatusCreated, phases)
}

func (h *PhaseHandler) getPhases(w http.ResponseWriter, r *http.Request, projectID string) {
	phases, err := h.store.GetPhases(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, phases)
}

func (h *PhaseHandler) getPhase(w http.ResponseWriter, r *http.Request, projectID, phaseType string) {
	phase, err := h.store.GetPhase(projectID, phaseType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, phase)
}

func (h *PhaseHandler) updatePhase(w http.ResponseWriter, r *http.Request, projectID, phaseType string) {
	var req models.UpdatePhaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Status != models.StatusPending &&
		req.Status != models.StatusInProgress &&
		req.Status != models.StatusCompleted {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	phase, err := h.store.UpdatePhase(projectID, phaseType, req.Status)
	if err != nil {
		if isValidationError(err) {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, phase)
}

func isValidationError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "前序阶段")
}

func (h *PhaseHandler) setPlannedDate(w http.ResponseWriter, r *http.Request, projectID, phaseType string) {
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

	phase, err := h.store.SetPlannedEndDate(projectID, phaseType, parsed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, phase)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
