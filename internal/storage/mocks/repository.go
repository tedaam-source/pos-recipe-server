package mocks

import (
	"context"
	"sync"
)

type MockHistoryRepository struct {
	mu           sync.Mutex
	SavedHistory []SavedEntry
	Err          error
}

type SavedEntry struct {
	HistoryID  uint64
	Expiration int64
}

func NewMockHistoryRepository() *MockHistoryRepository {
	return &MockHistoryRepository{
		SavedHistory: make([]SavedEntry, 0),
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
