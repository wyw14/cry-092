package handling

import (
	"time"

	"github.com/wyw14/cry-092/internal/domain/shared"
)

type PlanStatus string

const (
	PlanActive    PlanStatus = "active"
	PlanExtended  PlanStatus = "extended"
	PlanCompleted PlanStatus = "completed"
)

type Material struct {
	ID      string
	Title   string
	FileID  string
	AddedBy string
	AddedAt time.Time
}

type Plan struct {
	ID           string
	ProposalID   string
	AssignmentID string
	UnitID       string
	Steps        []string
	Materials    []Material
	DueAt        time.Time
	Status       PlanStatus
	Version      int64
}

func NewPlan(id, proposalID, assignmentID, unitID string, acceptedAt time.Time, calendar WorkdayCalendar, steps []string) (*Plan, error) {
	if len(steps) == 0 {
		return nil, shared.NewError("PLAN_STEPS_REQUIRED", "handling plan needs at least one step", nil)
	}
	return &Plan{ID: id, ProposalID: proposalID, AssignmentID: assignmentID, UnitID: unitID, Steps: append([]string(nil), steps...), DueAt: calendar.AddWorkdays(acceptedAt, 30), Status: PlanActive, Version: 1}, nil
}

func (p *Plan) AddMaterial(material Material) error {
	if p.Status == PlanCompleted {
		return shared.NewError("PLAN_COMPLETED", "completed plan cannot receive material", shared.ErrInvalidState)
	}
	material.AddedAt = material.AddedAt.UTC()
	p.Materials = append(p.Materials, material)
	p.Version++
	return nil
}

func (p *Plan) Complete(now time.Time) error {
	if p.Status != PlanActive && p.Status != PlanExtended {
		return shared.ErrInvalidState
	}
	p.Status = PlanCompleted
	p.Version++
	return nil
}
