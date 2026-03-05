package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRepo struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepo(pool *pgxpool.Pool) *SubscriptionRepo {
	return &SubscriptionRepo{pool: pool}
}

type CreateSubscriptionParams struct {
	ServiceName string
	Price       int
	UserID      string
	StartMonth  time.Time
	EndMonth    *time.Time
}

type SubscriptionRow struct {
	ID          string
	ServiceName string
	Price       int
	UserID      string
	StartMonth  time.Time
	EndMonth    *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (r *SubscriptionRepo) Create(ctx context.Context, p CreateSubscriptionParams) (SubscriptionRow, error) {
	const q = `
		INSERT INTO subscriptions (service_name, price, user_id, start_month, end_month)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, service_name, price, user_id, start_month, end_month, created_at, updated_at;
	`
	var row SubscriptionRow
	err := r.pool.QueryRow(ctx, q,
		p.ServiceName,
		p.Price,
		p.UserID,
		p.StartMonth,
		p.EndMonth,
	).Scan(
		&row.ID,
		&row.ServiceName,
		&row.Price,
		&row.UserID,
		&row.StartMonth,
		&row.EndMonth,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	return row, err
}
