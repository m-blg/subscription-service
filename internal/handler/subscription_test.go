package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"subscription-service/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MockService struct {
	CreateFunc             func(ctx context.Context, sub *model.Subscription) error
	GetByIDFunc            func(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	ListFunc               func(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]model.Subscription, error)
	UpdateFunc             func(ctx context.Context, sub *model.Subscription) error
	DeleteFunc             func(ctx context.Context, id uuid.UUID) error
	CalculateTotalCostFunc func(ctx context.Context, userID *uuid.UUID, serviceName string, from, until time.Time) (model.RUB, error)
}

func (m *MockService) Create(ctx context.Context, sub *model.Subscription) error {
	return m.CreateFunc(ctx, sub)
}
func (m *MockService) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *MockService) List(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]model.Subscription, error) {
	return m.ListFunc(ctx, userID, limit, offset)
}
func (m *MockService) Update(ctx context.Context, sub *model.Subscription) error {
	return m.UpdateFunc(ctx, sub)
}
func (m *MockService) Delete(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFunc(ctx, id)
}
func (m *MockService) CalculateTotalCost(ctx context.Context, userID *uuid.UUID, serviceName string, from, until time.Time) (model.RUB, error) {
	return m.CalculateTotalCostFunc(ctx, userID, serviceName, from, until)
}

func TestSubscriptionHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()

	tests := []struct {
		name           string
		url            string
		mockRet        []model.Subscription
		mockErr        error
		expectedStatus int
	}{
		{
			name: "List all subscriptions",
			url:  "/subscriptions?limit=5&offset=0",
			mockRet: []model.Subscription{
				{ID: uuid.New(), ServiceName: "Netflix"},
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "List subscriptions for user",
			url:  "/subscriptions?user_id=" + userID.String(),
			mockRet: []model.Subscription{
				{ID: uuid.New(), ServiceName: "Spotify", UserID: userID},
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid user_id",
			url:            "/subscriptions?user_id=invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockService{
				ListFunc: func(ctx context.Context, uid *uuid.UUID, l, o int) ([]model.Subscription, error) {
					return tt.mockRet, tt.mockErr
				},
			}
			h := NewSubscriptionHandler(mockSvc)
			r := gin.New()
			r.GET("/subscriptions", h.List)

			req, _ := http.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("List() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestSubscriptionHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()

	tests := []struct {
		name           string
		body           interface{}
		mockErr        error
		expectedStatus int
	}{
		{
			name: "Valid request",
			body: SubscriptionRequest{
				ServiceName: "Netflix",
				Price:       299,
				UserID:      userID,
				StartDate:   model.MonthYear(time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)),
			},
			mockErr:        nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Missing required fields",
			body: map[string]interface{}{
				"price": 299,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockService{
				CreateFunc: func(ctx context.Context, sub *model.Subscription) error {
					return tt.mockErr
				},
			}
			h := NewSubscriptionHandler(mockSvc)
			r := gin.New()
			r.POST("/subscriptions", h.Create)

			jsonBody, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Create() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestSubscriptionHandler_GetTotalCost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()

	tests := []struct {
		name           string
		url            string
		mockRet        model.RUB
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "Valid request",
			url:            "/subscriptions/total?user_id=" + userID.String() + "&from=01-2025&until=12-2025",
			mockRet:        1500,
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid date format",
			url:            "/subscriptions/total?user_id=" + userID.String() + "&from=2025-01&until=2025-12",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing user_id case",
			url:            "/subscriptions/total?from=01-2025&until=12-2025",
			mockRet:        1501,
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockService{
				CalculateTotalCostFunc: func(ctx context.Context, uid *uuid.UUID, sn string, f, u time.Time) (model.RUB, error) {
					return tt.mockRet, tt.mockErr
				},
			}
			h := NewSubscriptionHandler(mockSvc)
			r := gin.New()
			r.GET("/subscriptions/total", h.GetTotalCost)

			req, _ := http.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("GetTotalCost() status = %v, want %v, body: %s", w.Code, tt.expectedStatus, w.Body.String())
			}
		})
	}
}
