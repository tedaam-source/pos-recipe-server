package gmail

import (
	"fmt"

	gmail "google.golang.org/api/gmail/v1"
)

// ListMessageIDs returns a list of message IDs added since the given historyId
func (c *Client) ListMessageIDs(historyId uint64) ([]string, error) {
	if historyId == 0 {
		return []string{}, nil
	}

	// Запрашиваем не только добавленные письма, но и события по меткам
	resp, err := c.service.Users.History.List("me").
		StartHistoryId(historyId).
		HistoryTypes("messageAdded", "labelAdded", "labelRemoved").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list history: %w", err)
	}

	var messageIDs []string
	for _, h := range resp.History {
		// письма, добавленные в историю
		for _, m := range h.MessagesAdded {
			if m.Message != nil {
				messageIDs = append(messageIDs, m.Message.Id)
			}
		}
		// письма, получившие новые метки
		for _, m := range h.LabelsAdded {
			if m.Message != nil {
				messageIDs = append(messageIDs, m.Message.Id)
			}
		}
		// письма, у которых сняли метки
		for _, m := range h.LabelsRemoved {
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
