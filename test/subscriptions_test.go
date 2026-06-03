package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"CRUDL/db"
	"CRUDL/db/crud"
	"CRUDL/db/models"
	"CRUDL/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testRouter *chi.Mux
	testPool   *pgxpool.Pool
)

func TestMain(m *testing.M) {
	pool, err := db.Connect()
	if err != nil {
		fmt.Printf("Could not connect to database: %s\n", err)
		return
	}
	testPool = pool

	clearAllSubscriptions()

	repo := &crud.SubscriptionDB{}
	repo.SetDatabase(testPool)
	handler := &handlers.SubscriptionsHandler{}
	handler.SetDatabaseHandler(repo)

	testRouter = chi.NewRouter()
	testRouter.Post("/subscriptions", handler.Post)
	testRouter.Get("/subscriptions/{id}", handler.Get)
	testRouter.Patch("/subscriptions/{id}", handler.Patch)
	testRouter.Delete("/subscriptions/{id}", handler.Delete)
	testRouter.Get("/subscriptions", handler.List)

	code := m.Run()

	clearAllSubscriptions()
	testPool.Close()
	if code != 0 {
		panic("tests failed")
	}
}

func clearAllSubscriptions() {
	if testPool == nil {
		return
	}
	_, _ = testPool.Exec(context.Background(), "TRUNCATE TABLE subscriptions RESTART IDENTITY CASCADE")
}

func postJSON(url string, body interface{}) *httptest.ResponseRecorder {
	data, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

func get(url string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

func patchJSON(url string, body interface{}) *httptest.ResponseRecorder {
	data, _ := json.Marshal(body)
	req := httptest.NewRequest("PATCH", url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

func deleteReq(url string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("DELETE", url, nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

func TestCreateSubscription(t *testing.T) {
	clearAllSubscriptions()

	validBody := map[string]interface{}{
		"service_name": "Yandex Plus",
		"price":        400,
		"user_id":      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		"start_date":   "07-2025",
		"end_date":     "12-2025",
	}
	w := postJSON("/subscriptions", validBody)
	assert.Equal(t, http.StatusCreated, w.Code)

	var sub models.Subscription
	err := json.Unmarshal(w.Body.Bytes(), &sub)
	require.NoError(t, err)
	assert.Equal(t, "Yandex Plus", sub.ServiceName)
	assert.Equal(t, 400, sub.Price)
	assert.Equal(t, "60601fee-2bf1-4721-ae6f-7636e79a0cba", sub.UserID.String())
	assert.Equal(t, "2025-07-01", sub.StartDate.Format("2006-01-02"))
	require.NotNil(t, sub.EndDate)
	assert.Equal(t, "2025-12-01", sub.EndDate.Format("2006-01-02"))

	// Обязательное поле service_name
	invalidBody := map[string]interface{}{
		"price":      400,
		"user_id":    "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		"start_date": "07-2025",
	}
	w2 := postJSON("/subscriptions", invalidBody)
	assert.Equal(t, http.StatusBadRequest, w2.Code)

	// Неверный UUID
	badUUID := map[string]interface{}{
		"service_name": "Test",
		"price":        100,
		"user_id":      "not-a-uuid",
		"start_date":   "01-2025",
	}
	w3 := postJSON("/subscriptions", badUUID)
	assert.Equal(t, http.StatusBadRequest, w3.Code)

	// Отрицательная цена
	negativePrice := map[string]interface{}{
		"service_name": "Test",
		"price":        -10,
		"user_id":      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		"start_date":   "01-2025",
	}
	w4 := postJSON("/subscriptions", negativePrice)
	assert.Equal(t, http.StatusBadRequest, w4.Code)

	// end_date раньше start_date
	wrongPeriod := map[string]interface{}{
		"service_name": "Test",
		"price":        100,
		"user_id":      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		"start_date":   "12-2025",
		"end_date":     "05-2025",
	}
	w5 := postJSON("/subscriptions", wrongPeriod)
	assert.Equal(t, http.StatusBadRequest, w5.Code)
}

func TestGetSubscription(t *testing.T) {
	clearAllSubscriptions()

	createBody := map[string]interface{}{
		"service_name": "Netflix",
		"price":        599,
		"user_id":      "550e8400-e29b-41d4-a716-446655440000",
		"start_date":   "01-2025",
	}
	wCreate := postJSON("/subscriptions", createBody)
	require.Equal(t, http.StatusCreated, wCreate.Code)
	var created models.Subscription
	json.Unmarshal(wCreate.Body.Bytes(), &created)

	wGet := get("/subscriptions/" + strconv.FormatInt(created.ID, 10))
	assert.Equal(t, http.StatusOK, wGet.Code)
	var fetched models.Subscription
	json.Unmarshal(wGet.Body.Bytes(), &fetched)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, "Netflix", fetched.ServiceName)

	// Несуществующий ID
	wNotFound := get("/subscriptions/99999")
	assert.Equal(t, http.StatusNotFound, wNotFound.Code)

	// Некорректный ID, не число
	wInvalid := get("/subscriptions/abc")
	assert.Equal(t, http.StatusBadRequest, wInvalid.Code)
}

func TestUpdateSubscription(t *testing.T) {
	clearAllSubscriptions()

	createBody := map[string]interface{}{
		"service_name": "Spotify",
		"price":        399,
		"user_id":      "550e8400-e29b-41d4-a716-446655440000",
		"start_date":   "03-2025",
	}
	wCreate := postJSON("/subscriptions", createBody)
	require.Equal(t, http.StatusCreated, wCreate.Code)
	var created models.Subscription
	json.Unmarshal(wCreate.Body.Bytes(), &created)

	// Обновление цены
	patchBody := map[string]interface{}{"price": 499}
	wPatch := patchJSON("/subscriptions/"+strconv.FormatInt(created.ID, 10), patchBody)
	assert.Equal(t, http.StatusOK, wPatch.Code)
	var updated models.Subscription
	json.Unmarshal(wPatch.Body.Bytes(), &updated)
	assert.Equal(t, 499, updated.Price)

	// Обновление даты окончания
	patchEnd := map[string]interface{}{"end_date": "12-2025"}
	wPatch2 := patchJSON("/subscriptions/"+strconv.FormatInt(created.ID, 10), patchEnd)
	assert.Equal(t, http.StatusOK, wPatch2.Code)
	json.Unmarshal(wPatch2.Body.Bytes(), &updated)
	require.NotNil(t, updated.EndDate)
	assert.Equal(t, "2025-12-01", updated.EndDate.Format("2006-01-02"))

	// Пустое тело
	wEmpty := patchJSON("/subscriptions/"+strconv.FormatInt(created.ID, 10), map[string]interface{}{})
	assert.Equal(t, http.StatusBadRequest, wEmpty.Code)

	// Неверный ID
	wBadID := patchJSON("/subscriptions/99999", patchBody)
	assert.Equal(t, http.StatusNotFound, wBadID.Code)
}

func TestDeleteSubscription(t *testing.T) {
	clearAllSubscriptions()

	createBody := map[string]interface{}{
		"service_name": "Apple Music",
		"price":        299,
		"user_id":      "550e8400-e29b-41d4-a716-446655440000",
		"start_date":   "06-2025",
	}
	wCreate := postJSON("/subscriptions", createBody)
	require.Equal(t, http.StatusCreated, wCreate.Code)
	var created models.Subscription
	json.Unmarshal(wCreate.Body.Bytes(), &created)

	wDel := deleteReq("/subscriptions/" + strconv.FormatInt(created.ID, 10))
	assert.Equal(t, http.StatusNoContent, wDel.Code)

	// Повторное удаление
	wDelAgain := deleteReq("/subscriptions/" + strconv.FormatInt(created.ID, 10))
	assert.Equal(t, http.StatusNotFound, wDelAgain.Code)

	// Несуществующий ID
	wNotFound := deleteReq("/subscriptions/12345")
	assert.Equal(t, http.StatusNotFound, wNotFound.Code)
}

func TestListSubscriptions(t *testing.T) {
	clearAllSubscriptions()

	user1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	user2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	subs := []map[string]interface{}{
		{"service_name": "A", "price": 100, "user_id": user1.String(), "start_date": "01-2025"},
		{"service_name": "B", "price": 200, "user_id": user1.String(), "start_date": "02-2025"},
		{"service_name": "C", "price": 300, "user_id": user2.String(), "start_date": "03-2025"},
	}
	for _, s := range subs {
		w := postJSON("/subscriptions", s)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	// Фильтр по user1
	w := get("/subscriptions?user_id=" + user1.String())
	assert.Equal(t, http.StatusOK, w.Code)
	var list []models.Subscription
	json.Unmarshal(w.Body.Bytes(), &list)
	assert.Len(t, list, 2)

	// Фильтр по user1 и service_name = "A"
	w2 := get("/subscriptions?user_id=" + user1.String() + "&service_name=A")
	assert.Equal(t, http.StatusOK, w2.Code)
	json.Unmarshal(w2.Body.Bytes(), &list)
	assert.Len(t, list, 1)
	assert.Equal(t, "A", list[0].ServiceName)

	// Несуществующий user_id
	unknown := uuid.New()
	w3 := get("/subscriptions?user_id=" + unknown.String())
	assert.Equal(t, http.StatusOK, w3.Code)
	json.Unmarshal(w3.Body.Bytes(), &list)
	assert.Empty(t, list)

	// Отсутствует user_id
	w4 := get("/subscriptions")
	assert.Equal(t, http.StatusBadRequest, w4.Code)
}
