package server

import (
	"context"
	"errors"
	"github.com/Ostmind/subscriptionservice/internal/storage"
	"github.com/Ostmind/subscriptionservice/internal/subscription/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"strconv"
)

//go:generate mockgen -source=handlers.go -destination=mock/handlersrepository.go
type subscriptionManager interface {
	GetSubscriptionListByUserID(ctx context.Context, id uuid.UUID) ([]models.SubscriptionListDB, error)
	PostSubscription(ctx context.Context, sub models.SubscriptionListJSON) error
	UpdateSubscription(ctx context.Context, sub models.SubscriptionListJSON, id int) error
	DeleteSubscription(ctx context.Context, id int) error
	GetTotalPeriodCostByDatesAndServiceName(ctx context.Context, subList models.SubscriptionListToCostJSON) (int, error)
}

type controller struct {
	manager subscriptionManager
	logger  *slog.Logger
}

func NewSubscriptionHandler(manager storage.Repository, log *slog.Logger) *controller {
	return &controller{manager, log}
}

func (ctr controller) GetSubscriptionListByUserID(echo echo.Context) error {
	ctr.logger.Debug("Get Request for Subscription List")

	userCookie, err := echo.Cookie("userId")
	if err != nil {
		return echo.NoContent(http.StatusBadRequest)
	}

	uuidID, err := uuid.Parse(userCookie.Value)
	if err != nil {
		return echo.NoContent(http.StatusBadRequest)
	}

	res, err := ctr.manager.GetSubscriptionListByUserID(echo.Request().Context(), uuidID)
	if err != nil {

		if errors.Is(err, models.ErrNotFound) {
			return echo.NoContent(http.StatusNotFound)
		}
		return echo.NoContent(http.StatusInternalServerError)
	}

	var dtoSubList []models.SubscriptionListDTO

	for _, v := range res {
		listDTO := models.SubscriptionListDTO{StartDate: v.StartDate,
			Price:       v.Price,
			ServiceName: v.ServiceName,
		}
		dtoSubList = append(dtoSubList, listDTO)
	}

	return echo.JSON(http.StatusOK, dtoSubList)
}

func (ctr controller) PostSubscription(echo echo.Context) error {
	ctr.logger.Debug("Get Request for POST Subscription")

	var sub models.SubscriptionListJSON

	if err := echo.Bind(&sub); err != nil {
		return echo.NoContent(http.StatusBadRequest)
	}

	if err := ctr.manager.PostSubscription(echo.Request().Context(), sub); err != nil {
		if errors.Is(err, models.ErrUnique) {

			return echo.NoContent(http.StatusBadRequest)
		}

		return echo.NoContent(http.StatusInternalServerError)
	}

	return echo.NoContent(http.StatusOK)
}

func (ctr controller) DeleteSubscription(echo echo.Context) error {
	ctr.logger.Debug("Get Request for Delete Subscription")

	idStr := echo.QueryParam("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return echo.NoContent(http.StatusInternalServerError)
	}

	if err := ctr.manager.DeleteSubscription(echo.Request().Context(), id); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return echo.NoContent(http.StatusBadRequest)
		}
		return echo.NoContent(http.StatusInternalServerError)
	}

	return echo.NoContent(http.StatusOK)
}

func (ctr controller) UpdateSubscription(echo echo.Context) error {
	ctr.logger.Debug("Get Request for Update Subscription")

	var sub models.SubscriptionListJSON

	if err := echo.Bind(&sub); err != nil {
		return echo.NoContent(http.StatusBadRequest)
	}

	idStr := echo.QueryParam("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return echo.NoContent(http.StatusInternalServerError)
	}

	if err := ctr.manager.UpdateSubscription(echo.Request().Context(), sub, id); err != nil {
		return echo.NoContent(http.StatusInternalServerError)
	}

	return echo.NoContent(http.StatusOK)
}

func (ctr controller) GetTotalPeriodCostByDatesAndServiceName(echo echo.Context) error {
	ctr.logger.Debug("Get Request for Subscription Cost")

	var sub models.SubscriptionListToCostJSON

	if err := echo.Bind(&sub); err != nil {
		return echo.NoContent(http.StatusBadRequest)
	}

	res, err := ctr.manager.GetTotalPeriodCostByDatesAndServiceName(echo.Request().Context(), sub)
	if err != nil {
		return echo.NoContent(http.StatusInternalServerError)
	}

	return echo.JSON(http.StatusOK, map[string]int{"result": res})
}
