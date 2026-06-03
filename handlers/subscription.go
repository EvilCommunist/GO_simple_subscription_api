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

// @Summary      Создать новую подписку
// @Description  Создаёт запись о подписке. Даты передаются в формате MM-YYYY.
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        request body object true "Данные подписки" example({"service_name":"Yandex Plus","price":400,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"07-2025","end_date":"12-2025"})
// @Success      201  {object}  models.Subscription
// @Failure      400  {object}  map[string]interface{} "Ошибка валидации"
// @Failure      500  {object}  map[string]interface{} "Внутренняя ошибка"
// @Router       /subscriptions [post]
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

// @Summary      Получить подписку по ID
// @Description  Возвращает полную информацию о подписке.
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID подписки"
// @Success      200  {object}  models.Subscription
// @Failure      400  {object}  map[string]interface{} "Неверный ID"
// @Failure      404  {object}  map[string]interface{} "Подписка не найдена"
// @Failure      500  {object}  map[string]interface{} "Ошибка сервера"
// @Router       /subscriptions/{id} [get]
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

// @Summary      Обновить подписку
// @Description  Обновляет одно или несколько полей (price, start_date, end_date).
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id      path      int     true  "ID подписки"
// @Param        request body      object  true  "Поля для обновления" example({"price":450})
// @Success      200     {object}  models.Subscription
// @Failure      400     {object}  map[string]interface{} "Ошибка валидации"
// @Failure      404     {object}  map[string]interface{} "Подписка не найдена"
// @Failure      500     {object}  map[string]interface{} "Внутренняя ошибка"
// @Router       /subscriptions/{id} [patch]
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

// @Summary      Удалить подписку
// @Description  Удаляет подписку по ID.
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID подписки"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]interface{} "Неверный ID"
// @Failure      404  {object}  map[string]interface{} "Подписка не найдена"
// @Failure      500  {object}  map[string]interface{} "Внутренняя ошибка"
// @Router       /subscriptions/{id} [delete]
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

// @Summary      Получить список подписок
// @Description  Возвращает список подписок пользователя. Фильтрация по названию сервиса – опциональна.
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        user_id       query     string  true  "UUID пользователя"
// @Param        service_name  query     string  false "Название сервиса"
// @Success      200           {array}   models.Subscription
// @Failure      400           {object}  map[string]interface{} "Не указан user_id или неверный формат"
// @Failure      500           {object}  map[string]interface{} "Внутренняя ошибка"
// @Router       /subscriptions [get]
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
