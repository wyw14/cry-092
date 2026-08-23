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
		entries = domain.BuildRanking(period, aggregates)
		return s.Archives.SaveRanking(txCtx, entries)
	})
	return entries, err
}
