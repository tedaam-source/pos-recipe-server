package handlers

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"gagarin-soft/internal/services"
)

type PushHandler struct {
	Service *services.GmailWatchService
}

type PubSubMessage struct {
	Message struct {
		Data string `json:"data"`
	} `json:"message"`
}

type GmailPushData struct {
	EmailAddress string `json:"emailAddress"`
	HistoryID    uint64 `json:"historyId"`
}

func (h *PushHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req PubSubMessage
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}

	data, err := base64.StdEncoding.DecodeString(req.Message.Data)
	if err != nil {
		log.Printf("Failed to decode push data: %v", err)
		http.Error(w, "Invalid Data", http.StatusBadRequest)
		return
	}

	var pushData GmailPushData
	if err := json.Unmarshal(data, &pushData); err != nil {
		log.Printf("Failed to unmarshal push data: %v", err)
		w.WriteHeader(http.StatusOK) // Acknowledge to prevent retry loop
		return
	}

	log.Printf("Received push for historyId: %d", pushData.HistoryID)
	if err := h.Service.ProcessPushNotification(r.Context(), pushData.HistoryID); err != nil {
		log.Printf("Error processing push: %v", err)
		// Return 200 to acknowledge Pub/Sub, but log error
	}

	w.WriteHeader(http.StatusOK)
}
