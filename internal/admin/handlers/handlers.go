package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"gagarin-soft/internal/admin/config"
	"gagarin-soft/internal/admin/storage"
)

type Handler struct {
	cfg     *config.Config
	storage *storage.Storage
}

func NewHandler(cfg *config.Config, store *storage.Storage) *Handler {
	return &Handler{
		cfg:     cfg,
		storage: store,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" {
		from = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if to == "" {
		to = time.Now().Format("2006-01-02")
	}

	stats, err := h.storage.GetDailyStats(r.Context(), from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(stats)
}

func (h *Handler) GetFilters(w http.ResponseWriter, r *http.Request) {
	filters, err := h.storage.GetFilters(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(filters)
}

func (h *Handler) CreateFilter(w http.ResponseWriter, r *http.Request) {
	var f storage.Filter
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	f.UpdatedBy = getAdminEmail(r)

	if err := h.storage.CreateFilter(r.Context(), &f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) UpdateFilter(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var f storage.Filter
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	f.UpdatedBy = getAdminEmail(r)

	if err := h.storage.UpdateFilter(r.Context(), id, &f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteFilter(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.storage.DeleteFilter(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetEvents(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 {
			limit = l
		}
	}
	events, err := h.storage.GetEvents(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(events)
}

func (h *Handler) TriggerAction(w http.ResponseWriter, r *http.Request) {
	action := chi.URLParam(r, "action") // renew-watch, resync, reprocess

	// Delegate to worker implementation
	// For now just return OK
	// Real implementation needs to call Worker URL

	// TODO: Call h.cfg.WorkerBaseURL + ...

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Action triggered: " + action))
}

func getAdminEmail(r *http.Request) string {
	email := r.Header.Get("X-Goog-Authenticated-User-Email")
	if strings.HasPrefix(email, "accounts.google.com:") {
		email = strings.TrimPrefix(email, "accounts.google.com:")
	}
	return email
}
