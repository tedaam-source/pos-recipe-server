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
	if err := gormDB.AutoMigrate(&GmailWatchHistory{}); err != nil {
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
