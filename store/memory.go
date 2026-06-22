package store

import (
	"decoration/models"
	"decoration/notify"
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
	SetPlannedEndDate(projectID, phaseType string, plannedEndDate time.Time) (*models.ConstructionPhase, error)
	SetForeman(projectID, foremanName, foremanPhone string) (*models.ProjectInfo, error)
	GetForeman(projectID string) (*models.ProjectInfo, error)
	FindOverduePhases(now time.Time) []OverduePhase
	RecordAlert(alert *models.PhaseAlert)
	GetAlerts(projectID string) []models.PhaseAlert
}

type OverduePhase struct {
	ProjectID string
	PhaseType models.PhaseType
	PhaseName string
	OverdueDays int
}

type memoryStore struct {
	mu        sync.RWMutex
	phases    map[string][]models.ConstructionPhase
	foremen   map[string]models.ProjectInfo
	alerts    map[string][]models.PhaseAlert
	alertKeys map[string]time.Time
	notifier  notify.Notifier
}

func NewMemoryStore() PhaseStore {
	return &memoryStore{
		phases:    make(map[string][]models.ConstructionPhase),
		foremen:   make(map[string]models.ProjectInfo),
		alerts:    make(map[string][]models.PhaseAlert),
		alertKeys: make(map[string]time.Time),
		notifier:  notify.NewLogNotifier(),
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

func (s *memoryStore) SetPlannedEndDate(projectID, phaseType string, plannedEndDate time.Time) (*models.ConstructionPhase, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	phases, exists := s.phases[projectID]
	if !exists {
		return nil, errors.New("project not found")
	}

	now := time.Now()
	for i := range phases {
		if string(phases[i].PhaseType) == phaseType {
			phases[i].PlannedEndDate = &plannedEndDate
			phases[i].UpdatedAt = now
			s.phases[projectID] = phases
			updated := phases[i]
			return &updated, nil
		}
	}
	return nil, errors.New("phase not found")
}

func (s *memoryStore) SetForeman(projectID, foremanName, foremanPhone string) (*models.ProjectInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.phases[projectID]; !exists {
		return nil, errors.New("project not found")
	}

	now := time.Now()
	if existing, ok := s.foremen[projectID]; ok {
		existing.ForemanName = foremanName
		existing.ForemanPhone = foremanPhone
		existing.UpdatedAt = now
		s.foremen[projectID] = existing
		result := existing
		return &result, nil
	}

	info := models.ProjectInfo{
		ProjectID:    projectID,
		ForemanName:  foremanName,
		ForemanPhone: foremanPhone,
		UpdatedAt:    now,
		CreatedAt:    now,
	}
	s.foremen[projectID] = info
	result := info
	return &result, nil
}

func (s *memoryStore) GetForeman(projectID string) (*models.ProjectInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.phases[projectID]; !exists {
		return nil, errors.New("project not found")
	}

	info, ok := s.foremen[projectID]
	if !ok {
		return nil, nil
	}
	result := info
	return &result, nil
}

func (s *memoryStore) FindOverduePhases(now time.Time) []OverduePhase {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []OverduePhase
	for projectID, phases := range s.phases {
		for _, phase := range phases {
			if phase.Status == models.StatusCompleted {
				continue
			}
			if phase.PlannedEndDate == nil {
				continue
			}
			if now.After(*phase.PlannedEndDate) {
				diff := now.Sub(*phase.PlannedEndDate)
				days := int(diff.Hours() / 24)
				if days < 1 {
					days = 1
				}
				result = append(result, OverduePhase{
					ProjectID:   projectID,
					PhaseType:   phase.PhaseType,
					PhaseName:   phase.PhaseName,
					OverdueDays: days,
				})
			}
		}
	}
	return result
}

func (s *memoryStore) RecordAlert(alert *models.PhaseAlert) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.alerts[alert.ProjectID] = append(s.alerts[alert.ProjectID], *alert)

	key := fmt.Sprintf("%s:%s", alert.ProjectID, alert.PhaseType)
	s.alertKeys[key] = alert.NotifiedAt

	foreman, _ := s.foremen[alert.ProjectID]
	f := foreman
	s.notifier.Send(&f, alert)
}

func (s *memoryStore) GetAlerts(projectID string) []models.PhaseAlert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	src := s.alerts[projectID]
	result := make([]models.PhaseAlert, len(src))
	copy(result, src)
	return result
}
