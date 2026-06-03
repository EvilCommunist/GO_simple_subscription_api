package handlers

import (
	"CRUDL/db/crud"
	"CRUDL/db/models"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SubscriptionsHandler struct {
	conn *crud.SubscriptionDB
}

func (sh *SubscriptionsHandler) SetDatabaseHandler(conn *crud.SubscriptionDB) {
	sh.conn = conn
}

func (sh *SubscriptionsHandler) Post(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ServiceName string  `json:"service_name"`
		Price       int     `json:"price"`
		UserID      string  `json:"user_id"`
		StartDate   string  `json:"start_date"`
		EndDate     *string `json:"end_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		slog.Error("Invalid request acquired", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(request.UserID)
	if err != nil {
		slog.Error("Invalid user ID acquired", "error", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	startDate, err := time.Parse("01-2006", request.StartDate)
	if err != nil {
		slog.Error("Invalid start date acquired", "error", err)
		http.Error(w, "Invalid start date of subscription", http.StatusBadRequest)
		return
	}
	var endDate *time.Time = nil
	if request.EndDate != nil && *request.EndDate != "" {
		t, err := time.Parse("01-2006", *request.EndDate)
		if err != nil {
			slog.Error("Invalid end date acquired", "error", err)
			http.Error(w, "Invalid end date of subscription", http.StatusBadRequest)
			return
		}
		endDate = &t
		if endDate.Before(startDate) {
			slog.Error("Time of subscription expires is before it's beginning")
			http.Error(w, "Invalid subscription period", http.StatusBadRequest)
			return
		}
	}

	if strings.TrimSpace(request.ServiceName) == "" {
		slog.Error("Empty service name acquired")
		http.Error(w, "Invalid service name", http.StatusBadRequest)
		return
	}
	if request.Price < 0 {
		slog.Error("Negative price value acquired")
		http.Error(w, "Invalid price", http.StatusBadRequest)
		return
	}

	sub := models.Subscription{
		ServiceName: request.ServiceName,
		Price:       request.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	created, err := sh.conn.Create(sub, r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (sh *SubscriptionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		slog.Error("Invalid ID acquired", "error", err)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	sub, err := sh.conn.Read(id, r.Context())
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "ID cannot be negative"):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case strings.Contains(err.Error(), "no rows in result set"):
			http.Error(w, "Subscription not found", http.StatusNotFound)
		case strings.Contains(err.Error(), "Query refused"):
			http.Error(w, err.Error(), http.StatusInternalServerError)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sub)
}

func (sh *SubscriptionsHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Price     *int    `json:"price"`
		StartDate *string `json:"start_date"`
		EndDate   *string `json:"end_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		slog.Error("Invalid request acquired", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var (
		startDate *time.Time = nil
		endDate   *time.Time = nil
	)
	if request.StartDate != nil && *request.StartDate != "" {
		t, err := time.Parse("01-2006", *request.StartDate)
		if err != nil {
			slog.Error("Invalid start date acquired", "error", err)
			http.Error(w, "Invalid start date of subscription", http.StatusBadRequest)
			return
		}
		startDate = &t
	}
	if request.EndDate != nil && *request.EndDate != "" {
		t, err := time.Parse("01-2006", *request.EndDate)
		if err != nil {
			slog.Error("Invalid start date acquired", "error", err)
			http.Error(w, "Invalid start date of subscription", http.StatusBadRequest)
			return
		}
		endDate = &t
		if startDate != nil && endDate.Before(*startDate) {
			slog.Error("Time of subscription expires is before it's beginning")
			http.Error(w, "Invalid subscription period", http.StatusBadRequest)
			return
		}
	}

	updSub := models.SubscriptionUpdate{
		ID:        id,
		Price:     request.Price,
		StartDate: startDate,
		EndDate:   endDate,
	}

	updatedSub, err := sh.conn.Update(updSub, r.Context())
	if err != nil {
		slog.Error("Unable to update subscription", "error", err)
		switch {
		case strings.Contains(err.Error(), "No parameters"):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case strings.Contains(err.Error(), "no rows in result set"):
			http.Error(w, "Subscription not found", http.StatusNotFound)
		case strings.Contains(err.Error(), "Query refused"):
			http.Error(w, err.Error(), http.StatusInternalServerError)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedSub)
}

func (sh *SubscriptionsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		slog.Error("Invalid ID acquired", "error", err)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = sh.conn.Delete(id, r.Context())
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "Nothing was deleted"):
			http.Error(w, err.Error(), http.StatusNotFound)
		case strings.Contains(err.Error(), "Query refused"):
			http.Error(w, err.Error(), http.StatusInternalServerError)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (sh *SubscriptionsHandler) List(w http.ResponseWriter, r *http.Request) {
	uuidStr := r.URL.Query().Get("user_id")
	if uuidStr == "" {
		slog.Error("Empty UUID acquired")
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	uuid, err := uuid.Parse(uuidStr)
	if err != nil {
		slog.Error("Invalid UUID acquired", "error", err)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	var serviceName *string
	service := r.URL.Query().Get("service_name")
	if service != "" {
		serviceName = &service
	}

	subs, err := sh.conn.List(r.Context(), uuid, serviceName)

	if err != nil {
		switch {
		case strings.Contains(err.Error(), "Service name cannot be empty"):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case strings.Contains(err.Error(), "Query refused") ||
			strings.Contains(err.Error(), "Scan dropped with errors") ||
			strings.Contains(err.Error(), "Error of subscriptions iteration"):
			http.Error(w, err.Error(), http.StatusInternalServerError)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(subs)
}
