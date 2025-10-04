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

// GetSubscriptionListByUserID godoc
// @Summary Получить список подписок
// @Description Получает список подписок пользователя по userId из cookie
// @Tags subscriptions
// @Produce json
// @Param userId header string true "userId из cookie"
// @Success 200 {array} models.SubscriptionListDTO
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /subscription/users [get]
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

// PostSubscription godoc
// @Summary Создать новую подписку
// @Description Добавляет подписку на сервис
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body models.SubscriptionListJSON true "Данные подписки"
// @Success 200 {string} string "Подписка успешно создана"
// @Failure 400 {string} string "Неправильный запрос или дубликат подписки"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /subscriptions [post]
func (ctr controller) PostSubscription(echo echo.Context) error {
	ctr.logger.Debug("Get Request for POST Subscription")

	var sub models.SubscriptionListJSON

	if err := echo.Bind(&sub); err != nil {
		return echo.NoContent(http.StatusBadRequest)
	}

	if err := ctr.manager.PostSubscription(echo.Request().Context(), sub); err != nil {
		if errors.Is(err, models.ErrUnique) {

			return echo.JSON(http.StatusBadRequest, map[string]string{"result": "Неправильный запрос или дубликат подписки"})
		}

		return echo.JSON(http.StatusInternalServerError, map[string]string{"result": "Неправильный запрос или дубликат подписки"})
	}

	return echo.JSON(http.StatusOK, map[string]string{"result": "Подписка успешно создана"})
}

// DeleteSubscription godoc
// @Summary Удалить подписку
// @Description Удаление подписки по id из query-параметров
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id query int true "ID подписки для удаления"
// @Success 200 {string} string "Подписка успешно удалена"
// @Failure 400 {string} string "Некорректный id или не найдена подписка"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /subscriptions [delete]
func (ctr controller) DeleteSubscription(echo echo.Context) error {
	ctr.logger.Debug("Get Request for Delete Subscription")

	idStr := echo.QueryParam("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return echo.NoContent(http.StatusInternalServerError)
	}

	if err := ctr.manager.DeleteSubscription(echo.Request().Context(), id); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return echo.JSON(http.StatusBadRequest, map[string]string{"result": "Некорректный id или не найдена подписка"})
		}
		return echo.JSON(http.StatusInternalServerError, map[string]string{"result": "Внутренняя ошибка сервера"})
	}

	return echo.JSON(http.StatusOK, map[string]string{"result": "Подписка успешно удалена"})
}

// UpdateSubscription godoc
// @Summary Обновить подписку
// @Description Обновляет подписку по id переданному в query-параметрах
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id query int true "ID подписки для обновления"
// @Param subscription body models.SubscriptionListJSON true "Данные подписки для обновления"
// @Success 200 {string} string "Подписка успешно обновлена"
// @Failure 400 {string} string "Неправильный запрос или невалидные данные"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /subscriptions [put]
func (ctr controller) UpdateSubscription(echo echo.Context) error {
	ctr.logger.Debug("Get Request for Update Subscription")

	var sub models.SubscriptionListJSON

	if err := echo.Bind(&sub); err != nil {
		return echo.NoContent(http.StatusBadRequest)
	}

	idStr := echo.QueryParam("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return echo.JSON(http.StatusInternalServerError, map[string]string{"result": "Неправильный запрос или невалидные данные"})
	}

	if err := ctr.manager.UpdateSubscription(echo.Request().Context(), sub, id); err != nil {
		return echo.JSON(http.StatusInternalServerError, map[string]string{"result": "Внутренняя ошибка сервера"})
	}

	return echo.JSON(http.StatusOK, map[string]string{"result": "Подписка успешно обновлена"})
}

// GetTotalPeriodCostByDatesAndServiceName godoc
// @Summary Получить сумму стоимости по датам и имени сервиса
// @Description Возвращает общую стоимость подписок на сервис в указанный период
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param periodCost body models.SubscriptionListToCostJSON true "Параметры периода и имени сервиса"
// @Success 200 {object} map[string]int "Общая стоимость в поле result"
// @Failure 400 {string} string "Ошибка в параметрах запроса"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /subscriptions/cost [post]
func (ctr controller) GetTotalPeriodCostByDatesAndServiceName(echo echo.Context) error {
	ctr.logger.Debug("Get Request for Subscription Cost")

	var sub models.SubscriptionListToCostJSON

	if err := echo.Bind(&sub); err != nil {
		return echo.JSON(http.StatusBadRequest, map[string]string{"result": "Ошибка в параметрах запроса"})
	}

	res, err := ctr.manager.GetTotalPeriodCostByDatesAndServiceName(echo.Request().Context(), sub)
	if err != nil {
		return echo.JSON(http.StatusInternalServerError, map[string]string{"result": "Внутренняя ошибка сервера"})
	}

	return echo.JSON(http.StatusOK, map[string]int{"result": res})
}
