package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"subscription-service/internal/model"

	"github.com/google/uuid"
)

type MockRepository struct {
	CreateFunc       func(ctx context.Context, sub *model.Subscription) error
	GetByIDFunc      func(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	ListFunc         func(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]model.Subscription, error)
	UpdateFunc       func(ctx context.Context, sub *model.Subscription) error
	DeleteFunc       func(ctx context.Context, id uuid.UUID) error
	GetTotalCostFunc func(ctx context.Context, userID *uuid.UUID, serviceName string, from, to time.Time) (model.RUB, error)
}

func (m *MockRepository) Create(ctx context.Context, sub *model.Subscription) error {
	return m.CreateFunc(ctx, sub)
}
func (m *MockRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *MockRepository) List(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]model.Subscription, error) {
	return m.ListFunc(ctx, userID, limit, offset)
}
func (m *MockRepository) Update(ctx context.Context, sub *model.Subscription) error {
	return m.UpdateFunc(ctx, sub)
}
func (m *MockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFunc(ctx, id)
}
func (m *MockRepository) GetTotalCost(ctx context.Context, userID *uuid.UUID, serviceName string, from, to time.Time) (model.RUB, error) {
	return m.GetTotalCostFunc(ctx, userID, serviceName, from, to)
}

func TestSubscriptionService_CalculateTotalCost(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	userID := uuid.New()

	tests := []struct {
		name        string
		from        time.Time
		until       time.Time
		mockRet     model.RUB
		mockErr     error
		wantErr     bool
		expectedRes model.RUB
	}{
		{
			name:        "Valid period",
			from:        time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			until:       time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			mockRet:     1000,
			mockErr:     nil,
			wantErr:     false,
			expectedRes: 1000,
		},
		{
			name:        "Invalid period (from after until)",
			from:        time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			until:       time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr:     true,
			expectedRes: 0,
		},
		{
			name:        "Repository error",
			from:        time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			until:       time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			mockRet:     0,
			mockErr:     errors.New("db error"),
			wantErr:     true,
			expectedRes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{
				GetTotalCostFunc: func(ctx context.Context, uid *uuid.UUID, sn string, f, u time.Time) (model.RUB, error) {
					return tt.mockRet, tt.mockErr
				},
			}
			svc := NewSubscriptionService(repo, logger)

			got, err := svc.CalculateTotalCost(context.Background(), &userID, "", tt.from, tt.until)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateTotalCost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expectedRes {
				t.Errorf("CalculateTotalCost() got = %v, want %v", got, tt.expectedRes)
			}
		})
	}
}

func TestSubscriptionService_List(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	userID := uuid.New()

	tests := []struct {
		name    string
		userID  *uuid.UUID
		limit   int
		offset  int
		mockRet []model.Subscription
		mockErr error
		wantErr bool
	}{
		{
			name:   "List for user",
			userID: &userID,
			limit:  10,
			offset: 0,
			mockRet: []model.Subscription{
				{ID: uuid.New(), ServiceName: "Netflix", UserID: userID},
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "List all (userID nil)",
			userID:  nil,
			limit:   5,
			offset:  0,
			mockRet: []model.Subscription{{ID: uuid.New()}, {ID: uuid.New()}},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "Repository error",
			userID:  &userID,
			limit:   10,
			offset:  0,
			mockRet: nil,
			mockErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{
				ListFunc: func(ctx context.Context, uid *uuid.UUID, l, o int) ([]model.Subscription, error) {
					if uid != tt.userID {
						t.Errorf("List() userID = %v, want %v", uid, tt.userID)
					}
					if l != tt.limit {
						t.Errorf("List() limit = %v, want %v", l, tt.limit)
					}
					if o != tt.offset {
						t.Errorf("List() offset = %v, want %v", o, tt.offset)
					}
					return tt.mockRet, tt.mockErr
				},
			}
			svc := NewSubscriptionService(repo, logger)

			got, err := svc.List(context.Background(), tt.userID, tt.limit, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != len(tt.mockRet) {
				t.Errorf("List() got %d items, want %d", len(got), len(tt.mockRet))
			}
		})
	}
}
