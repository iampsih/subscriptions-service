package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
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

type ListSubscriptionsParams struct {
	UserID      *string
	ServiceName *string
	Limit       int
	Offset      int
}

type ListSubscriptionsResult struct {
	Items []SubscriptionRow
	Total int64
}

type UpdateSubscriptionParams struct {
	ID          string
	ServiceName string
	Price       int
	UserID      string
	StartMonth  time.Time
	EndMonth    *time.Time
}

type TotalParams struct {
	FromMonth   time.Time
	ToMonth     time.Time
	UserID      *string
	ServiceName *string
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

func (r *SubscriptionRepo) GetByID(ctx context.Context, id string) (SubscriptionRow, error) {
	const q = `
		SELECT id, service_name, price, user_id, start_month, end_month, created_at, updated_at
		FROM subscriptions
		WHERE id = $1;
	`

	var row SubscriptionRow

	err := r.pool.QueryRow(ctx, q, id).Scan(
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

func (r *SubscriptionRepo) Delete(ctx context.Context, id string) error {
	const q = `
		DELETE FROM subscriptions
		WHERE id = $1;
	`

	cmd, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *SubscriptionRepo) List(ctx context.Context, p ListSubscriptionsParams) (ListSubscriptionsResult, error) {
	const qCount = `
		SELECT COUNT(*)
		FROM subscriptions
		WHERE ($1::uuid IS NULL OR user_id = $1)
		AND ($2::text IS NULL OR service_name ILIKE '%' || $2 || '%');
	`
	var total int64
	if err := r.pool.QueryRow(ctx, qCount, p.UserID, p.ServiceName).Scan(&total); err != nil {
		return ListSubscriptionsResult{}, err
	}

	const qItems = `
		SELECT id, service_name, price, user_id, start_month, end_month, created_at, updated_at
		FROM subscriptions
		WHERE ($1::uuid IS NULL OR user_id = $1)
		AND ($2::text IS NULL OR service_name ILIKE '%' || $2 || '%')
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4;
	`

	rows, err := r.pool.Query(ctx, qItems, p.UserID, p.ServiceName, p.Limit, p.Offset)
	if err != nil {
		return ListSubscriptionsResult{}, err
	}
	defer rows.Close()

	items := make([]SubscriptionRow, 0, p.Limit)
	for rows.Next() {
		var row SubscriptionRow
		if err := rows.Scan(
			&row.ID,
			&row.ServiceName,
			&row.Price,
			&row.UserID,
			&row.StartMonth,
			&row.EndMonth,
			&row.CreatedAt,
			&row.UpdatedAt,
		); err != nil {
			return ListSubscriptionsResult{}, err
		}
		items = append(items, row)
	}
	if err := rows.Err(); err != nil {
		return ListSubscriptionsResult{}, err
	}

	return ListSubscriptionsResult{
		Items: items,
		Total: total,
	}, nil
}

func (r *SubscriptionRepo) Update(ctx context.Context, p UpdateSubscriptionParams) (SubscriptionRow, error) {
	const q = `
		UPDATE subscriptions
		SET service_name = $2,
			price = $3,
			user_id = $4,
			start_month = $5,
			end_month = $6,
			updated_at = now()
		WHERE id = $1
		RETURNING id, service_name, price, user_id, start_month, end_month, created_at, updated_at;
		`
	var row SubscriptionRow
	err := r.pool.QueryRow(ctx, q,
		p.ID, p.ServiceName, p.Price, p.UserID, p.StartMonth, p.EndMonth,
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

func (r *SubscriptionRepo) Total(ctx context.Context, p TotalParams) (int64, error) {
	const q = `
		WITH params AS (
		SELECT $1::date AS from_m, $2::date AS to_m
		)
		SELECT COALESCE(SUM(
		s.price * (
			(EXTRACT(YEAR FROM LEAST(COALESCE(s.end_month, p.to_m), p.to_m))::int * 12
			+ EXTRACT(MONTH FROM LEAST(COALESCE(s.end_month, p.to_m), p.to_m))::int)
			-
			(EXTRACT(YEAR FROM GREATEST(s.start_month, p.from_m))::int * 12
			+ EXTRACT(MONTH FROM GREATEST(s.start_month, p.from_m))::int)
			+ 1
		)
		), 0) AS total
		FROM subscriptions s
		CROSS JOIN params p
		WHERE
		GREATEST(s.start_month, p.from_m) <= LEAST(COALESCE(s.end_month, p.to_m), p.to_m)
		AND ($3::uuid IS NULL OR s.user_id = $3)
		AND ($4::text IS NULL OR s.service_name ILIKE '%' || $4 || '%');
	`
	var total int64
	err := r.pool.QueryRow(ctx, q, p.FromMonth, p.ToMonth, p.UserID, p.ServiceName).Scan(&total)
	return total, err
}
