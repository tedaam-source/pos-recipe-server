package storage

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type GmailWatchHistory struct {
	ID         uint64 `gorm:"primaryKey"`
	HistoryID  uint64 `gorm:"not null"`
	Expiration int64  `gorm:"not null"`
	CreatedAt  time.Time
}

// TableName overrides the default pluralization if needed, though 'events' and 'stats_daily' are standard.
func (Event) TableName() string {
	return "events"
}

func (DailyStat) TableName() string {
	return "stats_daily"
}

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(ctx context.Context, instanceConnectionName, dbUser, dbPass, dbName string) (*PostgresRepository, func() error, error) {
	d, err := cloudsqlconn.NewDialer(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init dialer: %w", err)
	}

	// Cleanup function to close the dialer
	cleanup := func() error {
		return d.Close()
	}

	dsn := fmt.Sprintf("user=%s password=%s database=%s", dbUser, dbPass, dbName)
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}

	// Configure the dialer
	config.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return d.Dial(ctx, instanceConnectionName)
	}

	dbURI := stdlib.RegisterConnConfig(config)
	sqlDB, err := sql.Open("pgx", dbURI)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to open sql connection: %w", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to open gorm connection: %w", err)
	}

	// AutoMigrate
	if err := gormDB.AutoMigrate(&GmailWatchHistory{}, &ProcessedEmail{}); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &PostgresRepository{db: gormDB}, cleanup, nil
}

func (r *PostgresRepository) SaveWatchStatus(ctx context.Context, historyID uint64, expiration int64) error {
	entry := GmailWatchHistory{
		HistoryID:  historyID,
		Expiration: expiration,
		CreatedAt:  time.Now(),
	}
	return r.db.WithContext(ctx).Create(&entry).Error
}

func (r *PostgresRepository) SaveProcessedEmail(ctx context.Context, email ProcessedEmail) error {
	if email.CreatedAt.IsZero() {
		email.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(&email).Error
}

func (r *PostgresRepository) RecordEvent(ctx context.Context, event Event) error {
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	// 'events' table uses UUID default gen_random_uuid(), so we check if ID is empty, let DB handle it.
	// However, GORM might try to insert zero value.
	// Best to use a map or Omit ID if empty.
	return r.db.WithContext(ctx).Omit("ID").Create(&event).Error
}

func (r *PostgresRepository) UpdateDailyStats(ctx context.Context, received, processedOk, processedError int) error {
	day := time.Now().Format("2006-01-02")

	// Atomic upsert via raw SQL because GORM upsert with increments is verbose
	// "stats_daily" (day, received, processed_ok, processed_error, last_event_at)
	query := `
		INSERT INTO stats_daily (day, received, processed_ok, processed_error, last_event_at)
		VALUES (?, ?, ?, ?, NOW())
		ON CONFLICT (day) DO UPDATE SET
			received = stats_daily.received + excluded.received,
			processed_ok = stats_daily.processed_ok + excluded.processed_ok,
			processed_error = stats_daily.processed_error + excluded.processed_error,
			last_event_at = NOW();
	`
	return r.db.WithContext(ctx).Exec(query, day, received, processedOk, processedError).Error
}
