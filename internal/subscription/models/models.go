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
	StartDate   pgtype.Date `json:"start_date"   example:"09-2025"`
	Price       int         `json:"price"        example:"400"`
	ServiceName string      `json:"service_name" example:"Netflix"`
}

type SubscriptionListJSON struct {
	UserID      uuid.UUID `json:"user_id"      example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string    `json:"start_date"   example:"09-2025"`
	Price       int       `json:"price"        example:"400"`
	ServiceName string    `json:"service_name" example:"Netflix"`
}

type SubscriptionListToCostJSON struct {
	UserID      uuid.UUID `json:"user_id"      example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string    `json:"start_date"   example:"09-2025"`
	EndDate     string    `json:"end_date"     example:"12-2025"`
	ServiceName []string  `json:"service_name" example:"Netflix,Yandex Plus,Spotify"`
}
