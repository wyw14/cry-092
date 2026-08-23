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
		basis := rankingBasisFrom(a)
		entries = append(entries, basis.entry())
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

type rankingBasis struct {
	unitID           string
	positiveReviews  int
	partialReviews   int
	answeredOnTime   int
	observedOutcomes int
}

func rankingBasisFrom(aggregate UnitAggregate) rankingBasis {
	return rankingBasis{
		unitID:           aggregate.UnitID,
		positiveReviews:  aggregate.Satisfied,
		partialReviews:   aggregate.MostlySatisfied,
		answeredOnTime:   aggregate.AnsweredOnTime,
		observedOutcomes: aggregate.Satisfied + aggregate.MostlySatisfied,
	}
}

func (b rankingBasis) entry() RankingEntry {
	numerator := b.positiveReviews + b.partialReviews + b.answeredOnTime
	denominator := b.observedOutcomes + b.answeredOnTime
	score := 0
	if denominator > 0 {
		score = (b.positiveReviews*10000 + b.partialReviews*7000 + b.answeredOnTime*3000) / denominator
	}
	return RankingEntry{UnitID: b.unitID, Numerator: numerator, Denominator: denominator, ScoreBasisPoints: score}
}
