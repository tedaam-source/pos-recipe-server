package gmail

import (
	"fmt"

	gmail "google.golang.org/api/gmail/v1"
)

// ListMessageIDs returns a list of message IDs added since the given historyId
func (c *Client) ListMessageIDs(historyId uint64) ([]string, error) {
	resp, err := c.service.Users.History.List("me").StartHistoryId(historyId).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list history: %w", err)
	}

	var messageIDs []string
	for _, h := range resp.History {
		for _, m := range h.MessagesAdded {
			if m.Message != nil {
				messageIDs = append(messageIDs, m.Message.Id)
			}
		}
	}
	return messageIDs, nil
}

// GetMessage returns the full message details including labels
func (c *Client) GetMessage(messageId string) (*gmail.Message, error) {
	return c.service.Users.Messages.Get("me", messageId).Do()
}
