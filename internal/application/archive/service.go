package archive

import (
	"context"
	"fmt"

	"github.com/wyw14/cry-092/internal/application"
	domain "github.com/wyw14/cry-092/internal/domain/archive"
)

type Service struct {
	Archives application.ArchiveRepository
	Audits   application.AuditRepository
	Tx       application.TransactionManager
	Clock    application.Clock
	IDs      application.IDGenerator
}

func (s Service) RebuildRanking(ctx context.Context, period domain.Period) ([]domain.RankingEntry, error) {
	if !period.Valid() {
		return nil, fmt.Errorf("invalid ranking period")
	}
	var entries []domain.RankingEntry
	err := s.Tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		aggregates, err := s.Archives.Aggregates(txCtx, period)
		if err != nil {
			return err
		}
		eligible := make([]domain.UnitAggregate, 0, len(aggregates))
		for _, aggregate := range aggregates {
			if aggregate.Eligible == 0 {
				continue
			}
			eligible = append(eligible, aggregate)
		}
		entries = domain.BuildRanking(period, eligible)
		return s.Archives.SaveRanking(txCtx, entries)
	})
	if err != nil {
		return nil, fmt.Errorf("rebuild unit ranking: %w", err)
	}
	return cloneRanking(entries), nil
}

func cloneRanking(entries []domain.RankingEntry) []domain.RankingEntry {
	result := make([]domain.RankingEntry, len(entries))
	copy(result, entries)
	return result
}
