package storage

import (
	"context"
)

type HistoryRepository interface {
	SaveWatchStatus(ctx context.Context, historyID uint64, expiration int64) error
}

type NoOpRepository struct{}

func (r *NoOpRepository) SaveWatchStatus(ctx context.Context, historyID uint64, expiration int64) error {
	// For now, we just log it in the main handler, or we could add structured logging here.
	// This satisfies the architecture requirement for separation of concerns.
	return nil
}
