package gmail

import (
	"context"
	"fmt"
	"net/http"

	gmail "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Client struct {
	service *gmail.Service
}

func NewClient(ctx context.Context, httpClient *http.Client) (*Client, error) {
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Gmail client: %w", err)
	}

	return &Client{
		service: srv,
	}, nil
}

type WatchResponse struct {
	HistoryId  uint64
	Expiration int64
}

func (c *Client) RenewWatch(topicName string) (*WatchResponse, error) {
	req := &gmail.WatchRequest{
		TopicName: topicName,
		LabelIds:  []string{"INBOX"}, // Listening to INBOX by default, can be parameterized if needed
	}

	resp, err := c.service.Users.Watch("me", req).Do()
	if err != nil {
		return nil, fmt.Errorf("gmail watch call failed: %w", err)
	}

	return &WatchResponse{
		HistoryId:  resp.HistoryId,
		Expiration: resp.Expiration,
	}, nil
}
