package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gagarin-soft/internal/config"
	"gagarin-soft/internal/handlers"
	"gagarin-soft/internal/services"
	"gagarin-soft/internal/storage/mocks"
)

// Reusing MockTokenManager and MockTransport from services test would be ideal,
// but for simplicity/isolation we redefine or place in a shared test package.
// Here we redefine specific to this test for clarity.

type MockAuthManager struct {
	Client *http.Client
}

func (m *MockAuthManager) GetRefreshToken(ctx context.Context, secretName string) (string, error) {
	return "mock-jwt", nil
}
func (m *MockAuthManager) GetHTTPClient(ctx context.Context, refreshToken string) *http.Client {
	return m.Client
}
func (m *MockAuthManager) Close() error { return nil }

type MockTransport struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

func TestRenewWatchHandler_ServeHTTP(t *testing.T) {
	// 1. Setup Dependencies
	cfg := &config.Config{
		ProjectID: "test-proj",
	}
	repo := mocks.NewMockHistoryRepository()
	
	mockTransport := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Mock Gmail API response
			respBody := `{"historyId": "999", "expiration": "88888888"}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(respBody)),
				Header:     make(http.Header),
			}, nil
		},
	}
	httpClient := &http.Client{Transport: mockTransport}
	authMgr := &MockAuthManager{Client: httpClient}

	svc := services.NewGmailWatchService(cfg, authMgr, repo)
	handler := &handlers.RenewWatchHandler{Service: svc}

	// 2. Create Request
	// Case A: Custom Topic in Body
	reqBody := map[string]string{"topicName": "projects/custom/topics/my-topic"}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/renew-watch", bytes.NewBuffer(bodyBytes))
	w := httptest.NewRecorder()

	// 3. Execute
	handler.ServeHTTP(w, req)

	// 4. Verify
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var respMap map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respMap); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if respMap["historyId"].(string) != "999" {
		t.Errorf("Expected historyId 999, got %v", respMap["historyId"])
	}
}
