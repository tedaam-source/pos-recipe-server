package handlers

import (
	"net/http"

	"gagarin-soft/internal/services"
)

type RenewWatchHandler struct {
	Service *services.GmailWatchService
}

func (h *RenewWatchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := h.Service.Renew(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}
