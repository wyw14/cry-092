package archive

import (
	"testing"
	"time"
)

func TestRankingKeepsConfiguredPeriodAndEligibleDenominator(t *testing.T) {
	period := Period{Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Label: "2025年度"}
	entries := BuildRanking(period, []UnitAggregate{
		{UnitID: "education", Eligible: 10, AnsweredOnTime: 8, Satisfied: 6, MostlySatisfied: 2},
		{UnitID: "transport", Eligible: 0},
	})
	if len(entries) != 2 {
		t.Fatalf("ranking omitted a unit from the configured population: %d", len(entries))
	}
	first := entries[0]
	if first.Period != period {
		t.Fatalf("ranking lost its statistical period: %+v", first.Period)
	}
	if first.Numerator != 16 || first.Denominator != 20 || first.ScoreBasisPoints != 4900 {
		t.Fatalf("ranking basis changed: %+v", first)
	}
	if entries[1].Denominator != 0 || entries[1].Period != period {
		t.Fatalf("zero-population unit lost its auditable basis: %+v", entries[1])
	}
}
