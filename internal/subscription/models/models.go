package models

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type SubscriptionListDB struct {
	UserID      uuid.UUID   `db:"user_id"`
	StartDate   pgtype.Date `db:"start_date"`
	Price       int         `db:"price"`
	ServiceName string      `db:"service_name"`
}

type SubscriptionListDTO struct {
	StartDate   pgtype.Date
	Price       int
	ServiceName string
}

type SubscriptionListJSON struct {
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	Price       int       `json:"price"`
	ServiceName string    `json:"service_name"`
}

type SubscriptionListToCostJSON struct {
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
	ServiceName []string  `json:"service_name"`
}
