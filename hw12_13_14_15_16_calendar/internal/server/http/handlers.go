package internalhttp

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/storage"
)

// CreateEventRequest represents the request body for creating an event.
type CreateEventRequest struct {
	ID           string        `json:"id"`
	Title        string        `json:"title"`
	DateTime     time.Time     `json:"dateTime"`
	Duration     time.Duration `json:"duration"`
	Description  string        `json:"description,omitempty"`
	UserID       string        `json:"userId"`
	NotifyBefore time.Duration `json:"notifyBefore,omitempty"`
}

// UpdateEventRequest represents the request body for updating an event.
type UpdateEventRequest struct {
	ID           string        `json:"id"`
	Title        string        `json:"title"`
	DateTime     time.Time     `json:"dateTime"`
	Duration     time.Duration `json:"duration"`
	Description  string        `json:"description,omitempty"`
	UserID       string        `json:"userId"`
	NotifyBefore time.Duration `json:"notifyBefore,omitempty"`
}

// DeleteEventRequest represents the request body for deleting an event.
type DeleteEventRequest struct {
	ID string `json:"id"`
}

// ListEventsRequest represents the request for listing events by date.
type ListEventsRequest struct {
	Date time.Time `json:"date"`
}

// EventResponse represents a single event in responses.
type EventResponse struct {
	ID           string        `json:"id"`
	Title        string        `json:"title"`
	DateTime     time.Time     `json:"dateTime"`
	Duration     time.Duration `json:"duration"`
	Description  string        `json:"description,omitempty"`
	UserID       string        `json:"userId"`
	NotifyBefore time.Duration `json:"notifyBefore,omitempty"`
}

// CreateEventResponse represents the response for creating an event.
type CreateEventResponse struct {
	ID string `json:"id"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Handlers struct holds dependencies for HTTP handlers.
type Handlers struct {
	app Application
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(app Application) *Handlers {
	return &Handlers{app: app}
}

// CreateEventHandler handles POST /events
func (h *Handlers) CreateEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	event := storage.Event{
		ID:           req.ID,
		Title:        req.Title,
		DateTime:     req.DateTime,
		Duration:     req.Duration,
		Description:  req.Description,
		UserID:       req.UserID,
		NotifyBefore: req.NotifyBefore,
	}

	if err := h.app.CreateEvent(r.Context(), event); err != nil {
		// TODO: Use proper error handling based on storage errors
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, CreateEventResponse{ID: event.ID})
}

// UpdateEventHandler handles PUT /events/{id}
func (h *Handlers) UpdateEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	event := storage.Event{
		ID:           req.ID,
		Title:        req.Title,
		DateTime:     req.DateTime,
		Duration:     req.Duration,
		Description:  req.Description,
		UserID:       req.UserID,
		NotifyBefore: req.NotifyBefore,
	}

	if err := h.app.UpdateEvent(r.Context(), event); err != nil {
		// TODO: Use proper error handling based on storage errors
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// DeleteEventHandler handles DELETE /events/{id}
func (h *Handlers) DeleteEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeleteEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.app.DeleteEvent(r.Context(), req.ID); err != nil {
		// TODO: Use proper error handling based on storage errors
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ListEventsForDayHandler handles GET /events/day?date=...
func (h *Handlers) ListEventsForDayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		respondWithError(w, http.StatusBadRequest, "date parameter is required")
		return
	}

	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid date format, use RFC3339")
		return
	}

	events, err := h.app.ListEventsForDay(r.Context(), date)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithEvents(w, events)
}

// ListEventsForWeekHandler handles GET /events/week?date=...
func (h *Handlers) ListEventsForWeekHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		respondWithError(w, http.StatusBadRequest, "date parameter is required")
		return
	}

	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid date format, use RFC3339")
		return
	}

	events, err := h.app.ListEventsForWeek(r.Context(), date)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithEvents(w, events)
}

// ListEventsForMonthHandler handles GET /events/month?date=...
func (h *Handlers) ListEventsForMonthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		respondWithError(w, http.StatusBadRequest, "date parameter is required")
		return
	}

	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid date format, use RFC3339")
		return
	}

	events, err := h.app.ListEventsForMonth(r.Context(), date)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithEvents(w, events)
}

// Helper functions

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, ErrorResponse{Error: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func respondWithEvents(w http.ResponseWriter, events []storage.Event) {
	eventResponses := make([]EventResponse, len(events))
	for i, event := range events {
		eventResponses[i] = EventResponse{
			ID:           event.ID,
			Title:        event.Title,
			DateTime:     event.DateTime,
			Duration:     event.Duration,
			Description:  event.Description,
			UserID:       event.UserID,
			NotifyBefore: event.NotifyBefore,
		}
	}
	respondWithJSON(w, http.StatusOK, eventResponses)
}
