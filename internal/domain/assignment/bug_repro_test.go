package assignment

import (
	"errors"
	"testing"
	"time"

	"github.com/wyw14/cry-092/internal/domain/shared"
)

func TestForeignUnitCannotAcceptOrHandleAssignedProposal(t *testing.T) {
	now := time.Date(2026, 8, 24, 9, 0, 0, 0, time.UTC)
	assignment := New("a1", "p1", "education-unit", "rule1", "supervisor", now)
	before := *assignment
	if err := assignment.Accept("transport-unit", "officer2", now.Add(time.Minute)); !errors.Is(err, shared.ErrForbidden) {
		t.Fatalf("foreign unit acceptance should be forbidden, got %v", err)
	}
	if assignment.UnitID != before.UnitID || assignment.Status != before.Status || assignment.Version != before.Version {
		t.Fatalf("rejected acceptance changed ownership: before=%+v after=%+v", before, *assignment)
	}
	if assignment.CanHandle("transport-unit") {
		t.Fatal("foreign unit became able to handle the proposal")
	}
}
