package services_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"gagarin-soft/internal/config"
	"gagarin-soft/internal/services"
	"gagarin-soft/internal/storage/mocks"
)

// MockTokenManager implements auth.TokenManager
type MockTokenManager struct {
	Client *http.Client
}

func (m *MockTokenManager) GetRefreshToken(ctx context.Context, secretName string) (string, error) {
	return "mock-refresh-token", nil
}

func (m *MockTokenManager) GetHTTPClient(ctx context.Context, refreshToken string) *http.Client {
	return m.Client
}

func (m *MockTokenManager) Close() error {
	return nil
}

// MockTransport intercepts HTTP requests
type MockTransport struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

func TestGmailWatchService_Renew(t *testing.T) {
	// 1. Setup Config
	cfg := &config.Config{
		ProjectID:        "test-project",
		GmailPubSubTopic: "projects/test-project/topics/test-topic",
	}

	// 2. Setup Mock Repo
	mockRepo := mocks.NewMockHistoryRepository()

	// 3. Setup Mock HTTP Client (to simulate Gmail API)
	mockTransport := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Verify URL
			if req.URL.Path != "/gmail/v1/users/me/watch" {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewBufferString("Not Found")),
				}, nil
			}

			// Return success response
			respBody := `{"historyId": "12345", "expiration": "1700000000000"}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(respBody)),
				Header:     make(http.Header),
			}, nil
		},
	}
	mockHTTPClient := &http.Client{Transport: mockTransport}

	// 4. Setup Mock Auth Manager
	mockAuth := &MockTokenManager{Client: mockHTTPClient}

	// 5. Initialize Service
	service := services.NewGmailWatchService(cfg, mockAuth, mockRepo)

	// 6. Execute
	ctx := context.Background()
	result, err := service.Renew(ctx)

	// 7. Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check result JSON
	var respMap map[string]interface{}
	if err := json.Unmarshal(result, &respMap); err != nil {
		t.Fatalf("Failed to parse result json: %v", err)
	}

	if respMap["historyId"].(string) != "12345" {
		t.Errorf("Expected historyId 12345, got %v", respMap["historyId"])
	}

	// Check Repo Save
	if len(mockRepo.SavedHistory) != 1 {
		t.Errorf("Expected 1 saved history entry, got %d", len(mockRepo.SavedHistory))
	} else {
		entry := mockRepo.SavedHistory[0]
		if entry.HistoryID != 12345 {
			t.Errorf("Expected saved historyId 12345, got %d", entry.HistoryID)
		}
	}
}
