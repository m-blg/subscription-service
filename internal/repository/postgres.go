package repository

import (
	"context"
	"errors"
	"time"

	"subscription-service/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRepo struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepo(pool *pgxpool.Pool) *SubscriptionRepo {
	return &SubscriptionRepo{pool: pool}
}

// NOTE: Updates sub with the generated fields from the database
func (r *SubscriptionRepo) Create(ctx context.Context, sub *model.Subscription) error {
	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)

	if err != nil {
		return err
	}
	return nil
}

func (r *SubscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at 
	          FROM subscriptions WHERE id = $1`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	sub, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.Subscription])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &sub, nil
}

// if `userID` is provided, filters subscriptions by user ID
func (r *SubscriptionRepo) List(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]model.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at 
	          FROM subscriptions 
	          WHERE ($1::UUID IS NULL OR user_id = $1)
	          ORDER BY start_date DESC
	          LIMIT $2 OFFSET $3`

	rows, _ := r.pool.Query(ctx, query, userID, limit, offset)
	subs, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.Subscription])
	if err != nil {
		return nil, err
	}
	return subs, nil
}

func (r *SubscriptionRepo) Update(ctx context.Context, sub *model.Subscription) error {
	query := `
		UPDATE subscriptions 
		SET service_name = $1, 
		    price = $2, 
		    user_id = $3,
		    start_date = $4, 
		    end_date = $5,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
		sub.ID,
	).Scan(&sub.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return err
	}
	return nil
}

func (r *SubscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *SubscriptionRepo) GetTotalCost(ctx context.Context, userID *uuid.UUID, serviceName string, from, to time.Time) (model.RUB, error) {
	query := `
		SELECT COALESCE(SUM(price), 0)
		FROM subscriptions
		WHERE ($1::uuid IS NULL OR user_id = $1)
		  AND ($2 = '' OR service_name = $2)
		  AND start_date <= $4
		  AND (end_date IS NULL OR end_date >= $3)`

	var total model.RUB
	err := r.pool.QueryRow(ctx, query, userID, serviceName, from, to).Scan(&total)
	return total, err
}
