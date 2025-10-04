package server_test

import (
	"errors"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Ostmind/subscriptionservice/internal/subscription/models"
	"github.com/Ostmind/subscriptionservice/internal/subscription/server/server"
	"github.com/Ostmind/subscriptionservice/internal/subscription/server/server/mock"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"log/slog"
)

func TestGetSubscriptionListByUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	e := echo.New()

	userID := uuid.MustParse("d4ae2ec1-3673-45c8-b823-7b28c99baff0")

	t.Run("Success", func(t *testing.T) {
		mockManager := mock_server.NewMocksubscriptionManager(ctrl)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "userId", Value: userID.String()})
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockManager.EXPECT().
			GetSubscriptionListByUserID(gomock.Any(), userID).
			Return([]models.SubscriptionListDB{
				{
					StartDate: func() pgtype.Date {
						var d pgtype.Date
						d.Time = time.Date(2023, 9, 0, 0, 0, 0, 0, time.UTC)
						d.Valid = true
						return d
					}(),
					Price:       100,
					ServiceName: "Spotify",
				},
				{
					StartDate: func() pgtype.Date {
						var d pgtype.Date
						d.Time = time.Date(2023, 10, 0, 0, 0, 0, 0, time.UTC)
						d.Valid = true
						return d
					}(),
					Price:       200,
					ServiceName: "Netflix",
				},
			}, nil)

		handler := server.NewSubscriptionHandler(mockManager, logger)
		if err := handler.GetSubscriptionListByUserID(c); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("ErrNotFound", func(t *testing.T) {
		mockManager := mock_server.NewMocksubscriptionManager(ctrl)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "userId", Value: userID.String()})
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockManager.EXPECT().
			GetSubscriptionListByUserID(gomock.Any(), userID).
			Return(nil, models.ErrNotFound)

		handler := server.NewSubscriptionHandler(mockManager, logger)
		if err := handler.GetSubscriptionListByUserID(c); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockManager := mock_server.NewMocksubscriptionManager(ctrl)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "userId", Value: userID.String()})
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockManager.EXPECT().
			GetSubscriptionListByUserID(gomock.Any(), userID).
			Return(nil, errors.New("db error"))

		handler := server.NewSubscriptionHandler(mockManager, logger)
		if err := handler.GetSubscriptionListByUserID(c); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})
}

func TestPostSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	e := echo.New()

	t.Run("BadRequest_BadJSON", func(t *testing.T) {
		mockManager := mock_server.NewMocksubscriptionManager(ctrl)

		body := strings.NewReader(`invalid json`)
		req := httptest.NewRequest(http.MethodPost, "/", body)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := server.NewSubscriptionHandler(mockManager, logger)

		if err := handler.PostSubscription(c); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("ErrUnique", func(t *testing.T) {
		mockManager := mock_server.NewMocksubscriptionManager(ctrl)

		jsonBody := `{"service_name": "Spotify", "price": 100, "start_date": "2023-09"}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockManager.EXPECT().
			PostSubscription(gomock.Any(), gomock.Any()).
			Return(models.ErrUnique)

		handler := server.NewSubscriptionHandler(mockManager, logger)

		if err := handler.PostSubscription(c); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockManager := mock_server.NewMocksubscriptionManager(ctrl)

		jsonBody := `{"service_name": "Spotify", "price": 100, "start_date": "2023-09"}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockManager.EXPECT().
			PostSubscription(gomock.Any(), gomock.Any()).
			Return(errors.New("db error"))

		handler := server.NewSubscriptionHandler(mockManager, logger)

		if err := handler.PostSubscription(c); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("Success", func(t *testing.T) {
		mockManager := mock_server.NewMocksubscriptionManager(ctrl)

		jsonBody := `{"service_name": "Spotify", "price": 100, "start_date": "2023-09"}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockManager.EXPECT().
			PostSubscription(gomock.Any(), gomock.Any()).
			Return(nil)

		handler := server.NewSubscriptionHandler(mockManager, logger)

		if err := handler.PostSubscription(c); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected %d, got %d", http.StatusOK, rec.Code)
		}
	})
}

func TestDeleteSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	e := echo.New()

	t.Run("Success", func(t *testing.T) {
		mockManager := mock_server.NewMocksubscriptionManager(ctrl)

		req := httptest.NewRequest(http.MethodDelete, "/?id=123", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockManager.EXPECT().
			DeleteSubscription(gomock.Any(), 123).
			Return(nil)

		handler := server.NewSubscriptionHandler(mockManager, logger)
		if err := handler.DeleteSubscription(c); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("BadRequest_InvalidID", func(t *testing.T) {
		mockManager := mock_server.NewMocksubscriptionManager(ctrl)

		req := httptest.NewRequest(http.MethodDelete, "/?id=abc", nil) // invalid int
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := server.NewSubscriptionHandler(mockManager, logger)
		if err := handler.DeleteSubscription(c); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		mockManager := mock_server.NewMocksubscriptionManager(ctrl)

		req := httptest.NewRequest(http.MethodDelete, "/?id=123", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockManager.EXPECT().
			DeleteSubscription(gomock.Any(), 123).
			Return(models.ErrNotFound)

		handler := server.NewSubscriptionHandler(mockManager, logger)
		if err := handler.DeleteSubscription(c); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockManager := mock_server.NewMocksubscriptionManager(ctrl)

		req := httptest.NewRequest(http.MethodDelete, "/?id=123", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockManager.EXPECT().
			DeleteSubscription(gomock.Any(), 123).
			Return(errors.New("db error"))

		handler := server.NewSubscriptionHandler(mockManager, logger)
		if err := handler.DeleteSubscription(c); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})
}

func TestUpdateSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	e := echo.New()

	t.Run("Success", func(t *testing.T) {
		mockManager := mock_server.NewMocksubscriptionManager(ctrl)

		jsonBody := `{"service_name": "Spotify", "price": 300, "start_date": "2025-09"}`
		req := httptest.NewRequest(http.MethodPut, "/?id=1", strings.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockManager.EXPECT().
			UpdateSubscription(gomock.Any(), gomock.Any(), 1).
			Return(nil)

		handler := server.NewSubscriptionHandler(mockManager, logger)

		if err := handler.UpdateSubscription(c); err != nil {
			t.Fatal(err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("BadRequest_InvalidID", func(t *testing.T) {
		mockManager := mock_server.NewMocksubscriptionManager(ctrl)

		jsonBody := `{"service_name": "Spotify", "price": 300, "start_date": "2025-09"}`
		req := httptest.NewRequest(http.MethodPut, "/?id=abc", strings.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := server.NewSubscriptionHandler(mockManager, logger)

		if err := handler.UpdateSubscription(c); err != nil {
			t.Fatal(err)
		}

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("BadRequest_BindError", func(t *testing.T) {
		mockManager := mock_server.NewMocksubscriptionManager(ctrl)

		req := httptest.NewRequest(http.MethodPut, "/?id=1", strings.NewReader(`invalid json`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := server.NewSubscriptionHandler(mockManager, logger)

		if err := handler.UpdateSubscription(c); err != nil {
			t.Fatal(err)
		}

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("InternalServerError_ManagerError", func(t *testing.T) {
		mockManager := mock_server.NewMocksubscriptionManager(ctrl)

		jsonBody := `{"service_name": "Spotify", "price": 300, "start_date": "2025-09"}`
		req := httptest.NewRequest(http.MethodPut, "/?id=1", strings.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockManager.EXPECT().
			UpdateSubscription(gomock.Any(), gomock.Any(), 1).
			Return(errors.New("db error"))

		handler := server.NewSubscriptionHandler(mockManager, logger)

		if err := handler.UpdateSubscription(c); err != nil {
			t.Fatal(err)
		}

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})
}

func TestGetTotalPeriodCostByDatesAndServiceName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	e := echo.New()

	t.Run("Success", func(t *testing.T) {
		mockManager := mock_server.NewMocksubscriptionManager(ctrl)

		jsonBody := `{"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"09-2025","end_date":"12-2025","service_name":["Spotify"]}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedResult := 1200

		mockManager.EXPECT().
			GetTotalPeriodCostByDatesAndServiceName(gomock.Any(), gomock.Any()).
			Return(expectedResult, nil)

		handler := server.NewSubscriptionHandler(mockManager, logger)
		if err := handler.GetTotalPeriodCostByDatesAndServiceName(c); err != nil {
			t.Fatal(err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		expectedJSON := `{"result":1200}`
		if strings.TrimSpace(rec.Body.String()) != expectedJSON {
			t.Errorf("expected body %s, got %s", expectedJSON, rec.Body.String())
		}
	})

	t.Run("BadRequest_BindError", func(t *testing.T) {
		mockManager := mock_server.NewMocksubscriptionManager(ctrl)

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`invalid json`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := server.NewSubscriptionHandler(mockManager, logger)

		if err := handler.GetTotalPeriodCostByDatesAndServiceName(c); err != nil {
			t.Fatal(err)
		}

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("InternalServerError_ManagerError", func(t *testing.T) {
		mockManager := mock_server.NewMocksubscriptionManager(ctrl)

		jsonBody := `{"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"09-2025","end_date":"12-2025","service_name":["Spotify"]}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockManager.EXPECT().
			GetTotalPeriodCostByDatesAndServiceName(gomock.Any(), gomock.Any()).
			Return(0, errors.New("db error"))

		handler := server.NewSubscriptionHandler(mockManager, logger)

		if err := handler.GetTotalPeriodCostByDatesAndServiceName(c); err != nil {
			t.Fatal(err)
		}

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})
}
