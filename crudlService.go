package main

// @title           Subscriptions API
// @version         1.0
// @description     REST сервис для управления подписками пользователей.

// @contact.name   Алексей
// @contact.url    https://github.com/EvilCommunist

// @license.name   Apache 2.0
// @license.url    http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8090
// @BasePath  /

// @schemes   http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"CRUDL/db"
	"CRUDL/db/crud"
	"CRUDL/handlers"

	httpSwagger "github.com/swaggo/http-swagger"

	_ "CRUDL/docs"
)

const (
	subscriptions = "/subscriptions"
	id            = "/{id}"
)

func main() {
	router := chi.NewRouter()
	conn, err := db.Connect()
	if err != nil {
		slog.Error("Unable to connect to database", "error", err)
		return
	}
	router.Get("/swagger/*", httpSwagger.WrapHandler)

	subDBHandler := &crud.SubscriptionDB{}
	subDBHandler.SetDatabase(conn)
	subHandler := handlers.SubscriptionsHandler{}
	subHandler.SetDatabaseHandler(subDBHandler)

	router.Post(subscriptions, subHandler.Post)
	router.Get(subscriptions+id, subHandler.Get)
	router.Patch(subscriptions+id, subHandler.Patch)
	router.Delete(subscriptions+id, subHandler.Delete)
	router.Get(subscriptions, subHandler.List)
	router.Get(subscriptions+"/totalCost", subHandler.TotalCost)

	http.ListenAndServe(":8090", router)
}
