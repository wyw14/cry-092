package archive

import (
	"sort"
	"time"
)

type Period struct {
	Start, End time.Time
	Label      string
}

func (p Period) Valid() bool { return p.Label != "" && p.End.After(p.Start) }

type UnitAggregate struct {
	UnitID          string
	Eligible        int
	AnsweredOnTime  int
	Satisfied       int
	MostlySatisfied int
}

type RankingEntry struct {
	UnitID           string
	Period           Period
	Numerator        int
	Denominator      int
	ScoreBasisPoints int
	Position         int
}

func BuildRanking(period Period, aggregates []UnitAggregate) []RankingEntry {
	entries := make([]RankingEntry, 0, len(aggregates))
	for _, a := range aggregates {
		score := 0
		if a.Eligible > 0 {
			score = (a.Satisfied*10000 + a.MostlySatisfied*7000 + a.AnsweredOnTime*3000) / (a.Eligible * 2)
		}
		entries = append(entries, RankingEntry{UnitID: a.UnitID, Period: period, Numerator: a.Satisfied + a.MostlySatisfied + a.AnsweredOnTime, Denominator: a.Eligible * 2, ScoreBasisPoints: score})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].ScoreBasisPoints == entries[j].ScoreBasisPoints {
			return entries[i].UnitID < entries[j].UnitID
		}
		return entries[i].ScoreBasisPoints > entries[j].ScoreBasisPoints
	})
	for i := range entries {
		entries[i].Position = i + 1
	}
	return entries
}
