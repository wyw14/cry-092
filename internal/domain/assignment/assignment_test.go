package assignment

import (
	"errors"
	"testing"
	"time"

	"github.com/wyw14/cry-092/internal/domain/shared"
)

func TestOnlyAssignedUnitCanAcceptAndHandle(t *testing.T) {
	a := New("a", "p", "unit-a", "rule", "supervisor", time.Now())
	if err := a.Accept("unit-b", "officer", time.Now()); !errors.Is(err, shared.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if err := a.Accept("unit-a", "officer", time.Now()); err != nil {
		t.Fatal(err)
	}
	if !a.CanHandle("unit-a") || a.CanHandle("unit-b") {
		t.Fatalf("invalid handling ownership: %+v", a)
	}
}

func TestRuleSetUsesHighestPriorityDeterministically(t *testing.T) {
	rule, ok := (RuleSet{Rules: []Rule{{ID: "b", Category: "education", UnitID: "u2", Priority: 10, Active: true}, {ID: "a", Category: "education", UnitID: "u1", Priority: 10, Active: true}, {ID: "high", Category: "education", UnitID: "u3", Priority: 20, Active: true}}}).Resolve("education")
	if !ok || rule.ID != "high" {
		t.Fatalf("unexpected rule %+v", rule)
	}
}
