package storage

import (
	"context"
	"github.com/Ostmind/subscriptionservice/internal/subscription/models"
	"github.com/google/uuid"
)

type Repository interface {
	GetSubscriptionListByUserID(ctx context.Context, id uuid.UUID) ([]models.SubscriptionListDB, error)
	PostSubscription(ctx context.Context, sub models.SubscriptionListJSON) error
	DeleteSubscription(ctx context.Context, id int) error
	UpdateSubscription(ctx context.Context, sub models.SubscriptionListJSON, id int) error
	GetTotalPeriodCostByDatesAndServiceName(ctx context.Context, subList models.SubscriptionListToCostJSON) (int, error)
}
