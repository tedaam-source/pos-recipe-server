package storage

import (
	"context"
	"fmt"
	"net"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool    *pgxpool.Pool
	cleanup func() error
}

func New(ctx context.Context, connString string, instanceConnectionName string) (*Storage, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	var cleanup func() error

	if instanceConnectionName != "" {
		// Use Cloud SQL Connector
		d, err := cloudsqlconn.NewDialer(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to init dialer: %w", err)
		}
		cleanup = func() error { return d.Close() }

		config.ConnConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return d.Dial(ctx, instanceConnectionName)
		}
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		if cleanup != nil {
			cleanup()
		}
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		if cleanup != nil {
			cleanup()
		}
		pool.Close()
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return &Storage{pool: pool, cleanup: cleanup}, nil
}

func (s *Storage) Close() {
	if s.cleanup != nil {
		s.cleanup()
	}
	s.pool.Close()
}

// --- Models ---

type Filter struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Enabled    bool      `json:"enabled"`
	Priority   int       `json:"priority"`
	GmailQuery string    `json:"gmail_query"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	UpdatedBy  string    `json:"updated_by"`
}

type DailyStat struct {
	Day            string    `json:"day"` // YYYY-MM-DD
	Received       int       `json:"received"`
	ProcessedOk    int       `json:"processed_ok"`
	ProcessedError int       `json:"processed_error"`
	LastEventAt    time.Time `json:"last_event_at"`
}

type Event struct {
	ID        string    `json:"id"`
	MessageID string    `json:"message_id"`
	FilterID  string    `json:"filter_id,omitempty"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// --- Methods ---

func (s *Storage) GetFilters(ctx context.Context) ([]Filter, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name, enabled, priority, gmail_query, created_at, updated_at, COALESCE(updated_by, '') FROM filters ORDER BY priority ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var filters []Filter
	for rows.Next() {
		var f Filter
		if err := rows.Scan(&f.ID, &f.Name, &f.Enabled, &f.Priority, &f.GmailQuery, &f.CreatedAt, &f.UpdatedAt, &f.UpdatedBy); err != nil {
			return nil, err
		}
		filters = append(filters, f)
	}
	return filters, nil
}

func (s *Storage) CreateFilter(ctx context.Context, f *Filter) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO filters (name, enabled, priority, gmail_query, updated_by) VALUES ($1, $2, $3, $4, $5)`,
		f.Name, f.Enabled, f.Priority, f.GmailQuery, f.UpdatedBy)
	return err
}

func (s *Storage) UpdateFilter(ctx context.Context, id string, f *Filter) error {
	_, err := s.pool.Exec(ctx, `UPDATE filters SET name=$1, enabled=$2, priority=$3, gmail_query=$4, updated_by=$5, updated_at=NOW() WHERE id=$6`,
		f.Name, f.Enabled, f.Priority, f.GmailQuery, f.UpdatedBy, id)
	return err
}

func (s *Storage) DeleteFilter(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM filters WHERE id=$1`, id)
	return err
}

func (s *Storage) GetDailyStats(ctx context.Context, from, to string) ([]DailyStat, error) {
	rows, err := s.pool.Query(ctx, `SELECT day, received, processed_ok, processed_error, last_event_at FROM stats_daily WHERE day >= $1 AND day <= $2 ORDER BY day DESC`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []DailyStat
	for rows.Next() {
		var st DailyStat
		var day time.Time
		if err := rows.Scan(&day, &st.Received, &st.ProcessedOk, &st.ProcessedError, &st.LastEventAt); err != nil {
			return nil, err
		}
		st.Day = day.Format("2006-01-02")
		stats = append(stats, st)
	}
	return stats, nil
}

func (s *Storage) GetEvents(ctx context.Context, limit int) ([]Event, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, message_id, COALESCE(filter_id::text, ''), status, COALESCE(error, ''), created_at FROM events ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.MessageID, &e.FilterID, &e.Status, &e.Error, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}
