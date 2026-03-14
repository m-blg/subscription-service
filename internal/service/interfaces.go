package service

import (
	"context"
	"subscription-service/internal/model"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, sub *model.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	List(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]model.Subscription, error)
	Update(ctx context.Context, sub *model.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetTotalCost(ctx context.Context, userID *uuid.UUID, serviceName string, from, to time.Time) (model.RUB, error)
}
