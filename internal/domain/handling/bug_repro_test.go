package handling

import (
	"testing"
	"time"
)

func TestPlanAndExtensionUseTheSameWorkdayCalendar(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	accepted := time.Date(2026, 9, 25, 10, 0, 0, 0, location)
	holiday := time.Date(2026, 10, 1, 0, 0, 0, 0, location)
	makeupSaturday := time.Date(2026, 10, 10, 0, 0, 0, 0, location)
	calendar := NewCalendar(location, []time.Time{holiday}, []time.Time{makeupSaturday})
	plan, err := NewPlan("plan", "proposal", "assignment", "unit", accepted, calendar, []string{"现场调研"})
	if err != nil {
		t.Fatal(err)
	}
	wantInitial := referenceWorkdays(calendar, accepted, 30)
	if !plan.DueAt.Equal(wantInitial) {
		t.Fatalf("initial deadline = %s, want %s", plan.DueAt, wantInitial)
	}
	extension, err := NewExtension("extension", plan.ID, "officer", "需要补充论证", 5, accepted)
	if err != nil {
		t.Fatal(err)
	}
	wantExtended := referenceWorkdays(calendar, plan.DueAt, 5)
	if err := extension.Review(plan, "supervisor", true, calendar, accepted); err != nil {
		t.Fatal(err)
	}
	if !plan.DueAt.Equal(wantExtended) {
		t.Fatalf("extended deadline = %s, want %s", plan.DueAt, wantExtended)
	}
}

func referenceWorkdays(calendar WorkdayCalendar, start time.Time, days int) time.Time {
	cursor := start.In(calendar.Location)
	for counted := 0; counted < days; {
		cursor = cursor.AddDate(0, 0, 1)
		if calendar.IsWorkday(cursor) {
			counted++
		}
	}
	return cursor.UTC()
}
