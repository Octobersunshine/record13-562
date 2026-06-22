package store

import (
	"decoration/models"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type PhaseStore interface {
	CreateProject(projectID string) ([]models.ConstructionPhase, error)
	GetPhases(projectID string) ([]models.ConstructionPhase, error)
	GetPhase(projectID, phaseType string) (*models.ConstructionPhase, error)
	UpdatePhase(projectID, phaseType string, status models.PhaseStatus) (*models.ConstructionPhase, error)
}

type memoryStore struct {
	mu     sync.RWMutex
	phases map[string][]models.ConstructionPhase
}

func NewMemoryStore() PhaseStore {
	return &memoryStore{
		phases: make(map[string][]models.ConstructionPhase),
	}
}

func (s *memoryStore) CreateProject(projectID string) ([]models.ConstructionPhase, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.phases[projectID]; exists {
		return nil, errors.New("project already exists")
	}

	now := time.Now()
	phaseTypes := []models.PhaseType{
		models.PhasePlumbingElectrical,
		models.PhaseMasonry,
		models.PhasePainting,
	}

	phases := make([]models.ConstructionPhase, 0, len(phaseTypes))
	for _, pt := range phaseTypes {
		phases = append(phases, models.ConstructionPhase{
			ID:        uuid.NewString(),
			ProjectID: projectID,
			PhaseType: pt,
			PhaseName: models.PhaseNames[pt],
			Status:    models.StatusPending,
			UpdatedAt: now,
			CreatedAt: now,
		})
	}

	s.phases[projectID] = phases
	return phases, nil
}

func (s *memoryStore) GetPhases(projectID string) ([]models.ConstructionPhase, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	phases, exists := s.phases[projectID]
	if !exists {
		return nil, errors.New("project not found")
	}

	result := make([]models.ConstructionPhase, len(phases))
	copy(result, phases)
	return result, nil
}

func (s *memoryStore) GetPhase(projectID, phaseType string) (*models.ConstructionPhase, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	phases, exists := s.phases[projectID]
	if !exists {
		return nil, errors.New("project not found")
	}

	for _, phase := range phases {
		if string(phase.PhaseType) == phaseType {
			return &phase, nil
		}
	}

	return nil, errors.New("phase not found")
}

func (s *memoryStore) UpdatePhase(projectID, phaseType string, status models.PhaseStatus) (*models.ConstructionPhase, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	phases, exists := s.phases[projectID]
	if !exists {
		return nil, errors.New("project not found")
	}

	targetIdx := -1
	for i := range phases {
		if string(phases[i].PhaseType) == phaseType {
			targetIdx = i
			break
		}
	}
	if targetIdx == -1 {
		return nil, errors.New("phase not found")
	}

	if status == models.StatusInProgress || status == models.StatusCompleted {
		prereq, hasPrereq := models.PhasePrerequisites[models.PhaseType(phaseType)]
		if hasPrereq {
			prereqCompleted := false
			for i := range phases {
				if phases[i].PhaseType == prereq && phases[i].Status == models.StatusCompleted {
					prereqCompleted = true
					break
				}
			}
			if !prereqCompleted {
				return nil, fmt.Errorf("前序阶段[%s]尚未完成，无法更新阶段[%s]",
					models.PhaseNames[prereq], models.PhaseNames[models.PhaseType(phaseType)])
			}
		}
	}

	now := time.Now()
	phases[targetIdx].Status = status
	phases[targetIdx].UpdatedAt = now
	if status == models.StatusCompleted {
		phases[targetIdx].CompletedAt = &now
	} else {
		phases[targetIdx].CompletedAt = nil
	}
	s.phases[projectID] = phases
	updated := phases[targetIdx]
	return &updated, nil
}
