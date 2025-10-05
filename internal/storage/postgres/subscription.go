package postgres

import (
	"context"
	"fmt"
	"github.com/Ostmind/subscriptionservice/internal/subscription/models"
	"github.com/google/uuid"
	"time"

	_ "github.com/lib/pq"
)

func (store *Storage) GetSubscriptionListByUserID(ctx context.Context, id uuid.UUID) (subscription []models.SubscriptionListDB, err error) {
	sqlStatement := `SELECT price, start_date, service_name FROM public.subscription WHERE user_id = $1;`

	rows, err := store.DB.Query(ctx, sqlStatement, id)
	if err != nil {
		return subscription, fmt.Errorf("failed to query DB %w", err)
	}
	defer rows.Close()

	found := false

	for rows.Next() {
		var t models.SubscriptionListDB

		if err := rows.Scan(&t.Price, &t.StartDate, &t.ServiceName); err != nil {
			return subscription, fmt.Errorf("scan Subscription List: %w", err)
		}

		subscription = append(subscription, t)
		found = true
	}

	if !found {
		return subscription, models.ErrNotFound
	}

	return subscription, nil
}

func (store *Storage) PostSubscription(ctx context.Context, sub models.SubscriptionListJSON) error {
	sqlStatement := `INSERT INTO subscription 
    				 (user_id, start_date, price, service_name) 
					 VALUES($1,$2,$3,$4);`

	startDateDB, err := time.Parse("02-2006", sub.StartDate)
	if err != nil {
		return fmt.Errorf("error adding to DB %w", err)
	}

	result, err := store.DB.Exec(ctx, sqlStatement, sub.UserID, startDateDB, sub.Price, sub.ServiceName)
	if err != nil {
		if !result.Insert() {
			return models.ErrUnique
		}

		return fmt.Errorf("error adding to DB %w", err)
	}
	return nil
}

func (store *Storage) DeleteSubscription(ctx context.Context, id int) error {
	sqlStatement := `DELETE FROM public.subscription WHERE id = $1;`

	result, err := store.DB.Exec(ctx, sqlStatement, id)
	if err != nil {
		return fmt.Errorf("error deleting from DB %w", err)
	}

	if result.RowsAffected() == 0 {
		return models.ErrNotFound
	}

	return nil
}

func (store *Storage) UpdateSubscription(ctx context.Context, sub models.SubscriptionListJSON, id int) error {
	sqlStatement := `UPDATE public.subscription SET 
                     user_id =$1, 
                     start_date=$2, 
                     price=$3, 
                     service_name=$4
                     WHERE id =$5;`

	startDateDB, err := time.Parse("02-2006", sub.StartDate)
	if err != nil {
		return fmt.Errorf("error adding to DB %w", err)
	}

	result, err := store.DB.Exec(ctx, sqlStatement, sub.UserID, startDateDB, sub.Price, sub.ServiceName, id)
	if err != nil {
		return fmt.Errorf("error updating DB %w", err)
	}

	if result.RowsAffected() == 0 {
		return models.ErrNotFound
	}

	return nil
}

func (store *Storage) GetTotalPeriodCostByDatesAndServiceName(ctx context.Context, subList models.SubscriptionListToCostJSON) (res int, err error) {
	sqlStatement := `SELECT SUM(price) 
				     FROM public.subscription 
	                 where user_id = $1
	                 and start_date >=$2
	                 and start_date <=$3`

	startDateDB, err := time.Parse("02-2006", subList.StartDate)
	if err != nil {
		return res, fmt.Errorf("error adding to DB %w", err)
	}

	endDateDB, err := time.Parse("02-2006", subList.EndDate)
	if err != nil {
		return res, fmt.Errorf("error adding to DB %w", err)
	}

	if len(subList.ServiceName) > 0 {
		sqlStatement += " and service_name = ANY($4)"

		rows := store.DB.QueryRow(ctx, sqlStatement, subList.UserID, startDateDB, endDateDB, subList.ServiceName)

		err = rows.Scan(&res)
		if err != nil {
			return res, fmt.Errorf("failed to parse DB %w", err)
		}

		return res, nil
	}

	rows := store.DB.QueryRow(ctx, sqlStatement, subList.UserID, startDateDB, endDateDB)

	err = rows.Scan(&res)
	if err != nil {
		return res, fmt.Errorf("failed to parse DB %w", err)
	}

	return res, nil
}
