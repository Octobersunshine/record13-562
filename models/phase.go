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
	ID             string       `json:"id"`
	ProjectID      string       `json:"project_id"`
	PhaseType      PhaseType    `json:"phase_type"`
	PhaseName      string       `json:"phase_name"`
	Status         PhaseStatus  `json:"status"`
	PlannedEndDate *time.Time   `json:"planned_end_date,omitempty"`
	CompletedAt    *time.Time   `json:"completed_at,omitempty"`
	UpdatedAt      time.Time    `json:"updated_at"`
	CreatedAt      time.Time    `json:"created_at"`
}

type ProjectInfo struct {
	ProjectID   string    `json:"project_id"`
	ForemanName string    `json:"foreman_name"`
	ForemanPhone string   `json:"foreman_phone"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type AlertSeverity string

const (
	SeverityWarning AlertSeverity = "warning"
	SeverityCritical AlertSeverity = "critical"
)

type PhaseAlert struct {
	ID          string         `json:"id"`
	ProjectID   string         `json:"project_id"`
	PhaseType   PhaseType      `json:"phase_type"`
	PhaseName   string         `json:"phase_name"`
	Severity    AlertSeverity  `json:"severity"`
	Message     string         `json:"message"`
	OverdueDays int            `json:"overdue_days"`
	NotifiedAt  time.Time      `json:"notified_at"`
}

type UpdatePhaseRequest struct {
	Status PhaseStatus `json:"status"`
}

type SetPlannedDateRequest struct {
	PlannedEndDate string `json:"planned_end_date"`
}

type SetForemanRequest struct {
	ForemanName  string `json:"foreman_name"`
	ForemanPhone string `json:"foreman_phone"`
}
