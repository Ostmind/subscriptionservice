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

	tests := []struct {
		name       string
		mockSetup  func(m *mock_server.MocksubscriptionManager)
		wantStatus int
	}{
		{
			name: "Success",
			mockSetup: func(m *mock_server.MocksubscriptionManager) {
				m.EXPECT().
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
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "ErrNotFound",
			mockSetup: func(m *mock_server.MocksubscriptionManager) {
				m.EXPECT().
					GetSubscriptionListByUserID(gomock.Any(), userID).
					Return(nil, models.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "InternalServerError",
			mockSetup: func(m *mock_server.MocksubscriptionManager) {
				m.EXPECT().
					GetSubscriptionListByUserID(gomock.Any(), userID).
					Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockManager := mock_server.NewMocksubscriptionManager(ctrl)
			tt.mockSetup(mockManager)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.AddCookie(&http.Cookie{Name: "userId", Value: userID.String()})
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := server.NewSubscriptionHandler(mockManager, logger)
			if err := handler.GetSubscriptionListByUserID(c); err != nil {
				t.Fatal(err)
			}
			if rec.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestPostSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	e := echo.New()

	tests := []struct {
		name       string
		jsonBody   string
		mockSetup  func(m *mock_server.MocksubscriptionManager)
		wantStatus int
	}{
		{
			name:     "BadRequest_BadJSON",
			jsonBody: `invalid json`,
			mockSetup: func(m *mock_server.MocksubscriptionManager) {
				// no mock calls expected
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "ErrUnique",
			jsonBody: `{"service_name": "Spotify", "price": 100, "start_date": "2023-09"}`,
			mockSetup: func(m *mock_server.MocksubscriptionManager) {
				m.EXPECT().
					PostSubscription(gomock.Any(), gomock.Any()).
					Return(models.ErrUnique)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "InternalServerError",
			jsonBody: `{"service_name": "Spotify", "price": 100, "start_date": "2023-09"}`,
			mockSetup: func(m *mock_server.MocksubscriptionManager) {
				m.EXPECT().
					PostSubscription(gomock.Any(), gomock.Any()).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:     "Success",
			jsonBody: `{"service_name": "Spotify", "price": 100, "start_date": "2023-09"}`,
			mockSetup: func(m *mock_server.MocksubscriptionManager) {
				m.EXPECT().
					PostSubscription(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockManager := mock_server.NewMocksubscriptionManager(ctrl)
			tt.mockSetup(mockManager)

			body := strings.NewReader(tt.jsonBody)
			req := httptest.NewRequest(http.MethodPost, "/", body)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := server.NewSubscriptionHandler(mockManager, logger)
			if err := handler.PostSubscription(c); err != nil {
				t.Fatal(err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestDeleteSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	e := echo.New()

	tests := []struct {
		name       string
		url        string
		mockSetup  func(m *mock_server.MocksubscriptionManager)
		wantStatus int
	}{
		{
			name: "Success",
			url:  "/?id=123",
			mockSetup: func(m *mock_server.MocksubscriptionManager) {
				m.EXPECT().
					DeleteSubscription(gomock.Any(), 123).
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "BadRequest_InvalidID",
			url:        "/?id=abc",
			mockSetup:  func(m *mock_server.MocksubscriptionManager) {}, // no expected mock calls (invalid ID parsing fails before mock)
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "NotFound",
			url:  "/?id=123",
			mockSetup: func(m *mock_server.MocksubscriptionManager) {
				m.EXPECT().
					DeleteSubscription(gomock.Any(), 123).
					Return(models.ErrNotFound)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "InternalServerError",
			url:  "/?id=123",
			mockSetup: func(m *mock_server.MocksubscriptionManager) {
				m.EXPECT().
					DeleteSubscription(gomock.Any(), 123).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockManager := mock_server.NewMocksubscriptionManager(ctrl)
			tt.mockSetup(mockManager)

			req := httptest.NewRequest(http.MethodDelete, tt.url, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := server.NewSubscriptionHandler(mockManager, logger)
			if err := handler.DeleteSubscription(c); err != nil {
				t.Fatal(err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestUpdateSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	e := echo.New()

	tests := []struct {
		name       string
		url        string
		jsonBody   string
		mockSetup  func(m *mock_server.MocksubscriptionManager)
		wantStatus int
	}{
		{
			name:     "Success",
			url:      "/?id=1",
			jsonBody: `{"service_name": "Spotify", "price": 300, "start_date": "2025-09"}`,
			mockSetup: func(m *mock_server.MocksubscriptionManager) {
				m.EXPECT().
					UpdateSubscription(gomock.Any(), gomock.Any(), 1).
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "BadRequest_InvalidID",
			url:        "/?id=abc",
			jsonBody:   `{"service_name": "Spotify", "price": 300, "start_date": "2025-09"}`,
			mockSetup:  func(m *mock_server.MocksubscriptionManager) {},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "BadRequest_BindError",
			url:        "/?id=1",
			jsonBody:   `invalid json`,
			mockSetup:  func(m *mock_server.MocksubscriptionManager) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "InternalServerError_ManagerError",
			url:      "/?id=1",
			jsonBody: `{"service_name": "Spotify", "price": 300, "start_date": "2025-09"}`,
			mockSetup: func(m *mock_server.MocksubscriptionManager) {
				m.EXPECT().
					UpdateSubscription(gomock.Any(), gomock.Any(), 1).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockManager := mock_server.NewMocksubscriptionManager(ctrl)
			tt.mockSetup(mockManager)

			body := strings.NewReader(tt.jsonBody)
			req := httptest.NewRequest(http.MethodPut, tt.url, body)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := server.NewSubscriptionHandler(mockManager, logger)
			if err := handler.UpdateSubscription(c); err != nil {
				t.Fatal(err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestGetTotalPeriodCostByDatesAndServiceName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	e := echo.New()

	tests := []struct {
		name       string
		jsonBody   string
		mockSetup  func(m *mock_server.MocksubscriptionManager)
		wantStatus int
		wantBody   string
	}{
		{
			name:     "Success",
			jsonBody: `{"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"09-2025","end_date":"12-2025","service_name":["Spotify"]}`,
			mockSetup: func(m *mock_server.MocksubscriptionManager) {
				m.EXPECT().
					GetTotalPeriodCostByDatesAndServiceName(gomock.Any(), gomock.Any()).
					Return(1200, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"result":1200}`,
		},
		{
			name:       "BadRequest_BindError",
			jsonBody:   `invalid json`,
			mockSetup:  func(m *mock_server.MocksubscriptionManager) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "InternalServerError_ManagerError",
			jsonBody: `{"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"09-2025","end_date":"12-2025","service_name":["Spotify"]}`,
			mockSetup: func(m *mock_server.MocksubscriptionManager) {
				m.EXPECT().
					GetTotalPeriodCostByDatesAndServiceName(gomock.Any(), gomock.Any()).
					Return(0, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockManager := mock_server.NewMocksubscriptionManager(ctrl)
			tt.mockSetup(mockManager)

			body := strings.NewReader(tt.jsonBody)
			req := httptest.NewRequest(http.MethodPost, "/", body)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := server.NewSubscriptionHandler(mockManager, logger)
			if err := handler.GetTotalPeriodCostByDatesAndServiceName(c); err != nil {
				t.Fatal(err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			if tt.wantBody != "" {
				if strings.TrimSpace(rec.Body.String()) != tt.wantBody {
					t.Errorf("expected body %s, got %s", tt.wantBody, rec.Body.String())
				}
			}
		})
	}
}
