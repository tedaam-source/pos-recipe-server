package mocks

import (
	"context"
	"sync"

	"gagarin-soft/internal/storage"
)

type MockHistoryRepository struct {
	mu           sync.Mutex
	SavedHistory []SavedEntry
	SavedEmails  []storage.ProcessedEmail
	Err          error
}

type SavedEntry struct {
	HistoryID  uint64
	Expiration int64
}

func NewMockHistoryRepository() *MockHistoryRepository {
	return &MockHistoryRepository{
		SavedHistory: make([]SavedEntry, 0),
		SavedEmails:  make([]storage.ProcessedEmail, 0),
	}
}

func (m *MockHistoryRepository) SaveWatchStatus(ctx context.Context, historyID uint64, expiration int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Err != nil {
		return m.Err
	}

	m.SavedHistory = append(m.SavedHistory, SavedEntry{
		HistoryID:  historyID,
		Expiration: expiration,
	})
	return nil
}

func (m *MockHistoryRepository) SaveProcessedEmail(ctx context.Context, email storage.ProcessedEmail) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Err != nil {
		return m.Err
	}
	m.SavedEmails = append(m.SavedEmails, email)
	return nil
}

func (m *MockHistoryRepository) RecordEvent(ctx context.Context, event storage.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Err
}

func (m *MockHistoryRepository) UpdateDailyStats(ctx context.Context, received, processedOk, processedError int) error {
	return m.Err
}
