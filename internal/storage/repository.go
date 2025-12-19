package storage

import (
	"context"
	"time"
)

type HistoryRepository interface {
	SaveWatchStatus(ctx context.Context, historyID uint64, expiration int64) error
	SaveProcessedEmail(ctx context.Context, email ProcessedEmail) error
	RecordEvent(ctx context.Context, event Event) error
	UpdateDailyStats(ctx context.Context, received, processedOk, processedError int) error
}

type ProcessedEmail struct {
	ID        uint64 `gorm:"primaryKey"`
	MessageID string `gorm:"uniqueIndex;not null"`
	HistoryID uint64 `gorm:"not null"`
	LabelIDs  string
	Snippet   string
	CreatedAt time.Time
}

// Event maps to the 'events' table created by admin service
type Event struct {
	ID        string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	MessageID string
	FilterID  string // Optional
	Status    string
	Error     string
	CreatedAt time.Time
}

// DailyStat maps to the 'stats_daily' table
type DailyStat struct {
	Day            string `gorm:"primaryKey;type:date"`
	Received       int
	ProcessedOk    int
	ProcessedError int
	LastEventAt    time.Time
}

type NoOpRepository struct{}

func (r *NoOpRepository) SaveWatchStatus(ctx context.Context, historyID uint64, expiration int64) error {
	return nil
}

func (r *NoOpRepository) SaveProcessedEmail(ctx context.Context, email ProcessedEmail) error {
	return nil
}

func (r *NoOpRepository) RecordEvent(ctx context.Context, event Event) error {
	return nil
}

func (r *NoOpRepository) UpdateDailyStats(ctx context.Context, received, processedOk, processedError int) error {
	return nil
}
