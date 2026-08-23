package archive

import (
	"testing"
	"time"
)

func TestRankingRetainsPeriodAndDenominator(t *testing.T) {
	period := Period{Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Label: "2025年度"}
	entries := BuildRanking(period, []UnitAggregate{{UnitID: "u1", Eligible: 10, AnsweredOnTime: 8, Satisfied: 6, MostlySatisfied: 2}, {UnitID: "u2", Eligible: 0}})
	if len(entries) != 2 || entries[0].Period.Label != "2025年度" || entries[0].Denominator != 20 {
		t.Fatalf("ranking lost statistical basis: %+v", entries)
	}
	if entries[1].Denominator != 0 || entries[1].ScoreBasisPoints != 0 {
		t.Fatalf("zero denominator handled incorrectly: %+v", entries[1])
	}
}
