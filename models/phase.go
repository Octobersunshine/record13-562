package models

import "time"

type PhaseType string

const (
	PhasePlumbingElectrical PhaseType = "plumbing_electrical"
	PhaseMasonry            PhaseType = "masonry"
	PhasePainting           PhaseType = "painting"
)

var PhaseNames = map[PhaseType]string{
	PhasePlumbingElectrical: "水电",
	PhaseMasonry:            "泥瓦",
	PhasePainting:           "油漆",
}

var PhaseOrder = []PhaseType{
	PhasePlumbingElectrical,
	PhaseMasonry,
	PhasePainting,
}

var PhasePrerequisites = map[PhaseType]PhaseType{
	PhaseMasonry:  PhasePlumbingElectrical,
	PhasePainting: PhaseMasonry,
}

type PhaseStatus string

const (
	StatusPending   PhaseStatus = "pending"
	StatusInProgress PhaseStatus = "in_progress"
	StatusCompleted PhaseStatus = "completed"
)

type ConstructionPhase struct {
	ID          string       `json:"id"`
	ProjectID   string       `json:"project_id"`
	PhaseType   PhaseType    `json:"phase_type"`
	PhaseName   string       `json:"phase_name"`
	Status      PhaseStatus  `json:"status"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	UpdatedAt   time.Time    `json:"updated_at"`
	CreatedAt   time.Time    `json:"created_at"`
}

type UpdatePhaseRequest struct {
	Status PhaseStatus `json:"status"`
}
