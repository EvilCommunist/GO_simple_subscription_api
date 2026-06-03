package crud

import (
	"CRUDL/db/models"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionDB struct {
	conn *pgxpool.Pool
}

func (sdb *SubscriptionDB) SetDatabase(conn *pgxpool.Pool) {
	sdb.conn = conn
}

func (sdb *SubscriptionDB) Create(sub models.Subscription, ctx context.Context) (*models.Subscription, error) {
	created := models.Subscription{}
	err := sdb.conn.QueryRow(ctx,
		fmt.Sprintf(`insert into %s (service_name, price, user_id, start_date, end_date)
		values ($1, $2, $3, $4, $5)
		returning id, service_name, price, user_id, start_date, end_date`,
			sub.GetTableName()), sub.ServiceName, sub.Price, sub.UserID, sub.StartDate,
		sub.EndDate).Scan(&created.ID, &created.ServiceName, &created.Price,
		&created.UserID, &created.StartDate, &created.EndDate)
	if err != nil {
		slog.Error("Query refused with error", "error", err)
		return nil, fmt.Errorf("Query refused with error %v", err)
	}
	return &created, nil
}

func (sdb *SubscriptionDB) Read(id int64, ctx context.Context) (*models.Subscription, error) {
	if id < 0 {
		return nil, fmt.Errorf("ID cannot be negative")
	}
	sub := models.Subscription{
		ID: id,
	}
	err := sdb.conn.QueryRow(ctx, fmt.Sprintf(`select service_name, price, user_id, 
	start_date, end_date from %s where id=$1`, sub.GetTableName()),
		id).Scan(&sub.ServiceName, &sub.Price,
		&sub.UserID, &sub.StartDate, &sub.EndDate)
	if err != nil {
		slog.Error("Query refused with error", "error", err)
		return nil, fmt.Errorf("Query refused with error %v", err)
	}
	return &sub, nil
}

func (sdb *SubscriptionDB) Update(sub models.SubscriptionUpdate, ctx context.Context) (*models.Subscription, error) {
	parametersSQL := []string{}
	parameters := []interface{}{}
	if sub.Price != nil {
		parameters = append(parameters, sub.Price)
		parametersSQL = append(parametersSQL, "price")
	}
	if sub.StartDate != nil {
		parameters = append(parameters, sub.StartDate)
		parametersSQL = append(parametersSQL, "start_date")
	}
	if sub.EndDate != nil {
		parameters = append(parameters, sub.EndDate)
		parametersSQL = append(parametersSQL, "end_date")
	}

	if len(parameters) == 0 {
		slog.Error("No parameters passed to update")
		return nil, fmt.Errorf("No parameters were passed to update")
	}

	querySet := ""
	for counter, name := range parametersSQL {
		querySet += fmt.Sprintf("%s = $%d, ", name, counter+1)
	}

	querySet = strings.TrimSuffix(querySet, ", ")
	subscriptionUpdated := models.Subscription{}
	err := sdb.conn.QueryRow(ctx, fmt.Sprintf(`update %s set %s
	where id=$%d returning id, service_name, price, user_id, start_date, end_date`,
		subscriptionUpdated.GetTableName(), querySet, len(parameters)+1),
		append(parameters, sub.ID)...).Scan(&subscriptionUpdated.ID, &subscriptionUpdated.ServiceName,
		&subscriptionUpdated.Price, &subscriptionUpdated.UserID,
		&subscriptionUpdated.StartDate, &subscriptionUpdated.EndDate)

	if err != nil {
		slog.Error("Query refused with error", "error", err)
		return nil, fmt.Errorf("Query refused with error %v", err)
	}
	return &subscriptionUpdated, nil
}

func (sdb *SubscriptionDB) Delete(id int64, ctx context.Context) error {
	sub := models.Subscription{}
	res, err := sdb.conn.Exec(ctx, fmt.Sprintf(`delete from %s where id=$1`,
		sub.GetTableName()), id)
	if err != nil {
		slog.Error("Query refused with error", "error", err)
		return fmt.Errorf("Query refused with error %v", err)
	}
	if res.RowsAffected() < 1 {
		slog.Error("Nothing was deleted")
		return fmt.Errorf("Nothing was deleted")
	}
	return nil
}

func (sdb *SubscriptionDB) List(ctx context.Context, userID uuid.UUID, serviceName *string) ([]models.Subscription, error) {
	queryCond := "\nwhere user_id = $1"
	args := []interface{}{}
	args = append(args, userID)
	if serviceName != nil {
		if *serviceName == "" {
			slog.Error("Service name cannot be empty")
			return nil, fmt.Errorf("Service name cannot be empty")
		}
		queryCond += " and service_name = $2"
		args = append(args, *serviceName)
	}
	sub := models.Subscription{}
	res, err := sdb.conn.Query(ctx, fmt.Sprintf(`select id, service_name, price, user_id, start_date, end_date
	from %s`+queryCond, sub.GetTableName()), args...)
	if err != nil {
		slog.Error("Query refused with error", "error", err)
		return nil, fmt.Errorf("Query refused with error %v", err)
	}
	defer res.Close()

	subscriptions := []models.Subscription{}
	for res.Next() {
		subscription := models.Subscription{}

		err := res.Scan(&subscription.ID, &subscription.ServiceName,
			&subscription.Price, &subscription.UserID,
			&subscription.StartDate, &subscription.EndDate)
		if err != nil {
			slog.Error("Scan dropped with error", "error", err)
			return nil, fmt.Errorf("Scan dropped with error %v", err)
		}

		subscriptions = append(subscriptions, subscription)
	}
	if err = res.Err(); err != nil {
		return nil, fmt.Errorf("Error of subscriptions iteration %v", err)
	}

	return subscriptions, nil
}
