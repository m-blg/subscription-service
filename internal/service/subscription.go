package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"subscription-service/internal/model"

	"github.com/google/uuid"
)

type SubscriptionService struct {
	repo   Repository
	logger *slog.Logger
}

func NewSubscriptionService(repo Repository, logger *slog.Logger) *SubscriptionService {
	return &SubscriptionService{
		repo:   repo,
		logger: logger.With("layer", "service"),
	}
}

func (s *SubscriptionService) Create(ctx context.Context, sub *model.Subscription) error {
	log := s.logger.With(
		"user_id", sub.UserID,
		"service_name", sub.ServiceName,
	)

	log.Debug("attempting to create subscription")

	if err := s.repo.Create(ctx, sub); err != nil {
		log.Error("failed to create subscription", "error", err)
		return err
	}

	log.Info("subscription created successfully", "id", sub.ID)
	return nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	log := s.logger.With("id", id)
	log.Debug("attempting to get subscription by ID")

	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("failed to get subscription by ID", "error", err)
		return nil, err
	}

	log.Info("subscription found", "id", sub.ID)
	return sub, nil
}

func (s *SubscriptionService) List(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]model.Subscription, error) {
	log := s.logger.With("user_id", userID, "limit", limit, "offset", offset)
	log.Debug("attempting to list subscriptions")

	subs, err := s.repo.List(ctx, userID, limit, offset)
	if err != nil {
		log.Error("failed to list subscriptions", "error", err)
		return nil, err
	}

	log.Info("subscriptions found", "count", len(subs))
	return subs, nil
}

func (s *SubscriptionService) Update(ctx context.Context, sub *model.Subscription) error {
	log := s.logger.With("id", sub.ID)
	log.Debug("attempting to update subscription")

	if err := s.repo.Update(ctx, sub); err != nil {
		log.Error("failed to update subscription", "error", err)
		return err
	}

	log.Info("subscription updated successfully", "id", sub.ID)
	return nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	log := s.logger.With("id", id)
	log.Debug("attempting to delete subscription")

	if err := s.repo.Delete(ctx, id); err != nil {
		log.Error("failed to delete subscription", "error", err)
		return err
	}

	log.Info("subscription deleted successfully", "id", id)
	return nil
}

func (s *SubscriptionService) CalculateTotalCost(ctx context.Context, userID *uuid.UUID, serviceName string, from, until time.Time) (model.RUB, error) {
	layout := model.LayoutMMYYYY
	periodStr := from.Format(layout) + " until " + until.Format(layout)
	log := s.logger.With("user_id", userID, "service_name", serviceName, "period", periodStr)
	log.Debug("attempting to calculate total cost")

	if from.After(until) {
		log.Debug("from.After(until) validation failed")
		return 0, fmt.Errorf("%w: start date cannot be after end date", model.ErrValidation)
	}

	totalCost, err := s.repo.GetTotalCost(ctx, userID, serviceName, from, until)
	if err != nil {
		log.Error("failed to calculate total cost", "error", err)
		return 0, err
	}

	log.Info("total cost calculated", "amount", totalCost)
	return totalCost, nil
}
