package main

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"CRUDL/db"
	"CRUDL/db/crud"
	"CRUDL/handlers"
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

	subDBHandler := &crud.SubscriptionDB{}
	subDBHandler.SetDatabase(conn)
	subHandler := handlers.SubscriptionsHandler{}
	subHandler.SetDatabaseHandler(subDBHandler)

	router.Post(subscriptions, subHandler.Post)
	router.Get(subscriptions+id, subHandler.Get)
	router.Patch(subscriptions+id, subHandler.Patch)
	router.Delete(subscriptions+id, subHandler.Delete)
	router.Get(subscriptions, subHandler.List)

	http.ListenAndServe(":8090", router)
}
