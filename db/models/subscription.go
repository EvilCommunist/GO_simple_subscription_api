package models

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          int64
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
}

func (sub *Subscription) GetTableName() string {
	return "subscriptions"
}

type SubscriptionUpdate struct {
	ID        int64
	Price     *int
	StartDate *time.Time
	EndDate   *time.Time
}
