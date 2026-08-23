package handling

import (
	"testing"
	"time"
)

func TestWorkdayCalendarHonorsHolidayAndMakeupDay(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	holiday := time.Date(2026, 10, 1, 0, 0, 0, 0, loc)
	makeup := time.Date(2026, 10, 10, 0, 0, 0, 0, loc)
	calendar := NewCalendar(loc, []time.Time{holiday}, []time.Time{makeup})
	if calendar.IsWorkday(holiday) {
		t.Fatal("holiday counted as workday")
	}
	if !calendar.IsWorkday(makeup) {
		t.Fatal("makeup Saturday not counted")
	}
	start := time.Date(2026, 9, 30, 9, 0, 0, 0, loc)
	due := calendar.AddWorkdays(start, 2).In(loc)
	if due.Format("2006-01-02") != "2026-10-05" {
		t.Fatalf("unexpected due date %s", due)
	}
}

func TestExtensionUpdatesPlanWithSameCalendar(t *testing.T) {
	loc := time.UTC
	calendar := NewCalendar(loc, nil, nil)
	now := time.Date(2026, 8, 24, 0, 0, 0, 0, loc)
	plan, err := NewPlan("plan", "p", "a", "unit", now, calendar, []string{"现场调研"})
	if err != nil {
		t.Fatal(err)
	}
	original := plan.DueAt
	extension, err := NewExtension("e", "plan", "unit-user", "补充论证", 5, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := extension.Review(plan, "supervisor", true, calendar, now); err != nil {
		t.Fatal(err)
	}
	if got := calendar.WorkdaysBetween(original, plan.DueAt); got != 5 {
		t.Fatalf("expected 5 workdays, got %d", got)
	}
}
