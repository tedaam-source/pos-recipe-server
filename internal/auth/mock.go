package auth

import (
	"context"
	"log"
	"net/http"
)

type MockManager struct{}

func NewMockManager() *MockManager {
	return &MockManager{}
}

func (m *MockManager) Close() error {
	return nil
}

func (m *MockManager) GetRefreshToken(ctx context.Context, secretName string) (string, error) {
	log.Printf("[MOCK] GetRefreshToken called for %s", secretName)
	return "mock-refresh-token", nil
}

func (m *MockManager) GetHTTPClient(ctx context.Context, refreshToken string) *http.Client {
	log.Printf("[MOCK] GetHTTPClient called with token %s", refreshToken)
	return http.DefaultClient
}
