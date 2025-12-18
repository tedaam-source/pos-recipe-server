package auth

import (
	"context"
	"fmt"
	"hash/crc32"
	"net/http"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Manager struct {
	secretsClient *secretmanager.Client
	projectID     string
	clientID      string
	clientSecret  string
}

func NewManager(ctx context.Context, projectID, clientID, clientSecret string) (*Manager, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret manager client: %w", err)
	}

	return &Manager{
		secretsClient: client,
		projectID:     projectID,
		clientID:      clientID,
		clientSecret:  clientSecret,
	}, nil
}

func (m *Manager) Close() error {
	return m.secretsClient.Close()
}

// GetRefreshToken retrieves the refresh token from Secret Manager
func (m *Manager) GetRefreshToken(ctx context.Context, secretName string) (string, error) {
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", m.projectID, secretName)

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	result, err := m.secretsClient.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret %s: %w", name, err)
	}

	// Verify data checksum.
	crc32c := crc32.MakeTable(crc32.Castagnoli)
	checksum := int64(crc32.Checksum(result.Payload.Data, crc32c))
	if checksum != *result.Payload.DataCrc32C {
		return "", fmt.Errorf("data corruption detected for secret %s", name)
	}

	return string(result.Payload.Data), nil
}

// GetHTTPClient returns an authenticated HTTP client using the refresh token
func (m *Manager) GetHTTPClient(ctx context.Context, refreshToken string) *http.Client {
	config := &oauth2.Config{
		ClientID:     m.clientID,
		ClientSecret: m.clientSecret,
		Endpoint:     google.Endpoint,
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(-1 * time.Hour), // Force refresh
	}

	return config.Client(ctx, token)
}
